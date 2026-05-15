import { BrowserRouter as Router, Routes, Route, useLocation } from "react-router-dom";
import { useEffect, useState } from "react";
import GameInterface from "./components/layout/main/GameInterface.tsx";
import CreateGamePage from "./components/pages/CreateGamePage.tsx";
import JoinGamePage from "./components/pages/JoinGamePage.tsx";
import CardsPage from "./components/pages/CardsPage.tsx";
import GameLandingPage from "./components/pages/GameLandingPage.tsx";
import ReconnectingPage from "./components/pages/ReconnectingPage.tsx";
import { globalWebSocketManager } from "./services/globalWebSocketManager.ts";
import { SpaceBackgroundProvider, useSpaceBackground } from "./contexts/SpaceBackgroundContext.tsx";
import { SoundProvider } from "./contexts/SoundContext.tsx";
import { NotificationProvider } from "./contexts/NotificationContext.tsx";
import { World3DSettingsProvider } from "./contexts/World3DSettingsContext.tsx";
import NotificationContainer from "./components/ui/notifications/NotificationContainer.tsx";
import { audioService } from "./services/audioService.ts";
import { skyboxCache } from "./services/SkyboxCache.ts";
import MainMenuHamburger from "./components/ui/buttons/MainMenuHamburger.tsx";
import SpaceBackground from "./components/3d/SpaceBackground.tsx";
import LoadingOverlay from "./components/game/view/LoadingOverlay.tsx";
import { useAppPhaseStore, showsSpaceBackground, showsMenuChrome } from "./stores/appPhaseStore.ts";
import FeedbackWindow from "./components/ui/debug/FeedbackWindow.tsx";
import { WindowManagerProvider } from "./components/ui/debug/WindowManager.tsx";
import { APP_VERSION } from "./config.ts";
import "./App.css";

function App() {
  const [isWebSocketReady, setIsWebSocketReady] = useState(false);

  useEffect(() => {
    const initializeWebSocket = async () => {
      try {
        await globalWebSocketManager.initialize();
        setIsWebSocketReady(true);
      } catch (error) {
        console.error("Failed to initialize WebSocket:", error);
        setIsWebSocketReady(true);
      }
    };

    void initializeWebSocket();
  }, []);

  if (!isWebSocketReady) {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "100vh",
          background: "#000011",
          color: "white",
          fontSize: "18px",
        }}
      >
        Connecting to server...
      </div>
    );
  }

  return (
    <SoundProvider>
      <SpaceBackgroundProvider>
        <World3DSettingsProvider>
          <div className="App" style={{ margin: 0, padding: 0 }}>
            <Router>
              <NotificationProvider>
                <AppWithBackground />
                <NotificationContainer />
              </NotificationProvider>
            </Router>
          </div>
        </World3DSettingsProvider>
      </SpaceBackgroundProvider>
    </SoundProvider>
  );
}

function routeForPathname(pathname) {
  if (pathname === "/create") {
    return "create";
  }
  if (pathname === "/join") {
    return "join";
  }
  if (pathname === "/cards") {
    return "cards";
  }
  if (pathname === "/reconnecting") {
    return "reconnecting";
  }
  return "landing";
}

function AppWithBackground() {
  const location = useLocation();
  const { isLoaded, error } = useSpaceBackground();
  const [overlayVisible, setOverlayVisible] = useState(() => !skyboxCache.isReady());
  const phase = useAppPhaseStore((s) => s.phase);
  const setPhase = useAppPhaseStore((s) => s.setPhase);

  const inMenuRoute = ["/", "/create", "/join", "/cards", "/reconnecting"].includes(
    location.pathname,
  );
  const showSpaceBackgroundLayer = showsSpaceBackground(phase);
  const showMenuChrome = showsMenuChrome(phase);

  useEffect(() => {
    if (!inMenuRoute) {
      return;
    }
    const route = routeForPathname(location.pathname);
    if (phase.kind !== "menu" || phase.route !== route) {
      setPhase({ kind: "menu", route });
    }
  }, [inMenuRoute, location.pathname, phase, setPhase]);

  const skyboxReady = isLoaded || !!error;

  useEffect(() => {
    if (!inMenuRoute) {
      setOverlayVisible(false);
    } else if (!skyboxCache.isReady()) {
      setOverlayVisible(true);
    }
  }, [inMenuRoute]);

  useEffect(() => {
    if (showSpaceBackgroundLayer && isLoaded) {
      audioService.playAmbient();
    } else if (!showSpaceBackgroundLayer && location.pathname !== "/game") {
      audioService.stopAmbient();
    }
  }, [showSpaceBackgroundLayer, isLoaded, location.pathname]);

  return (
    <>
      <div
        style={{
          opacity: showSpaceBackgroundLayer ? 1 : 0,
          transition: "opacity 1500ms ease-out",
          pointerEvents: showSpaceBackgroundLayer ? "auto" : "none",
        }}
      >
        <SpaceBackground animationSpeed={0.5} overlayOpacity={0.3} />
      </div>
      {inMenuRoute && overlayVisible && (
        <LoadingOverlay
          isLoaded={skyboxReady}
          onTransitionEnd={() => setOverlayVisible(false)}
          showDelayMs={0}
          minDurationMs={500}
        />
      )}
      {showMenuChrome && !overlayVisible && <MainMenuHamburger />}
      {showMenuChrome && !overlayVisible && <MenuFooter />}
      <Routes>
        <Route path="/" element={<GameLandingPage />} />
        <Route path="/create" element={<CreateGamePage />} />
        <Route path="/join" element={<JoinGamePage />} />
        <Route path="/cards" element={<CardsPage />} />
        <Route path="/reconnecting" element={<ReconnectingPage />} />
        <Route path="/game/:gameId" element={<GameInterface />} />
        <Route path="/game" element={<GameInterface />} />
      </Routes>
    </>
  );
}

function MenuFooter() {
  const [showFeedbackWindow, setShowFeedbackWindow] = useState(false);

  useEffect(() => {
    const handleToggleFeedback = () => setShowFeedbackWindow((prev) => !prev);
    window.addEventListener("toggle-feedback-window", handleToggleFeedback);
    return () => window.removeEventListener("toggle-feedback-window", handleToggleFeedback);
  }, []);

  return (
    <>
      <div className="fixed bottom-[16px] left-[16px] right-[16px] flex items-center justify-between text-white/30 text-xs select-none z-20 pointer-events-none">
        <span className="pointer-events-auto">
          {APP_VERSION}
          <span className="mx-1">|</span>
          <button
            className="hover:text-white/70 transition-colors cursor-pointer"
            onClick={() => window.dispatchEvent(new CustomEvent("toggle-feedback-window"))}
          >
            Feedback
          </button>
        </span>
        <a
          href="/cards"
          className="pointer-events-auto bg-space-black-darker/90 border-2 border-space-blue-500 rounded-lg font-orbitron font-semibold text-white text-sm py-1.5 px-3 no-underline inline-block backdrop-blur-space hover:border-space-blue-400 hover:shadow-[0_0_12px_rgba(255,255,255,0.15)] transition-all duration-200"
        >
          View Cards
        </a>
      </div>
      <WindowManagerProvider>
        <FeedbackWindow
          isVisible={showFeedbackWindow}
          onClose={() => setShowFeedbackWindow(false)}
          gameState={null}
        />
      </WindowManagerProvider>
    </>
  );
}

export default App;
