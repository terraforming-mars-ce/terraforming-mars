import React from "react";
import Game3DView from "../../game/view/Game3DView.tsx";
import { GameDto } from "@/types/generated/api-types.ts";

interface MainContentDisplayProps {
  gameState: GameDto;
  animateHexEntrance?: boolean;
  onSkyboxReady?: () => void;
  onGpuReady?: () => void;
  showUI?: boolean;
  uiAnimationClass?: string;
}

const MainContentDisplay: React.FC<MainContentDisplayProps> = ({
  gameState,
  animateHexEntrance = false,
  onSkyboxReady,
  onGpuReady,
  showUI = true,
  uiAnimationClass = "",
}) => {
  return (
    <Game3DView
      gameState={gameState}
      animateHexEntrance={animateHexEntrance}
      onSkyboxReady={onSkyboxReady}
      onGpuReady={onGpuReady}
      showUI={showUI}
      uiAnimationClass={uiAnimationClass}
    />
  );
};

export default MainContentDisplay;
