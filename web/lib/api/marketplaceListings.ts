import { API_BASE_URL } from "../constants";
import type { MarketListing } from "../types";

export async function fetchMarketplaceListings(
  cardId: string
): Promise<MarketListing[]> {
  const res = await fetch(`${API_BASE_URL}/api/cards/${cardId}/listings`);
  if (!res.ok) throw new Error(`Failed to fetch listings: ${res.status}`);
  return res.json();
}
