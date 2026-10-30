"use client";

import Image from "next/image";
import { useFormatter, useTranslations } from "next-intl";
import type { PokemonCard } from "@/lib/types";
import { useModalStore } from "@/lib/store";
import { cn } from "@/lib/utils";

interface CatalogTableProps {
  cards: PokemonCard[];
}

export function CatalogTable({ cards }: CatalogTableProps) {
  const t = useTranslations("cards");
  const format = useFormatter();
  const openCardDetail = useModalStore((s) => s.openCardDetail);

  return (
    <div className="overflow-x-auto rounded-xs border border-border">
      <table className="w-full min-w-[34rem] text-sm">
        <thead className="bg-card text-left text-[11px] uppercase tracking-wider text-muted-foreground">
          <tr>
            <th scope="col" className="w-16 px-4 py-2.5">
              <span className="sr-only">{t("columns.image")}</span>
            </th>
            <th scope="col" className="px-4 py-2.5 font-medium">{t("columns.card")}</th>
            <th scope="col" className="w-32 px-4 py-2.5 font-medium">{t("columns.number")}</th>
            <th scope="col" className="w-40 px-4 py-2.5 text-right font-medium">{t("columns.lowestPrice")}</th>
            <th scope="col" className="w-36 px-4 py-2.5 text-right font-medium">{t("columns.listings")}</th>
          </tr>
        </thead>
        <tbody>
          {cards.map((card) => {
            const onSale = card.min_price != null;
            return (
              <tr
                key={card.id}
                onClick={() => openCardDetail(card)}
                className={cn(
                  "cursor-pointer border-t border-border transition-colors hover:bg-accent",
                  !onSale && "text-muted-foreground",
                )}
              >
                <td className="px-4 py-2">
                  <div className="relative h-14 w-10 overflow-hidden rounded-xs bg-muted">
                    <Image src={card.image_url} alt="" fill sizes="40px" className="object-cover" />
                  </div>
                </td>
                <td className="max-w-0 px-4 py-2">
                  {/* Keyboard access: Enter on the button bubbles to the row's click handler. */}
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      className={cn(
                        "truncate rounded-sm text-left text-[15px] font-semibold hover:text-primary",
                        "focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
                        onSale && "text-foreground",
                      )}
                    >
                      {card.name}
                    </button>
                    {card.variant_count > 1 && (
                      <span className="shrink-0 rounded border border-border px-1.5 py-px text-[10px] text-muted-foreground">
                        {t("variants", { count: card.variant_count })}
                      </span>
                    )}
                  </div>
                  <p className="mt-0.5 truncate text-xs text-muted-foreground">{card.series}</p>
                </td>
                <td className="px-4 py-2 font-mono text-xs text-muted-foreground">{card.card_no}</td>
                <td className="px-4 py-2 text-right font-mono text-[15px] font-medium tabular-nums">
                  {onSale
                    ? format.number(card.min_price as number, { style: "currency", currency: "KRW", maximumFractionDigits: 0 })
                    : t("noData")}
                </td>
                <td className="px-4 py-2 text-right">
                  {onSale ? (
                    <span className="rounded bg-primary/15 px-2 py-0.5 text-[11px] font-medium text-primary">
                      {t("listingCount", { count: card.listing_count })}
                    </span>
                  ) : (
                    <span className="text-[11px]">{t("notOnSale")}</span>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
