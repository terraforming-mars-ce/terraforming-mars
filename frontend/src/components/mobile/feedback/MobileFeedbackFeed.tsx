import { useEffect, useRef, useState } from "react";
import type { CardDto, TriggeredEffectDto } from "@/types/generated/api-types.ts";
import type { PlayedCardNotification } from "@/hooks/usePlayedCardNotification.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import GameButton from "../../ui/buttons/GameButton.tsx";
import CardTagList from "../../ui/cards/CardTagList.tsx";
import TriggeredEffectToast, {
  groupTriggeredEffects,
  type TriggeredEffectItem,
} from "../../ui/cards/TriggeredEffectToast.tsx";
import MobileCardDetail from "../MobileCardDetail.tsx";
import { useMobileGame } from "../MobileGameContext.tsx";

const TOAST_DURATION_MS = 4000;
const MAX_VISIBLE = 3;
const DEFAULT_COLOR = "#6496ff";

interface EffectEntry {
  seq: number;
  playerId: string;
  item: TriggeredEffectItem;
}

interface MobileFeedbackFeedProps {
  triggeredEffects: TriggeredEffectDto[];
  playedCardNotification: PlayedCardNotification | null;
  isPlayedCardPinned: boolean;
  onPlayedCardTogglePin?: () => void;
  onPlayedCardAdvance?: () => void;
}

function useDismissTimer(onDismiss: () => void, paused: boolean) {
  const onDismissRef = useRef(onDismiss);
  onDismissRef.current = onDismiss;

  useEffect(() => {
    if (paused) {
      return undefined;
    }
    const timer = setTimeout(() => onDismissRef.current(), TOAST_DURATION_MS);
    return () => clearTimeout(timer);
  }, [paused]);
}

function EffectToast({
  entry,
  playerName,
  playerColor,
  animationClass,
  onDismiss,
}: {
  entry: EffectEntry;
  playerName: string;
  playerColor: string;
  animationClass: string;
  onDismiss: () => void;
}) {
  useDismissTimer(onDismiss, false);
  return (
    <div
      className={`pointer-events-auto max-w-full cursor-pointer ${animationClass}`}
      onClick={onDismiss}
    >
      <TriggeredEffectToast
        item={entry.item}
        header={
          <span
            className="font-orbitron text-[11px] font-bold tracking-wider leading-none"
            style={{ color: playerColor }}
          >
            {playerName}
          </span>
        }
      />
    </div>
  );
}

function PlayedCardToast({
  notification,
  isPinned,
  animationClass,
  onTogglePin,
  onDismiss,
  onView,
}: {
  notification: PlayedCardNotification;
  isPinned: boolean;
  animationClass: string;
  onTogglePin?: () => void;
  onDismiss: () => void;
  onView: () => void;
}) {
  const wasPinnedRef = useRef(isPinned);
  useDismissTimer(onDismiss, isPinned);

  useEffect(() => {
    if (wasPinnedRef.current && !isPinned) {
      onDismiss();
    }
    wasPinnedRef.current = isPinned;
  }, [isPinned, onDismiss]);

  const color = notification.playerColor || DEFAULT_COLOR;

  return (
    <div
      className={`pointer-events-auto w-full flex items-center gap-1.5 pl-2.5 pr-1 py-0.5 bg-[rgba(10,10,15,0.96)] border border-white/15 shadow-[0_2px_8px_rgba(0,0,0,0.5)] ${animationClass}`}
      style={{ borderLeft: `3px solid ${color}` }}
    >
      <div
        className={`flex-1 min-w-0 flex flex-col justify-center min-h-11 ${isPinned ? "cursor-default" : "cursor-pointer"}`}
        onClick={() => {
          if (!isPinned) {
            onDismiss();
          }
        }}
      >
        <span className="font-orbitron text-[11px] leading-tight truncate">
          <span style={{ color }}>{notification.playerName}</span>
          <span className="text-white/60"> played</span>
        </span>
        <span className="flex items-center gap-1.5 min-w-0">
          <span className="text-[13px] font-bold leading-tight text-white truncate">
            {notification.card.name}
          </span>
          <CardTagList
            card={notification.card}
            size="sm"
            className="shrink-0 flex items-center gap-0.5"
          />
        </span>
      </div>
      {onTogglePin && (
        <GameButton
          size="sm"
          emphasis="quiet"
          selected={isPinned}
          aria-pressed={isPinned}
          className="shrink-0 !min-h-11 !px-2.5 !text-[11px]"
          onClick={onTogglePin}
        >
          {isPinned ? "Unpin" : "Pin"}
        </GameButton>
      )}
      <GameButton size="sm" className="shrink-0 !min-h-11 !px-3 !text-[11px]" onClick={onView}>
        View
      </GameButton>
    </div>
  );
}

export default function MobileFeedbackFeed({
  triggeredEffects,
  playedCardNotification,
  isPlayedCardPinned,
  onPlayedCardTogglePin,
  onPlayedCardAdvance,
}: MobileFeedbackFeedProps) {
  const { gameState, playerColorMap } = useMobileGame();
  const reducedMotion = useReducedMotion();
  const [effects, setEffects] = useState<EffectEntry[]>([]);
  const [playedSeq, setPlayedSeq] = useState<{ id: number; seq: number } | null>(null);
  const [viewCard, setViewCard] = useState<CardDto | null>(null);
  const prevTriggeredEffectsRef = useRef(triggeredEffects);
  const seqRef = useRef(0);

  useEffect(() => {
    if (triggeredEffects === prevTriggeredEffectsRef.current) {
      return;
    }
    prevTriggeredEffectsRef.current = triggeredEffects;
    const playerIds = [...new Set(triggeredEffects.map((effect) => effect.playerId))];
    const entries = playerIds.flatMap((playerId) =>
      groupTriggeredEffects(triggeredEffects, playerId).map((item) => ({
        seq: seqRef.current++,
        playerId,
        item,
      })),
    );
    if (entries.length > 0) {
      setEffects((prev) => [...prev, ...entries].slice(-MAX_VISIBLE));
    }
  }, [triggeredEffects]);

  const playedId = playedCardNotification?.id;
  useEffect(() => {
    if (playedId === undefined) {
      setPlayedSeq(null);
      return;
    }
    setPlayedSeq({ id: playedId, seq: seqRef.current++ });
  }, [playedId]);

  const dismissEffect = (seq: number) => {
    setEffects((prev) => prev.filter((entry) => entry.seq !== seq));
  };

  const advancePlayed = () => {
    onPlayedCardAdvance?.();
  };

  const playerNames = new Map<string, string>();
  if (gameState.currentPlayer) {
    playerNames.set(gameState.currentPlayer.id, gameState.currentPlayer.name);
  }
  for (const other of gameState.otherPlayers ?? []) {
    playerNames.set(other.id, other.name);
  }

  const showPlayed =
    playedCardNotification !== null && playedSeq !== null && playedSeq.id === playedId;
  const visibleEffects = effects.slice(showPlayed ? 1 - MAX_VISIBLE : -MAX_VISIBLE);
  const animationClass = reducedMotion ? "" : "animate-[fadeIn_200ms_ease-out]";

  const toasts = visibleEffects.map((entry) => ({
    seq: entry.seq,
    node: (
      <EffectToast
        key={`effect-${entry.seq}`}
        entry={entry}
        playerName={playerNames.get(entry.playerId) ?? "Unknown"}
        playerColor={playerColorMap.get(entry.playerId) ?? DEFAULT_COLOR}
        animationClass={animationClass}
        onDismiss={() => dismissEffect(entry.seq)}
      />
    ),
  }));
  if (showPlayed && playedCardNotification) {
    toasts.push({
      seq: playedSeq.seq,
      node: (
        <PlayedCardToast
          key={`played-${playedCardNotification.id}`}
          notification={playedCardNotification}
          isPinned={isPlayedCardPinned}
          animationClass={animationClass}
          onTogglePin={onPlayedCardTogglePin}
          onDismiss={advancePlayed}
          onView={() => setViewCard(playedCardNotification.card)}
        />
      ),
    });
  }
  toasts.sort((a, b) => b.seq - a.seq);

  return (
    <>
      {toasts.length > 0 && (
        <div
          className="fixed flex flex-col items-center gap-1.5 pointer-events-none"
          style={{
            zIndex: Z_INDEX.MOBILE_FEEDBACK,
            top: "calc(var(--hud-top-h) + var(--safe-top) + 8px)",
            left: "calc(var(--hud-rail-left-w) + var(--safe-left) + 8px)",
            right: "calc(var(--hud-rail-right-w) + var(--safe-right) + 8px)",
          }}
        >
          <div className="w-full max-w-[360px] flex flex-col items-center gap-1.5">
            {toasts.map((toast) => toast.node)}
          </div>
        </div>
      )}
      {viewCard && <MobileCardDetail card={viewCard} onClose={() => setViewCard(null)} />}
    </>
  );
}
