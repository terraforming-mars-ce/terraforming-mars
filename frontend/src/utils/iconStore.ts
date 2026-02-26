/**
 * Centralized icon asset path lookup store.
 * All icon path mappings are defined here to avoid duplication across components.
 */

export const RESOURCE_ICONS: { [key: string]: string } = {
  credit: "/assets/resources/megacredit.png",
  steel: "/assets/resources/steel.png",
  titanium: "/assets/resources/titanium.png",
  plant: "/assets/resources/plant.png",
  energy: "/assets/resources/power.png",
  power: "/assets/resources/power.png",
  heat: "/assets/resources/heat.png",
  microbe: "/assets/resources/microbe.png",
  animal: "/assets/resources/animal.png",
  floater: "/assets/resources/floater.png",
  science: "/assets/resources/science.png",
  asteroid: "/assets/resources/asteroid.png",
  disease: "/assets/resources/disease.png",
  tr: "/assets/resources/tr.png",
  "terraform-rating": "/assets/resources/tr.png",
  fighter: "/assets/resources/fighter.png",
  camp: "/assets/resources/camp.png",
  preservation: "/assets/resources/preservation.png",
  data: "/assets/resources/data.png",
  specialized: "/assets/resources/specialized-robot.png",
  "specialized-robot": "/assets/resources/specialized-robot.png",
  delegate: "/assets/resources/director.png",
  director: "/assets/resources/director.png",
  influence: "/assets/misc/influence.png",
  // Production variants
  "credit-production": "/assets/resources/megacredit.png",
  "steel-production": "/assets/resources/steel.png",
  "titanium-production": "/assets/resources/titanium.png",
  "plant-production": "/assets/resources/plant.png",
  "energy-production": "/assets/resources/power.png",
  "heat-production": "/assets/resources/heat.png",
};

export const TAG_ICONS: { [key: string]: string } = {
  earth: "/assets/tags/earth.png",
  science: "/assets/tags/science.png",
  plant: "/assets/tags/plant.png",
  microbe: "/assets/tags/microbe.png",
  animal: "/assets/tags/animal.png",
  power: "/assets/tags/power.png",
  space: "/assets/tags/space.png",
  building: "/assets/tags/building.png",
  city: "/assets/tags/city.png",
  jovian: "/assets/tags/jovian.png",
  venus: "/assets/tags/venus.png",
  event: "/assets/tags/event.png",
  "mars-tag": "/assets/tags/mars.png",
  moon: "/assets/tags/moon.png",
  wild: "/assets/tags/wild.png",
  wildlife: "/assets/tags/wild.png", // Use wild.png as fallback
  clone: "/assets/tags/clone.png",
  crime: "/assets/tags/crime.png",
};

export const TILE_ICONS: { [key: string]: string } = {
  "city-placement": "/assets/tiles/city.png",
  "ocean-placement": "/assets/tiles/ocean.png",
  "greenery-placement": "/assets/tiles/greenery_no_O2.png",
  "land-claim": "/assets/tiles/special.png",
  "city-tile": "/assets/tiles/city.png",
  "ocean-tile": "/assets/tiles/ocean.png",
  "greenery-tile": "/assets/tiles/greenery_no_O2.png",
  "greenery-placed": "/assets/tiles/greenery_no_O2.png", // For triggered effects
  "ocean-placed": "/assets/tiles/ocean.png", // For triggered effects
  "city-placed": "/assets/tiles/city.png", // For triggered effects
  "volcano-tile": "/assets/tiles/hazard.png",
  "volcano-placement": "/assets/tiles/hazard.png",
};

export const GLOBAL_PARAMETER_ICONS: { [key: string]: string } = {
  temperature: "/assets/global-parameters/temperature.png",
  oxygen: "/assets/global-parameters/oxygen.png",
  ocean: "/assets/tiles/ocean.png",
  venus: "/assets/global-parameters/venus.png",
};

export const SPECIAL_ICONS: { [key: string]: string } = {
  "card-draw": "/assets/resources/card.png",
  "card-take": "/assets/resources/card.png",
  "card-peek": "/assets/resources/card.png",
  card: "/assets/misc/corpCard.png",
  tag: "/assets/tags/wild.png",
  discount: "/assets/resources/megacredit.png",
  milestone: "/assets/misc/checkmark.png",
  award: "/assets/misc/first-player.png",
  asterisk: "/assets/misc/asterisc.png",
  asterisc: "/assets/misc/asterisc.png", // Support both spellings
  mars: "/assets/mars.png",
  arrow: "/assets/misc/arrow.png",
};

/**
 * Get the icon path for a given icon type.
 * Searches across all icon categories.
 * Resources are checked first since card behaviors use resource types,
 * and tags are checked after for tag displays.
 */
export function getIconPath(iconType: string): string | null {
  const cleanType = iconType?.toLowerCase().replace(/[_\s]/g, "-");

  // Check resources first (used in card behaviors for inputs/outputs)
  if (RESOURCE_ICONS[cleanType]) {
    return RESOURCE_ICONS[cleanType];
  }

  // Check tags (for tag displays on cards)
  if (TAG_ICONS[cleanType]) {
    return TAG_ICONS[cleanType];
  }

  // Check tiles
  if (TILE_ICONS[cleanType]) {
    return TILE_ICONS[cleanType];
  }

  // Check global parameters
  if (GLOBAL_PARAMETER_ICONS[cleanType]) {
    return GLOBAL_PARAMETER_ICONS[cleanType];
  }

  // Check special icons
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
