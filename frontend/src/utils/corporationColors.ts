// Corporation brand colors for card border styling
// These are derived from each corporation's logo/branding

export const corporationBorderColors: Record<string, string> = {
  // Base game corporations
  credicor: "#800080", // purple
  ecoline: "#00b400", // green
  helion: "#e6e600", // yellow
  "mining-guild": "#c9380e", // orange-red
  "interplanetary-cinematics": "#cc3333", // red
  inventrix: "#6bb5c7", // cyan
  phobolog: "#6b3fa0", // purple
  "tharsis-republic": "#ff5f00", // orange
  thorgate: "#808080", // grey
  "united-nations-mars-initiative": "#027dc3", // blue

  // Corporate Era corporations
  teractor: "#ff4500", // orangered
  "saturn-systems": "#6b3fa0", // purple
  aphrodite: "#ff8c00", // orange
  celestic: "#17b9ec", // cyan-blue
  manutech: "#e63900", // red-orange
  "morning-star-inc": "#ffffff", // white
  viron: "#808080", // grey

  // Prelude corporations
  "cheung-shing-mars": "#ff0000", // red
  "point-luna": "#ffffff", // white
  "robinson-industries": "#cccccc", // silver/grey
  "valley-trust": "#027dc3", // blue
  vitor: "#ff4500", // orangered

  // Colonies corporations
  aridor: "#cc3333", // red background
  arklight: "#000089", // deep blue
  polyphemos: "#cc3333", // red
  poseidon: "#4169e1", // royal blue
  "stormcraft-incorporated": "#ff8c00", // orange
  "lakefront-resorts": "#ffffff", // white
  pristar: "#ff5d21", // orange
  "septem-tribus": "#ffffff", // white
  "terralabs-research": "#ffffff", // white
  "utopia-invest": "#00aa00", // green
  factorum: "#ff8c00", // orange
  "mons-insurance": "#8b4513", // brown
  philares: "#ff5858", // red
  "arcadian-communities": "#eeeeee", // light grey
  recyclon: "#ff0000", // red
  "splice-tactical-genomics": "#ff0000", // red

  // Turmoil corporations
  astrodrill: "#ffcc00", // yellow/gold
  "pharmacy-union": "#ffffff", // white
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
