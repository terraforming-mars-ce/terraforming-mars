import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type FocusEvent,
  type KeyboardEvent,
  type MouseEvent,
  type PointerEvent,
  type RefObject,
} from "react";

const TAP_MOVE_TOLERANCE_PX = 10;

let closeActiveReveal: (() => void) | null = null;

function hasKeyboardFocus(element: HTMLElement): boolean {
  const focused = document.activeElement;
  return !!focused && element.contains(focused) && focused.matches(":focus-visible");
}

export interface TapRevealTriggerProps<T extends HTMLElement> {
  ref: RefObject<T | null>;
  onPointerEnter: (event: PointerEvent<T>) => void;
  onPointerLeave: (event: PointerEvent<T>) => void;
  onPointerDown: (event: PointerEvent<T>) => void;
  onPointerUp: (event: PointerEvent<T>) => void;
  onPointerCancel: () => void;
  onClick: (event: MouseEvent<T>) => void;
  onFocus: (event: FocusEvent<T>) => void;
  onBlur: () => void;
  onKeyDown: (event: KeyboardEvent<T>) => void;
  "aria-expanded": boolean | undefined;
}

export interface TapReveal<T extends HTMLElement> {
  open: boolean;
  close: () => void;
  triggerRef: RefObject<T | null>;
  triggerProps: TapRevealTriggerProps<T>;
}

/**
 * Hover reveal for mouse pointers, tap-to-toggle for touch and pen. Touch taps on the
 * trigger stop propagating so they never also select, inspect or drag what contains it.
 */
export function useTapReveal<T extends HTMLElement = HTMLElement>({
  enabled = true,
}: { enabled?: boolean } = {}): TapReveal<T> {
  const triggerRef = useRef<T>(null);
  const [isOpen, setIsOpen] = useState(false);
  const pressRef = useRef<{ x: number; y: number } | null>(null);
  const lastPointerTypeRef = useRef("mouse");
  const open = enabled && isOpen;

  const close = useCallback(() => setIsOpen(false), []);
  const show = useCallback(() => {
    if (closeActiveReveal && closeActiveReveal !== close) {
      closeActiveReveal();
    }
    closeActiveReveal = close;
    setIsOpen(true);
  }, [close]);

  useEffect(() => {
    if (!enabled) {
      setIsOpen(false);
    }
  }, [enabled]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const handlePointerDown = (event: globalThis.PointerEvent) => {
      if (triggerRef.current?.contains(event.target as Node)) {
        return;
      }
      close();
    };
    const handleScroll = (event: Event) => {
      const trigger = triggerRef.current;
      if (!trigger || !(event.target instanceof Node) || event.target.contains(trigger)) {
        close();
      }
    };
    document.addEventListener("pointerdown", handlePointerDown, true);
    window.addEventListener("scroll", handleScroll, true);
    window.addEventListener("resize", close);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown, true);
      window.removeEventListener("scroll", handleScroll, true);
      window.removeEventListener("resize", close);
      if (closeActiveReveal === close) {
        closeActiveReveal = null;
      }
    };
  }, [open, close]);

  const triggerProps: TapRevealTriggerProps<T> = {
    ref: triggerRef,
    onPointerEnter: (event) => {
      lastPointerTypeRef.current = event.pointerType;
      if (enabled && event.pointerType === "mouse") {
        show();
      }
    },
    onPointerLeave: (event) => {
      if (event.pointerType === "mouse" && !hasKeyboardFocus(event.currentTarget)) {
        close();
      }
    },
    onPointerDown: (event) => {
      lastPointerTypeRef.current = event.pointerType;
      if (!enabled || event.pointerType === "mouse") {
        return;
      }
      event.stopPropagation();
      pressRef.current = { x: event.clientX, y: event.clientY };
    },
    onPointerUp: (event) => {
      if (!enabled || event.pointerType === "mouse") {
        return;
      }
      event.stopPropagation();
      const press = pressRef.current;
      pressRef.current = null;
      if (
        !press ||
        Math.hypot(event.clientX - press.x, event.clientY - press.y) > TAP_MOVE_TOLERANCE_PX
      ) {
        return;
      }
      if (isOpen) {
        close();
      } else {
        show();
      }
    },
    onPointerCancel: () => {
      pressRef.current = null;
    },
    onClick: (event) => {
      if (enabled && lastPointerTypeRef.current !== "mouse") {
        event.stopPropagation();
      }
    },
    onFocus: (event) => {
      if (enabled && hasKeyboardFocus(event.currentTarget)) {
        show();
      }
    },
    onBlur: close,
    onKeyDown: (event) => {
      if (event.key === "Escape" && open) {
        event.stopPropagation();
        close();
      }
    },
    "aria-expanded": enabled ? open : undefined,
  };

  return { open, close, triggerRef, triggerProps };
}
