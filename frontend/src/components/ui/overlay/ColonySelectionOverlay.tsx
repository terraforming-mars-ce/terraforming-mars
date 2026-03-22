import React, { useState, useMemo } from "react";
import {
  PendingColonySelectionDto,
  ColonyTileDto,
  ColonyOutputDto,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import ColonySteps, { mapOutputTypeToIcon } from "../popover/ColonySteps.tsx";
import GameButton from "../buttons/GameButton.tsx";
import {
  OVERLAY_BACKGROUND_CLASS,
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
} from "./overlayStyles.ts";

interface PlayerInfo {
  id: string;
  name: string;
  color: string;
}

interface ColonySelectionOverlayProps {
  isOpen: boolean;
  pendingSelection: PendingColonySelectionDto;
  colonyTiles: ColonyTileDto[];
  viewingPlayerId: string;
  allPlayers: PlayerInfo[];
  onConfirm: (colonyId: string) => void;
}

const ColonySelectionOverlay: React.FC<ColonySelectionOverlayProps> = ({
  isOpen,
  pendingSelection,
  colonyTiles,
  viewingPlayerId,
  allPlayers,
  onConfirm,
}) => {
  const [selectedColonyId, setSelectedColonyId] = useState<string | null>(null);

  const availableColonies = useMemo(() => {
    const availableIds = new Set(pendingSelection.availableColonyIds);
    return colonyTiles.filter((colony) => {
      if (!availableIds.has(colony.id)) {
        return false;
      }
      const isFull = colony.playerColonies.length >= colony.colonies.length;
      if (isFull) {
        return false;
      }
      if (
        !pendingSelection.allowDuplicatePlayerColony &&
        colony.playerColonies.includes(viewingPlayerId)
      ) {
        return false;
      }
      return true;
    });
  }, [colonyTiles, pendingSelection, viewingPlayerId]);

  const getPlayerColor = (playerId: string): string => {
    return allPlayers.find((p) => p.id === playerId)?.color ?? "#666";
  };

  const getPlayerName = (playerId: string): string => {
    return allPlayers.find((p) => p.id === playerId)?.name ?? "Unknown";
  };

  if (!isOpen || availableColonies.length === 0) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      <div className={OVERLAY_BACKGROUND_CLASS} />

      <div className={OVERLAY_CONTAINER_CLASS}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Place Colony</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Select a colony to place from {pendingSelection.source}.
          </p>
        </div>

        <div className="flex-1 overflow-y-auto p-6 space-y-3">
          {availableColonies.map((colony) => {
            const isSelected = selectedColonyId === colony.id;
            const nextSlotIndex = colony.playerColonies.length;
            const reward = colony.colonies[nextSlotIndex]?.reward ?? [];

            return (
              <button
                key={colony.id}
                type="button"
                className={`w-full text-left p-4 rounded-lg border-2 transition-all cursor-pointer ${
                  isSelected
                    ? "border-white/60 bg-white/10"
                    : "border-white/10 bg-white/[0.02] hover:border-white/30 hover:bg-white/[0.04]"
                }`}
                style={{
                  borderColor: isSelected ? colony.style.color : undefined,
                }}
                onClick={() => setSelectedColonyId(colony.id)}
              >
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-white text-sm font-bold font-orbitron m-0">{colony.name}</h3>
                  {reward.length > 0 && (
                    <div className="flex items-center gap-1.5">
                      <span className="text-[10px] font-orbitron text-white/40 uppercase tracking-wider">
                        Reward
                      </span>
                      <OutputDisplay outputs={reward} />
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

                <div className="flex items-center gap-2 mt-2 text-[10px] text-white/40">
                  <span className="font-orbitron uppercase tracking-wider">Colony Bonus</span>
                  <OutputDisplay outputs={colony.colonyBonus} />
                </div>
              </button>
            );
          })}
        </div>

        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="text-sm text-white/70">
            {selectedColonyId
              ? `Selected: ${availableColonies.find((c) => c.id === selectedColonyId)?.name}`
              : "Select a colony"}
          </div>
          <GameButton
            size="lg"
            onClick={() => {
              if (selectedColonyId) {
                onConfirm(selectedColonyId);
              }
            }}
            disabled={!selectedColonyId}
          >
            Confirm Placement
          </GameButton>
        </div>
      </div>
    </div>
  );
};

const OutputDisplay: React.FC<{ outputs: ColonyOutputDto[] }> = ({ outputs }) => {
  return (
    <span className="inline-flex items-center gap-0.5">
      {outputs.map((output, i) => {
        const icon = mapOutputTypeToIcon(output.type);
        const useAmountProp = output.type === "credit" || output.type === "credit-production";
        return (
          <span key={i} className="inline-flex items-center gap-0.5">
            {!useAmountProp && output.amount > 1 && (
              <span className="text-xs text-white/70 font-orbitron font-bold">{output.amount}</span>
            )}
            <GameIcon
              iconType={icon}
              amount={useAmountProp ? output.amount : undefined}
              size="small"
            />
          </span>
        );
      })}
    </span>
  );
};

export default ColonySelectionOverlay;
