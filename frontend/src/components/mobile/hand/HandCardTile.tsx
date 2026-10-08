import { useEffect, useRef } from "react";
import type { PlayerCardDto } from "@/types/generated/api-types.ts";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import GameCard from "../../ui/cards/GameCard.tsx";
import FittedCard from "../../ui/cards/FittedCard.tsx";
import { GAME_CARD_NATURAL_WIDTH } from "../../ui/cards/CardFitHeightContext.ts";

interface HandCardTileProps {
  card: PlayerCardDto;
  height: number | undefined;
  highlighted: boolean;
  onSelect?: (card: PlayerCardDto) => void;
}

const GLOW_FRAMES: Keyframe[] = [
  { boxShadow: "0 0 0 2px #8fb8ca, 0 0 18px 4px rgba(143, 184, 202, 0.75)" },
  { boxShadow: "0 0 0 2px rgba(143, 184, 202, 0), 0 0 18px 4px rgba(143, 184, 202, 0)" },
];
const GLOW_MS = 600;

export default function HandCardTile({ card, height, highlighted, onSelect }: HandCardTileProps) {
  const ref = useRef<HTMLButtonElement>(null);
  const reducedMotion = useReducedMotion();

  useEffect(() => {
    if (!highlighted || reducedMotion) {
      return;
    }
    const animation = ref.current?.animate(GLOW_FRAMES, { duration: GLOW_MS, easing: "ease-out" });
    return () => animation?.cancel();
  }, [highlighted, reducedMotion]);

  const outline = highlighted && reducedMotion ? "outline-2 outline-[#8fb8ca]" : "outline-none";

  return (
    <button
      ref={ref}
      type="button"
      data-hand-card-id={card.id}
      aria-label={card.name}
      className={`shrink-0 snap-start text-left cursor-pointer select-none ${outline} focus-visible:ring-2 focus-visible:ring-[#8fb8ca]`}
      onClick={() => onSelect?.(card)}
    >
      <FittedCard naturalWidth={GAME_CARD_NATURAL_WIDTH} height={height}>
        <GameCard card={card} />
      </FittedCard>
    </button>
  );
}
