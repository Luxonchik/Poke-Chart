# PokeChart — Roadmap

Документ разбит на два блока:
- **Часть 1** — критические фиксы (баги, безопасность, надёжность)
- **Часть 2** — новая логика (price_history, жизненный цикл объявлений, расширение скрапера)

Каждый шаг независим и может быть сделан отдельным коммитом.

---

## Часть 1. Критические фиксы

### Шаг 1. DB-строка из переменной окружения

**Файл:** `services/api/cmd/api/main.go`

Сейчас:
```go
connStr := "postgres://admin:admin_password@localhost:5433/pokecart?sslmode=disable"
```

Заменить на:
```go
connStr := os.Getenv("DATABASE_URL")
if connStr == "" {
    connStr = "postgres://admin:admin_password@localhost:5433/pokecart?sslmode=disable"
}
```

И добавить в `services/api/.env`:
```
DATABASE_URL=postgres://admin:admin_password@localhost:5433/pokecart?sslmode=disable
```

---

### Шаг 2. Логировать ошибки db.Exec в воркере

**Файл:** `services/api/internal/workers/vision_worker.go`

Все вызовы `db.Exec(...)` без проверки ошибок — обернуть:
```go
if _, err := db.Exec("UPDATE market_listings SET status = 'pending' ...", ...); err != nil {
    log.Printf("⚠️ Не удалось обновить статус: %v", err)
}
```
Затронуты строки: 53, 62, 64, 67, 103, 107.

---

### Шаг 3. Убрать утечку внутренних ошибок наружу

**Файл:** `services/api/internal/handlers/listing_handler.go:55`

Сейчас:
```go
http.Error(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
```

Заменить на:
```go
log.Printf("❌ Ошибка CreateListing: %v", err)
http.Error(w, "Internal server error", http.StatusInternalServerError)
```

---

### Шаг 4. ON CONFLICT — один запрос вместо двух

**Файл:** `services/api/internal/repository/listing_repo.go`

Заменить INSERT + отдельный UPDATE на:
```sql
INSERT INTO market_listings (external_id, title, price, image_url, url, status, created_at, boosted_at)
VALUES ($1, $2, $3, $4, $5, 'pending', NULLIF($6, '')::timestamp, NULLIF($7, '')::timestamp)
ON CONFLICT (external_id) DO UPDATE SET
    image_url  = COALESCE(NULLIF(EXCLUDED.image_url, ''),  market_listings.image_url),
    created_at = COALESCE(market_listings.created_at, EXCLUDED.created_at),
    boosted_at = COALESCE(market_listings.boosted_at, EXCLUDED.boosted_at)
```

`RowsAffected` при `DO UPDATE` возвращает 1, поэтому `isNew` определять по другому признаку:
```go
// Добавить в SELECT перед вставкой: EXISTS(SELECT 1 FROM market_listings WHERE external_id=$1)
// или хранить xmax в RETURNING — проще всего просто отдельным SELECT до вставки.
```
Самый простой вариант — `SELECT EXISTS(...)` до `INSERT`, вернуть результат.

---

### Шаг 5. Убрать test-эндпоинт из продакшена

**Файл:** `services/api/internal/routes/routes.go`

```go
// Удалить эту строку:
mux.HandleFunc("POST /api/test-vision", handlers.TestVisionHandler)
```

Если нужен для отладки — добавить флаг `DEBUG=true` в `.env` и регистрировать условно.

---

### Шаг 6. Исправить двойную фильтрацию в скрапере

**Файл:** `services/scraper-carrot/cmd/main.go`

`ScrapeService.Scrape()` уже фильтрует. В `main.go` убрать повторную проверку:
```go
// Удалить:
if !cardFilter.IsValid(item) {
    continue
}
```
И убрать передачу `cardFilter` в `main.go` (он нужен только внутри `ScrapeService`).

---

### Шаг 7. URL скрапера и адрес API из конфига

**Файл:** `services/scraper-carrot/cmd/main.go` и `output/console_printer.go`

```go
// main.go
searchURL := os.Getenv("DAANGN_SEARCH_URL")
if searchURL == "" {
    searchURL = "https://www.daangn.com/kr/buy-sell/s/?in=역삼동-6035&search=포켓몬+EX"
}
```

```go
// console_printer.go
apiURL := os.Getenv("API_URL")
if apiURL == "" {
    apiURL = "http://localhost:8080"
}
response, err := http.Post(apiURL+"/api/listings", ...)
```

Добавить `services/scraper-carrot/.env`:
```
DAANGN_SEARCH_URL=https://www.daangn.com/kr/buy-sell/s/?in=역삼동-6035&search=포켓몬+EX
API_URL=http://localhost:8080
```

---

### Шаг 8. Ограничить `computePoWNonce` по итерациям

**Файл:** `services/scraper-carrot/internal/client/daangn_client.go`

```go
func computePoWNonce(challenge string, difficulty int) (int, error) {
    if difficulty <= 0 {
        return 0, nil
    }
    targetPrefix := strings.Repeat("0", difficulty)
    for nonce := 0; nonce < 10_000_000; nonce++ {
        hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", challenge, nonce)))
        if strings.HasPrefix(hex.EncodeToString(hash[:]), targetPrefix) {
            return nonce, nil
        }
    }
    return 0, fmt.Errorf("PoW не решён за 10M итераций (difficulty=%d)", difficulty)
}
```

---

### Шаг 9. Удалить закомментированный код

**Файл:** `services/scraper-carrot/internal/output/console_printer.go`

Удалить весь блок `// func PrintSummary` и `// func PrintArticle` — они не используются.

---

### Шаг 10. Починить `listing_repo` — REPLACE в SQL как миграция

**Файл:** `shared/init.sql` — добавить одноразовую миграцию:
```sql
UPDATE market_listings
SET url = REPLACE(url, 'https://www.daangn.comhttps://www.daangn.com', 'https://www.daangn.com')
WHERE url LIKE 'https://www.daangn.comhttps://%';
```

После этого убрать `REPLACE(...)` из SELECT в `listing_repo.go:23`.

---

## Часть 2. Price History и жизненный цикл объявлений

### Шаг 11. Новая схема БД

**Файл:** `shared/init.sql` — добавить в конец:

```sql
-- Жизненный цикл объявления
ALTER TABLE market_listings ADD COLUMN IF NOT EXISTS closed_at TIMESTAMP;
-- closed_at IS NULL  → объявление активно
-- closed_at NOT NULL → объявление закрыто (продано или снято)

-- История цен: один снимок на каждое "активное" состояние объявления
CREATE TABLE IF NOT EXISTS price_snapshots (
    id          SERIAL PRIMARY KEY,
    card_id     TEXT NOT NULL REFERENCES master_cards(id),
    listing_id  INTEGER NOT NULL REFERENCES market_listings(id),
    price       INTEGER NOT NULL,
    source      TEXT NOT NULL DEFAULT 'daangn', -- площадка
    snapped_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_price_snapshots_card_id   ON price_snapshots(card_id);
CREATE INDEX IF NOT EXISTS idx_price_snapshots_snapped_at ON price_snapshots(snapped_at);
CREATE INDEX IF NOT EXISTS idx_market_listings_status     ON market_listings(status);
CREATE INDEX IF NOT EXISTS idx_market_listings_master_card ON market_listings(master_card_id);
```

**Почему так, а не хранить историю прямо в `market_listings`?**
`price_snapshots` — это агрегированная аналитическая таблица. Одно объявление может
появляться много раз (скрапер обходит каждые 5 минут), поэтому снимки цены фиксируются
отдельно и не дублируют данные объявления.

---

### Шаг 12. Записывать снимок цены при матчинге карточки

**Файл:** `services/api/internal/workers/vision_worker.go`

После успешного `UPDATE market_listings SET status = 'active'` добавить:
```go
_, err = db.Exec(`
    INSERT INTO price_snapshots (card_id, listing_id, price, source)
    VALUES ($1, $2, $3, 'daangn')
    ON CONFLICT DO NOTHING
`, masterCard.ID, id, /* price нужно достать из market_listings */ )
```

Для этого при получении задачи из очереди — тянуть и `price`:
```go
// Изменить RETURNING в воркере:
RETURNING id, image_url, price
// и добавить переменную:
var price int
err := db.QueryRow(query).Scan(&id, &imageUrl, &price)
```

---

### Шаг 13. Отслеживать закрытые объявления

Сейчас объявления никогда не закрываются. Нужен отдельный воркер (или второй цикл
в существующем) — **checker**, который раз в час/день проходит по `active`-листингам
и проверяет, ещё ли они живы на Daangn.

**Файл:** `services/api/internal/workers/listing_checker.go` (новый файл)

Логика:
```
1. SELECT id, url FROM market_listings WHERE status = 'active' AND closed_at IS NULL
2. HEAD-запрос на url (или GET с проверкой 404)
3. Если 404 или признак "объявление закрыто" → UPDATE SET status='closed', closed_at=NOW()
```

Это позволит на фронте показывать реальные "проданные" цены, а не только "хотят продать".

---

### Шаг 14. Новый API-эндпоинт: история цен карточки

**Файл:** `services/api/internal/handlers/price_handler.go` (новый файл)

```
GET /api/cards/{cardId}/price-history?days=30
```

Возвращает:
```json
{
  "card_id": "sv1-025",
  "median": 15000,
  "min": 8000,
  "max": 25000,
  "snapshots": [
    { "price": 15000, "snapped_at": "2026-06-01T12:00:00Z" },
    ...
  ]
}
```

SQL:
```sql
SELECT price, snapped_at
FROM price_snapshots
WHERE card_id = $1
  AND snapped_at >= NOW() - INTERVAL '30 days'
ORDER BY snapped_at ASC
```

Агрегаты считать прямо в Go или вторым запросом:
```sql
SELECT
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY price) AS median,
    MIN(price),
    MAX(price)
FROM price_snapshots
WHERE card_id = $1 AND snapped_at >= NOW() - INTERVAL '30 days'
```

---

### Шаг 15. Расширить скрапер: несколько запросов / вся Корея

**Файл:** `services/scraper-carrot/cmd/main.go`

Вместо одного URL — список:
```go
searchQueries := []string{
    "포켓몬 EX",
    "포켓몬 SAR",
    "포켓몬 카드",
}
regions := []string{
    "역삼동-6035",
    "강남구-1168",
    // добавлять по мере необходимости
}
```

Или использовать эндпоинт без фильтра региона (проверить, поддерживает ли Daangn).

---

### Шаг 16. Показывать цену на фронте

**Порядок работы:**

1. `GET /api/cards` — уже есть, добавить `min_price` и `median_price` в ответ
   (одним JOIN к `price_snapshots` за последние 30 дней)

2. `GET /api/cards/{cardId}/price-history` — новый эндпоинт (Шаг 14)

3. На странице карточки — график цены (shadcn + Recharts или простая таблица снимков)

**SQL для обогащения `/api/cards`:**
```sql
SELECT
    mc.id, mc.name, mc.image_url, mc.card_no, mc.rarity, mc.series,
    COUNT(DISTINCT ml.id) AS listing_count,
    MIN(ps.price)         AS min_price,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY ps.price) AS median_price
FROM master_cards mc
INNER JOIN market_listings ml ON ml.master_card_id = mc.id AND ml.status = 'active'
LEFT JOIN price_snapshots ps  ON ps.card_id = mc.id
    AND ps.snapped_at >= NOW() - INTERVAL '30 days'
GROUP BY mc.id
ORDER BY listing_count DESC
LIMIT $1 OFFSET $2
```

---

## Порядок выполнения (рекомендуемый)

```
Шаги 1-3   → безопасность и надёжность (30 мин)
Шаги 4-6   → чистота кода (30 мин)
Шаги 7-10  → конфигурация и мелкие фиксы (20 мин)
Шаг 11     → миграция схемы (10 мин, требует пересоздания БД или ALTER)
Шаги 12-13 → воркеры записи и проверки (2-3 часа)
Шаг 14     → новый эндпоинт (1 час)
Шаги 15-16 → скрапер + фронт (по ситуации)
```

---

## Что НЕ входит в этот план (намеренно)

- Авторизация / роли — не нужна пока нет личного кабинета
- Другие площадки (Joonggonara, Bunjang) — после того как pipeline с Daangn стабилен
- Автопополнение `master_cards` — отдельная большая задача (скрапинг официального сайта TCG)
- Полнотекстовый поиск — после того как базовый функционал готов
