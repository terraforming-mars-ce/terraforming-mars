import React, { useEffect, useState } from "react";
import GameCard from "../cards/GameCard.tsx";
import GameIcon from "../display/GameIcon.tsx";
import {
  PendingCardDrawSelectionDto,
  ResourceTypeCredit,
} from "../../../types/generated/api-types.ts";
import {
  OVERLAY_BACKGROUND_CLASS,
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

interface CardDrawSelectionOverlayProps {
  isOpen: boolean;
  selection: PendingCardDrawSelectionDto;
  playerCredits: number;
  onConfirm: (cardsToTake: string[], cardsToBuy: string[]) => void;
}

const CardDrawSelectionOverlay: React.FC<CardDrawSelectionOverlayProps> = ({
  isOpen,
  selection,
  playerCredits,
  onConfirm,
}) => {
  const [cardsToTake, setCardsToTake] = useState<string[]>([]);
  const [cardsToBuy, setCardsToBuy] = useState<string[]>([]);
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [showPlayability, setShowPlayability] = useState(false);

  // Initialize selection when overlay opens
  useEffect(() => {
    if (isOpen && selection.availableCards.length > 0) {
      setCardsToTake([]);
      setCardsToBuy([]);
      setShowConfirmation(false);
    }
  }, [isOpen, selection.availableCards]);

  if (!isOpen || selection.availableCards.length === 0) return null;

  // Pure card-draw: All shown cards must be taken (no choice)
  // Peek+Draw/Take: Some cards must/can be taken (player has choice)
  const isCardDraw =
    selection.maxBuyCount === 0 && selection.freeTakeCount === selection.availableCards.length;

  const getTitleAndDescription = (): {
    title: string;
    description: string;
  } => {
    if (isCardDraw) {
      return {
        title: "New cards",
        description: "",
      };
    }

    if (selection.playAsPrelude) {
      return {
        title: "Select prelude",
        description: "Select 1 prelude card to play",
      };
    }

    // For all peek/take/buy scenarios, use consistent "Select cards" title
    const maxCards = selection.freeTakeCount + selection.maxBuyCount;
    return {
      title: "Select cards",
      description: `Select up to ${maxCards} card${maxCards !== 1 ? "s" : ""}`,
    };
  };

  const canAffordBuy = (): boolean => {
    const currentBuyCost = cardsToBuy.length * selection.cardBuyCost;
    return currentBuyCost + selection.cardBuyCost <= playerCredits;
  };

  const handleCardSelect = (cardId: string) => {
    // For card-draw scenarios, auto-select all cards
    if (isCardDraw) {
      return;
    }

    // Reset confirmation when user selects cards
    if (showConfirmation) {
      setShowConfirmation(false);
    }

    // If card is already in take list, remove it
    if (cardsToTake.includes(cardId)) {
      setCardsToTake((prev) => prev.filter((id) => id !== cardId));
      return;
    }

    // If card is already in buy list, remove it
    if (cardsToBuy.includes(cardId)) {
      setCardsToBuy((prev) => prev.filter((id) => id !== cardId));
      return;
    }

    // Try to add to free take list first
    if (cardsToTake.length < selection.freeTakeCount) {
      setCardsToTake((prev) => [...prev, cardId]);
    } else if (cardsToBuy.length < selection.maxBuyCount && canAffordBuy()) {
      // Otherwise add to buy list if we can afford it
      setCardsToBuy((prev) => [...prev, cardId]);
    }
  };

  const handleConfirm = () => {
    // For card-draw, automatically select all cards
    if (isCardDraw) {
      const allCardIds = selection.availableCards.map((c) => c.id);
      onConfirm(allCardIds, []);
      return;
    }

    const totalSelected = cardsToTake.length + cardsToBuy.length;
    const maxAllowed = selection.freeTakeCount + selection.maxBuyCount;

    if (totalSelected > maxAllowed) {
      return; // Invalid selection
    }

    // Require confirmation in two scenarios:
    // 1. Discarding all cards (totalSelected === 0)
    // 2. Not taking all available free cards (cardsToTake.length < freeTakeCount)
    const needsConfirmation = totalSelected === 0 || cardsToTake.length < selection.freeTakeCount;

    if (needsConfirmation && !showConfirmation) {
      // First click - show confirmation
      setShowConfirmation(true);
      return;
    }

    // Second click or no confirmation needed - proceed with selection
    onConfirm(cardsToTake, cardsToBuy);
  };

  const getButtonText = (): string => {
    if (isCardDraw) {
      return "Return";
    }

    // For peek/take/buy scenarios
    const totalSelected = cardsToTake.length + cardsToBuy.length;

    // Check if confirmation is needed
    const needsConfirmation = totalSelected === 0 || cardsToTake.length < selection.freeTakeCount;

    if (needsConfirmation && showConfirmation) {
      return "Confirm Selection";
    }

    if (totalSelected === 0) {
      return "Discard all";
    }

    // Show buy count if any cards are being bought
    if (cardsToBuy.length > 0) {
      return cardsToBuy.length === 1 ? "Buy 1 card" : `Buy ${cardsToBuy.length} cards`;
    }

    // Otherwise just confirm the free selection
    return "Confirm Selection";
  };

  const titleInfo = getTitleAndDescription();
  const totalBuyCost = cardsToBuy.length * selection.cardBuyCost;
  const totalSelected = cardsToTake.length + cardsToBuy.length;
  // For peek scenarios, allow any selection from 0 to max (including discarding all)
  const isValidSelection =
    isCardDraw || totalSelected <= selection.freeTakeCount + selection.maxBuyCount;

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      {/* Translucent background */}
      <div className={OVERLAY_BACKGROUND_CLASS} />

      {/* Content container */}
      <div className={OVERLAY_CONTAINER_CLASS}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <div className="flex items-center justify-between w-full">
            <div>
              <h2 className={OVERLAY_TITLE_CLASS}>{titleInfo.title}</h2>
              {titleInfo.description && (
                <p className={OVERLAY_DESCRIPTION_CLASS}>{titleInfo.description}</p>
              )}
            </div>
            {!isCardDraw && (
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
              const isSelected = cardsToTake.includes(card.id) || cardsToBuy.includes(card.id);

              return (
                <div key={card.id} className="relative">
                  <GameCard
                    card={card}
                    isSelected={isSelected}
                    onSelect={handleCardSelect}
                    animationDelay={index * 100}
                    showCheckbox={!isCardDraw}
                  />
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
          {!selection.playAsPrelude && (
            <div className="flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between max-[768px]:flex-wrap">
              <div className="flex items-center gap-3">
                <span className={RESOURCE_LABEL_CLASS}>Your Credits:</span>
                <GameIcon iconType={ResourceTypeCredit} amount={playerCredits} size="large" />
              </div>
              {totalBuyCost > 0 && (
                <div className="flex items-center gap-3">
                  <span className={RESOURCE_LABEL_CLASS}>Buy Cost:</span>
                  <GameIcon iconType={ResourceTypeCredit} amount={totalBuyCost} size="large" />
                </div>
              )}
            </div>
          )}

          <div className="flex items-center gap-6 ml-auto max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3">
            <div className="text-sm">
              {isCardDraw ? (
                <span className="text-white/70">
                  Drawing {selection.freeTakeCount} card
                  {selection.freeTakeCount !== 1 ? "s" : ""}
                </span>
              ) : showConfirmation ? (
                <span className="text-[#ff9800]">
                  {totalSelected === 0
                    ? "Are you sure you want to discard all?"
                    : (() => {
                        const remainingFreeTakes = selection.freeTakeCount - cardsToTake.length;
                        return `You can take ${remainingFreeTakes} more card${remainingFreeTakes !== 1 ? "s" : ""} for free. Confirm?`;
                      })()}
                </span>
              ) : (
                (() => {
                  const discardCount = selection.availableCards.length - totalSelected;
                  return discardCount > 0 ? (
                    <span className="text-white/70">
                      Discard {discardCount} card
                      {discardCount !== 1 ? "s" : ""}
                    </span>
                  ) : null;
                })()
              )}
            </div>
            <div className="flex gap-3 items-center">
              <GameButton
                size="lg"
                onClick={handleConfirm}
                disabled={!isValidSelection || totalBuyCost > playerCredits}
                className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
              >
                {getButtonText()}
              </GameButton>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CardDrawSelectionOverlay;
