import React, { useMemo } from "react";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import {
  AdminCommandRequest,
  AdminCommandTypeStartTileSelection,
  PlaceableTileTypeDto,
  StartTileSelectionAdminCommand,
} from "../../../types/generated/api-types.ts";
import { useWindowDrag, useWindowManager } from "./WindowManager.tsx";

interface TilePlacerWindowProps {
  playerId: string;
  playerName: string;
  placeableTileTypes: PlaceableTileTypeDto[];
  onClose: () => void;
}

const WINDOW_ID = "tile-placer";
const WINDOW_WIDTH = 220;

interface TileGroup {
  label: string;
  tiles: { type: string; label: string }[];
}

const TilePlacerWindow: React.FC<TilePlacerWindowProps> = ({
  playerId,
  playerName,
  placeableTileTypes,
  onClose,
}) => {
  const tileGroups = useMemo((): TileGroup[] => {
    const groupMap = new Map<string, { type: string; label: string }[]>();
    const groupOrder: string[] = [];
    for (const tile of placeableTileTypes) {
      if (!groupMap.has(tile.group)) {
        groupMap.set(tile.group, []);
        groupOrder.push(tile.group);
      }
      groupMap.get(tile.group)!.push({ type: tile.type, label: tile.label });
    }
    return groupOrder.map((group) => ({ label: group, tiles: groupMap.get(group)! }));
  }, [placeableTileTypes]);
  const { position, isDragging, handleMouseDown } = useWindowDrag({
    windowId: WINDOW_ID,
    width: WINDOW_WIDTH,
    height: 600,
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

      <div
        style={{
          display: "flex",
          flexDirection: "column",
          gap: "10px",
          maxHeight: "400px",
          overflowY: "auto",
        }}
      >
        {tileGroups.map((group) => (
          <div key={group.label}>
            <div
              style={{
                color: "#888",
                fontSize: "10px",
                fontWeight: "600",
                textTransform: "uppercase",
                letterSpacing: "0.5px",
                marginBottom: "4px",
                paddingLeft: "2px",
              }}
            >
              {group.label}
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: "4px" }}>
              {group.tiles.map((tile) => (
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
                    padding: "8px 12px",
                    background: "rgba(200, 50, 50, 0.1)",
                    border: "1px solid rgba(200, 50, 50, 0.3)",
                    borderRadius: "6px",
                    color: "white",
                    fontSize: "12px",
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
        ))}
      </div>
    </div>
  );
};

export default TilePlacerWindow;
