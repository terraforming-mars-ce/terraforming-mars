import { useCallback, useState } from "react";

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

export function useCardInspection() {
  const [inspections, setInspections] = useState<CardInspectionSession[]>([]);

  const clearInspection = useCallback(
    () => setInspections((current) => (current.length ? [] : current)),
    [],
  );

  const closeInspection = useCallback((restoreFocus = false) => {
    setInspections((current) =>
      current.map((entry) => (entry.closing ? entry : { ...entry, closing: true, restoreFocus })),
    );
  }, []);

  const inspectCard = useCallback((cardId: string, source: HTMLElement) => {
    setInspections((current) => {
      const existing = current.find((entry) => entry.cardId === cardId);
      const returning = current.map((entry) => ({ ...entry, closing: true, restoreFocus: false }));
      if (existing && !existing.closing) {
        return returning;
      }
      return [
        ...returning.filter((entry) => entry.cardId !== cardId),
        { cardId, source, closing: false, restoreFocus: false },
      ];
    });
  }, []);

  const finishInspection = useCallback((finished: CardInspectionSession) => {
    setInspections((current) => current.filter((entry) => entry !== finished));
    if (finished.restoreFocus && finished.source.isConnected) {
      requestAnimationFrame(() => finished.source.focus({ preventScroll: true }));
    }
  }, []);

  return { inspections, inspectCard, closeInspection, finishInspection, clearInspection };
}
