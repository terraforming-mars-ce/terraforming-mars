import React from "react";

interface CorporationArtwork {
  name: string;
  file: string;
  width: number;
  height: number;
  viewBox: string;
}

const corporationLogos: Record<string, CorporationArtwork> = {
  credicor: {
    name: "CrediCor",
    file: "credicor.webp",
    width: 768,
    height: 768,
    viewBox: "79 124 612 505",
  },
  ecoline: {
    name: "Ecoline",
    file: "ecoline.webp",
    width: 768,
    height: 384,
    viewBox: "29 101 715 159",
  },
  helion: {
    name: "Helion",
    file: "helion.webp",
    width: 768,
    height: 480,
    viewBox: "4 103 762 250",
  },
  "interplanetary-cinematics": {
    name: "Interplanetary Cinematics",
    file: "interplanetary-cinematics.webp",
    width: 768,
    height: 432,
    viewBox: "25 105 718 221",
  },
  inventrix: {
    name: "Inventrix",
    file: "inventrix.webp",
    width: 768,
    height: 512,
    viewBox: "27 162 717 183",
  },
  "mining-guild": {
    name: "Mining Guild",
    file: "mining-guild.webp",
    width: 768,
    height: 465,
    viewBox: "41 45 687 372",
  },
  phobolog: {
    name: "PhoboLog",
    file: "phobolog.webp",
    width: 768,
    height: 512,
    viewBox: "38 73 693 355",
  },
  "tharsis-republic": {
    name: "Tharsis Republic",
    file: "tharsis-republic.webp",
    width: 768,
    height: 404,
    viewBox: "52 52 663 302",
  },
  thorgate: {
    name: "ThorGate",
    file: "thorgate.webp",
    width: 768,
    height: 384,
    viewBox: "24 108 719 151",
  },
  "united-nations-mars-initiative": {
    name: "United Nations Mars Initiative",
    file: "unmi.webp",
    width: 768,
    height: 452,
    viewBox: "20 37 728 378",
  },
  "saturn-systems": {
    name: "Saturn Systems",
    file: "saturn-systems.webp",
    width: 768,
    height: 512,
    viewBox: "40 89 689 300",
  },
  teractor: {
    name: "Teractor",
    file: "teractor.webp",
    width: 768,
    height: 384,
    viewBox: "9 126 750 124",
  },
  aridor: {
    name: "Aridor",
    file: "aridor.webp",
    width: 768,
    height: 439,
    viewBox: "45 28 664 382",
  },
  arklight: {
    name: "Arklight",
    file: "arklight.webp",
    width: 768,
    height: 530,
    viewBox: "85 46 598 462",
  },
  polyphemos: {
    name: "Polyphemos",
    file: "polyphemos.webp",
    width: 768,
    height: 512,
    viewBox: "35 103 699 306",
  },
  poseidon: {
    name: "Poseidon",
    file: "poseidon.webp",
    width: 768,
    height: 512,
    viewBox: "38 114 692 284",
  },
  "stormcraft-incorporated": {
    name: "Stormcraft Incorporated",
    file: "stormcraft-incorporated.webp",
    width: 768,
    height: 384,
    viewBox: "22 114 731 147",
  },
  "cheung-shing-mars": {
    name: "Cheung Shing Mars",
    file: "cheung-shing-mars.webp",
    width: 768,
    height: 640,
    viewBox: "48 93 674 445",
  },
  "point-luna": {
    name: "Point Luna",
    file: "point-luna.webp",
    width: 768,
    height: 349,
    viewBox: "23 49 720 252",
  },
  "robinson-industries": {
    name: "Robinson Industries",
    file: "robinson-industries.webp",
    width: 768,
    height: 512,
    viewBox: "25 157 719 171",
  },
  "valley-trust": {
    name: "Valley Trust",
    file: "valley-trust.webp",
    width: 768,
    height: 640,
    viewBox: "56 90 657 460",
  },
  vitor: { name: "Vitor", file: "vitor.webp", width: 768, height: 384, viewBox: "30 86 714 192" },
  "lakefront-resorts": {
    name: "Lakefront Resorts",
    file: "lakefront-resorts.webp",
    width: 768,
    height: 512,
    viewBox: "25 115 719 264",
  },
  pristar: {
    name: "Pristar",
    file: "pristar.webp",
    width: 768,
    height: 512,
    viewBox: "140 57 489 398",
  },
  "septem-tribus": {
    name: "Septem Tribus",
    file: "septem-tribus.webp",
    width: 768,
    height: 512,
    viewBox: "63 16 642 462",
  },
  "terralabs-research": {
    name: "TerraLabs Research",
    file: "terralabs-research.webp",
    width: 768,
    height: 512,
    viewBox: "51 58 669 374",
  },
  "utopia-invest": {
    name: "Utopia Invest",
    file: "utopia-invest.webp",
    width: 768,
    height: 512,
    viewBox: "36 108 695 282",
  },
  aphrodite: {
    name: "Aphrodite",
    file: "aphrodite.webp",
    width: 768,
    height: 432,
    viewBox: "21 113 726 205",
  },
  celestic: {
    name: "Celestic",
    file: "celestic.webp",
    width: 768,
    height: 512,
    viewBox: "32 105 704 286",
  },
  manutech: {
    name: "Manutech",
    file: "manutech.webp",
    width: 768,
    height: 432,
    viewBox: "33 124 702 176",
  },
  "morning-star-inc": {
    name: "Morning Star Inc.",
    file: "morning-star-inc.webp",
    width: 768,
    height: 512,
    viewBox: "38 126 696 262",
  },
  viron: { name: "Viron", file: "viron.webp", width: 768, height: 512, viewBox: "33 131 702 220" },
  "arcadian-communities": {
    name: "Arcadian Communities",
    file: "arcadian-communities.webp",
    width: 768,
    height: 512,
    viewBox: "34 73 699 363",
  },
  recyclon: {
    name: "Recyclon",
    file: "recyclon.webp",
    width: 768,
    height: 768,
    viewBox: "40 91 684 580",
  },
  "splice-tactical-genomics": {
    name: "Splice Tactical Genomics",
    file: "splice-tactical-genomics.webp",
    width: 768,
    height: 512,
    viewBox: "31 102 710 299",
  },
  factorum: {
    name: "Factorum",
    file: "factorum.webp",
    width: 768,
    height: 512,
    viewBox: "52 81 665 329",
  },
  "mons-insurance": {
    name: "Mons Insurance",
    file: "mons-insurance.webp",
    width: 768,
    height: 384,
    viewBox: "32 79 703 218",
  },
  philares: {
    name: "Philares",
    file: "philares.webp",
    width: 768,
    height: 512,
    viewBox: "28 143 715 225",
  },
  astrodrill: {
    name: "AstroDrill",
    file: "astrodrill.webp",
    width: 768,
    height: 512,
    viewBox: "27 154 719 181",
  },
  "pharmacy-union": {
    name: "Pharmacy Union",
    file: "pharmacy-union.webp",
    width: 768,
    height: 427,
    viewBox: "77 66 615 285",
  },
};

function CorporationLogo({ logo, className }: { logo: CorporationArtwork; className: string }) {
  const filterId = React.useId();

  return (
    <svg
      role="img"
      aria-label={logo.name}
      viewBox={logo.viewBox}
      className={`block max-w-full ${className}`}
    >
      <defs>
        <filter id={filterId} colorInterpolationFilters="sRGB">
          {/* Key out the charcoal matte while keeping the approved artwork’s colors. */}
          <feColorMatrix
            in="SourceGraphic"
            type="matrix"
            values="0 0 0 0 0  0 0 0 0 0  0 0 0 0 0  3 3 3 0 -0.945"
          />
          <feComposite in="SourceGraphic" operator="in" />
        </filter>
      </defs>
      <image
        href={`/assets/corporations/${logo.file}`}
        width={logo.width}
        height={logo.height}
        preserveAspectRatio="xMidYMid meet"
        filter={`url(#${filterId})`}
      />
    </svg>
  );
}

export function getCorporationLogo(
  corporationName: string,
  className = "w-[220px] h-[110px]",
): React.ReactNode | null {
  const key = corporationName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  const logo = corporationLogos[key];
  if (!logo) {
    return null;
  }

  return <CorporationLogo logo={logo} className={className} />;
}
