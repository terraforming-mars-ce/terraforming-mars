import React, { useState, useEffect } from "react";
import GameButton from "../buttons/GameButton.tsx";
import { GameFlowPopover, GameFlowTitle, GameFlowFooter } from "./GameFlowPopover.tsx";

interface AmountSelectionPopoverProps {
  cardName: string;
  resourceLabel: string;
  maxAmount: number;
  onAmountSelect: (amount: number) => void;
  onCancel: () => void;
  isVisible: boolean;
}

const AmountSelectionPopover: React.FC<AmountSelectionPopoverProps> = ({
  cardName,
  resourceLabel,
  maxAmount,
  onAmountSelect,
  onCancel,
  isVisible,
}) => {
  const [amount, setAmount] = useState(maxAmount);

  useEffect(() => {
    if (isVisible) {
      setAmount(maxAmount);
    }
  }, [isVisible, maxAmount]);

  const handleConfirm = () => {
    onAmountSelect(amount);
  };

  return (
    <GameFlowPopover isVisible={isVisible} onClose={onCancel} className="min-w-[280px]">
      <GameFlowTitle>
        <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
          Select Amount
        </h3>
        <div className="text-white/60 text-xs text-shadow-glow mt-1">{cardName}</div>
      </GameFlowTitle>

      <div className="p-5 flex flex-col items-center gap-4">
        <div className="text-white/80 text-sm text-center">
          How many <span className="text-space-blue-300 font-semibold">{resourceLabel}</span>?
        </div>

        <div className="flex items-center gap-4">
          <button
            className="w-9 h-9 rounded-lg bg-space-blue-600/50 border border-space-blue-500/60 text-white text-lg font-bold cursor-pointer transition-all duration-200 hover:bg-space-blue-500/60 disabled:opacity-30 disabled:cursor-default"
            onClick={() => setAmount((a) => Math.max(0, a - 1))}
            disabled={amount <= 0}
          >
            -
          </button>

          <div className="min-w-[60px] text-center">
            <span className="text-white text-3xl font-orbitron font-bold text-shadow-glow">
              {amount}
            </span>
            <span className="text-white/40 text-sm ml-1">/ {maxAmount}</span>
          </div>

          <button
            className="w-9 h-9 rounded-lg bg-space-blue-600/50 border border-space-blue-500/60 text-white text-lg font-bold cursor-pointer transition-all duration-200 hover:bg-space-blue-500/60 disabled:opacity-30 disabled:cursor-default"
            onClick={() => setAmount((a) => Math.min(maxAmount, a + 1))}
            disabled={amount >= maxAmount}
          >
            +
          </button>
        </div>

        {maxAmount > 1 && (
          <input
            type="range"
            min={0}
            max={maxAmount}
            value={amount}
            onChange={(e) => setAmount(parseInt(e.target.value))}
            className="w-full accent-space-blue-500"
          />
        )}
      </div>

      <GameFlowFooter className="gap-3">
        <GameButton buttonType="secondary" variant="info" size="sm" onClick={onCancel}>
          Cancel
        </GameButton>
        <GameButton buttonType="primary" variant="success" size="sm" onClick={handleConfirm}>
          Confirm
        </GameButton>
      </GameFlowFooter>
    </GameFlowPopover>
  );
};

export default AmountSelectionPopover;
