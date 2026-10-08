import React from "react";
import Game3DView from "../../game/view/Game3DView.tsx";
import TileInspectTooltip from "./TileInspectTooltip.tsx";
import PlacementConfirmBar from "../overlay/PlacementConfirmBar.tsx";
import { GameDto } from "@/types/generated/api-types.ts";

interface MainContentDisplayProps {
  gameState: GameDto;
  animateHexEntrance?: boolean;
  startDark?: boolean;
  tilesHidden?: boolean;
  onSkyboxReady?: () => void;
  onGpuReady?: () => void;
  showUI?: boolean;
  uiAnimationClass?: string;
}

const MainContentDisplay: React.FC<MainContentDisplayProps> = ({
  gameState,
  animateHexEntrance = false,
  startDark = false,
  tilesHidden = false,
  onSkyboxReady,
  onGpuReady,
  showUI = true,
  uiAnimationClass = "",
}) => {
  return (
    <>
      <Game3DView
        gameState={gameState}
        animateHexEntrance={animateHexEntrance}
        startDark={startDark}
        tilesHidden={tilesHidden}
        onSkyboxReady={onSkyboxReady}
        onGpuReady={onGpuReady}
        showUI={showUI}
        uiAnimationClass={uiAnimationClass}
      />
      <TileInspectTooltip gameState={gameState} visible={showUI} />
      <PlacementConfirmBar gameState={gameState} visible={showUI} />
    </>
  );
};

export default MainContentDisplay;
