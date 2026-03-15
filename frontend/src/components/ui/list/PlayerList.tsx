import { useMemo, useState, useCallback, useImperativeHandle, forwardRef } from "react";
import {
  PlayerDto,
  OtherPlayerDto,
  GamePhase,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import PlayerCard from "../cards/PlayerCard.tsx";
import GameButton from "../buttons/GameButton.tsx";
import { GameFlowPopover, GameFlowTitle, GameFlowFooter } from "../popover/GameFlowPopover.tsx";

export interface PlayerListHandle {
  requestSkipAction: () => void;
}

function measureTextWidth(text: string, font: string): number {
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");
  if (!ctx) {
    return text.length * 8;
  }
  ctx.font = font;
  return ctx.measureText(text).width;
}

interface PlayerListProps {
  players: (PlayerDto | OtherPlayerDto)[];
  currentPlayer: PlayerDto | null;
  turnPlayerId: string;
  currentPhase?: GamePhase;
  hostPlayerId?: string;
  pendingTilePlayerId?: string;
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
    pendingTilePlayerId,
    triggeredEffects = [],
    onPlayerClick,
    onKickPlayer,
    onConvertToBot,
  },
  ref,
) {
  const isActionPhase = currentPhase === "action";

  const { minNameWidth, minCardWidth } = useMemo(() => {
    if (players.length === 0) {
      return { minNameWidth: 0, minCardWidth: 0 };
    }
    const nameFont = "bold 14px Orbitron";
    const actionsFont = "bold 10px Orbitron";
    const trFont = "bold 14px Orbitron";
    const buttonFont = "bold 9px Orbitron";
    const actionsSuffix = " 00/00";

    let maxNameWidth = 0;
    let maxTrWidth = 0;
    for (const p of players) {
      const nameWidth = measureTextWidth(p.name, nameFont);
      const actionsWidth = measureTextWidth(actionsSuffix, actionsFont);
      const totalNameWidth = nameWidth + actionsWidth + 6;
      if (totalNameWidth > maxNameWidth) {
        maxNameWidth = totalNameWidth;
      }
      const trWidth = measureTextWidth(String(p.terraformRating), trFont);
      if (trWidth > maxTrWidth) {
        maxTrWidth = trWidth;
      }
    }
    const nameCol = Math.ceil(maxNameWidth);
    // TR box: text + px-2.5 (10px*2) + border (1px*2) = trWidth + 22
    const trBoxWidth = Math.ceil(maxTrWidth) + 22;
    // Button: text + px-3 (12px*2) + border (1px*2) = textWidth + 26
    const buttonWidth = Math.ceil(measureTextWidth("PASS", buttonFont)) + 26;
    // Card: border-left(8) + pl(8) + nameCol + gap(12) + trBox + gap(8) + button + right-pad(20) + clip-cut(8)
    const cardWidth = 8 + 8 + nameCol + 12 + trBoxWidth + 8 + buttonWidth + 20 + 8;
    return { minNameWidth: nameCol, minCardWidth: Math.ceil(cardWidth) };
  }, [players]);

  const [showPassConfirmation, setShowPassConfirmation] = useState(false);

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
          hasPendingTile={player.id === pendingTilePlayerId}
          triggeredEffects={triggeredEffects}
          onPlayerClick={onPlayerClick}
          onKickPlayer={onKickPlayer}
          onConvertToBot={onConvertToBot}
          minNameWidth={minNameWidth}
          minCardWidth={minCardWidth}
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
          <div className="text-white/60 text-xs text-shadow-glow mt-1">
            You still have unused actions this generation.
          </div>
        </GameFlowTitle>
        <GameFlowFooter className="gap-3">
          <GameButton buttonType="secondary" size="sm" onClick={dismissPassConfirmation}>
            Cancel
          </GameButton>
          <GameButton variant="warn" size="sm" onClick={() => void handleConfirmPass()}>
            Pass Anyway
          </GameButton>
        </GameFlowFooter>
      </GameFlowPopover>
    </div>
  );
});

export default PlayerList;
