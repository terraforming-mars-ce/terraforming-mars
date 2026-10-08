import type { ComponentType, ReactNode } from "react";
import type { MobileScreenId } from "@/stores/mobileUiStore.ts";

function LineIcon({ children }: { children: ReactNode }) {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {children}
    </svg>
  );
}

export const ActionsIcon = () => (
  <LineIcon>
    <circle cx="3" cy="12" r="1.1" fill="currentColor" stroke="none" />
    <circle cx="8" cy="12" r="1.1" fill="currentColor" stroke="none" />
    <path d="M12 12h10m-4-4 4 4-4 4" />
  </LineIcon>
);

export const ProjectsIcon = () => (
  <LineIcon>
    <path d="M4 21V9l8-5 8 5v12" />
    <path d="M9 21v-6h6v6" />
  </LineIcon>
);

export const TableauIcon = () => (
  <LineIcon>
    <rect x="3" y="4" width="7" height="7" />
    <rect x="14" y="4" width="7" height="7" />
    <rect x="3" y="14" width="7" height="7" />
    <rect x="14" y="14" width="7" height="7" />
  </LineIcon>
);

export const PlayersIcon = () => (
  <LineIcon>
    <circle cx="9" cy="8" r="3.5" />
    <path d="M2.5 20c.8-3.5 3.4-5.5 6.5-5.5s5.7 2 6.5 5.5" />
    <path d="M16 4.8a3.5 3.5 0 0 1 0 6.4M18 14.8c1.8.8 3 2.6 3.5 5.2" />
  </LineIcon>
);

export const LogIcon = () => (
  <LineIcon>
    <path d="M4 5h16v11H9l-5 4z" />
    <path d="M8 9h8M8 12h5" />
  </LineIcon>
);

export const HandIcon = () => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 14 14"
    fill="none"
    stroke="currentColor"
    strokeWidth="0.85"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden="true"
  >
    <path d="M6.546.857a.475.475 0 0 1 .581-.335l6.02 1.612a.475.475 0 0 1 .337.581l-2.31 8.618a.475.475 0 0 1-.582.335l-6.02-1.612a.475.475 0 0 1-.336-.581z" />
    <path d="M6.108 2.535L.852 3.944a.475.475 0 0 0-.336.581l2.308 8.618a.475.475 0 0 0 .582.335l3.01-.806" />
  </svg>
);

export const SCREEN_ICONS: Record<MobileScreenId, ComponentType> = {
  hand: HandIcon,
  actions: ActionsIcon,
  projects: ProjectsIcon,
  tableau: TableauIcon,
  players: PlayersIcon,
  log: LogIcon,
};
