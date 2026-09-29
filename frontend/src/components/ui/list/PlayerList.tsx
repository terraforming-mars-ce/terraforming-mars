import { useState, useCallback, useImperativeHandle, forwardRef } from "react";
import {
  PlayerDto,
  OtherPlayerDto,
  GamePhase,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { isPlayerActionPhase } from "@/utils/actionUtils.ts";
import PlayerCard from "../cards/PlayerCard.tsx";
import GameButton from "../buttons/GameButton.tsx";
import { GameFlowPopover, GameFlowTitle, GameFlowFooter } from "../popover/GameFlowPopover.tsx";

export interface PlayerListHandle {
  requestSkipAction: () => void;
}

interface PlayerListProps {
  players: (PlayerDto | OtherPlayerDto)[];
  currentPlayer: PlayerDto | null;
  turnPlayerId: string;
  currentPhase?: GamePhase;
  hostPlayerId?: string;
  triggeredEffects?: TriggeredEffectDto[];
  onPlayerClick?: (player: PlayerDto | OtherPlayerDto) => void;
  onKickPlayer?: (playerId: string) => void;
  onConvertToBot?: (playerId: string) => void;
}

const PlayerList = forwardRef<PlayerListHandle, PlayerListProps>(function PlayerList(
  {
    players,
    currentPlayer,
    turnPlayerId,
    currentPhase,
    hostPlayerId,
    triggeredEffects = [],
    onPlayerClick,
    onKickPlayer,
    onConvertToBot,
  },
  ref,
) {
  const isActionPhase = isPlayerActionPhase(currentPhase);

  const [showPassConfirmation, setShowPassConfirmation] = useState(false);
  const [passWarningReason, setPassWarningReason] = useState("");

  const dismissPassConfirmation = useCallback(() => {
    setShowPassConfirmation(false);
  }, []);

  const doSkipAction = useCallback(async () => {
    try {
      await globalWebSocketManager.skipAction();
    } catch (error) {
      console.error("Failed to skip action:", error);
    }
  }, []);

  const handleSkipAction = useCallback(async () => {
    if (!currentPlayer) {
      await doSkipAction();
      return;
    }

    const isPassing = currentPlayer.availableActions === 2 || currentPlayer.availableActions === -1;
    if (!isPassing) {
      await doSkipAction();
      return;
    }

    const hasPlayableCards = currentPlayer.cards.some((c) => c.available);
    const hasUsableActions = currentPlayer.actions.some((a) => a.available);

    if (!hasPlayableCards && !hasUsableActions) {
      await doSkipAction();
      return;
    }

    let reason = "You have unused card actions this generation.";
    if (hasPlayableCards && hasUsableActions) {
      reason = "You have playable cards and unused card actions.";
    } else if (hasPlayableCards) {
      reason = "You have cards you can still play.";
    }
    setPassWarningReason(reason);
    setShowPassConfirmation(true);
  }, [currentPlayer, doSkipAction]);

  const handleConfirmPass = useCallback(async () => {
    setShowPassConfirmation(false);
    await doSkipAction();
  }, [doSkipAction]);

  useImperativeHandle(
    ref,
    () => ({
      requestSkipAction: () => void handleSkipAction(),
    }),
    [handleSkipAction],
  );

  return (
    <div className="flex flex-col gap-0 overflow-y-auto overflow-x-visible max-h-[calc(100vh-200px)] [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden">
      {players.map((player) => (
        <PlayerCard
          key={player.id}
          player={player}
          playerColor={player.color || "#6496ff"}
          isCurrentPlayer={player.id === currentPlayer?.id}
          isCurrentTurn={player.id === turnPlayerId}
          isActionPhase={isActionPhase}
          isHost={currentPlayer?.id === hostPlayerId}
          onSkipAction={handleSkipAction}
          triggeredEffects={triggeredEffects}
          onPlayerClick={onPlayerClick}
          onKickPlayer={onKickPlayer}
          onConvertToBot={onConvertToBot}
        />
      ))}
      <GameFlowPopover
        isVisible={showPassConfirmation}
        onClose={dismissPassConfirmation}
        type="immediate"
      >
        <GameFlowTitle>
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            Actions Available
          </h3>
          <div className="text-white/60 text-xs text-shadow-glow mt-1">{passWarningReason}</div>
        </GameFlowTitle>
        <GameFlowFooter className="gap-3">
          <GameButton emphasis="secondary" size="sm" onClick={dismissPassConfirmation}>
            Cancel
          </GameButton>
          <GameButton tone="warn" size="sm" onClick={() => void handleConfirmPass()}>
            Pass Anyway
          </GameButton>
        </GameFlowFooter>
      </GameFlowPopover>
    </div>
  );
});

export default PlayerList;
