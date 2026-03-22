import { create } from "zustand";

export interface AsteroidEvent {
  id: number;
  cardName: string;
  playerName: string;
  playerColor: string;
}

interface AsteroidEventState {
  queue: AsteroidEvent[];
  enqueue: (event: Omit<AsteroidEvent, "id">) => void;
  dequeue: () => void;
}

let nextId = 0;

export const useAsteroidEventStore = create<AsteroidEventState>((set) => ({
  queue: [],
  enqueue: (event) =>
    set((state) => ({
      queue: [...state.queue, { ...event, id: nextId++ }],
    })),
  dequeue: () =>
    set((state) => ({
      queue: state.queue.slice(1),
    })),
}));
