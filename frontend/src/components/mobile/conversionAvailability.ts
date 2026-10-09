import {
  PlayerStatusSelection,
  PlayerStatusTile,
  StandardProjectConvertHeatToTemperature,
  StandardProjectConvertPlantsToGreenery,
  type GameDto,
  type PlayerDto,
  type StandardProject,
} from "@/types/generated/api-types.ts";
import { findResourceConversion } from "@/utils/resourceConversionUtils.ts";

export interface ConversionAvailability {
  plants: boolean;
  heat: boolean;
}

export function getConversionAvailability(
  gameState: GameDto,
  currentPlayer: PlayerDto | null,
): ConversionAvailability {
  const blocked =
    currentPlayer?.status === PlayerStatusTile || currentPlayer?.status === PlayerStatusSelection;
  if (blocked) {
    return { plants: false, heat: false };
  }
  const isAvailable = (projectType: StandardProject) =>
    !!findResourceConversion(gameState, projectType)?.available;
  return {
    plants: isAvailable(StandardProjectConvertPlantsToGreenery),
    heat:
      isAvailable(StandardProjectConvertHeatToTemperature) &&
      (gameState.globalParameters?.temperature ?? -30) < 8,
  };
}
