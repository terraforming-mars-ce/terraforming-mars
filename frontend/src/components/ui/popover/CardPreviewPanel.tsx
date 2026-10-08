import { useCardPlayFlowStore } from "@/stores/cardPlayFlowStore";
import React, { useEffect } from "react";
import { createPortal } from "react-dom";
import { CardDto } from "@/types/generated/api-types.ts";
import { cardImage } from "@/assets";
import { Z_INDEX } from "@/constants/zIndex.ts";
import GameCard from "../cards/GameCard.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";
import GameButton from "../buttons/GameButton.tsx";
import { InfoGlyph } from "../display/InfoTooltip.tsx";

const noop = () => {};
const stopPropagation = (event: React.SyntheticEvent) => event.stopPropagation();

const CardPreview: React.FC<{ card: CardDto }> = ({ card }) => {
  if (card.type === "corporation") {
    return (
      <div style={{ transform: "scale(0.55)", transformOrigin: "top left", width: 400, height: 0 }}>
        <CorporationCard
          card={card}
          isSelected={false}
          onSelect={noop}
          showCheckbox={false}
          disableInteraction={true}
        />
      </div>
    );
  }
  return (
    <div className="card-size">
      <GameCard card={card} />
    </div>
  );
};

const CardPreviewPanel: React.FC<{ card: CardDto | null }> = ({ card }) => {
  const inPlayFlow = useCardPlayFlowStore((state) => state.playSession?.phase === "choosing");

  return (
    <div
      className={`
        ${inPlayFlow ? "hidden lg:block absolute right-full top-0 mr-4" : "hidden md:block absolute left-full top-0 ml-4"}
        transition-all duration-200 ease-out
        ${card ? "opacity-100 translate-x-0" : "opacity-0 translate-x-2 pointer-events-none"}
      `}
    >
      {card && <CardPreview card={card} />}
    </div>
  );
};

export const CardPreviewThumbnail: React.FC<{ cardId: string }> = ({ cardId }) => {
  const artwork = cardImage(cardId);
  return (
    <div className="hidden pointer-coarse:block w-10 h-14 shrink-0 overflow-hidden border border-white/20 bg-black/40">
      {artwork && (
        <img
          {...artwork}
          sizes="40px"
          alt=""
          loading="lazy"
          decoding="async"
          className="w-full h-full object-cover"
        />
      )}
    </div>
  );
};

export const CardPreviewButton: React.FC<{ card: CardDto; onPreview: (card: CardDto) => void }> = ({
  card,
  onPreview,
}) => (
  <GameButton
    emphasis="quiet"
    clickSound={false}
    aria-label={`Show ${card.name}`}
    className="hidden pointer-coarse:inline-flex size-11 p-0 shrink-0 text-white/70"
    onClick={(event) => {
      event.stopPropagation();
      onPreview(card);
    }}
  >
    <InfoGlyph />
  </GameButton>
);

export const CardPreviewOverlay: React.FC<{ card: CardDto | null; onClose: () => void }> = ({
  card,
  onClose,
}) => {
  useEffect(() => {
    if (!card) {
      return;
    }
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        event.stopImmediatePropagation();
        onClose();
      }
    };
    window.addEventListener("keydown", handleKeyDown, true);
    return () => window.removeEventListener("keydown", handleKeyDown, true);
  }, [card, onClose]);

  if (!card) {
    return null;
  }

  return createPortal(
    <div
      data-overlay-layer
      role="dialog"
      aria-modal="true"
      aria-label={card.name}
      className="fixed inset-0 flex bg-black/70 p-4 overflow-y-auto overscroll-contain animate-[fadeIn_150ms_ease-in]"
      style={{ zIndex: Z_INDEX.CARD_PREVIEW_OVERLAY }}
      onClick={(event) => {
        event.stopPropagation();
        onClose();
      }}
      onTouchMove={stopPropagation}
      onWheel={stopPropagation}
    >
      <div className="m-auto" onClick={stopPropagation}>
        {card.type === "corporation" ? (
          <div className="w-[400px] max-w-full">
            <CorporationCard
              card={card}
              isSelected={false}
              onSelect={noop}
              showCheckbox={false}
              disableInteraction={true}
            />
          </div>
        ) : (
          <CardPreview card={card} />
        )}
      </div>
    </div>,
    document.body,
  );
};

export default CardPreviewPanel;
