import { create } from "zustand";

export type MobileScreenId = "hand" | "actions" | "projects" | "tableau" | "players" | "log";

export type MobileProjectsTab = "standard" | "milestones" | "awards" | "colonies" | "funding";

export type MobileLogTab = "log" | "chat";

export interface MobileScreenParams {
  projectsTab?: MobileProjectsTab;
  tableauPlayerId?: string | null;
  playerId?: string | null;
  logTab?: MobileLogTab;
}

interface MobileUiState {
  activeScreen: MobileScreenId | null;
  projectsTab: MobileProjectsTab;
  tableauPlayerId: string | null;
  focusedPlayerId: string | null;
  logTab: MobileLogTab;
  lastSeenChatCount: number;

  open: (screen: MobileScreenId, params?: MobileScreenParams) => void;
  close: () => void;
  setProjectsTab: (tab: MobileProjectsTab) => void;
  setLogTab: (tab: MobileLogTab) => void;
  markChatSeen: (count: number) => void;
}

export const useMobileUiStore = create<MobileUiState>((set) => ({
  activeScreen: null,
  projectsTab: "standard",
  tableauPlayerId: null,
  focusedPlayerId: null,
  logTab: "log",
  lastSeenChatCount: 0,

  open: (screen, params = {}) =>
    set((state) => ({
      activeScreen: screen,
      projectsTab: params.projectsTab ?? state.projectsTab,
      tableauPlayerId:
        screen === "tableau" ? (params.tableauPlayerId ?? null) : state.tableauPlayerId,
      focusedPlayerId: params.playerId ?? null,
      logTab: params.logTab ?? state.logTab,
    })),
  close: () => set({ activeScreen: null }),
  setProjectsTab: (tab) => set({ projectsTab: tab }),
  setLogTab: (tab) => set({ logTab: tab }),
  markChatSeen: (count) => set({ lastSeenChatCount: count }),
}));
