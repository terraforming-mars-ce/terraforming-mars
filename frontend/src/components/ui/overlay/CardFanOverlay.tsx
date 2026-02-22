import React, { useState, useEffect, useRef, useCallback } from "react";
import GameCard from "../cards/GameCard.tsx";
import { PlayerCardDto } from "@/types/generated/api-types.ts";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

function clamp(n: number, min: number, max: number): number {
  return Math.min(Math.max(n, min), max);
}

// --- Layout constants ---
const SPACING = 120;
const FLAT_RADIUS = 2.5;
const TURN_RADIUS = 4;
const MAX_ROTATE = 22;
const EDGE_DROP = 18;
const BASE_Y_OFFSET = 160;
const VISIBLE_RADIUS = 4;
const CULL_RADIUS = VISIBLE_RADIUS + 1;

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

interface CardFanOverlayProps {
  cards: PlayerCardDto[];
  hideWhenModalOpen?: boolean;
  onCardSelect?: (cardId: string) => void;
  onPlayCard?: (cardId: string) => Promise<void>;
}

const CardFanOverlay: React.FC<CardFanOverlayProps> = ({
  cards,
  hideWhenModalOpen = false,
  onCardSelect,
  onPlayCard,
}) => {
  const [scrollPos, setScrollPos] = useState(0);
  const [cardOrder, setCardOrder] = useState<string[]>([]);
  const [highlightedCard, setHighlightedCard] = useState<string | null>(null);
  const [draggedCard, setDraggedCard] = useState<string | null>(null);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const [dragPosition, setDragPosition] = useState({ x: 0, y: 0 });
  const [dragStartPosition, setDragStartPosition] = useState({ x: 0, y: 0 });
  const [isInThrowZone, setIsInThrowZone] = useState(false);
  const [returningCard, setReturningCard] = useState<string | null>(null);

  const handRef = useRef<HTMLDivElement>(null);
  const cardsRef = useRef(cards);
  const dragIntentRef = useRef(false);
  const capturedCardRef = useRef<HTMLDivElement | null>(null);
  const lastPointerXRef = useRef(0);

  const { playCardHoverSound } = useSoundEffects();

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
  const scrollTargetRef = useRef(0);
  const scrollAnimRef = useRef(0);
  const isAnimatingRef = useRef(false);

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
      scrollTargetRef.current = clamp(scrollTargetRef.current + delta * WHEEL_SCALE, 0, maxScroll);
      startScrollAnimation();
    },
    [draggedCard, highlightedCard, cardOrder.length, startScrollAnimation],
  );

  useEffect(() => {
    const el = handRef.current;
    if (!el) return;
    el.addEventListener("wheel", handleWheel, { passive: false });
    return () => el.removeEventListener("wheel", handleWheel);
  }, [handleWheel]);

  // --- Pointer events for drag ---
  const handlePointerDown = (cardId: string, e: React.PointerEvent<HTMLDivElement>) => {
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
            await onPlayCard(cardId);
            setDraggedCard(null);
            setIsInThrowZone(false);

            setHighlightedCard(null);
            return;
          } catch (error) {
            console.error("Failed to play card:", error);
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
    [draggedCard, dragStartPosition, highlightedCard, onPlayCard, onCardSelect, playCardHoverSound],
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

  if (hideWhenModalOpen || cards.length === 0) {
    return null;
  }

  return (
    <div className="card-fan-overlay" ref={handRef}>
      {cards.map((card) => {
        const index = cardOrder.indexOf(card.id);
        if (index === -1) return null;

        const isDraggedCard = draggedCard === card.id;

        const absD = Math.abs(index - scrollPos);
        if (absD > CULL_RADIUS && !isDraggedCard) return null;

        const edgeOpacity = isDraggedCard || absD <= VISIBLE_RADIUS ? 1 : 0;
        const isDragging = isDraggedCard && dragIntentRef.current;
        const isReturning = returningCard === card.id;
        const isHighlighted = highlightedCard === card.id;

        const base = getCardTransform(index, scrollPos);
        let finalX = base.x;
        let finalY = base.y;
        let finalRotation = base.rotation;
        let finalScale = base.scale;
        let finalZ = base.z;

        // Selected overlay (click to raise)
        if (isHighlighted && !isDraggedCard) {
          finalY += SELECTED_LIFT;
          finalScale = SELECTED_SCALE;
          finalRotation = 0;
          finalZ = 2000;
        }

        // Dragged card follows pointer
        let isDragRaised = false;
        if (isDraggedCard && !isReturning) {
          const containerRect = handRef.current?.getBoundingClientRect();
          if (containerRect) {
            finalX = dragPosition.x + dragOffset.x - (containerRect.left + containerRect.width / 2);
            finalY = dragPosition.y + dragOffset.y - containerRect.bottom;
            finalRotation = 0;
            finalScale = 1;
            finalZ = 3000;

            const dragDeltaY = dragPosition.y - dragStartPosition.y;
            isDragRaised = dragIntentRef.current && dragDeltaY < DRAG_RAISE_THRESHOLD;
          }
        }

        const showErrors =
          !card.available &&
          card.errors.length > 0 &&
          (isHighlighted || (isDraggedCard && isDragRaised));

        const classNames = [
          "card-fan-card",
          isDragging && !isReturning ? "is-dragging" : "",
          !isDraggedCard && draggedCard ? "is-reordering" : "",
          isReturning ? "is-returning" : "",
          isDraggedCard && isInThrowZone && card.available ? "is-throw-zone" : "",
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
            }}
            onPointerDown={(e) => handlePointerDown(card.id, e)}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerUp}
          >
            <GameCard
              card={card}
              isSelected={
                isHighlighted || (isDraggedCard && isInThrowZone && card.available === true)
              }
              onSelect={() => {}}
              animationDelay={-1}
            />
            {!card.available && card.errors.length > 0 && (
              <div className={`card-fan-error-panel ${showErrors ? "is-visible" : ""}`}>
                {card.errors.map((err, i) => (
                  <div key={i} className="card-fan-error-item">
                    {err.message}
                  </div>
                ))}
              </div>
            )}
          </div>
        );
      })}

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
  );
};

export default CardFanOverlay;
