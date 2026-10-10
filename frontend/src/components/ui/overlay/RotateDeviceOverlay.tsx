import { Z_INDEX } from "@/constants/zIndex.ts";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

export default function RotateDeviceOverlay() {
  const { isCoarsePointer, isPortrait, isPhoneScreen } = useLayoutMode();
  const isPhonePortrait = isCoarsePointer && isPortrait && isPhoneScreen;
  if (!isPhonePortrait) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 flex flex-col items-center justify-center gap-6 bg-black px-8 text-center text-white select-none"
      style={{ zIndex: Z_INDEX.ROTATE_DEVICE_OVERLAY }}
    >
      <svg
        viewBox="0 0 120 120"
        className="h-28 w-28 text-white/80 motion-safe:animate-[rotateDevice_2.4s_ease-in-out_infinite]"
        aria-hidden="true"
      >
        <rect
          x="38"
          y="14"
          width="44"
          height="92"
          rx="8"
          fill="none"
          stroke="currentColor"
          strokeWidth="4"
        />
        <line
          x1="52"
          y1="96"
          x2="68"
          y2="96"
          stroke="currentColor"
          strokeWidth="4"
          strokeLinecap="round"
        />
      </svg>
      <div className="font-orbitron text-lg font-semibold tracking-widest uppercase">
        Rotate your device
      </div>
      <div className="text-sm text-white/60">Open Mars is played in landscape.</div>
    </div>
  );
}
