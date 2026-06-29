import { Search } from "lucide-react";
import { Input } from "../ui/input";

export function Header() {
  return (
    <header className="sticky top-0 z-50 w-full bg-[#0c0c11]/72 backdrop-blur-xl border-b border-white/10">
      <div className="container mx-auto flex h-16 items-center justify-between px-4">
        {/* Логотип */}
        <div className="flex items-center gap-2 text-white ">
          <span className="text-xl font-bold tracking-tight ">
            Poké<span className="text-red-600">Track </span>
          </span>
        </div>

        {/* Центральная часть: Строка поиска */}
        <div className="hidden flex-1 items-center justify-center px-8 md:flex">
          <div className="relative w-full max-w-lg">
            <Search className="absolute left-2.5 top-2 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search card, sets: 레시라무 ex or SVP020..."
              className="w-full rounded-full bg-muted pl-9 pr-4 focus-visible:ring-1"
            />
          </div>
        </div>
      </div>
    </header>
  );
}
