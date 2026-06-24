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
    id SERIAL PRIMARY KEY,
    external_id TEXT UNIQUE,
    title TEXT,
    price INTEGER,
    image_url TEXT,
    status TEXT,
    url TEXT,
    created_at TIMESTAMP,
    boosted_at TIMESTAMP,
    scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ai_error_msg TEXT,
    ai_card_no TEXT,
    ai_series TEXT,
    master_card_id TEXT REFERENCES master_cards(id) -- Ссылка на эталон
);

ALTER TABLE market_listings
ADD COLUMN IF NOT EXISTS image_url TEXT;

ALTER TABLE market_listings ADD COLUMN IF NOT EXISTS ai_card_no TEXT;
ALTER TABLE market_listings ADD COLUMN IF NOT EXISTS ai_series TEXT;