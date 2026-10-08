import type { PlayerCardDto } from "@/types/generated/api-types.ts";
import { getZIndex } from "@/constants/zIndex.ts";
import MobileCardDetail from "../MobileCardDetail.tsx";
import GameButton from "../../ui/buttons/GameButton.tsx";
import CardStatusMessages from "../../ui/display/CardStatusMessages.tsx";
import { useHorizontalSwipe } from "./useHorizontalSwipe.ts";

interface HandCardDetailProps {
  card: PlayerCardDto;
  showPlay: boolean;
  onPlay: (card: PlayerCardDto) => void;
  onClose: () => void;
  onPrevious?: () => void;
  onNext?: () => void;
}

function Chevron({ direction }: { direction: "left" | "right" }) {
  const points = direction === "left" ? "15 18 9 12 15 6" : "9 18 15 12 9 6";
  return (
    <svg
      width="22"
      height="22"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <polyline points={points} />
    </svg>
  );
}

function EdgeButton({
  direction,
  label,
  onClick,
}: {
  direction: "left" | "right";
  label: string;
  onClick: () => void;
}) {
  const side =
    direction === "left"
      ? { left: "calc(var(--safe-left) + 4px)" }
      : { right: "calc(var(--safe-right) + 4px)" };
  return (
    <button
      type="button"
      aria-label={label}
      className="fixed top-1/2 -translate-y-1/2 w-11 h-11 flex items-center justify-center rounded-full bg-black/65 border border-white/25 text-white cursor-pointer active:bg-white/15"
      style={{ ...side, zIndex: getZIndex("MOBILE_CARD_DETAIL", 1) }}
      onClick={onClick}
    >
      <Chevron direction={direction} />
    </button>
  );
}

export default function HandCardDetail({
  card,
  showPlay,
  onPlay,
  onClose,
  onPrevious,
  onNext,
}: HandCardDetailProps) {
  const swipe = useHorizontalSwipe({ onSwipeLeft: onNext, onSwipeRight: onPrevious });
  const hasStatus = !card.available || (card.warnings?.length ?? 0) > 0;

  const actions = showPlay ? (
    <GameButton size="lg" disabled={!card.available} onClick={() => onPlay(card)}>
      Play
    </GameButton>
  ) : undefined;

  return (
    <div {...swipe}>
      <MobileCardDetail
        card={card}
        notice={hasStatus ? <CardStatusMessages card={card} /> : undefined}
        actions={actions}
        onClose={onClose}
      />
      {onPrevious && <EdgeButton direction="left" label="Previous card" onClick={onPrevious} />}
      {onNext && <EdgeButton direction="right" label="Next card" onClick={onNext} />}
    </div>
  );
}
