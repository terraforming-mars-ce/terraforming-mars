import { create } from "zustand";

interface CardDragState {
  isDraggingCard: boolean;

  startCardDrag: () => void;
  endCardDrag: () => void;
}

export const useCardDragStore = create<CardDragState>((set) => ({
  isDraggingCard: false,

  startCardDrag: () => set({ isDraggingCard: true }),
  endCardDrag: () => set({ isDraggingCard: false }),
}));
