import React from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";
import { TileHighlightMode } from "../../game/board/Tile.tsx";
import { TileVPIndicator } from "../../ui/overlay/EndGameOverlay.tsx";
import BottomResourceBar, {
  BottomResourceBarCallbacks,
} from "../../ui/overlay/BottomResourceBar.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
import { StandardProject } from "../../../types/cards.tsx";
import {
  GameDto,
  PlayerDto,
  OtherPlayerDto,
  CardDto,
  TriggeredEffectDto,
} from "../../../types/generated/api-types.ts";

export type TransitionPhase =
  | "idle"
  | "lobby"
  | "loading"
  | "fadeOutLobby"
  | "animateUI"
  | "complete";

interface GameLayoutProps {
  gameState: GameDto;
  currentPlayer: PlayerDto | null;
  playedCards?: CardDto[];
  corporationCard?: CardDto | null;
  showCardSelection?: boolean;
  transitionPhase?: TransitionPhase;
  animateHexEntrance?: boolean;
  changedPaths?: Set<string>;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
  triggeredEffects?: TriggeredEffectDto[];
  bottomBarCallbacks?: BottomResourceBarCallbacks;
  onStandardProjectSelect?: (project: StandardProject) => void;
  onLeaveGame?: () => void;
  onSkyboxReady?: () => void;
  onGpuReady?: () => void;
}

const GameLayout: React.FC<GameLayoutProps> = ({
  gameState,
  currentPlayer,
  playedCards = [],
  corporationCard = null,
  showCardSelection = false,
  transitionPhase = "idle",
  animateHexEntrance = false,
  changedPaths = new Set(),
  tileHighlightMode,
  vpIndicators = [],
  triggeredEffects = [],
  bottomBarCallbacks,
  onStandardProjectSelect,
  onLeaveGame,
  onSkyboxReady,
  onGpuReady,
}) => {
  // Create a map of all players (current + others) for easy lookup
  const playerMap = new Map<string, PlayerDto | OtherPlayerDto>();
  if (gameState?.currentPlayer) {
    playerMap.set(gameState.currentPlayer.id, gameState.currentPlayer);
  }
  gameState?.otherPlayers?.forEach((otherPlayer) => {
    playerMap.set(otherPlayer.id, otherPlayer);
  });

  // Construct allPlayers using the turn order from the backend
  const allPlayers: (PlayerDto | OtherPlayerDto)[] =
    (gameState?.turnOrder
      ?.map((playerId) => playerMap.get(playerId))
      .filter((player) => player !== undefined) as (PlayerDto | OtherPlayerDto)[]) || [];

  // Find the current turn player for the right sidebar
  const currentTurnPlayer =
    allPlayers.find((player) => player.id === gameState?.currentTurn) || null;

  const showUI =
    transitionPhase === "animateUI" || transitionPhase === "complete" || transitionPhase === "idle";
  const isAnimatingIn = transitionPhase === "animateUI";
  const uiAnimationClass = isAnimatingIn ? "animate-[uiFadeIn_1200ms_ease-out_both]" : "";

  return (
    <div className="relative w-screen h-screen bg-[#000000] text-white overflow-hidden">
      {/* CSS animations for transition */}
      <style>{`
        @keyframes uiFadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
      `}</style>

      {/* Game content takes full screen - hidden during lobby (SpaceBackground shown instead) */}
      {transitionPhase !== "lobby" && (
        <div className="absolute inset-0">
          <MainContentDisplay
            gameState={gameState}
            tileHighlightMode={tileHighlightMode}
            vpIndicators={vpIndicators}
            animateHexEntrance={animateHexEntrance}
            onSkyboxReady={onSkyboxReady}
            onGpuReady={onGpuReady}
            showUI={showUI}
            uiAnimationClass={uiAnimationClass}
          />
        </div>
      )}

      {/* TopMenuBar overlays on top */}
      {showUI && !showCardSelection && (
        <div className={uiAnimationClass}>
          <TopMenuBar
            gameState={gameState}
            currentPlayer={currentPlayer}
            onStandardProjectSelect={onStandardProjectSelect}
            onLeaveGame={onLeaveGame}
            gameId={gameState?.id}
          />
        </div>
      )}

      {/* Overlay Components */}
      {showUI && (
        <div className={uiAnimationClass}>
          <LeftSidebar
            players={allPlayers}
            currentPlayer={currentPlayer}
            turnPlayerId={gameState?.currentTurn || ""}
            currentPhase={gameState?.currentPhase}
            hasPendingTilePlacement={!!currentPlayer?.pendingTileSelection}
            triggeredEffects={triggeredEffects}
          />

          <RightSidebar
            globalParameters={gameState?.globalParameters}
            generation={gameState?.generation}
            currentPlayer={currentTurnPlayer}
          />

          <PlayerOverlay players={allPlayers} currentPlayer={currentPlayer} />
        </div>
      )}

      {showUI && !showCardSelection && (
        <div className={uiAnimationClass}>
          <BottomResourceBar
            currentPlayer={currentPlayer}
            gameState={gameState}
            playedCards={playedCards}
            changedPaths={changedPaths}
            callbacks={bottomBarCallbacks}
            gameId={gameState?.id}
            corporation={corporationCard}
          />
        </div>
      )}
    </div>
  );
};

export default GameLayout;
