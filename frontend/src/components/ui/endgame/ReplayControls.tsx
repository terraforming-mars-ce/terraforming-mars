import { FC, ReactNode, useState, useMemo } from "react";

interface GenerationMarker {
  index: number;
  generation: number;
}

interface ReplayControlsProps {
  currentIndex: number;
  totalStates: number;
  isPlaying: boolean;
  playbackSpeed: number;
  currentLabel: string;
  onPlay: () => void;
  onPause: () => void;
  onStepForward: () => void;
  onStepBackward: () => void;
  onSeek: (index: number) => void;
  onSpeedChange: (speed: number) => void;
  generationMarkers?: GenerationMarker[];
  rightSlot?: ReactNode;
}

const SPEED_OPTIONS: { value: number; label: string }[] = [
  { value: 0.25, label: "x0.5" },
  { value: 0.5, label: "x1" },
  { value: 1, label: "x2" },
  { value: 2, label: "x4" },
];

const ReplayControls: FC<ReplayControlsProps> = ({
  currentIndex,
  totalStates,
  isPlaying,
  playbackSpeed,
  currentLabel,
  onPlay,
  onPause,
  onStepForward,
  onStepBackward,
  onSeek,
  onSpeedChange,
  generationMarkers = [],
  rightSlot,
}) => {
  const [isSliderHovered, setIsSliderHovered] = useState(false);

  const progressPct = useMemo(() => {
    const max = Math.max(totalStates - 1, 1);
    return (currentIndex / max) * 100;
  }, [currentIndex, totalStates]);

  const markerPositions = useMemo(() => {
    const max = Math.max(totalStates - 1, 1);
    return generationMarkers.map((m) => ({
      pct: (m.index / max) * 100,
      label: `Gen ${m.generation}`,
    }));
  }, [generationMarkers, totalStates]);

  const sliderHeight = isSliderHovered ? 10 : 4;
  const thumbSize = isSliderHovered ? 14 : 0;

  return (
    <div className="space-y-1">
      {/* Custom timeline scrubber */}
      <div
        className="relative w-full cursor-pointer"
        style={{ height: 20, display: "flex", alignItems: "center" }}
        onMouseEnter={() => setIsSliderHovered(true)}
        onMouseLeave={() => setIsSliderHovered(false)}
      >
        <input
          type="range"
          min={0}
          max={Math.max(totalStates - 1, 0)}
          value={currentIndex}
          onChange={(e) => onSeek(parseInt(e.target.value, 10))}
          className="absolute inset-0 w-full opacity-0 cursor-pointer"
          style={{ height: 20, zIndex: 2 }}
        />
        <div
          className="w-full relative"
          style={{ height: sliderHeight, transition: "height 150ms ease" }}
        >
          <div
            className="absolute inset-0 rounded-full"
            style={{ backgroundColor: "rgba(255,255,255,0.15)" }}
          />
          <div
            className="absolute top-0 left-0 h-full rounded-full"
            style={{
              width: `${progressPct}%`,
              backgroundColor: "rgba(255,255,255,0.7)",
              boxShadow: "0 0 8px rgba(255,255,255,0.4), 0 0 2px rgba(255,255,255,0.6)",
              transition: "width 400ms ease-out",
            }}
          />
          {/* Generation marker ticks — only on hover */}
          {markerPositions.map((m, i) => (
            <div
              key={i}
              className="absolute"
              style={{
                left: `${m.pct}%`,
                top: "50%",
                transform: "translate(-50%, -50%)",
                opacity: isSliderHovered ? 1 : 0,
                transition: "opacity 150ms ease",
                pointerEvents: "none",
              }}
            >
              <div
                style={{
                  width: 2,
                  height: isSliderHovered ? 18 : 8,
                  backgroundColor: "rgba(255,255,255,0.4)",
                  transition: "height 150ms ease",
                }}
              />
              <div
                className="text-white/40 font-orbitron whitespace-nowrap"
                style={{
                  fontSize: 9,
                  position: "absolute",
                  top: -16,
                  left: "50%",
                  transform: "translateX(-50%)",
                }}
              >
                {m.label}
              </div>
            </div>
          ))}
          <div
            className="absolute top-1/2 rounded-full bg-white"
            style={{
              left: `${progressPct}%`,
              width: thumbSize,
              height: thumbSize,
              transform: "translate(-50%, -50%)",
              transition: "left 400ms ease-out, width 150ms ease, height 150ms ease",
            }}
          />
        </div>
      </div>

      {/* Controls row */}
      <div className="relative flex items-center justify-center">
        {/* Label — absolutely positioned left */}
        <div className="absolute left-0 text-sm text-white/60 font-orbitron truncate">
          {currentLabel}
          <span className="ml-2 text-white/30 tabular-nums text-xs">
            ({currentIndex + 1} / {totalStates})
          </span>
        </div>

        {/* Speed selector — absolutely positioned right of label */}
        <div className="absolute flex items-center gap-1" style={{ left: "35%" }}>
          {SPEED_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => onSpeedChange(opt.value)}
              className={`px-1.5 py-0.5 text-xs rounded transition-colors cursor-pointer ${
                playbackSpeed === opt.value
                  ? "bg-white/20 text-white"
                  : "text-white/40 hover:text-white/70"
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>

        {/* Centered playback controls */}
        <div className="flex items-center gap-1">
          <button
            onClick={onStepBackward}
            disabled={currentIndex === 0}
            className="p-2 text-white/60 hover:text-white disabled:text-white/20 transition-colors cursor-pointer disabled:cursor-default"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <path d="M6 6h2v12H6zm3.5 6l8.5 6V6z" />
            </svg>
          </button>

          <button
            onClick={isPlaying ? onPause : onPlay}
            className="p-2 text-white/80 hover:text-white transition-colors cursor-pointer"
          >
            {isPlaying ? (
              <svg width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
              </svg>
            ) : (
              <svg width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
                <path d="M8 5v14l11-7z" />
              </svg>
            )}
          </button>

          <button
            onClick={onStepForward}
            disabled={currentIndex >= totalStates - 1}
            className="p-2 text-white/60 hover:text-white disabled:text-white/20 transition-colors cursor-pointer disabled:cursor-default"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <path d="M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z" />
            </svg>
          </button>
        </div>

        {/* Right slot — absolutely positioned right */}
        {rightSlot && <div className="absolute right-0">{rightSlot}</div>}
      </div>
    </div>
  );
};

export default ReplayControls;
