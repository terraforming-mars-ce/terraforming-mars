import {
  createContext,
  useContext,
  useLayoutEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import BackButton from "./buttons/BackButton.tsx";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";
import { useVisualViewport } from "@/hooks/useVisualViewport.ts";
import { useRenderPause } from "@/stores/renderPauseStore.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

const PanelAnimationContext = createContext(false);

export function AnimatedHeight({ children }: { children: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);
  const [height, setHeight] = useState<number>();
  useLayoutEffect(() => {
    const element = ref.current;
    if (!element) return;
    const observer = new ResizeObserver(() => setHeight(element.getBoundingClientRect().height));
    observer.observe(element);
    setHeight(element.getBoundingClientRect().height);
    return () => observer.disconnect();
  }, []);
  useLayoutEffect(() => {
    if (ref.current) setHeight(ref.current.getBoundingClientRect().height);
  });
  return (
    <div
      className="overflow-hidden transition-[height] duration-200 ease-out motion-reduce:transition-none"
      style={{ height }}
    >
      <div ref={ref}>{children}</div>
    </div>
  );
}

export function SaveFeedback({
  message,
  error = false,
  className = "",
}: {
  message: string;
  error?: boolean;
  className?: string;
}) {
  const parentAnimates = useContext(PanelAnimationContext);
  const content = message ? (
    <p
      role={error ? "alert" : "status"}
      className={`py-3 compact:py-1 text-sm break-words ${error ? "text-red-300" : "text-white/70"} ${className}`}
    >
      {message}
    </p>
  ) : null;
  return parentAnimates ? content : <AnimatedHeight>{content}</AnimatedHeight>;
}

export default function SaveGameLayout({
  title,
  expanded = false,
  onBack,
  footer,
  children,
}: {
  title: string;
  expanded?: boolean;
  onBack: () => void;
  footer?: ReactNode;
  children: ReactNode;
}) {
  const { isCompact } = useLayoutMode();
  const viewport = useVisualViewport(isCompact);
  useRenderPause("save-page", isCompact);
  if (isCompact) {
    return (
      <main
        className="fixed inset-x-0 flex flex-col bg-black text-white text-left"
        style={{
          zIndex: Z_INDEX.MOBILE_SCREEN,
          top: viewport.offsetTop,
          height: viewport.height,
          paddingTop: "var(--safe-top)",
          paddingBottom: "var(--safe-bottom)",
          paddingLeft: "var(--safe-left)",
          paddingRight: "var(--safe-right)",
        }}
      >
        <header className="flex shrink-0 items-center gap-4 border-b border-white/15 px-3 min-h-14">
          <BackButton onClick={onBack} className="min-h-11" />
          <h1 className="font-orbitron text-lg font-semibold m-0">{title}</h1>
        </header>
        <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain p-4">{children}</div>
        {footer && (
          <footer className="shrink-0 border-t border-white/15 px-4 py-2">{footer}</footer>
        )}
      </main>
    );
  }
  return (
    <>
      <div
        className="menu-chrome-top-left menu-chrome-back"
        style={{ zIndex: Z_INDEX.TOP_MENU_BAR }}
      >
        <BackButton onClick={onBack} />
      </div>
      <main
        className="menu-shell text-white text-left relative transition-[width] duration-200 ease-out motion-reduce:transition-none"
        style={{
          zIndex: Z_INDEX.UI_BASE,
          width: `min(${expanded ? 1100 : 450}px, calc(100% - 48px - var(--safe-left) - var(--safe-right)))`,
        }}
      >
        <h1 className="menu-title">{title}</h1>
        <div className="game-panel [&>div:has([role=listbox])]:overflow-visible">
          <PanelAnimationContext.Provider value={true}>
            <AnimatedHeight>
              <div
                style={{
                  // Measure the destination layout immediately so wrapping cannot
                  // restart the height transition while the panel widens.
                  width: `min(${expanded ? 1100 : 450}px, calc(100vw - 48px - var(--safe-left) - var(--safe-right)))`,
                }}
              >
                <div className="p-6 sm:p-8">{children}</div>
                {footer && <footer className="px-8 pb-8">{footer}</footer>}
              </div>
            </AnimatedHeight>
          </PanelAnimationContext.Provider>
        </div>
      </main>
    </>
  );
}
