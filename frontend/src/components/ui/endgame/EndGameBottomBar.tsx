import { FC, useState, useMemo, useEffect, useRef } from "react";
import type { GameDto, GameHistoryEntryDto } from "../../../types/generated/api-types";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { getPhaseDisplayName } from "../../../constants/gameConstants";
import { useVPCounting } from "../../../contexts/VPCountingContext";
import GameGraphs from "./GameGraphs.tsx";
import ReplayControls from "./ReplayControls.tsx";
import VPPhaseTabsOverlay from "./VPPhaseTabsOverlay.tsx";
import BackButton from "../buttons/BackButton.tsx";

const ANGLE_INDENT = 14;
const BUTTON_HEIGHT = 32;
const BUTTON_SPACING = 4;
const BORDER_COLOR = "rgba(60,60,70,0.7)";

interface ParaButtonConfig {
  id: string;
  label: string;
  accentColor: string;
  width: number;
  isActive?: boolean;
  visible?: boolean;
  onClick: () => void;
}

function ParallelogramButtonGroup({
  buttons,
  direction = "right",
}: {
  buttons: ParaButtonConfig[];
  direction?: "right" | "left" | "open";
}) {
  const visibleButtons = buttons.filter((b) => b.visible !== false);

  return (
    <div className="flex items-center">
      {visibleButtons.map((btn, idx) => (
        <ParaButton
          key={btn.id}
          config={btn}
          index={idx}
          total={visibleButtons.length}
          direction={direction}
        />
      ))}
    </div>
  );
}

function ParaButton({
  config,
  index,
  total,
  direction,
}: {
  config: ParaButtonConfig;
  index: number;
  total: number;
  direction: "right" | "left" | "open";
}) {
  const [isHovered, setIsHovered] = useState(false);

  const w = config.width;
  const h = BUTTON_HEIGHT;

  const isFlat =
    direction === "open" ? false : direction === "right" ? index === 0 : index === total - 1;

  let fillPoints: string;
  let accentLine: { x1: number; y1: number; x2: number; y2: number };
  let angledEdge: { x1: number; y1: number; x2: number; y2: number } | null;
  let secondAngledEdge: { x1: number; y1: number; x2: number; y2: number } | null = null;

  if (direction === "left") {
    fillPoints = isFlat
      ? `${ANGLE_INDENT},0 ${w},0 ${w},${h} 0,${h}`
      : `${ANGLE_INDENT},0 ${w},0 ${w - ANGLE_INDENT},${h} 0,${h}`;
    accentLine = { x1: ANGLE_INDENT, y1: 0, x2: w, y2: 0 };
    angledEdge = { x1: ANGLE_INDENT, y1: 0, x2: 0, y2: h };
    secondAngledEdge = isFlat ? null : { x1: w, y1: 0, x2: w - ANGLE_INDENT, y2: h };
  } else {
    fillPoints = isFlat
      ? `0,0 ${w - ANGLE_INDENT},0 ${w},${h} 0,${h}`
      : `0,0 ${w - ANGLE_INDENT},0 ${w},${h} ${ANGLE_INDENT},${h}`;
    accentLine = { x1: isFlat ? 0 : 0, y1: 0, x2: w - ANGLE_INDENT, y2: 0 };
    angledEdge = { x1: w - ANGLE_INDENT, y1: 0, x2: w, y2: h };
    secondAngledEdge = isFlat ? null : { x1: 0, y1: 0, x2: ANGLE_INDENT, y2: h };
  }

  const showAccent = isHovered || config.isActive;
  const topStrokeColor = showAccent ? config.accentColor : BORDER_COLOR;
  const topStrokeWidth = showAccent ? 3 : 1;

  return (
    <button
      onClick={config.onClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className="relative cursor-pointer"
      style={{
        width: w,
        height: h,
        marginLeft: index === 0 ? 0 : -ANGLE_INDENT + BUTTON_SPACING,
        zIndex: Z_INDEX.UI_BASE - index,
      }}
    >
      <svg
        className="absolute inset-0 w-full h-full"
        viewBox={`0 0 ${w} ${h}`}
        preserveAspectRatio="none"
      >
        <polygon
          points={fillPoints}
          fill={isHovered ? "rgba(20,20,25,0.95)" : "rgba(10,10,15,0.95)"}
        />
        <line
          x1={accentLine.x1}
          y1={accentLine.y1}
          x2={accentLine.x2}
          y2={accentLine.y2}
          stroke={topStrokeColor}
          strokeWidth={topStrokeWidth}
        />
        <line
          x1={angledEdge.x1}
          y1={angledEdge.y1}
          x2={angledEdge.x2}
          y2={angledEdge.y2}
          stroke={BORDER_COLOR}
          strokeWidth="1"
        />
        {secondAngledEdge && (
          <line
            x1={secondAngledEdge.x1}
            y1={secondAngledEdge.y1}
            x2={secondAngledEdge.x2}
            y2={secondAngledEdge.y2}
            stroke={BORDER_COLOR}
            strokeWidth="1"
          />
        )}
      </svg>
      <div
        className="absolute inset-0 flex items-center justify-center font-orbitron text-xs tracking-wider"
        style={{
          color: isHovered ? "#ffffff" : "rgba(255,255,255,0.8)",
          paddingLeft: ANGLE_INDENT / 2,
          paddingRight: ANGLE_INDENT / 2,
        }}
      >
        {config.label}
      </div>
    </button>
  );
}

interface EndGameBottomBarProps {
  game: GameDto;
  playerId: string;
  historyEntries?: GameHistoryEntryDto[];
  activePanel: "score" | "graphs" | "replay";
  onPanelChange?: (panel: "score" | "graphs" | "replay") => void;
  isReplayActive?: boolean;
  replayIndex?: number;
  replayTotal?: number;
  replayPlaying?: boolean;
  replaySpeed?: number;
  onReplayPlay?: () => void;
  onReplayPause?: () => void;
  onReplaySeek?: (index: number) => void;
  onReplayStepForward?: () => void;
  onReplayStepBackward?: () => void;
  onReplaySpeedChange?: (speed: number) => void;
  replaySpectatePlayerId?: string | null;
  onReplaySpectatePlayerChange?: (playerId: string | null) => void;
}

const EndGameBottomBar: FC<EndGameBottomBarProps> = ({
  game,
  playerId,
  historyEntries,
  activePanel,
  onPanelChange,
  isReplayActive,
  replayIndex,
  replayTotal,
  replayPlaying,
  replaySpeed,
  onReplayPlay,
  onReplayPause,
  onReplaySeek,
  onReplayStepForward,
  onReplayStepBackward,
  onReplaySpeedChange,
  replaySpectatePlayerId,
  onReplaySpectatePlayerChange,
}) => {
  const { state: vpState, controls: vpControls } = useVPCounting();

  useEffect(() => {
    if (activePanel !== "graphs" || !onPanelChange) {
      return;
    }
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        onPanelChange("score");
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [activePanel, onPanelChange]);

  const allScores = game.finalScores ?? [];
  const sortedScores = [...allScores].sort((a, b) => b.vpBreakdown.totalVP - a.vpBreakdown.totalVP);

  const maxVP = Math.max(...sortedScores.map((s) => s.vpBreakdown.totalVP), 1);
  const isCounting = vpState.isActive && !vpState.isComplete;

  const playerColors = useMemo(() => {
    const colors: Record<string, string> = {};
    if (game.currentPlayer) {
      colors[game.currentPlayer.id] = game.currentPlayer.color || "#ffffff";
    }
    for (const p of game.otherPlayers ?? []) {
      colors[p.id] = p.color || "#ffffff";
    }
    return colors;
  }, [game]);

  const playerNames = useMemo(() => {
    const names: Record<string, string> = {};
    if (game.currentPlayer) {
      names[game.currentPlayer.id] = game.currentPlayer.name;
    }
    for (const p of game.otherPlayers ?? []) {
      names[p.id] = p.name;
    }
    return names;
  }, [game]);

  const allPlayers = useMemo(() => {
    const players: { id: string; name: string }[] = [];
    if (game.currentPlayer) {
      players.push({ id: game.currentPlayer.id, name: game.currentPlayer.name });
    }
    for (const p of game.otherPlayers ?? []) {
      players.push({ id: p.id, name: p.name });
    }
    return players;
  }, [game]);

  const replayLabel = useMemo(() => {
    if (!isReplayActive || !historyEntries?.length) {
      return "";
    }
    const entry = historyEntries[replayIndex ?? 0];
    if (!entry) {
      return "";
    }
    return `Generation ${entry.generation} - ${getPhaseDisplayName(entry.phase)}`;
  }, [isReplayActive, historyEntries, replayIndex]);

  const generationMarkers = useMemo(() => {
    if (!historyEntries?.length) return [];
    const markers: { index: number; generation: number }[] = [];
    let lastGen = -1;
    for (let i = 0; i < historyEntries.length; i++) {
      if (historyEntries[i].generation !== lastGen) {
        lastGen = historyEntries[i].generation;
        markers.push({ index: i, generation: lastGen });
      }
    }
    return markers;
  }, [historyEntries]);

  const actionButtons: ParaButtonConfig[] = [
    {
      id: "replay-vp",
      label: "\u27F3",
      accentColor: "#ffffff",
      width: 50,
      visible: activePanel === "score" && vpState.isComplete && !isCounting,
      onClick: vpControls.start,
    },
    {
      id: "skip",
      label: "Skip",
      accentColor: "#ffffff",
      width: 80,
      visible: isCounting,
      onClick: vpControls.skip,
    },
  ];

  if (allScores.length === 0) {
    return (
      <div
        className="fixed bottom-0 left-0 right-0 p-4 flex items-center justify-center"
        style={{ zIndex: Z_INDEX.STANDARD_MODAL }}
      >
        <div className="text-center">
          <p className="text-red-400 font-orbitron mb-3">No scores available</p>
        </div>
      </div>
    );
  }

  return (
    <>
      {/* Graphs fullscreen overlay */}
      <div
        className={`fixed inset-0 bg-black/70 backdrop-blur-lg flex flex-col transition-opacity duration-500 ${
          activePanel === "graphs" ? "opacity-100" : "opacity-0 pointer-events-none"
        }`}
        style={{ zIndex: Z_INDEX.MENU_DROPDOWN }}
      >
        {onPanelChange && (
          <div className="flex-shrink-0 px-4 pt-4">
            <BackButton onClick={() => onPanelChange("score")}>Back to Score</BackButton>
          </div>
        )}
        {historyEntries && (
          <div className="flex-1 flex items-center justify-center min-h-0 px-[5%] pb-4">
            <div className="w-full h-full">
              <GameGraphs
                entries={historyEntries}
                playerColors={playerColors}
                playerNames={playerNames}
              />
            </div>
          </div>
        )}
      </div>

      {/* Replay controls overlay — blurred strip below the top menu bar */}
      <div
        className={`fixed top-[60px] max-lg:top-[50px] left-0 right-0 bg-black/50 flex justify-center transition-opacity duration-500 ${
          activePanel === "replay" ? "opacity-100" : "opacity-0 pointer-events-none"
        }`}
        style={{ paddingTop: 10, paddingBottom: 10, zIndex: Z_INDEX.MENU_DROPDOWN }}
      >
        <div className="w-[60%]">
          <ReplayControls
            currentIndex={replayIndex ?? 0}
            totalStates={replayTotal ?? 0}
            isPlaying={replayPlaying ?? false}
            playbackSpeed={replaySpeed ?? 1}
            currentLabel={replayLabel}
            onPlay={onReplayPlay ?? (() => {})}
            onPause={onReplayPause ?? (() => {})}
            onStepForward={onReplayStepForward ?? (() => {})}
            onStepBackward={onReplayStepBackward ?? (() => {})}
            onSeek={onReplaySeek ?? (() => {})}
            onSpeedChange={onReplaySpeedChange ?? (() => {})}
            generationMarkers={generationMarkers}
            rightSlot={
              <div className="flex items-center gap-2">
                <span className="text-white/50 text-xs font-orbitron">VIEW AS</span>
                {allPlayers.map((p) => (
                  <button
                    key={p.id}
                    onClick={() =>
                      onReplaySpectatePlayerChange?.(replaySpectatePlayerId === p.id ? null : p.id)
                    }
                    className="cursor-pointer transition-all duration-200"
                    style={{
                      width: replaySpectatePlayerId === p.id ? 18 : 14,
                      height: replaySpectatePlayerId === p.id ? 18 : 14,
                      borderRadius: "50%",
                      backgroundColor: playerColors[p.id] || "#ffffff",
                      opacity: replaySpectatePlayerId === p.id ? 1 : 0.5,
                      border: "none",
                      boxShadow:
                        replaySpectatePlayerId === p.id
                          ? `0 0 8px ${playerColors[p.id] || "#ffffff"}`
                          : "none",
                    }}
                    aria-label={p.name}
                  />
                ))}
              </div>
            }
          />
        </div>
      </div>

      {/* Score bottom bar with faded background */}
      <div
        className={`fixed bottom-0 left-0 right-0 transition-opacity duration-500 ${
          activePanel === "score" ? "opacity-100" : "opacity-0 pointer-events-none"
        }`}
        style={{ zIndex: Z_INDEX.STANDARD_MODAL }}
      >
        <div className="bg-black/50 backdrop-blur-sm" style={{ paddingTop: 16, paddingBottom: 12 }}>
          {/* Phase tabs */}
          <VPPhaseTabsOverlay
            phases={vpControls.phases}
            currentPhaseIndex={vpState.currentPhaseIndex}
            isActive={vpState.isActive}
          />

          {/* Score bars */}
          <div className="space-y-1.5" style={{ marginLeft: "20%", marginRight: "20%" }}>
            {sortedScores.map((score) => (
              <PlayerBar
                key={score.playerId}
                playerId={score.playerId}
                playerName={score.playerName}
                totalVP={score.vpBreakdown.totalVP}
                displayVP={
                  isCounting
                    ? (vpState.playerAccumulatedVP[score.playerId] ?? 0)
                    : score.vpBreakdown.totalVP
                }
                maxVP={maxVP}
                playerColor={playerColors[score.playerId] || "#ffffff"}
                isYou={score.playerId === playerId}
                isActivePlayer={isCounting && vpState.activePlayerId === score.playerId}
                isDimmed={isCounting && vpState.activePlayerId !== score.playerId}
              />
            ))}
          </div>

          {/* Controls row — only VP replay/skip buttons */}
          {actionButtons.some((b) => b.visible !== false) && (
            <div className="relative flex items-center px-4 pb-1 pt-1">
              <div className="ml-auto">
                <ParallelogramButtonGroup buttons={actionButtons} direction="left" />
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  );
};

const BAR_BASE_MS = 800;
const BAR_MS_PER_PERCENT = 30;

function PlayerBar({
  playerName,
  displayVP,
  maxVP,
  playerColor,
  isYou,
  isActivePlayer,
  isDimmed,
}: {
  playerId: string;
  playerName: string;
  totalVP: number;
  displayVP: number;
  maxVP: number;
  playerColor: string;
  isYou: boolean;
  isActivePlayer: boolean;
  isDimmed: boolean;
}) {
  const barWidth = maxVP > 0 ? (displayVP / maxVP) * 100 : 0;
  const prevWidthRef = useRef(barWidth);

  const widthDelta = Math.abs(barWidth - prevWidthRef.current);
  const isResetting = barWidth < prevWidthRef.current;
  const transitionMs = isResetting ? 400 : BAR_BASE_MS + widthDelta * BAR_MS_PER_PERCENT;

  useEffect(() => {
    prevWidthRef.current = barWidth;
  }, [barWidth]);

  const barRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [visualVP, setVisualVP] = useState(displayVP);
  const [visualLeft, setVisualLeft] = useState(0);
  const rafRef = useRef(0);

  useEffect(() => {
    const tick = () => {
      if (barRef.current && containerRef.current) {
        const barPx = barRef.current.getBoundingClientRect().width;
        const containerPx = containerRef.current.getBoundingClientRect().width;
        const visualPct = containerPx > 0 ? barPx / containerPx : 0;
        setVisualVP(Math.round(visualPct * maxVP));
        setVisualLeft(barPx);
      }
      rafRef.current = requestAnimationFrame(tick);
    };
    rafRef.current = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(rafRef.current);
  }, [maxVP]);

  return (
    <div className="flex items-center gap-3">
      <div
        className="text-base font-orbitron truncate shrink-0 transition-opacity duration-300"
        style={{ width: 120, color: "#ffffff", opacity: isDimmed ? 0.4 : 1 }}
      >
        {playerName}
        {isYou && <span className="text-white/30 text-xs ml-1">(you)</span>}
      </div>
      <div className="flex-1 h-2 relative" ref={containerRef}>
        <div
          ref={barRef}
          className="h-full ease-out"
          style={{
            width: `${Math.min(barWidth, 100)}%`,
            backgroundColor: playerColor,
            transform: "skewX(-10deg)",
            opacity: isDimmed ? 0.3 : 0.85,
            boxShadow: isActivePlayer ? `0 0 10px 2px ${playerColor}` : "none",
            transition: `width ${transitionMs}ms ease-out, opacity 300ms ease-out, box-shadow 300ms ease-out`,
          }}
        />
        <span
          className="absolute top-1/2 -translate-y-1/2 font-orbitron text-xs tabular-nums text-white/80 whitespace-nowrap transition-opacity duration-300"
          style={{
            left: visualLeft,
            paddingLeft: "4px",
            opacity: isDimmed ? 0.4 : 1,
          }}
        >
          {visualVP} VP
        </span>
      </div>
    </div>
  );
}

export default EndGameBottomBar;
