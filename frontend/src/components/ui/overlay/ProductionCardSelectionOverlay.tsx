import React from "react";
import GameCard from "../cards/GameCard.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import { useCardSelection } from "../../../hooks/useCardSelection.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_BACKGROUND_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_CARDS_CONTAINER_CLASS,
  OVERLAY_CARDS_INNER_CLASS,
  OVERLAY_FOOTER_CLASS,
  OVERLAY_FOOTER_LEFT_CLASS,
  OVERLAY_FOOTER_RIGHT_CLASS,
  RESOURCE_LABEL_CLASS,
  RESOURCE_DISPLAY_CLASS,
} from "./overlayStyles.ts";
import GameMenuButton from "../buttons/GameMenuButton.tsx";

interface ProductionCardSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  playerCredits: number;
  onSelectCards: (selectedCardIds: string[]) => void;
  onReturn: () => void;
  initialSelectedCardIds?: string[];
  onSelectionChange?: (ids: string[]) => void;
}

const ProductionCardSelectionOverlay: React.FC<ProductionCardSelectionOverlayProps> = ({
  isOpen,
  cards,
  playerCredits,
  onSelectCards,
  onReturn,
  initialSelectedCardIds,
  onSelectionChange,
}) => {
  const {
    selectedCardIds,
    totalCost,
    showConfirmation,
    isValidSelection,
    handleCardSelect,
    handleConfirm,
  } = useCardSelection({
    cards,
    isOpen,
    playerCredits,
    costPerCard: 3,
    minCards: 0,
    initialSelectedCardIds,
    onSelectionChange,
  });

  if (!isOpen || cards.length === 0) return null;

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      {/* Translucent background */}
      <div className={OVERLAY_BACKGROUND_CLASS} />

      {/* Content container */}
      <div className={OVERLAY_CONTAINER_CLASS}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Select Cards to Buy</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Choose cards to buy for your next turn. Each card costs 3 MC.
          </p>
        </div>

        {/* Cards display */}
        <div className={OVERLAY_CARDS_CONTAINER_CLASS}>
          <div className={OVERLAY_CARDS_INNER_CLASS}>
            {cards.map((card, index) => {
              const cardIndex = selectedCardIds.indexOf(card.id);
              const isSelected = cardIndex !== -1;

              return (
                <GameCard
                  key={card.id}
                  card={card}
                  isSelected={isSelected}
                  onSelect={handleCardSelect}
                  animationDelay={index * 100}
                  showCheckbox={true}
                />
              );
            })}
          </div>
        </div>

        {/* Footer with cost and confirm button */}
        <div className={OVERLAY_FOOTER_CLASS}>
          <div className={OVERLAY_FOOTER_LEFT_CLASS}>
            <div className={RESOURCE_DISPLAY_CLASS}>
              <span className={RESOURCE_LABEL_CLASS}>Your Credits:</span>
              <GameIcon iconType={ResourceTypeCredit} amount={playerCredits} size="large" />
            </div>
            <div className={RESOURCE_DISPLAY_CLASS}>
              <span className={RESOURCE_LABEL_CLASS}>Total Cost:</span>
              {totalCost > 0 ? (
                <GameIcon iconType={ResourceTypeCredit} amount={totalCost} size="large" />
              ) : (
                <span className="!text-[#4caf50] font-bold tracking-[1px]">FREE</span>
              )}
            </div>
          </div>

          <div className={OVERLAY_FOOTER_RIGHT_CLASS}>
            <div className="text-sm">
              {selectedCardIds.length === 0 ? (
                showConfirmation ? (
                  <span className="text-[#ff9800]">
                    Are you sure you don't want to buy any cards?
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
              <GameMenuButton variant="text" size="md" onClick={onReturn}>
                Hide
              </GameMenuButton>
              <GameMenuButton
                variant="primary"
                size="lg"
                onClick={() => handleConfirm(onSelectCards)}
                disabled={!isValidSelection}
                className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
              >
                {showConfirmation
                  ? "Confirm Skip"
                  : selectedCardIds.length === 0
                    ? "Skip Buy Cards"
                    : "Buy Cards"}
              </GameMenuButton>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProductionCardSelectionOverlay;
