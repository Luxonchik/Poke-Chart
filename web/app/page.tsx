import { InfiniteCardGrid } from "@/components/cards/InfiniteCardGrid";
import { CardDetailModal } from "@/components/cards/CardDetailModal";
import { Trending } from "@/components/features/landing/Trending";

export default function Home() {
  return (
    <main className="max-w-7xl mx-auto px-7 pt-8 pb-20">
      <Trending />
      <InfiniteCardGrid />
      <CardDetailModal />
    </main>
  );
}
