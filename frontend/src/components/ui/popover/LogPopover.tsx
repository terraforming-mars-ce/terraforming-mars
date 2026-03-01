import React, { useState, useEffect, useRef, useMemo, useCallback } from "react";
import { StateDiffDto, GameDto, CalculatedOutputDto } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import GameIcon from "@/components/ui/display/GameIcon.tsx";
import VictoryPointIcon from "@/components/ui/display/VictoryPointIcon.tsx";
import BehaviorSection from "@/components/ui/cards/BehaviorSection";
import { GamePopover } from "../GamePopover";

interface LogPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  anchorRef: React.RefObject<HTMLElement>;
  gameId: string;
  gameState?: GameDto;
}

const TagIcon: React.FC<{ tag: string }> = ({ tag }) => {
  return <GameIcon iconType={tag} size="small" />;
};

const resourceTypeToIconType: Record<string, string> = {
  credits: "credit",
  steel: "steel",
  titanium: "titanium",
  plants: "plant",
  energy: "energy",
  heat: "heat",
  "credits-production": "credit-production",
  "steel-production": "steel-production",
  "titanium-production": "titanium-production",
  "plants-production": "plant-production",
  "energy-production": "energy-production",
  "heat-production": "heat-production",
  tr: "tr",
  oxygen: "oxygen",
  temperature: "temperature",
  "ocean-placement": "ocean-placement",
  "greenery-placement": "greenery-placement",
  "city-placement": "city-placement",
};

const TILE_PLACEMENT_TYPES = ["ocean-placement", "greenery-placement", "city-placement"];
const GLOBAL_PARAMETER_TYPES = ["temperature", "oxygen"];

const BEHAVIOR_OUTPUT_TYPES = [...TILE_PLACEMENT_TYPES, ...GLOBAL_PARAMETER_TYPES];

const CalculatedOutputsDisplay: React.FC<{
  outputs: CalculatedOutputDto[];
  showAll?: boolean;
  excludeBehaviors?: boolean;
}> = ({ outputs, showAll = false, excludeBehaviors = false }) => {
  const outputsToShow = showAll
    ? outputs.filter(
        (o) =>
          o.amount !== 0 && (!excludeBehaviors || !BEHAVIOR_OUTPUT_TYPES.includes(o.resourceType)),
      )
    : outputs.filter(
        (o) =>
          o.isScaled &&
          o.amount !== 0 &&
          (!excludeBehaviors || !BEHAVIOR_OUTPUT_TYPES.includes(o.resourceType)),
      );

  if (outputsToShow.length === 0) return null;

  return (
    <div className="mt-1 flex flex-wrap items-center gap-2 px-1">
      <span className="text-[10px] text-gray-400 uppercase tracking-wider">Gained:</span>
      {outputsToShow.map((output, index) => {
        const iconType = resourceTypeToIconType[output.resourceType] || output.resourceType;
        return (
          <div key={index} className="flex items-center gap-0.5">
            <GameIcon iconType={iconType} amount={output.amount} size="small" />
          </div>
        );
      })}
    </div>
  );
};

interface LogEntryProps {
  diff: StateDiffDto;
  playerNames: Map<string, string>;
}

interface PlayerTurnGroup {
  playerId: string;
  entries: StateDiffDto[];
}

interface LogGroup {
  generation: number;
  playerTurns: PlayerTurnGroup[];
}

function hexToPlayerStyle(hex: string): { border: string; bg: string; text: string } {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return {
    border: `rgba(${r}, ${g}, ${b}, 0.4)`,
    bg: `rgba(${r}, ${g}, ${b}, 0.05)`,
    text: hex,
  };
}

const DEFAULT_PLAYER_STYLE = {
  border: "rgba(100, 200, 255, 0.4)",
  bg: "rgba(100, 200, 255, 0.05)",
  text: "#64c8ff",
};

const groupLogsByGeneration = (logs: StateDiffDto[]): LogGroup[] => {
  if (logs.length === 0) return [];

  const groups: LogGroup[] = [];
  let currentGeneration = 1;

  for (const log of logs) {
    if (log.changes?.generation) {
      currentGeneration = log.changes.generation.new;
    }

    let genGroup = groups.find((g) => g.generation === currentGeneration);
    if (!genGroup) {
      genGroup = { generation: currentGeneration, playerTurns: [] };
      groups.push(genGroup);
    }

    const lastTurn = genGroup.playerTurns[genGroup.playerTurns.length - 1];
    if (lastTurn && lastTurn.playerId === log.playerId) {
      lastTurn.entries.push(log);
    } else {
      genGroup.playerTurns.push({ playerId: log.playerId, entries: [log] });
    }
  }

  return groups;
};

interface GenerationDividerProps {
  generation: number;
  entryCount: number;
}

const GenerationDivider: React.FC<GenerationDividerProps> = ({ generation, entryCount }) => (
  <div className="sticky top-0 z-10 flex items-center gap-2 py-2 px-3 bg-[#1a1a2e] border-b border-[rgba(100,200,255,0.3)]">
    <div className="flex items-center gap-1.5">
      <span className="text-[10px] font-bold uppercase tracking-wider text-[#64c8ff]">
        Generation {generation}
      </span>
    </div>
    <div className="flex-1 h-px bg-gradient-to-r from-[rgba(100,200,255,0.3)] to-transparent" />
    <span className="text-[9px] text-gray-500">{entryCount} actions</span>
  </div>
);

interface PlayerTurnSectionProps {
  playerName: string;
  entries: StateDiffDto[];
  playerColor: string | undefined;
  playerNames: Map<string, string>;
}

const PlayerTurnSection: React.FC<PlayerTurnSectionProps> = ({
  playerName,
  entries,
  playerColor,
  playerNames,
}) => {
  const color = playerColor ? hexToPlayerStyle(playerColor) : DEFAULT_PLAYER_STYLE;

  return (
    <div
      className="rounded-lg mb-2 overflow-hidden"
      style={{
        borderLeft: `3px solid ${color.border}`,
        backgroundColor: color.bg,
      }}
    >
      <div
        className="flex items-center gap-2 py-1.5 px-3"
        style={{ borderBottom: `1px solid ${color.border}` }}
      >
        <div className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: color.text }} />
        <span
          className="text-[10px] font-semibold uppercase tracking-wider"
          style={{ color: color.text }}
        >
          {playerName}
        </span>
        <span className="text-[9px] text-gray-500">
          {entries.length} action{entries.length !== 1 ? "s" : ""}
        </span>
      </div>
      <div className="flex flex-col">
        {entries.map((diff) => (
          <LogEntry key={diff.sequenceNumber} diff={diff} playerNames={playerNames} />
        ))}
      </div>
    </div>
  );
};

const LogEntry: React.FC<LogEntryProps> = ({ diff, playerNames }) => {
  const isCardPlay = diff.sourceType === "card_play";
  const isCardAction = diff.sourceType === "card_action";
  const isStandardProject = diff.sourceType === "standard_project";
  const isResourceConvert = diff.sourceType === "resource_convert";
  const isGameEvent = diff.sourceType === "game_event";
  const isCardSource = isCardPlay || isCardAction;
  const isBehaviorSource = isStandardProject || isResourceConvert || isGameEvent;
  const displayData = diff.displayData;

  const playerName = playerNames.get(diff.playerId) || "Unknown";

  const cardTags = displayData?.tags || [];
  const vpConditions = displayData?.vpConditions || [];
  const behaviorsToShow = displayData?.behaviors || [];

  // Determine the choice display mode
  const choiceDisplayInfo = useMemo(() => {
    if (diff.choiceIndex === undefined || diff.choiceIndex === null) {
      return { hasChoices: false, type: "none" as const };
    }

    // Check if the behavior has a single behavior with choices array (e.g., Artificial Photosynthesis)
    if (
      behaviorsToShow.length === 1 &&
      behaviorsToShow[0].choices &&
      behaviorsToShow[0].choices.length > 0
    ) {
      return {
        hasChoices: true,
        type: "within-behavior" as const,
        choices: behaviorsToShow[0].choices,
      };
    }

    // Check if there are multiple behaviors (OR between behaviors)
    if (isCardPlay && behaviorsToShow.length > 1) {
      return { hasChoices: true, type: "between-behaviors" as const };
    }

    return { hasChoices: false, type: "none" as const };
  }, [diff.choiceIndex, behaviorsToShow, isCardPlay]);

  return (
    <div className="relative flex flex-col gap-1 py-2 px-3 hover:bg-white/5 rounded transition-colors border-b border-[rgba(100,200,255,0.2)] last:border-b-0">
      <div className="flex items-center gap-2">
        <span className="text-xs text-[#64c8ff] font-medium shrink-0">{playerName}</span>
        <span className="text-sm text-white truncate font-medium">{diff.source}</span>
        {isCardAction && (
          <span className="bg-[linear-gradient(135deg,rgba(100,200,255,0.3)_0%,rgba(80,160,220,0.4)_100%)] text-[#64c8ff] text-[8px] font-semibold uppercase tracking-[0.3px] py-0.5 px-1.5 rounded-lg border border-[rgba(100,200,255,0.4)] shrink-0">
            action
          </span>
        )}
        {isCardPlay && cardTags.length > 0 && (
          <div className="flex items-center gap-1 shrink-0">
            {cardTags.map((tag, i) => (
              <TagIcon key={i} tag={tag} />
            ))}
          </div>
        )}
        {isCardPlay && vpConditions.length > 0 && (
          <div className="shrink-0">
            <VictoryPointIcon vpConditions={vpConditions} />
          </div>
        )}
      </div>

      {choiceDisplayInfo.hasChoices && choiceDisplayInfo.type === "within-behavior" ? (
        // Single behavior with choices array - render each choice separately
        <div className="mt-1 flex flex-col gap-1">
          {choiceDisplayInfo.choices!.map((choice, choiceIndex) => {
            const isChosen = choiceIndex === diff.choiceIndex;
            // Create a synthetic behavior from the choice for display
            const syntheticBehavior = {
              ...behaviorsToShow[0],
              choices: undefined,
              inputs: choice.inputs,
              outputs: choice.outputs,
            };
            return (
              <div
                key={choiceIndex}
                className={`[&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none scale-90 origin-left ${!isChosen ? "opacity-40 grayscale" : ""}`}
              >
                <BehaviorSection behaviors={[syntheticBehavior]} greyOutAll={!isChosen} />
              </div>
            );
          })}
        </div>
      ) : choiceDisplayInfo.hasChoices && choiceDisplayInfo.type === "between-behaviors" ? (
        // Multiple behaviors - render each behavior with highlighting based on behavior index
        <div className="mt-1 flex flex-col gap-1">
          {behaviorsToShow.map((behavior, index) => {
            const isChosen = index === diff.choiceIndex;
            return (
              <div
                key={index}
                className={`[&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none scale-90 origin-left ${!isChosen ? "opacity-40 grayscale" : ""}`}
              >
                <BehaviorSection behaviors={[behavior]} greyOutAll={!isChosen} />
              </div>
            );
          })}
        </div>
      ) : (
        behaviorsToShow.length > 0 && (
          <div className="mt-1 [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none scale-90 origin-left">
            <BehaviorSection behaviors={behaviorsToShow} />
          </div>
        )
      )}

      {diff.calculatedOutputs && diff.calculatedOutputs.length > 0 && (
        <CalculatedOutputsDisplay
          outputs={diff.calculatedOutputs}
          showAll={!isCardSource || isCardAction}
          excludeBehaviors={isBehaviorSource}
        />
      )}

      {!displayData && !isCardSource && !isBehaviorSource && (
        <div className="text-xs text-gray-400">{diff.description}</div>
      )}
    </div>
  );
};

const LogPopover: React.FC<LogPopoverProps> = ({
  isVisible,
  onClose,
  anchorRef,
  gameId,
  gameState,
}) => {
  const [logs, setLogs] = useState<StateDiffDto[]>([]);
  const lastSequenceRef = useRef<number>(0);
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const playerNames = useMemo(() => {
    const names = new Map<string, string>();
    if (gameState?.currentPlayer) {
      names.set(gameState.currentPlayer.id, gameState.currentPlayer.name);
    }
    if (gameState?.otherPlayers) {
      gameState.otherPlayers.forEach((p) => {
        names.set(p.id, p.name);
      });
    }
    return names;
  }, [gameState?.currentPlayer, gameState?.otherPlayers]);

  // Handle incoming log updates via WebSocket
  const handleLogUpdate = useCallback((newLogs: StateDiffDto[]) => {
    setLogs((prev) => {
      // Deduplicate by sequence number
      const existingSeqs = new Set(prev.map((l) => l.sequenceNumber));
      const uniqueNewLogs = newLogs.filter((l) => !existingSeqs.has(l.sequenceNumber));
      if (uniqueNewLogs.length === 0) return prev;
      return [...prev, ...uniqueNewLogs];
    });
    if (newLogs.length > 0) {
      const maxSeq = Math.max(...newLogs.map((l) => l.sequenceNumber));
      if (maxSeq > lastSequenceRef.current) {
        lastSequenceRef.current = maxSeq;
      }
    }
  }, []);

  // Subscribe to WebSocket log updates
  useEffect(() => {
    globalWebSocketManager.on("log-update", handleLogUpdate);
    return () => {
      globalWebSocketManager.off("log-update", handleLogUpdate);
    };
  }, [handleLogUpdate]);

  // Clear logs when game changes, then request all logs from backend
  useEffect(() => {
    setLogs([]);
    lastSequenceRef.current = 0;
    if (gameId) {
      void globalWebSocketManager.requestLogs();
    }
  }, [gameId]);

  useEffect(() => {
    if (!isVisible || logs.length === 0) return;
    requestAnimationFrame(() => {
      const el = scrollContainerRef.current;
      if (el) {
        el.scrollTop = el.scrollHeight;
      }
    });
  }, [isVisible, logs]);

  const groupedLogs = useMemo(() => groupLogsByGeneration(logs), [logs]);

  const playerColorMap = useMemo(() => {
    const map = new Map<string, string>();
    if (gameState?.currentPlayer?.color) {
      map.set(gameState.currentPlayer.id, gameState.currentPlayer.color);
    }
    gameState?.otherPlayers?.forEach((p) => {
      if (p.color) map.set(p.id, p.color);
    });
    return map;
  }, [gameState]);

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="log"
      header={{
        title: "Game Log",
        badge: logs.length > 0 ? `${logs.length} entries` : undefined,
      }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={350}
      maxHeight={400}
      contentRef={scrollContainerRef}
    >
      {logs.length === 0 ? (
        <div className="flex items-center justify-center py-10 px-5">
          <span className="font-orbitron text-sm text-white/50">No logs</span>
        </div>
      ) : (
        <div className="flex flex-col">
          {groupedLogs.map((group) => {
            const totalEntries = group.playerTurns.reduce((sum, t) => sum + t.entries.length, 0);
            return (
              <div key={group.generation} className="flex flex-col">
                <GenerationDivider generation={group.generation} entryCount={totalEntries} />
                <div className="p-2 flex flex-col">
                  {group.playerTurns.map((turn, turnIndex) => (
                    <PlayerTurnSection
                      key={`${turn.playerId}-${turnIndex}`}
                      playerName={playerNames.get(turn.playerId) || "Unknown"}
                      entries={turn.entries}
                      playerColor={playerColorMap.get(turn.playerId)}
                      playerNames={playerNames}
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </GamePopover>
  );
};

export default LogPopover;
