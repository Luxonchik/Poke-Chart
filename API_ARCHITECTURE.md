# PokeChart — Архитектура API и поток данных

## Обзор системы

Проект состоит из 4 сервисов + PostgreSQL:

```
┌─────────────────┐     ┌──────────────┐     ┌──────────────┐
│  scraper-carrot │────▶│   API (Go)   │◀────│  web (Next)  │
│  (Go, 당근마켓)  │POST │  :9000       │ GET │  :4000       │
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
task api       # Go API на :9000
task web       # Next.js на :4000
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

**Запуск:** `task scraper` (бесконечный цикл, каждые 5 минут). Отладка одного прохода: `go run ./debug` из папки сервиса — печатает, сколько объявлений на странице и сколько прошло фильтр.

**Что делает:**
1. Запрашивает страницу поиска `daangn.com` ("포켓몬 EX", район 역삼동; адрес — `DAANGN_SEARCH_URL`)
2. Достаёт `window.__remixContext` и берёт объявления из `state.loaderData["routes/kr.search.buy-sell._index"].buySellArticles` — они отрендерены на сервере, отдельного API и PoW больше нет (сайт изменился в октябре 2026)
3. Если этого роута на странице нет — возвращает ошибку `unexpected page structure`, а не пустой список, чтобы смена вёрстки была видна в логе
4. Фильтрует объявления: цена ≥ ₩3 000, без цен-заглушек (123456, 12345), без стоп-слов и оптовых лотов
5. Отправляет весь снимок одним запросом на `POST /api/sync` (`API_URL`, по умолчанию `http://localhost:9000`)

**Поток данных:**
```
당근마켓 HTML ──▶ __remixContext ──▶ buySellArticles ──▶ фильтр ──▶ POST /api/sync
```

### 3. `services/api` — Go REST API

**Запуск:** `task api` (порт :9000)

**Структура:**
```
services/api/
├── cmd/api/main.go              # Точка входа, CORS, запуск воркера
├── internal/
│   ├── models/
│   │   ├── card.go              # Card, CardWithListingCount
│   │   ├── listing.go           # MarketListing, CreateListingRequest
│   │   ├── stats.go             # MarketStats
│   │   └── vision.go            # VisionResult
│   ├── repository/
│   │   ├── card_repo.go         # CRUD для master_cards
│   │   └── listing_repo.go      # CRUD для market_listings, сводка GetMarketStats
│   ├── handlers/
│   │   ├── card_handler.go      # GET /api/cards, GET /api/catalog
│   │   ├── stats_handler.go     # GET /api/stats
│   │   ├── listing_handler.go   # POST /api/listings, GET /api/cards/{id}/listings
│   │   └── vision_handler.go    # POST /api/test-vision
│   ├── routes/routes.go         # Регистрация маршрутов
│   ├── workers/vision_worker.go # Фоновый AI-воркер
│   └── services/vision/gemini.go # Интеграция с Gemini API
```

#### Эндпоинты

##### `GET /api/cards?page=1`

Возвращает карты из `master_cards`, у которых есть хотя бы одно **активное** объявление на рынке. Отсортированы по количеству объявлений (DESC).

**SQL:** карты схлопываются в «печати» (см. ниже), INNER JOIN `market_listings` WHERE `status = 'active'`, GROUP BY, ORDER BY `listing_count DESC`.

**Печати и зеркальные версии.** В официальном справочнике обычная и зеркальная (holo) версии одной карты — две строки `master_cards` с разными `id` и одинаковыми `name`, `card_no`, `series`; признака «зеркальная» в данных нет. `GET /api/cards` и `GET /api/catalog` группируют такие строки в одну (`printingsSubquery` в `card_repo.go`): представитель — наименьший `id`, объявления и `min_price` считаются по всем версиям, `variant_count` — сколько версий объединено. `GET /api/cards/{cardId}/listings` тоже возвращает объявления всех версий этой печати. Данные в БД не меняются.

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
    "listing_count": 5,
    "min_price": 42000,
    "variant_count": 1
  }
]
```

`min_price` — минимальная цена среди активных объявлений в KRW (`null`, если объявлений нет).

Пагинация: 30 карт на страницу. `page=1` → offset 0, `page=2` → offset 30.

##### `GET /api/catalog?page=1&q=...&on_sale=1`

Полный справочник `master_cards`, включая карты без объявлений. Используется для каталога на главной.

**SQL:** печати (как у `/api/cards`), LEFT JOIN `market_listings` (только `status = 'active'`), GROUP BY, ORDER BY `listing_count DESC, id` — сначала карты с объявлениями.

**Поиск:** `q` — подстрока имени или номера карты, без учёта регистра (`ILIKE`). Спецсимволы `%`, `_`, `\` экранируются, длина обрезается до 100 символов. Пустой `q` — без фильтра.

**Только в продаже:** `on_sale=1` оставляет карты, у которых есть хотя бы одно активное объявление (`HAVING COUNT(ml.id) > 0`).

**Ответ:** тот же формат, что у `GET /api/cards` (у карт без объявлений `listing_count` = 0, `min_price` = `null`). Пагинация: 30 карт на страницу.

##### `GET /api/stats`

Сводные цифры для полосы статистики на главной.

**Ответ:**
```json
{
  "cards_tracked": 5012,
  "active_listings": 14,
  "ai_processed": 63,
  "last_scraped_at": "2026-10-27T06:30:00Z"
}
```

- `cards_tracked` — строк в `master_cards`
- `active_listings` — объявления со `status = 'active'`
- `ai_processed` — объявления, которые воркер уже разобрал (`active`, `rejected`, `error`, `closed`)
- `last_scraped_at` — `MAX(scraped_at)` по `market_listings`, `null` если объявлений нет

##### `GET /api/cards/{cardId}/listings`

Возвращает все активные объявления с рынка для конкретной карты, включая её другие версии (та же печать).

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
3. Скачивает `image_url`, отправляет в Gemini для распознавания (модель — `GEMINI_MODEL`, по умолчанию `gemini-3.8-flash`; список доступных: `GET https://generativelanguage.googleapis.com/v1beta/models?key=…`)
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

**Ошибки Gemini:**
- `429` / `503` (модель перегружена, `vision.ErrModelBusy`) — объявление возвращается в `pending` без штрафа, воркер делает паузу с экспоненциальным backoff: 30 с → 1 мин → 2 мин → 4 мин → 5 мин (максимум). Первый успешный ответ сбрасывает паузу.
- Любая другая ошибка (битая картинка, невалидный JSON) — счётчик попыток хранится в начале `ai_error_msg` (`[2/3] причина`). После 3 попыток → `status = 'error'`.

**Лог:** одна строка на объявление с итогом и тем, что распознал Gemini:
```
🔍 ID=7690 → active: 리자몽 ex 201/165  (AI: 리자몽 ex · SAR · 201/165)
🔍 ID=7691 → error: нет в справочнике  (AI: 메가리자몽X ex · SAR · 110/080)
🔍 ID=7692 → rejected: не одиночная карта
🔍 ID=7693 → отложено: Gemini временно перегружен (статус 503)
⏸ Gemini перегружен, пауза 30s
```

## Фронтенд (web)

**Фреймворк:** Next.js 16 + React 19, TanStack Query, Zustand, Tailwind, next-intl (ko/en)

### Поток данных на фронте

```
GET /api/stats ──▶ useMarketStats() ──▶ StatsBar — полоса с цифрами

GET /api/cards?page=1 ──▶ useOnMarketCards() ──▶ OnMarketCarousel — карты с активными объявлениями

CatalogToolbar ──▶ useCatalogFilterStore { query (debounce 300 мс), onSale } — Zustand

GET /api/catalog?page=N&q=<query>&on_sale=1
  ↓
useCatalogCards({ query, onSale }) — TanStack useInfiniteQuery (keepPreviousData)
  ↓
CatalogSection — загрузка, пустые состояния, infinite scroll, переключатель вида (useState)
  ├── CatalogGrid → CardTile — сетка (по умолчанию)
  └── CatalogTable — таблица
  ↓ (клик по карте)
useModalStore.openCardDetail(card) — Zustand
  ↓
CardDetailModal — показывает:
  ├── Имя, серия, редкость, card_no (из выбранной карты)
  ├── Минимальная рыночная цена (из реальных листингов)
  ├── GET /api/cards/{id}/listings → useMarketplaceListings()
  │   └── MarketplaceListings — список объявлений с ценами и ссылками
  └── usePriceHistory() — история цен: пока мок, в модалке показан размытым с подписью «нет данных»
```

### Ключевые файлы фронта

| Файл | Что делает |
|------|-----------|
| `lib/api/cards.ts` | `fetchCards({ page, path, query, onSale })` → GET /api/cards или /api/catalog |
| `lib/api/stats.ts` | `fetchMarketStats()` → GET /api/stats |
| `lib/api/marketplaceListings.ts` | `fetchMarketplaceListings(cardId)` → GET /api/cards/{id}/listings |
| `lib/api/priceHistory.ts` | Мок-данные (TODO: нет таблицы в БД) |
| `lib/queries/useCards.ts` | `cardKeys`, `useCatalogCards(filters)`, `useOnMarketCards()`, `useInfiniteCards(path)` (для `InfiniteCardGrid`, сейчас не подключён) |
| `lib/queries/useStats.ts` | `statsKeys`, `useMarketStats()` |
| `lib/queries/useMarketplaceListings.ts` | `useMarketplaceListings(cardId)` |
| `lib/types.ts` | `PokemonCard`, `MarketListing`, `PricePoint` |
| `lib/store.ts` | Zustand: модалка выбранной карты, фильтры каталога (`query`, `onSale`) |
| `components/features/landing/CatalogSection.tsx` | Секция каталога: данные, состояния, infinite scroll, вид Grid/Table |
| `components/cards/CatalogToolbar.tsx` | Поиск, фильтр «On sale», переключатель вида |
| `components/cards/CatalogGrid.tsx` | Сетка плиток каталога |
| `components/cards/CardTile.tsx` | Плитка карты: картинка, имя, цена, номер |
| `components/cards/CatalogTable.tsx` | Каталог таблицей |
| `components/cards/InfiniteCardGrid.tsx` | Старый грид карт (сейчас не подключён) |
| `components/cards/PokemonCard.tsx` | Одна карточка (с бейджем listing_count) |
| `components/cards/CardDetailModal.tsx` | Модалка с деталями + листингами |
| `components/cards/MarketplaceListings.tsx` | Список объявлений с 당근마켓 |
| `components/layout/Container.tsx` | Общие горизонтальные границы (до 1600px) для шапки, полос и контента |
| `components/layout/Header.tsx` | Шапка: логотип, ссылки на секции `#market` / `#catalog` |
| `components/layout/HeaderSearch.tsx` | Прежнее поле поиска в шапке (сейчас не подключено, поиск перенесён в `CatalogToolbar`) |
| `components/features/landing/StatsBar.tsx` | Полоса статистики во всю ширину под шапкой |
| `components/features/landing/OnMarketCarousel.tsx` | Карусель (Embla, перетаскивание, без петли) карт с активными объявлениями |

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
