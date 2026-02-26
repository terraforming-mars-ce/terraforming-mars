import React, { useState, useRef, useEffect, useCallback } from "react";
import { GameDto, PlayerDto } from "@/types/generated/api-types.ts";
import { StandardProject } from "@/types/cards.tsx";
import { Z_INDEX } from "@/constants/zIndex";
import SoundToggleButton from "../../ui/buttons/SoundToggleButton.tsx";
import StandardProjectPopover from "../../ui/popover/StandardProjectPopover.tsx";
import MilestonePopover from "../../ui/popover/MilestonePopover.tsx";
import AwardPopover from "../../ui/popover/AwardPopover.tsx";
import { GamePopover } from "../../ui/GamePopover";
import { useHoverSound } from "@/hooks/useHoverSound.ts";

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
}

const ParallelogramButton: React.FC<ParallelogramButtonProps> = ({
  index,
  width,
  height,
  color,
  children,
  onClick,
  buttonRef,
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
  const leftEdge = isFirst
    ? null
    : { x1: 0, y1: 0, x2: ANGLE_INDENT, y2: height };

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
      className="relative pointer-events-auto cursor-pointer"
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
          fill={isHovered ? "rgba(20,20,25,0.95)" : "rgba(10,10,15,0.95)"}
        />
        <line
          x1={topEdge.x1}
          y1={topEdge.y1}
          x2={topEdge.x2}
          y2={topEdge.y2}
          stroke={isHovered ? color : BORDER_COLOR}
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
            isHovered ? "text-white" : "text-white/80"
          }`}
        >
          {children}
        </span>
      </div>
    </button>
  );
};

interface TopMenuBarProps {
  gameState: GameDto;
  currentPlayer?: PlayerDto | null;
  onStandardProjectSelect?: (project: StandardProject) => void;
  onLeaveGame?: () => void;
  gameId?: string;
}

const TopMenuBar: React.FC<TopMenuBarProps> = ({
  gameState,
  currentPlayer,
  onStandardProjectSelect,
  onLeaveGame,
  gameId,
}) => {
  const [menuOpen, setMenuOpen] = useState(false);
  const [hamburgerHovered, setHamburgerHovered] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(
    !!document.fullscreenElement,
  );
  const hamburgerButtonRef = useRef<HTMLButtonElement>(null);
  const menuItemHover = useHoverSound();

  useEffect(() => {
    const handleFullscreenChange = () =>
      setIsFullscreen(!!document.fullscreenElement);
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () =>
      document.removeEventListener("fullscreenchange", handleFullscreenChange);
  }, []);

  const handleToggleFullscreen = useCallback(() => {
    if (isFullscreen) {
      void document.exitFullscreen();
    } else {
      void document.documentElement.requestFullscreen();
    }
    setMenuOpen(false);
  }, [isFullscreen]);

  const [showStandardProjectsPopover, setShowStandardProjectsPopover] =
    useState(false);
  const [showMilestonePopover, setShowMilestonePopover] = useState(false);
  const [showAwardPopover, setShowAwardPopover] = useState(false);
  const standardProjectsButtonRef = useRef<HTMLButtonElement>(null);
  const milestonesButtonRef = useRef<HTMLButtonElement>(null);
  const awardsButtonRef = useRef<HTMLButtonElement>(null);

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

  const handleStandardProjectSelect = (project: StandardProject) => {
    setShowStandardProjectsPopover(false);
    onStandardProjectSelect?.(project);
  };

  const menuItems = [
    { id: "projects" as const, label: "STANDARD PROJECTS", color: "#4a90e2" },
    { id: "milestones" as const, label: "MILESTONES", color: "#ff6b35" },
    { id: "awards" as const, label: "AWARDS", color: "#f39c12" },
  ];

  const handleTabClick = (tabId: "milestones" | "projects" | "awards") => {
    if (currentPlayer?.pendingTileSelection) return;

    if (tabId === "projects") {
      setShowStandardProjectsPopover((prev) => !prev);
    } else if (tabId === "milestones") {
      setShowMilestonePopover((prev) => !prev);
    } else if (tabId === "awards") {
      setShowAwardPopover((prev) => !prev);
    }
  };

  // Get the appropriate ref for each button
  const getButtonRef = (itemId: "projects" | "milestones" | "awards") => {
    if (itemId === "projects") return standardProjectsButtonRef;
    if (itemId === "milestones") return milestonesButtonRef;
    if (itemId === "awards") return awardsButtonRef;
    return null;
  };

  const calcScale = () => Math.min(1, Math.max(0.75, window.innerWidth / 2200));
  const [topBarScale, setTopBarScale] = useState(calcScale);

  useEffect(() => {
    const handleResize = () => setTopBarScale(calcScale());
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const buttonWidths = [250, 190, 160];
  const buttonHeight = 40;

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
          {menuItems.map((item, index) => (
            <ParallelogramButton
              key={item.id}
              index={index}
              total={menuItems.length}
              width={buttonWidths[index]}
              height={buttonHeight}
              color={item.color}
              onClick={() => handleTabClick(item.id)}
              buttonRef={
                getButtonRef(
                  item.id,
                ) as React.RefObject<HTMLButtonElement | null>
              }
            >
              {item.label}
            </ParallelogramButton>
          ))}
        </div>

        <div
          className="origin-top-right"
          style={{ transform: `scale(${topBarScale})` }}
        >
          <button
            ref={hamburgerButtonRef}
            onClick={() => setMenuOpen(!menuOpen)}
            onMouseEnter={() => setHamburgerHovered(true)}
            onMouseLeave={() => setHamburgerHovered(false)}
            aria-label="Menu"
            className="relative pointer-events-auto cursor-pointer"
            style={{ width: HAMBURGER_WIDTH, height: buttonHeight }}
          >
            <svg
              className="absolute inset-0 w-full h-full"
              viewBox={`0 0 ${HAMBURGER_WIDTH} ${buttonHeight}`}
              preserveAspectRatio="none"
            >
              <polygon
                points={`${ANGLE_INDENT},0 ${HAMBURGER_WIDTH},0 ${HAMBURGER_WIDTH},${buttonHeight} 0,${buttonHeight}`}
                fill={
                  hamburgerHovered
                    ? "rgba(20,20,25,0.95)"
                    : "rgba(10,10,15,0.95)"
                }
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
        >
          <div className="py-1">
            <button
              onClick={() => {
                menuItemHover.onClick?.();
                void handleCopyGameLink();
              }}
              onMouseEnter={menuItemHover.onMouseEnter}
              className="w-full flex items-center gap-3 px-4 py-3 text-white text-sm hover:bg-white/10 transition-colors text-left"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
              </svg>
              Copy game link
            </button>
            <div className="border-t border-[#333]" />
            <SoundToggleButton />
            <div className="border-t border-[#333]" />
            <button
              onClick={handleToggleFullscreen}
              className="w-full flex items-center gap-3 px-4 py-3 text-white text-sm hover:bg-white/10 transition-colors text-left"
            >
              {isFullscreen ? (
                <>
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <polyline points="4 14 10 14 10 20" />
                    <polyline points="20 10 14 10 14 4" />
                    <line x1="14" y1="10" x2="21" y2="3" />
                    <line x1="3" y1="21" x2="10" y2="14" />
                  </svg>
                  Exit Fullscreen
                </>
              ) : (
                <>
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <polyline points="15 3 21 3 21 9" />
                    <polyline points="9 21 3 21 3 15" />
                    <line x1="21" y1="3" x2="14" y2="10" />
                    <line x1="3" y1="21" x2="10" y2="14" />
                  </svg>
                  Fullscreen
                </>
              )}
            </button>
            <div className="border-t border-[#333]" />
            <button
              onClick={() => {
                menuItemHover.onClick?.();
                setMenuOpen(false);
                window.dispatchEvent(
                  new CustomEvent("toggle-performance-window"),
                );
              }}
              onMouseEnter={menuItemHover.onMouseEnter}
              className="w-full flex items-center gap-3 px-4 py-3 text-white text-sm hover:bg-white/10 transition-colors text-left"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <polyline points="22,12 18,12 15,21 9,3 6,12 2,12" />
              </svg>
              Performance
            </button>
            <div className="border-t border-[#333]" />
            <button
              onClick={() => {
                menuItemHover.onClick?.();
                handleLeaveGame();
              }}
              onMouseEnter={menuItemHover.onMouseEnter}
              className="w-full flex items-center gap-3 px-4 py-3 text-red-400 text-sm hover:bg-white/10 transition-colors text-left"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16 17 21 12 16 7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
              Leave game
            </button>
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
    </div>
  );
};

export default TopMenuBar;
