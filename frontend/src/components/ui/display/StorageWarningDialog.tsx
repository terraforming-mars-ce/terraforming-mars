import GameButton from "@/components/ui/buttons/GameButton.tsx";
import React from "react";
import { createPortal } from "react-dom";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface StorageWarningDialogProps {
  message: string;
  onCancel: () => void;
  onContinue: () => void;
}

const StorageWarningDialog: React.FC<StorageWarningDialogProps> = ({
  message,
  onCancel,
  onContinue,
}) => {
  return createPortal(
    <div
      className="fixed inset-0 flex items-center justify-center animate-[fadeIn_0.2s_ease]"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
      onClick={(e) => e.stopPropagation()}
    >
      <div className="absolute inset-0 bg-black/50" onClick={onCancel} />
      <div className="relative z-[1] game-panel game-panel-clipped game-window p-5 max-w-[340px]">
        <h3 className="font-orbitron text-sm font-bold text-amber-400 m-0 mb-2">
          No Storage Available
        </h3>
        <p className="text-white/70 text-xs mb-4 leading-relaxed">{message}</p>
        <div className="flex gap-2 justify-end">
          <GameButton
            emphasis="quiet"
            className="px-3 py-1.5 rounded-none text-xs font-orbitron text-white/50 hover:text-white/80 transition-colors cursor-pointer"
            onClick={onCancel}
          >
            Cancel
          </GameButton>
          <GameButton
            emphasis="quiet"
            className="px-3 py-1.5 rounded-none text-xs font-orbitron font-bold bg-amber-500/20 border border-amber-500/40 text-amber-400 hover:bg-amber-500/30 transition-colors cursor-pointer"
            onClick={onContinue}
          >
            Continue Anyway
          </GameButton>
        </div>
      </div>
    </div>,
    document.body,
  );
};

export default StorageWarningDialog;
