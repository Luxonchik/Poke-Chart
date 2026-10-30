"use client";

import { useOnMarketCards } from "@/lib/queries/useCards";
import { PokemonCard } from "@/components/cards/PokemonCard";
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";

export function OnMarketCarousel() {
  const { data: cards, status } = useOnMarketCards();

  if (status !== "success" || cards.length === 0) return null;

  return (
    <Carousel opts={{ dragFree: true, align: "start" }} className="mb-12">
      <div className="absolute -top-12 right-0 hidden gap-2 md:flex">
        <CarouselPrevious className="static translate-y-0 border-border bg-card text-foreground hover:bg-accent" />
        <CarouselNext className="static translate-y-0 border-border bg-card text-foreground hover:bg-accent" />
      </div>
      <CarouselContent className="-ml-5 cursor-grab active:cursor-grabbing">
        {cards.map((card) => (
          <CarouselItem key={card.id} className="basis-60 pl-5">
            <PokemonCard card={card} />
          </CarouselItem>
        ))}
      </CarouselContent>
    </Carousel>
  );
}
