import FloatingWindow from "./FloatingWindow.tsx";
import React, { useState, useEffect } from "react";
import { useWindowDrag, useWindowManager } from "./WindowManager.tsx";
import { GameDto } from "../../../types/generated/api-types.ts";
import SidebarNav, { type ActiveItem } from "./SidebarNav.tsx";
import GameStatePanel from "./panels/GameStatePanel.tsx";
import PlayerResourcesPage from "./panels/PlayerResourcesPage.tsx";
import PlayerBehaviorPage from "./panels/PlayerBehaviorPage.tsx";
import PlaceTilePage from "./panels/PlaceTilePage.tsx";
import GameCommandsPage from "./panels/GameCommandsPage.tsx";
import World3DCameraPage from "./panels/World3DCameraPage.tsx";
import World3DSunPage from "./panels/World3DSunPage.tsx";
import World3DSkyboxPage from "./panels/World3DSkyboxPage.tsx";

const WINDOW_ID = "admin-tools";
const WINDOW_WIDTH = 780;
const EXCLUDE_SELECTORS = [".tree-expand-toggle", ".tree-node-content", ".debug-content-area"];

interface DebugDropdownProps {
  isVisible: boolean;
  onClose: () => void;
  gameState: GameDto | null;
  changedPaths?: Set<string>;
}

const DebugDropdown: React.FC<DebugDropdownProps> = ({
  isVisible,
  onClose,
  gameState,
  changedPaths = new Set(),
}) => {
  const [activeItem, setActiveItem] = useState<ActiveItem>("game-state");
  const [selectedPlayerId, setSelectedPlayerId] = useState<string>("");

  const { position, handleMouseDown } = useWindowDrag({
    windowId: WINDOW_ID,
    width: WINDOW_WIDTH,
    height: () => window.innerHeight * 0.7,
    excludeSelectors: EXCLUDE_SELECTORS,
    isVisible,
  });

  const { getZIndex } = useWindowManager();

  const allPlayers = gameState ? [gameState.currentPlayer, ...gameState.otherPlayers] : [];

  useEffect(() => {
    if (allPlayers.length > 0 && !selectedPlayerId) {
      setSelectedPlayerId(allPlayers[0].id);
    }
  }, [allPlayers.length]);

  if (!isVisible) {
    return null;
  }

  const is3DItem = (item: ActiveItem) => item.startsWith("3d-");
  const isCommandItem = (item: ActiveItem) => item !== "game-state" && !is3DItem(item);

  const renderContent = () => {
    if (activeItem === "game-state") {
      return <GameStatePanel gameState={gameState} changedPaths={changedPaths} />;
    }

    if (activeItem === "3d-camera") {
      return <World3DCameraPage />;
    }
    if (activeItem === "3d-sun") {
      return <World3DSunPage />;
    }
    if (activeItem === "3d-skybox") {
      return <World3DSkyboxPage />;
    }

    if (!gameState) {
      return (
        <div style={{ color: "#666", textAlign: "center", padding: "20px" }}>
          No game state available
        </div>
      );
    }

    const playerProps = {
      gameState,
      selectedPlayerId,
      onPlayerChange: setSelectedPlayerId,
    };

    switch (activeItem) {
      case "player-resources":
        return <PlayerResourcesPage {...playerProps} />;
      case "player-behavior":
        return <PlayerBehaviorPage {...playerProps} />;
      case "place-tile":
        return <PlaceTilePage {...playerProps} />;
      case "game-commands":
        return <GameCommandsPage gameState={gameState} />;
      default:
        return null;
    }
  };

  return (
    <FloatingWindow
      title="Admin Tools"
      onClose={onClose}
      onMouseDown={handleMouseDown}
      style={{
        top: position.y,
        left: position.x,
        width: WINDOW_WIDTH,
        maxHeight: "70vh",
        zIndex: getZIndex(WINDOW_ID),
      }}
    >
      <div
        style={{
          display: "flex",
          flex: 1,
          overflow: "hidden",
          minHeight: 0,
        }}
      >
        <SidebarNav
          activeItem={activeItem}
          onSelectItem={setActiveItem}
          developmentMode={gameState?.settings.developmentMode ?? false}
        />

        <div
          className="debug-content-area"
          style={{
            flex: 1,
            overflow: isCommandItem(activeItem) ? "visible" : "auto",
            padding: "16px",
            display: "flex",
            flexDirection: "column",
          }}
          onMouseDown={(e) => e.stopPropagation()}
        >
          {renderContent()}
        </div>
      </div>
    </FloatingWindow>
  );
};

export default DebugDropdown;
