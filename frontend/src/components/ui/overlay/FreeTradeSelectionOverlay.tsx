import React, { useState, useMemo } from "react";
import {
  PendingFreeTradeSelectionDto,
  ColonyTileDto,
  ColonyOutputDto,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import ColonySteps, { mapOutputTypeToIcon } from "../popover/ColonySteps.tsx";
import GameButton from "../buttons/GameButton.tsx";

interface PlayerInfo {
  id: string;
  name: string;
  color: string;
}

interface FreeTradeSelectionOverlayProps {
  isOpen: boolean;
  pendingSelection: PendingFreeTradeSelectionDto;
  colonyTiles: ColonyTileDto[];
  viewingPlayerId: string;
  tradeFleetAvailable: boolean;
  allPlayers: PlayerInfo[];
  onConfirm: (colonyId: string) => void;
}

const FreeTradeSelectionOverlay: React.FC<FreeTradeSelectionOverlayProps> = ({
  isOpen,
  pendingSelection,
  colonyTiles,
  viewingPlayerId,
  tradeFleetAvailable,
  allPlayers,
  onConfirm,
}) => {
  const [selectedColonyId, setSelectedColonyId] = useState<string | null>(null);

  const tradeableIds = useMemo(
    () => new Set(pendingSelection.availableColonyIds),
    [pendingSelection],
  );

  const isColonyTradeable = (colony: ColonyTileDto): boolean => {
    return tradeableIds.has(colony.id) && !colony.tradedThisGen && tradeFleetAvailable;
  };

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
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />

      <div className="relative z-[1] w-[480px] max-h-[80vh] flex flex-col bg-space-black-darker/95 border border-space-blue-500 rounded-lg overflow-hidden shadow-glow-lg">
        <div className="px-4 py-3 border-b border-white/10 flex items-center justify-between">
          <h2 className="font-orbitron text-base font-bold text-white tracking-wider m-0">
            Free Trade
          </h2>
          <span className="text-xs text-white/40">{pendingSelection.source}</span>
        </div>

        {/* Trade fleet status */}
        {!tradeFleetAvailable && (
          <div className="px-4 py-2 bg-red-900/20 border-b border-red-500/30">
            <span className="text-xs text-red-400 font-orbitron">No trade fleet available</span>
          </div>
        )}

        {/* Ships display */}
        <div className="px-4 py-2 border-b border-white/10 flex flex-wrap items-center gap-2">
          <span className="text-[10px] font-orbitron text-white/50 uppercase tracking-wider">
            Ships:
          </span>
          {allPlayers.map((player) => (
            <div key={player.id} className="flex items-center gap-1">
              <div className="w-3 h-3 rounded-sm" style={{ backgroundColor: player.color }} />
              <span className="text-[10px] font-orbitron text-white/60">{player.name}</span>
            </div>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto p-3 space-y-2">
          {colonyTiles.map((colony) => {
            const tradeable = isColonyTradeable(colony);
            const isSelected = selectedColonyId === colony.id;
            const markerOutput = colony.steps[colony.markerPosition]?.outputs ?? [];

            const viewerColonyCount = colony.playerColonies.filter(
              (id) => id === viewingPlayerId,
            ).length;
            const tradeGainOutputs: ColonyOutputDto[] = [...markerOutput];
            if (viewerColonyCount > 0) {
              for (const bonus of colony.colonyBonus) {
                const scaledAmount = bonus.amount * viewerColonyCount;
                const existing = tradeGainOutputs.find((o) => o.type === bonus.type);
                if (existing) {
                  tradeGainOutputs[tradeGainOutputs.indexOf(existing)] = {
                    ...existing,
                    amount: existing.amount + scaledAmount,
                  };
                } else {
                  tradeGainOutputs.push({ ...bonus, amount: scaledAmount });
                }
              }
            }

            return (
              <button
                key={colony.id}
                type="button"
                className={`w-full text-left px-3 py-2.5 rounded border transition-all ${
                  !tradeable
                    ? "border-white/5 bg-white/[0.01] opacity-40 cursor-default"
                    : isSelected
                      ? "bg-white/10 cursor-pointer"
                      : "border-white/10 bg-white/[0.02] hover:border-white/25 hover:bg-white/[0.04] cursor-pointer"
                }`}
                style={{
                  borderColor: isSelected && tradeable ? colony.style.color : undefined,
                }}
                onClick={() => {
                  if (tradeable) {
                    setSelectedColonyId(colony.id);
                  }
                }}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <h3 className="text-white text-xs font-bold font-orbitron m-0">
                      {colony.name}
                    </h3>
                    {colony.tradedThisGen && (
                      <span className="text-[9px] font-orbitron text-white/30 uppercase">
                        Traded
                      </span>
                    )}
                  </div>
                  {tradeable && (
                    <div className="flex items-center gap-1">
                      <span className="text-white/30 text-sm">→</span>
                      <OutputDisplay outputs={tradeGainOutputs} />
                    </div>
                  )}
                </div>

                <ColonySteps
                  steps={colony.steps}
                  markerPosition={colony.markerPosition}
                  tradeStepBonus={colony.tradeStepBonus}
                  playerColonies={colony.playerColonies}
                  maxSlots={colony.colonies.length}
                  getPlayerColor={getPlayerColor}
                  getPlayerName={getPlayerName}
                />

                <div className="flex items-center gap-1.5 mt-1.5 text-[9px] text-white/40">
                  <span className="font-orbitron uppercase tracking-wider">Colony Bonus</span>
                  <OutputDisplay outputs={colony.colonyBonus} />
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
            Confirm Trade
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

export default FreeTradeSelectionOverlay;
