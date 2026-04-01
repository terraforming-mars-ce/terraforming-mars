import React, { useState, useRef, useEffect, useMemo } from "react";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext.tsx";
import {
  GameDto,
  GamePhaseAction,
  GamePhaseInitApplyCorp,
  GamePhaseInitApplyPrelude,
  GameStatusActive,
  PlayerDto,
  SpectatorDto,
} from "@/types/generated/api-types.ts";
import { StandardProject } from "@/types/cards.tsx";
import { Z_INDEX } from "@/constants/zIndex";
import { canPerformActions } from "@/utils/actionUtils.ts";
import GameHamburgerMenu from "../../ui/GameHamburgerMenu.tsx";
import TravelPopover from "../../ui/popover/TravelPopover.tsx";
import StandardProjectPopover from "../../ui/popover/StandardProjectPopover.tsx";
import MilestonePopover from "../../ui/popover/MilestonePopover.tsx";
import AwardPopover from "../../ui/popover/AwardPopover.tsx";
import ColonyPopover from "../../ui/popover/ColonyPopover.tsx";
import ProjectFundingPopover from "../../ui/popover/ProjectFundingPopover.tsx";
import { GamePopover } from "../../ui/GamePopover";
import { useHoverSound } from "@/hooks/useHoverSound.ts";
import { HamburgerIcon, EyeIcon } from "../../ui/menuIcons.tsx";

const ANGLE_INDENT = 20;
const BUTTON_SPACING = 6;
const BORDER_COLOR = "rgba(60,60,70,0.7)";
const HAMBURGER_WIDTH = 65;
const HAMBURGER_COLOR = "#ffffff";
const TRAVEL_WIDTH = 140;
const TRAVEL_COLOR = "#7eb8da";

const MilestoneAlertIndicator: React.FC<{ visible: boolean; top: number }> = ({ visible, top }) => (
  <div
    className={`absolute left-1/2 -translate-x-1/2 pointer-events-none transition-opacity duration-500 ${
      visible ? "opacity-100" : "opacity-0"
    }`}
    style={{ top }}
  >
    <span
      className="font-orbitron font-bold text-[11px] leading-none animate-[milestoneIndicatorPulse_2s_ease-in-out_infinite] flex items-center justify-center w-5 h-5 rounded-full border border-yellow-400/70"
      style={{ color: "#facc15", paddingTop: 1 }}
    >
      !
    </span>
  </div>
);

type EdgeStyle = "slope-left" | "slope-right" | "flat";

interface ParallelogramButtonProps {
  width: number;
  height: number;
  color: string;
  children: React.ReactNode;
  onClick: () => void;
  buttonRef: React.RefObject<HTMLButtonElement | null>;
  isActive?: boolean;
  leftEdge?: EdgeStyle;
  rightEdge?: EdgeStyle;
  className?: string;
  style?: React.CSSProperties;
}

const ParallelogramButton: React.FC<ParallelogramButtonProps> = ({
  width,
  height,
  color,
  children,
  onClick,
  buttonRef,
  isActive = false,
  leftEdge: left = "flat",
  rightEdge: right = "flat",
  className = "",
  style: extraStyle,
}) => {
  const [isHovered, setIsHovered] = useState(false);
  const hoverSound = useHoverSound();
  const active = isHovered || isActive;

  const ai = ANGLE_INDENT;
  const w = width;
  const h = height;

  // slope-left (\): top indented, bottom at edge
  // slope-right (/): top at edge, bottom indented
  const tl = left === "slope-left" ? ai : 0;
  const bl = left === "slope-right" ? ai : 0;
  const tr = right === "slope-left" ? w - ai : w;
  const br = right === "slope-right" ? w - ai : w;

  const fillPoints = `${tl},0 ${tr},0 ${br},${h} ${bl},${h}`;

  const edges: Array<{ x1: number; y1: number; x2: number; y2: number }> = [];
  if (left === "slope-left") {
    edges.push({ x1: ai, y1: 0, x2: 0, y2: h });
  } else if (left === "slope-right") {
    edges.push({ x1: 0, y1: 0, x2: ai, y2: h });
  }
  if (right === "slope-left") {
    edges.push({ x1: w - ai, y1: 0, x2: w, y2: h });
  } else if (right === "slope-right") {
    edges.push({ x1: w, y1: 0, x2: w - ai, y2: h });
  }

  return (
    <button
      ref={buttonRef}
      onClick={() => {
        hoverSound.onClick?.();
        onClick();
      }}
      onMouseEnter={() => {
        setIsHovered(true);
        hoverSound.onMouseEnter?.();
      }}
      onMouseLeave={() => setIsHovered(false)}
      className={`relative pointer-events-auto cursor-pointer outline-none ${className}`}
      style={{ width, height, ...extraStyle }}
    >
      <svg
        className="absolute inset-0 w-full h-full"
        viewBox={`0 0 ${w} ${h}`}
        preserveAspectRatio="none"
      >
        <polygon
          points={fillPoints}
          fill={active ? "rgba(20,20,25,0.95)" : "rgba(10,10,15,0.95)"}
        />
        <line
          x1={tl}
          y1={0}
          x2={tr}
          y2={0}
          stroke={active ? color : BORDER_COLOR}
          strokeWidth="3"
        />
        {edges.map((e, i) => (
          <line
            key={i}
            x1={e.x1}
            y1={e.y1}
            x2={e.x2}
            y2={e.y2}
            stroke={BORDER_COLOR}
            strokeWidth="2"
          />
        ))}
      </svg>
      <div className="relative z-10 h-full flex items-center justify-center px-4">
        <span
          className={`font-orbitron font-bold text-sm uppercase tracking-wider transition-colors duration-200 ${
            active ? "text-white" : "text-white/80"
          }`}
        >
          {children}
        </span>
      </div>
    </button>
  );
};

const ENDGAME_ACCENT = "#3b82f6";

function EndgameTabButton({
  label,
  width,
  height,
  isFirst,
  isActive,
  onClick,
}: {
  label: string;
  width: number;
  height: number;
  isFirst: boolean;
  isActive: boolean;
  onClick: () => void;
}) {
  const [isHovered, setIsHovered] = useState(false);
  const hoverSound = useHoverSound();

  // "Left" direction: angled edges mirror the right-side style
  const fillPoints = isFirst
    ? `${ANGLE_INDENT},0 ${width},0 ${width - ANGLE_INDENT},${height} 0,${height}`
    : `${ANGLE_INDENT},0 ${width},0 ${width - ANGLE_INDENT},${height} 0,${height}`;

  const showAccent = isActive || isHovered;

  return (
    <button
      onClick={() => {
        hoverSound.onClick?.();
        onClick();
      }}
      onMouseEnter={() => {
        setIsHovered(true);
        hoverSound.onMouseEnter?.();
      }}
      onMouseLeave={() => setIsHovered(false)}
      className="relative pointer-events-auto cursor-pointer outline-none"
      style={{
        width,
        height,
        marginRight: -ANGLE_INDENT + BUTTON_SPACING,
      }}
    >
      <svg
        className="absolute inset-0 w-full h-full"
        viewBox={`0 0 ${width} ${height}`}
        preserveAspectRatio="none"
      >
        <polygon
          points={fillPoints}
          fill={isHovered ? "rgba(20,20,25,0.95)" : "rgba(10,10,15,0.95)"}
        />
        <line
          x1={ANGLE_INDENT}
          y1={0}
          x2={width}
          y2={0}
          stroke={showAccent ? ENDGAME_ACCENT : BORDER_COLOR}
          strokeWidth="3"
        />
        <line
          x1={width}
          y1={0}
          x2={width - ANGLE_INDENT}
          y2={height}
          stroke={BORDER_COLOR}
          strokeWidth="2"
        />
        <line x1={ANGLE_INDENT} y1={0} x2={0} y2={height} stroke={BORDER_COLOR} strokeWidth="2" />
      </svg>
      <div className="relative z-10 h-full flex items-center justify-center px-4">
        <span
          className={`font-orbitron font-bold text-sm uppercase tracking-wider transition-colors duration-200 ${
            showAccent ? "text-white" : "text-white/60"
          }`}
        >
          {label}
        </span>
      </div>
    </button>
  );
}

interface EndgamePanelButton {
  id: "score" | "graphs" | "replay";
  label: string;
  width: number;
}

const ENDGAME_BUTTONS: EndgamePanelButton[] = [
  { id: "score", label: "SCORE", width: 120 },
  { id: "graphs", label: "GRAPHS", width: 130 },
  { id: "replay", label: "REPLAY", width: 130 },
];

interface TopMenuBarProps {
  gameState: GameDto;
  currentPlayer?: PlayerDto | null;
  onStandardProjectSelect?: (project: StandardProject) => void;
  onLeaveGame?: () => void;
  onEndGame?: () => void;
  gameId?: string;
  isEndgame?: boolean;
  activeEndgamePanel?: "score" | "graphs" | "replay";
  onEndgamePanelChange?: (panel: "score" | "graphs" | "replay") => void;
  hasHistory?: boolean;
}

const TopMenuBar: React.FC<TopMenuBarProps> = ({
  gameState,
  currentPlayer,
  onStandardProjectSelect,
  onLeaveGame,
  onEndGame,
  gameId,
  isEndgame = false,
  activeEndgamePanel,
  onEndgamePanelChange,
  hasHistory = false,
}) => {
  const { activePlanet } = usePlanetFocus();
  const [menuOpen, setMenuOpen] = useState(false);
  const [showTravelPopover, setShowTravelPopover] = useState(false);
  const travelButtonRef = useRef<HTMLButtonElement>(null);

  const [spectatorsOpen, setSpectatorsOpen] = useState(false);
  const hamburgerButtonRef = useRef<HTMLButtonElement>(null);
  const eyeButtonRef = useRef<HTMLButtonElement>(null);

  const [showStandardProjectsPopover, setShowStandardProjectsPopover] = useState(false);
  const [showMilestonePopover, setShowMilestonePopover] = useState(false);
  const [showAwardPopover, setShowAwardPopover] = useState(false);
  const [showColonyPopover, setShowColonyPopover] = useState(false);
  const [showProjectFundingPopover, setShowProjectFundingPopover] = useState(false);
  const standardProjectsButtonRef = useRef<HTMLButtonElement>(null);
  const milestonesButtonRef = useRef<HTMLButtonElement>(null);
  const awardsButtonRef = useRef<HTMLButtonElement>(null);
  const coloniesButtonRef = useRef<HTMLButtonElement>(null);
  const projectFundingButtonRef = useRef<HTMLButtonElement>(null);

  const handleStandardProjectSelect = (project: StandardProject) => {
    setShowStandardProjectsPopover(false);
    onStandardProjectSelect?.(project);
  };

  const isInitPhase =
    gameState?.currentPhase === GamePhaseInitApplyCorp ||
    gameState?.currentPhase === GamePhaseInitApplyPrelude;

  const hasColonies = (gameState?.colonies?.length ?? 0) > 0;
  const hasProjectFunding = (gameState?.projectFunding?.length ?? 0) > 0;

  const hasEligibleMilestones = useMemo(() => {
    const isGameActive = gameState?.status === GameStatusActive;
    const isActionPhase = gameState?.currentPhase === GamePhaseAction;
    const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
    if (!isGameActive || !isActionPhase || !isCurrentPlayerTurn || !canPerformActions(gameState)) {
      return false;
    }
    const milestones = gameState?.currentPlayer?.milestones ?? [];
    return milestones.some(
      (m) => !m.isClaimed && !(m.errors ?? []).some((e) => e.category === "requirement"),
    );
  }, [gameState]);

  const menuItems: { id: string; label: string; color: string }[] = [
    { id: "projects", label: "STANDARD PROJECTS", color: "#4a90e2" },
    { id: "milestones", label: "MILESTONES", color: "#ff6b35" },
    { id: "awards", label: "AWARDS", color: "#f39c12" },
    ...(hasColonies ? [{ id: "colonies", label: "COLONIES", color: "#7c6fc4" }] : []),
    ...(hasProjectFunding ? [{ id: "funding", label: "FUNDING", color: "#10b981" }] : []),
  ];

  const handleTabClick = (tabId: string) => {
    if (currentPlayer?.pendingTileSelection) return;

    if (tabId === "projects") {
      setShowStandardProjectsPopover((prev) => !prev);
    } else if (tabId === "milestones") {
      setShowMilestonePopover((prev) => !prev);
    } else if (tabId === "awards") {
      setShowAwardPopover((prev) => !prev);
    } else if (tabId === "colonies") {
      setShowColonyPopover((prev) => !prev);
    } else if (tabId === "funding") {
      setShowProjectFundingPopover((prev) => !prev);
    }
  };

  const getButtonRef = (itemId: string) => {
    if (itemId === "projects") return standardProjectsButtonRef;
    if (itemId === "milestones") return milestonesButtonRef;
    if (itemId === "awards") return awardsButtonRef;
    if (itemId === "colonies") return coloniesButtonRef;
    if (itemId === "funding") return projectFundingButtonRef;
    return null;
  };

  const calcScale = () => Math.min(1, Math.max(0.75, window.innerWidth / 2200));
  const [topBarScale, setTopBarScale] = useState(calcScale);

  useEffect(() => {
    const handleResize = () => setTopBarScale(calcScale());
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const spectators: SpectatorDto[] = gameState?.spectators || [];
  const baseWidths = [250, 190, 160];
  let buttonWidths = [...baseWidths];
  if (hasColonies) {
    buttonWidths.push(170);
  }
  if (hasProjectFunding) {
    buttonWidths.push(160);
  }
  const buttonHeight = 40;
  const EYE_WIDTH = 68;

  return (
    <>
      <div
        className="bg-transparent relative pointer-events-none"
        style={{ zIndex: Z_INDEX.TOP_MENU_BAR }}
      >
        <div className="flex justify-between items-center px-2 h-[60px] max-lg:h-[50px] max-md:flex-wrap">
          <div
            className={`flex max-md:order-2 max-md:flex-[0_0_100%] max-md:mt-2.5 origin-top-left transition-opacity duration-500 ease-in-out ${activePlanet === "solar-system" ? "opacity-0 pointer-events-none" : "opacity-100"}`}
            style={{ transform: `scale(${topBarScale})` }}
          >
            {!isInitPhase &&
              !isEndgame &&
              menuItems.map((item, index) => (
                <div key={item.id} className="relative">
                  <ParallelogramButton
                    width={buttonWidths[index]}
                    height={buttonHeight}
                    color={item.color}
                    onClick={() => handleTabClick(item.id)}
                    buttonRef={getButtonRef(item.id) as React.RefObject<HTMLButtonElement | null>}
                    leftEdge={index === 0 ? "flat" : "slope-right"}
                    rightEdge="slope-left"
                    style={{
                      marginLeft: index === 0 ? 0 : -ANGLE_INDENT + BUTTON_SPACING,
                      zIndex: Z_INDEX.UI_BASE - index,
                    }}
                    isActive={
                      (item.id === "projects" && showStandardProjectsPopover) ||
                      (item.id === "milestones" && showMilestonePopover) ||
                      (item.id === "awards" && showAwardPopover) ||
                      (item.id === "colonies" && showColonyPopover) ||
                      (item.id === "funding" && showProjectFundingPopover)
                    }
                  >
                    {item.label}
                  </ParallelogramButton>
                  {item.id === "milestones" && (
                    <MilestoneAlertIndicator
                      visible={hasEligibleMilestones && !showMilestonePopover}
                      top={buttonHeight + 6}
                    />
                  )}
                </div>
              ))}
          </div>

          <div style={{ marginRight: HAMBURGER_WIDTH * topBarScale }}>
            <div
              className="origin-top-right flex items-center"
              style={{ transform: `scale(${topBarScale})` }}
            >
              {isEndgame && onEndgamePanelChange && (
                <div className="flex items-center pointer-events-auto">
                  {ENDGAME_BUTTONS.filter(
                    (btn) => (btn.id !== "graphs" && btn.id !== "replay") || hasHistory,
                  ).map((btn, idx) => {
                    const isActive = activeEndgamePanel === btn.id;
                    return (
                      <EndgameTabButton
                        key={btn.id}
                        label={btn.label}
                        width={btn.width}
                        height={buttonHeight}
                        isFirst={idx === 0}
                        isActive={isActive}
                        onClick={() => onEndgamePanelChange(btn.id)}
                      />
                    );
                  })}
                </div>
              )}
              {spectators.length > 0 && (
                <ParallelogramButton
                  width={EYE_WIDTH}
                  height={buttonHeight}
                  color="#7eb8da"
                  onClick={() => setSpectatorsOpen(!spectatorsOpen)}
                  buttonRef={eyeButtonRef}
                  isActive={spectatorsOpen}
                  leftEdge="slope-left"
                  rightEdge="slope-right"
                  style={{ marginRight: -ANGLE_INDENT + BUTTON_SPACING }}
                >
                  <EyeIcon />
                </ParallelogramButton>
              )}
              <ParallelogramButton
                width={TRAVEL_WIDTH}
                height={buttonHeight}
                color={TRAVEL_COLOR}
                onClick={() => setShowTravelPopover((prev) => !prev)}
                buttonRef={travelButtonRef}
                isActive={showTravelPopover}
                leftEdge="slope-left"
                rightEdge="slope-right"
                style={{ marginRight: -ANGLE_INDENT + BUTTON_SPACING }}
              >
                TRAVEL
              </ParallelogramButton>
            </div>
          </div>
        </div>

        <TravelPopover
          isVisible={showTravelPopover}
          onClose={() => setShowTravelPopover(false)}
          anchorRef={travelButtonRef}
        />

        <StandardProjectPopover
          isVisible={showStandardProjectsPopover}
          onClose={() => setShowStandardProjectsPopover(false)}
          onProjectSelect={handleStandardProjectSelect}
          gameState={gameState}
          anchorRef={standardProjectsButtonRef}
        />

        <MilestonePopover
          isVisible={showMilestonePopover}
          onClose={() => setShowMilestonePopover(false)}
          gameState={gameState}
          anchorRef={milestonesButtonRef}
        />

        <AwardPopover
          isVisible={showAwardPopover}
          onClose={() => setShowAwardPopover(false)}
          gameState={gameState}
          anchorRef={awardsButtonRef}
        />

        {hasColonies && (
          <ColonyPopover
            isVisible={showColonyPopover}
            onClose={() => setShowColonyPopover(false)}
            gameState={gameState}
            anchorRef={coloniesButtonRef}
          />
        )}

        {hasProjectFunding && (
          <ProjectFundingPopover
            isVisible={showProjectFundingPopover}
            onClose={() => setShowProjectFundingPopover(false)}
            gameState={gameState}
            anchorRef={projectFundingButtonRef}
          />
        )}

        {spectators.length > 0 && (
          <GamePopover
            isVisible={spectatorsOpen}
            onClose={() => setSpectatorsOpen(false)}
            position={{
              type: "anchor",
              anchorRef: eyeButtonRef,
              placement: "below",
            }}
            theme="menu"
            width={180}
            maxHeight="auto"
            animation="slideDown"
            excludeRef={eyeButtonRef}
            zIndex={Z_INDEX.TOP_MENU_ALWAYS_ON_TOP + 1}
          >
            <div className="py-2 px-3">
              <div className="text-white/40 text-[10px] font-orbitron font-bold uppercase tracking-wider mb-2">
                Spectators ({spectators.length})
              </div>
              <div className="flex flex-col gap-1.5">
                {spectators.map((s) => (
                  <div key={s.id} className="flex items-center gap-2">
                    <span
                      className="w-2 h-2 rounded-full shrink-0"
                      style={{ backgroundColor: s.color }}
                    />
                    <span className="text-white/80 text-sm truncate">{s.name}</span>
                  </div>
                ))}
              </div>
            </div>
          </GamePopover>
        )}
      </div>
      <div
        className="fixed right-0 px-2 h-[60px] max-lg:h-[50px] flex items-center pointer-events-none origin-top-right"
        style={{
          zIndex: Z_INDEX.TOP_MENU_ALWAYS_ON_TOP,
          transform: `scale(${topBarScale})`,
          top: topBarScale < 1 ? 1 : 0,
        }}
      >
        <ParallelogramButton
          width={HAMBURGER_WIDTH}
          height={buttonHeight}
          color={HAMBURGER_COLOR}
          onClick={() => setMenuOpen(!menuOpen)}
          buttonRef={hamburgerButtonRef}
          isActive={menuOpen}
          leftEdge="slope-left"
          rightEdge="flat"
        >
          <HamburgerIcon />
        </ParallelogramButton>
      </div>
      <GameHamburgerMenu
        isOpen={menuOpen}
        onClose={() => setMenuOpen(false)}
        anchorRef={hamburgerButtonRef}
        gameId={gameId}
        isHost={currentPlayer?.id === gameState.hostPlayerId}
        onLeaveGame={onLeaveGame}
        onEndGame={currentPlayer?.id === gameState.hostPlayerId ? onEndGame : undefined}
      />
    </>
  );
};

export default TopMenuBar;
