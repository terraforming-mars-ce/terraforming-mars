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
  hasPendingTilePlacement?: boolean;
  triggeredEffects?: TriggeredEffectDto[];
  onPlayerClick?: (player: PlayerDto | OtherPlayerDto) => void;
}

const LeftSidebar: React.FC<LeftSidebarProps> = ({
  players,
  currentPlayer,
  turnPlayerId,
  currentPhase,
  hasPendingTilePlacement = false,
  triggeredEffects = [],
  onPlayerClick,
}) => {
  return (
    <div className="absolute top-[15%] left-0 z-10 w-[240px] h-[calc(85vh-120px)] bg-transparent py-[15px] flex flex-col overflow-visible pointer-events-none">
      <PlayerList
        players={players}
        currentPlayer={currentPlayer}
        turnPlayerId={turnPlayerId}
        currentPhase={currentPhase}
        hasPendingTilePlacement={hasPendingTilePlacement}
        triggeredEffects={triggeredEffects}
        onPlayerClick={onPlayerClick}
      />
    </div>
  );
};

export default LeftSidebar;
