import { ColonyOutputDto } from "@/types/generated/api-types.ts";

export interface PlayerInfo {
  id: string;
  name: string;
  color: string;
}

export const STORAGE_RESOURCE_TYPES = ["floater", "microbe", "animal"];

export function hasStorageCardForType(
  resourceType: string,
  playedCards: { resourceStorage?: { type: string } }[],
  corporation?: { resourceStorage?: { type: string } } | null,
): boolean {
  if (playedCards.some((c) => c.resourceStorage?.type === resourceType)) {
    return true;
  }
  if (corporation?.resourceStorage?.type === resourceType) {
    return true;
  }
  return false;
}

export function getStorageWarning(
  outputs: ColonyOutputDto[],
  playedCards: { resourceStorage?: { type: string } }[],
  corporation?: { resourceStorage?: { type: string } } | null,
): string | null {
  for (const output of outputs) {
    if (
      STORAGE_RESOURCE_TYPES.includes(output.type) &&
      output.amount > 0 &&
      !hasStorageCardForType(output.type, playedCards, corporation)
    ) {
      return `You have no cards that can store ${output.type}. Resources will be lost.`;
    }
  }
  return null;
}
