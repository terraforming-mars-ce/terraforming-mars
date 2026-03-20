// Frontend card type definitions for Terraforming Mars
import {
  StandardProjectSellPatents,
  StandardProjectPowerPlant,
  StandardProjectAsteroid,
  StandardProjectAquifer,
  StandardProjectGreenery,
  StandardProjectCity,
} from "./generated/api-types.ts";

export enum CardType {
  AUTOMATED = "automated",
  ACTIVE = "active",
  EVENT = "event",
  CORPORATION = "corporation",
  PRELUDE = "prelude",
}

export enum CardTag {
  BUILDING = "building",
  SPACE = "space",
  POWER = "power",
  SCIENCE = "science",
  MICROBE = "microbe",
  ANIMAL = "animal",
  PLANT = "plant",
  EARTH = "earth",
  JOVIAN = "jovian",
  CITY = "city",
  VENUS = "venus",
  MARS = "mars",
  MOON = "moon",
  WILD = "wild",
  EVENT = "event",
  CLONE = "clone",
}

// Standard Project type - re-export generated constants for convenience
export type StandardProject = string;
export const StandardProject = {
  SELL_PATENTS: StandardProjectSellPatents,
  POWER_PLANT: StandardProjectPowerPlant,
  ASTEROID: StandardProjectAsteroid,
  AQUIFER: StandardProjectAquifer,
  GREENERY: StandardProjectGreenery,
  CITY: StandardProjectCity,
} as const;
