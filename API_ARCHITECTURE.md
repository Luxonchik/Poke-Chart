# PokeChart — Архитектура API и поток данных

## Обзор системы

Проект состоит из 4 сервисов + PostgreSQL:

```
┌─────────────────┐     ┌──────────────┐     ┌──────────────┐
│  scraper-carrot │────▶│   API (Go)   │◀────│  web (Next)  │
│  (Go, 당근마켓)  │POST │  :8080       │ GET │  :3001       │
└─────────────────┘     └──────┬───────┘     └──────────────┘
                               │
                        ┌──────▼───────┐
                        │  PostgreSQL  │
                        │  :5433       │
                        └──────┬───────┘
                               │
                        ┌──────▼───────┐
                        │ master-seeder│
                        │  (одноразово)│
                        └──────────────┘
```

## Запуск

```bash
# Всё сразу (БД + API + Web)
task

# По отдельности
task db        # Docker: PostgreSQL на порту 5433
task api       # Go API на :8080
task web       # Next.js на :3001
task scraper   # Скрапер 당근마켓
```

## База данных

**Подключение:** `postgres://admin:admin_password@localhost:5433/pokecart`

### Таблица `master_cards` — эталонный справочник

Содержит ~5000 официальных покемон-карт, спарсенных с pokemoncard.co.kr через `master-seeder`.

| Колонка    | Тип  | Описание                                    |
|------------|------|---------------------------------------------|
| id         | TEXT PK | ID с сайта pokemoncard.co.kr             |
| name       | TEXT | Название карты (корейский, напр. "피카츄 ex") |
| image_url  | TEXT | URL официального изображения                |
| card_no    | TEXT | Номер карты (напр. "017/086")               |
| rarity     | TEXT | Редкость (SAR, UR, SR, AR, RR, ex и т.д.)   |
| series     | TEXT | Серия (напр. "스칼렛&바이올렛")               |

### Таблица `market_listings` — объявления с рынка

Содержит объявления, спарсенные скрапером с 당근마켓. Каждое объявление проходит через AI-воркер для распознавания и привязки к эталонной карте.

| Колонка        | Тип       | Описание                                      |
|----------------|-----------|-----------------------------------------------|
| id             | SERIAL PK | Автоинкремент                                 |
| external_id    | TEXT UQ   | ID объявления на 당근마켓                       |
| title          | TEXT      | Заголовок объявления                           |
| price          | INTEGER   | Цена в KRW                                    |
| image_url      | TEXT      | URL фото из объявления                         |
| url            | TEXT      | Ссылка на объявление                           |
| status         | TEXT      | Статус обработки (см. ниже)                    |
| master_card_id | TEXT FK   | Ссылка на `master_cards.id` (NULL до привязки) |
| ai_error_msg   | TEXT      | Сообщение об ошибке AI                         |
| ai_card_no     | TEXT      | Номер карты, распознанный AI                   |
| ai_series      | TEXT      | Серия, распознанная AI                         |
| created_at     | TIMESTAMP | Дата создания объявления на 당근마켓             |
| boosted_at     | TIMESTAMP | Дата буста объявления                          |
| scraped_at     | TIMESTAMP | Когда мы спарсили                              |
| updated_at     | TIMESTAMP | Последнее обновление записи                    |

**Статусы `status`:**

```
pending ──▶ processing ──▶ active     (AI распознал + нашёл в master_cards)
                      ├──▶ error      (AI распознал, но карты нет в справочнике)
                      └──▶ rejected   (фото — не покемон-карта)
```

## Сервисы

### 1. `services/master-seeder` — Сидер эталонных карт

**Запуск:** одноразовый скрипт, заполняет `master_cards`.

**Что делает:**
1. Ходит на `pokemoncard.co.kr` через POST-запросы к пагинации
2. Для каждой карты парсит HTML-страницу детали: имя, card_no, rarity, series
3. Записывает в `master_cards` с `ON CONFLICT (id) DO UPDATE`

**Источник image_url:** `https://cards.image.pokemonkorea.co.kr/data/` + путь из API

### 2. `services/scraper-carrot` — Скрапер 당근마켓

**Запуск:** `task scraper` (бесконечный цикл, каждые 5 минут)

**Что делает:**
1. Запрашивает `daangn.com` с поиском "포켓몬 EX" в районе 역삼동
2. Парсит HTML через Remix-контекст (`window.__remixContext`)
3. Фильтрует объявления (отбрасывает нерелевантные по ключевым словам)
4. Отправляет каждое валидное объявление на `POST /api/listings`

**Поток данных:**
```
당근마켓 HTML ──▶ RemixContext JSON ──▶ []Article ──▶ фильтр ──▶ POST /api/listings
```

### 3. `services/api` — Go REST API

**Запуск:** `task api` (порт :8080)

**Структура:**
```
services/api/
├── cmd/api/main.go              # Точка входа, CORS, запуск воркера
├── internal/
│   ├── models/
│   │   ├── card.go              # Card, CardWithListingCount
│   │   ├── listing.go           # MarketListing, CreateListingRequest
│   │   └── vision.go            # VisionResult
│   ├── repository/
│   │   ├── card_repo.go         # CRUD для master_cards
│   │   └── listing_repo.go      # CRUD для market_listings
│   ├── handlers/
│   │   ├── card_handler.go      # GET /api/cards
│   │   ├── listing_handler.go   # POST /api/listings, GET /api/cards/{id}/listings
│   │   └── vision_handler.go    # POST /api/test-vision
│   ├── routes/routes.go         # Регистрация маршрутов
│   ├── workers/vision_worker.go # Фоновый AI-воркер
│   └── services/vision/gemini.go # Интеграция с Gemini API
```

#### Эндпоинты

##### `GET /api/cards?page=1`

Возвращает карты из `master_cards`, у которых есть хотя бы одно **активное** объявление на рынке. Отсортированы по количеству объявлений (DESC).

**SQL:** INNER JOIN `market_listings` WHERE `status = 'active'`, GROUP BY, ORDER BY `listing_count DESC`.

**Ответ:**
```json
[
  {
    "id": "SV7s-017",
    "name": "피카츄 ex",
    "image_url": "https://cards.image.pokemonkorea.co.kr/data/...",
    "card_no": "017/086",
    "rarity": "SAR",
    "series": "스칼렛&바이올렛 「스텔라 미라클」",
    "listing_count": 5
  }
]
```

Пагинация: 30 карт на страницу. `page=1` → offset 0, `page=2` → offset 30.

##### `GET /api/cards/{cardId}/listings`

Возвращает все активные объявления с рынка для конкретной карты.

**Ответ:**
```json
[
  {
    "id": 42,
    "daangn_id": "abc123",
    "title": "피카츄 SAR 미개봉",
    "price": 45000,
    "image_url": "https://...",
    "source_url": "https://www.daangn.com/...",
    "master_card_id": "SV7s-017",
    "status": "active",
    "created_at": "2025-06-20T10:30:00Z",
    "updated_at": "2025-06-20T12:00:00Z"
  }
]
```

Отсортированы по цене ASC (самые дешёвые первыми).

##### `POST /api/listings`

Принимает объявления от скрапера.

**Тело запроса:**
```json
{
  "daangn_id": "abc123",
  "title": "포켓몬 카드 피카츄",
  "price": 50000,
  "image_url": "https://...",
  "source_url": "https://www.daangn.com/...",
  "created_at": "2025-06-20T10:30:00+09:00",
  "boosted_at": ""
}
```

**Ответ:** `201` (новое) или `200` (дубликат по `external_id`).

##### `POST /api/test-vision`

Тестовый эндпоинт для проверки распознавания карты через Gemini.

**Тело:** `{"image_url": "https://..."}`

### 4. AI Vision Worker (фоновый воркер)

Запускается как горутина при старте API. Работает бесконечно.

**Что делает:**
1. Берёт одну запись из `market_listings` с `status = 'pending'` (с `FOR UPDATE SKIP LOCKED`)
2. Ставит `status = 'processing'`
3. Скачивает `image_url`, отправляет в **Gemini 2.5 Flash** для распознавания
4. Gemini возвращает: `name`, `rarity`, `card_no`, `series`, `is_pokemon_card`
5. Если не покемон-карта → `status = 'rejected'`
6. Ищет совпадение в `master_cards` (стратегии по приоритету):
   - `FindBestByCardNoThenName` — по номеру карты + уточнение по имени
   - `FindCandidateCardsByNameAndSeries` → если >1 кандидат, сравнение изображений через Gemini
   - `FindCardByNameAndSeries` — по имени и серии
   - `FindCardByNameAndRarity` — фоллбэк по имени и редкости
7. Если совпадение найдено → `status = 'active'`, `master_card_id = <id>`
8. Если не найдено → `status = 'error'`

**Rate limit:** 12 секунд между запросами к Gemini (лимит Google: 5 RPM).

## Фронтенд (web)

**Фреймворк:** Next.js 16 + React 19, TanStack Query, Zustand, Tailwind, next-intl (ko/en)

### Поток данных на фронте

```
GET /api/cards?page=N
  ↓
useInfiniteCards() — TanStack useInfiniteQuery
  ↓
InfiniteCardGrid — рендерит карточки с infinite scroll
  ↓ (клик по карте)
useModalStore.openCardDetail(card) — Zustand
  ↓
CardDetailModal — показывает:
  ├── Имя, серия, редкость, card_no (из выбранной карты)
  ├── Минимальная рыночная цена (из реальных листингов)
  ├── GET /api/cards/{id}/listings → useMarketplaceListings()
  │   └── MarketplaceListings — список объявлений с ценами и ссылками
  └── usePriceHistory() — история цен (пока мок, нет таблицы в БД)
```

### Ключевые файлы фронта

| Файл | Что делает |
|------|-----------|
| `lib/api/cards.ts` | `fetchCards(page)` → GET /api/cards?page=N |
| `lib/api/marketplaceListings.ts` | `fetchMarketplaceListings(cardId)` → GET /api/cards/{id}/listings |
| `lib/api/priceHistory.ts` | Мок-данные (TODO: нет таблицы в БД) |
| `lib/queries/useCards.ts` | `useInfiniteCards()` — infinite scroll хук |
| `lib/queries/useMarketplaceListings.ts` | `useMarketplaceListings(cardId)` |
| `lib/types.ts` | `PokemonCard`, `MarketListing`, `PricePoint` |
| `lib/store.ts` | Zustand: модалка выбранной карты |
| `components/cards/InfiniteCardGrid.tsx` | Грид карт с infinite scroll |
| `components/cards/PokemonCard.tsx` | Одна карточка (с бейджем listing_count) |
| `components/cards/CardDetailModal.tsx` | Модалка с деталями + листингами |
| `components/cards/MarketplaceListings.tsx` | Список объявлений с 당근마켓 |

## Общий поток: от объявления до экрана

```
1. Скрапер парсит 당근마켓 каждые 5 минут
2. POST /api/listings → market_listings (status=pending)
3. Vision Worker берёт pending → Gemini распознаёт → ищет в master_cards
4. Если нашёл → status=active, master_card_id=<id>
5. GET /api/cards → INNER JOIN → только карты с active листингами
6. Фронт показывает эти карты, отсортированные по listing_count DESC
7. Клик → модалка → GET /api/cards/{id}/listings → реальные объявления
```
