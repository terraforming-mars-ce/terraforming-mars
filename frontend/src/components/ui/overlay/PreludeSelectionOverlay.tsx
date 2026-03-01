import React, { useState, useCallback, useEffect } from "react";
import GameCard from "../cards/GameCard.tsx";
import { CardDto } from "../../../types/generated/api-types.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
} from "./overlayStyles.ts";
import GameMenuButton from "../buttons/GameMenuButton.tsx";

interface PreludeSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  maxSelectable: number;
  onSelectCards: (selectedCardIds: string[]) => void;
}

const PreludeSelectionOverlay: React.FC<PreludeSelectionOverlayProps> = ({
  isOpen,
  cards,
  maxSelectable,
  onSelectCards,
}) => {
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  useEffect(() => {
    if (isOpen) {
      setSelectedIds([]);
    }
  }, [isOpen]);

  const handleCardSelect = useCallback(
    (cardId: string) => {
      setSelectedIds((prev) => {
        if (prev.includes(cardId)) {
          return prev.filter((id) => id !== cardId);
        }
        if (prev.length >= maxSelectable) {
          return prev;
        }
        return [...prev, cardId];
      });
    },
    [maxSelectable],
  );

  const handleConfirm = useCallback(() => {
    if (selectedIds.length === maxSelectable) {
      onSelectCards(selectedIds);
    }
  }, [selectedIds, maxSelectable, onSelectCards]);

  if (!isOpen || cards.length === 0) return null;

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      <div className={OVERLAY_CONTAINER_CLASS}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Select Prelude Cards</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Choose {maxSelectable} prelude cards. These are played for free and give you a head
            start.
          </p>
        </div>

        <div className="flex-1 overflow-x-auto overflow-y-hidden p-8 flex items-center bg-[radial-gradient(ellipse_at_center,rgba(139,69,19,0.1)_0%,transparent_70%)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-white/20 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-white/30 max-[768px]:p-5">
          <div className="flex gap-6 mx-auto py-5 max-[768px]:gap-4">
            {cards.map((card, index) => (
              <GameCard
                key={card.id}
                card={card}
                isSelected={selectedIds.includes(card.id)}
                onSelect={handleCardSelect}
                animationDelay={index * 100}
                showCheckbox={true}
              />
            ))}
          </div>
        </div>

        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="text-sm text-white/70">
            {selectedIds.length} / {maxSelectable} selected
          </div>
          <GameMenuButton
            variant="primary"
            size="lg"
            onClick={handleConfirm}
            disabled={selectedIds.length !== maxSelectable}
            className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
          >
            Confirm Selection
          </GameMenuButton>
        </div>
      </div>
    </div>
  );
};

export default PreludeSelectionOverlay;
