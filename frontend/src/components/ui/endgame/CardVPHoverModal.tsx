import { FC } from "react";
import type { FinalScoreDto } from "../../../types/generated/api-types";
import { Z_INDEX } from "@/constants/zIndex.ts";
import GameIcon from "../display/GameIcon";

interface CardVPHoverModalProps {
  playerScore: FinalScoreDto | undefined;
}

/** Card VP hover modal - shows all cards with VP for a player */
const CardVPHoverModal: FC<CardVPHoverModalProps> = ({ playerScore }) => {
  if (!playerScore) return null;

  const cardVPDetails = playerScore.vpBreakdown.cardVPDetails ?? [];
  const cardsWithVP = cardVPDetails.filter((card) => card.totalVP > 0);

  if (cardsWithVP.length === 0) {
    return (
      <div
        className="fixed inset-0 flex items-center justify-center pointer-events-none"
        style={{ zIndex: Z_INDEX.MENU_DROPDOWN }}
      >
        <div className="bg-space-black/95 backdrop-blur-lg border-2 border-purple-500/50 rounded-xl p-6 max-w-md w-full mx-4 shadow-2xl">
          <div className="flex items-center gap-2 mb-4">
            <GameIcon iconType="card" size="medium" />
            <span className="text-purple-400 font-orbitron text-sm">
              {playerScore.playerName}&apos;s Cards
            </span>
          </div>
          <p className="text-white/60 text-center">No cards with VP</p>
        </div>
      </div>
    );
  }

  return (
    <div
      className="fixed inset-0 flex items-center justify-center pointer-events-none"
      style={{ zIndex: Z_INDEX.MENU_DROPDOWN }}
    >
      <div className="bg-space-black/95 backdrop-blur-lg border-2 border-purple-500/50 rounded-xl p-6 max-w-lg w-full mx-4 shadow-2xl max-h-[70vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <GameIcon iconType="card" size="medium" />
            <span className="text-purple-400 font-orbitron text-sm">
              {playerScore.playerName}&apos;s Cards
            </span>
          </div>
          <div className="bg-purple-600/30 px-3 py-1 rounded-lg">
            <span className="text-purple-300 font-orbitron text-sm">
              Total: {playerScore.vpBreakdown.cardVP} VP
            </span>
          </div>
        </div>

        {/* Cards list */}
        <div className="space-y-3">
          {cardsWithVP.map((card, idx) => (
            <div key={idx} className="bg-purple-900/30 border border-purple-500/30 rounded-lg p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-white font-medium">{card.cardName}</span>
                <span className="text-purple-400 font-orbitron font-bold">+{card.totalVP} VP</span>
              </div>
              {/* Condition breakdown */}
              {card.conditions.map((condition, condIdx) => (
                <div key={condIdx} className="text-xs text-white/60 flex items-center gap-2">
                  <span
                    className={`
                      uppercase px-1.5 py-0.5 rounded text-[10px]
                      ${condition.conditionType === "fixed" ? "bg-blue-600/50 text-blue-200" : ""}
                      ${condition.conditionType === "per" ? "bg-green-600/50 text-green-200" : ""}
                      ${condition.conditionType === "once" ? "bg-yellow-600/50 text-yellow-200" : ""}
                    `}
                  >
                    {condition.conditionType}
                  </span>
                  <span>{condition.explanation}</span>
                </div>
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default CardVPHoverModal;
