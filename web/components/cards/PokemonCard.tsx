"use client";

import Image from "next/image";
import { useFormatter, useTranslations } from "next-intl";
import type { PokemonCard as PokemonCardType } from "@/lib/types";
import { useModalStore } from "@/lib/store";
import { useCardTilt } from "./useCardTilt";

interface PokemonCardProps {
  card: PokemonCardType;
}

export function PokemonCard({ card }: PokemonCardProps) {
  const { cardRef, foilRef, glareRef, onMouseMove, onMouseLeave } =
    useCardTilt();
  const openCardDetail = useModalStore((s) => s.openCardDetail);
  const t = useTranslations("cards");
  const format = useFormatter();

  return (
    <div className="perspective-[800px]">
      <div
        ref={cardRef}
        onMouseMove={onMouseMove}
        onMouseLeave={onMouseLeave}
        onClick={() => openCardDetail(card)}
        className="cursor-pointer overflow-hidden rounded-md border border-border bg-card transition-transform duration-150 ease-out"
        style={{ transformStyle: "preserve-3d" }}
      >
        <div
          className="relative isolate aspect-[63/88] overflow-hidden"
          style={{ transformStyle: "flat" }}
        >
          <Image
            src={card.image_url}
            alt={card.name}
            fill
            className="object-cover"
            sizes="240px"
          />

          <div
            ref={foilRef}
            className="pointer-events-none absolute inset-0 z-20 opacity-0 transition-opacity duration-300"
          />

          <div
            ref={glareRef}
            className="pointer-events-none absolute inset-0 z-20 opacity-0 transition-opacity duration-250 mix-blend-soft-light"
          />
        </div>

        <div className="flex flex-col gap-1 px-2.5 py-2">
          <h3 className="truncate text-sm font-semibold text-foreground">
            {card.name}
          </h3>
          {card.min_price != null && (
            <div className="flex items-center justify-between gap-2">
              <span className="font-mono text-sm font-medium tabular-nums text-foreground">
                {format.number(card.min_price, {
                  style: "currency",
                  currency: "KRW",
                  maximumFractionDigits: 0,
                })}
              </span>
              <span className="shrink-0 rounded bg-primary/15 px-1.5 py-0.5 text-[10px] font-medium text-primary">
                {t("listingCount", { count: card.listing_count })}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
