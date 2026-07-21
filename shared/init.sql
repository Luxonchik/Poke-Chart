-- NOTE: This schema is expected by services/api.
-- master_cards is the "official reference" table.
CREATE TABLE IF NOT EXISTS master_cards (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    image_url TEXT,
    card_no TEXT,
    rarity TEXT,
    series TEXT
);

CREATE TABLE IF NOT EXISTS market_listings (
    id          SERIAL PRIMARY KEY,
    source      TEXT NOT NULL DEFAULT 'daangn',  -- площадка: daangn / joonggonara / bunjang / ...
    external_id TEXT NOT NULL,
    title       TEXT,
    price       INTEGER,
    image_url   TEXT,
    url         TEXT,
    status      TEXT,                             -- pending / processing / active / rejected / error / closed
    created_at  TIMESTAMP,
    boosted_at  TIMESTAMP,
    closed_at   TIMESTAMP,                        -- NULL = ещё активно, NOT NULL = исчезло с площадки
    scraped_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ai_error_msg TEXT,
    ai_card_no   TEXT,
    ai_series    TEXT,
    master_card_id TEXT REFERENCES master_cards(id),
    CONSTRAINT market_listings_source_external_id_key UNIQUE (source, external_id)
);

CREATE TABLE IF NOT EXISTS price_snapshots (
    id         SERIAL PRIMARY KEY,
    card_id    TEXT NOT NULL REFERENCES master_cards(id),
    listing_id INTEGER NOT NULL REFERENCES market_listings(id),
    price      INTEGER NOT NULL,
    source     TEXT NOT NULL,
    snapped_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_market_listings_status      ON market_listings(status);
CREATE INDEX IF NOT EXISTS idx_market_listings_master_card ON market_listings(master_card_id);
CREATE INDEX IF NOT EXISTS idx_market_listings_source      ON market_listings(source);
CREATE INDEX IF NOT EXISTS idx_price_snapshots_card_id     ON price_snapshots(card_id);
CREATE INDEX IF NOT EXISTS idx_price_snapshots_snapped_at  ON price_snapshots(snapped_at);

-- Миграции для существующих инсталляций (ALTER IF NOT EXISTS идемпотентны)
ALTER TABLE market_listings ADD COLUMN IF NOT EXISTS source       TEXT NOT NULL DEFAULT 'daangn';
ALTER TABLE market_listings ADD COLUMN IF NOT EXISTS closed_at    TIMESTAMP;
ALTER TABLE market_listings ADD COLUMN IF NOT EXISTS ai_card_no   TEXT;
ALTER TABLE market_listings ADD COLUMN IF NOT EXISTS ai_series    TEXT;

-- Одноразовая чистка битых URL
UPDATE market_listings
SET url = REPLACE(url, 'https://www.daangn.comhttps://www.daangn.com', 'https://www.daangn.com')
WHERE url LIKE 'https://www.daangn.comhttps://%';
