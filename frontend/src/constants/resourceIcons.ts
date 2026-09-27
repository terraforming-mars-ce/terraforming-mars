/**
 * Resource and game element icon paths
 */

export const RESOURCE_ICONS = {
  credit: "/assets/resources/megacredit.png",
  steel: "/assets/resources/steel.png",
  titanium: "/assets/resources/titanium.png",
  plant: "/assets/resources/plant.png",
  energy: "/assets/resources/power.png",
  heat: "/assets/resources/heat.png",
  oxygen: "/assets/global-parameters/oxygen.png",
  ocean: "/assets/tiles/ocean.png",
  tr: "/assets/resources/tr.png",
} as const;

export const GLOBAL_PARAM_ICONS = {
  temperature: "/assets/global-parameters/temperature.png",
  oxygen: "/assets/global-parameters/oxygen.png",
  ocean: "/assets/tiles/ocean.png",
} as const;

export const MISC_ICONS = {
  production: "/assets/misc/production.png",
} as const;

export type ResourceType = keyof typeof RESOURCE_ICONS;
export type GlobalParamType = keyof typeof GLOBAL_PARAM_ICONS;
