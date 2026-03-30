import React, { useEffect, useState } from "react";
import GameCard from "../cards/GameCard.tsx";
import {
  PendingCardDiscardSelectionDto,
  PlayerCardDto,
} from "../../../types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
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
} from "./overlayStyles.ts";
import GameButton from "../buttons/GameButton.tsx";

interface CardDiscardSelectionOverlayProps {
  isOpen: boolean;
  selection: PendingCardDiscardSelectionDto;
  handCards: PlayerCardDto[];
  onConfirm: (cardsToDiscard: string[]) => void;
}

const CardDiscardSelectionOverlay: React.FC<CardDiscardSelectionOverlayProps> = ({
  isOpen,
  selection,
  handCards,
  onConfirm,
}) => {
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [showConfirmation, setShowConfirmation] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setSelectedCardIds([]);
      setShowConfirmation(false);
    }
  }, [isOpen]);

  if (!isOpen || handCards.length === 0) return null;

  const isOptional = selection.minCards === 0;

  const handleCardSelect = (cardId: string) => {
    if (showConfirmation) {
      setShowConfirmation(false);
    }

    if (selectedCardIds.includes(cardId)) {
      setSelectedCardIds((prev) => prev.filter((id) => id !== cardId));
      return;
    }

    if (selectedCardIds.length < selection.maxCards) {
      setSelectedCardIds((prev) => [...prev, cardId]);
    }
  };

  const handleConfirm = () => {
    if (selectedCardIds.length === 0 && isOptional && !showConfirmation) {
      setShowConfirmation(true);
      return;
    }

    if (selectedCardIds.length < selection.minCards) {
      return;
    }

    onConfirm(selectedCardIds);
  };

  const handleSkip = () => {
    if (!showConfirmation) {
      setShowConfirmation(true);
      return;
    }
    onConfirm([]);
  };

  return (
    <div
      className="fixed inset-0 flex items-center justify-center"
      style={{ zIndex: Z_INDEX.CORPORATION_SELECTION }}
    >
      <div className={OVERLAY_BACKDROP_BLUR_CLASS} />
      <div className={OVERLAY_BACKDROP_TINT_CLASS} />

      <div className={OVERLAY_CONTAINER_CLASS}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Discard to Draw</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            {selection.source}: Select up to {selection.maxCards} card
            {selection.maxCards !== 1 ? "s" : ""} to discard
            {isOptional ? " (optional)" : ""}.
            {selectedCardIds.length > 0
              ? ` You will draw ${selectedCardIds.length} new card${selectedCardIds.length !== 1 ? "s" : ""}.`
              : ""}
          </p>
        </div>

        <div className={OVERLAY_CARDS_CONTAINER_CLASS}>
          <div className={OVERLAY_CARDS_INNER_CLASS}>
            {handCards.map((card, index) => {
              const isSelected = selectedCardIds.includes(card.id);
              return (
                <div key={card.id} className="relative">
                  <GameCard
                    card={card}
                    isSelected={isSelected}
                    onSelect={handleCardSelect}
                    animationDelay={index * 100}
                    showCheckbox={true}
                  />
                </div>
              );
            })}
          </div>
        </div>

        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="text-sm">
            {selectedCardIds.length === 0 ? (
              showConfirmation ? (
                <span className="text-[#ff9800]">Are you sure you want to skip?</span>
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
            {isOptional && (
              <GameButton buttonType="textonly" size="md" onClick={handleSkip}>
                {showConfirmation && selectedCardIds.length === 0 ? "Confirm Skip" : "Skip"}
              </GameButton>
            )}
            <GameButton
              size="lg"
              onClick={handleConfirm}
              disabled={selectedCardIds.length < selection.minCards}
              className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
            >
              {showConfirmation && selectedCardIds.length === 0
                ? "Confirm Skip"
                : selectedCardIds.length === 0
                  ? "Confirm"
                  : `Discard ${selectedCardIds.length} & Draw`}
            </GameButton>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CardDiscardSelectionOverlay;
