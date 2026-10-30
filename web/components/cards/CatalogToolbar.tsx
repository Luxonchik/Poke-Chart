"use client";

import { useEffect, useState } from "react";
import { LayoutGrid, Rows3, Search } from "lucide-react";
import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { useCatalogFilterStore } from "@/lib/store";
import type { CatalogView } from "@/lib/types";
import { cn } from "@/lib/utils";

const SEARCH_DEBOUNCE_MS = 300;

interface CatalogToolbarProps {
  view: CatalogView;
  onViewChange: (view: CatalogView) => void;
}

export function CatalogToolbar({ view, onViewChange }: CatalogToolbarProps) {
  const t = useTranslations("cards");
  const { query, onSale, setQuery, toggleOnSale } = useCatalogFilterStore();
  const [value, setValue] = useState(query);

  useEffect(() => {
    const id = setTimeout(() => setQuery(value.trim()), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(id);
  }, [value, setQuery]);

  const views: { id: CatalogView; label: string; Icon: typeof LayoutGrid }[] = [
    { id: "grid", label: t("view.grid"), Icon: LayoutGrid },
    { id: "table", label: t("view.table"), Icon: Rows3 },
  ];

  return (
    <div className="flex flex-wrap items-center gap-2">
      <form role="search" onSubmit={(e) => e.preventDefault()} className="relative w-full sm:w-80">
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

      <button
        type="button"
        aria-pressed={onSale}
        onClick={toggleOnSale}
        className={cn(
          "h-9 rounded-full border px-4 text-sm transition-colors",
          onSale
            ? "border-primary bg-primary/15 text-primary"
            : "border-border text-muted-foreground hover:bg-accent hover:text-foreground",
        )}
      >
        {t("onSale")}
      </button>

      <div role="group" aria-label={t("view.label")} className="ml-auto flex overflow-hidden rounded-md border border-border">
        {views.map(({ id, label, Icon }) => (
          <button
            key={id}
            type="button"
            aria-pressed={view === id}
            onClick={() => onViewChange(id)}
            className={cn(
              "flex h-9 items-center gap-1.5 px-3 text-sm transition-colors not-first:border-l not-first:border-border",
              view === id ? "bg-accent text-foreground" : "text-muted-foreground hover:text-foreground",
            )}
          >
            <Icon className="h-4 w-4" aria-hidden />
            {label}
          </button>
        ))}
      </div>
    </div>
  );
}
