import { createContext, useContext, useState, useCallback, useEffect, useRef } from "react";
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
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });

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
      setIsDragging(true);
      setDragStart({
        x: e.clientX - position.x,
        y: e.clientY - position.y,
      });
    },
    [position, bringToFront, windowId],
  );

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging) return;
      const h = getHeight();
      const screenWidth = window.innerWidth;
      const screenHeight = window.innerHeight;

      const minX = -(width / 2);
      const maxX = screenWidth - width / 2;
      const minY = -(h / 2);
      const maxY = screenHeight - h / 2;

      setPosition({
        x: Math.max(minX, Math.min(maxX, e.clientX - dragStart.x)),
        y: Math.max(minY, Math.min(maxY, e.clientY - dragStart.y)),
      });
    },
    [isDragging, dragStart, width, getHeight],
  );

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  useEffect(() => {
    if (!isDragging) return;
    document.body.style.userSelect = "none";
    document.body.style.cursor = "grabbing";
    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    return () => {
      document.body.style.userSelect = "";
      document.body.style.cursor = "";
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
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
