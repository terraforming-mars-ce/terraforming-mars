export interface CardPackInfo {
  id: string;
  label: string;
  description: string;
  cardCount?: string;
  wip?: boolean;
  lockedOn?: boolean;
}

export const CARD_PACKS: CardPackInfo[] = [
  {
    id: "base-game",
    label: "Base Game",
    cardCount: "147 cards",
    description:
      "Includes tested cards with comprehensive test coverage. All cards have verified implementations.",
    lockedOn: true,
  },
  {
    id: "prelude",
    label: "Prelude",
    cardCount: "47 cards",
    description:
      "Each player receives 4 prelude cards and keeps 2 for free. These give an early boost with resources, production, or other effects.",
  },
  {
    id: "colonies",
    label: "Colonies",
    cardCount: "54 cards",
    description:
      "Adds colony tiles that players can trade with and build on. Trade costs 3 energy, building costs 17 MC.",
  },
  {
    id: "project-funding",
    label: "Project Funding",
    description:
      "Themed projects players invest in by purchasing seats. Funders get tier-based rewards and all players get a completion effect.",
    wip: true,
  },
  {
    id: "experimental",
    label: "Experimental",
    cardCount: "4 cards",
    description:
      "Experimental cards with new mechanics: extra actions, bonus tags, special tiles, and tile destruction.",
  },
];

export const VENUS_PACK: CardPackInfo = {
  id: "venus-next",
  label: "Venus Next",
  cardCount: "54 cards",
  description:
    "Adds the Venus globe with tile placements and the Venus global parameter track. Venus Next expansion cards are automatically included.",
};
