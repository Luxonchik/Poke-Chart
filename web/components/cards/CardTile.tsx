"use client";

import Image from "next/image";
import { useFormatter, useTranslations } from "next-intl";
import type { PokemonCard } from "@/lib/types";
import { useModalStore } from "@/lib/store";
import { cn } from "@/lib/utils";

interface CardTileProps {
  card: PokemonCard;
}

export function CardTile({ card }: CardTileProps) {
  const t = useTranslations("cards");
  const format = useFormatter();
  const openCardDetail = useModalStore((s) => s.openCardDetail);
  const onSale = card.min_price != null;

  return (
    <button
      type="button"
      onClick={() => openCardDetail(card)}
      className={cn(
        "group flex flex-col overflow-hidden rounded-md border border-border bg-card text-left transition-colors",
        "hover:border-primary/60 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
        !onSale && "opacity-75 hover:opacity-100",
      )}
    >
      <div className="relative aspect-[63/88] w-full bg-muted">
        <Image
          src={card.image_url}
          alt={card.name}
          fill
          sizes="(max-width: 640px) 50vw, 180px"
          className="object-cover"
        />
      </div>
      <div className="flex w-full flex-col gap-0.5 px-2 py-1.5">
        <span className="truncate text-[13px] font-semibold text-foreground">{card.name}</span>
        <div className="flex items-center justify-between gap-2 font-mono text-xs tabular-nums">
          <span className={onSale ? "text-foreground" : "text-muted-foreground"}>
            {onSale
              ? format.number(card.min_price as number, { style: "currency", currency: "KRW", maximumFractionDigits: 0 })
              : t("noData")}
          </span>
          <span className="truncate text-muted-foreground">{card.card_no}</span>
        </div>
      </div>
    </button>
  );
}
