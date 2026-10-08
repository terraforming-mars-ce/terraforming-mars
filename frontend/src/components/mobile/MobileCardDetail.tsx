import { useRef, type ReactNode } from "react";
import {
  CardTypeCorporation,
  type CardDto,
  type PlayerCardDto,
} from "@/types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useRenderPause } from "@/stores/renderPauseStore.ts";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import { useBackDismiss } from "@/hooks/useBackDismiss.ts";
import GameCard from "../ui/cards/GameCard.tsx";
import CorporationCard from "../ui/cards/CorporationCard.tsx";
import CloseButton from "../ui/buttons/CloseButton.tsx";
import FittedCard from "../ui/cards/FittedCard.tsx";

interface MobileCardDetailProps {
  card: CardDto | PlayerCardDto;
  actions?: ReactNode;
  notice?: ReactNode;
  onClose: () => void;
}

const INSPECTION_CARD_WIDTH = 360;
const CORPORATION_CARD_WIDTH = 320;
const SIDE_COLUMN_WIDTH = 200;
const COLUMN_GAP = 20;

const noop = () => {};

export default function MobileCardDetail({
  card,
  actions,
  notice,
  onClose,
}: MobileCardDetailProps) {
  const reducedMotion = useReducedMotion();
  const boundsRef = useRef<HTMLDivElement>(null);
  useRenderPause("mobile-card-detail", true);
  useBackDismiss(onClose);

  const isCorporation = card.type === CardTypeCorporation;
  const hasSideColumn = !!actions || !!notice;
  const reservedWidth = hasSideColumn ? 2 * (SIDE_COLUMN_WIDTH + COLUMN_GAP) : 0;
  const animationClass = reducedMotion ? "" : "animate-[modalSlideIn_180ms_ease-out]";

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={card.name}
      className="fixed inset-0 h-dvh text-white"
      style={{ zIndex: Z_INDEX.MOBILE_CARD_DETAIL }}
    >
      <div
        className="absolute inset-0 bg-black/80 backdrop-blur-sm cursor-default"
        onClick={onClose}
      />
      <div
        ref={boundsRef}
        className={`relative h-full grid grid-cols-[1fr_auto_1fr] items-center pointer-events-none ${animationClass}`}
        style={{
          gap: COLUMN_GAP,
          paddingTop: "calc(var(--safe-top) + 12px)",
          paddingBottom: "calc(var(--safe-bottom) + 12px)",
          paddingLeft: "calc(var(--safe-left) + 56px)",
          paddingRight: "calc(var(--safe-right) + 56px)",
        }}
      >
        <div className="col-start-2 flex justify-center">
          {isCorporation ? (
            <FittedCard
              naturalWidth={CORPORATION_CARD_WIDTH}
              boundsRef={boundsRef}
              reservedWidth={reservedWidth}
              className="pointer-events-auto"
            >
              <CorporationCard
                card={card}
                isSelected={false}
                onSelect={noop}
                disableInteraction
                catalog
              />
            </FittedCard>
          ) : (
            <FittedCard
              naturalWidth={INSPECTION_CARD_WIDTH}
              boundsRef={boundsRef}
              reservedWidth={reservedWidth}
              className="pointer-events-auto"
            >
              <div data-card-inspection>
                <GameCard card={card} presentation="inspection" dimUnavailable={false} />
              </div>
            </FittedCard>
          )}
        </div>

        {hasSideColumn && (
          <div
            className="col-start-3 justify-self-start max-w-full max-h-full overflow-y-auto overscroll-contain flex flex-col justify-center gap-3 pointer-events-auto"
            style={{ width: SIDE_COLUMN_WIDTH }}
          >
            {notice && <div className="text-[13px] leading-snug">{notice}</div>}
            {actions && (
              <div className="flex flex-col gap-2 [&>*]:w-full [&>*]:min-h-11">{actions}</div>
            )}
          </div>
        )}
      </div>
      <div
        className="absolute"
        style={{ top: "calc(var(--safe-top) + 8px)", right: "calc(var(--safe-right) + 8px)" }}
      >
        <CloseButton onClick={onClose} label="Close card" />
      </div>
    </div>
  );
}
