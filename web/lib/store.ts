import { create } from "zustand";
import type { PokemonCard } from "./types";

interface ModalState {
  activeModal: "card-detail" | null;
  selectedCard: PokemonCard | null;
  openCardDetail: (card: PokemonCard) => void;
  closeModal: () => void;
}

export const useModalStore = create<ModalState>((set) => ({
  activeModal: null,
  selectedCard: null,
  openCardDetail: (card) => set({ activeModal: "card-detail", selectedCard: card }),
  closeModal: () => set({ activeModal: null, selectedCard: null }),
}));
