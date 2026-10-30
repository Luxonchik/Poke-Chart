import { API_BASE_URL } from "../constants";
import type { MarketStats } from "../types";

export async function fetchMarketStats(): Promise<MarketStats> {
  const res = await fetch(`${API_BASE_URL}/api/stats`);
  if (!res.ok) throw new Error(`Failed to fetch stats: ${res.status}`);
  return res.json();
}
