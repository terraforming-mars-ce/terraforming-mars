import { getZIndex } from "@/constants/zIndex.ts";
import { useBackDismiss } from "@/hooks/useBackDismiss.ts";
import type { PlayerHostAction } from "@/hooks/usePlayerHostActions.ts";
import GameButton from "../../ui/buttons/GameButton.tsx";

interface PlayerActionSheetProps {
  playerName: string;
  actions: PlayerHostAction[];
  onClose: () => void;
}

export default function PlayerActionSheet({
  playerName,
  actions,
  onClose,
}: PlayerActionSheetProps) {
  useBackDismiss(onClose);

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={`Actions for ${playerName}`}
      className="fixed inset-0 flex items-end justify-center"
      style={{ zIndex: getZIndex("LOCAL", 1) }}
    >
      <div className="absolute inset-0 bg-black/60 cursor-default" onClick={onClose} />
      <div
        className="relative w-full max-w-[420px] game-panel game-panel-clipped game-window flex flex-col gap-2 px-4 pt-3"
        style={{ paddingBottom: "calc(var(--safe-bottom) + 12px)" }}
      >
        <div className="font-orbitron text-[13px] font-bold uppercase tracking-wider text-white/70 truncate">
          {playerName}
        </div>
        {actions.map((action) => (
          <GameButton
            key={action.id}
            emphasis="secondary"
            tone={action.tone === "danger" ? "error" : "info"}
            className="!min-h-11 w-full font-orbitron !text-[13px]"
            onClick={() => {
              onClose();
              action.run();
            }}
          >
            {action.label}
          </GameButton>
        ))}
        <GameButton
          emphasis="quiet"
          className="!min-h-11 w-full font-orbitron !text-[13px]"
          onClick={onClose}
        >
          Cancel
        </GameButton>
      </div>
    </div>
  );
}
