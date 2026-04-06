import { FC, useEffect, useState } from "react";
import GameIcon from "../display/GameIcon";
import { CardVPDetailDto } from "@/types/generated/api-types";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface CardVPOverlayProps {
  /** Per-card VP details */
  cardVPDetails: CardVPDetailDto[];
  /** Current card index being displayed */
  currentCardIndex: number;
  /** Whether the overlay is visible */
  isVisible: boolean;
  /** Total VP from all cards */
  totalCardVP: number;
}

/**
 * CardVPOverlay - Center-screen overlay showing detailed card VP breakdown
 * Shows one card at a time with VP math calculation
 */
const CardVPOverlay: FC<CardVPOverlayProps> = ({
  cardVPDetails,
  currentCardIndex,
  isVisible,
  totalCardVP,
}) => {
  const [animateIn, setAnimateIn] = useState(false);

  useEffect(() => {
    if (!isVisible) {
      return;
    }
    setAnimateIn(false);
    // Trigger animation after brief delay
    const timeout = setTimeout(() => setAnimateIn(true), 50);
    return () => clearTimeout(timeout);
  }, [isVisible, currentCardIndex]);

  if (!isVisible || cardVPDetails.length === 0) {
    return null;
  }

  const currentCard =
    currentCardIndex < cardVPDetails.length ? cardVPDetails[currentCardIndex] : null;

  if (!currentCard) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 flex items-center justify-center pointer-events-none"
      style={{ zIndex: Z_INDEX.MENU_DROPDOWN }}
    >
      <div
        className={`
          bg-space-black/95 backdrop-blur-lg border-2 border-purple-500/50 rounded-xl
          p-6 max-w-md w-full mx-4 shadow-2xl
          transform transition-all duration-300
          ${animateIn ? "opacity-100 scale-100 translate-y-0" : "opacity-0 scale-95 translate-y-4"}
        `}
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <GameIcon iconType="card" size="medium" />
            <span className="text-purple-400 font-orbitron text-sm">
              Card {currentCardIndex + 1} of {cardVPDetails.length}
            </span>
          </div>
          <div className="bg-purple-600/30 px-3 py-1 rounded-lg">
            <span className="text-purple-300 font-orbitron text-sm">Total: {totalCardVP} VP</span>
          </div>
        </div>

        {/* Card Name */}
        <h3 className="text-xl font-orbitron text-white mb-4 text-center">
          {currentCard.cardName}
        </h3>

        {/* VP Breakdown */}
        <div className="space-y-3">
          {currentCard.conditions.map((condition, idx) => (
            <div key={idx} className="bg-purple-900/30 border border-purple-500/30 rounded-lg p-4">
              {/* Condition type badge */}
              <div className="flex items-center justify-between mb-2">
                <span
                  className={`
                  text-xs uppercase tracking-wider px-2 py-1 rounded
                  ${condition.conditionType === "fixed" ? "bg-blue-600/50 text-blue-200" : ""}
                  ${condition.conditionType === "per" ? "bg-green-600/50 text-green-200" : ""}
                  ${condition.conditionType === "once" ? "bg-yellow-600/50 text-yellow-200" : ""}
                `}
                >
                  {condition.conditionType}
                </span>
                <span className="text-2xl font-orbitron font-bold text-purple-400">
                  +{condition.totalVP} VP
                </span>
              </div>

              {/* Math explanation */}
              <p className="text-white/80 text-sm">{condition.explanation}</p>

              {/* Detailed breakdown for "per" conditions */}
              {condition.conditionType === "per" && condition.count > 0 && (
                <div className="mt-2 pt-2 border-t border-purple-500/20">
                  <div className="flex items-center gap-2 text-xs text-white/60">
                    <span>Count: {condition.count}</span>
                    <span className="text-white/40">|</span>
                    <span>Triggers: {condition.actualTriggers}</span>
                    {condition.maxTrigger !== null && condition.maxTrigger !== undefined && (
                      <>
                        <span className="text-white/40">|</span>
                        <span>Max: {condition.maxTrigger}</span>
                      </>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Card total */}
        <div className="mt-4 pt-4 border-t border-purple-500/30 flex items-center justify-between">
          <span className="text-white/60 text-sm">Card VP Total</span>
          <span className="text-3xl font-orbitron font-bold text-purple-400">
            {currentCard.totalVP} VP
          </span>
        </div>

        {/* Progress dots */}
        <div className="flex items-center justify-center gap-2 mt-4">
          {cardVPDetails.map((_, idx) => (
            <div
              key={idx}
              className={`
                w-2 h-2 rounded-full transition-all duration-300
                ${idx === currentCardIndex ? "bg-purple-400 scale-125" : "bg-purple-800"}
              `}
            />
          ))}
        </div>
      </div>
    </div>
  );
};

export default CardVPOverlay;
