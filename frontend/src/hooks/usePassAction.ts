import { useCallback, useState } from "react";
import {
  PlayerStatusSelection,
  PlayerStatusTile,
  type GamePhase,
  type PlayerDto,
} from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { isPlayerActionPhase } from "@/utils/actionUtils.ts";

interface UsePassActionOptions {
  player: PlayerDto | null;
  turnPlayerId?: string;
  currentPhase?: GamePhase;
}

export interface PassConfirmation {
  open: boolean;
  reason: string;
  confirm: () => void;
  dismiss: () => void;
}

export interface PassAction {
  canPass: boolean;
  label: "PASS" | "SKIP";
  requestPass: () => void;
  confirmation: PassConfirmation;
}

async function skipAction() {
  try {
    await globalWebSocketManager.skipAction();
  } catch (error) {
    console.error("Failed to skip action:", error);
  }
}

function passWarningReason(player: PlayerDto): string | null {
  const isPassing = player.availableActions === 2 || player.availableActions === -1;
  if (!isPassing) {
    return null;
  }
  const hasPlayableCards = player.cards.some((c) => c.available);
  const hasUsableActions = player.actions.some((a) => a.available);
  if (hasPlayableCards && hasUsableActions) {
    return "You have playable cards and unused card actions.";
  }
  if (hasPlayableCards) {
    return "You have cards you can still play.";
  }
  if (hasUsableActions) {
    return "You have unused card actions this generation.";
  }
  return null;
}

export function usePassAction({
  player,
  turnPlayerId,
  currentPhase,
}: UsePassActionOptions): PassAction {
  const [open, setOpen] = useState(false);
  const [reason, setReason] = useState("");

  const isBlocked = player?.status === PlayerStatusTile || player?.status === PlayerStatusSelection;
  const canPass =
    !!player && player.id === turnPlayerId && isPlayerActionPhase(currentPhase) && !isBlocked;
  const label = player?.availableActions === -1 || player?.availableActions === 2 ? "PASS" : "SKIP";

  const requestPass = useCallback(() => {
    const warning = player ? passWarningReason(player) : null;
    if (!warning) {
      void skipAction();
      return;
    }
    setReason(warning);
    setOpen(true);
  }, [player]);

  const confirm = useCallback(() => {
    setOpen(false);
    void skipAction();
  }, []);

  const dismiss = useCallback(() => {
    setOpen(false);
  }, []);

  return { canPass, label, requestPass, confirmation: { open, reason, confirm, dismiss } };
}
