import { API_BASE_URL } from "../constants";
import type { PokemonCard } from "../types";

export async function fetchCards(page: number): Promise<PokemonCard[]> {
  const res = await fetch(`${API_BASE_URL}/api/cards?page=${page}`);
  if (!res.ok) throw new Error(`Failed to fetch cards: ${res.status}`);
  return res.json();
}
