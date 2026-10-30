import { getTranslations } from "next-intl/server";
import { OnMarketCarousel } from "./OnMarketCarousel";

export async function Trending() {
  const t = await getTranslations("landing");

  return (
    <section id="market" className="scroll-mt-24">
      <div className="mb-8 text-foreground">
        <h1 className="text-3xl font-bold tracking-tight">{t("title")}</h1>
        <p className="text-muted-foreground mt-2">{t("subtitle")}</p>
      </div>
      <OnMarketCarousel />
    </section>
  );
}
