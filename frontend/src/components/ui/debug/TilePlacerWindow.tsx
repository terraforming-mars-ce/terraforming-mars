import React from "react";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import {
  AdminCommandRequest,
  AdminCommandTypeStartTileSelection,
  StartTileSelectionAdminCommand,
} from "../../../types/generated/api-types.ts";
import { useWindowDrag, useWindowManager } from "./WindowManager.tsx";

interface TilePlacerWindowProps {
  playerId: string;
  playerName: string;
  onClose: () => void;
}

const WINDOW_ID = "tile-placer";
const WINDOW_WIDTH = 220;

const TILE_TYPES = [
  { type: "city", label: "City" },
  { type: "greenery", label: "Greenery" },
  { type: "ocean", label: "Ocean" },
  { type: "volcano", label: "Volcano" },
  { type: "clear", label: "Clear" },
];

const TilePlacerWindow: React.FC<TilePlacerWindowProps> = ({ playerId, playerName, onClose }) => {
  const { position, isDragging, handleMouseDown } = useWindowDrag({
    windowId: WINDOW_ID,
    width: WINDOW_WIDTH,
    height: 300,
    defaultPosition:
      typeof window !== "undefined"
        ? { x: window.innerWidth - WINDOW_WIDTH - 40, y: 60 }
        : undefined,
  });

  const { getZIndex } = useWindowManager();

  const handleTileClick = async (tileType: string) => {
    const command: StartTileSelectionAdminCommand = {
      playerId: playerId,
      tileType: tileType,
    };
    const adminRequest: AdminCommandRequest = {
      commandType: AdminCommandTypeStartTileSelection as any,
      payload: command,
    };
    try {
      await globalWebSocketManager.sendAdminCommand(adminRequest);
    } catch (error) {
      console.error("Failed to send tile placer command:", error);
    }
  };

  const accentColor = "rgb(200, 50, 50)";

  return (
    <div
      onMouseDown={handleMouseDown}
      style={{
        position: "fixed",
        top: `${position.y}px`,
        left: `${position.x}px`,
        width: `${WINDOW_WIDTH}px`,
        background: "rgba(0, 0, 0, 0.95)",
        border: `2px solid ${accentColor}`,
        borderRadius: "8px",
        padding: "12px",
        zIndex: getZIndex(WINDOW_ID),
        display: "flex",
        flexDirection: "column",
        boxShadow: `0 4px 20px rgba(200, 50, 50, 0.3)`,
        cursor: isDragging ? "grabbing" : "default",
        transition: isDragging ? "none" : "top 0.2s ease-out, left 0.2s ease-out",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "10px",
          paddingBottom: "10px",
          borderBottom: "1px solid rgba(200, 50, 50, 0.3)",
          userSelect: "none",
          cursor: "grab",
        }}
      >
        <div>
          <h3
            style={{
              margin: 0,
              color: accentColor,
              fontSize: "14px",
              display: "flex",
              alignItems: "center",
              gap: "6px",
            }}
          >
            <span style={{ opacity: 0.7, fontSize: "10px" }}>⋮⋮</span>
            Tile Placer
          </h3>
          <div style={{ color: "#888", fontSize: "10px", marginTop: "2px" }}>{playerName}</div>
        </div>
        <button
          onClick={onClose}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            background: "none",
            border: "none",
            color: "#abb2bf",
            fontSize: "18px",
            cursor: "pointer",
            padding: "0 4px",
          }}
        >
          ×
        </button>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
        {TILE_TYPES.map((tile) => (
          <button
            key={tile.type}
            onClick={() => void handleTileClick(tile.type)}
            onMouseDown={(e) => e.stopPropagation()}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = "rgba(200, 50, 50, 0.25)";
              e.currentTarget.style.borderColor = "rgba(200, 50, 50, 0.6)";
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = "rgba(200, 50, 50, 0.1)";
              e.currentTarget.style.borderColor = "rgba(200, 50, 50, 0.3)";
            }}
            style={{
              padding: "10px 14px",
              background: "rgba(200, 50, 50, 0.1)",
              border: "1px solid rgba(200, 50, 50, 0.3)",
              borderRadius: "6px",
              color: "white",
              fontSize: "13px",
              fontWeight: "500",
              cursor: "pointer",
              transition: "all 0.15s ease",
              textAlign: "left",
            }}
          >
            {tile.label}
          </button>
        ))}
      </div>
    </div>
  );
};

export default TilePlacerWindow;
