import type { PokemonCard } from "@/lib/types";
import { CardTile } from "./CardTile";

interface CatalogGridProps {
  cards: PokemonCard[];
}

export function CatalogGrid({ cards }: CatalogGridProps) {
  return (
    <div className="grid grid-cols-2 gap-3.5 sm:grid-cols-[repeat(auto-fill,minmax(9.5rem,1fr))]">
      {cards.map((card) => (
        <CardTile key={card.id} card={card} />
      ))}
    </div>
  );
}
