import React, { useEffect, useRef, useState } from "react";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { CARD_PACKS, VENUS_PACK } from "@/constants/cardPacks.ts";
import type { GameDto, MapPreviewTile } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import GameButton from "../buttons/GameButton.tsx";
import MapPreview from "./MapPreview.tsx";
import LobbyPickerDropdown, { type LobbyPickerItem } from "./LobbyPickerDropdown.tsx";

interface LobbyMapInfoPanelProps {
  game: GameDto;
  playerId: string;
}

const PACK_LABELS: Record<string, string> = {
  ...Object.fromEntries(CARD_PACKS.map((p) => [p.id, p.label])),
  [VENUS_PACK.id]: VENUS_PACK.label,
};

const LOCKED_PACK_IDS = new Set(CARD_PACKS.filter((p) => p.lockedOn).map((p) => p.id));

const MAP_FADE_MS = 300;

interface MapLayer {
  key: number;
  tiles: MapPreviewTile[];
}

const LobbyMapInfoPanel: React.FC<LobbyMapInfoPanelProps> = ({ game, playerId }) => {
  const isHost = game.hostPlayerId === playerId;
  const availableMaps = game.settings.availableMaps || [];
  const activeMap = availableMaps.find((m) => m.id === game.settings.mapId);
  const mapName = activeMap?.name || game.settings.mapId || "Tharsis";
  const mapTiles = activeMap?.tiles || [];

  const keyCounterRef = useRef(0);
  const lastMapIdRef = useRef(game.settings.mapId);
  const [layers, setLayers] = useState<MapLayer[]>(() => [{ key: 0, tiles: mapTiles }]);
  const [topKey, setTopKey] = useState(0);
  const [entered, setEntered] = useState(false);

  useEffect(() => {
    const id = requestAnimationFrame(() => setEntered(true));
    return () => cancelAnimationFrame(id);
  }, []);

  useEffect(() => {
    if (lastMapIdRef.current === game.settings.mapId) {
      return;
    }
    lastMapIdRef.current = game.settings.mapId;

    keyCounterRef.current += 1;
    const newKey = keyCounterRef.current;
    setLayers((prev) => [...prev, { key: newKey, tiles: mapTiles }]);
    // Defer the opacity flip to the next frame so the new layer
    // animates from opacity-0 → opacity-100 instead of mounting visible.
    const raf = requestAnimationFrame(() => setTopKey(newKey));
    const cleanup = window.setTimeout(() => {
      setLayers((prev) => prev.filter((l) => l.key === newKey));
    }, MAP_FADE_MS + 50);
    return () => {
      cancelAnimationFrame(raf);
      window.clearTimeout(cleanup);
    };
  }, [game.settings.mapId, mapTiles]);

  const activePacks = [...(game.settings.cardPacks || [])];
  if (game.settings.venusNextEnabled && !activePacks.includes(VENUS_PACK.id)) {
    activePacks.push(VENUS_PACK.id);
  }

  if (availableMaps.length === 0) {
    return null;
  }

  const handleMapSelect = (mapId: string) => {
    if (mapId === game.settings.mapId) {
      return;
    }
    void globalWebSocketManager.updateGameSettings({ mapId });
  };

  const handlePackAdd = (packId: string) => {
    if (packId === VENUS_PACK.id) {
      void globalWebSocketManager.updateGameSettings({ venusNextEnabled: true });
      return;
    }
    const next = [...(game.settings.cardPacks || []), packId];
    void globalWebSocketManager.updateGameSettings({ cardPacks: next });
  };

  const handlePackRemove = (packId: string) => {
    if (packId === VENUS_PACK.id) {
      void globalWebSocketManager.updateGameSettings({ venusNextEnabled: false });
      return;
    }
    const next = (game.settings.cardPacks || []).filter((p) => p !== packId);
    void globalWebSocketManager.updateGameSettings({ cardPacks: next });
  };

  const mapItems: LobbyPickerItem[] = availableMaps.map((m) => ({
    id: m.id,
    name: m.name,
    description: m.description,
  }));

  const addablePacks: LobbyPickerItem[] = [...CARD_PACKS, VENUS_PACK]
    .filter((p) => {
      if (p.lockedOn) {
        return false;
      }
      if (p.id === VENUS_PACK.id) {
        return !game.settings.venusNextEnabled;
      }
      return !(game.settings.cardPacks || []).includes(p.id);
    })
    .map((p) => ({
      id: p.id,
      name: p.label,
      description: p.description,
    }));

  return (
    <div
      className={`hidden xl:flex fixed flex-col gap-3 w-[260px] transition-opacity duration-300 ease-out ${
        entered ? "opacity-100" : "opacity-0"
      }`}
      style={{
        zIndex: Z_INDEX.MENU_DROPDOWN,
        top: "50%",
        right: "calc(50% - 225px - 32px - 260px)",
        transform: "translateY(-50%)",
      }}
    >
      <div className="bg-space-black-darker/85 border border-space-blue-600/50 rounded-xl p-4 backdrop-blur-space shadow-[0_10px_40px_rgba(0,0,0,0.5)]">
        {isHost ? (
          <div className="mb-3">
            <LobbyPickerDropdown
              items={mapItems}
              selectedId={game.settings.mapId}
              onSelect={handleMapSelect}
              trigger={({ open, toggle }) => (
                <button
                  type="button"
                  onClick={toggle}
                  className="w-full flex items-center justify-between gap-2 bg-black/40 border border-space-blue-600/50 rounded-lg px-3 py-2 text-white font-orbitron text-xs font-semibold uppercase tracking-wide transition-colors cursor-pointer hover:border-space-blue-400"
                >
                  <span>{mapName}</span>
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className={`transition-transform ${open ? "rotate-180" : ""}`}
                  >
                    <polyline points="6 9 12 15 18 9" />
                  </svg>
                </button>
              )}
            />
          </div>
        ) : (
          <h3 className="font-orbitron text-white/80 text-xs font-semibold uppercase tracking-wide mb-3 text-center">
            {mapName}
          </h3>
        )}
        <div className="relative flex justify-center">
          {/* Reserve space using an invisible copy of the current tiles. */}
          <div className="invisible">
            <MapPreview tiles={mapTiles} />
          </div>
          {layers.map((layer) => (
            <div
              key={layer.key}
              className="absolute inset-0 flex justify-center transition-opacity ease-out"
              style={{
                transitionDuration: `${MAP_FADE_MS}ms`,
                opacity: layer.key === topKey ? 1 : 0,
              }}
            >
              <MapPreview tiles={layer.tiles} />
            </div>
          ))}
        </div>
      </div>

      <div className="bg-space-black-darker/85 border border-space-blue-600/50 rounded-xl p-4 backdrop-blur-space shadow-[0_10px_40px_rgba(0,0,0,0.5)]">
        <h3 className="font-orbitron text-white/80 text-xs font-semibold uppercase tracking-wide mb-2 text-center">
          Card Packs
        </h3>
        <div className="flex flex-col gap-1.5">
          {activePacks.length === 0 ? (
            <div className="text-white/40 text-xs italic">None</div>
          ) : (
            activePacks.map((pack) => {
              const removable = isHost && !LOCKED_PACK_IDS.has(pack);
              return (
                <div
                  key={pack}
                  className="flex items-center justify-between gap-2 bg-black/30 border border-white/10 rounded-md px-2 py-1.5"
                >
                  <span className="font-orbitron text-white text-xs tracking-wide">
                    {PACK_LABELS[pack] || pack}
                  </span>
                  {removable && (
                    <button
                      type="button"
                      onClick={() => handlePackRemove(pack)}
                      aria-label={`Remove ${PACK_LABELS[pack] || pack}`}
                      className="text-white/40 hover:text-red-400 transition-colors cursor-pointer p-0.5"
                    >
                      <svg
                        width="12"
                        height="12"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      >
                        <line x1="18" y1="6" x2="6" y2="18" />
                        <line x1="6" y1="6" x2="18" y2="18" />
                      </svg>
                    </button>
                  )}
                </div>
              );
            })
          )}
        </div>
        {isHost && (
          <div className="mt-3">
            <LobbyPickerDropdown
              items={addablePacks}
              onSelect={handlePackAdd}
              emptyMessage="All packs enabled"
              trigger={({ toggle }) => (
                <GameButton buttonType="secondary" size="sm" onClick={toggle} className="w-full">
                  + Add
                </GameButton>
              )}
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default LobbyMapInfoPanel;
