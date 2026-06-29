"use client";

import Image from "next/image";
import type { PokemonCard as PokemonCardType } from "@/lib/types";
import { useModalStore } from "@/lib/store";
import { useCardTilt } from "./useCardTilt";

interface PokemonCardProps {
  card: PokemonCardType;
}

export function PokemonCard({ card }: PokemonCardProps) {
  const { cardRef, foilRef, glareRef, onMouseMove, onMouseLeave } = useCardTilt();
  const openCardDetail = useModalStore((s) => s.openCardDetail);

  return (
    <div className="perspective-[800px]">
      <div
        ref={cardRef}
        onMouseMove={onMouseMove}
        onMouseLeave={onMouseLeave}
        onClick={() => openCardDetail(card)}
        className="cursor-pointer overflow-hidden rounded-2xl bg-[#14141b] p-3.5 transition-transform duration-150 ease-out"
        style={{ transformStyle: "preserve-3d" }}
      >
        <div
          className="relative isolate overflow-hidden rounded-xl aspect-[20/27]"
          style={{ transformStyle: "flat" }}
        >
          <Image
            src={card.image_url}
            alt={card.name}
            fill
            className="object-contain"
            sizes="(max-width: 768px) 50vw, (max-width: 1200px) 25vw, 20vw"
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

        <div className="mt-2.5 px-0.5 pb-0.5">
          <h3 className="text-sm font-semibold leading-tight truncate text-white">
            {card.name}
          </h3>
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground truncate">
              {card.series}
            </p>
            {card.listing_count > 0 && (
              <span className="text-[10px] font-medium text-green-400 bg-green-500/15 px-1.5 py-0.5 rounded-md shrink-0 ml-1">
                {card.listing_count}
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
