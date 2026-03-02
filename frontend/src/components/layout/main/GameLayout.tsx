import React, { useState, useCallback } from "react";
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
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import GameMenuModal from "../../ui/overlay/GameMenuModal.tsx";
import GameMenuButton from "../../ui/buttons/GameMenuButton.tsx";

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
  showStartingSelection?: boolean;
  transitionPhase?: TransitionPhase;
  animateHexEntrance?: boolean;
  changedPaths?: Set<string>;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
  triggeredEffects?: TriggeredEffectDto[];
  bottomBarCallbacks?: BottomResourceBarCallbacks;
  onStandardProjectSelect?: (project: StandardProject) => void;
  onLeaveGame?: () => void;
  onEndGame?: () => void;
  onSkyboxReady?: () => void;
  onGpuReady?: () => void;
  onPlayerClick?: (player: PlayerDto | OtherPlayerDto) => void;
  spectatingPlayer?: PlayerDto | OtherPlayerDto | null;
  spectatingCorporation?: CardDto | null;
  spectatePlayerColor?: string;
  onStopSpectating?: () => void;
}

const GameLayout: React.FC<GameLayoutProps> = ({
  gameState,
  currentPlayer,
  playedCards = [],
  corporationCard = null,
  showStartingSelection = false,
  transitionPhase = "idle",
  animateHexEntrance = false,
  changedPaths = new Set(),
  tileHighlightMode,
  vpIndicators = [],
  triggeredEffects = [],
  bottomBarCallbacks,
  onStandardProjectSelect,
  onLeaveGame,
  onEndGame,
  onSkyboxReady,
  onGpuReady,
  onPlayerClick,
  spectatingPlayer,
  spectatingCorporation,
  spectatePlayerColor,
  onStopSpectating,
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

  const [pendingAction, setPendingAction] = useState<{
    type: "kick" | "convertToBot";
    playerId: string;
    playerName: string;
  } | null>(null);

  const handleKickPlayer = useCallback(
    (playerId: string) => {
      const player = allPlayers.find((p) => p.id === playerId);
      setPendingAction({ type: "kick", playerId, playerName: player?.name || "Unknown" });
    },
    [allPlayers],
  );

  const handleConvertToBot = useCallback(
    (playerId: string) => {
      const player = allPlayers.find((p) => p.id === playerId);
      setPendingAction({ type: "convertToBot", playerId, playerName: player?.name || "Unknown" });
    },
    [allPlayers],
  );

  const handleConfirmAction = async () => {
    if (!pendingAction) return;
    setPendingAction(null);
    try {
      if (pendingAction.type === "kick") {
        await globalWebSocketManager.kickPlayer(pendingAction.playerId);
      } else {
        await globalWebSocketManager.convertToBot(pendingAction.playerId);
      }
    } catch (error) {
      console.error("Failed to execute action:", error);
    }
  };

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
      {showUI && !showStartingSelection && (
        <div className={uiAnimationClass}>
          <TopMenuBar
            gameState={gameState}
            currentPlayer={currentPlayer}
            onStandardProjectSelect={onStandardProjectSelect}
            onLeaveGame={onLeaveGame}
            onEndGame={onEndGame}
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
            hostPlayerId={gameState?.hostPlayerId}
            hasPendingTilePlacement={!!currentPlayer?.pendingTileSelection}
            triggeredEffects={triggeredEffects}
            onPlayerClick={onPlayerClick}
            onKickPlayer={handleKickPlayer}
            onConvertToBot={handleConvertToBot}
          />

          <RightSidebar
            globalParameters={gameState?.globalParameters}
            generation={gameState?.generation}
            currentPlayer={currentTurnPlayer}
            showVenus={gameState?.settings?.venusNextEnabled}
          />

          <PlayerOverlay players={allPlayers} currentPlayer={currentPlayer} />
        </div>
      )}

      {showUI && !showStartingSelection && (
        <div className={uiAnimationClass}>
          <BottomResourceBar
            currentPlayer={currentPlayer}
            gameState={gameState}
            playedCards={playedCards}
            changedPaths={changedPaths}
            callbacks={bottomBarCallbacks}
            gameId={gameState?.id}
            corporation={corporationCard}
            spectatingPlayer={spectatingPlayer}
            spectatingCorporation={spectatingCorporation}
            spectatePlayerColor={spectatePlayerColor}
            onStopSpectating={onStopSpectating}
          />
        </div>
      )}

      {pendingAction?.type === "kick" && (
        <GameMenuModal
          title="Kick player?"
          showBackdrop={true}
          onClose={() => setPendingAction(null)}
          zIndex={10000}
        >
          <p className="text-white/80 text-center mb-6">
            <span className="font-bold text-white">{pendingAction.playerName}</span> will be removed
            from the game and cannot rejoin.
          </p>
          <div className="flex gap-4 justify-center">
            <GameMenuButton variant="secondary" onClick={() => setPendingAction(null)}>
              Cancel
            </GameMenuButton>
            <GameMenuButton variant="error" onClick={() => void handleConfirmAction()}>
              Kick
            </GameMenuButton>
          </div>
        </GameMenuModal>
      )}

      {pendingAction?.type === "convertToBot" && (
        <GameMenuModal
          title="Convert to bot?"
          showBackdrop={true}
          onClose={() => setPendingAction(null)}
          zIndex={10000}
        >
          <p className="text-white/80 text-center mb-6">
            <span className="font-bold text-white">{pendingAction.playerName}</span> will be
            replaced by a bot. This cannot be undone.
          </p>
          <div className="flex gap-4 justify-center">
            <GameMenuButton variant="secondary" onClick={() => setPendingAction(null)}>
              Cancel
            </GameMenuButton>
            <GameMenuButton variant="error" onClick={() => void handleConfirmAction()}>
              Convert
            </GameMenuButton>
          </div>
        </GameMenuModal>
      )}
    </div>
  );
};

export default GameLayout;
