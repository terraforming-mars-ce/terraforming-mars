export const CARD_TYPE_COLORS = {
  automated: "#4caf50",
  active: "#2196f3",
  event: "#f44336",
  corporation: "#ffc107",
  prelude: "#e91e63",
} as const;

export const getCardTypeColor = (type: string): string => {
  return CARD_TYPE_COLORS[type.toLowerCase() as keyof typeof CARD_TYPE_COLORS] || "#4a90e2";
};
