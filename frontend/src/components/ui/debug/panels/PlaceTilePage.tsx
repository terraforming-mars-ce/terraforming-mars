import React from "react";
import { globalWebSocketManager } from "../../../../services/globalWebSocketManager.ts";
import {
  GameDto,
  AdminCommandRequest,
  AdminCommandTypeStartTileSelection,
  StartTileSelectionAdminCommand,
} from "../../../../types/generated/api-types.ts";
import PlayerSelector from "../PlayerSelector.tsx";
import { TILE_ICONS } from "../../../../utils/iconStore.ts";

interface PlaceTilePageProps {
  gameState: GameDto;
  selectedPlayerIds: string[];
  onPlayerChange: (ids: string[]) => void;
}

const ICON_MAP: Record<string, string | undefined> = {
  city: TILE_ICONS["city-tile"],
  greenery: TILE_ICONS["greenery-tile"],
  ocean: TILE_ICONS["ocean-tile"],
  volcano: TILE_ICONS["volcano-tile"],
  "land-claim": TILE_ICONS["land-claim"],
  "ecological-zone": TILE_ICONS["greenery-tile"],
  "natural-preserve": TILE_ICONS["greenery-tile"],
  mohole: TILE_ICONS["volcano-tile"],
  "nuclear-zone": TILE_ICONS["tile-placement"],
  mining: TILE_ICONS["tile-placement"],
  restricted: TILE_ICONS["tile-placement"],
  colony: "/assets/tiles/colony.png",
};

const secondaryButtonStyle = {
  padding: "4px 8px",
  background: "transparent",
  border: "1px solid rgba(59, 130, 246, 0.5)",
  borderRadius: "6px",
  color: "#93c5fd",
  fontSize: "10px",
  cursor: "pointer" as const,
  fontWeight: "500" as const,
  display: "inline-flex",
  alignItems: "center",
  gap: "4px",
};

const TILE_GROUPS: {
  label: string;
  extraPadding?: boolean;
  tiles: { type: string; label: string }[];
}[] = [
  {
    label: "Base",
    tiles: [
      { type: "city", label: "City" },
      { type: "greenery", label: "Greenery" },
      { type: "ocean", label: "Ocean" },
    ],
  },
  {
    label: "Special",
    tiles: [
      { type: "volcano", label: "Volcano" },
      { type: "nuclear-zone", label: "Nuclear Zone" },
      { type: "mining", label: "Mining" },
      { type: "natural-preserve", label: "Nat. Preserve" },
      { type: "land-claim", label: "Land Claim" },
    ],
  },
  {
    label: "",
    tiles: [
      { type: "ecological-zone", label: "Eco Zone" },
      { type: "mohole", label: "Mohole" },
      { type: "restricted", label: "Restricted" },
      { type: "colony", label: "Colony" },
    ],
  },
  {
    label: "",
    extraPadding: true,
    tiles: [{ type: "clear", label: "Clear Tile" }],
  },
];

const PlaceTilePage: React.FC<PlaceTilePageProps> = ({
  gameState,
  selectedPlayerIds,
  onPlayerChange,
}) => {
  const allPlayers = [gameState.currentPlayer, ...gameState.otherPlayers];
  const playerId = selectedPlayerIds[0];

  const sendCommand = async (commandType: string, payload: any) => {
    const req: AdminCommandRequest = { commandType: commandType as any, payload };
    try {
      await globalWebSocketManager.sendAdminCommand(req);
    } catch (error) {
      console.error("Failed to send admin command:", error);
    }
  };

  const handleTileSelection = async (tileType: string) => {
    if (!playerId) {
      return;
    }
    const command: StartTileSelectionAdminCommand = { playerId, tileType };
    await sendCommand(AdminCommandTypeStartTileSelection, command);
  };

  const players = allPlayers.map((p) => ({ id: p.id, name: p.name }));

  return (
    <div>
      <PlayerSelector players={players} selectedIds={selectedPlayerIds} onChange={onPlayerChange} />

      <div style={{ marginTop: "12px" }}>
        {TILE_GROUPS.map((group, idx) => (
          <div
            key={group.label || `group-${idx}`}
            style={{ marginBottom: "6px", paddingTop: group.extraPadding ? "8px" : undefined }}
          >
            <div style={{ display: "flex", alignItems: "center", gap: "6px" }}>
              <span style={{ color: "#888", fontSize: "10px", minWidth: "45px" }}>
                {group.label ? `${group.label}:` : ""}
              </span>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "4px" }}>
                {group.tiles.map((tile) => {
                  const iconSrc = ICON_MAP[tile.type];
                  return (
                    <button
                      key={tile.type}
                      onClick={() => void handleTileSelection(tile.type)}
                      style={secondaryButtonStyle}
                    >
                      {tile.type === "clear" ? (
                        <span style={{ fontSize: "12px", lineHeight: 1 }}>✕</span>
                      ) : iconSrc ? (
                        <img
                          src={iconSrc}
                          alt=""
                          style={{ width: "14px", height: "14px", objectFit: "contain" }}
                        />
                      ) : null}
                      {tile.label}
                    </button>
                  );
                })}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default PlaceTilePage;
