# API Endpoints

Source of truth: `internal/routes/routes.go`.

```mermaid
flowchart LR
  Client[Client / Web / Scraper] --> Mux[http.ServeMux]

  Mux --> CardsRoute["GET /api/cards"]
  CardsRoute --> CardHandler["CardHandler.GetCards"]
  CardHandler --> CardRepo["CardRepo.GetAllCards(limit=30, offset)"]
  CardRepo --> MasterCards[(master_cards)]

  Mux --> ListingsRoute["POST /api/listings"]
  ListingsRoute --> ListingHandler["ListingHandler.ReceiveFromScraper"]
  ListingHandler --> ListingRepo["ListingRepo.CreateListing"]
  ListingRepo --> MarketListings[(market_listings)]

  Mux --> VisionRoute["POST /api/test-vision"]
  VisionRoute --> VisionHandler["TestVisionHandler"]
  VisionHandler --> VisionService["vision.RecognizeCard(image_url)"]
  VisionService --> Gemini[Gemini API]
```

## Endpoint Map

| Method | Path | Handler | Input | Success response | Storage/service |
| --- | --- | --- | --- | --- | --- |
| `GET` | `/api/cards?page=1` | `CardHandler.GetCards` | Query `page`, optional positive integer. Defaults to `1`. | `200` JSON array of cards with active listings. Empty result is `[]`. | Reads `master_cards` INNER JOIN `market_listings`. |
| `GET` | `/api/catalog?page=1&q=&on_sale=1` | `CardHandler.GetCatalog` | Query `page` (same rules as `/api/cards`), optional `q` search string, optional `on_sale=1`. | `200` JSON array of all cards. | Reads `master_cards` LEFT JOIN `market_listings`. |
| `GET` | `/api/stats` | `StatsHandler.GetStats` | — | `200` JSON object with market totals. | Reads `master_cards`, `market_listings`. |
| `POST` | `/api/listings` | `ListingHandler.ReceiveFromScraper` | JSON listing payload from scraper. `daangn_id` is required. | `201` when inserted, `200` when duplicate already existed. | Inserts/updates `market_listings`. |
| `POST` | `/api/test-vision` | `TestVisionHandler` | JSON with `image_url`. | `200` JSON vision result. | Calls `vision.RecognizeCard`. |

## GET /api/cards

Returns cards from `master_cards`, ordered by `id`.

Query params:

| Name | Type | Required | Default | Notes |
| --- | --- | --- | --- | --- |
| `page` | integer | No | `1` | Values `<= 0` or invalid values fall back to `1`. Page size is fixed at `30`. |

Response item:

```json
{
  "id": "string",
  "name": "string",
  "image_url": "string",
  "card_no": "string",
  "rarity": "string",
  "series": "string",
  "listing_count": 0,
  "min_price": 42000,
  "variant_count": 1
}
```

Rows of `master_cards` with the same `name`, `card_no` and `series` (regular and mirror versions) are returned as one item: `id` is the lowest id, `listing_count`/`min_price` cover all versions, `variant_count` is how many were merged.

`min_price` is the lowest active listing price in KRW, or `null` when there are none.

Errors:

| Status | Cause |
| --- | --- |
| `500` | Database query failed. |

## GET /api/catalog

Returns every printing from `master_cards`, cards with active listings first (`listing_count DESC, id`), including cards without listings. Same response item as `GET /api/cards`; `listing_count` is `0` for cards with no active listings.

Query params:

| Name | Type | Required | Default | Notes |
| --- | --- | --- | --- | --- |
| `page` | integer | No | `1` | Same rules as `/api/cards`. |
| `q` | string | No | `""` | Case-insensitive substring of `name` or `card_no`. LIKE wildcards are escaped; truncated to 100 characters. |
| `on_sale` | `1` | No | — | When `1`, only printings with at least one active listing. |

## GET /api/stats

Response:

```json
{
  "cards_tracked": 5012,
  "active_listings": 14,
  "ai_processed": 63,
  "last_scraped_at": "2026-10-27T06:30:00Z"
}
```

`ai_processed` counts listings in a final worker status (`active`, `rejected`, `error`, `closed`). `last_scraped_at` is `null` when there are no listings.

Errors:

| Status | Cause |
| --- | --- |
| `500` | Database query failed. |

## POST /api/listings

Receives listings from the scraper and stores them in `market_listings`.

Request:

```json
{
  "daangn_id": "string",
  "title": "string",
  "price": 0,
  "image_url": "string",
  "source_url": "string",
  "created_at": "2026-06-24 12:00:00",
  "boosted_at": "2026-06-24 12:00:00"
}
```

Behavior:

- Inserts a new row with `status = "pending"`.
- Uses `external_id = daangn_id`.
- If `external_id` already exists, does not insert a duplicate.
- For duplicates, fills missing `image_url`, `created_at`, and `boosted_at` when possible.

Success responses:

```json
{
  "status": "saved",
  "message": "new listing saved"
}
```

```json
{
  "status": "ignored",
  "message": "listing already exists"
}
```

Errors:

| Status | Cause |
| --- | --- |
| `400` | Invalid JSON. |
| `400` | Missing `daangn_id`. |
| `500` | Database insert/update failed. |

## POST /api/test-vision

Temporary endpoint for testing card recognition by image URL.

Request:

```json
{
  "image_url": "https://example.com/card.jpg"
}
```

Response:

```json
{
  "name": "string",
  "rarity": "string",
  "card_no": "string",
  "is_pokemon_card": true
}
```

Errors:

| Status | Cause |
| --- | --- |
| `400` | Invalid JSON or missing expected request shape. |
| `500` | Vision recognition failed. |

