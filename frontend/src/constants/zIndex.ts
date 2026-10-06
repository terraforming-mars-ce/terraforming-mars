/**
 * Centralized z-index values for consistent layering across the application
 *
 * Organized in logical groups with clear separation between layers:
 * - Base layer (0-9): Game board, background elements
 * - UI layer (10-99): Basic UI components, overlays
 * - Navigation layer (100-199): Menu bars, sidebars
 * - Overlay layer (200-999): Tooltips, dropdowns, cards
 * - Modal layer (1000-8999): Standard modals, popups
 * - App layer (35000+): App-wide views, confirmations and critical overlays
 */

export const Z_INDEX = {
  CONTROL_DECORATION: -1,
  // Base Layer (0-9)
  GAME_BOARD_BACKGROUND: 0,
  GAME_BOARD_BASE: 1,
  GAME_BOARD_TILES: 2,
  GAME_BOARD_EFFECTS: 5,
  GAME_BOARD_OVERLAY: 10,

  // UI Layer (10-99)
  UI_BASE: 10,
  COST_DISPLAY: 20,
  TILE_PLACEMENT_PROMPT: 50,
  PLAYER_OVERLAY: 90,

  // Navigation Layer (100-199)
  // Pre-game corporation showcase: above the board, below selection popovers and the hamburger
  SHOWCASE: 105,
  TOP_MENU_BAR: 100,
  BOTTOM_RESOURCE_BAR: 100,
  LEFT_SIDEBAR: 110,
  RIGHT_SIDEBAR: 110,

  // Overlay Layer (200-999)
  CARDS_HAND_OVERLAY: 200,
  CARDS_PREVIEW: 250,
  TOOLTIPS: 300,
  DROPDOWNS: 400,
  CARD_HOVER: 500,
  CARD_SELECTED: 600,

  // Decorative scroll hint above the per-card index band, below highlighted(2000)/dragged(3000)
  CARD_FAN_SCROLL_HINT: 1500,
  CARD_FAN_HIGHLIGHTED: 2000,
  CARD_FAN_DRAGGED: 3000,

  // Modal Layer (1000-8999)
  MENU_DROPDOWN: 1000,
  STANDARD_MODAL: 2000,
  CARD_DETAIL_MODAL: 3000,
  CORPORATION_SELECTION: 5000,
  PENDING_ACTION_BACKDROP: 4500,

  // Popover Layer (10000+)
  POPOVER: 10001,
  SELECTION_BACKDROP: 10002,
  SELECTION_POPOVER: 10003,
  LOBBY_SETTINGS_MODAL: 10500,
  IMMEDIATE_BACKDROP: 30000,
  IMMEDIATE_POPOVER: 30001,

  // Loading overlay - above everything
  LOADING_OVERLAY: 99999,

  // Debug Windows Layer (20000-20099)
  DEBUG_WINDOWS: 20000,

  // Always-on-top UI (above debug windows)
  TOP_MENU_ALWAYS_ON_TOP: 20100,
  EXPANDED_CARD_FAN: 20201,

  DEBUG_OVERLAY: 9999,

  // App-wide layers stay above all gameplay selections, including immediate flows.
  APP_OVERLAY: 35000,
  CONFIRMATION_MODAL: 40000,
  SYSTEM_NOTIFICATIONS: 41000,
  ERROR_OVERLAYS: 42000,
} as const;

// Type for z-index values
export type ZIndexValue = (typeof Z_INDEX)[keyof typeof Z_INDEX];

// Helper function to get z-index with optional offset
export const getZIndex = (level: keyof typeof Z_INDEX, offset: number = 0): number => {
  return Z_INDEX[level] + offset;
};
