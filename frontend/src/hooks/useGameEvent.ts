import { useCallback, useRef, useState } from "react";

export interface GameEvent {
  id: number;
  title: string;
  achievementName?: string;
  playerName?: string;
  playerColor?: string;
  duration: number;
  size?: "normal" | "large";
}

type GameEventInput = Omit<GameEvent, "id">;

export function useGameEvent() {
  const queueRef = useRef<GameEvent[]>([]);
  const nextIdRef = useRef(0);
  const [currentEvent, setCurrentEvent] = useState<GameEvent | null>(null);

  const showNext = useCallback(() => {
    queueRef.current.shift();
    setCurrentEvent(queueRef.current[0] ?? null);
  }, []);

  const enqueue = useCallback((input: GameEventInput) => {
    const event: GameEvent = { ...input, id: nextIdRef.current++ };
    queueRef.current.push(event);
    setCurrentEvent((current) => current ?? queueRef.current[0] ?? null);
  }, []);

  const dismissCurrent = useCallback(() => {
    showNext();
  }, [showNext]);

  return { currentEvent, enqueue, dismissCurrent };
}
