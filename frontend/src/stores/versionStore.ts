import { create } from "zustand";
import { APP_VERSION } from "@/config.ts";
import { isLocalBuild } from "@/utils/version.ts";

interface VersionState {
  serverVersion: string | null;
  dismissedVersion: string | null;
  changelogOpen: boolean;
  changelogFocus: string | null;
  setServerVersion: (serverVersion: string) => void;
  dismissUpdate: (version: string) => void;
  openChangelog: (focus?: string) => void;
  closeChangelog: () => void;
}

export const useVersionStore = create<VersionState>((set) => ({
  serverVersion: null,
  dismissedVersion: null,
  changelogOpen: false,
  changelogFocus: null,
  setServerVersion: (serverVersion) => set({ serverVersion }),
  dismissUpdate: (dismissedVersion) => set({ dismissedVersion }),
  openChangelog: (focus) => set({ changelogOpen: true, changelogFocus: focus ?? null }),
  closeChangelog: () => set({ changelogOpen: false, changelogFocus: null }),
}));

export function availableUpdate(state: VersionState): string | null {
  const { serverVersion } = state;
  if (serverVersion === null || serverVersion === APP_VERSION) {
    return null;
  }
  if (isLocalBuild(serverVersion) || isLocalBuild(APP_VERSION)) {
    return null;
  }
  return serverVersion;
}

export function showsUpdateNotice(state: VersionState): boolean {
  const update = availableUpdate(state);
  return update !== null && update !== state.dismissedVersion;
}
