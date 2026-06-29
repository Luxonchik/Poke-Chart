import { getTranslations } from "next-intl/server";

export async function Trending() {
  const t = await getTranslations("landing");

  return (
    <div className="mb-8 text-amber-50">
      <h1 className="text-3xl font-bold tracking-tight">{t("title")}</h1>
      <p className="text-muted-foreground mt-2">{t("subtitle")}</p>
    </div>
  );
}
