import React, { useState, useMemo } from "react";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import GameCard from "../cards/GameCard.tsx";
import { GameModal, GameModalHeader, GameModalContent } from "../GameModal";

interface CardsPlayedModalProps {
  isVisible: boolean;
  onClose: () => void;
  cards: CardDto[];
}

const CardsPlayedModal: React.FC<CardsPlayedModalProps> = ({ isVisible, onClose, cards }) => {
  const [searchQuery, setSearchQuery] = useState("");

  const totalCost = useMemo(() => cards.reduce((sum, card) => sum + card.cost, 0), [cards]);

  const filteredCards = useMemo(() => {
    if (!searchQuery.trim()) return cards;
    const query = searchQuery.toLowerCase();

    const matchesCard = (card: CardDto): boolean => {
      // Name, description, type, pack
      if (card.name.toLowerCase().includes(query)) return true;
      if (card.description?.toLowerCase().includes(query)) return true;
      if (card.type?.toLowerCase().includes(query)) return true;
      if (String(card.cost).includes(query)) return true;

      // Tags
      if (card.tags?.some((tag) => tag.toLowerCase().includes(query))) return true;

      // Requirements
      if (card.requirements?.description?.toLowerCase().includes(query)) return true;
      if (
        card.requirements?.items?.some((req) => {
          if (req.type?.toLowerCase().includes(query)) return true;
          if (req.tag?.toLowerCase().includes(query)) return true;
          if (req.resource?.toLowerCase().includes(query)) return true;
          return false;
        })
      )
        return true;

      // Behaviors (descriptions, resource types in inputs/outputs)
      if (
        card.behaviors?.some((b) => {
          if (b.description?.toLowerCase().includes(query)) return true;
          if (b.inputs?.some((io) => io.type?.toLowerCase().includes(query))) return true;
          if (b.outputs?.some((io) => io.type?.toLowerCase().includes(query))) return true;
          return false;
        })
      )
        return true;

      // Resource storage
      if (card.resourceStorage?.type?.toLowerCase().includes(query)) return true;
      if (card.resourceStorage?.description?.toLowerCase().includes(query)) return true;

      // VP conditions
      if (
        card.vpConditions?.some((vp) => {
          if (vp.condition?.toLowerCase().includes(query)) return true;
          if (vp.description?.toLowerCase().includes(query)) return true;
          return false;
        })
      )
        return true;

      return false;
    };

    return cards.filter(matchesCard);
  }, [cards, searchQuery]);

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
      className="bg-black/50 border border-[var(--modal-accent)]/40 rounded-md text-white py-1.5 px-3 text-sm w-[200px] placeholder:text-white/40 outline-none focus:border-[var(--modal-accent)]/70"
    />
  );

  return (
    <GameModal
      isVisible={isVisible}
      onClose={onClose}
      theme="cardsPlayed"
      size="full"
      className="h-[90vh]"
    >
      <GameModalHeader
        title="Played Cards"
        stats={statsContent}
        controls={controlsContent}
        onClose={onClose}
      />

      <GameModalContent>
        {filteredCards.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <h3 className="text-white/70 text-lg font-orbitron m-0">No Cards Found</h3>
          </div>
        ) : (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(215px,1fr))] gap-x-0 gap-y-[50px] justify-items-center pt-[30px]">
            {filteredCards.map((card) => (
              <div key={card.id} className="w-full max-w-[240px]">
                <GameCard card={card} isSelected={false} onSelect={() => {}} animationDelay={-1} />
              </div>
            ))}
          </div>
        )}
      </GameModalContent>
    </GameModal>
  );
};

export default CardsPlayedModal;
