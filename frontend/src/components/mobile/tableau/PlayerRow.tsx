import { forwardRef } from "react";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import { usePlayerHostActions, type PlayerHostAction } from "@/hooks/usePlayerHostActions.ts";
import GameButton from "../../ui/buttons/GameButton.tsx";
import PlayerStatusChips from "./PlayerStatusChips.tsx";
import ResourceProductionGrid from "./ResourceProductionGrid.tsx";
import { type TableauPlayer } from "./tableauPlayers.ts";
import GameIcon from "../../ui/display/GameIcon.tsx";
import { ResourceTypeTR } from "@/types/generated/api-types.ts";

interface PlayerRowProps {
  player: TableauPlayer;
  color: string;
  isMe: boolean;
  isTurn: boolean;
  isActionPhase: boolean;
  isHost: boolean;
  highlighted: boolean;
  onOpenTableau: (playerId: string) => void;
  onOpenActions: (player: TableauPlayer, actions: PlayerHostAction[]) => void;
  onKickPlayer: (playerId: string) => void;
  onConvertToBot?: (playerId: string) => void;
}

const PlayerRow = forwardRef<HTMLDivElement, PlayerRowProps>(function PlayerRow(
  {
    player,
    color,
    isMe,
    isTurn,
    isActionPhase,
    isHost,
    highlighted,
    onOpenTableau,
    onOpenActions,
    onKickPlayer,
    onConvertToBot,
  },
  ref,
) {
  const hostActions = usePlayerHostActions({
    player,
    isHost,
    isCurrentPlayer: isMe,
    onKickPlayer,
    onConvertToBot,
  });
  const inactive = player.isExited || (player.playerType !== "bot" && !player.isConnected);
  let background = "";
  if (highlighted) {
    background = "bg-white/15";
  } else if (isMe) {
    background = "bg-white/[0.06]";
  }

  return (
    <div
      ref={ref}
      className={`flex items-stretch min-h-14 border-b border-white/10 transition-colors duration-500 ${background}`}
    >
      <span
        className={`shrink-0 ${isTurn ? "w-2" : "w-1.5"}`}
        style={{ backgroundColor: color, boxShadow: isTurn ? `0 0 10px ${color}` : undefined }}
        aria-hidden="true"
      />
      <button
        type="button"
        aria-label={`Open ${isMe ? "your" : `${player.name}'s`} tableau`}
        className={`flex-1 min-w-0 flex items-center gap-3 px-3 py-1.5 text-left cursor-pointer ${inactive ? "opacity-50" : ""}`}
        onClick={() => onOpenTableau(player.id)}
      >
        <span className="flex-1 min-w-0 flex flex-col gap-1">
          <span className="flex flex-wrap items-center gap-x-2 gap-y-1 min-w-0">
            <span
              className={`truncate max-w-full font-orbitron text-[14px] ${isMe ? "font-bold text-white" : "text-white/90"}`}
            >
              {player.name}
            </span>
            <PlayerStatusChips
              player={player}
              isMe={isMe}
              isTurn={isTurn}
              isActionPhase={isActionPhase}
            />
          </span>
          <span className="flex items-center gap-3 min-w-0">
            {player.corporation &&
              (getCorporationLogo(player.corporation.name, "w-12 h-5 shrink-0", "48px") ?? (
                <span className="truncate font-orbitron text-[11px] uppercase tracking-wider text-white/55">
                  {player.corporation.name}
                </span>
              ))}
            <span className="shrink-0 flex items-center gap-1">
              <GameIcon iconType={ResourceTypeTR} size="small" />
              <span className="font-orbitron text-[14px] font-bold tabular-nums text-white">
                {player.terraformRating}
              </span>
            </span>
          </span>
        </span>
        <span className="shrink-0">
          <ResourceProductionGrid player={player} size="compact" layout="row" />
        </span>
      </button>
      {hostActions.length > 0 && (
        <GameButton
          emphasis="quiet"
          aria-label={`Host actions for ${player.name}`}
          className="!p-0 !min-h-11 w-11 h-11 shrink-0 self-center mr-1 !text-xl leading-none"
          onClick={() => onOpenActions(player, hostActions)}
        >
          ⋯
        </GameButton>
      )}
    </div>
  );
});

export default PlayerRow;
