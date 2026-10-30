import Link from "next/link";
import { getTranslations } from "next-intl/server";
import { Container } from "./Container";

export async function Header() {
  const t = await getTranslations("header");

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border bg-background/80 backdrop-blur-xl">
      <Container className="flex h-16 items-center gap-8">
        <Link href="/" className="text-xl font-bold tracking-tight text-foreground">
          Poké<span className="text-primary">Track</span>
        </Link>

        <nav className="hidden gap-6 text-sm sm:flex">
          <Link href="/#market" className="text-muted-foreground transition-colors hover:text-foreground">
            {t("market")}
          </Link>
          <Link href="/#catalog" className="text-muted-foreground transition-colors hover:text-foreground">
            {t("catalog")}
          </Link>
        </nav>
      </Container>
    </header>
  );
}
