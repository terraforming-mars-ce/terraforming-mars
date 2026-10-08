import type { ComponentType, ReactNode } from "react";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useRenderPause } from "@/stores/renderPauseStore.ts";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import { useBackDismiss } from "@/hooks/useBackDismiss.ts";
import { useVisualViewport } from "@/hooks/useVisualViewport.ts";
import CloseButton from "../ui/buttons/CloseButton.tsx";
import GameButton from "../ui/buttons/GameButton.tsx";

export interface MobileScreenTab {
  id: string;
  label: string;
  badge?: number | boolean;
}

interface MobileScreenProps {
  title: string;
  icon: ComponentType;
  tabs?: MobileScreenTab[];
  activeTab?: string;
  onTabChange?: (tabId: string) => void;
  onClose: () => void;
  footer?: ReactNode;
  keyboardAware?: boolean;
  children: ReactNode;
}

function TabBadge({ badge }: { badge: number | boolean }) {
  if (badge === false || badge === 0) {
    return null;
  }
  if (badge === true) {
    return <span className="w-1.5 h-1.5 rounded-full bg-amber-300" aria-hidden="true" />;
  }
  return (
    <span className="min-w-[18px] px-1 rounded-full bg-white/15 text-[11px] leading-[18px] text-center">
      {badge}
    </span>
  );
}

export default function MobileScreen({
  title,
  icon: Icon,
  tabs,
  activeTab,
  onTabChange,
  onClose,
  footer,
  keyboardAware = false,
  children,
}: MobileScreenProps) {
  const reducedMotion = useReducedMotion();
  const viewport = useVisualViewport(keyboardAware);
  useRenderPause("mobile-screen", true);
  useBackDismiss(onClose);

  const animationClass = reducedMotion ? "" : "animate-[modalSlideIn_180ms_ease-out]";

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={title}
      className={`fixed inset-x-0 ${keyboardAware ? "" : "top-0 h-dvh"} flex flex-col text-white bg-[rgba(3,3,4,0.96)] ${animationClass}`}
      style={{
        zIndex: Z_INDEX.MOBILE_SCREEN,
        top: keyboardAware ? viewport.offsetTop : undefined,
        height: keyboardAware ? viewport.height : undefined,
        paddingTop: "var(--safe-top)",
        paddingLeft: "var(--safe-left)",
        paddingRight: "var(--safe-right)",
        paddingBottom: "var(--safe-bottom)",
      }}
    >
      <div className="flex items-center gap-3 h-11 shrink-0 pl-4 border-b border-white/15 bg-black/40">
        <span className="shrink-0 flex items-center text-white/90" aria-hidden="true">
          <Icon />
        </span>
        <div
          role="tablist"
          className="flex-1 min-w-0 flex items-center gap-1.5 overflow-x-auto overscroll-contain [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
        >
          {tabs?.map((tab) => {
            const selected = tab.id === activeTab;
            return (
              <GameButton
                key={tab.id}
                role="tab"
                aria-selected={selected}
                emphasis="quiet"
                size="xs"
                className={`shrink-0 !min-h-11 h-11 !px-3 whitespace-nowrap font-orbitron !text-[11px] tracking-wider uppercase border-b-2 ${selected ? "border-[#8fb8ca] !text-white" : "border-transparent"}`}
                onClick={() => onTabChange?.(tab.id)}
              >
                <span className="flex items-center gap-1.5">
                  {tab.label}
                  {tab.badge !== undefined && <TabBadge badge={tab.badge} />}
                </span>
              </GameButton>
            );
          })}
        </div>
        <CloseButton onClick={onClose} label={`Close ${title}`} />
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto overscroll-contain">{children}</div>
      {footer && (
        <div className="shrink-0 flex items-center justify-end gap-3 px-4 py-2 border-t border-white/15 bg-black/40">
          {footer}
        </div>
      )}
    </div>
  );
}
