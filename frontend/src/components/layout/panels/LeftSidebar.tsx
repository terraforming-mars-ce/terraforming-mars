import { forwardRef } from "react";
import {
  PlayerDto,
  OtherPlayerDto,
  GamePhase,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";
import PlayerList, { PlayerListHandle } from "@/components/ui/list/PlayerList.tsx";

interface LeftSidebarProps {
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

const LeftSidebar = forwardRef<PlayerListHandle, LeftSidebarProps>(function LeftSidebar(
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
  return (
    <div className="absolute top-[15%] left-0 z-10 h-[calc(85vh-120px)] bg-transparent py-[15px] flex flex-col overflow-visible pointer-events-none">
      <PlayerList
        ref={ref}
        players={players}
        currentPlayer={currentPlayer}
        turnPlayerId={turnPlayerId}
        currentPhase={currentPhase}
        hostPlayerId={hostPlayerId}
        triggeredEffects={triggeredEffects}
        onPlayerClick={onPlayerClick}
        onKickPlayer={onKickPlayer}
        onConvertToBot={onConvertToBot}
      />
    </div>
  );
});

export default LeftSidebar;
