const LOCAL_BUILD = "localbuild";
const VERSION_PATTERN = /^v(\d+)(?:\.(\d+))?(?:\.(\d+))?/;

export function isLocalBuild(version: string): boolean {
  return version === LOCAL_BUILD;
}

// Untagged builds are "<git describe>_<UTC timestamp>"; the timestamp is noise to players
export function displayVersion(version: string): string {
  return version.split("_")[0];
}

function versionParts(version: string): [number, number, number] | null {
  const match = VERSION_PATTERN.exec(version);
  if (!match) {
    return null;
  }
  return [Number(match[1]), Number(match[2] ?? 0), Number(match[3] ?? 0)];
}

// Orders release tags numerically (v7 < v7.0.1 < v7.1.0); untagged builds compare as their base tag
export function compareVersions(a: string, b: string): number {
  const partsA = versionParts(a);
  const partsB = versionParts(b);
  if (!partsA || !partsB) {
    return 0;
  }
  for (let i = 0; i < 3; i++) {
    if (partsA[i] !== partsB[i]) {
      return partsA[i] < partsB[i] ? -1 : 1;
    }
  }
  return 0;
}
