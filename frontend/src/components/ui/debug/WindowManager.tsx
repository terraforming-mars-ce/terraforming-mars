import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import type { ReactNode, MouseEvent as ReactMouseEvent } from "react";
import { Z_INDEX } from "../../../constants/zIndex";

interface WindowManagerContextType {
  bringToFront: (windowId: string) => void;
  getZIndex: (windowId: string) => number;
  registerWindow: (windowId: string) => void;
  unregisterWindow: (windowId: string) => void;
}

const WindowManagerContext = createContext<WindowManagerContextType | null>(null);

const BASE_Z_INDEX = Z_INDEX.DEBUG_WINDOWS;

export function WindowManagerProvider({ children }: { children: ReactNode }) {
  const [focusStack, setFocusStack] = useState<string[]>([]);

  const registerWindow = useCallback((windowId: string) => {
    setFocusStack((prev) => {
      if (prev.includes(windowId)) return prev;
      return [...prev, windowId];
    });
  }, []);

  const unregisterWindow = useCallback((windowId: string) => {
    setFocusStack((prev) => prev.filter((id) => id !== windowId));
  }, []);

  const bringToFront = useCallback((windowId: string) => {
    setFocusStack((prev) => {
      if (prev.at(-1) === windowId) {
        return prev;
      }
      const filtered = prev.filter((id) => id !== windowId);
      return [...filtered, windowId];
    });
  }, []);

  const getZIndex = useCallback(
    (windowId: string) => {
      const index = focusStack.indexOf(windowId);
      if (index === -1) return BASE_Z_INDEX;
      return BASE_Z_INDEX + index;
    },
    [focusStack],
  );

  return (
    <WindowManagerContext.Provider
      value={{ bringToFront, getZIndex, registerWindow, unregisterWindow }}
    >
      {children}
    </WindowManagerContext.Provider>
  );
}

export function useWindowManager() {
  const context = useContext(WindowManagerContext);
  if (!context) {
    throw new Error("useWindowManager must be used within WindowManagerProvider");
  }
  return context;
}

interface UseWindowDragOptions {
  windowId: string;
  width: number;
  height: number | (() => number);
  defaultPosition?: { x: number; y: number };
  excludeSelectors?: string[];
  isVisible?: boolean;
}

export function useWindowDrag({
  windowId,
  width,
  height,
  defaultPosition,
  excludeSelectors,
  isVisible = true,
}: UseWindowDragOptions) {
  const { bringToFront, registerWindow, unregisterWindow } = useWindowManager();
  const excludeSelectorsRef = useRef(excludeSelectors);
  excludeSelectorsRef.current = excludeSelectors;
  const prevVisibleRef = useRef(isVisible);

  useEffect(() => {
    registerWindow(windowId);
    return () => unregisterWindow(windowId);
  }, [windowId, registerWindow, unregisterWindow]);

  useEffect(() => {
    if (isVisible && !prevVisibleRef.current) {
      bringToFront(windowId);
    }
    prevVisibleRef.current = isVisible;
  }, [isVisible, bringToFront, windowId]);

  const getHeight = useCallback(() => (typeof height === "function" ? height() : height), [height]);

  const [position, setPosition] = useState(() => {
    if (defaultPosition) return defaultPosition;
    if (typeof window === "undefined") return { x: 100, y: 60 };
    return {
      x: (window.innerWidth - width) / 2,
      y: 60,
    };
  });

  const [isDragging, setIsDragging] = useState(false);
  const drag = useRef<{
    element: HTMLElement;
    x: number;
    y: number;
    startX: number;
    startY: number;
    offsetX: number;
    offsetY: number;
    frame: number;
  } | null>(null);
  useLayoutEffect(() => {
    if (!isDragging && drag.current) {
      cancelAnimationFrame(drag.current.frame);
      drag.current.element.style.transform = "";
      drag.current.element.style.willChange = "";
      drag.current = null;
    }
  }, [isDragging]);

  const handleMouseDown = useCallback(
    (e: ReactMouseEvent) => {
      bringToFront(windowId);

      const target = e.target as HTMLElement;

      if (
        target.tagName === "BUTTON" ||
        target.tagName === "INPUT" ||
        target.closest("button") ||
        target.closest("input")
      ) {
        return;
      }

      const selectors = excludeSelectorsRef.current;
      if (selectors) {
        for (const selector of selectors) {
          if (target.closest(selector)) return;
        }
      }

      e.preventDefault();
      const element = e.currentTarget as HTMLElement;
      const rect = element.getBoundingClientRect();
      drag.current = {
        element,
        x: rect.left,
        y: rect.top,
        startX: rect.left,
        startY: rect.top,
        offsetX: e.clientX - rect.left,
        offsetY: e.clientY - rect.top,
        frame: 0,
      };
      element.style.willChange = "transform";
      setIsDragging(true);
    },
    [bringToFront, windowId],
  );

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      const current = drag.current;
      if (!current) {
        return;
      }
      const renderedWidth = Math.min(width, window.innerWidth - 32);
      current.x = Math.max(
        16,
        Math.min(window.innerWidth - renderedWidth - 16, e.clientX - current.offsetX),
      );
      current.y = Math.max(16, Math.min(window.innerHeight - 80, e.clientY - current.offsetY));
      if (!current.frame) {
        current.frame = requestAnimationFrame(() => {
          current.frame = 0;
          current.element.style.transform = `translate3d(${current.x - current.startX}px, ${current.y - current.startY}px, 0)`;
        });
      }
    },
    [width],
  );

  const handleMouseUp = useCallback(() => {
    if (drag.current) {
      setPosition({ x: drag.current.x, y: drag.current.y });
    }
    setIsDragging(false);
  }, []);

  useEffect(() => {
    if (!isDragging) return;
    document.body.style.userSelect = "none";
    document.body.style.cursor = "default";
    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    window.addEventListener("blur", handleMouseUp);
    return () => {
      document.body.style.userSelect = "";
      document.body.style.cursor = "";
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
      window.removeEventListener("blur", handleMouseUp);
      if (drag.current) {
        cancelAnimationFrame(drag.current.frame);
        drag.current.frame = 0;
        drag.current.element.style.transform = "";
        drag.current.element.style.willChange = "";
      }
    };
  }, [isDragging, handleMouseMove, handleMouseUp]);

  useEffect(() => {
    const handleResize = () => {
      const h = getHeight();
      const screenWidth = window.innerWidth;
      const screenHeight = window.innerHeight;

      setPosition((prev) => ({
        x: Math.max(-(width / 2), Math.min(screenWidth - width / 2, prev.x)),
        y: Math.max(-(h / 2), Math.min(screenHeight - h / 2, prev.y)),
      }));
    };
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, [width, getHeight]);

  return { position, isDragging, handleMouseDown };
}
