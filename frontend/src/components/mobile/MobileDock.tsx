import type { ReactNode } from "react";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { canPerformActions } from "@/utils/actionUtils.ts";
import { useMobileUiStore, type MobileScreenId } from "@/stores/mobileUiStore.ts";
import {
  ActionsIcon,
  HandIcon,
  LogIcon,
  PlayersIcon,
  ProjectsIcon,
  TableauIcon,
} from "./screenIcons.tsx";
import { useMobileGame } from "./MobileGameContext.tsx";
import { getConversionAvailability } from "./conversionAvailability.ts";

interface DockEntry {
  id: MobileScreenId;
  label: string;
  icon: ReactNode;
  badge?: number | boolean;
}

function DockBadge({ badge }: { badge: number | boolean | undefined }) {
  if (badge === undefined || badge === false || badge === 0) {
    return null;
  }
  if (badge === true) {
    return (
      <span
        className="absolute top-1.5 right-[calc(50%-18px)] w-2 h-2 rounded-full bg-amber-300 shadow-[0_0_6px_rgba(252,211,77,0.8)]"
        aria-hidden="true"
      />
    );
  }
  return (
    <span className="absolute top-1 left-[calc(50%+6px)] min-w-[18px] h-[18px] px-1 rounded-full bg-[#2c5a8a] border border-[#8fb8ca]/60 font-orbitron text-[11px] font-bold leading-[16px] text-center text-white tabular-nums">
      {badge}
    </span>
  );
}

interface MobileDockProps {
  className?: string;
}

export default function MobileDock({ className = "" }: MobileDockProps) {
  const { gameState, currentPlayer, isSpectator, chatMessages } = useMobileGame();
  const activeScreen = useMobileUiStore((s) => s.activeScreen);
  const lastSeenChatCount = useMobileUiStore((s) => s.lastSeenChatCount);
  const open = useMobileUiStore((s) => s.open);

  const canAct = canPerformActions(gameState);
  const conversions = getConversionAvailability(gameState, currentPlayer);
  const availableCardActions = currentPlayer?.actions?.filter((a) => a.available).length ?? 0;
  const availableConversions = Number(conversions.plants) + Number(conversions.heat);
  const actionsBadge = canAct ? availableCardActions + availableConversions : 0;

  const viewer = gameState.currentPlayer;
  const projectsAvailable =
    canAct &&
    (!!viewer?.standardProjects?.some((p) => p.available) ||
      !!viewer?.milestones?.some((m) => m.available) ||
      !!viewer?.awards?.some((a) => a.available) ||
      !!gameState.colonies?.some((c) => c.tradeAvailable || c.buildAvailable) ||
      !!gameState.projectFunding?.some((p) => p.canBuySeat));

  const entries: DockEntry[] = [
    ...(isSpectator
      ? []
      : [
          {
            id: "hand" as const,
            label: "Hand",
            icon: <HandIcon />,
            badge: currentPlayer?.cards?.length ?? 0,
          },
          { id: "actions" as const, label: "Actions", icon: <ActionsIcon />, badge: actionsBadge },
        ]),
    { id: "projects", label: "Projects", icon: <ProjectsIcon />, badge: projectsAvailable },
    { id: "tableau", label: "Tableau", icon: <TableauIcon /> },
    { id: "players", label: "Players", icon: <PlayersIcon /> },
    { id: "log", label: "Log", icon: <LogIcon />, badge: chatMessages.length > lastSeenChatCount },
  ];

  return (
    <nav
      aria-label="Game dock"
      className={`fixed bottom-0 inset-x-0 flex items-stretch bg-[rgba(3,3,4,0.92)] border-t border-white/10 pointer-events-auto ${className}`}
      style={{
        zIndex: Z_INDEX.MOBILE_HUD,
        height: "calc(var(--hud-dock-h) + var(--safe-bottom))",
        paddingBottom: "var(--safe-bottom)",
        paddingLeft: "var(--safe-left)",
        paddingRight: "var(--safe-right)",
      }}
    >
      {entries.map((entry) => {
        const active = activeScreen === entry.id;
        return (
          <button
            key={entry.id}
            type="button"
            aria-pressed={active}
            className={`relative flex-1 min-w-0 min-h-11 flex flex-col items-center justify-center gap-0.5 cursor-pointer border-t-2 transition-colors duration-150 ${active ? "border-[#8fb8ca] text-white bg-white/10" : "border-transparent text-[#aeb6c8] active:bg-white/5"}`}
            onClick={() => open(entry.id)}
          >
            {entry.icon}
            <span className="font-orbitron text-[11px] font-semibold tracking-wider uppercase leading-none">
              {entry.label}
            </span>
            <DockBadge badge={entry.badge} />
          </button>
        );
      })}
    </nav>
  );
}
