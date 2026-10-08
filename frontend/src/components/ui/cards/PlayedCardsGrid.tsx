import { useMemo } from "react";
import { CardDto } from "@/types/generated/api-types.ts";
import { cardDescriptionToPlainText } from "@/utils/cardDescription.ts";
import GameCard from "./GameCard.tsx";

function matchesCard(card: CardDto, query: string): boolean {
  if (card.name.toLowerCase().includes(query)) {
    return true;
  }
  if (cardDescriptionToPlainText(card.description).toLowerCase().includes(query)) {
    return true;
  }
  if (card.type?.toLowerCase().includes(query)) {
    return true;
  }
  if (String(card.cost).includes(query)) {
    return true;
  }
  if (card.tags?.some((tag) => tag.toLowerCase().includes(query))) {
    return true;
  }
  if (card.requirements?.description?.toLowerCase().includes(query)) {
    return true;
  }
  if (
    card.requirements?.items?.some(
      (req) =>
        req.type?.toLowerCase().includes(query) ||
        req.tag?.toLowerCase().includes(query) ||
        req.resource?.toLowerCase().includes(query),
    )
  ) {
    return true;
  }
  if (
    card.behaviors?.some(
      (b) =>
        b.description?.toLowerCase().includes(query) ||
        b.inputs?.some((io) => io.type?.toLowerCase().includes(query)) ||
        b.outputs?.some((io) => io.type?.toLowerCase().includes(query)),
    )
  ) {
    return true;
  }
  if (card.resourceStorage?.type?.toLowerCase().includes(query)) {
    return true;
  }
  if (card.resourceStorage?.description?.toLowerCase().includes(query)) {
    return true;
  }
  return (
    card.vpConditions?.some(
      (vp) =>
        vp.condition?.toLowerCase().includes(query) ||
        vp.description?.toLowerCase().includes(query),
    ) ?? false
  );
}

export function filterPlayedCards(cards: CardDto[], searchQuery: string): CardDto[] {
  if (!searchQuery.trim()) {
    return cards;
  }
  const query = searchQuery.toLowerCase();
  return cards.filter((card) => matchesCard(card, query));
}

interface PlayedCardsGridProps {
  cards: CardDto[];
  searchQuery?: string;
}

export default function PlayedCardsGrid({ cards, searchQuery = "" }: PlayedCardsGridProps) {
  const filteredCards = useMemo(() => filterPlayedCards(cards, searchQuery), [cards, searchQuery]);

  if (filteredCards.length === 0) {
    return <NoPlayedCards />;
  }

  return (
    <div className="grid grid-cols-[repeat(auto-fill,minmax(215px,1fr))] gap-x-0 gap-y-[50px] justify-items-center pt-[30px]">
      {filteredCards.map((card) => (
        <div key={card.id} className="card-size">
          <GameCard card={card} isSelected={false} />
        </div>
      ))}
    </div>
  );
}

export function NoPlayedCards() {
  return (
    <div className="flex items-center justify-center h-full">
      <h3 className="text-white/70 text-lg font-orbitron m-0">No Cards Found</h3>
    </div>
  );
}
