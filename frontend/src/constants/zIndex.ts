/**
 * Centralized z-index values for consistent layering across the application
 *
 * Organized in logical groups with clear separation between layers:
 * - Base layer (0-9): Game board, background elements
 * - UI layer (10-99): Basic UI components, overlays
 * - Navigation layer (100-199): Menu bars, sidebars
 * - Overlay layer (200-999): Tooltips, dropdowns, cards
 * - Modal layer (1000-8999): Standard modals, popups
 * - Critical layer (9000-9999): System-critical overlays
 */

export const Z_INDEX = {
  // Base Layer (0-9)
  GAME_BOARD_BACKGROUND: 0,
  GAME_BOARD_BASE: 1,
  GAME_BOARD_TILES: 2,
  GAME_BOARD_EFFECTS: 5,
  GAME_BOARD_OVERLAY: 10,

  // UI Layer (10-99)
  UI_BASE: 10,
  COST_DISPLAY: 20,
  PLAYER_OVERLAY: 90,

  // Navigation Layer (100-199)
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

  // Modal Layer (1000-8999)
  MENU_DROPDOWN: 1000,
  STANDARD_MODAL: 2000,
  CARD_DETAIL_MODAL: 3000,
  CONFIRMATION_MODAL: 4000,
  CORPORATION_SELECTION: 5000,

  // Popover Layer (10000+)
  POPOVER: 10001,
  SELECTION_POPOVER: 10002,

  // Debug Windows Layer (20000-20099)
  DEBUG_WINDOWS: 20000,

  // Always-on-top UI (above debug windows)
  TOP_MENU_ALWAYS_ON_TOP: 20100,

  // Critical Layer (9000-9999)
  SYSTEM_NOTIFICATIONS: 9000,
  ERROR_OVERLAYS: 9500,
  DEBUG_OVERLAY: 9999,
} as const;

// Type for z-index values
export type ZIndexValue = (typeof Z_INDEX)[keyof typeof Z_INDEX];

// Helper function to get z-index with optional offset
export const getZIndex = (level: keyof typeof Z_INDEX, offset: number = 0): number => {
  return Z_INDEX[level] + offset;
};
