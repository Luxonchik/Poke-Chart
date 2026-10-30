"use client";

import { useFormatter, useNow, useTranslations } from "next-intl";
import { Container } from "@/components/layout/Container";
import { useMarketStats } from "@/lib/queries/useStats";

const EMPTY_VALUE = "—";

export function StatsBar() {
  const t = useTranslations("landing.stats");
  const format = useFormatter();
  const now = useNow({ updateInterval: 60_000 });
  const { data } = useMarketStats();

  const lastSync = data?.last_scraped_at
    ? format.relativeTime(new Date(data.last_scraped_at), now)
    : EMPTY_VALUE;

  const items = [
    { id: "cards", label: t("cardsTracked"), value: data ? format.number(data.cards_tracked) : EMPTY_VALUE },
    { id: "listings", label: t("activeListings"), value: data ? format.number(data.active_listings) : EMPTY_VALUE },
    { id: "ai", label: t("aiProcessed"), value: data ? format.number(data.ai_processed) : EMPTY_VALUE },
    { id: "sync", label: t("lastSync"), value: lastSync },
  ];

  return (
    <div className="border-b border-border">
      <Container>
        <dl className="grid grid-cols-2 gap-y-2 py-3 md:grid-cols-4 md:py-0">
          {items.map((item) => (
            <div
              key={item.id}
              className="flex flex-col gap-1.5 py-2 md:border-l md:border-border md:py-5 md:pl-8 md:first:border-l-0 md:first:pl-0"
            >
              <dt className="text-[11px] uppercase tracking-wider text-muted-foreground">{item.label}</dt>
              <dd className="font-mono text-2xl font-medium tabular-nums text-foreground">{item.value}</dd>
            </div>
          ))}
        </dl>
      </Container>
    </div>
  );
}
