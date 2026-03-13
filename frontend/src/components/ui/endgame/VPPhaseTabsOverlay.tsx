import { FC } from "react";
import type { VPPhase } from "../../../contexts/VPCountingContext";

const ANGLE_INDENT = 14;
const TAB_WIDTH = 110;
const TAB_HEIGHT = 32;
const TAB_SPACING = 4;
const BORDER_COLOR = "rgba(60,60,70,0.7)";

interface VPPhaseTabsOverlayProps {
  phases: VPPhase[];
  currentPhaseIndex: number;
  isActive: boolean;
}

const VPPhaseTabsOverlay: FC<VPPhaseTabsOverlayProps> = ({
  phases,
  currentPhaseIndex,
  isActive,
}) => {
  if (!isActive && currentPhaseIndex < 0) {
    return null;
  }

  return (
    <div className="flex items-center justify-center pb-1">
      {phases.map((phase, index) => (
        <PhaseTab
          key={phase.id}
          phase={phase}
          index={index}
          currentPhaseIndex={currentPhaseIndex}
          glow={index === currentPhaseIndex}
        />
      ))}
    </div>
  );
};

interface PhaseTabProps {
  phase: VPPhase;
  index: number;
  currentPhaseIndex: number;
  glow: boolean;
}

const PhaseTab: FC<PhaseTabProps> = ({ phase, index, currentPhaseIndex, glow }) => {
  const isCurrent = index === currentPhaseIndex;

  const w = TAB_WIDTH;
  const h = TAB_HEIGHT;

  const fillPoints = `0,0 ${w - ANGLE_INDENT},0 ${w},${h} ${ANGLE_INDENT},${h}`;
  const glowGradId = `phase-glow-grad-${index}`;
  const glowFilterId = `phase-glow-filter-${index}`;

  const topStrokeColor = isCurrent ? "#ffffff" : BORDER_COLOR;
  const topStrokeWidth = isCurrent ? 3 : 1;

  const textOpacity = isCurrent ? 1 : 0.7;

  return (
    <div
      className="relative"
      style={{
        width: w,
        height: h,
        marginLeft: index === 0 ? 0 : -ANGLE_INDENT + TAB_SPACING,
        zIndex: 10 - index,
        opacity: isCurrent ? 1 : 0.4,
        transition: "opacity 300ms ease-out",
      }}
    >
      <svg
        className="absolute inset-0 w-full h-full"
        viewBox={`0 0 ${w} ${h}`}
        preserveAspectRatio="none"
      >
        {glow && (
          <defs>
            <linearGradient
              id={glowGradId}
              x1="0"
              y1="0"
              x2={w - ANGLE_INDENT}
              y2="0"
              gradientUnits="userSpaceOnUse"
            >
              <stop offset="0%" stopColor="rgba(255,255,255,0)" />
              <stop offset="30%" stopColor="rgba(255,255,255,0.6)" />
              <stop offset="50%" stopColor="rgba(255,255,255,0.8)" />
              <stop offset="70%" stopColor="rgba(255,255,255,0.6)" />
              <stop offset="100%" stopColor="rgba(255,255,255,0)" />
            </linearGradient>
            <filter id={glowFilterId}>
              <feGaussianBlur stdDeviation="3" />
            </filter>
          </defs>
        )}
        <polygon points={fillPoints} fill="rgba(10,10,15,0.95)" />
        {glow && (
          <line
            x1={0}
            y1={2}
            x2={w - ANGLE_INDENT}
            y2={2}
            stroke={`url(#${glowGradId})`}
            strokeWidth={8}
            filter={`url(#${glowFilterId})`}
          />
        )}
        <line
          x1={0}
          y1={0}
          x2={w - ANGLE_INDENT}
          y2={0}
          stroke={topStrokeColor}
          strokeWidth={topStrokeWidth}
        />
        <line x1={w - ANGLE_INDENT} y1={0} x2={w} y2={h} stroke={BORDER_COLOR} strokeWidth="1" />
        <line x1={0} y1={0} x2={ANGLE_INDENT} y2={h} stroke={BORDER_COLOR} strokeWidth="1" />
      </svg>
      <div
        className="absolute inset-0 flex items-center justify-center font-orbitron text-xs tracking-wider"
        style={{
          color: "rgba(255,255,255,0.8)",
          opacity: textOpacity,
          paddingLeft: ANGLE_INDENT / 2,
          paddingRight: ANGLE_INDENT / 2,
        }}
      >
        {phase.label}
      </div>
    </div>
  );
};

export default VPPhaseTabsOverlay;
