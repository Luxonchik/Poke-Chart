"use client";

import { keepPreviousData, useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { fetchCards } from "@/lib/api/cards";
import { CARDS_PAGE_SIZE } from "@/lib/constants";
import type { CardsEndpoint, CatalogFilters } from "@/lib/types";

export const cardKeys = {
  all: ["cards"] as const,
  list: (path: CardsEndpoint) => [...cardKeys.all, "list", path] as const,
  catalog: (filters: CatalogFilters) => [...cardKeys.all, "catalog", filters] as const,
  onMarket: () => [...cardKeys.all, "on-market"] as const,
};

export function useInfiniteCards(path: CardsEndpoint = "/api/cards") {
  return useInfiniteQuery({
    queryKey: cardKeys.list(path),
    queryFn: ({ pageParam }) => fetchCards({ page: pageParam, path }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.length === CARDS_PAGE_SIZE ? allPages.length + 1 : undefined,
  });
}

export function useCatalogCards(filters: CatalogFilters) {
  return useInfiniteQuery({
    queryKey: cardKeys.catalog(filters),
    queryFn: ({ pageParam }) => fetchCards({ page: pageParam, path: "/api/catalog", ...filters }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.length === CARDS_PAGE_SIZE ? allPages.length + 1 : undefined,
    // Keep the current cards on screen while a new search is loading.
    placeholderData: keepPreviousData,
  });
}

/** First page of cards that currently have active listings. */
export function useOnMarketCards() {
  return useQuery({
    queryKey: cardKeys.onMarket(),
    queryFn: () => fetchCards({ page: 1, path: "/api/cards" }),
  });
}
