import { useState, useCallback, useRef, useEffect } from "react";
import type { GameHistoryEntryDto } from "../types/generated/api-types";

interface UseGameReplayReturn {
  currentEntry: GameHistoryEntryDto | null;
  currentIndex: number;
  totalStates: number;
  isPlaying: boolean;
  playbackSpeed: number;
  isActive: boolean;
  play: () => void;
  pause: () => void;
  seekTo: (index: number) => void;
  stepForward: () => void;
  stepBackward: () => void;
  setPlaybackSpeed: (speed: number) => void;
  start: () => void;
  exit: () => void;
}

export function useGameReplay(entries: GameHistoryEntryDto[]): UseGameReplayReturn {
  const [isActive, setIsActive] = useState(false);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearTimer = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // Auto-advance when playing
  useEffect(() => {
    if (!isPlaying || !isActive || entries.length === 0) return;

    const delay = 1000 / playbackSpeed;
    timerRef.current = setTimeout(() => {
      setCurrentIndex((prev) => {
        if (prev >= entries.length - 1) {
          setIsPlaying(false);
          return prev;
        }
        return prev + 1;
      });
    }, delay);

    return clearTimer;
  }, [isPlaying, currentIndex, playbackSpeed, entries.length, isActive, clearTimer]);

  const play = useCallback(() => setIsPlaying(true), []);
  const pause = useCallback(() => {
    setIsPlaying(false);
    clearTimer();
  }, [clearTimer]);

  const seekTo = useCallback(
    (index: number) => {
      setCurrentIndex(Math.max(0, Math.min(index, entries.length - 1)));
    },
    [entries.length],
  );

  const stepForward = useCallback(() => {
    pause();
    setCurrentIndex((prev) => Math.min(prev + 1, entries.length - 1));
  }, [entries.length, pause]);

  const stepBackward = useCallback(() => {
    pause();
    setCurrentIndex((prev) => Math.max(prev - 1, 0));
  }, [pause]);

  const start = useCallback(() => {
    setIsActive(true);
    setCurrentIndex(0);
    setIsPlaying(false);
  }, []);

  const exit = useCallback(() => {
    setIsActive(false);
    setIsPlaying(false);
    setCurrentIndex(0);
    clearTimer();
  }, [clearTimer]);

  const currentEntry = isActive && entries.length > 0 ? (entries[currentIndex] ?? null) : null;

  return {
    currentEntry,
    currentIndex,
    totalStates: entries.length,
    isPlaying,
    playbackSpeed,
    isActive,
    play,
    pause,
    seekTo,
    stepForward,
    stepBackward,
    setPlaybackSpeed,
    start,
    exit,
  };
}
