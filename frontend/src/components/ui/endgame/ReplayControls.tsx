import GameButton from "@/components/ui/buttons/GameButton.tsx";
import {
  FC,
  ReactNode,
  useState,
  useMemo,
  useRef,
  type PointerEvent as ReactPointerEvent,
} from "react";
import { getZIndex } from "@/constants/zIndex.ts";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

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
  const { isCompact, isCoarsePointer } = useLayoutMode();
  const isTouch = isCompact || isCoarsePointer;
  const trackRef = useRef<HTMLDivElement>(null);
  const scrubPointerRef = useRef<number | null>(null);
  const lastSeekRef = useRef<number | null>(null);

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

  const showDetail = isTouch || isSliderHovered;
  const scrubberHeight = isTouch ? 44 : 20;
  let sliderHeight = isSliderHovered ? 10 : 4;
  let thumbSize = isSliderHovered ? 14 : 0;
  if (isTouch) {
    sliderHeight = 6;
    thumbSize = 18;
  }

  const seekToClientX = (clientX: number) => {
    const track = trackRef.current;
    if (!track) {
      return;
    }
    const rect = track.getBoundingClientRect();
    if (rect.width === 0) {
      return;
    }
    const ratio = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width));
    const index = Math.round(ratio * Math.max(totalStates - 1, 0));
    if (index === lastSeekRef.current) {
      return;
    }
    lastSeekRef.current = index;
    onSeek(index);
  };

  const handleScrubStart = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (!isTouch) {
      return;
    }
    scrubPointerRef.current = event.pointerId;
    lastSeekRef.current = currentIndex;
    event.currentTarget.setPointerCapture(event.pointerId);
    seekToClientX(event.clientX);
  };
  const handleScrubMove = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (scrubPointerRef.current !== event.pointerId) {
      return;
    }
    seekToClientX(event.clientX);
  };
  const handleScrubEnd = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (scrubPointerRef.current !== event.pointerId) {
      return;
    }
    scrubPointerRef.current = null;
  };

  const touchTargetClass = isTouch ? " min-w-[44px] min-h-[44px]" : "";

  const playbackButtons = (
    <div className="flex items-center gap-1">
      <GameButton
        emphasis="quiet"
        onClick={onStepBackward}
        disabled={currentIndex === 0}
        aria-label="Step back"
        className={`p-2 text-white/60 hover:text-white disabled:text-white/20 transition-colors cursor-pointer disabled:cursor-default${touchTargetClass}`}
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M6 6h2v12H6zm3.5 6l8.5 6V6z" />
        </svg>
      </GameButton>

      <GameButton
        emphasis="quiet"
        onClick={isPlaying ? onPause : onPlay}
        aria-label={isPlaying ? "Pause" : "Play"}
        className={`p-2 text-white/80 hover:text-white transition-colors cursor-pointer${touchTargetClass}`}
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
      </GameButton>

      <GameButton
        emphasis="quiet"
        onClick={onStepForward}
        disabled={currentIndex >= totalStates - 1}
        aria-label="Step forward"
        className={`p-2 text-white/60 hover:text-white disabled:text-white/20 transition-colors cursor-pointer disabled:cursor-default${touchTargetClass}`}
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z" />
        </svg>
      </GameButton>
    </div>
  );

  const speedButtons = SPEED_OPTIONS.map((opt) => (
    <GameButton
      emphasis="quiet"
      key={opt.value}
      onClick={() => onSpeedChange(opt.value)}
      className={`px-1.5 py-0.5 text-xs rounded-none transition-colors cursor-pointer${touchTargetClass} ${
        playbackSpeed === opt.value ? "bg-white/20 text-white" : "text-white/40 hover:text-white/70"
      }`}
    >
      {opt.label}
    </GameButton>
  ));

  return (
    <div className="space-y-1">
      {/* Custom timeline scrubber */}
      <div
        className="relative w-full cursor-pointer"
        style={{
          height: scrubberHeight,
          display: "flex",
          alignItems: "center",
          touchAction: isTouch ? "none" : undefined,
        }}
        onMouseEnter={() => setIsSliderHovered(true)}
        onMouseLeave={() => setIsSliderHovered(false)}
        onPointerDown={handleScrubStart}
        onPointerMove={handleScrubMove}
        onPointerUp={handleScrubEnd}
        onPointerCancel={handleScrubEnd}
      >
        <input
          type="range"
          min={0}
          max={Math.max(totalStates - 1, 0)}
          value={currentIndex}
          onChange={(e) => onSeek(parseInt(e.target.value, 10))}
          aria-label="Replay position"
          className={`absolute inset-0 w-full opacity-0 cursor-pointer ${isTouch ? "pointer-events-none" : ""}`}
          style={{ height: scrubberHeight, zIndex: getZIndex("LOCAL", 2) }}
        />
        <div
          ref={trackRef}
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
          {markerPositions.map((m, i) => (
            <div
              key={i}
              className="absolute"
              style={{
                left: `${m.pct}%`,
                top: "50%",
                transform: "translate(-50%, -50%)",
                opacity: showDetail ? 1 : 0,
                transition: "opacity 150ms ease",
                pointerEvents: "none",
              }}
            >
              <div
                style={{
                  width: 2,
                  height: showDetail ? 18 : 8,
                  backgroundColor: "rgba(255,255,255,0.4)",
                  transition: "height 150ms ease",
                }}
              />
              <div
                className="text-white/40 font-orbitron whitespace-nowrap"
                style={{
                  fontSize: isTouch ? 11 : 9,
                  position: "absolute",
                  top: isTouch ? -18 : -16,
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

      {isTouch ? (
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <div className="min-w-0 flex-1 basis-[140px] truncate font-orbitron text-[12px] text-white/60">
            {currentLabel}
            <span className="ml-2 text-[11px] tabular-nums text-white/30">
              ({currentIndex + 1} / {totalStates})
            </span>
          </div>
          {playbackButtons}
          <div className="flex items-center gap-1">{speedButtons}</div>
          {rightSlot && <div className="flex items-center">{rightSlot}</div>}
        </div>
      ) : (
        <div className="relative flex items-center justify-center">
          <div className="absolute left-0 text-sm text-white/60 font-orbitron truncate">
            {currentLabel}
            <span className="ml-2 text-white/30 tabular-nums text-xs">
              ({currentIndex + 1} / {totalStates})
            </span>
          </div>

          <div className="absolute flex items-center gap-1" style={{ left: "35%" }}>
            {speedButtons}
          </div>

          {playbackButtons}

          {rightSlot && <div className="absolute right-0">{rightSlot}</div>}
        </div>
      )}
    </div>
  );
};

export default ReplayControls;
