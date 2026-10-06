import { ResourceType } from "@/types/generated/api-types.ts";

export const RESOURCE_COLORS: Record<ResourceType, string> = {
  credit: "#f1c40f", // Gold
  steel: "#d2691e", // Brown/orangy
  titanium: "#95a5a6", // Grey
  plant: "#27ae60", // Green
  energy: "#9b59b6", // Purple
  heat: "#ff4500", // Red/orange
};

export const RESOURCE_NAMES: Record<ResourceType, string> = {
  credit: "Credits",
  steel: "Steel",
  titanium: "Titanium",
  plant: "Plants",
  energy: "Energy",
  heat: "Heat",
  floater: "Floaters",
  microbe: "Microbes",
  animal: "Animals",
  asteroid: "Asteroids",
  fighter: "Fighters",
  science: "Science resources",
  "card-resource": "Resources",
};

export const getResourceColor = (resourceType: ResourceType): string => {
  return RESOURCE_COLORS[resourceType];
};

const SINGULAR_RESOURCE_NAMES: Record<ResourceType, string> = {
  credit: "Credit",
  plant: "Plant",
  floater: "Floater",
  microbe: "Microbe",
  animal: "Animal",
  asteroid: "Asteroid",
  fighter: "Fighter",
  science: "Science resource",
  "card-resource": "Resource",
};

export const getResourceName = (resourceType: ResourceType, amount?: number): string => {
  if (resourceType.endsWith("-production")) {
    return `${getResourceName(resourceType.replace("-production", ""), 1)} production`;
  }
  if (amount === 1 && SINGULAR_RESOURCE_NAMES[resourceType]) {
    return SINGULAR_RESOURCE_NAMES[resourceType];
  }
  return RESOURCE_NAMES[resourceType] ?? resourceType.replaceAll("-", " ");
};
