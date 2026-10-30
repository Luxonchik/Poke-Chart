"use client";

import { useEffect, useState } from "react";
import { useInView } from "react-intersection-observer";
import { useFormatter, useTranslations } from "next-intl";
import { CatalogGrid } from "@/components/cards/CatalogGrid";
import { CatalogTable } from "@/components/cards/CatalogTable";
import { CatalogToolbar } from "@/components/cards/CatalogToolbar";
import { useCatalogCards } from "@/lib/queries/useCards";
import { useMarketStats } from "@/lib/queries/useStats";
import { useCatalogFilterStore } from "@/lib/store";
import type { CatalogView } from "@/lib/types";
import { cn } from "@/lib/utils";

export function CatalogSection() {
  const t = useTranslations("cards");
  const format = useFormatter();
  const [view, setView] = useState<CatalogView>("grid");
  const query = useCatalogFilterStore((s) => s.query);
  const onSale = useCatalogFilterStore((s) => s.onSale);
  const { data: stats } = useMarketStats();
  const { ref, inView } = useInView({ rootMargin: "600px" });

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPlaceholderData, status } =
    useCatalogCards({ query, onSale });

  useEffect(() => {
    if (inView && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage]);

  const cards = data?.pages.flat() ?? [];

  return (
    <section id="catalog" className="scroll-mt-24 space-y-4">
      <div className="flex items-baseline gap-3">
        <h2 className="text-xl font-semibold text-foreground">{t("title")}</h2>
        {stats && (
          <span className="font-mono text-sm text-muted-foreground">
            {t("total", { count: format.number(stats.cards_tracked) })}
          </span>
        )}
      </div>

      <CatalogToolbar view={view} onViewChange={setView} />

      {status === "pending" && <p className="py-10 text-center text-muted-foreground">{t("loading")}</p>}
      {status === "error" && <p className="py-10 text-center text-destructive">{t("error")}</p>}
      {status === "success" && cards.length === 0 && (
        <p className="py-10 text-center text-muted-foreground">
          {query ? t("empty", { query }) : t("emptyFiltered")}
        </p>
      )}

      {cards.length > 0 && (
        <div className={cn("transition-opacity", isPlaceholderData && "opacity-60")}>
          {view === "grid" ? <CatalogGrid cards={cards} /> : <CatalogTable cards={cards} />}
        </div>
      )}

      {cards.length > 0 && (
        <div ref={ref} className="flex h-10 items-center justify-center text-sm text-muted-foreground">
          {isFetchingNextPage && <span className="animate-pulse">{t("loadingMore")}</span>}
          {!hasNextPage && <span>{t("noMore")}</span>}
        </div>
      )}
    </section>
  );
}
