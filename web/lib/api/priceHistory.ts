import type { PricePoint, PriceHistoryPeriod } from "../types";

function createRng(seed: number): () => number {
  let s = Math.abs(seed) || 1;
  return () => {
    s = (s * 16807) % 2147483647;
    return s / 2147483647;
  };
}

function hashCode(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (Math.imul(31, hash) + str.charCodeAt(i)) | 0;
  }
  return Math.abs(hash);
}

const PERIOD_DAYS: Record<PriceHistoryPeriod, number> = {
  "7D": 7,
  "30D": 30,
  "90D": 90,
};

// TODO: Replace mock body with fetch(`${API_BASE_URL}/api/cards/${cardId}/price-history?period=${period}`)
export async function fetchPriceHistory(
  cardId: string,
  period: PriceHistoryPeriod
): Promise<PricePoint[]> {
  const seed = hashCode(cardId + period);
  const rng = createRng(seed);
  const days = PERIOD_DAYS[period];

  const basePrice = 15000 + (seed % 85000);
  const trend = seed % 3 === 0 ? -0.003 : 0.004;

  const points: PricePoint[] = [];
  let price = basePrice;
  const now = new Date();

  for (let i = days - 1; i >= 0; i--) {
    const date = new Date(now);
    date.setDate(date.getDate() - i);

    const noise = (rng() - 0.5) * 0.06;
    price = price * (1 + trend + noise);
    price = Math.max(price, 1000);

    points.push({
      date: date.toISOString().split("T")[0],
      price: Math.round(price),
    });
  }

  return points;
}
