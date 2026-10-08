import { useState } from "react";
import type { PlayerCardDto } from "@/types/generated/api-types.ts";

function compareHandCards(a: PlayerCardDto, b: PlayerCardDto): number {
  if (a.available !== b.available) {
    return a.available ? -1 : 1;
  }
  if (a.effectiveCost !== b.effectiveCost) {
    return a.effectiveCost - b.effectiveCost;
  }
  return a.name.localeCompare(b.name);
}

function handKey(cards: PlayerCardDto[]): string {
  return cards
    .map((card) => card.id)
    .sort()
    .join("|");
}

function sortedIds(cards: PlayerCardDto[]): string[] {
  return [...cards].sort(compareHandCards).map((card) => card.id);
}

/** Sorts playable first, then cost, then name; re-sorts only when the set of cards changes. */
export function useStableHandOrder(cards: PlayerCardDto[]): PlayerCardDto[] {
  const key = handKey(cards);
  const [order, setOrder] = useState(() => ({ key, ids: sortedIds(cards) }));

  let ids = order.ids;
  if (order.key !== key) {
    ids = sortedIds(cards);
    setOrder({ key, ids });
  }

  const byId = new Map(cards.map((card) => [card.id, card]));
  return ids.flatMap((id) => {
    const card = byId.get(id);
    return card ? [card] : [];
  });
}
