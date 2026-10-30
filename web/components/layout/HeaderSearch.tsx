"use client";

import { useEffect, useState } from "react";
import { Search } from "lucide-react";
import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { useCatalogFilterStore } from "@/lib/store";

const SEARCH_DEBOUNCE_MS = 300;

export function HeaderSearch() {
  const t = useTranslations("header");
  const query = useCatalogFilterStore((s) => s.query);
  const setQuery = useCatalogFilterStore((s) => s.setQuery);
  const [value, setValue] = useState(query);

  useEffect(() => {
    const id = setTimeout(() => setQuery(value.trim()), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(id);
  }, [value, setQuery]);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setQuery(value.trim());
    document.getElementById("catalog")?.scrollIntoView({ behavior: "smooth" });
  }

  return (
    <form role="search" onSubmit={handleSubmit} className="relative ml-auto hidden w-full max-w-sm md:block">
      <Search className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        type="search"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        aria-label={t("searchLabel")}
        placeholder={t("searchPlaceholder")}
        className="h-9 bg-card pl-9"
      />
    </form>
  );
}
