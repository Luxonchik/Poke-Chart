import { CardDetailModal } from "@/components/cards/CardDetailModal";
import { CatalogSection } from "@/components/features/landing/CatalogSection";
import { StatsBar } from "@/components/features/landing/StatsBar";
import { Trending } from "@/components/features/landing/Trending";
import { Container } from "@/components/layout/Container";

export default function Home() {
  return (
    <main>
      <StatsBar />
      <Container className="pt-10 pb-20">
        <Trending />
        <CatalogSection />
      </Container>
      <CardDetailModal />
    </main>
  );
}
