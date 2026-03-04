import React from "react";
import {
  PlayerDto,
  OtherPlayerDto,
  GamePhase,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import PlayerCard from "../cards/PlayerCard.tsx";

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

const PlayerList: React.FC<PlayerListProps> = ({
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
}) => {
  const isActionPhase = currentPhase === "action";

  const handleSkipAction = async () => {
    try {
      await globalWebSocketManager.skipAction();
    } catch (error) {
      console.error("Failed to skip action:", error);
    }
  };

  return (
    <div className="flex flex-col w-full gap-0 overflow-y-auto overflow-x-visible max-h-[calc(100vh-200px)] [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden">
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
        />
      ))}
    </div>
  );
};

export default PlayerList;
