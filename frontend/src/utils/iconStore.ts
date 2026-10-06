import { assetUrl } from "@/assets";
/**
 * Centralized icon asset path lookup store.
 * All icon path mappings are defined here to avoid duplication across components.
 */

export const RESOURCE_ICONS: { [key: string]: string } = {
  credit: assetUrl("icons/resources/megacredit", 128),
  steel: assetUrl("icons/resources/steel", 128),
  titanium: assetUrl("icons/resources/titanium", 128),
  plant: assetUrl("icons/resources/plant", 128),
  energy: assetUrl("icons/resources/power", 128),
  power: assetUrl("icons/resources/power", 128),
  heat: assetUrl("icons/resources/heat", 128),
  microbe: assetUrl("icons/resources/microbe", 128),
  animal: assetUrl("icons/resources/animal", 128),
  floater: assetUrl("icons/resources/floater", 128),
  science: assetUrl("icons/resources/science", 128),
  asteroid: assetUrl("icons/resources/asteroid", 128),
  disease: assetUrl("icons/resources/disease", 128),
  tr: assetUrl("icons/resources/tr", 128),
  "terraform-rating": assetUrl("icons/resources/tr", 128),
  fighter: assetUrl("icons/resources/fighter", 128),
  camp: assetUrl("icons/resources/camp", 128),
  preservation: assetUrl("icons/resources/preservation", 128),
  data: assetUrl("icons/resources/data", 128),
  specialized: assetUrl("icons/resources/specialized-robot", 128),
  "specialized-robot": assetUrl("icons/resources/specialized-robot", 128),
  delegate: assetUrl("icons/resources/director", 128),
  director: assetUrl("icons/resources/director", 128),
  influence: assetUrl("icons/ui/influence", 128),
  // Production variants
  "credit-production": assetUrl("icons/resources/megacredit", 128),
  "steel-production": assetUrl("icons/resources/steel", 128),
  "titanium-production": assetUrl("icons/resources/titanium", 128),
  "plant-production": assetUrl("icons/resources/plant", 128),
  "energy-production": assetUrl("icons/resources/power", 128),
  "heat-production": assetUrl("icons/resources/heat", 128),
};

export const TAG_ICONS: { [key: string]: string } = {
  earth: assetUrl("icons/tags/earth", 128),
  science: assetUrl("icons/tags/science", 128),
  plant: assetUrl("icons/tags/plant", 128),
  microbe: assetUrl("icons/tags/microbe", 128),
  animal: assetUrl("icons/tags/animal", 128),
  power: assetUrl("icons/tags/power", 128),
  space: assetUrl("icons/tags/space", 128),
  building: assetUrl("icons/tags/building", 128),
  city: assetUrl("icons/tags/city", 128),
  jovian: assetUrl("icons/tags/jovian", 128),
  venus: assetUrl("icons/tags/venus", 128),
  event: assetUrl("icons/tags/event", 128),
  "mars-tag": assetUrl("icons/tags/mars", 128),
  moon: assetUrl("icons/tags/moon", 128),
  wild: assetUrl("icons/tags/wild", 128),
  wildlife: assetUrl("icons/tags/wild", 128), // Use wild.png as fallback
  clone: assetUrl("icons/tags/clone", 128),
  crime: assetUrl("icons/tags/crime", 128),
};

export const TILE_ICONS: { [key: string]: string } = {
  "city-placement": assetUrl("icons/placements/city", 128),
  "ocean-placement": assetUrl("icons/placements/ocean", 128),
  "greenery-placement": assetUrl("icons/placements/greenery-no-oxygen", 128),
  "land-claim": assetUrl("icons/placements/special", 128),
  "tile-placement": assetUrl("icons/placements/special", 128),
  tile: assetUrl("icons/placements/special", 128),
  "city-tile": assetUrl("icons/placements/city", 128),
  "ocean-tile": assetUrl("icons/placements/ocean", 128),
  "greenery-tile": assetUrl("icons/placements/greenery-no-oxygen", 128),
  "greenery-placed": assetUrl("icons/placements/greenery-no-oxygen", 128), // For triggered effects
  "ocean-placed": assetUrl("icons/placements/ocean", 128), // For triggered effects
  "city-placed": assetUrl("icons/placements/city", 128), // For triggered effects
  "volcano-tile": assetUrl("icons/placements/volcano", 128),
  "volcano-placement": assetUrl("icons/placements/volcano", 128),
  "colony-tile-add": assetUrl("icons/placements/colony", 128),
  colony: assetUrl("icons/placements/colony", 128),
  "colony-count": assetUrl("icons/placements/colony", 128),
  "colony-bonus": assetUrl("icons/placements/colony", 128),
  greenery: assetUrl("icons/placements/greenery-no-oxygen", 128),
  trade: assetUrl("icons/placements/trade", 128),
  "trade-fleet": assetUrl("icons/placements/trade", 128),
  "colony-track-step": assetUrl("icons/placements/colony", 128),
};

export const GLOBAL_PARAMETER_ICONS: { [key: string]: string } = {
  temperature: assetUrl("icons/parameters/temperature", 128),
  oxygen: assetUrl("icons/parameters/oxygen", 128),
  ocean: assetUrl("icons/placements/ocean", 128),
  venus: assetUrl("icons/parameters/venus", 128),
};

export const SPECIAL_ICONS: { [key: string]: string } = {
  "card-draw": assetUrl("icons/resources/card", 128),
  "card-take": assetUrl("icons/resources/card", 128),
  "card-peek": assetUrl("icons/resources/card", 128),
  "card-discard": assetUrl("icons/resources/card", 128),
  "card-buy": assetUrl("icons/resources/card", 128),
  card: assetUrl("icons/resources/card", 128),
  tag: assetUrl("icons/tags/wild", 128),
  discount: assetUrl("icons/resources/megacredit", 128),
  milestone: assetUrl("icons/ui/checkmark", 128),
  award: assetUrl("icons/ui/first-player", 128),
  asterisk: assetUrl("icons/ui/asterisc", 128),
  asterisc: assetUrl("icons/ui/asterisc", 128), // Support both spellings
  mars: assetUrl("icons/ui/mars", 128),
  arrow: assetUrl("icons/ui/arrow", 128),
  "card-resource": assetUrl("icons/resources/wild", 128),
};

/**
 * Get the icon path for a given icon type.
 * Searches across all icon categories.
 * Resources are checked first since card behaviors use resource types,
 * and tags are checked after for tag displays.
 */
export function getIconPath(iconType: string): string | null {
  const cleanType = iconType?.toLowerCase().replace(/[_\s]/g, "-");

  if (RESOURCE_ICONS[cleanType]) {
    return RESOURCE_ICONS[cleanType];
  }

  if (TILE_ICONS[cleanType]) {
    return TILE_ICONS[cleanType];
  }

  if (GLOBAL_PARAMETER_ICONS[cleanType]) {
    return GLOBAL_PARAMETER_ICONS[cleanType];
  }

  if (TAG_ICONS[cleanType]) {
    return TAG_ICONS[cleanType];
  }

  if (SPECIAL_ICONS[cleanType]) {
    return SPECIAL_ICONS[cleanType];
  }

  // Handle production suffix
  if (cleanType.includes("-production")) {
    const baseResourceType = cleanType.replace("-production", "");
    if (RESOURCE_ICONS[baseResourceType]) {
      return RESOURCE_ICONS[baseResourceType];
    }
  }

  // Handle -tag suffix to force tag icon lookup (e.g., "plant-tag" -> tag icon for plant)
  if (cleanType.endsWith("-tag")) {
    const baseTagType = cleanType.replace("-tag", "");
    if (TAG_ICONS[baseTagType]) {
      return TAG_ICONS[baseTagType];
    }
  }

  return null;
}

/**
 * Check if an icon type is a tag.
 */
export function isTagIcon(iconType: string): boolean {
  const cleanType = iconType?.toLowerCase().replace(/[_\s]/g, "-");
  return TAG_ICONS[cleanType] !== undefined;
}

/**
 * Get the tag icon path specifically.
 * Use this when you know you need the tag icon (e.g., for card tag displays).
 */
export function getTagIconPath(iconType: string): string | null {
  const cleanType = iconType?.toLowerCase().replace(/[_\s]/g, "-");
  return TAG_ICONS[cleanType] || null;
}

/**
 * Check if an icon type is a tile.
 */
export function isTileIcon(iconType: string): boolean {
  const cleanType = iconType?.toLowerCase().replace(/[_\s]/g, "-");
  return TILE_ICONS[cleanType] !== undefined;
}

/**
 * Check if an icon type is a global parameter.
 */
export function isGlobalParameterIcon(iconType: string): boolean {
  const cleanType = iconType?.toLowerCase().replace(/[_\s]/g, "-");
  return GLOBAL_PARAMETER_ICONS[cleanType] !== undefined;
}
