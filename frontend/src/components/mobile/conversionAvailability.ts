import {
  PlayerStatusSelection,
  PlayerStatusTile,
  type GameDto,
  type PlayerDto,
} from "@/types/generated/api-types.ts";

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
  const projects = gameState.currentPlayer?.standardProjects ?? [];
  const isAvailable = (projectType: string) =>
    !!projects.find((project) => project.projectType === projectType)?.available;
  return {
    plants: isAvailable("convert-plants-to-greenery"),
    heat:
      isAvailable("convert-heat-to-temperature") &&
      (gameState.globalParameters?.temperature ?? -30) < 8,
  };
}
