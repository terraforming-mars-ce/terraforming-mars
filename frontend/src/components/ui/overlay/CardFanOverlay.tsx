import React, {
  useState,
  useEffect,
  useRef,
  useCallback,
  useMemo,
  useImperativeHandle,
  forwardRef,
} from "react";
import GameCard from "../cards/GameCard.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";
import BlurredOverlay from "./BlurredOverlay.tsx";
import { PlayerCardDto } from "@/types/generated/api-types.ts";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

function clamp(n: number, min: number, max: number): number {
  return Math.min(Math.max(n, min), max);
}

// --- Fan layout constants ---
const SPACING = 120;
const CARD_WIDTH = 200;
const CARD_HEIGHT = 280;
const MAX_PANEL_WIDTH = 640;
const FAN_PADDING = 40;
const FLAT_RADIUS = 2.5;
const TURN_RADIUS = 4;
const MAX_ROTATE = 22;
const EDGE_DROP = 18;
const BASE_Y_OFFSET = 160;
const VISIBLE_RADIUS = 3;
const CULL_RADIUS = VISIBLE_RADIUS + 1;

// --- Expanded layout constants ---
const EXPANDED_SPACING = 216;

const SELECTED_LIFT = -180;
const SELECTED_SCALE = 1.12;

const THROW_DISTANCE_THRESHOLD = 120;
const THROW_Y_THRESHOLD = -80;
const DRAG_THRESHOLD = 14;
const DRAG_RAISE_THRESHOLD = -60;
const WHEEL_SCALE = 0.005;

interface CardTransform {
  x: number;
  y: number;
  rotation: number;
  scale: number;
  z: number;
}

function getCardTransform(i: number, scrollPos: number): CardTransform {
  const d = i - scrollPos;
  const absD = Math.abs(d);

  const x = d * SPACING;

  let rotation = 0;
  let t = 0;
  if (absD > FLAT_RADIUS) {
    t = clamp((absD - FLAT_RADIUS) / TURN_RADIUS, 0, 1);
    rotation = Math.sign(d) * Math.pow(t, 1.2) * MAX_ROTATE;
  }

  const y = BASE_Y_OFFSET + t * t * EDGE_DROP;

  const z = i;

  return { x, y, rotation, scale: 1, z };
}

export interface CardFanOverlayHandle {
  toggleExpand: () => void;
  collapse: () => void;
  readonly isExpanded: boolean;
}

interface CardFanOverlayProps {
  cards: PlayerCardDto[];
  hideWhenModalOpen?: boolean;
  onCardSelect?: (cardId: string) => void;
  onPlayCard?: (cardId: string) => Promise<void>;
}

const CardFanOverlay = forwardRef<CardFanOverlayHandle, CardFanOverlayProps>(
  ({ cards, hideWhenModalOpen = false, onCardSelect, onPlayCard }, ref) => {
    const [scrollPos, setScrollPos] = useState(0);
    const [cardOrder, setCardOrder] = useState<string[]>([]);
    const [highlightedCard, setHighlightedCard] = useState<string | null>(null);
    const [draggedCard, setDraggedCard] = useState<string | null>(null);
    const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
    const [dragPosition, setDragPosition] = useState({ x: 0, y: 0 });
    const [dragStartPosition, setDragStartPosition] = useState({ x: 0, y: 0 });
    const [isInThrowZone, setIsInThrowZone] = useState(false);
    const [returningCard, setReturningCard] = useState<string | null>(null);
    const [isExpanded, setIsExpanded] = useState(false);
    const [isTransitioning, setIsTransitioning] = useState(false);
    const [flyingAwayGhost, setFlyingAwayGhost] = useState<{
      card: PlayerCardDto;
      x: number;
      y: number;
      scale: number;
      animating: boolean;
    } | null>(null);

    const handRef = useRef<HTMLDivElement>(null);
    const cardsRef = useRef(cards);
    const dragIntentRef = useRef(false);
    const capturedCardRef = useRef<HTMLDivElement | null>(null);
    const lastPointerXRef = useRef(0);

    const { playCardHoverSound } = useSoundEffects();

    const [windowWidth, setWindowWidth] = useState(window.innerWidth);
    const [windowHeight, setWindowHeight] = useState(window.innerHeight);

    useEffect(() => {
      const handleResize = () => {
        setWindowWidth(window.innerWidth);
        setWindowHeight(window.innerHeight);
      };
      window.addEventListener("resize", handleResize);
      return () => window.removeEventListener("resize", handleResize);
    }, []);

    const fanScale = useMemo(() => {
      const panelWidth = Math.min(MAX_PANEL_WIDTH, Math.max(480, (windowWidth - 700) / 2));
      const availableWidth = windowWidth - 2 * panelWidth - FAN_PADDING;
      const maxVisibleWidth = 2 * VISIBLE_RADIUS * SPACING + CARD_WIDTH;
      if (maxVisibleWidth <= availableWidth) return 1;
      return Math.min(1, Math.max(0.5, (availableWidth / maxVisibleWidth) * 1.2));
    }, [windowWidth]);

    const expandedBaseY = useMemo(() => {
      return -(windowHeight / 2 - CARD_HEIGHT / 2 - 48);
    }, [windowHeight]);

    const expandedCullRadius = useMemo(() => {
      return Math.ceil(windowWidth / EXPANDED_SPACING / 2) + 2;
    }, [windowWidth]);

    // --- Expand / Collapse ---
    const handleExpand = useCallback(() => {
      setIsTransitioning(true);
      setIsExpanded(true);
      setHighlightedCard(null);
      setTimeout(() => setIsTransitioning(false), 350);
    }, []);

    const scrollTargetRef = useRef(0);
    const scrollAnimRef = useRef(0);
    const isAnimatingRef = useRef(false);

    const handleCollapse = useCallback(
      (cardId?: string) => {
        setIsTransitioning(true);
        setIsExpanded(false);
        if (cardId) {
          const cardIndex = cardOrder.indexOf(cardId);
          if (cardIndex >= 0) {
            scrollTargetRef.current = cardIndex;
            setScrollPos(cardIndex);
            setHighlightedCard(cardId);
          }
        }
        setTimeout(() => setIsTransitioning(false), 350);
      },
      [cardOrder],
    );

    useImperativeHandle(
      ref,
      () => ({
        toggleExpand: () => {
          if (isExpanded) {
            handleCollapse();
          } else {
            handleExpand();
          }
        },
        collapse: () => {
          if (isExpanded) {
            handleCollapse();
          }
        },
        get isExpanded() {
          return isExpanded;
        },
      }),
      [isExpanded, handleCollapse, handleExpand],
    );

    useEffect(() => {
      const handleKeyDown = (e: KeyboardEvent) => {
        if (e.key === "Escape" && isExpanded) {
          e.stopImmediatePropagation();
          handleCollapse();
        }
      };
      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }, [isExpanded, handleCollapse]);

    useEffect(() => {
      cardsRef.current = cards;
    }, [cards]);

    // Sync cardOrder when cards prop changes (additions/removals)
    useEffect(() => {
      setCardOrder((prev) => {
        const propIds = new Set(cards.map((c) => c.id));
        const kept = prev.filter((id) => propIds.has(id));
        const keptSet = new Set(kept);
        const added = cards.filter((c) => !keptSet.has(c.id)).map((c) => c.id);
        const next = [...kept, ...added];
        const isFirstRender = prev.length === 0;
        const result = isFirstRender ? cards.map((c) => c.id) : next;
        // Center scroll on first load
        if (isFirstRender && result.length > 0) {
          const center = (result.length - 1) / 2;
          setScrollPos(center);
          scrollTargetRef.current = center;
        }
        return result;
      });
    }, [cards]);

    // Clamp scrollPos when cards change
    useEffect(() => {
      const maxScroll = Math.max(cardOrder.length - 1, 0);
      setScrollPos((prev) => {
        const clamped = clamp(prev, 0, maxScroll);
        scrollTargetRef.current = clamped;
        return clamped;
      });
      // Clear selection if card disappeared
      if (highlightedCard && !cardOrder.includes(highlightedCard)) {
        setHighlightedCard(null);
      }
    }, [cardOrder, highlightedCard]);

    // --- Wheel scrolling ---
    const animateScroll = useCallback(() => {
      setScrollPos((prev) => {
        const diff = scrollTargetRef.current - prev;
        if (Math.abs(diff) < 0.01) {
          isAnimatingRef.current = false;
          return scrollTargetRef.current;
        }
        scrollAnimRef.current = requestAnimationFrame(animateScroll);
        return prev + diff * 0.25;
      });
    }, []);

    const startScrollAnimation = useCallback(() => {
      if (!isAnimatingRef.current) {
        isAnimatingRef.current = true;
        scrollAnimRef.current = requestAnimationFrame(animateScroll);
      }
    }, [animateScroll]);

    useEffect(() => {
      return () => cancelAnimationFrame(scrollAnimRef.current);
    }, []);

    const handleWheel = useCallback(
      (e: WheelEvent) => {
        // When dragging, only allow scroll if cursor is near the card fan
        if (draggedCard) {
          const containerRect = handRef.current?.getBoundingClientRect();
          if (!containerRect || e.clientY < containerRect.bottom - 200) return;
        }
        e.preventDefault();
        if (highlightedCard) {
          setHighlightedCard(null);
        }
        const delta = e.deltaY || e.deltaX;
        const maxScroll = Math.max(cardOrder.length - 1, 0);
        scrollTargetRef.current = clamp(
          scrollTargetRef.current + delta * WHEEL_SCALE,
          0,
          maxScroll,
        );
        startScrollAnimation();
      },
      [draggedCard, highlightedCard, cardOrder.length, startScrollAnimation],
    );

    useEffect(() => {
      if (isExpanded) {
        document.addEventListener("wheel", handleWheel, { passive: false });
        return () => document.removeEventListener("wheel", handleWheel);
      } else {
        const el = handRef.current;
        if (!el) return;
        el.addEventListener("wheel", handleWheel, { passive: false });
        return () => el.removeEventListener("wheel", handleWheel);
      }
    }, [handleWheel, isExpanded]);

    useEffect(() => {
      if (!isExpanded) return;
      const handleKeyDown = (e: KeyboardEvent) => {
        if (e.key === "ArrowLeft" || e.key === "ArrowRight") {
          e.preventDefault();
          if (highlightedCard) {
            setHighlightedCard(null);
          }
          const maxScroll = Math.max(cardOrder.length - 1, 0);
          const step = e.key === "ArrowLeft" ? -1 : 1;
          scrollTargetRef.current = clamp(scrollTargetRef.current + step, 0, maxScroll);
          startScrollAnimation();
        }
      };
      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }, [isExpanded, highlightedCard, cardOrder.length, startScrollAnimation]);

    // --- Pointer events for drag (collapsed mode only) ---
    const handlePointerDown = (cardId: string, e: React.PointerEvent<HTMLDivElement>) => {
      if (isExpanded) return;
      e.preventDefault();
      const cardEl = e.currentTarget;
      cardEl.setPointerCapture(e.pointerId);
      capturedCardRef.current = cardEl;

      dragIntentRef.current = false;

      const cardIndex = cardOrder.indexOf(cardId);
      const transform = getCardTransform(cardIndex, scrollPos);
      const containerRect = handRef.current?.getBoundingClientRect();

      if (containerRect) {
        const cardScreenX = containerRect.left + containerRect.width / 2 + transform.x;
        // Include selected lift so the card doesn't snap down on grab
        const isSelected = highlightedCard === cardId;
        const liftY = isSelected ? SELECTED_LIFT : 0;
        const cardScreenY = containerRect.bottom + transform.y + liftY;

        setDragOffset({
          x: cardScreenX - e.clientX,
          y: cardScreenY - e.clientY,
        });
      }

      setDraggedCard(cardId);
      setDragPosition({ x: e.clientX, y: e.clientY });
      setDragStartPosition({ x: e.clientX, y: e.clientY });
      setIsInThrowZone(false);
    };

    const tryReorder = useCallback(
      (pointerX: number, pointerY: number) => {
        if (!draggedCard || !dragIntentRef.current) return;
        const containerRect = handRef.current?.getBoundingClientRect();
        if (!containerRect) return;
        const cursorNearFan = pointerY >= containerRect.bottom - 80;
        if (!cursorNearFan) return;
        const relativeX = pointerX - (containerRect.left + containerRect.width / 2);
        const targetSlot = clamp(
          Math.round(relativeX / SPACING + scrollPos),
          0,
          cardOrder.length - 1,
        );
        setCardOrder((prev) => {
          const currentIdx = prev.indexOf(draggedCard);
          if (currentIdx === -1 || currentIdx === targetSlot) return prev;
          const next = [...prev];
          next.splice(currentIdx, 1);
          next.splice(targetSlot, 0, draggedCard);
          return next;
        });
      },
      [draggedCard, scrollPos, cardOrder.length],
    );

    // Re-run reorder when scrollPos changes during drag
    useEffect(() => {
      if (draggedCard && dragIntentRef.current) {
        tryReorder(lastPointerXRef.current, dragPosition.y);
      }
    }, [scrollPos, draggedCard, tryReorder, dragPosition.y]);

    const handlePointerMove = useCallback(
      (e: React.PointerEvent<HTMLDivElement>) => {
        if (!draggedCard) return;

        const deltaX = e.clientX - dragStartPosition.x;
        const deltaY = e.clientY - dragStartPosition.y;
        const movedDist = Math.sqrt(deltaX * deltaX + deltaY * deltaY);

        if (!dragIntentRef.current && movedDist > DRAG_THRESHOLD) {
          dragIntentRef.current = true;
        }

        if (!dragIntentRef.current) return;

        lastPointerXRef.current = e.clientX;
        setDragPosition({ x: e.clientX, y: e.clientY });

        tryReorder(e.clientX, e.clientY);

        const throwDist = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
        const isUpward = deltaY < THROW_Y_THRESHOLD;
        setIsInThrowZone(throwDist > THROW_DISTANCE_THRESHOLD && isUpward);
      },
      [draggedCard, dragStartPosition, tryReorder],
    );

    const handlePointerUp = useCallback(
      async (e: React.PointerEvent<HTMLDivElement>) => {
        const cardId = draggedCard;
        if (!cardId) return;

        if (capturedCardRef.current) {
          capturedCardRef.current.releasePointerCapture(e.pointerId);
          capturedCardRef.current = null;
        }

        const wasDrag = dragIntentRef.current;
        dragIntentRef.current = false;

        if (!wasDrag) {
          // This was a click, not a drag
          setDraggedCard(null);
          setIsInThrowZone(false);

          // Toggle selection
          if (highlightedCard === cardId) {
            setHighlightedCard(null);
          } else {
            void playCardHoverSound();
            setHighlightedCard(cardId);
            onCardSelect?.(cardId);
          }

          return;
        }

        // Drag ended — check throw
        const deltaX = e.clientX - dragStartPosition.x;
        const deltaY = e.clientY - dragStartPosition.y;
        const dist = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
        const isUpward = deltaY < THROW_Y_THRESHOLD;
        const isThrow = dist > THROW_DISTANCE_THRESHOLD && isUpward;

        if (isThrow && onPlayCard) {
          const cardData = cardsRef.current.find((c) => c.id === cardId);
          if (cardData?.available) {
            try {
              const containerRect = handRef.current?.getBoundingClientRect();
              const releaseX = containerRect
                ? e.clientX + dragOffset.x - (containerRect.left + containerRect.width / 2)
                : 0;
              const releaseY = containerRect ? e.clientY + dragOffset.y - containerRect.bottom : 0;

              setFlyingAwayGhost({
                card: cardData,
                x: releaseX,
                y: releaseY,
                scale: fanScale,
                animating: false,
              });
              setDraggedCard(null);
              setIsInThrowZone(false);
              setHighlightedCard(null);

              requestAnimationFrame(() => {
                requestAnimationFrame(() => {
                  setFlyingAwayGhost((prev) =>
                    prev ? { ...prev, scale: prev.scale * 0.85, animating: true } : prev,
                  );
                });
              });

              setTimeout(() => {
                setFlyingAwayGhost(null);
              }, 400);

              await onPlayCard(cardId);
              return;
            } catch (error) {
              console.error("Failed to play card:", error);
              setFlyingAwayGhost(null);
            }
          }
        }

        // Return card to hand with animation
        setReturningCard(cardId);
        setDraggedCard(null);
        setIsInThrowZone(false);

        setTimeout(() => {
          setReturningCard(null);
        }, 400);
      },
      [
        draggedCard,
        dragStartPosition,
        dragOffset,
        fanScale,
        highlightedCard,
        onPlayCard,
        onCardSelect,
        playCardHoverSound,
      ],
    );

    // --- Click outside to deselect ---
    useEffect(() => {
      const handleDocumentClick = (event: MouseEvent) => {
        if (handRef.current && !handRef.current.contains(event.target as Node)) {
          setHighlightedCard(null);
        }
      };
      document.addEventListener("click", handleDocumentClick);
      return () => document.removeEventListener("click", handleDocumentClick);
    }, []);

    if (hideWhenModalOpen || (cards.length === 0 && !flyingAwayGhost)) {
      return null;
    }

    const activeCullRadius = isExpanded ? expandedCullRadius : CULL_RADIUS;
    const activeVisibleRadius = isExpanded ? expandedCullRadius : VISIBLE_RADIUS;

    return (
      <>
        <BlurredOverlay visible={isExpanded} onClose={() => handleCollapse()} zIndex={20200}>
          <div />
        </BlurredOverlay>

        <div
          className="card-fan-overlay"
          ref={handRef}
          style={isExpanded ? { zIndex: Z_INDEX.TOP_MENU_ALWAYS_ON_TOP } : undefined}
        >
          {cards.map((card) => {
            const index = cardOrder.indexOf(card.id);
            if (index === -1) return null;

            const isDraggedCard = draggedCard === card.id;
            if (flyingAwayGhost?.card.id === card.id) return null;

            const absD = Math.abs(index - scrollPos);
            if (absD > activeCullRadius && !isDraggedCard) return null;

            const edgeOpacity = isExpanded || isDraggedCard || absD <= activeVisibleRadius ? 1 : 0;
            const isDragging = isDraggedCard && dragIntentRef.current;
            const isReturning = returningCard === card.id;
            const isHighlighted = highlightedCard === card.id;

            let finalX: number;
            let finalY: number;
            let finalRotation: number;
            let finalScale: number;
            let finalZ: number;
            let showErrors = false;

            if (isExpanded) {
              const d = index - scrollPos;
              finalX = d * EXPANDED_SPACING;
              finalY = expandedBaseY;
              finalRotation = 0;
              finalScale = 1;
              finalZ = index;
            } else {
              const base = getCardTransform(index, scrollPos);
              finalX = base.x * fanScale;
              finalY = base.y * fanScale;
              finalRotation = base.rotation;
              finalScale = fanScale;
              finalZ = base.z;

              if (isHighlighted && !isDraggedCard) {
                finalY += SELECTED_LIFT * fanScale;
                finalScale = SELECTED_SCALE * fanScale;
                finalRotation = 0;
                finalZ = 2000;
              }

              let isDragRaised = false;
              if (isDraggedCard && !isReturning) {
                const containerRect = handRef.current?.getBoundingClientRect();
                if (containerRect) {
                  finalX =
                    dragPosition.x + dragOffset.x - (containerRect.left + containerRect.width / 2);
                  finalY = dragPosition.y + dragOffset.y - containerRect.bottom;
                  finalRotation = 0;
                  finalScale = fanScale;
                  finalZ = 3000;

                  const dragDeltaY = dragPosition.y - dragStartPosition.y;
                  isDragRaised = dragIntentRef.current && dragDeltaY < DRAG_RAISE_THRESHOLD;
                }
              }

              const hasErrors = !card.available && card.errors.length > 0;
              const hasWarnings = (card.warnings?.length ?? 0) > 0;
              showErrors =
                (hasErrors || hasWarnings) && (isHighlighted || (isDraggedCard && isDragRaised));
            }

            const staggerDelay = isTransitioning
              ? `${Math.abs(index - Math.round(scrollPos)) * 25}ms`
              : undefined;

            const classNames = [
              "card-fan-card",
              isTransitioning ? "is-mode-transition" : "",
              !isExpanded && isDragging && !isReturning ? "is-dragging" : "",
              !isExpanded && !isDraggedCard && draggedCard ? "is-reordering" : "",
              !isExpanded && isReturning ? "is-returning" : "",
              !isExpanded && isDraggedCard && isInThrowZone && card.available
                ? "is-throw-zone"
                : "",
              isExpanded ? "is-expanded" : "",
            ]
              .filter(Boolean)
              .join(" ");

            return (
              <div
                key={card.id}
                data-card-id={card.id}
                className={classNames}
                style={{
                  transform: `translate(${finalX}px, ${finalY}px) rotate(${finalRotation}deg) scale(${finalScale})`,
                  zIndex: finalZ,
                  opacity: edgeOpacity,
                  pointerEvents: edgeOpacity === 0 ? "none" : undefined,
                  transitionDelay: staggerDelay,
                }}
                onPointerDown={isExpanded ? undefined : (e) => handlePointerDown(card.id, e)}
                onPointerMove={isExpanded ? undefined : handlePointerMove}
                onPointerUp={isExpanded ? undefined : handlePointerUp}
                onClick={isExpanded ? () => handleCollapse(card.id) : undefined}
              >
                <GameCard
                  card={card}
                  isSelected={
                    !isExpanded &&
                    (isHighlighted || (isDraggedCard && isInThrowZone && card.available === true))
                  }
                  onSelect={() => {}}
                  animationDelay={-1}
                />
                {!isExpanded &&
                  (!card.available || (card.warnings && card.warnings.length > 0)) && (
                    <div className={`card-fan-error-panel ${showErrors ? "is-visible" : ""}`}>
                      {!card.available &&
                        card.errors.map((err, i) => (
                          <div key={i} className="card-fan-error-item">
                            {err.message}
                          </div>
                        ))}
                      {card.warnings &&
                        card.warnings.map((warn, i) => (
                          <div key={i} className="card-fan-warning-item">
                            {warn.message}
                          </div>
                        ))}
                    </div>
                  )}
              </div>
            );
          })}

          {flyingAwayGhost && (
            <div
              className={`card-fan-card ${flyingAwayGhost.animating ? "is-flying-away" : "is-flying-away-start"}`}
              style={{
                transform: `translate(${flyingAwayGhost.x}px, ${flyingAwayGhost.y}px) scale(${flyingAwayGhost.scale})`,
                zIndex: Z_INDEX.CARD_DETAIL_MODAL,
              }}
            >
              <GameCard
                card={flyingAwayGhost.card}
                isSelected={true}
                onSelect={() => {}}
                animationDelay={-1}
              />
            </div>
          )}

          <style>{`
        .card-fan-overlay {
          position: fixed;
          bottom: 48px;
          left: 50%;
          transform: translateX(-50%);
          width: 0;
          height: 300px;
          z-index: 1100;
          pointer-events: none;
        }

        .card-fan-card {
          position: absolute;
          bottom: 0;
          left: 50%;
          margin-left: -100px;
          cursor: pointer;
          transform-origin: bottom center;
          pointer-events: auto;
          user-select: none;
          touch-action: none;
          transition: transform 180ms ease, filter 180ms ease, opacity 180ms ease;
        }

        .card-fan-card.is-mode-transition {
          transition: transform 280ms cubic-bezier(0.25, 0.46, 0.45, 0.94),
                      opacity 210ms ease !important;
        }

        .card-fan-card.is-expanded {
          cursor: pointer;
        }

        .card-fan-card.is-dragging {
          transition: none;
          cursor: grabbing;
        }

        .card-fan-card.is-reordering {
          transition: transform 200ms ease, opacity 180ms ease;
        }

        .card-fan-card.is-returning {
          transition: transform 400ms cubic-bezier(0.25, 0.46, 0.45, 0.94),
                      filter 180ms ease,
                      opacity 180ms ease;
        }

        .card-fan-card.is-throw-zone {
          filter: brightness(1.15);
        }

        .card-fan-card.is-flying-away-start {
          transition: none !important;
          pointer-events: none !important;
        }

        .card-fan-card.is-flying-away {
          transition: transform 350ms ease-out, opacity 350ms ease-out !important;
          opacity: 0 !important;
          pointer-events: none !important;
        }



        .card-fan-error-panel {
          position: absolute;
          left: 100%;
          top: 0;
          margin-left: 10px;
          width: 180px;
          display: flex;
          flex-direction: column;
          gap: 4px;
          pointer-events: none;
          opacity: 0;
          transform: translateX(-6px);
          transition: opacity 200ms ease, transform 200ms ease;
        }

        .card-fan-error-panel.is-visible {
          opacity: 1;
          transform: translateX(0);
        }

        .card-fan-error-item {
          background: rgba(10, 10, 15, 0.95);
          border: 1px solid rgba(231, 76, 60, 0.6);
          border-left: 3px solid #e74c3c;
          color: rgba(255, 255, 255, 0.9);
          font-size: 12px;
          line-height: 1.4;
          padding: 8px 10px;
          white-space: normal;
        }

        .card-fan-warning-item {
          background: rgba(10, 10, 15, 0.95);
          border: 1px solid rgba(255, 193, 7, 0.6);
          border-left: 3px solid #ffc107;
          color: rgba(255, 255, 255, 0.9);
          font-size: 12px;
          line-height: 1.4;
          padding: 8px 10px;
          white-space: normal;
        }

        @media (max-width: 1200px) {
          .card-fan-overlay {
            height: 250px;
          }
        }

        @media (max-width: 768px) {
          .card-fan-overlay {
            height: 200px;
          }
        }
      `}</style>
        </div>
      </>
    );
  },
);

export default CardFanOverlay;
