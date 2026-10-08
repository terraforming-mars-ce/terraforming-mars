import React from "react";
import CardChoice from "../cards/CardChoice.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import { useCardSelection } from "../../../hooks/useCardSelection.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_CONTAINER_STYLE,
  OVERLAY_BACKDROP_BLUR_CLASS,
  OVERLAY_BACKDROP_TINT_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_FOOTER_CLASS,
  OVERLAY_FOOTER_LEFT_CLASS,
  OVERLAY_FOOTER_RIGHT_CLASS,
  RESOURCE_LABEL_CLASS,
  RESOURCE_DISPLAY_CLASS,
  OVERLAY_ACTION_BUTTON_CLASS,
  OVERLAY_ROOT_SAFE_AREA_CLASS,
} from "./overlayStyles.ts";
import CardSelectionRow from "./CardSelectionRow.tsx";
import OverlayDescription from "./OverlayDescription.tsx";
import GameButton from "../buttons/GameButton.tsx";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

interface ProductionCardSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  playerCredits: number;
  costPerCard: number;
  allowRandomBuy?: boolean;
  onSelectCards: (selectedCardIds: string[], options?: { randomBuy?: boolean }) => void;
  onReturn: () => void;
  initialSelectedCardIds?: string[];
  onSelectionChange?: (ids: string[]) => void;
}

const ProductionCardSelectionOverlay: React.FC<ProductionCardSelectionOverlayProps> = ({
  isOpen,
  cards,
  playerCredits,
  costPerCard,
  allowRandomBuy = false,
  onSelectCards,
  onReturn,
  initialSelectedCardIds,
  onSelectionChange,
}) => {
  const { isCompact } = useLayoutMode();
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
    costPerCard,
    minCards: 0,
    initialSelectedCardIds,
    onSelectionChange,
  });

  if (!isOpen || cards.length === 0) {
    return null;
  }

  const iconSize = isCompact ? "medium" : "large";

  return (
    <div
      className={`fixed inset-0 flex items-center justify-center ${OVERLAY_ROOT_SAFE_AREA_CLASS}`}
      style={{ zIndex: Z_INDEX.CORPORATION_SELECTION }}
    >
      <div className={OVERLAY_BACKDROP_BLUR_CLASS} />
      <div className={OVERLAY_BACKDROP_TINT_CLASS} />

      {/* Content container */}
      <div className={OVERLAY_CONTAINER_CLASS} style={OVERLAY_CONTAINER_STYLE}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Select Cards to Buy</h2>
          <OverlayDescription>
            Choose cards to buy for your next turn. Each card costs {costPerCard} MC.
          </OverlayDescription>
        </div>

        {/* Cards display */}
        <CardSelectionRow>
          {cards.map((card, index) => {
            const cardIndex = selectedCardIds.indexOf(card.id);
            const isSelected = cardIndex !== -1;

            return (
              <CardChoice
                key={card.id}
                card={card}
                isSelected={isSelected}
                onSelect={handleCardSelect}
                animationDelay={index * 100}
                showCheckbox={true}
              />
            );
          })}
        </CardSelectionRow>

        {/* Footer with cost and confirm button */}
        <div className={OVERLAY_FOOTER_CLASS}>
          <div className={OVERLAY_FOOTER_LEFT_CLASS}>
            <div className={RESOURCE_DISPLAY_CLASS}>
              <span className={RESOURCE_LABEL_CLASS}>
                {isCompact ? "Credits" : "Available payment:"}
              </span>
              <GameIcon iconType={ResourceTypeCredit} amount={playerCredits} size={iconSize} />
            </div>
            <div className={RESOURCE_DISPLAY_CLASS}>
              <span className={RESOURCE_LABEL_CLASS}>{isCompact ? "Cost" : "Total Cost:"}</span>
              {totalCost > 0 ? (
                <GameIcon iconType={ResourceTypeCredit} amount={totalCost} size={iconSize} />
              ) : (
                <span className="!text-[#4caf50] font-bold tracking-[1px]">FREE</span>
              )}
            </div>
          </div>

          <div className={OVERLAY_FOOTER_RIGHT_CLASS}>
            {!isCompact && (
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
            )}
            <div className="flex gap-3 items-center">
              <GameButton
                emphasis="quiet"
                size="md"
                className={OVERLAY_ACTION_BUTTON_CLASS}
                onClick={onReturn}
              >
                Hide
              </GameButton>
              {allowRandomBuy && selectedCardIds.length === 0 && (
                <GameButton
                  emphasis="secondary"
                  size="lg"
                  onClick={() => onSelectCards([], { randomBuy: true })}
                  disabled={playerCredits < costPerCard}
                  className={OVERLAY_ACTION_BUTTON_CLASS}
                >
                  <span className="inline-flex items-center gap-2">
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    >
                      <rect x="3" y="3" width="18" height="18" rx="2" />
                      <circle cx="8" cy="8" r="1" fill="currentColor" />
                      <circle cx="16" cy="8" r="1" fill="currentColor" />
                      <circle cx="8" cy="16" r="1" fill="currentColor" />
                      <circle cx="16" cy="16" r="1" fill="currentColor" />
                      <circle cx="12" cy="12" r="1" fill="currentColor" />
                    </svg>
                    Buy 1 Random
                  </span>
                </GameButton>
              )}
              <GameButton
                size="lg"
                onClick={() => handleConfirm(onSelectCards)}
                disabled={!isValidSelection}
                className={OVERLAY_ACTION_BUTTON_CLASS}
              >
                {showConfirmation
                  ? "Confirm Skip"
                  : selectedCardIds.length === 0
                    ? "Skip Buy Cards"
                    : "Buy Cards"}
              </GameButton>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProductionCardSelectionOverlay;
