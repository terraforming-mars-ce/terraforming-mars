import React, { useState, useMemo } from "react";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import PlayedCardsGrid from "../cards/PlayedCardsGrid.tsx";
import { GameModal, GameModalHeader, GameModalContent } from "../GameModal";

interface CardsPlayedModalProps {
  isVisible: boolean;
  onClose: () => void;
  cards: CardDto[];
}

const CardsPlayedModal: React.FC<CardsPlayedModalProps> = ({ isVisible, onClose, cards }) => {
  const [searchQuery, setSearchQuery] = useState("");

  const totalCost = useMemo(() => cards.reduce((sum, card) => sum + card.cost, 0), [cards]);

  const statsContent = (
    <div className="flex items-center gap-3">
      <div className="text-white/80 text-xs bg-[rgba(150,100,255,0.2)] py-1 px-2.5 rounded-md border border-[rgba(150,100,255,0.3)]">
        {cards.length} cards played
      </div>
      <GameIcon iconType={ResourceTypeCredit} amount={totalCost} size="medium" />
    </div>
  );

  const controlsContent = (
    <input
      type="text"
      value={searchQuery}
      onChange={(e) => setSearchQuery(e.target.value)}
      placeholder="Search cards..."
      spellCheck={false}
      autoComplete="off"
      className="bg-black/50 border border-[var(--modal-accent)]/40 rounded-none text-white py-1.5 px-3 text-sm w-[200px] placeholder:text-white/40 outline-none focus:border-[var(--modal-accent)]/70"
    />
  );

  return (
    <GameModal
      isVisible={isVisible}
      onClose={onClose}
      theme="cardsPlayed"
      size="full"
      className="h-[90dvh]"
    >
      <GameModalHeader
        title="Played Cards"
        stats={statsContent}
        controls={controlsContent}
        onClose={onClose}
      />

      <GameModalContent>
        <PlayedCardsGrid cards={cards} searchQuery={searchQuery} />
      </GameModalContent>
    </GameModal>
  );
};

export default CardsPlayedModal;
