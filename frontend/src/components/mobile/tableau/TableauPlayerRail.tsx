import GameButton from "../../ui/buttons/GameButton.tsx";

interface TableauPlayerRailProps {
  color: string;
  playerName: string;
  canSwitch: boolean;
  onPrevious: () => void;
  onNext: () => void;
}

function Chevron({ direction }: { direction: "left" | "right" }) {
  const points = direction === "left" ? "15 18 9 12 15 6" : "9 18 15 12 9 6";
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <polyline points={points} />
    </svg>
  );
}

export default function TableauPlayerRail({
  color,
  playerName,
  canSwitch,
  onPrevious,
  onNext,
}: TableauPlayerRailProps) {
  return (
    <div
      className="shrink-0 w-12 h-full flex flex-col items-center justify-between border-r border-white/10 bg-black/30"
      style={{ boxShadow: `inset 4px 0 0 ${color}` }}
    >
      {canSwitch && (
        <GameButton
          emphasis="quiet"
          aria-label="Previous player"
          className="!p-0 !min-h-11 w-11 h-11 shrink-0 flex items-center justify-center"
          onClick={onPrevious}
        >
          <Chevron direction="left" />
        </GameButton>
      )}
      <span
        className="flex-1 min-h-0 flex items-center justify-center overflow-hidden font-orbitron text-[12px] font-bold tracking-wider uppercase [writing-mode:vertical-rl] rotate-180 truncate"
        style={{ color }}
      >
        {playerName}
      </span>
      {canSwitch && (
        <GameButton
          emphasis="quiet"
          aria-label="Next player"
          className="!p-0 !min-h-11 w-11 h-11 shrink-0 flex items-center justify-center"
          onClick={onNext}
        >
          <Chevron direction="right" />
        </GameButton>
      )}
    </div>
  );
}
