import {
  GameDto,
  PlayerEffectDto,
  PlayerResourceConversionDto,
  StandardProject,
  StandardProjectConvertPlantsToGreenery,
  StandardProjectConvertHeatToTemperature,
} from "../types/generated/api-types.ts";

const BASE_PLANTS_FOR_GREENERY = 8;
const BASE_HEAT_FOR_TEMPERATURE = 8;

export function findResourceConversion(
  gameState: GameDto | undefined,
  projectType: StandardProject,
): PlayerResourceConversionDto | undefined {
  return gameState?.currentPlayer?.resourceConversions?.find(
    (conversion) => conversion.projectType === projectType,
  );
}

/**
 * Calculates the required plants for greenery conversion considering player discounts
 * @param playerEffects - The player's active effects
 * @returns The required number of plants after discounts (minimum 1)
 */
export function calculatePlantsForGreenery(playerEffects?: PlayerEffectDto[]): number {
  return calculateDiscountedCost(
    BASE_PLANTS_FOR_GREENERY,
    playerEffects,
    StandardProjectConvertPlantsToGreenery,
  );
}

/**
 * Calculates the required heat for temperature conversion considering player discounts
 * @param playerEffects - The player's active effects
 * @returns The required number of heat after discounts (minimum 1)
 */
export function calculateHeatForTemperature(playerEffects?: PlayerEffectDto[]): number {
  return calculateDiscountedCost(
    BASE_HEAT_FOR_TEMPERATURE,
    playerEffects,
    StandardProjectConvertHeatToTemperature,
  );
}

/**
 * Generic discount calculator for resource conversions
 * @param baseCost - The base cost before discounts
 * @param playerEffects - The player's active effects
 * @param standardProject - The standard project type
 * @returns The final cost after discounts (minimum 1)
 */
function calculateDiscountedCost(
  baseCost: number,
  playerEffects: PlayerEffectDto[] | undefined,
  standardProject: string,
): number {
  if (!playerEffects || playerEffects.length === 0) {
    return baseCost;
  }

  let totalDiscount = 0;

  for (const effect of playerEffects) {
    for (const output of effect.behavior.outputs ?? []) {
      if (output.type !== "discount") {
        continue;
      }

      // Check if any selector targets this standard project
      const matchesStandardProject = output.selectors?.some((selector) =>
        selector.standardProjects?.includes(standardProject as never),
      );

      if (!matchesStandardProject) {
        continue;
      }

      totalDiscount += output.amount;
    }
  }

  const finalCost = baseCost - totalDiscount;
  return Math.max(finalCost, 1);
}
