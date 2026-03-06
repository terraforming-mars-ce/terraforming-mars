import React, { useRef, useState, useEffect } from "react";
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
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [activeItem, setActiveItem] = useState<ActiveItem>("game-state");
  const [selectedPlayerIds, setSelectedPlayerIds] = useState<string[]>([]);

  const { position, isDragging, handleMouseDown } = useWindowDrag({
    windowId: WINDOW_ID,
    width: WINDOW_WIDTH,
    height: () => window.innerHeight * 0.7,
    excludeSelectors: EXCLUDE_SELECTORS,
    isVisible,
  });

  const { getZIndex } = useWindowManager();

  const allPlayers = gameState ? [gameState.currentPlayer, ...gameState.otherPlayers] : [];

  useEffect(() => {
    if (allPlayers.length > 0 && selectedPlayerIds.length === 0) {
      setSelectedPlayerIds([allPlayers[0].id]);
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
      selectedPlayerIds,
      onPlayerChange: setSelectedPlayerIds,
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
    <div
      ref={dropdownRef}
      className="debug-dropdown"
      onMouseDown={handleMouseDown}
      style={{
        position: "fixed",
        top: `${position.y}px`,
        left: `${position.x}px`,
        width: `${WINDOW_WIDTH}px`,
        maxHeight: "70vh",
        background: "rgb(0, 0, 0)",
        border: "2px solid #3b82f6",
        borderRadius: "8px",
        zIndex: getZIndex(WINDOW_ID),
        overflow: "hidden",
        display: "flex",
        flexDirection: "column",
        boxShadow: "0 4px 20px rgba(59, 130, 246, 0.3)",
        cursor: isDragging ? "default" : "default",
        transition: isDragging ? "none" : "top 0.2s ease-out, left 0.2s ease-out",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: "12px 16px",
          borderBottom: "1px solid #333",
          userSelect: "none",
          cursor: "default",
        }}
      >
        <h3
          className="font-orbitron"
          style={{
            margin: 0,
            color: "#3b82f6",
            fontSize: "16px",
            display: "flex",
            alignItems: "center",
            gap: "8px",
          }}
        >
          <span style={{ opacity: 0.7, fontSize: "12px" }}>&#x22ee;&#x22ee;</span>
          Admin Tools
        </h3>
        <button
          onClick={onClose}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            background: "none",
            border: "none",
            color: "#abb2bf",
            fontSize: "20px",
            cursor: "pointer",
            padding: "0 4px",
          }}
        >
          ×
        </button>
      </div>

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
    </div>
  );
};

export default DebugDropdown;
