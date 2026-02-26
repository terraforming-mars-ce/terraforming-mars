import React from "react";
import Game3DView from "../../game/view/Game3DView.tsx";
import { TileHighlightMode } from "../../game/board/Tile.tsx";
import { TileVPIndicator } from "../../ui/overlay/EndGameOverlay.tsx";
import { GameDto } from "@/types/generated/api-types.ts";

interface MainContentDisplayProps {
  gameState: GameDto;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
  animateHexEntrance?: boolean;
  onSkyboxReady?: () => void;
  onGpuReady?: () => void;
  showUI?: boolean;
  uiAnimationClass?: string;
}

const MainContentDisplay: React.FC<MainContentDisplayProps> = ({
  gameState,
  tileHighlightMode,
  vpIndicators = [],
  animateHexEntrance = false,
  onSkyboxReady,
  onGpuReady,
  showUI = true,
  uiAnimationClass = "",
}) => {
  return (
    <Game3DView
      gameState={gameState}
      tileHighlightMode={tileHighlightMode}
      vpIndicators={vpIndicators}
      animateHexEntrance={animateHexEntrance}
      onSkyboxReady={onSkyboxReady}
      onGpuReady={onGpuReady}
      showUI={showUI}
      uiAnimationClass={uiAnimationClass}
    />
  );
};

export default MainContentDisplay;
