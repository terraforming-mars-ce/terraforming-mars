import { IconDisplayInfo } from "../types.ts";
import {
  type ResourceCondition,
  isProduction as isProductionType,
  getPer,
  getVariableAmount,
} from "@/types/resourceConditions.ts";

/**
 * Enhanced resource display analysis that considers space constraints.
 * Determines whether to display resources individually or with a number indicator.
 *
 * @param resource - Resource object with type and amount
 * @param availableSpace - Maximum number of icons that can fit horizontally
 * @param forceCompact - Whether to force compact display mode
 * @returns Display information including mode and icon count
 */
export const analyzeResourceDisplayWithConstraints = (
  resource: ResourceCondition & { forceNumberFormat?: boolean },
  availableSpace: number,
  forceCompact: boolean = false,
): IconDisplayInfo => {
  const resourceType = resource.type || "unknown";
  const amount = resource.amount ?? 1;
  const hasPer = getPer(resource);
  const isProduction = isProductionType(resource);
  const variableAmount = getVariableAmount(resource);

  if (variableAmount) {
    return {
      resourceType,
      amount,
      displayMode: "number",
      iconCount: 1,
      variableAmount: true,
    };
  }

  // Include every tag in combined per conditions when budgeting layout space.
  if (isProduction && hasPer) {
    return {
      resourceType,
      amount,
      displayMode: "number", // Always use number format for per conditions
      iconCount: 1 + Math.max(1, hasPer.tags?.length ?? 0),
    };
  }

  // Use individual display for amounts ≤3 (unless compact mode forces earlier threshold)
  const maxIndividualIcons = forceCompact || resource.forceNumberFormat ? 1 : 3;
  const absoluteAmount = Math.abs(amount);
  const useIndividual =
    absoluteAmount > 0 && absoluteAmount <= maxIndividualIcons && absoluteAmount <= availableSpace;

  return {
    resourceType,
    amount,
    displayMode: useIndividual ? "individual" : "number",
    iconCount: useIndividual ? absoluteAmount : 1,
  };
};

export const coordinateDisplayModes = (
  resources: ResourceCondition[],
  compact = false,
  paired = false,
): Map<ResourceCondition, IconDisplayInfo> => {
  const infos = resources.map((resource) => ({
    resource,
    info: analyzeResourceDisplayWithConstraints(resource, Infinity, compact),
  }));
  const numeric =
    paired &&
    infos.some(
      ({ resource, info }) => !resource.type.startsWith("credit") && info.displayMode === "number",
    );
  return new Map(
    infos.map(({ resource, info }) => [
      resource,
      numeric ? { ...info, displayMode: "number", iconCount: 1 } : info,
    ]),
  );
};

export function orderIndependentResources(resources: ResourceCondition[], compact = false) {
  const hasExternalNumber = (resource: ResourceCondition) => {
    if (resource.type.startsWith("credit")) {
      return false;
    }
    return (
      resource.type === "trade-fleet" ||
      resource.type === "colony-track-step" ||
      analyzeResourceDisplayWithConstraints(resource, Infinity, compact).displayMode === "number"
    );
  };
  return [
    ...resources.filter(hasExternalNumber),
    ...resources.filter((resource) => !hasExternalNumber(resource)),
  ];
}

/**
 * Analyzes and consolidates card outputs for optimal display.
 * Implements smart consolidation rules:
 * - If peek == buy, show only buy icon
 * - If peek == take, show only take icon
 * - Otherwise, show both separately
 * - Pure card-draw shows no badge
 *
 * @param outputs - Array of output resources
 * @returns Array of consolidated card display items
 */
export interface CardDisplayItem {
  amount: number;
  badgeType: "peek" | "take" | "buy" | "discard" | "none";
  isAttack: boolean;
  keepAmount?: number;
}

const isAttackTarget = (target: string | undefined): boolean =>
  target === "any-player" || target === "all-opponents" || (target?.startsWith("steal-") ?? false);

export const analyzeCardOutputs = (outputs: ResourceCondition[]): CardDisplayItem[] => {
  // Separate card outputs by target (self vs opponents)
  let selfDraw = 0;
  let selfPeek = 0;
  let selfTake = 0;
  let selfBuy = 0;
  let selfDiscard = 0;
  let attackDraw = 0;
  let attackPeek = 0;
  let attackTake = 0;
  let attackBuy = 0;
  let attackDiscard = 0;

  outputs.forEach((output) => {
    const type = output.type;
    const amount = output.amount ?? 0;
    const attack = isAttackTarget(output.target);

    switch (type) {
      case "card-draw":
        if (attack) attackDraw += amount;
        else selfDraw += amount;
        break;
      case "card-peek":
        if (attack) attackPeek += amount;
        else selfPeek += amount;
        break;
      case "card-take":
        if (attack) attackTake += amount;
        else selfTake += amount;
        break;
      case "card-buy":
        if (attack) attackBuy += amount;
        else selfBuy += amount;
        break;
      case "card-discard":
        if (attack) attackDiscard += amount;
        else selfDiscard += amount;
        break;
    }
  });

  const hasSelf = selfDraw + selfPeek + selfTake + selfBuy + selfDiscard > 0;
  const hasAttack = attackDraw + attackPeek + attackTake + attackBuy + attackDiscard > 0;

  if (!hasSelf && !hasAttack) return [];

  const result: CardDisplayItem[] = [];

  const consolidateGroup = (
    draw: number,
    peek: number,
    take: number,
    buy: number,
    discard: number,
    isAttack: boolean,
  ) => {
    if (discard > 0) {
      result.push({ amount: discard, badgeType: "discard", isAttack });
    }

    // Pure card-draw (no peek/take/buy)
    if (draw > 0 && peek === 0 && take === 0 && buy === 0) {
      result.push({ amount: draw, badgeType: "none", isAttack });
      return;
    }

    // Peek + Buy with equal amounts -> consolidate to buy only
    if (peek > 0 && buy > 0 && peek === buy && take === 0) {
      result.push({ amount: buy, badgeType: "buy", isAttack });
      return;
    }

    if (peek > take && take > 0 && buy === 0 && draw === 0) {
      result.push({ amount: peek, badgeType: "peek", isAttack, keepAmount: take });
      return;
    }

    // Peek + Take with equal amounts -> consolidate to take only
    if (peek > 0 && take > 0 && peek === take && buy === 0) {
      result.push({ amount: take, badgeType: "take", isAttack });
      return;
    }

    // Show separately
    if (peek > 0) result.push({ amount: peek, badgeType: "peek", isAttack });
    if (take > 0) result.push({ amount: take, badgeType: "take", isAttack });
    if (buy > 0) result.push({ amount: buy, badgeType: "buy", isAttack });
    if (draw > 0) result.push({ amount: draw, badgeType: "none", isAttack });
  };

  if (hasSelf) consolidateGroup(selfDraw, selfPeek, selfTake, selfBuy, selfDiscard, false);
  if (hasAttack)
    consolidateGroup(attackDraw, attackPeek, attackTake, attackBuy, attackDiscard, true);

  return result;
};
