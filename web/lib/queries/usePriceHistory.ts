"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchPriceHistory } from "@/lib/api/priceHistory";
import type { PriceHistoryPeriod } from "@/lib/types";

export function usePriceHistory(cardId: string, period: PriceHistoryPeriod) {
  return useQuery({
    queryKey: ["priceHistory", cardId, period],
    queryFn: () => fetchPriceHistory(cardId, period),
    enabled: cardId !== "",
  });
}
