import React, { useMemo, useState } from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ColonyTileDto,
  ColonyOutputDto,
  ResourceTypeCredit,
  ResourceTypeEnergy,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { webSocketService } from "@/services/webSocketService.ts";
import { canPerformActions } from "@/utils/actionUtils.ts";
import { GamePopover } from "../GamePopover";
import ColonySteps, { getTradeExpression, mapOutputTypeToIcon } from "./ColonySteps.tsx";

type ColonyMode = "trade" | "build";

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
    void webSocketService.tradeWithColony(colonyId);
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
      position={{ type: "fixed", top: 60, left: 20 }}
      theme="colonies"
      excludeRef={anchorRef}
      header={{
        title: "Colonies",
        rightContent: toggleButton,
        showCloseButton: true,
      }}
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
  getPlayerColor: (playerId: string) => string;
  getPlayerName: (playerId: string) => string;
  onTrade: (colonyId: string) => void;
  onBuild: (colonyId: string) => void;
}

const ColonyTileCard: React.FC<ColonyTileCardProps> = ({
  colony,
  mode,
  canAct,
  getPlayerColor,
  getPlayerName,
  onTrade,
  onBuild,
}) => {
  const canTrade = canAct && colony.tradeAvailable;
  const canBuild = canAct && colony.buildAvailable;

  const isDisabled = mode === "trade" ? !canTrade : !canBuild;
  const dimmed = canAct && isDisabled;

  const markerOutput = colony.steps[colony.markerPosition]?.outputs ?? [];
  const buildReward = colony.colonies[0]?.reward ?? [];
  const tradeExpression = getTradeExpression(colony.steps);

  return (
    <div
      className={`rounded-lg border py-2.5 px-[15px] transition-all duration-200 ${
        dimmed ? "opacity-50" : ""
      }`}
      style={{ borderColor: colony.style.color + "60" }}
    >
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

      {/* Info row: cost → gain */}
      <div className="flex items-center gap-1.5 mb-2 h-7">
        {mode === "trade" ? (
          <>
            <span className="text-xs text-white/70 font-orbitron font-bold">3</span>
            <GameIcon iconType={ResourceTypeEnergy} size="small" />
            <span className="text-white/30 text-sm">→</span>
            <CostDisplay outputs={markerOutput} />
          </>
        ) : (
          <>
            <GameIcon iconType={ResourceTypeCredit} amount={17} size="small" />
            <span className="text-white/30 text-sm">→</span>
            <GameIcon iconType="colony-tile" size="small" />
            <CostDisplay outputs={buildReward} />
          </>
        )}
      </div>

      {/* Trade track label + expression */}
      <div className="flex items-center justify-between mb-1">
        <span className="text-white/50 text-[10px] font-orbitron uppercase tracking-wider">
          Trade
        </span>
        {tradeExpression && (
          <span className="inline-flex items-center gap-0.5">
            {tradeExpression.type === "x-icon" && (
              <span className="text-xs text-white/70 font-orbitron font-bold">X</span>
            )}
            <GameIcon iconType={tradeExpression.icon} size="small" />
          </span>
        )}
      </div>

      {/* Trade steps with colony slots */}
      <div className="mb-1.5">
        <ColonySteps
          steps={colony.steps}
          markerPosition={colony.markerPosition}
          playerColonies={colony.playerColonies}
          maxSlots={colony.colonies.length}
          getPlayerColor={getPlayerColor}
        />
      </div>

      {/* Colony bonus */}
      <div className="flex items-center gap-2 text-[10px] text-white/40">
        <span className="font-orbitron uppercase tracking-wider">Bonus</span>
        <OutputDisplay outputs={colony.colonyBonus} />
      </div>
    </div>
  );
};

interface OutputDisplayProps {
  outputs: ColonyOutputDto[];
}

const OutputDisplay: React.FC<OutputDisplayProps> = ({ outputs }) => {
  return (
    <span className="inline-flex items-center gap-0.5">
      {outputs.map((output, i) => (
        <span key={i} className="inline-flex items-center gap-0.5">
          <span className="text-[10px]">{output.amount}</span>
          <GameIcon iconType={mapOutputTypeToIcon(output.type)} size="small" />
        </span>
      ))}
    </span>
  );
};

const CostDisplay: React.FC<OutputDisplayProps> = ({ outputs }) => {
  return (
    <span className="inline-flex items-center gap-0.5">
      {outputs.map((output, i) => (
        <span key={i} className="inline-flex items-center gap-0.5">
          {output.amount > 1 && (
            <span className="text-xs text-white/70 font-orbitron font-bold">{output.amount}</span>
          )}
          <GameIcon iconType={mapOutputTypeToIcon(output.type)} size="small" />
        </span>
      ))}
    </span>
  );
};

export default ColonyPopover;
