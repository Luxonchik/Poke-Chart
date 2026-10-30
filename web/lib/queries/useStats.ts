"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchMarketStats } from "@/lib/api/stats";

export const statsKeys = {
  market: ["stats", "market"] as const,
};

export function useMarketStats() {
  return useQuery({
    queryKey: statsKeys.market,
    queryFn: fetchMarketStats,
    staleTime: 60_000,
  });
}
