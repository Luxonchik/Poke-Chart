"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchMarketplaceListings } from "@/lib/api/marketplaceListings";

export function useMarketplaceListings(cardId: string) {
  return useQuery({
    queryKey: ["marketplaceListings", cardId],
    queryFn: () => fetchMarketplaceListings(cardId),
    enabled: cardId !== "",
  });
}
