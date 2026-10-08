import { useContext, useEffect, useRef, useState, type ReactNode } from "react";
import GameCard, { type GameCardProps } from "./GameCard.tsx";
import FittedCard from "./FittedCard.tsx";
import { CardFitHeightContext, GAME_CARD_NATURAL_WIDTH } from "./CardFitHeightContext.ts";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";
import { useLongPress } from "@/hooks/useLongPress.ts";
import CardDetailPortal from "../overlay/CardDetailPortal.tsx";

interface CardChoiceProps extends Omit<GameCardProps, "presentation" | "description"> {
  onSelect: (cardId: string) => void;
  animationDelay?: number;
  disabled?: boolean;
  detailNotice?: ReactNode;
}

export default function CardChoice({
  onSelect,
  animationDelay = 0,
  disabled = false,
  detailNotice,
  ...props
}: CardChoiceProps) {
  const { isCompact } = useLayoutMode();
  const fitHeight = useContext(CardFitHeightContext);
  const [detailOpen, setDetailOpen] = useState(false);
  const longPress = useLongPress<HTMLDivElement>(() => setDetailOpen(true));
  const { playCardHoverSound } = useSoundEffects();
  const pendingSound = useRef(false);
  useEffect(() => {
    if (pendingSound.current) {
      pendingSound.current = false;
      void playCardHoverSound();
    }
  }, [props.isSelected, playCardHoverSound]);
  useEffect(() => {
    if (!isCompact) {
      setDetailOpen(false);
    }
  }, [isCompact]);
  const activate = () => {
    if (disabled) {
      return;
    }
    pendingSound.current = true;
    onSelect(props.card.id);
  };
  const animationClass = animationDelay >= 0 ? "motion-safe:animate-[fadeInUp_0.5s_ease_both]" : "";
  const animationStyle =
    animationDelay >= 0 ? { animationDelay: `${animationDelay}ms` } : undefined;
  const fitted = isCompact && fitHeight !== null;
  const choice = (
    <div
      role="checkbox"
      aria-checked={props.isSelected ?? false}
      aria-disabled={disabled}
      aria-label={props.card.name}
      tabIndex={disabled ? -1 : 0}
      className={`card-size ${fitted && props.showCheckbox ? "pb-3" : ""} select-none outline-none focus-visible:ring-2 focus-visible:ring-blue-400 transition-[filter] duration-200 ${disabled ? "cursor-default grayscale brightness-50" : "cursor-pointer"} ${isCompact ? "" : animationClass}`}
      style={isCompact ? undefined : animationStyle}
      {...(isCompact ? longPress : {})}
      onClick={activate}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          activate();
        }
      }}
    >
      <GameCard {...props} />
    </div>
  );

  if (!isCompact) {
    return choice;
  }

  return (
    <div className={`relative ${animationClass}`} style={animationStyle}>
      {fitted ? (
        <FittedCard naturalWidth={GAME_CARD_NATURAL_WIDTH} height={fitHeight}>
          {choice}
        </FittedCard>
      ) : (
        choice
      )}
      {detailOpen && (
        <CardDetailPortal
          card={props.card}
          notice={detailNotice}
          onClose={() => setDetailOpen(false)}
        />
      )}
    </div>
  );
}
