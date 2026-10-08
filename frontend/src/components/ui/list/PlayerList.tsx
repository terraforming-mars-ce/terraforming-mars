import { useImperativeHandle, forwardRef } from "react";
import {
  PlayerDto,
  OtherPlayerDto,
  GamePhase,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";
import { usePassAction } from "@/hooks/usePassAction.ts";
import { isPlayerActionPhase } from "@/utils/actionUtils.ts";
import PlayerCard from "../cards/PlayerCard.tsx";
import PassConfirmDialog from "../modals/PassConfirmDialog.tsx";

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

  const { requestPass, confirmation } = usePassAction({
    player: currentPlayer,
    turnPlayerId,
    currentPhase,
  });

  useImperativeHandle(
    ref,
    () => ({
      requestSkipAction: requestPass,
    }),
    [requestPass],
  );

  return (
    <div className="flex flex-col gap-0 overflow-y-auto overflow-x-visible max-h-[calc(100dvh-200px)] [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden">
      {players.map((player) => (
        <PlayerCard
          key={player.id}
          player={player}
          playerColor={player.color || "#6496ff"}
          isCurrentPlayer={player.id === currentPlayer?.id}
          isCurrentTurn={player.id === turnPlayerId}
          isActionPhase={isActionPhase}
          isHost={currentPlayer?.id === hostPlayerId}
          onSkipAction={requestPass}
          triggeredEffects={triggeredEffects}
          onPlayerClick={onPlayerClick}
          onKickPlayer={onKickPlayer}
          onConvertToBot={onConvertToBot}
        />
      ))}
      <PassConfirmDialog confirmation={confirmation} />
    </div>
  );
});

export default PlayerList;
