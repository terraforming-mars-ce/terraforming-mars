import React, { useEffect, useRef, useState } from "react";
import { Z_INDEX } from "@/constants/zIndex";

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
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);
  const [amount, setAmount] = useState(maxAmount);

  useEffect(() => {
    if (isVisible) {
      setAmount(maxAmount);
    }
  }, [isVisible, maxAmount]);

  const handleCancelClick = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onCancel();
    }, 200);
  };

  const handleConfirm = () => {
    onAmountSelect(amount);
  };

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        handleCancelClick();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        event.button === 0 &&
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node)
      ) {
        handleCancelClick();
      }
    };

    const preventScroll = (event: WheelEvent | TouchEvent) => {
      event.preventDefault();
      event.stopPropagation();
    };

    if (isVisible) {
      document.body.style.overflow = "hidden";
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("wheel", preventScroll, { passive: false });
      document.addEventListener("touchmove", preventScroll, { passive: false });
    }

    return () => {
      document.body.style.overflow = "";
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("wheel", preventScroll);
      document.removeEventListener("touchmove", preventScroll);
    };
  }, [isVisible, onCancel]);

  if (!isVisible) return null;

  return (
    <div
      className="fixed top-0 left-0 right-0 bottom-0 flex items-center justify-center pointer-events-auto overflow-hidden"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div
        className={`
          min-w-[280px] w-fit max-w-[90vw]
          bg-space-black-darker/95
          border-2 border-space-blue-500
          rounded-xl
          shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_rgba(30,60,150,1)]
          backdrop-blur-space
          flex flex-col overflow-hidden isolate
          pointer-events-auto
          ${isClosing ? "animate-fadeOut" : "animate-popIn"}
        `}
        ref={popoverRef}
      >
        {/* Header */}
        <div className="py-[15px] px-5 bg-black/40 border-b border-b-space-blue-500/60">
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            Select Amount
          </h3>
          <div className="text-white/60 text-xs text-shadow-glow mt-1">{cardName}</div>
        </div>

        {/* Amount Selector */}
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

          {/* Slider */}
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

        {/* Footer */}
        <div className="px-4 py-3 bg-black/40 border-t border-space-blue-500/60 flex justify-center gap-3">
          <button
            className="
              bg-space-blue-600/50
              border-2 border-space-blue-500/60
              rounded-md text-white text-xs font-semibold
              px-6 py-2 cursor-pointer
              transition-all duration-200
              text-shadow-glow font-orbitron
              shadow-[0_0_8px_rgba(30,60,150,0.4)]
              hover:bg-space-blue-500/60
              hover:border-space-blue-500/80

              hover:shadow-[0_0_12px_rgba(30,60,150,0.6)]
            "
            onClick={handleCancelClick}
          >
            Cancel
          </button>
          <button
            className="
              bg-green-600/60
              border-2 border-green-500/60
              rounded-md text-white text-xs font-semibold
              px-6 py-2 cursor-pointer
              transition-all duration-200
              text-shadow-glow font-orbitron
              shadow-[0_0_8px_rgba(34,197,94,0.4)]
              hover:bg-green-500/60
              hover:border-green-500/80

              hover:shadow-[0_0_12px_rgba(34,197,94,0.6)]
            "
            onClick={handleConfirm}
          >
            Confirm
          </button>
        </div>
      </div>

      <style>{`
        @keyframes popIn {
          from {
            opacity: 0;
            transform: scale(0.9) translateY(-20px);
          }
          to {
            opacity: 1;
            transform: scale(1) translateY(0);
          }
        }

        @keyframes fadeOut {
          from {
            opacity: 1;
          }
          to {
            opacity: 0;
          }
        }

        .animate-popIn {
          animation: popIn 0.25s ease-out;
        }

        .animate-fadeOut {
          animation: fadeOut 0.2s ease-out forwards;
        }
      `}</style>
    </div>
  );
};

export default AmountSelectionPopover;
