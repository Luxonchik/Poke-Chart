"use client";

import { useState } from "react";
import Image from "next/image";
import { useTranslations, useFormatter } from "next-intl";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useModalStore } from "@/lib/store";
import { usePriceHistory } from "@/lib/queries/usePriceHistory";
import { useMarketplaceListings } from "@/lib/queries/useMarketplaceListings";
import { PriceHistoryChart } from "./PriceHistoryChart";
import { MarketplaceListings } from "./MarketplaceListings";
import type { PriceHistoryPeriod } from "@/lib/types";

const PERIODS: PriceHistoryPeriod[] = ["7D", "30D", "90D"];

const PERIOD_LABEL_KEYS: Record<PriceHistoryPeriod, string> = {
  "7D": "period7D",
  "30D": "period30D",
  "90D": "period90D",
};

export function CardDetailModal() {
  const t = useTranslations("cardDetail");
  const format = useFormatter();
  const { activeModal, selectedCard, closeModal } = useModalStore();
  const [period, setPeriod] = useState<PriceHistoryPeriod>("30D");

  const isOpen = activeModal === "card-detail" && selectedCard !== null;
  const cardId = selectedCard?.id ?? "";

  const { data: priceHistory } = usePriceHistory(cardId, period);
  const { data: listings } = useMarketplaceListings(cardId);

  if (!selectedCard) return null;

  const lowestListingPrice = listings && listings.length > 0
    ? Math.min(...listings.map((l) => l.price))
    : null;

  const currentPrice = priceHistory?.[priceHistory.length - 1]?.price;
  const oldPrice = priceHistory?.[0]?.price;
  const priceChange =
    currentPrice && oldPrice ? ((currentPrice - oldPrice) / oldPrice) * 100 : null;

  return (
    <Dialog open={isOpen} onOpenChange={(open) => { if (!open) closeModal(); }}>
      <DialogContent
        className="sm:max-w-3xl bg-[#14141b] border-white/10 text-white p-0 gap-0 overflow-hidden"
        showCloseButton
      >
        <DialogTitle className="sr-only">{selectedCard.name}</DialogTitle>

        <div className="grid gap-0 md:grid-cols-[1fr_1.5fr]">
          {/* Left — card preview */}
          <div className="flex flex-col gap-4 bg-[#0c0c11] p-5">
            <div className="relative aspect-[20/27] overflow-hidden rounded-xl">
              <Image
                src={selectedCard.image_url}
                alt={selectedCard.name}
                fill
                className="object-contain"
                sizes="(max-width: 768px) 80vw, 30vw"
              />
            </div>
            <Button
              variant="outline"
              className="w-full border-white/10 text-white hover:bg-white/5"
            >
              {t("viewFullDetails")}
            </Button>
          </div>

          {/* Right — details */}
          <div className="flex flex-col gap-5 p-5 overflow-y-auto max-h-[80vh]">
            {/* Name + series + rarity + card_no */}
            <div>
              <h2 className="text-xl font-bold">{selectedCard.name}</h2>
              <div className="mt-1 flex items-center gap-2">
                <span className="text-sm text-muted-foreground">{selectedCard.series}</span>
                <Badge
                  variant="secondary"
                  className="bg-white/10 text-white/70 border-0 text-xs"
                >
                  {selectedCard.rarity}
                </Badge>
              </div>
              {selectedCard.card_no && (
                <span className="text-xs text-muted-foreground mt-1 block">
                  #{selectedCard.card_no}
                </span>
              )}
            </div>

            {/* Price block — from real listings */}
            {lowestListingPrice != null && (
              <div className="flex items-baseline gap-3">
                <span className="text-2xl font-bold">
                  {format.number(lowestListingPrice, {
                    style: "currency",
                    currency: "KRW",
                    maximumFractionDigits: 0,
                  })}
                </span>
                <span className="text-xs text-muted-foreground">
                  {t("lowestMarketPrice")}
                </span>
              </div>
            )}

            {/* Price history */}
            <div>
              <h3 className="text-sm font-medium mb-3">{t("priceHistory")}</h3>
              <div className="flex gap-1 mb-3">
                {PERIODS.map((p) => (
                  <button
                    key={p}
                    onClick={() => setPeriod(p)}
                    className={cn(
                      "px-3 py-1 text-xs font-medium rounded-lg transition-colors",
                      period === p
                        ? "bg-white/15 text-white"
                        : "text-muted-foreground hover:text-white hover:bg-white/5"
                    )}
                  >
                    {t(PERIOD_LABEL_KEYS[p])}
                  </button>
                ))}
              </div>
              {priceHistory && <PriceHistoryChart data={priceHistory} />}
            </div>

            {/* Marketplace listings */}
            {listings && listings.length > 0 && (
              <div>
                <h3 className="text-sm font-medium mb-3">
                  {t("whereItsSelling")} ({listings.length})
                </h3>
                <MarketplaceListings listings={listings} />
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
