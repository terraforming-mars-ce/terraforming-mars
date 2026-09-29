import { useState } from "react";
import { createStore } from "zustand/vanilla";
import { coldStartTrace } from "@/services/performanceStore.ts";

export interface CardInspectionDrag {
  pointerId: number;
  clientX: number;
  clientY: number;
  grabX: number;
  grabY: number;
}

export interface CardInspectionSession {
  cardId: string;
  source: HTMLElement;
  closing: boolean;
  restoreFocus: boolean;
}

interface CardInspectionState {
  inspections: CardInspectionSession[];
  clearInspection: () => void;
  closeInspection: (restoreFocus?: boolean) => void;
  inspectCard: (cardId: string, source: HTMLElement) => void;
  finishInspection: (finished: CardInspectionSession) => void;
}

function createInspectionStore() {
  return createStore<CardInspectionState>((set) => ({
    inspections: [],
    clearInspection: () => set((state) => (state.inspections.length ? { inspections: [] } : state)),
    closeInspection: (restoreFocus = false) =>
      set(({ inspections }) => ({
        inspections: inspections.map((entry) =>
          entry.closing ? entry : { ...entry, closing: true, restoreFocus },
        ),
      })),
    inspectCard: (cardId, source) => {
      coldStartTrace.begin("card-inspection", { cardId });
      set(({ inspections }) => {
        const existing = inspections.find((entry) => entry.cardId === cardId);
        const returning = inspections.map((entry) => ({
          ...entry,
          closing: true,
          restoreFocus: false,
        }));
        if (existing && !existing.closing) {
          return { inspections: returning };
        }
        return {
          inspections: [
            ...returning.filter((entry) => entry.cardId !== cardId),
            { cardId, source, closing: false, restoreFocus: false },
          ],
        };
      });
    },
    finishInspection: (finished) => {
      set(({ inspections }) => ({
        inspections: inspections.filter((entry) => entry !== finished),
      }));
      if (finished.restoreFocus && finished.source.isConnected) {
        requestAnimationFrame(() => finished.source.focus({ preventScroll: true }));
      }
    },
  }));
}

export type CardInspectionStore = ReturnType<typeof createInspectionStore>;

export function useCardInspection() {
  return useState(createInspectionStore)[0];
}
