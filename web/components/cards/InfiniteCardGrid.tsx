"use client";

import { useInView } from "react-intersection-observer";
import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { useInfiniteCards } from "@/lib/queries/useCards";
import { PokemonCard } from "./PokemonCard";

export function InfiniteCardGrid() {
  const t = useTranslations("cards");
  const { ref, inView } = useInView();

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    status,
  } = useInfiniteCards("/api/catalog");

  useEffect(() => {
    if (inView && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage]);

  if (status === "pending") return <div className="text-center py-10 text-muted-foreground">{t("loading")}</div>;
  if (status === "error") return <div className="text-center py-10 text-destructive">{t("error")}</div>;

  return (
    <section id="catalog" className="scroll-mt-24 space-y-8">
      <div className="text-foreground">
        <h2 className="text-2xl font-bold tracking-tight">{t("title")}</h2>
        <p className="text-muted-foreground mt-2">{t("subtitle")}</p>
      </div>
      <div className="grid grid-cols-[repeat(auto-fill,minmax(14.25rem,1fr))] gap-7">
        {data.pages.map((page, i) => (
          <div className="contents" key={i}>
            {page.map((card) => (
              <PokemonCard key={card.id} card={card} />
            ))}
          </div>
        ))}
      </div>

      <div ref={ref} className="h-10 flex items-center justify-center">
        {isFetchingNextPage && <span className="text-muted-foreground animate-pulse">{t("loadingMore")}</span>}
        {!hasNextPage && <span className="text-muted-foreground">{t("noMore")}</span>}
      </div>
    </section>
  );
}
