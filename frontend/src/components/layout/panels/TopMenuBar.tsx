import React, { useState, useRef, useEffect, useCallback } from "react";
import {
  GameDto,
  GamePhaseInitApplyCorp,
  GamePhaseInitApplyPrelude,
  PlayerDto,
  SpectatorDto,
} from "@/types/generated/api-types.ts";
import { StandardProject } from "@/types/cards.tsx";
import { Z_INDEX } from "@/constants/zIndex";
import SoundToggleButton from "../../ui/buttons/SoundToggleButton.tsx";
import StandardProjectPopover from "../../ui/popover/StandardProjectPopover.tsx";
import MilestonePopover from "../../ui/popover/MilestonePopover.tsx";
import AwardPopover from "../../ui/popover/AwardPopover.tsx";
import ColonyPopover from "../../ui/popover/ColonyPopover.tsx";
import ProjectFundingPopover from "../../ui/popover/ProjectFundingPopover.tsx";
import { GamePopover } from "../../ui/GamePopover";
import { useHoverSound } from "@/hooks/useHoverSound.ts";
import {
  MenuPopoverItem,
  MenuPopoverDivider,
  MenuPopoverVersion,
} from "../../ui/MenuPopoverItem.tsx";
import {
  CopyIcon,
  FullscreenIcon,
  ExitFullscreenIcon,
  PerformanceIcon,
  FeedbackIcon,
  LeaveIcon,
  EndGameIcon,
} from "../../ui/menuIcons.tsx";

const ANGLE_INDENT = 20;
const BUTTON_SPACING = 6;
const BORDER_COLOR = "rgba(60,60,70,0.7)";
const HAMBURGER_WIDTH = 56;
const HAMBURGER_COLOR = "#ffffff";

interface ParallelogramButtonProps {
  index: number;
  total: number;
  width: number;
  height: number;
  color: string;
  children: React.ReactNode;
  onClick: () => void;
  buttonRef: React.RefObject<HTMLButtonElement | null>;
  isActive?: boolean;
}

const ParallelogramButton: React.FC<ParallelogramButtonProps> = ({
  index,
  width,
  height,
  color,
  children,
  onClick,
  buttonRef,
  isActive = false,
}) => {
  const [isHovered, setIsHovered] = useState(false);
  const hoverSound = useHoverSound();
  const isFirst = index === 0;

  // For parallelogram: both angled edges slant the same direction (\)
  // First button: flat left, angled right
  // Others: angled left (top-left to bottom-right), angled right
  const fillPoints = isFirst
    ? `0,0 ${width - ANGLE_INDENT},0 ${width},${height} 0,${height}`
    : `0,0 ${width - ANGLE_INDENT},0 ${width},${height} ${ANGLE_INDENT},${height}`;

  const topEdge = isFirst
    ? { x1: 0, y1: 0, x2: width - ANGLE_INDENT, y2: 0 }
    : { x1: 0, y1: 0, x2: width - ANGLE_INDENT, y2: 0 };

  const rightEdge = { x1: width - ANGLE_INDENT, y1: 0, x2: width, y2: height };

  // Left edge for non-first: from (0, 0) to (ANGLE_INDENT, height) - same slant direction as right
  const leftEdge = isFirst ? null : { x1: 0, y1: 0, x2: ANGLE_INDENT, y2: height };

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
      className="relative pointer-events-auto cursor-pointer outline-none"
      style={{
        width,
        height,
        marginLeft: isFirst ? 0 : -ANGLE_INDENT + BUTTON_SPACING,
        zIndex: 10 - index,
      }}
    >
      <svg
        className="absolute inset-0 w-full h-full"
        viewBox={`0 0 ${width} ${height}`}
        preserveAspectRatio="none"
      >
        <polygon
          points={fillPoints}
          fill={isHovered || isActive ? "rgba(20,20,25,0.95)" : "rgba(10,10,15,0.95)"}
        />
        <line
          x1={topEdge.x1}
          y1={topEdge.y1}
          x2={topEdge.x2}
          y2={topEdge.y2}
          stroke={isHovered || isActive ? color : BORDER_COLOR}
          strokeWidth="3"
        />
        <line
          x1={rightEdge.x1}
          y1={rightEdge.y1}
          x2={rightEdge.x2}
          y2={rightEdge.y2}
          stroke={BORDER_COLOR}
          strokeWidth="2"
        />
        {leftEdge && (
          <line
            x1={leftEdge.x1}
            y1={leftEdge.y1}
            x2={leftEdge.x2}
            y2={leftEdge.y2}
            stroke={BORDER_COLOR}
            strokeWidth="2"
          />
        )}
      </svg>
      <div className="relative z-10 h-full flex items-center justify-center px-4">
        <span
          className={`font-orbitron font-bold text-sm uppercase tracking-wider transition-colors duration-200 ${
            isHovered || isActive ? "text-white" : "text-white/80"
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
  const [menuOpen, setMenuOpen] = useState(false);
  const [hamburgerHovered, setHamburgerHovered] = useState(false);
  const [eyeHovered, setEyeHovered] = useState(false);
  const [spectatorsOpen, setSpectatorsOpen] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(!!document.fullscreenElement);
  const hamburgerButtonRef = useRef<HTMLButtonElement>(null);
  const eyeButtonRef = useRef<HTMLButtonElement>(null);
  const menuItemHover = useHoverSound();

  useEffect(() => {
    const handleFullscreenChange = () => setIsFullscreen(!!document.fullscreenElement);
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () => document.removeEventListener("fullscreenchange", handleFullscreenChange);
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setMenuOpen((prev) => !prev);
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  const handleToggleFullscreen = useCallback(() => {
    if (isFullscreen) {
      void document.exitFullscreen();
    } else {
      void document.documentElement.requestFullscreen();
    }
    setMenuOpen(false);
  }, [isFullscreen]);

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

  const handleCopyGameLink = async () => {
    if (gameId) {
      const url = `${window.location.origin}/game/${gameId}`;
      await navigator.clipboard.writeText(url);
      setMenuOpen(false);
    }
  };

  const handleLeaveGame = () => {
    setMenuOpen(false);
    onLeaveGame?.();
  };

  const handleEndGame = () => {
    setMenuOpen(false);
    onEndGame?.();
  };

  const handleStandardProjectSelect = (project: StandardProject) => {
    setShowStandardProjectsPopover(false);
    onStandardProjectSelect?.(project);
  };

  const isInitPhase =
    gameState?.currentPhase === GamePhaseInitApplyCorp ||
    gameState?.currentPhase === GamePhaseInitApplyPrelude;

  const hasColonies = (gameState?.colonyTiles?.length ?? 0) > 0;
  const hasProjectFunding = (gameState?.projectFunding?.length ?? 0) > 0;

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
    <div
      className="bg-transparent relative pointer-events-none"
      style={{ zIndex: Z_INDEX.TOP_MENU_ALWAYS_ON_TOP }}
    >
      <div className="flex justify-between items-center px-2 h-[60px] max-lg:h-[50px] max-md:flex-wrap">
        <div
          className="flex max-md:order-2 max-md:flex-[0_0_100%] max-md:mt-2.5 origin-top-left"
          style={{ transform: `scale(${topBarScale})` }}
        >
          {!isInitPhase &&
            !isEndgame &&
            menuItems.map((item, index) => (
              <ParallelogramButton
                key={item.id}
                index={index}
                total={menuItems.length}
                width={buttonWidths[index]}
                height={buttonHeight}
                color={item.color}
                onClick={() => handleTabClick(item.id)}
                buttonRef={getButtonRef(item.id) as React.RefObject<HTMLButtonElement | null>}
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
            ))}
        </div>

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
            <button
              ref={eyeButtonRef}
              onClick={() => setSpectatorsOpen(!spectatorsOpen)}
              onMouseEnter={() => setEyeHovered(true)}
              onMouseLeave={() => setEyeHovered(false)}
              aria-label="Spectators"
              className="relative pointer-events-auto cursor-pointer outline-none"
              style={{
                width: EYE_WIDTH,
                height: buttonHeight,
                marginRight: -ANGLE_INDENT + BUTTON_SPACING,
              }}
            >
              <svg
                className="absolute inset-0 w-full h-full"
                viewBox={`0 0 ${EYE_WIDTH} ${buttonHeight}`}
                preserveAspectRatio="none"
              >
                <polygon
                  points={`${ANGLE_INDENT},0 ${EYE_WIDTH},0 ${EYE_WIDTH - ANGLE_INDENT},${buttonHeight} 0,${buttonHeight}`}
                  fill={eyeHovered ? "rgba(20,20,25,0.95)" : "rgba(10,10,15,0.95)"}
                />
                <line
                  x1={ANGLE_INDENT}
                  y1={0}
                  x2={EYE_WIDTH}
                  y2={0}
                  stroke={eyeHovered ? "#7eb8da" : BORDER_COLOR}
                  strokeWidth="3"
                />
                <line
                  x1={EYE_WIDTH}
                  y1={0}
                  x2={EYE_WIDTH - ANGLE_INDENT}
                  y2={buttonHeight}
                  stroke={BORDER_COLOR}
                  strokeWidth="2"
                />
                <line
                  x1={ANGLE_INDENT}
                  y1={0}
                  x2={0}
                  y2={buttonHeight}
                  stroke={BORDER_COLOR}
                  strokeWidth="2"
                />
              </svg>
              <div className="relative z-10 h-full flex items-center justify-center">
                <svg
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke={eyeHovered ? "white" : "rgba(255,255,255,0.8)"}
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
              </div>
            </button>
          )}
          <button
            ref={hamburgerButtonRef}
            onClick={() => setMenuOpen(!menuOpen)}
            onMouseEnter={() => setHamburgerHovered(true)}
            onMouseLeave={() => setHamburgerHovered(false)}
            aria-label="Menu"
            data-overlay-layer
            className="relative pointer-events-auto cursor-pointer outline-none"
            style={{ width: HAMBURGER_WIDTH, height: buttonHeight }}
          >
            <svg
              className="absolute inset-0 w-full h-full"
              viewBox={`0 0 ${HAMBURGER_WIDTH} ${buttonHeight}`}
              preserveAspectRatio="none"
            >
              <polygon
                points={`${ANGLE_INDENT},0 ${HAMBURGER_WIDTH},0 ${HAMBURGER_WIDTH},${buttonHeight} 0,${buttonHeight}`}
                fill={hamburgerHovered ? "rgba(20,20,25,0.95)" : "rgba(10,10,15,0.95)"}
              />
              <line
                x1={ANGLE_INDENT}
                y1={0}
                x2={HAMBURGER_WIDTH}
                y2={0}
                stroke={hamburgerHovered ? HAMBURGER_COLOR : BORDER_COLOR}
                strokeWidth="3"
              />
              <line
                x1={ANGLE_INDENT}
                y1={0}
                x2={0}
                y2={buttonHeight}
                stroke={BORDER_COLOR}
                strokeWidth="2"
              />
            </svg>
            <div
              className="relative z-10 h-full flex items-center justify-center"
              style={{ paddingLeft: ANGLE_INDENT / 2 }}
            >
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke={hamburgerHovered ? "white" : "rgba(255,255,255,0.8)"}
                strokeWidth="2"
                strokeLinecap="round"
              >
                <line x1="3" y1="6" x2="21" y2="6" />
                <line x1="3" y1="12" x2="21" y2="12" />
                <line x1="3" y1="18" x2="21" y2="18" />
              </svg>
            </div>
          </button>
        </div>

        <GamePopover
          isVisible={menuOpen}
          onClose={() => setMenuOpen(false)}
          position={{
            type: "anchor",
            anchorRef: hamburgerButtonRef,
            placement: "below",
          }}
          theme="menu"
          width={200}
          maxHeight="auto"
          animation="slideDown"
          excludeRef={hamburgerButtonRef}
          zIndex={Z_INDEX.TOP_MENU_ALWAYS_ON_TOP + 1}
          overlayLayer
        >
          <div className="py-1">
            <MenuPopoverItem
              icon={<CopyIcon />}
              label="Copy game link"
              onClick={() => {
                menuItemHover.onClick?.();
                void handleCopyGameLink();
              }}
              onMouseEnter={menuItemHover.onMouseEnter}
            />
            <MenuPopoverDivider />
            <SoundToggleButton />
            <MenuPopoverDivider />
            <MenuPopoverItem
              icon={isFullscreen ? <ExitFullscreenIcon /> : <FullscreenIcon />}
              label={isFullscreen ? "Exit Fullscreen" : "Fullscreen"}
              onClick={handleToggleFullscreen}
            />
            <MenuPopoverDivider />
            <MenuPopoverItem
              icon={<PerformanceIcon />}
              label="Performance"
              onClick={() => {
                menuItemHover.onClick?.();
                setMenuOpen(false);
                window.dispatchEvent(new CustomEvent("toggle-performance-window"));
              }}
              onMouseEnter={menuItemHover.onMouseEnter}
            />
            <MenuPopoverDivider />
            <MenuPopoverItem
              icon={<FeedbackIcon />}
              label="Feedback"
              onClick={() => {
                menuItemHover.onClick?.();
                setMenuOpen(false);
                window.dispatchEvent(new CustomEvent("toggle-feedback-window"));
              }}
              onMouseEnter={menuItemHover.onMouseEnter}
            />
            <MenuPopoverDivider />
            <MenuPopoverItem
              icon={<LeaveIcon />}
              label="Leave game"
              variant="danger"
              onClick={() => {
                menuItemHover.onClick?.();
                handleLeaveGame();
              }}
              onMouseEnter={menuItemHover.onMouseEnter}
            />
            {currentPlayer?.id === gameState.hostPlayerId && (
              <>
                <MenuPopoverDivider />
                <MenuPopoverItem
                  icon={<EndGameIcon />}
                  label="End game"
                  variant="danger"
                  onClick={() => {
                    menuItemHover.onClick?.();
                    handleEndGame();
                  }}
                  onMouseEnter={menuItemHover.onMouseEnter}
                />
              </>
            )}
            <MenuPopoverDivider />
            <MenuPopoverVersion />
          </div>
        </GamePopover>
      </div>

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
  );
};

export default TopMenuBar;
