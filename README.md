# PokeChart

Price aggregator and tracker for Pokemon TCG cards on the Korean secondary market. It collects marketplace listings, identifies the card on each photo with an AI vision model, matches it against a master card database, and surfaces price trends through a web interface.

> Personal learning project. Built to explore Go microservices, AI vision integration, and full-stack data pipelines. Not affiliated with Nintendo, The Pokemon Company, or any marketplace.

## What it does

- Collects public "for sale" listings for Pokemon cards from a Korean marketplace.
- Runs each listing photo through **Gemini Vision** to recognize the card (name, rarity, card number, series) - in Korean.
- Matches recognized listings against a master card database, using a second image-comparison pass to disambiguate look-alikes.
- Stores listings and builds **price history** per card.
- Serves everything to a Next.js web app with an infinite card grid, card detail modal, and price charts.

## Architecture

Four services plus PostgreSQL, orchestrated with Taskfile and Docker Compose:

```
┌─────────────────┐   POST    ┌──────────────┐   GET    ┌──────────────┐
│ scraper-carrot  │ ────────► │   API (Go)   │ ◄─────── │  web (Next)  │
│ (Go collector)  │  listings │    :8080     │  cards   │    :3001     │
└─────────────────┘           └──────┬───────┘          └──────────────┘
                                      │
                              ┌───────▼────────┐
                              │   PostgreSQL   │
                              │     :5433      │
                              └───────┬────────┘
                                      │ seed (one-off)
                              ┌───────▼────────┐
                              │ master-seeder  │
                              └────────────────┘

Vision worker (inside API) ──► Gemini API for card recognition & matching
```

| Service | Language | Role |
| --- | --- | --- |
| `services/api` | Go | REST API, vision worker, card/listing repositories |
| `services/scraper-carrot` | Go | Collects marketplace listings, posts them to the API |
| `services/master-seeder` | Go | One-off loader for the master card database |
| `web` | Next.js / TypeScript | Frontend: card grid, detail modal, price history charts |
| PostgreSQL | - | `master_cards`, `market_listings`, price history |

## Tech stack

**Backend:** Go, `net/http`, PostgreSQL, Gemini Vision API
**Frontend:** Next.js (App Router), TypeScript, TanStack Query, Tailwind CSS, Zustand
**Infra:** Docker Compose, Taskfile

## Key engineering details

- **AI vision with hallucination guards.** The recognizer doesn't trust the model blindly: card numbers are regex-validated, rarity tokens are stripped from names, series labels are normalized, and the image-matching pass only accepts a candidate ID it was actually given, with a confidence threshold. Temperature is pinned to 0.
- **Two-stage card matching.** First recognize the card from a single photo; when several master cards look alike, a second Gemini pass compares the listing photo against reference images and picks the best match (or declines).
- **Resilient collection.** The collector parses the marketplace's embedded page state, runs on a fixed interval, and de-duplicates listings by marketplace ID on the API side.

## Getting started

Prerequisites: Go 1.22+, Node 20+, Docker, [Task](https://taskfile.dev).

```bash
# 1. Configure environment
cp services/api/.env.example services/api/.env
cp services/scraper-carrot/.env.example services/scraper-carrot/.env
cp web/.env.example web/.env
# then fill in GEMINI_API_KEY in services/api/.env

# 2. Start PostgreSQL (Docker) + API + web
task db        # PostgreSQL on :5433
task api       # Go API on :8080
task web       # Next.js on :3001

# 3. Seed the master card database (one-off)
task seed

# 4. Run the collector
task scraper
```

## API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/cards?page=1` | List master cards (page size 30) |
| `GET` | `/api/cards/{cardId}/listings` | Market listings for a card |
| `POST` | `/api/sync` | Ingest listings from the collector |

## Disclaimer

This is an educational project. Collected data is public marketplace information used only for personal, non-commercial analysis. The project respects reasonable request rates and does not redistribute personal data. Pokemon and Pokemon TCG are trademarks of their respective owners.
