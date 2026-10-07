import { useEffect, useLayoutEffect, useRef, useState, type ReactNode } from "react";
import type { CardDto, GameDto, InitPhaseDto } from "@/types/generated/api-types.ts";
import { GamePhaseInitApplyPrelude } from "@/types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import { audioService } from "@/services/audioService.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import { CARD_RELEASE_DURATION_MS, playCardRelease } from "@/utils/cardReleaseAnimation.ts";
import { SHOWCASE_FADE_MS } from "@/constants/gameConstants.ts";
import { useShowcaseTileHold } from "@/hooks/useShowcaseTileHold.ts";
import {
  corpShown,
  isShowcaseStepEnd,
  showcaseControllerId,
  showcasePlayers,
  shownPreludes,
  type ShowcasePlayer,
} from "@/utils/showcase.ts";
import CorporationCard from "../cards/CorporationCard.tsx";
import GameCard from "../cards/GameCard.tsx";
import GameButton from "../buttons/GameButton.tsx";

type Viewport = { width: number; height: number };

const CORP_CARD_WIDTH = 400;
/** Typical rendered heights, used only to pick a scale; boxes are sized from the real height. */
const CORP_CARD_HEIGHT = 400;
const CORP_MAX_SCALE = 1.2;
const PRELUDE_CARD_WIDTH = 360;
const PRELUDE_CARD_HEIGHT = 500;
const PRELUDE_GAP = 40;
const PRELUDE_MAX_SCALE = 0.8;
/** Summary preludes are drawn at this fraction of the corp card's scale, two fitting under it. */
const SUMMARY_PRELUDE_RATIO = 0.5;
const SUMMARY_GAP = 16;
const NAME_SPACE = 80;
const BOTTOM_BAR_SPACE = 96;
const ROSTER_COLUMN_GAP = 32;

const noop = () => {};

interface Props {
  game: GameDto;
  initPhase: InitPhaseDto;
}

export default function CorporationShowcaseOverlay({ game, initPhase }: Props) {
  const reducedMotion = useReducedMotion();
  const [viewport, setViewport] = useState<Viewport>({
    width: window.innerWidth,
    height: window.innerHeight,
  });
  const [viewingPlayerId, setViewingPlayerId] = useState<string | null>(null);
  const [sentVersion, setSentVersion] = useState<number | null>(null);

  useEffect(() => {
    const resize = () => setViewport({ width: window.innerWidth, height: window.innerHeight });
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, []);

  const players = showcasePlayers(game);
  const isPreludePhase = game.currentPhase === GamePhaseInitApplyPrelude;
  const isRoster = initPhase.stage === "roster";
  const livePlayer = players.find((p) => p.id === initPhase.currentPlayerId) ?? null;
  const liveIndex = livePlayer ? players.indexOf(livePlayer) : -1;
  const pending = initPhase.hasPendingSelection;
  const holdingAfterChoice = useShowcaseTileHold(pending);
  const faded = pending || holdingAfterChoice;
  const animate = !reducedMotion;

  const viewerId = game.isSpectator ? null : (game.currentPlayer?.id ?? null);
  const isController = viewerId !== null && showcaseControllerId(game) === viewerId;
  const stepEnd = isShowcaseStepEnd(game) && !faded;
  const nextSent = sentVersion === initPhase.confirmVersion;
  const viewingPlayer = players.find((p) => p.id === viewingPlayerId) ?? null;

  const handleNext = () => {
    setViewingPlayerId(null);
    setSentVersion(initPhase.confirmVersion);
    void globalWebSocketManager.confirmInitAdvance();
  };

  let content: ReactNode = null;
  if (viewingPlayer) {
    const index = players.indexOf(viewingPlayer);
    content = (
      <PlayerSummary
        key={`history:${viewingPlayer.id}`}
        player={viewingPlayer}
        corporation={corpShown(game, index) ? viewingPlayer.corporation : undefined}
        preludes={shownPreludes(viewingPlayer)}
        scale={summaryScale(viewport)}
      />
    );
  } else if (isRoster) {
    content = <RosterSlide players={players} viewport={viewport} />;
  } else if (livePlayer && isPreludePhase) {
    content = (
      <PreludeSlide
        key={`prelude:${livePlayer.id}`}
        player={livePlayer}
        preludes={initPhase.preludes}
        played={initPhase.preludesPlayed}
        viewport={viewport}
        animate={animate}
      />
    );
  } else if (livePlayer?.corporation) {
    content = (
      <CorpSlide
        key={`corp:${livePlayer.id}`}
        player={livePlayer}
        corporation={livePlayer.corporation}
        played={initPhase.stage === "applied"}
        viewport={viewport}
        animate={animate}
      />
    );
  }

  return (
    <div
      className={`fixed inset-0 flex flex-col transition-opacity ease-out ${faded ? "pointer-events-none opacity-0" : "pointer-events-auto opacity-100"}`}
      style={{
        zIndex: Z_INDEX.SHOWCASE,
        transitionDuration: `${SHOWCASE_FADE_MS}ms`,
        // Let the card that caused the choice finish playing before the board takes over.
        transitionDelay: faded ? `${CARD_RELEASE_DURATION_MS}ms` : "0ms",
      }}
      role="dialog"
      aria-label="Corporation showcase"
      aria-hidden={faded}
    >
      <style>{SHOWCASE_KEYFRAMES}</style>
      <div className="absolute inset-0 bg-black/65 backdrop-blur-sm animate-[showcaseFadeIn_400ms_ease-out_both]" />
      <main className="relative flex min-h-0 flex-1 items-center justify-center px-4">
        {content}
      </main>
      <footer className="relative flex items-center gap-3 px-8 pb-6 max-md:px-4">
        <div className="flex min-w-0 flex-1 flex-wrap items-center gap-2">
          {players.map((p, i) => {
            const shown = corpShown(game, i);
            const isLive = !isRoster && i === liveIndex;
            const isViewed = viewingPlayerId === p.id;
            return (
              <button
                key={p.id}
                type="button"
                disabled={!shown}
                onClick={() => setViewingPlayerId(p.id)}
                className={`flex items-center border px-3 py-1.5 font-orbitron text-[11px] uppercase tracking-[0.1em] transition-colors duration-300 ${shown ? "cursor-pointer hover:border-white/50" : "cursor-default opacity-50"} ${isViewed ? "border-white bg-white/15" : ""} ${!isViewed && isLive ? "border-white/60 bg-white/10" : ""} ${!isViewed && !isLive ? "border-white/15 bg-black/40" : ""}`}
              >
                <span className="h-2 w-2 shrink-0" style={{ background: p.color }} />
                <span className="ml-2 text-white/90">{p.name}</span>
                <span
                  className={`grid transition-[grid-template-columns,opacity] duration-500 ease-out ${shown && p.corporation ? "grid-cols-[1fr] opacity-100" : "grid-cols-[0fr] opacity-0"}`}
                >
                  <span className="overflow-hidden whitespace-nowrap text-white/60">
                    <span className="px-2 text-white/40">·</span>
                    {p.corporation?.name}
                  </span>
                </span>
              </button>
            );
          })}
          {viewingPlayer && (
            <GameButton size="sm" emphasis="quiet" onClick={() => setViewingPlayerId(null)}>
              Follow host
            </GameButton>
          )}
        </div>
        {isController ? (
          <GameButton size="sm" disabled={!stepEnd || nextSent} onClick={handleNext}>
            {isRoster ? "Start game" : "Next"}
          </GameButton>
        ) : (
          <span
            className={`font-orbitron text-xs uppercase tracking-[0.15em] text-white/60 transition-opacity duration-300 ${stepEnd ? "opacity-100" : "opacity-0"}`}
          >
            Waiting for host…
          </span>
        )}
      </footer>
    </div>
  );
}

function availableHeight(viewport: Viewport): number {
  return viewport.height - NAME_SPACE - BOTTOM_BAR_SPACE - 48;
}

function summaryScale(viewport: Viewport): number {
  const columnHeight = CORP_CARD_HEIGHT + PRELUDE_CARD_HEIGHT * SUMMARY_PRELUDE_RATIO + SUMMARY_GAP;
  return Math.min(
    CORP_MAX_SCALE,
    availableHeight(viewport) / columnHeight,
    (viewport.width - 64) / CORP_CARD_WIDTH,
  );
}

/**
 * Lays out a card at a scale: the box takes the card's real rendered height times the
 * scale, so scaled cards sit snugly without guessing heights.
 */
function ScaledBox({
  width,
  scale,
  estimatedHeight,
  children,
}: {
  width: number;
  scale: number;
  estimatedHeight: number;
  children: ReactNode;
}) {
  const innerRef = useRef<HTMLDivElement | null>(null);
  const [height, setHeight] = useState(estimatedHeight);

  useLayoutEffect(() => {
    const element = innerRef.current;
    if (!element) {
      return;
    }
    const measure = () => setHeight(element.offsetHeight);
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  return (
    <div style={{ width: width * scale, height: height * scale }}>
      <div
        ref={innerRef}
        style={{ width, transform: `scale(${scale})`, transformOrigin: "top left" }}
      >
        {children}
      </div>
    </div>
  );
}

type CardPlayState = "waiting" | "releasing" | "gone";

/**
 * Plays a showcase card once `played` turns true: the normal card-play release animation
 * and sound, after which the card stays gone. A card already played on mount is gone.
 */
function useCardPlay(played: boolean, animate: boolean) {
  const ref = useRef<HTMLDivElement | null>(null);
  const handled = useRef(played);
  const [state, setState] = useState<CardPlayState>(played ? "gone" : "waiting");

  useLayoutEffect(() => {
    if (!played || handled.current) {
      return;
    }
    handled.current = true;
    void audioService.playSound("card-played");
    const element = ref.current;
    if (!element || !animate) {
      setState("gone");
      return;
    }
    setState("releasing");
    const animation = playCardRelease(element);
    animation.addEventListener("finish", () => setState("gone"));
  }, [played, animate]);

  return { ref, state };
}

const FLIP_DELAY_MS = 150;
const FLIP_STAGGER_MS = 200;

const FLIP_CLASS = "animate-[showcaseFlipIn_700ms_cubic-bezier(0.2,0.8,0.2,1)_both]";

function CorpSlide({
  player,
  corporation,
  played,
  viewport,
  animate,
}: {
  player: ShowcasePlayer;
  corporation: CardDto;
  played: boolean;
  viewport: Viewport;
  animate: boolean;
}) {
  const { ref, state } = useCardPlay(played, animate);
  const scale = Math.min(
    CORP_MAX_SCALE,
    (viewport.width - 64) / CORP_CARD_WIDTH,
    availableHeight(viewport) / CORP_CARD_HEIGHT,
  );

  return (
    <div className="flex flex-col items-center gap-6">
      <PlayerName player={player} />
      <div
        className={animate ? FLIP_CLASS : ""}
        style={{ animationDelay: animate ? `${FLIP_DELAY_MS}ms` : undefined }}
      >
        <div ref={ref} style={{ visibility: state === "gone" ? "hidden" : "visible" }}>
          <ScaledBox width={CORP_CARD_WIDTH} scale={scale} estimatedHeight={CORP_CARD_HEIGHT}>
            <CorpCard card={corporation} />
          </ScaledBox>
        </div>
      </div>
    </div>
  );
}

function PreludeSlide({
  player,
  preludes,
  played,
  viewport,
  animate,
}: {
  player: ShowcasePlayer;
  preludes: CardDto[];
  played: number;
  viewport: Viewport;
  animate: boolean;
}) {
  const count = Math.max(1, preludes.length);
  const rowWidth = count * PRELUDE_CARD_WIDTH + (count - 1) * PRELUDE_GAP;
  const scale = Math.min(
    PRELUDE_MAX_SCALE,
    (viewport.width - 64) / rowWidth,
    availableHeight(viewport) / PRELUDE_CARD_HEIGHT,
  );

  return (
    <div className="flex flex-col items-center gap-6">
      <PlayerName player={player} />
      <div className="flex items-start" style={{ gap: PRELUDE_GAP * scale }}>
        {preludes.map((card, i) => (
          <PreludeCard
            key={card.id}
            card={card}
            played={i < played}
            index={i}
            scale={scale}
            animate={animate}
          />
        ))}
      </div>
    </div>
  );
}

function PreludeCard({
  card,
  played,
  index,
  scale,
  animate,
}: {
  card: CardDto;
  played: boolean;
  index: number;
  scale: number;
  animate: boolean;
}) {
  const { ref, state } = useCardPlay(played, animate);
  const flipDelay = FLIP_DELAY_MS + index * FLIP_STAGGER_MS;
  return (
    <div
      className={animate ? FLIP_CLASS : ""}
      style={{ animationDelay: animate ? `${flipDelay}ms` : undefined }}
    >
      <div ref={ref} style={{ visibility: state === "gone" ? "hidden" : "visible" }}>
        <ScaledBox width={PRELUDE_CARD_WIDTH} scale={scale} estimatedHeight={PRELUDE_CARD_HEIGHT}>
          <GameCard
            card={card}
            presentation="inspection"
            moduleState={state === "releasing" ? "releasing" : "idle"}
          />
        </ScaledBox>
      </div>
    </div>
  );
}

/** A player's shown cards for looking back: the corporation on top, played preludes below. */
function PlayerSummary({
  player,
  corporation,
  preludes,
  scale,
}: {
  player: ShowcasePlayer;
  corporation: CardDto | undefined;
  preludes: CardDto[];
  scale: number;
}) {
  const preludeScale = scale * SUMMARY_PRELUDE_RATIO;
  const preludeGap = CORP_CARD_WIDTH * scale - 2 * PRELUDE_CARD_WIDTH * preludeScale;
  return (
    <div className="flex flex-col items-center gap-4 animate-[showcaseFadeIn_400ms_ease-out_both]">
      <PlayerName player={player} />
      <div className="flex flex-col items-center" style={{ gap: SUMMARY_GAP }}>
        {corporation && (
          <ScaledBox width={CORP_CARD_WIDTH} scale={scale} estimatedHeight={CORP_CARD_HEIGHT}>
            <CorpCard card={corporation} />
          </ScaledBox>
        )}
        {preludes.length > 0 && (
          <div className="flex items-start justify-center" style={{ gap: Math.max(8, preludeGap) }}>
            {preludes.map((card) => (
              <ScaledBox
                key={card.id}
                width={PRELUDE_CARD_WIDTH}
                scale={preludeScale}
                estimatedHeight={PRELUDE_CARD_HEIGHT}
              >
                <GameCard card={card} presentation="inspection" />
              </ScaledBox>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function CorpCard({ card }: { card: CardDto }) {
  return (
    <CorporationCard
      card={card}
      isSelected={false}
      onSelect={noop}
      disableInteraction={true}
      borderColor={getCorporationBorderColor(card.name)}
    />
  );
}

function PlayerName({
  player,
  size = "large",
}: {
  player: ShowcasePlayer;
  size?: "large" | "small";
}) {
  const sizeClass = size === "large" ? "text-3xl max-md:text-xl" : "text-lg";
  return (
    <span
      className={`font-orbitron font-bold uppercase tracking-[0.15em] [text-shadow:0_2px_8px_rgba(0,0,0,0.8)] animate-[showcaseFadeIn_400ms_ease-out_both] ${sizeClass}`}
      style={{ color: player.color }}
    >
      {player.name}
    </span>
  );
}

/** The final step: every player's corporation side by side. */
function RosterSlide({ players, viewport }: { players: ShowcasePlayer[]; viewport: Viewport }) {
  const withCorp = players.filter(
    (p): p is ShowcasePlayer & { corporation: CardDto } => p.corporation !== undefined,
  );
  const count = Math.max(1, withCorp.length);
  const scale = Math.min(
    0.9,
    availableHeight(viewport) / CORP_CARD_HEIGHT,
    (viewport.width - 64 - (count - 1) * ROSTER_COLUMN_GAP) / (count * CORP_CARD_WIDTH),
  );
  return (
    <div className="flex items-start justify-center" style={{ gap: ROSTER_COLUMN_GAP }}>
      {withCorp.map((p, i) => (
        <div
          key={p.id}
          className="flex flex-col items-center gap-3 animate-[showcaseFadeIn_500ms_ease-out_both]"
          style={{ animationDelay: `${i * 120}ms` }}
        >
          <PlayerName player={p} size="small" />
          <ScaledBox width={CORP_CARD_WIDTH} scale={scale} estimatedHeight={CORP_CARD_HEIGHT}>
            <CorpCard card={p.corporation} />
          </ScaledBox>
        </div>
      ))}
    </div>
  );
}

const SHOWCASE_KEYFRAMES = `
  @keyframes showcaseFadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @keyframes showcaseFlipIn {
    from { opacity: 0; transform: perspective(1200px) rotateY(80deg) scale(0.9); }
    to { opacity: 1; transform: perspective(1200px) rotateY(0deg) scale(1); }
  }
`;
