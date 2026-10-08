import { BotPersonaChip, PlayerChip } from "../../ui/display/BotChips.tsx";
import type { TableauPlayer } from "./tableauPlayers.ts";

const CHIP_BASE = "tracking-[0.5px] border [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]";

interface PlayerStatusChipsProps {
  player: TableauPlayer;
  isMe: boolean;
  isTurn: boolean;
  isActionPhase: boolean;
}

export default function PlayerStatusChips({
  player,
  isMe,
  isTurn,
  isActionPhase,
}: PlayerStatusChipsProps) {
  const isBot = player.playerType === "bot";
  const isDisconnected = !isBot && !player.isConnected;
  const showTurn = isTurn && isActionPhase && !player.passed;
  const actionsLabel =
    player.availableActions === -1 ? "∞" : `${player.availableActions}/${player.totalActions}`;

  return (
    <div className="player-chip-group gap-1 [&_.player-chip]:text-[11px]">
      {isMe && (
        <PlayerChip
          className={`${CHIP_BASE} bg-[rgba(60,100,150,0.8)] text-white border-[rgba(80,130,180,0.7)]`}
        >
          YOU
        </PlayerChip>
      )}
      {showTurn && (
        <PlayerChip
          className={`${CHIP_BASE} bg-[rgba(50,100,160,0.6)] text-white border-[rgba(80,140,200,0.8)]`}
        >
          TURN {actionsLabel}
        </PlayerChip>
      )}
      {player.passed && !player.isExited && (
        <PlayerChip
          className={`${CHIP_BASE} bg-[rgba(80,80,90,0.6)] text-[rgb(140,140,150)] border-[rgba(60,60,70,0.7)]`}
        >
          PASSED
        </PlayerChip>
      )}
      {isBot && <BotPersonaChip persona={player.botPersona} botStatus={player.botStatus} />}
      {isBot && player.botStatus === "thinking" && (
        <PlayerChip
          className={`${CHIP_BASE} bg-[rgba(120,80,200,0.6)] text-[rgb(200,180,255)] border-[rgba(140,100,220,0.7)]`}
        >
          THINKING
        </PlayerChip>
      )}
      {player.isExited && (
        <PlayerChip
          className={`${CHIP_BASE} bg-[rgba(180,60,60,0.4)] text-[rgb(220,140,140)] border-[rgba(180,60,60,0.5)]`}
        >
          EXITED
        </PlayerChip>
      )}
      {isDisconnected && !player.isExited && (
        <PlayerChip
          className={`${CHIP_BASE} bg-[rgba(180,60,60,0.4)] text-[rgb(220,140,140)] border-[rgba(180,60,60,0.5)]`}
        >
          DISCONNECTED
        </PlayerChip>
      )}
    </div>
  );
}
