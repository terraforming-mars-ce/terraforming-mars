import { useCallback, useRef, useState } from "react";
import { CardDto } from "@/types/generated/api-types.ts";

export interface PlayedCardNotification {
  id: number;
  card: CardDto;
  playerName: string;
  playerColor: string;
}

type PlayedCardNotificationInput = Omit<PlayedCardNotification, "id">;

export function usePlayedCardNotification() {
  const queueRef = useRef<PlayedCardNotification[]>([]);
  const nextIdRef = useRef(0);
  const [currentNotification, setCurrentNotification] = useState<PlayedCardNotification | null>(
    null,
  );
  const [isPinned, setIsPinned] = useState(false);

  const advance = useCallback(() => {
    queueRef.current.shift();
    setIsPinned(false);
    setCurrentNotification(queueRef.current[0] ?? null);
  }, []);

  const enqueue = useCallback((input: PlayedCardNotificationInput) => {
    const notification: PlayedCardNotification = { ...input, id: nextIdRef.current++ };
    queueRef.current.push(notification);
    setCurrentNotification((current) => current ?? queueRef.current[0] ?? null);
  }, []);

  const togglePin = useCallback(() => {
    setIsPinned((prev) => !prev);
  }, []);

  return { currentNotification, isPinned, enqueue, togglePin, advance };
}
