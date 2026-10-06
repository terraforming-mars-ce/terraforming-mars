import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useCardPlayFlowStore, type CardPlaySession } from "@/stores/cardPlayFlowStore";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { Z_INDEX } from "@/constants/zIndex";
import { CARD_RELEASE_KEYFRAMES, CARD_RELEASE_TIMING } from "@/utils/cardReleaseAnimation";
import CardInspection from "./CardInspection";
import GameCard from "../cards/GameCard";

const noop = () => {};

export default function CardPlayPresentation({
  chatBounds,
}: {
  chatBounds: DOMRectReadOnly | null;
}) {
  const session = useCardPlayFlowStore((state) => state.playSession);
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
  const compact = viewport.width < 1024 || viewport.height < 720;
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
      data-layout={compact ? "stacked" : "side"}
      className={
        compact
          ? "fixed inset-x-0 top-16 bottom-24 flex flex-col items-center gap-4 overflow-y-auto overscroll-contain px-3 pointer-events-none"
          : "fixed inset-0 pointer-events-none"
      }
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <CardInspection
        card={session.card}
        inspection={inspection}
        playSession={session}
        compact={compact}
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
