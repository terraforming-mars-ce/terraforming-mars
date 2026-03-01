import React from "react";
import GameCard from "../cards/GameCard.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import { useCardSelection } from "../../../hooks/useCardSelection.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
  RESOURCE_LABEL_CLASS,
  RESOURCE_DISPLAY_CLASS,
} from "./overlayStyles.ts";
import GameMenuButton from "../buttons/GameMenuButton.tsx";

interface StartingCardSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  playerCredits: number;
  onSelectCards: (selectedCardIds: string[]) => void;
}

const StartingCardSelectionOverlay: React.FC<StartingCardSelectionOverlayProps> = ({
  isOpen,
  cards,
  playerCredits,
  onSelectCards,
}) => {
  const {
    selectedCardIds,
    totalCost,
    showConfirmation,
    isValidSelection,
    handleCardSelect,
    handleConfirm: handleCardConfirm,
  } = useCardSelection({
    cards,
    isOpen,
    playerCredits,
    costPerCard: 3,
    minCards: 0,
  });

  if (!isOpen || cards.length === 0) return null;

  const handleConfirm = () => {
    handleCardConfirm((cardIds) => {
      onSelectCards(cardIds);
    });
  };

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      <div className={OVERLAY_CONTAINER_CLASS}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Select Starting Cards</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Choose your starting cards. Each card costs 3 MC.
          </p>
        </div>

        <div className="flex-1 overflow-x-auto overflow-y-hidden p-8 flex items-center bg-[radial-gradient(ellipse_at_center,rgba(139,69,19,0.1)_0%,transparent_70%)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-white/20 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-white/30 max-[768px]:p-5">
          <div className="flex gap-6 mx-auto py-5 max-[768px]:gap-4">
            {cards.map((card, index) => {
              const isSelected = selectedCardIds.includes(card.id);

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

        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between">
            <div className={RESOURCE_DISPLAY_CLASS}>
              <span className={RESOURCE_LABEL_CLASS}>Your Credits:</span>
              <GameIcon iconType={ResourceTypeCredit} amount={playerCredits} size="large" />
            </div>
            <div className={RESOURCE_DISPLAY_CLASS}>
              <span className={RESOURCE_LABEL_CLASS}>Total Cost:</span>
              {totalCost > 0 ? (
                <>
                  <GameIcon iconType={ResourceTypeCredit} amount={totalCost} size="large" />
                  <span className="text-white/70 text-sm">
                    ({selectedCardIds.length} card
                    {selectedCardIds.length !== 1 ? "s" : ""} selected)
                  </span>
                </>
              ) : (
                <span className="!text-[#4caf50] font-bold tracking-[1px]">FREE</span>
              )}
            </div>
          </div>

          <div className="flex items-center gap-4 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3">
            {showConfirmation && (
              <div className="text-sm">
                <span className="text-[#ff9800]">
                  Are you sure you don't want to select any cards?
                </span>
              </div>
            )}

            <GameMenuButton
              variant="primary"
              size="lg"
              onClick={handleConfirm}
              disabled={!isValidSelection}
              className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
            >
              {showConfirmation ? "Confirm Skip" : "Confirm Selection"}
            </GameMenuButton>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StartingCardSelectionOverlay;
