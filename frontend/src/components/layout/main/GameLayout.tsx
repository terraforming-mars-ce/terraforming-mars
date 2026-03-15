import { useState, useCallback, forwardRef } from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import type { PlayerListHandle } from "../../ui/list/PlayerList.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";

import BottomResourceBar, {
  BottomResourceBarCallbacks,
} from "../../ui/overlay/BottomResourceBar.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
import PlayedCardNotificationOverlay from "../../ui/overlay/PlayedCardNotification.tsx";
import type { PlayedCardNotification } from "@/hooks/usePlayedCardNotification.ts";
import { StandardProject } from "../../../types/cards.tsx";
import {
  ChatMessageDto,
  GameDto,
  GamePhaseComplete,
  GamePhaseInitApplyCorp,
  GamePhaseInitApplyPrelude,
  PlayerDto,
  OtherPlayerDto,
  CardDto,
  TriggeredEffectDto,
} from "../../../types/generated/api-types.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import GameMenuModal from "../../ui/overlay/GameMenuModal.tsx";
import GameButton from "../../ui/buttons/GameButton.tsx";
import ChatOverlay from "../../ui/overlay/ChatOverlay.tsx";

export type TransitionPhase =
  | "idle"
  | "lobby"
  | "loading"
  | "fadeOutLobby"
  | "marsRevealed"
  | "animateUI"
  | "complete";

interface GameLayoutProps {
  gameState: GameDto;
  currentPlayer: PlayerDto | null;
  playedCards?: CardDto[];
  corporationCard?: CardDto | null;
  showCorporation?: boolean;
  initTurnPlayerId?: string | null;
  showStartingSelection?: boolean;
  transitionPhase?: TransitionPhase;
  animateHexEntrance?: boolean;
  startDark?: boolean;
  tilesHidden?: boolean;
  changedPaths?: Set<string>;
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
  isGameSpectator?: boolean;
  chatMessages?: ChatMessageDto[];
  onSendChatMessage?: (message: string) => void;
  isLobbyPhase?: boolean;
  playerColorMap?: Map<string, string>;
  endgameFadeUI?: boolean;
  isEndgame?: boolean;
  activeEndgamePanel?: "score" | "graphs" | "replay";
  onEndgamePanelChange?: (panel: "score" | "graphs" | "replay") => void;
  hasHistory?: boolean;
  playedCardNotification?: PlayedCardNotification | null;
  isPlayedCardPinned?: boolean;
  onPlayedCardTogglePin?: () => void;
  onPlayedCardAdvance?: () => void;
}

const GameLayout = forwardRef<PlayerListHandle, GameLayoutProps>(function GameLayout(
  {
    gameState,
    currentPlayer,
    playedCards = [],
    corporationCard = null,
    showCorporation = true,
    initTurnPlayerId = null,
    showStartingSelection = false,
    transitionPhase = "idle",
    animateHexEntrance = false,
    startDark = false,
    tilesHidden = false,
    changedPaths = new Set(),
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
    isGameSpectator = false,
    chatMessages,
    onSendChatMessage,
    isLobbyPhase = false,
    playerColorMap,
    endgameFadeUI = false,
    isEndgame = false,
    activeEndgamePanel,
    onEndgamePanelChange,
    hasHistory = false,
    playedCardNotification,
    isPlayedCardPinned,
    onPlayedCardTogglePin,
    onPlayedCardAdvance,
  },
  ref,
) {
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
  const endgameFadeClass = endgameFadeUI ? "opacity-0 pointer-events-none" : "opacity-100";

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
            animateHexEntrance={animateHexEntrance}
            startDark={startDark}
            tilesHidden={tilesHidden}
            onSkyboxReady={onSkyboxReady}
            onGpuReady={onGpuReady}
            showUI={showUI}
            uiAnimationClass={uiAnimationClass}
          />
        </div>
      )}

      {/* TopMenuBar overlays on top — always visible in endgame */}
      {showUI && !showStartingSelection && (
        <div
          className={`${uiAnimationClass} ${isEndgame ? "opacity-100" : endgameFadeClass} transition-opacity duration-700 ease-in-out`}
        >
          <TopMenuBar
            gameState={gameState}
            currentPlayer={currentPlayer}
            onStandardProjectSelect={onStandardProjectSelect}
            onLeaveGame={onLeaveGame}
            onEndGame={onEndGame}
            gameId={gameState?.id}
            isEndgame={isEndgame}
            activeEndgamePanel={activeEndgamePanel}
            onEndgamePanelChange={onEndgamePanelChange}
            hasHistory={hasHistory}
          />
        </div>
      )}

      {/* Chat overlay - rendered before sidebars so it's behind them in z-order */}
      {showUI && !showStartingSelection && chatMessages && onSendChatMessage && (
        <div className={uiAnimationClass}>
          <ChatOverlay
            messages={chatMessages}
            onSendMessage={onSendChatMessage}
            isLobby={isLobbyPhase}
            isEndgame={endgameFadeUI}
            playerColorMap={playerColorMap}
          />
        </div>
      )}

      {/* Overlay Components */}
      {showUI && (
        <div
          className={`${uiAnimationClass} ${endgameFadeClass} transition-opacity duration-700 ease-in-out`}
        >
          <LeftSidebar
            ref={ref}
            players={allPlayers}
            currentPlayer={currentPlayer}
            turnPlayerId={
              gameState?.currentPhase === GamePhaseInitApplyCorp ||
              gameState?.currentPhase === GamePhaseInitApplyPrelude
                ? initTurnPlayerId || ""
                : gameState?.currentTurn || ""
            }
            currentPhase={gameState?.currentPhase}
            hostPlayerId={gameState?.hostPlayerId}
            pendingTilePlayerId={
              gameState?.initPhase?.hasPendingTiles
                ? gameState.initPhase.currentPlayerId
                : currentPlayer?.pendingTileSelection
                  ? currentPlayer.id
                  : undefined
            }
            triggeredEffects={triggeredEffects}
            onPlayerClick={onPlayerClick}
            onKickPlayer={handleKickPlayer}
            onConvertToBot={gameState?.settings?.hasClaudeApiKey ? handleConvertToBot : undefined}
          />

          <RightSidebar
            globalParameters={gameState?.globalParameters}
            generation={gameState?.generation}
            currentPlayer={currentTurnPlayer}
            showVenus={gameState?.settings?.venusNextEnabled}
          />

          <PlayerOverlay players={allPlayers} currentPlayer={currentPlayer} />

          {playedCardNotification && onPlayedCardTogglePin && onPlayedCardAdvance && (
            <PlayedCardNotificationOverlay
              notification={playedCardNotification}
              isPinned={isPlayedCardPinned ?? false}
              onTogglePin={onPlayedCardTogglePin}
              onAdvance={onPlayedCardAdvance}
            />
          )}
        </div>
      )}

      {showUI &&
        !showStartingSelection &&
        (gameState?.currentPhase !== GamePhaseComplete ||
          (isEndgame && activeEndgamePanel === "replay")) && (
          <div className={uiAnimationClass}>
            <BottomResourceBar
              currentPlayer={currentPlayer}
              gameState={gameState}
              playedCards={playedCards}
              changedPaths={changedPaths}
              callbacks={bottomBarCallbacks}
              gameId={gameState?.id}
              corporation={corporationCard}
              showCorporation={showCorporation}
              spectatingPlayer={spectatingPlayer}
              spectatingCorporation={spectatingCorporation}
              spectatePlayerColor={spectatePlayerColor}
              onStopSpectating={onStopSpectating}
              isGameSpectator={isGameSpectator}
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
            <GameButton buttonType="secondary" onClick={() => setPendingAction(null)}>
              Cancel
            </GameButton>
            <GameButton variant="error" onClick={() => void handleConfirmAction()}>
              Kick
            </GameButton>
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
            <GameButton buttonType="secondary" onClick={() => setPendingAction(null)}>
              Cancel
            </GameButton>
            <GameButton variant="error" onClick={() => void handleConfirmAction()}>
              Convert
            </GameButton>
          </div>
        </GameMenuModal>
      )}
    </div>
  );
});

export default GameLayout;
