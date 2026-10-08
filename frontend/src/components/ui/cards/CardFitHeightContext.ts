import { createContext } from "react";

/** Design width of a compact `GameCard`; fitted cards are laid out at this width and scaled. */
export const GAME_CARD_NATURAL_WIDTH = 200;

/** Height a card row gives each card in compact mode; `null` renders cards at design size. */
export const CardFitHeightContext = createContext<number | null>(null);
