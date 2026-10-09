/**
 * TabManager handles multi-tab prevention for the Terraforming Mars game.
 * Ensures only one tab can have an active game session at a time.
 */

import { serverScopedKey } from "./gateway.ts";

const TAB_STORAGE_KEY_PREFIX = serverScopedKey("terraforming-mars-active-tab");
const HEARTBEAT_INTERVAL = 5000; // 5 seconds
const TAB_TIMEOUT = 10000; // 10 seconds

interface TabState {
  tabId: string;
  timestamp: number;
  gameId: string | null;
  playerName: string | null;
}

export class TabManager {
  private tabId: string;
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private isActive = false;
  private gameId: string | null = null;
  private playerName: string | null = null;

  constructor() {
    this.tabId = this.generateTabId();
    this.setupEventListeners();
  }

  /**
   * Attempts to claim the active tab status for a game session
   */
  claimTab(gameId: string, playerName: string): Promise<boolean> {
    return new Promise((resolve) => {
      this.gameId = gameId;
      this.playerName = playerName;

      const existingTab = this.getActiveTab();

      if (!existingTab || this.isTabExpired(existingTab)) {
        this.setActiveTab();
        this.startHeartbeat();
        this.isActive = true;
        resolve(true);
        return;
      }

      if (existingTab.tabId === this.tabId) {
        this.setActiveTab();
        this.startHeartbeat();
        this.isActive = true;
        resolve(true);
        return;
      }

      resolve(false);
    });
  }

  /**
   * Releases the active tab status
   */
  releaseTab(): void {
    this.stopHeartbeat();
    this.removeActiveTab();
    this.isActive = false;
    this.gameId = null;
    this.playerName = null;
  }

  /**
   * Checks if this tab is currently the active tab
   */
  isActiveTab(): boolean {
    if (!this.isActive) {
      return false;
    }

    const activeTab = this.getActiveTab();
    return activeTab?.tabId === this.tabId && !this.isTabExpired(activeTab);
  }

  /**
   * Gets information about the currently active tab
   */
  getActiveTabInfo(): { gameId: string; playerName: string } | null {
    const activeTab = this.getActiveTab();
    if (!activeTab || this.isTabExpired(activeTab)) {
      return null;
    }

    return {
      gameId: activeTab.gameId || "unknown",
      playerName: activeTab.playerName || "unknown",
    };
  }

  /**
   * Force takes over the tab (used when user confirms they want to take over)
   */
  forceTakeOver(gameId: string, playerName: string): void {
    this.gameId = gameId;
    this.playerName = playerName;
    this.setActiveTab();
    this.startHeartbeat();
    this.isActive = true;
  }

  private generateTabId(): string {
    return `tab-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private getActiveTab(): TabState | null {
    if (!this.playerName) return null;

    try {
      const key = `${TAB_STORAGE_KEY_PREFIX}-${this.playerName}`;
      const stored = localStorage.getItem(key);
      return stored ? JSON.parse(stored) : null;
    } catch {
      return null;
    }
  }

  private setActiveTab(): void {
    if (!this.playerName) return;

    const tabState: TabState = {
      tabId: this.tabId,
      timestamp: Date.now(),
      gameId: this.gameId,
      playerName: this.playerName,
    };

    const key = `${TAB_STORAGE_KEY_PREFIX}-${this.playerName}`;
    localStorage.setItem(key, JSON.stringify(tabState));
  }

  private removeActiveTab(): void {
    if (!this.playerName) return;

    const activeTab = this.getActiveTab();
    if (activeTab?.tabId === this.tabId) {
      const key = `${TAB_STORAGE_KEY_PREFIX}-${this.playerName}`;
      localStorage.removeItem(key);
    }
  }

  private isTabExpired(tabState: TabState): boolean {
    return Date.now() - tabState.timestamp > TAB_TIMEOUT;
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();

    this.heartbeatInterval = setInterval(() => {
      if (this.isActive) {
        this.setActiveTab();
      } else {
        this.stopHeartbeat();
      }
    }, HEARTBEAT_INTERVAL);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private setupEventListeners(): void {
    window.addEventListener("beforeunload", () => {
      this.releaseTab();
    });

    document.addEventListener("visibilitychange", () => {
      if (document.hidden) {
        this.stopHeartbeat();
      } else if (this.isActive) {
        this.startHeartbeat();
      }
    });

    window.addEventListener("storage", (event) => {
      const expectedKey = this.playerName ? `${TAB_STORAGE_KEY_PREFIX}-${this.playerName}` : null;

      if (event.key === expectedKey && this.isActive) {
        const newTabState = event.newValue ? JSON.parse(event.newValue) : null;

        if (newTabState && newTabState.tabId !== this.tabId) {
          this.stopHeartbeat();
          this.isActive = false;
        }
      }
    });
  }

  /**
   * Clean up resources
   */
  destroy(): void {
    this.releaseTab();
  }
}

// Singleton instance
let tabManagerInstance: TabManager | null = null;

export const getTabManager = (): TabManager => {
  if (!tabManagerInstance) {
    tabManagerInstance = new TabManager();
  }
  return tabManagerInstance;
};
