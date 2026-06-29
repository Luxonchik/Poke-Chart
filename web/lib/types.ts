export interface PokemonCard {
  id: string;
  name: string;
  image_url: string;
  card_no: string;
  rarity: string;
  series: string;
  listing_count: number;
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
