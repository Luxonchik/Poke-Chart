"use client";

import { useTranslations, useFormatter } from "next-intl";
import { Badge } from "@/components/ui/badge";
import type { MarketListing } from "@/lib/types";

interface MarketplaceListingsProps {
  listings: MarketListing[];
}

export function MarketplaceListings({ listings }: MarketplaceListingsProps) {
  const t = useTranslations("cardDetail");
  const format = useFormatter();

  const lowestPrice = Math.min(...listings.map((l) => l.price));

  return (
    <div className="space-y-2">
      {listings.map((listing) => (
        <a
          key={listing.id}
          href={listing.source_url}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-3 rounded-xl bg-white/5 px-3 py-2.5 hover:bg-white/10 transition-colors"
        >
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white bg-[#FF8A00]">
            당
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-white truncate">
                {listing.title}
              </span>
              {listing.price === lowestPrice && (
                <Badge className="bg-green-500/15 text-green-400 border-0 text-[10px] px-1.5 py-0">
                  {t("lowest")}
                </Badge>
              )}
            </div>
            <span className="text-xs text-muted-foreground">
              당근마켓
            </span>
          </div>

          <span className="text-sm font-semibold text-white whitespace-nowrap">
            {format.number(listing.price, {
              style: "currency",
              currency: "KRW",
              maximumFractionDigits: 0,
            })}
          </span>
        </a>
      ))}
    </div>
  );
}
