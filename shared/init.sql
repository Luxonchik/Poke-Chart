CREATE TABLE IF NOT EXISTS master_cards (
    id SERIAL PRIMARY KEY,
    name_ko TEXT NOT NULL,
    rarity TEXT,
    set_name TEXT,
    card_number TEXT,
    image_url TEXT,
    UNIQUE(name_ko, set_name, card_number) -- Чтобы не было дублей
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
    master_card_id INTEGER REFERENCES master_cards(id) -- Ссылка на эталон
);

ALTER TABLE market_listings
ADD COLUMN IF NOT EXISTS image_url TEXT;