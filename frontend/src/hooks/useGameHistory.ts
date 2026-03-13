import { useState, useEffect, useRef } from "react";
import type { GameHistoryEntryDto } from "../types/generated/api-types";
import { apiService } from "../services/apiService";

const HISTORY_PHASES = [
  "waiting_for_game_start",
  "init_apply_corp",
  "init_apply_prelude",
  "action",
];
const HISTORY_POLICY = "everyAction";

interface UseGameHistoryReturn {
  entries: GameHistoryEntryDto[];
  isLoading: boolean;
  error: string | null;
}

export function useGameHistory(gameId: string | null): UseGameHistoryReturn {
  const [entries, setEntries] = useState<GameHistoryEntryDto[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fetchedRef = useRef<string | null>(null);

  useEffect(() => {
    if (!gameId || fetchedRef.current === gameId) return;
    fetchedRef.current = gameId;

    setIsLoading(true);
    setError(null);

    apiService
      .getGameHistory(gameId, { phases: HISTORY_PHASES, policy: HISTORY_POLICY })
      .then((data) => {
        setEntries(data);
        setIsLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setIsLoading(false);
      });
  }, [gameId]);

  return { entries, isLoading, error };
}
