import React, { useState, useMemo } from "react";
import { PendingColonySelectionDto, ColonyTileDto } from "@/types/generated/api-types.ts";
import ColonySteps from "../popover/ColonySteps.tsx";
import GameButton from "../buttons/GameButton.tsx";
import ColonyOutputDisplay from "../display/ColonyOutputDisplay.tsx";
import { PlayerInfo } from "@/utils/colonyUtils.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface ColonySelectionOverlayProps {
  isOpen: boolean;
  pendingSelection: PendingColonySelectionDto;
  colonyTiles: ColonyTileDto[];
  allPlayers: PlayerInfo[];
  onConfirm: (colonyId: string) => void;
}

const ColonySelectionOverlay: React.FC<ColonySelectionOverlayProps> = ({
  isOpen,
  pendingSelection,
  colonyTiles,
  allPlayers,
  onConfirm,
}) => {
  const [selectedColonyId, setSelectedColonyId] = useState<string | null>(null);

  const selectableIds = useMemo(
    () => new Set(pendingSelection.availableColonyIds),
    [pendingSelection],
  );

  const getPlayerColor = (playerId: string): string => {
    return allPlayers.find((p) => p.id === playerId)?.color ?? "#666";
  };

  const getPlayerName = (playerId: string): string => {
    return allPlayers.find((p) => p.id === playerId)?.name ?? "Unknown";
  };

  if (!isOpen) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 flex items-center justify-center"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div className="absolute inset-0 backdrop-blur-sm" />
      <div className="absolute inset-0 bg-black/60 animate-[fadeIn_0.3s_ease]" />

      <div className="relative z-[1] w-[480px] max-h-[80vh] flex flex-col bg-space-black-darker/95 border border-space-blue-500 rounded-lg overflow-hidden shadow-glow-lg">
        <div className="px-4 py-3 border-b border-white/10 flex items-center justify-between">
          <h2 className="font-orbitron text-base font-bold text-white tracking-wider m-0">
            Place Colony
          </h2>
          <span className="text-xs text-white/40">{pendingSelection.source}</span>
        </div>

        <div className="flex-1 overflow-y-auto p-3 space-y-2">
          {colonyTiles.map((colony) => {
            const selectable = selectableIds.has(colony.id);
            const isSelected = selectedColonyId === colony.id;
            const nextSlotIndex = colony.playerColonies.length;
            const reward = colony.colonies[nextSlotIndex]?.reward ?? [];

            return (
              <button
                key={colony.id}
                type="button"
                className={`w-full text-left px-3 py-2.5 rounded border transition-all ${
                  !selectable
                    ? "border-white/5 bg-white/[0.01] opacity-40 cursor-default"
                    : isSelected
                      ? "bg-white/10 cursor-pointer"
                      : "border-white/10 bg-white/[0.02] hover:border-white/25 hover:bg-white/[0.04] cursor-pointer"
                }`}
                style={{
                  borderColor: isSelected && selectable ? colony.style.color : undefined,
                }}
                onClick={() => {
                  if (selectable) {
                    setSelectedColonyId(colony.id);
                  }
                }}
              >
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-white text-xs font-bold font-orbitron m-0">{colony.name}</h3>
                  {reward.length > 0 && (
                    <div className="flex items-center gap-1">
                      <span className="text-[9px] font-orbitron text-white/40 uppercase">
                        Reward
                      </span>
                      <ColonyOutputDisplay outputs={reward} />
                    </div>
                  )}
                </div>

                <ColonySteps
                  steps={colony.steps}
                  markerPosition={colony.markerPosition}
                  playerColonies={colony.playerColonies}
                  maxSlots={colony.colonies.length}
                  getPlayerColor={getPlayerColor}
                  getPlayerName={getPlayerName}
                />

                <div className="flex items-center gap-1.5 mt-1.5 text-[9px] text-white/40">
                  <span className="font-orbitron uppercase tracking-wider">Colony Bonus</span>
                  <ColonyOutputDisplay outputs={colony.colonyBonus} />
                </div>
              </button>
            );
          })}
        </div>

        <div className="px-4 py-3 border-t border-white/10 flex items-center justify-end">
          <GameButton
            size="sm"
            onClick={() => {
              if (selectedColonyId) {
                onConfirm(selectedColonyId);
              }
            }}
            disabled={!selectedColonyId}
          >
            Confirm
          </GameButton>
        </div>
      </div>
    </div>
  );
};

export default ColonySelectionOverlay;
