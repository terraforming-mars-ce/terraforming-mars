import { useEffect, useMemo } from "react";
import {
  GamePhaseInitApplyCorp,
  GamePhaseInitApplyPrelude,
  type PlayerDto,
} from "@/types/generated/api-types.ts";
import { useAppPhaseStore } from "@/stores/appPhaseStore.ts";
import { useGameStore } from "@/stores/gameStore.ts";
import { useUIOverlayStore } from "@/stores/uiOverlayStore.ts";
import { useMobileUiStore } from "@/stores/mobileUiStore.ts";
import { usePlanetFocus } from "@/contexts/PlanetFocusContext.tsx";
import MainContentDisplay from "../ui/display/MainContentDisplay.tsx";
import CardPlayPresentation from "../ui/overlay/CardPlayPresentation.tsx";
import {
  CardInspections,
  SolarSystemFade,
  type GameLayoutProps,
} from "../layout/main/GameLayout.tsx";
import GameConfirmDialogs from "../layout/main/GameConfirmDialogs.tsx";
import { MobileGameContext, type MobileGameContextValue } from "./MobileGameContext.tsx";
import MobileScreenHost from "./MobileScreenHost.tsx";
import MobileTopBar from "./MobileTopBar.tsx";
import MobileResourceRail from "./MobileResourceRail.tsx";
import MobileStatusRail from "./MobileStatusRail.tsx";
import MobileDock from "./MobileDock.tsx";
import MobileFeedbackFeed from "./feedback/MobileFeedbackFeed.tsx";

export interface MobileGameLayoutProps extends GameLayoutProps {
  onPlayCard: (cardId: string) => void;
  onInspectCard: (cardId: string, source: HTMLElement) => void;
  isReplay: boolean;
}

const noop = () => {};

function hasServerSelection(player: PlayerDto | undefined): boolean {
  if (!player) {
    return false;
  }
  return Boolean(
    player.pendingCardDrawSelection ||
    player.pendingCardSelection ||
    (player.pendingTileSelection?.availableHexes?.length ?? 0) > 0 ||
    player.pendingColonySelection ||
    player.pendingColonyResourceSelection ||
    player.pendingAwardFundSelection ||
    player.pendingFreeTradeSelection ||
    player.pendingEffectSelection ||
    player.pendingResourceRemovalSelection ||
    player.pendingCardReveal ||
    player.pendingBehaviorResolutions?.length,
  );
}

function useCloseScreenOnBlockingInteraction() {
  const serverSelection = useGameStore((s) => hasServerSelection(s.game?.currentPlayer));
  const overlaySelection = useUIOverlayStore(
    (s) =>
      s.showStartingSelection ||
      s.showPendingCardSelection ||
      s.showCardDrawSelection ||
      s.showResourceRemovalSelection ||
      s.showColonyResourceSelection ||
      s.showColonyPlacementSelection ||
      s.showFreeTradeSelection ||
      (s.showProductionPhaseModal && !s.isProductionModalHidden),
  );
  const blocking = serverSelection || overlaySelection;

  useEffect(() => {
    if (blocking) {
      useMobileUiStore.getState().close();
    }
  }, [blocking]);
}

export default function MobileGameLayout({
  gameState,
  inspectionStore,
  inspectionHand,
  inspectionBlocked,
  onInspectionDrag,
  currentPlayer,
  initTurnPlayerId = null,
  showStartingSelection = false,
  animateHexEntrance = false,
  startDark = false,
  tilesHidden = false,
  bottomBarCallbacks,
  onStandardProjectSelect,
  onLeaveGame,
  onEndGame,
  onSkyboxReady,
  onGpuReady,
  onPlayerClick,
  spectatingPlayer,
  onStopSpectating,
  isGameSpectator = false,
  chatMessages,
  onSendChatMessage,
  playerColorMap,
  onPlayCard,
  onInspectCard,
  isReplay,
  triggeredEffects = [],
  playedCardNotification = null,
  isPlayedCardPinned = false,
  onPlayedCardTogglePin,
  onPlayedCardAdvance,
}: MobileGameLayoutProps) {
  const phase = useAppPhaseStore((s) => s.phase);
  const showUI =
    phase.kind === "showcase" ||
    phase.kind === "animateUI" ||
    phase.kind === "playing" ||
    phase.kind === "completed";
  const isShowcase = phase.kind === "showcase";
  const isCompleted = phase.kind === "completed";
  const showHud = showUI && !isShowcase && !isCompleted;
  const isAnimatingIn = phase.kind === "animateUI";
  const uiAnimationClass = isAnimatingIn ? "animate-[uiFadeIn_1200ms_ease-out_both]" : "";
  const earlyUiAnimationClass =
    isShowcase || isAnimatingIn ? "animate-[uiFadeIn_1200ms_ease-out_both]" : "";
  const { activePlanet } = usePlanetFocus();
  const inSolarSystem = activePlanet === "solar-system";

  const isInitPhase =
    gameState.currentPhase === GamePhaseInitApplyCorp ||
    gameState.currentPhase === GamePhaseInitApplyPrelude;
  const turnPlayerId = (isInitPhase ? initTurnPlayerId : gameState.currentTurn) || "";

  useCloseScreenOnBlockingInteraction();

  useEffect(() => {
    if (!showHud) {
      useMobileUiStore.getState().close();
    }
  }, [showHud]);

  const contextValue = useMemo<MobileGameContextValue>(
    () => ({
      gameState,
      currentPlayer,
      isSpectator: isGameSpectator,
      isReplay,
      playerColorMap: playerColorMap ?? new Map(),
      onPlayCard,
      onInspectCard,
      onStandardProjectSelect: onStandardProjectSelect ?? noop,
      onActionSelect: bottomBarCallbacks?.onActionSelect ?? noop,
      onConvertPlantsToGreenery: bottomBarCallbacks?.onConvertPlantsToGreenery ?? noop,
      onConvertHeatToTemperature: bottomBarCallbacks?.onConvertHeatToTemperature ?? noop,
      onPlayerClick: onPlayerClick ?? noop,
      onSendChatMessage: onSendChatMessage ?? noop,
      chatMessages: chatMessages ?? [],
      onLeaveGame: onLeaveGame ?? noop,
      onEndGame: onEndGame ?? noop,
    }),
    [
      gameState,
      currentPlayer,
      isGameSpectator,
      isReplay,
      playerColorMap,
      onPlayCard,
      onInspectCard,
      onStandardProjectSelect,
      bottomBarCallbacks,
      onPlayerClick,
      onSendChatMessage,
      chatMessages,
      onLeaveGame,
      onEndGame,
    ],
  );

  return (
    <MobileGameContext.Provider value={contextValue}>
      <div className="relative w-full h-dvh text-white overflow-hidden">
        <style>{`
          @keyframes uiFadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
          }
        `}</style>

        {phase.kind !== "lobby" && (
          <div className="absolute inset-0 bg-[#000000]">
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

        {showUI && (
          <SolarSystemFade>
            <CardPlayPresentation chatBounds={null} />
            <CardInspections
              store={inspectionStore}
              hand={inspectionHand}
              blocked={inspectionBlocked}
              chatBounds={null}
              onDrag={onInspectionDrag}
            />
          </SolarSystemFade>
        )}

        {showUI && !showStartingSelection && (
          <MobileTopBar
            showStatus={!isShowcase}
            showActions={showHud}
            spectatingPlayer={spectatingPlayer}
            onStopSpectating={onStopSpectating}
            className={earlyUiAnimationClass}
          />
        )}
        {showHud && !showStartingSelection && !inSolarSystem && (
          <MobileResourceRail spectatingPlayer={spectatingPlayer} className={uiAnimationClass} />
        )}
        {showUI && !isCompleted && (
          <MobileStatusRail
            turnPlayerId={turnPlayerId}
            showParameters={showHud && !inSolarSystem}
            className={earlyUiAnimationClass}
          />
        )}
        {showHud && !showStartingSelection && !inSolarSystem && (
          <MobileDock className={uiAnimationClass} />
        )}

        {showHud && !showStartingSelection && (
          <MobileFeedbackFeed
            triggeredEffects={triggeredEffects}
            playedCardNotification={playedCardNotification}
            isPlayedCardPinned={isPlayedCardPinned}
            onPlayedCardTogglePin={onPlayedCardTogglePin}
            onPlayedCardAdvance={onPlayedCardAdvance}
          />
        )}

        <MobileScreenHost />
        <GameConfirmDialogs />
      </div>
    </MobileGameContext.Provider>
  );
}
