import React, { useState } from "react";
import GameCard from "../cards/GameCard.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { PendingCardSelectionDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import { useCardSelection } from "../../../hooks/useCardSelection.ts";
import {
  OVERLAY_BACKDROP_BLUR_CLASS,
  OVERLAY_BACKDROP_TINT_CLASS,
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_CARDS_CONTAINER_CLASS,
  OVERLAY_CARDS_INNER_CLASS,
  OVERLAY_FOOTER_CLASS,
  RESOURCE_LABEL_CLASS,
} from "./overlayStyles.ts";
import GameButton from "../buttons/GameButton.tsx";

interface PendingCardSelectionOverlayProps {
  isOpen: boolean;
  selection: PendingCardSelectionDto;
  playerCredits: number;
  onSelectCards: (selectedCardIds: string[]) => void;
  onCancel?: () => void;
}

interface TitleInfo {
  title: string;
  description: string;
}

const PendingCardSelectionOverlay: React.FC<PendingCardSelectionOverlayProps> = ({
  isOpen,
  selection,
  playerCredits,
  onSelectCards,
  onCancel,
}) => {
  const {
    selectedCardIds,
    totalCost,
    totalReward,
    showConfirmation,
    isValidSelection,
    handleCardSelect,
    handleConfirm: handleCardConfirm,
    canAffordCard,
  } = useCardSelection({
    cards: selection.availableCards,
    isOpen,
    playerCredits,
    getCardCost: (cardId) => selection.cardCosts[cardId] || 0,
    getCardReward: (cardId) => selection.cardRewards[cardId] || 0,
    minCards: selection.minCards,
    maxCards: selection.maxCards,
  });

  const [showPlayability, setShowPlayability] = useState(false);

  if (!isOpen || selection.availableCards.length === 0) return null;

  const isSellPatents = selection.source === "sell-patents";

  const getTitleAndDescription = (source: string): TitleInfo => {
    switch (source) {
      case "sell-patents":
        return {
          title: "Sell Cards for Credits",
          description: "Select cards to sell. Each card gives you 1 MC.",
        };
      case "research-phase":
        return {
          title: "Buy Research Cards",
          description: "Select cards to purchase. Each card costs 3 MC.",
        };
      default:
        return {
          title: "Select Cards",
          description: "Choose cards from the available options.",
        };
    }
  };

  const getCardBadge = (
    cardId: string,
  ): { type: "reward" | "cost" | "free"; value?: number } | null => {
    const cost = selection.cardCosts[cardId] || 0;
    const reward = selection.cardRewards[cardId] || 0;

    if (reward > 0) {
      return { type: "reward", value: reward };
    } else if (cost > 0) {
      return { type: "cost", value: cost };
    } else if (cost === 0 && reward === 0) {
      return { type: "free" };
    }
    return null;
  };

  const handleConfirm = () => {
    handleCardConfirm(onSelectCards);
  };

  const handleCancel = () => {
    if (onCancel) {
      onCancel();
    } else {
      // If no cancel handler, treat as selecting 0 cards (if allowed)
      if (selection.minCards === 0) {
        onSelectCards([]);
      }
    }
  };

  const titleInfo = getTitleAndDescription(selection.source);

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center">
      <div className={OVERLAY_BACKDROP_BLUR_CLASS} />
      <div className={OVERLAY_BACKDROP_TINT_CLASS} />

      {/* Content container */}
      <div className={OVERLAY_CONTAINER_CLASS}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <div className="flex items-center justify-between w-full">
            <div>
              <h2 className={OVERLAY_TITLE_CLASS}>{titleInfo.title}</h2>
              <p className={OVERLAY_DESCRIPTION_CLASS}>{titleInfo.description}</p>
            </div>
            {!isSellPatents && (
              <GameButton
                buttonType="secondary"
                size="sm"
                onClick={() => setShowPlayability((prev) => !prev)}
              >
                {showPlayability ? "Hide Playability" : "Show Playability"}
              </GameButton>
            )}
          </div>
        </div>

        {/* Cards display */}
        <div className={OVERLAY_CARDS_CONTAINER_CLASS}>
          <div className={OVERLAY_CARDS_INNER_CLASS}>
            {selection.availableCards.map((card, index) => {
              const isSelected = selectedCardIds.includes(card.id);
              const badge = getCardBadge(card.id);
              const canAfford = canAffordCard(card.id);

              return (
                <div key={card.id} className="relative">
                  <GameCard
                    card={card}
                    isSelected={isSelected}
                    onSelect={handleCardSelect}
                    animationDelay={index * 100}
                    showCheckbox={true}
                  />
                  {/* Cost/Reward Badge (hidden for sell-patents since reward is obvious) */}
                  {badge && selection.source !== "sell-patents" && (
                    <div
                      className={`absolute top-2 right-2 px-2 py-1 rounded-md font-bold text-sm shadow-lg ${
                        badge.type === "cost"
                          ? "bg-[#f44336] text-white"
                          : "bg-[#4caf50] text-white"
                      }`}
                    >
                      {badge.type === "cost"
                        ? `${badge.value} MC`
                        : badge.type === "reward"
                          ? `+${badge.value} MC`
                          : "FREE"}
                    </div>
                  )}
                  {/* Unaffordable overlay (buy cost) */}
                  {!canAfford && !isSelected && (
                    <div className="absolute inset-0 bg-black/60 rounded-lg flex items-center justify-center">
                      <span className="text-white/80 font-bold">Can't Afford</span>
                    </div>
                  )}
                  {/* Playability errors */}
                  {showPlayability && !card.available && card.errors.length > 0 && (
                    <div className="absolute top-full left-0 right-0 mt-1 flex flex-col gap-0.5 z-10 max-h-24 overflow-y-auto">
                      {card.errors.map((err, i) => (
                        <div
                          key={i}
                          className="bg-[rgba(10,10,15,0.95)] border border-[rgba(231,76,60,0.6)] border-l-[3px] border-l-[#e74c3c] text-white/90 text-[10px] leading-tight px-1.5 py-1 rounded-sm"
                        >
                          {err.message}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Footer with cost and confirm button */}
        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between max-[768px]:flex-wrap">
            <div className="flex items-center gap-3">
              <span className={RESOURCE_LABEL_CLASS}>Your Credits:</span>
              <GameIcon iconType={ResourceTypeCredit} amount={playerCredits} size="large" />
            </div>
            {totalCost > 0 && (
              <div className="flex items-center gap-3">
                <span className={RESOURCE_LABEL_CLASS}>Total Cost:</span>
                <GameIcon iconType={ResourceTypeCredit} amount={totalCost} size="large" />
              </div>
            )}
            {totalReward > 0 && (
              <div className="flex items-center gap-3">
                <span className={RESOURCE_LABEL_CLASS}>Gain:</span>
                <GameIcon iconType={ResourceTypeCredit} amount={totalReward} size="large" />
              </div>
            )}
          </div>

          <div className="flex items-center gap-6 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3">
            <div className="text-sm">
              {selectedCardIds.length === 0 ? (
                showConfirmation ? (
                  <span className="text-[#ff9800]">
                    Are you sure you don't want to select any cards?
                  </span>
                ) : (
                  <span className="text-white/70">No cards selected</span>
                )
              ) : (
                <span className="text-white/70">
                  {selectedCardIds.length} card
                  {selectedCardIds.length !== 1 ? "s" : ""} selected
                </span>
              )}
            </div>
            <div className="flex gap-3 items-center">
              {(onCancel || selection.minCards === 0) && (
                <GameButton buttonType="textonly" size="md" onClick={handleCancel}>
                  Cancel
                </GameButton>
              )}
              <GameButton
                size="lg"
                onClick={handleConfirm}
                disabled={!isValidSelection || totalCost > playerCredits}
                className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
              >
                {showConfirmation
                  ? "Confirm Skip"
                  : selection.source === "sell-patents"
                    ? "Sell Cards"
                    : "Confirm Selection"}
              </GameButton>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PendingCardSelectionOverlay;
