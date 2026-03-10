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
import MainMenuSettingsButton from "./components/ui/buttons/MainMenuSettingsButton.tsx";
import SpaceBackground from "./components/3d/SpaceBackground.tsx";
import LoadingOverlay from "./components/game/view/LoadingOverlay.tsx";
import "./App.css";

function App() {
  const [isWebSocketReady, setIsWebSocketReady] = useState(false);

  // Initialize WebSocket connection once on app startup
  useEffect(() => {
    const initializeWebSocket = async () => {
      try {
        // console.log("Initializing global WebSocket connection...");
        await globalWebSocketManager.initialize();
        // console.log("Global WebSocket connection ready");
        setIsWebSocketReady(true);
      } catch (error) {
        console.error("Failed to initialize WebSocket:", error);
        // Continue running app even if WebSocket fails initially
        // It will retry connection when needed
        setIsWebSocketReady(true); // Allow app to continue
      }
    };

    void initializeWebSocket();
  }, []); // Empty dependency array - runs once on app mount

  // Show loading while WebSocket is initializing
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

function AppWithBackground() {
  const location = useLocation();
  const { isLoaded, error } = useSpaceBackground();
  const [overlayVisible, setOverlayVisible] = useState(() => !skyboxCache.isReady());
  const showSpaceBackground = ["/", "/create", "/join"].includes(location.pathname);

  const skyboxReady = isLoaded || !!error;

  useEffect(() => {
    if (!showSpaceBackground) {
      setOverlayVisible(false);
    } else if (!skyboxCache.isReady()) {
      setOverlayVisible(true);
    }
  }, [showSpaceBackground]);

  useEffect(() => {
    if (showSpaceBackground && isLoaded) {
      audioService.playAmbient();
    } else if (!showSpaceBackground && location.pathname !== "/game") {
      audioService.stopAmbient();
    }
  }, [showSpaceBackground, isLoaded, location.pathname]);

  return (
    <>
      {showSpaceBackground && <SpaceBackground animationSpeed={0.5} overlayOpacity={0.3} />}
      {showSpaceBackground && overlayVisible && (
        <LoadingOverlay isLoaded={skyboxReady} onTransitionEnd={() => setOverlayVisible(false)} />
      )}
      {showSpaceBackground && !overlayVisible && <MainMenuSettingsButton />}
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

export default App;
