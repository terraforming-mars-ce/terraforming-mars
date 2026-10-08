import { useCallback, useEffect, useState } from "react";
import type { LogUpdatePayload, StateDiffDto } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";

export function useGameLogs(gameId: string | undefined): StateDiffDto[] {
  const [logs, setLogs] = useState<StateDiffDto[]>([]);

  const handleLogUpdate = useCallback(({ logs: newLogs }: LogUpdatePayload) => {
    setLogs((prev) => {
      const existingSeqs = new Set(prev.map((l) => l.sequenceNumber));
      const uniqueNewLogs = newLogs.filter((l) => !existingSeqs.has(l.sequenceNumber));
      if (uniqueNewLogs.length === 0) {
        return prev;
      }
      return [...prev, ...uniqueNewLogs];
    });
  }, []);

  useEffect(() => {
    globalWebSocketManager.on("log-update", handleLogUpdate);
    return () => {
      globalWebSocketManager.off("log-update", handleLogUpdate);
    };
  }, [handleLogUpdate]);

  useEffect(() => {
    setLogs([]);
    if (gameId) {
      void globalWebSocketManager.requestLogs();
    }
  }, [gameId]);

  return logs;
}
