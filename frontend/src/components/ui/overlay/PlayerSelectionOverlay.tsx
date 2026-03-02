import React from "react";
import { GameDto, OtherPlayerDto, PlayerDto } from "../../../types/generated/api-types.ts";
import { getCorporationLogo } from "../../../utils/corporationLogos.tsx";
import GameMenuModal from "./GameMenuModal.tsx";

interface PlayerSelectionOverlayProps {
  game: GameDto;
  onSelectPlayer: (playerId: string, playerName: string) => void;
  onCancel: () => void;
  visible?: boolean;
  onExited?: () => void;
}

const PlayerSelectionOverlay: React.FC<PlayerSelectionOverlayProps> = ({
  game,
  onSelectPlayer,
  onCancel,
  visible,
  onExited,
}) => {
  const allPlayers: (PlayerDto | OtherPlayerDto)[] = [
    ...(game.currentPlayer ? [game.currentPlayer] : []),
    ...(game.otherPlayers || []),
  ];

  const orderedPlayers =
    game.turnOrder && game.turnOrder.length > 0
      ? game.turnOrder
          .map((pid) => allPlayers.find((p) => p.id === pid))
          .filter((player): player is PlayerDto | OtherPlayerDto => player !== undefined)
      : allPlayers;

  return (
    <GameMenuModal
      title="Reconnect to Game"
      subtitle="Select a player to reconnect as"
      onBack={onCancel}
      visible={visible}
      onExited={onExited}
    >
      <div className="mb-6">
        <h3 className="text-white text-sm font-semibold mb-2 uppercase tracking-wide">Players</h3>
        <div className="flex flex-col gap-2">
          {orderedPlayers.map((player) => {
            const isExited = player.isExited;
            const isConnected = player.isConnected;
            const canSelect = !isConnected && !isExited;

            return (
              <button
                key={player.id}
                onClick={() => canSelect && onSelectPlayer(player.id, player.name)}
                disabled={!canSelect}
                className={`flex justify-between items-center py-3 px-4 bg-black/40 rounded-lg border transition-all text-left w-full ${
                  canSelect
                    ? "border-space-blue-600/50 hover:border-space-blue-400 hover:bg-black/60 cursor-pointer"
                    : "border-white/10 opacity-50 cursor-default"
                }`}
              >
                <div className="flex items-center gap-3">
                  {player.corporation && (
                    <div className="w-[80px] h-6 flex-shrink-0 flex items-center justify-start overflow-hidden">
                      <div className="origin-left scale-[0.2]">
                        {getCorporationLogo(
                          player.corporation.name.toLowerCase() as Parameters<
                            typeof getCorporationLogo
                          >[0],
                        )}
                      </div>
                    </div>
                  )}
                  <span className="text-white text-sm font-medium">{player.name}</span>
                </div>
                {isExited ? (
                  <span className="px-1.5 py-px rounded-lg text-[8px] font-semibold uppercase tracking-[0.3px] shadow-[0_1px_2px_rgba(0,0,0,0.2)] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white border border-[rgba(231,76,60,0.5)]">
                    EXITED
                  </span>
                ) : !isConnected ? (
                  <span className="px-1.5 py-px rounded-lg text-[8px] font-semibold uppercase tracking-[0.3px] shadow-[0_1px_2px_rgba(0,0,0,0.2)] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white border border-[rgba(231,76,60,0.5)]">
                    DISCONNECTED
                  </span>
                ) : null}
              </button>
            );
          })}
        </div>
      </div>

      <p className="text-white/40 text-xs text-center">Only disconnected players can be selected</p>
    </GameMenuModal>
  );
};

export default PlayerSelectionOverlay;
