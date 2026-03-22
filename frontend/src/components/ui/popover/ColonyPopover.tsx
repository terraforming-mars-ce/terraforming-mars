import React, { useMemo, useState } from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ColonyTileDto,
  ColonyOutputDto,
  ResourceTypeCredit,
  ResourceTypeEnergy,
  ResourceTypeTitanium,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { webSocketService } from "@/services/webSocketService.ts";
import { canPerformActions } from "@/utils/actionUtils.ts";
import { GamePopover, GamePopoverItem } from "../GamePopover";
import ColonySteps, { getTradeExpression, mapOutputTypeToIcon } from "./ColonySteps.tsx";

type ColonyMode = "trade" | "build";
type TradePaymentType = "credits" | "energy" | "titanium";

interface ColonyPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

interface PlayerInfo {
  id: string;
  name: string;
  color: string;
}

const ColonyPopover: React.FC<ColonyPopoverProps> = ({
  isVisible,
  onClose,
  gameState,
  anchorRef,
}) => {
  const [mode, setMode] = useState<ColonyMode>("trade");

  const resources = gameState?.currentPlayer?.resources;
  const canAffordCredits = (resources?.credits ?? 0) >= 9;
  const canAffordEnergy = (resources?.energy ?? 0) >= 3;
  const canAffordTitanium = (resources?.titanium ?? 0) >= 3;

  const defaultPayment = (): TradePaymentType => {
    if (canAffordCredits) return "credits";
    if (canAffordEnergy) return "energy";
    if (canAffordTitanium) return "titanium";
    return "credits";
  };

  const [tradePayment, setTradePayment] = useState<TradePaymentType>(defaultPayment);

  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
  const canAct =
    isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);

  const colonyTiles = gameState?.colonyTiles ?? [];

  const allPlayers: PlayerInfo[] = useMemo(() => {
    if (!gameState) return [];
    const players: PlayerInfo[] = [];
    if (gameState.currentPlayer?.id) {
      players.push({
        id: gameState.currentPlayer.id,
        name: gameState.currentPlayer.name,
        color: gameState.currentPlayer.color,
      });
    }
    for (const p of gameState.otherPlayers) {
      players.push({ id: p.id, name: p.name, color: p.color });
    }
    return players;
  }, [gameState]);

  const getPlayerColor = (playerId: string): string => {
    return allPlayers.find((p) => p.id === playerId)?.color ?? "#666";
  };

  const getPlayerName = (playerId: string): string => {
    return allPlayers.find((p) => p.id === playerId)?.name ?? "Unknown";
  };

  const handleTrade = (colonyId: string) => {
    if (!canAct) return;
    void webSocketService.tradeWithColony(colonyId, tradePayment);
  };

  const handleBuild = (colonyId: string) => {
    if (!canAct) return;
    void webSocketService.buildColony(colonyId);
  };

  const toggleButton = (
    <div className="flex rounded overflow-hidden border border-white/20">
      <button
        className={`px-3 py-1 text-[11px] font-orbitron font-bold transition-colors cursor-pointer ${
          mode === "trade"
            ? "bg-white/20 text-white"
            : "bg-transparent text-white/40 hover:text-white/60"
        }`}
        onClick={() => setMode("trade")}
      >
        Trade
      </button>
      <button
        className={`px-3 py-1 text-[11px] font-orbitron font-bold transition-colors cursor-pointer ${
          mode === "build"
            ? "bg-white/20 text-white"
            : "bg-transparent text-white/40 hover:text-white/60"
        }`}
        onClick={() => setMode("build")}
      >
        Build
      </button>
    </div>
  );

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "below" }}
      theme="colonies"
      excludeRef={anchorRef}
      header={
        mode === "trade"
          ? {
              title: "",
              badge: (
                <TradePaymentSelector
                  selected={tradePayment}
                  onSelect={setTradePayment}
                  canAffordCredits={canAffordCredits}
                  canAffordEnergy={canAffordEnergy}
                  canAffordTitanium={canAffordTitanium}
                />
              ),
              rightContent: toggleButton,
            }
          : {
              title: "",
              rightContent: toggleButton,
            }
      }
      width={560}
      maxHeight="80vh"
      animation="slideDown"
      className="!bg-space-black-darker"
    >
      <div className="p-2 space-y-2">
        {colonyTiles.map((colony) => (
          <ColonyTileCard
            key={colony.id}
            colony={colony}
            mode={mode}
            canAct={canAct}
            viewingPlayerId={gameState?.viewingPlayerId ?? ""}
            tradePayment={tradePayment}
            getPlayerColor={getPlayerColor}
            getPlayerName={getPlayerName}
            onTrade={handleTrade}
            onBuild={handleBuild}
          />
        ))}
      </div>
    </GamePopover>
  );
};

interface ColonyTileCardProps {
  colony: ColonyTileDto;
  mode: ColonyMode;
  canAct: boolean;
  viewingPlayerId: string;
  tradePayment: TradePaymentType;
  getPlayerColor: (playerId: string) => string;
  getPlayerName: (playerId: string) => string;
  onTrade: (colonyId: string) => void;
  onBuild: (colonyId: string) => void;
}

const TRADE_PAYMENT_CONFIG: Record<TradePaymentType, { icon: string; amount: number }> = {
  credits: { icon: ResourceTypeCredit, amount: 9 },
  energy: { icon: ResourceTypeEnergy, amount: 3 },
  titanium: { icon: ResourceTypeTitanium, amount: 3 },
};

const ColonyTileCard: React.FC<ColonyTileCardProps> = ({
  colony,
  mode,
  canAct,
  viewingPlayerId,
  tradePayment,
  getPlayerColor,
  getPlayerName,
  onTrade,
  onBuild,
}) => {
  const canTrade = canAct && colony.tradeAvailable;
  const canBuild = canAct && colony.buildAvailable;

  const isDisabled = mode === "trade" ? !canTrade : !canBuild;
  const dimmed = canAct && isDisabled;

  const boostedPosition = Math.min(
    colony.markerPosition + (colony.tradeStepBonus ?? 0),
    colony.steps.length - 1,
  );
  const markerOutput = colony.steps[boostedPosition]?.outputs ?? [];
  const buildReward = colony.colonies[0]?.reward ?? [];
  const tradeExpression = getTradeExpression(colony.steps);

  const viewerColonyCount = colony.playerColonies.filter((id) => id === viewingPlayerId).length;
  const tradeGainOutputs: ColonyOutputDto[] = useMemo(() => {
    const combined = [...markerOutput];
    if (viewerColonyCount > 0) {
      for (const bonus of colony.colonyBonus) {
        const scaledAmount = bonus.amount * viewerColonyCount;
        const existing = combined.find((o) => o.type === bonus.type);
        if (existing) {
          combined[combined.indexOf(existing)] = {
            ...existing,
            amount: existing.amount + scaledAmount,
          };
        } else {
          combined.push({ ...bonus, amount: scaledAmount });
        }
      }
    }
    return combined;
  }, [markerOutput, viewerColonyCount, colony.colonyBonus]);

  return (
    <GamePopoverItem state={dimmed ? "disabled" : "available"} borderColor={colony.style.color}>
      {/* Header: name, traded chip, action button */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <h3 className="text-white text-sm font-bold font-orbitron m-0">{colony.name}</h3>
          {colony.tradedThisGen && colony.traderId && (
            <span
              className="px-2 py-0.5 rounded-full text-[10px] font-orbitron text-white"
              style={{ backgroundColor: getPlayerColor(colony.traderId) }}
            >
              {getPlayerName(colony.traderId)}
            </span>
          )}
        </div>

        {canAct && (
          <div>
            {mode === "trade" ? (
              <button
                className={`px-3 py-1 rounded text-xs font-semibold font-orbitron transition-all cursor-pointer ${
                  canTrade
                    ? "bg-white/15 hover:bg-white/25 text-white"
                    : "bg-gray-600/30 text-gray-500"
                }`}
                onClick={(e) => {
                  e.stopPropagation();
                  if (canTrade) onTrade(colony.id);
                }}
                disabled={!canTrade}
              >
                Trade
              </button>
            ) : (
              <button
                className={`px-3 py-1 rounded text-xs font-semibold font-orbitron transition-all cursor-pointer ${
                  canBuild
                    ? "bg-white/15 hover:bg-white/25 text-white"
                    : "bg-gray-600/30 text-gray-500"
                }`}
                onClick={(e) => {
                  e.stopPropagation();
                  if (canBuild) onBuild(colony.id);
                }}
                disabled={!canBuild}
              >
                Build
              </button>
            )}
          </div>
        )}
      </div>

      {/* Info row: cost → gain (layered for cross-fade without re-mount) */}
      <div className="relative h-7">
        {(["credits", "energy", "titanium"] as TradePaymentType[]).map((pt) => {
          const config = TRADE_PAYMENT_CONFIG[pt];
          return (
            <div
              key={pt}
              className="absolute inset-0 flex items-center gap-1.5"
              style={{
                opacity: mode === "trade" && !colony.tradedThisGen && tradePayment === pt ? 1 : 0,
                transition: mode === "trade" ? "opacity 300ms" : "none",
              }}
            >
              {pt !== "credits" && (
                <span className="text-xs text-white/70 font-orbitron font-bold">
                  {config.amount}
                </span>
              )}
              <GameIcon
                iconType={config.icon}
                amount={pt === "credits" ? config.amount : undefined}
                size="small"
              />
              <span className="text-white/30 text-sm">→</span>
              <CostDisplay outputs={tradeGainOutputs} />
            </div>
          );
        })}
        <div
          className="absolute inset-0 flex items-center"
          style={{
            opacity: mode === "trade" && colony.tradedThisGen ? 1 : 0,
            transition: mode === "trade" ? "opacity 300ms" : "none",
          }}
        >
          <span className="text-[10px] font-orbitron font-bold text-white/30 uppercase tracking-wider">
            Already Traded
          </span>
        </div>
        <div
          className="absolute inset-0 flex items-center gap-1.5"
          style={{
            opacity: mode === "build" ? 1 : 0,
            transition: mode === "build" ? "opacity 300ms" : "none",
          }}
        >
          <GameIcon iconType={ResourceTypeCredit} amount={17} size="small" />
          <span className="text-white/30 text-sm">→</span>
          <GameIcon iconType="colony-tile" size="small" />
          <CostDisplay outputs={buildReward} />
        </div>
      </div>

      <div style={{ height: "1px", background: "rgba(255,255,255,0.15)", margin: "16px 0" }} />

      {/* Trade steps with colony slots */}
      <div className="mb-3">
        <ColonySteps
          steps={colony.steps}
          markerPosition={colony.markerPosition}
          tradeStepBonus={colony.tradeStepBonus}
          playerColonies={colony.playerColonies}
          maxSlots={colony.colonies.length}
          getPlayerColor={getPlayerColor}
          getPlayerName={getPlayerName}
        />
      </div>

      {/* Colony bonus + trade gain expression */}
      <div className="flex items-center justify-between text-[10px] text-white/40">
        <div className="flex items-center gap-2">
          <span className="font-orbitron uppercase tracking-wider">Colony Bonus</span>
          <OutputDisplay outputs={colony.colonyBonus} />
        </div>
        {tradeExpression && (
          <div className="flex items-center gap-2">
            <span className="font-orbitron uppercase tracking-wider">Trade Gain</span>
            <span className="inline-flex items-center gap-0.5">
              {tradeExpression.type === "x-icon" && !tradeExpression.isCreditType && (
                <span className="text-[10px] text-white/70 font-orbitron font-bold">X</span>
              )}
              <GameIcon
                iconType={tradeExpression.icon}
                amount={
                  tradeExpression.type === "x-icon" && tradeExpression.isCreditType
                    ? "X"
                    : undefined
                }
                size="small"
              />
            </span>
          </div>
        )}
      </div>
    </GamePopoverItem>
  );
};

interface TradePaymentSelectorProps {
  selected: TradePaymentType;
  onSelect: (type: TradePaymentType) => void;
  canAffordCredits: boolean;
  canAffordEnergy: boolean;
  canAffordTitanium: boolean;
}

const TradePaymentSelector: React.FC<TradePaymentSelectorProps> = ({
  selected,
  onSelect,
  canAffordCredits,
  canAffordEnergy,
  canAffordTitanium,
}) => {
  const options: { type: TradePaymentType; icon: string; canAfford: boolean }[] = [
    { type: "credits", icon: ResourceTypeCredit, canAfford: canAffordCredits },
    { type: "energy", icon: ResourceTypeEnergy, canAfford: canAffordEnergy },
    { type: "titanium", icon: ResourceTypeTitanium, canAfford: canAffordTitanium },
  ];

  return (
    <div className="flex rounded overflow-hidden border border-white/20">
      {options.map((opt) => {
        const isSelected = selected === opt.type;
        return (
          <button
            key={opt.type}
            className={`flex items-center px-1.5 py-0.5 transition-colors cursor-pointer ${
              isSelected
                ? "bg-white/20 text-white"
                : opt.canAfford
                  ? "bg-transparent text-white/40 hover:text-white/60"
                  : "bg-transparent text-white/20 opacity-40"
            }`}
            onClick={() => onSelect(opt.type)}
          >
            <GameIcon
              iconType={opt.icon}
              amount={opt.type === "credits" ? "X" : undefined}
              size="small"
            />
          </button>
        );
      })}
    </div>
  );
};

interface OutputDisplayProps {
  outputs: ColonyOutputDto[];
}

const OutputDisplay: React.FC<OutputDisplayProps> = ({ outputs }) => {
  return (
    <span className="inline-flex items-center gap-0.5">
      {outputs.map((output, i) => {
        const icon = mapOutputTypeToIcon(output.type);
        const useAmountProp = output.type === "credit" || output.type === "credit-production";
        return (
          <span key={i} className="inline-flex items-center gap-0.5">
            {!useAmountProp && (
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

const CostDisplay: React.FC<OutputDisplayProps> = ({ outputs }) => {
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

export default ColonyPopover;
