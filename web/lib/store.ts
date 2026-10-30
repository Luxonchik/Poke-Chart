import { create } from "zustand";
import type { PokemonCard } from "./types";

interface ModalState {
  activeModal: "card-detail" | null;
  selectedCard: PokemonCard | null;
  openCardDetail: (card: PokemonCard) => void;
  closeModal: () => void;
}

interface CatalogFilterState {
  query: string;
  onSale: boolean;
  setQuery: (query: string) => void;
  toggleOnSale: () => void;
}

export const useCatalogFilterStore = create<CatalogFilterState>((set) => ({
  query: "",
  onSale: false,
  setQuery: (query) => set({ query }),
  toggleOnSale: () => set((s) => ({ onSale: !s.onSale })),
}));

export const useModalStore = create<ModalState>((set) => ({
  activeModal: null,
  selectedCard: null,
  openCardDetail: (card) => set({ activeModal: "card-detail", selectedCard: card }),
  closeModal: () => set({ activeModal: null, selectedCard: null }),
}));
