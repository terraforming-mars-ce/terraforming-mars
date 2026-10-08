import { useEffect, useRef, type MouseEvent, type PointerEvent } from "react";

interface LongPressOptions {
  delayMs?: number;
  moveTolerancePx?: number;
}

export interface LongPressPoint {
  x: number;
  y: number;
}

export interface LongPressHandlers<T extends Element> {
  onPointerDown: (event: PointerEvent<T>) => void;
  onPointerMove: (event: PointerEvent<T>) => void;
  onPointerUp: () => void;
  onPointerCancel: () => void;
  onClickCapture: (event: MouseEvent<T>) => void;
  onContextMenu: (event: MouseEvent<T>) => void;
}

const GUARD_SETTLE_MS = 400;

function swallow(event: Event) {
  event.preventDefault();
  event.stopPropagation();
}

/**
 * Whatever the long press opens usually ends up under the finger, so the browser's touch
 * context menu and the trailing click land on it instead of on the trigger. Swallow both
 * document-wide until the finger lifts.
 */
function guardFollowUpEvents(): () => void {
  let settleTimer: ReturnType<typeof setTimeout> | null = null;
  const release = () => {
    if (settleTimer !== null) {
      clearTimeout(settleTimer);
    }
    document.removeEventListener("contextmenu", swallow, true);
    document.removeEventListener("click", swallowClick, true);
    document.removeEventListener("pointerup", settle, true);
    document.removeEventListener("pointercancel", settle, true);
  };
  const swallowClick = (event: Event) => {
    swallow(event);
    document.removeEventListener("click", swallowClick, true);
  };
  const settle = () => {
    if (settleTimer === null) {
      settleTimer = setTimeout(release, GUARD_SETTLE_MS);
    }
  };
  document.addEventListener("contextmenu", swallow, true);
  document.addEventListener("click", swallowClick, true);
  document.addEventListener("pointerup", settle, true);
  document.addEventListener("pointercancel", settle, true);
  return release;
}

/**
 * Long-press for touch and pen pointers; mouse users keep right-click. Suppresses the click
 * that follows a long press and the browser's own touch context menu.
 */
export function useLongPress<T extends Element = Element>(
  onLongPress: (point: LongPressPoint) => void,
  { delayMs = 500, moveTolerancePx = 10 }: LongPressOptions = {},
): LongPressHandlers<T> {
  const onLongPressRef = useRef(onLongPress);
  onLongPressRef.current = onLongPress;
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const startRef = useRef<LongPressPoint | null>(null);
  const touchActiveRef = useRef(false);
  const firedRef = useRef(false);

  const releaseGuardRef = useRef<(() => void) | null>(null);

  const cancel = () => {
    if (timerRef.current !== null) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    startRef.current = null;
  };

  useEffect(
    () => () => {
      cancel();
      releaseGuardRef.current?.();
    },
    [],
  );

  return {
    onPointerDown: (event) => {
      cancel();
      firedRef.current = false;
      touchActiveRef.current = event.pointerType !== "mouse";
      if (!touchActiveRef.current || !event.isPrimary) {
        return;
      }
      const start = { x: event.clientX, y: event.clientY };
      startRef.current = start;
      timerRef.current = setTimeout(() => {
        timerRef.current = null;
        startRef.current = null;
        firedRef.current = true;
        releaseGuardRef.current?.();
        releaseGuardRef.current = guardFollowUpEvents();
        navigator.vibrate?.(10);
        onLongPressRef.current(start);
      }, delayMs);
    },
    onPointerMove: (event) => {
      const start = startRef.current;
      if (!start) {
        return;
      }
      if (Math.hypot(event.clientX - start.x, event.clientY - start.y) > moveTolerancePx) {
        cancel();
      }
    },
    onPointerUp: cancel,
    onPointerCancel: cancel,
    onClickCapture: (event) => {
      if (firedRef.current) {
        firedRef.current = false;
        event.preventDefault();
        event.stopPropagation();
      }
    },
    onContextMenu: (event) => {
      if (touchActiveRef.current) {
        event.preventDefault();
        event.stopPropagation();
      }
    },
  };
}
