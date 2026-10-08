import { useRef, type MouseEvent, type PointerEvent, type TouchEvent } from "react";

interface SwipeOptions {
  onSwipeLeft?: () => void;
  onSwipeRight?: () => void;
  thresholdPx?: number;
}

interface Point {
  x: number;
  y: number;
}

/**
 * Horizontal swipe from touch events (not cancelled when the browser starts a scroll)
 * plus mouse drags. Suppresses the click that ends a swipe.
 */
export function useHorizontalSwipe({ onSwipeLeft, onSwipeRight, thresholdPx = 50 }: SwipeOptions) {
  const startRef = useRef<Point | null>(null);
  const swipedRef = useRef(false);

  const begin = (point: Point) => {
    startRef.current = point;
    swipedRef.current = false;
  };

  const end = (point: Point) => {
    const start = startRef.current;
    startRef.current = null;
    if (!start) {
      return;
    }
    const dx = point.x - start.x;
    const dy = point.y - start.y;
    if (Math.abs(dx) < thresholdPx || Math.abs(dx) <= Math.abs(dy)) {
      return;
    }
    swipedRef.current = true;
    if (dx < 0) {
      onSwipeLeft?.();
    } else {
      onSwipeRight?.();
    }
  };

  return {
    onTouchStart: (event: TouchEvent) => {
      if (event.touches.length !== 1) {
        startRef.current = null;
        return;
      }
      const touch = event.touches[0];
      begin({ x: touch.clientX, y: touch.clientY });
    },
    onTouchEnd: (event: TouchEvent) => {
      const touch = event.changedTouches[0];
      if (!touch) {
        startRef.current = null;
        return;
      }
      end({ x: touch.clientX, y: touch.clientY });
    },
    onTouchCancel: () => {
      startRef.current = null;
    },
    onPointerDown: (event: PointerEvent) => {
      if (event.pointerType === "mouse" && event.button === 0) {
        begin({ x: event.clientX, y: event.clientY });
      }
    },
    onPointerUp: (event: PointerEvent) => {
      if (event.pointerType === "mouse") {
        end({ x: event.clientX, y: event.clientY });
      }
    },
    onClickCapture: (event: MouseEvent) => {
      if (swipedRef.current) {
        swipedRef.current = false;
        event.preventDefault();
        event.stopPropagation();
      }
    },
  };
}
