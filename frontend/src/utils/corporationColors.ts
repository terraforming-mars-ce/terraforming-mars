// Accents from the approved corporation logos, shared by cards and the resource bar.

export const corporationBorderColors: Record<string, string> = {
  // Base game corporations
  credicor: "#bda4e6",
  ecoline: "#8ce329",
  helion: "#ffb500",
  "mining-guild": "#c79555",
  "interplanetary-cinematics": "#ef5c52",
  inventrix: "#00cbea",
  phobolog: "#9266ff",
  "tharsis-republic": "#c45e3f",
  thorgate: "#00aeef",
  "united-nations-mars-initiative": "#79b7ec",

  // Corporate Era corporations
  teractor: "#d6ed66",
  "saturn-systems": "#e76b4a",
  aphrodite: "#efb5b5",
  celestic: "#8faeff",
  manutech: "#ec8c43",
  "morning-star-inc": "#ffa91c",
  viron: "#4be1b2",

  // Prelude corporations
  "cheung-shing-mars": "#dd4e2d",
  "point-luna": "#43c8c6",
  "robinson-industries": "#edaa4d",
  "valley-trust": "#8fd2ab",
  vitor: "#cda950",

  // Colonies corporations
  aridor: "#f3ad18",
  arklight: "#8fdddd",
  polyphemos: "#984bff",
  poseidon: "#22c8d0",
  "stormcraft-incorporated": "#96c6e5",
  "lakefront-resorts": "#94cdd0",
  pristar: "#98b686",
  "septem-tribus": "#cc6848",
  "terralabs-research": "#2fc9ef",
  "utopia-invest": "#b59aef",
  factorum: "#cf7047",
  "mons-insurance": "#9fceee",
  philares: "#2fc4c9",
  "arcadian-communities": "#a3b988",
  recyclon: "#a9ed2c",
  "splice-tactical-genomics": "#cd5596",

  // Turmoil corporations
  astrodrill: "#ffb60b",
  "pharmacy-union": "#56d2bf",
  ecotec: "#00aa00", // green

  // Promo corporations
  "tycho-magnetics": "#808080", // grey
  "kuiper-cooperative": "#627edb", // blue
  spire: "#aaaaaa", // grey
  sagitta: "#52be52", // green
  "palladin-shipping": "#dada73", // yellow
  "nirgal-enterprises": "#e29911", // orange/gold
};

export function getCorporationBorderColor(corporationName: string): string {
  const key = corporationName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  return corporationBorderColors[key] || "#ffc107"; // default gold
}
