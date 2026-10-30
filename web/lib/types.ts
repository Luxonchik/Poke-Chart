export interface PokemonCard {
  id: string;
  name: string;
  image_url: string;
  card_no: string;
  rarity: string;
  series: string;
  listing_count: number;
  /** Lowest active listing price in KRW; null when the card has no listings. */
  min_price: number | null;
  /** How many master_cards rows share this printing (regular + mirror = 2). */
  variant_count: number;
}

export type CardsEndpoint = "/api/cards" | "/api/catalog";

export interface CardsParams {
  page: number;
  path: CardsEndpoint;
  /** Name or card number substring; only applied by /api/catalog. */
  query?: string;
  /** Only cards with active listings; only applied by /api/catalog. */
  onSale?: boolean;
}

export interface CatalogFilters {
  query: string;
  onSale: boolean;
}

export type CatalogView = "grid" | "table";

export interface MarketStats {
  cards_tracked: number;
  active_listings: number;
  ai_processed: number;
  last_scraped_at: string | null;
}

export type PriceHistoryPeriod = "7D" | "30D" | "90D";

export interface PricePoint {
  date: string;
  price: number;
}

export interface MarketListing {
  id: number;
  daangn_id: string;
  title: string;
  price: number;
  image_url: string;
  source_url: string;
  master_card_id: string;
  status: string;
  created_at: string;
  updated_at: string;
}
