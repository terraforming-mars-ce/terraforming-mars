import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useCardPlayFlowStore, type CardPlaySession } from "@/stores/cardPlayFlowStore";
import { usePaymentStore } from "@/stores/paymentStore";
import { useGameStore } from "@/stores/gameStore";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { useLayoutMode } from "@/hooks/useLayoutMode";
import type { PlayerCardDto } from "@/types/generated/api-types";
import { Z_INDEX } from "@/constants/zIndex";
import { CARD_RELEASE_KEYFRAMES, CARD_RELEASE_TIMING } from "@/utils/cardReleaseAnimation";
import CardInspection from "./CardInspection";
import GameCard from "../cards/GameCard";
import FittedCard from "../cards/FittedCard";
import { GAME_CARD_NATURAL_WIDTH } from "../cards/CardFitHeightContext";

const noop = () => {};

type CardPlayFlowState = ReturnType<typeof useCardPlayFlowStore.getState>;

function stagedCardId(state: CardPlayFlowState): string | null {
  if (state.playSession?.phase === "choosing") {
    return state.playSession.card.id;
  }
  if (state.showChoiceSelection && state.cardPendingChoice) {
    return state.cardPendingChoice.id;
  }
  if (state.showCardStorageSelection && state.pendingCardStorage) {
    return state.pendingCardStorage.cardId;
  }
  if (state.showTargetPlayerSelection && state.pendingTargetPlayer) {
    return state.pendingTargetPlayer.cardId;
  }
  if (state.showAmountSelection && state.pendingVariableAmount?.type === "play-card") {
    return state.pendingVariableAmount.cardId;
  }
  if (state.showCardResourceSelection && state.pendingCardResourceInput?.type === "play-card") {
    return state.pendingCardResourceInput.cardId;
  }
  return null;
}

export default function CardPlayPresentation({
  chatBounds,
}: {
  chatBounds: DOMRectReadOnly | null;
}) {
  const { isCompact } = useLayoutMode();
  const session = useCardPlayFlowStore((state) => state.playSession);
  if (isCompact) {
    return <CompactPlayStage session={session} />;
  }
  if (!session) {
    return null;
  }
  return <Presentation key={session.id} session={session} chatBounds={chatBounds} />;
}

function Presentation({
  session,
  chatBounds,
}: {
  session: CardPlaySession;
  chatBounds: DOMRectReadOnly | null;
}) {
  const [viewport, setViewport] = useState({
    width: window.innerWidth,
    height: window.innerHeight,
  });
  const reducedMotion = useReducedMotion();
  const directRef = useRef<HTMLDivElement>(null);
  const stacked = viewport.width < 1024 || viewport.height < 720;
  const inspection = useMemo(
    () => ({
      cardId: session.card.id,
      source: session.source,
      closing: session.phase === "returning",
      restoreFocus: session.phase === "returning",
    }),
    [session.card.id, session.source, session.phase === "returning"],
  );
  const finish = useCallback(() => {
    useCardPlayFlowStore.getState().finishPlayAnimation(session.id);
  }, [session.id]);
  const host = useCallback((element: HTMLDivElement | null) => {
    useCardPlayFlowStore.getState().setPlayPromptHost(element);
  }, []);

  useLayoutEffect(() => {
    const resize = () => setViewport({ width: window.innerWidth, height: window.innerHeight });
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, []);

  const departing = session.phase === "submitting" || session.phase === "reconnecting";
  const returning = session.phase === "returning";
  useLayoutEffect(() => {
    if (session.inspected || session.animationDone || (!departing && !returning)) {
      return;
    }
    const element = directRef.current;
    if (!element || reducedMotion) {
      finish();
      return;
    }
    const frames: Keyframe[] = [...CARD_RELEASE_KEYFRAMES];
    if (returning) {
      const bounds = session.source.getBoundingClientRect();
      frames[1] = {
        transform: `translate(${bounds.x - session.origin.x}px, ${bounds.y - session.origin.y}px)`,
        opacity: 0,
      };
    }
    const animation = element.animate(frames, CARD_RELEASE_TIMING);
    animation.onfinish = finish;
    return () => {
      animation.onfinish = null;
      animation.cancel();
    };
  }, [
    session.id,
    session.inspected,
    departing,
    returning,
    session.animationDone,
    reducedMotion,
    finish,
  ]);

  if (session.animationDone) {
    return null;
  }
  if (!session.inspected) {
    return createPortal(
      <div
        className="fixed pointer-events-none"
        aria-hidden="true"
        style={{
          left: session.origin.x,
          top: session.origin.y,
          zIndex: Z_INDEX.CARD_DETAIL_MODAL,
        }}
      >
        <div ref={directRef}>
          <div
            className="card-size"
            style={{ transform: `scale(${session.origin.scale})`, transformOrigin: "top left" }}
          >
            <GameCard card={session.card} isSelected moduleState="releasing" />
          </div>
        </div>
      </div>,
      document.body,
    );
  }
  return createPortal(
    <div
      data-card-play-stage
      data-layout={stacked ? "stacked" : "side"}
      className={
        stacked
          ? "fixed inset-x-0 top-16 bottom-24 flex flex-col items-center gap-4 overflow-y-auto overscroll-contain px-3 pointer-events-none"
          : "fixed inset-0 pointer-events-none"
      }
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <CardInspection
        card={session.card}
        inspection={inspection}
        playSession={session}
        sheet={stacked}
        onReturned={finish}
        chatBounds={chatBounds}
        onClose={noop}
        onDragStart={noop}
      />
      <div
        ref={host}
        data-card-play-prompt
        className="fixed inset-0 flex items-center justify-center pointer-events-none px-3"
      />
    </div>,
    document.body,
  );
}

function CompactPlayStage({ session }: { session: CardPlaySession | null }) {
  const flowCardId = useCardPlayFlowStore(stagedCardId);
  const paymentCardId = usePaymentStore((state) =>
    state.pending?.intent.action === "play-card" ? (state.pending.intent.cardId ?? null) : null,
  );
  const cardId = flowCardId ?? paymentCardId;
  const handCard = useGameStore((state) =>
    cardId ? state.currentPlayer?.cards.find((candidate) => candidate.id === cardId) : undefined,
  );
  const card = handCard ?? (session?.card.id === cardId ? session?.card : undefined);
  const host = useCallback((element: HTMLDivElement | null) => {
    useCardPlayFlowStore.getState().setPlayPromptHost(element);
  }, []);

  useLayoutEffect(() => {
    if (!session || session.animationDone) {
      return;
    }
    const store = useCardPlayFlowStore.getState();
    if (session.phase === "choosing") {
      store.inspectionArrived(session.id);
    } else if (session.phase !== "preparing") {
      store.finishPlayAnimation(session.id);
    }
  }, [session]);

  if (!card) {
    return null;
  }
  return createPortal(
    <div
      data-card-play-stage
      className="fixed inset-0 animate-[fadeIn_150ms_ease-out]"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div className="absolute inset-0 bg-black/70" />
      <div
        className="absolute flex gap-3"
        style={{
          top: "calc(var(--safe-top) + 8px)",
          right: "calc(var(--safe-right) + 8px)",
          bottom: "calc(var(--safe-bottom) + 8px)",
          left: "calc(var(--safe-left) + 8px)",
        }}
      >
        <FittedPlayCard card={card} />
        <div
          ref={host}
          data-card-play-prompt
          className="w-[60%] min-w-0 h-full flex items-center justify-center pointer-events-none"
        />
      </div>
    </div>,
    document.body,
  );
}

function FittedPlayCard({ card }: { card: PlayerCardDto }) {
  const columnRef = useRef<HTMLDivElement>(null);
  return (
    <div
      ref={columnRef}
      data-overlay-layer
      className="w-[40%] min-w-0 h-full flex items-center justify-center"
    >
      <FittedCard naturalWidth={GAME_CARD_NATURAL_WIDTH} boundsRef={columnRef} maxScale={2}>
        <GameCard card={card} dimUnavailable={false} />
      </FittedCard>
    </div>
  );
}
