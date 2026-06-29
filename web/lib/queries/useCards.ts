"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import { fetchCards } from "@/lib/api/cards";
import { CARDS_PAGE_SIZE } from "@/lib/constants";

export function useInfiniteCards() {
  return useInfiniteQuery({
    queryKey: ["cards"],
    queryFn: ({ pageParam }) => fetchCards(pageParam),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.length === CARDS_PAGE_SIZE ? allPages.length + 1 : undefined,
  });
}
