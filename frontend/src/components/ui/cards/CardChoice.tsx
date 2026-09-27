import { useEffect, useRef } from "react";
import GameCard, { type GameCardProps } from "./GameCard.tsx";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

interface CardChoiceProps extends Omit<GameCardProps, "presentation" | "description"> {
  onSelect: (cardId: string) => void;
  animationDelay?: number;
}

export default function CardChoice({ onSelect, animationDelay = 0, ...props }: CardChoiceProps) {
  const { playCardHoverSound } = useSoundEffects();
  const pendingSound = useRef(false);
  useEffect(() => {
    if (pendingSound.current) {
      pendingSound.current = false;
      void playCardHoverSound();
    }
  }, [props.isSelected, playCardHoverSound]);
  const activate = () => {
    pendingSound.current = true;
    onSelect(props.card.id);
  };
  return (
    <div
      role="checkbox"
      aria-checked={props.isSelected ?? false}
      aria-label={props.card.name}
      tabIndex={0}
      className={`card-size cursor-pointer select-none outline-none focus-visible:ring-2 focus-visible:ring-blue-400 ${animationDelay >= 0 ? "motion-safe:animate-[fadeInUp_0.5s_ease_both]" : ""}`}
      style={animationDelay >= 0 ? { animationDelay: `${animationDelay}ms` } : undefined}
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
}
