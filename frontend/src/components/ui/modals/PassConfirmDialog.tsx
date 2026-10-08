import type { PassConfirmation } from "@/hooks/usePassAction.ts";
import GameButton from "../buttons/GameButton.tsx";
import { GameFlowPopover, GameFlowTitle, GameFlowFooter } from "../popover/GameFlowPopover.tsx";

interface PassConfirmDialogProps {
  confirmation: PassConfirmation;
}

export default function PassConfirmDialog({ confirmation }: PassConfirmDialogProps) {
  return (
    <GameFlowPopover isVisible={confirmation.open} onClose={confirmation.dismiss} type="immediate">
      <GameFlowTitle>
        <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
          Actions Available
        </h3>
        <div className="text-white/60 text-xs text-shadow-glow mt-1">{confirmation.reason}</div>
      </GameFlowTitle>
      <GameFlowFooter className="gap-3">
        <GameButton emphasis="secondary" size="sm" onClick={confirmation.dismiss}>
          Cancel
        </GameButton>
        <GameButton tone="warn" size="sm" onClick={confirmation.confirm}>
          Pass Anyway
        </GameButton>
      </GameFlowFooter>
    </GameFlowPopover>
  );
}
