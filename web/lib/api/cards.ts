import { API_BASE_URL } from "../constants";
import type { CardsParams, PokemonCard } from "../types";

export async function fetchCards({ page, path, query, onSale }: CardsParams): Promise<PokemonCard[]> {
  const search = new URLSearchParams({ page: String(page) });
  if (query) search.set("q", query);
  if (onSale) search.set("on_sale", "1");

  const res = await fetch(`${API_BASE_URL}${path}?${search}`);
  if (!res.ok) throw new Error(`Failed to fetch cards: ${res.status}`);
  return res.json();
}
