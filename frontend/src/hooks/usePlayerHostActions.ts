import { useMemo } from "react";
import { OtherPlayerDto, PlayerDto } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";

export interface PlayerHostAction {
  id: "convert-to-bot" | "retry-bot" | "kick";
  label: string;
  tone: "danger" | "neutral";
  run: () => void;
}

interface PlayerHostActionsInput {
  player: PlayerDto | OtherPlayerDto;
  isHost: boolean;
  isCurrentPlayer: boolean;
  onKickPlayer?: (playerId: string) => void;
  onConvertToBot?: (playerId: string) => void;
}

export function usePlayerHostActions({
  player,
  isHost,
  isCurrentPlayer,
  onKickPlayer,
  onConvertToBot,
}: PlayerHostActionsInput): PlayerHostAction[] {
  const playerId = player.id;
  const isExited = player.isExited;
  const isBot = player.playerType === "bot";
  const botFailed = player.botStatus === "failed";

  return useMemo(() => {
    const actions: PlayerHostAction[] = [];
    if (!isHost) {
      return actions;
    }
    if (!isCurrentPlayer && !isExited && !isBot && onConvertToBot) {
      actions.push({
        id: "convert-to-bot",
        label: "Convert to bot",
        tone: "danger",
        run: () => onConvertToBot(playerId),
      });
    }
    if (isBot && botFailed) {
      actions.push({
        id: "retry-bot",
        label: "Retry bot",
        tone: "neutral",
        run: () => void globalWebSocketManager.retryBot(playerId),
      });
    }
    if (!isCurrentPlayer && !isExited && onKickPlayer) {
      actions.push({
        id: "kick",
        label: "Kick player",
        tone: "danger",
        run: () => onKickPlayer(playerId),
      });
    }
    return actions;
  }, [isHost, isCurrentPlayer, isExited, isBot, botFailed, onConvertToBot, onKickPlayer, playerId]);
}
