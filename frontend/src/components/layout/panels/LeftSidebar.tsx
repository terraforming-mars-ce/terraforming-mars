import React from "react";
import {
  PlayerDto,
  OtherPlayerDto,
  GamePhase,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";
import PlayerList from "@/components/ui/list/PlayerList.tsx";

interface LeftSidebarProps {
  players: (PlayerDto | OtherPlayerDto)[];
  currentPlayer: PlayerDto | null;
  turnPlayerId: string;
  currentPhase?: GamePhase;
  hostPlayerId?: string;
  hasPendingTilePlacement?: boolean;
  triggeredEffects?: TriggeredEffectDto[];
  onPlayerClick?: (player: PlayerDto | OtherPlayerDto) => void;
  onKickPlayer?: (playerId: string) => void;
  onConvertToBot?: (playerId: string) => void;
}

const LeftSidebar: React.FC<LeftSidebarProps> = ({
  players,
  currentPlayer,
  turnPlayerId,
  currentPhase,
  hostPlayerId,
  hasPendingTilePlacement = false,
  triggeredEffects = [],
  onPlayerClick,
  onKickPlayer,
  onConvertToBot,
}) => {
  return (
    <div className="absolute top-[15%] left-0 z-10 w-[240px] h-[calc(85vh-120px)] bg-transparent py-[15px] flex flex-col overflow-visible pointer-events-none">
      <PlayerList
        players={players}
        currentPlayer={currentPlayer}
        turnPlayerId={turnPlayerId}
        currentPhase={currentPhase}
        hostPlayerId={hostPlayerId}
        hasPendingTilePlacement={hasPendingTilePlacement}
        triggeredEffects={triggeredEffects}
        onPlayerClick={onPlayerClick}
        onKickPlayer={onKickPlayer}
        onConvertToBot={onConvertToBot}
      />
    </div>
  );
};

export default LeftSidebar;
