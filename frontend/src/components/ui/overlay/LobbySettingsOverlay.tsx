import React, { useState } from "react";
import { CARD_PACKS, VENUS_PACK } from "@/constants/cardPacks.ts";
import type { GameDto, UpdateGameSettingsRequest } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import GameButton from "../buttons/GameButton.tsx";
import InfoTooltip from "../display/InfoTooltip.tsx";
import LobbyPickerDropdown, { type LobbyPickerItem } from "../lobby/LobbyPickerDropdown.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_FOOTER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
} from "./overlayStyles.ts";

interface LobbySettingsOverlayProps {
  game: GameDto;
  playerId: string;
  isOpen: boolean;
  onClose: () => void;
}

type Tab = "general" | "cardPacks" | "ai";

const LobbySettingsOverlay: React.FC<LobbySettingsOverlayProps> = ({
  game,
  playerId,
  isOpen,
  onClose,
}) => {
  const isHost = game.hostPlayerId === playerId;
  const [activeTab, setActiveTab] = useState<Tab>("general");
  const [claudeToken, setClaudeToken] = useState("");
  const [tokenStatus, setTokenStatus] = useState<"idle" | "saving" | "saved">("idle");

  if (!isOpen || !isHost) {
    return null;
  }

  const settings = game.settings;
  const availableMaps = settings.availableMaps || [];
  const currentMapName =
    availableMaps.find((m) => m.id === settings.mapId)?.name || settings.mapId || "Tharsis";
  const playerCount = (game.currentPlayer?.id ? 1 : 0) + (game.otherPlayers?.length || 0);

  const mapItems: LobbyPickerItem[] = availableMaps.map((m) => ({
    id: m.id,
    name: m.name,
    description: m.description,
  }));

  const dispatch = (patch: UpdateGameSettingsRequest) => {
    void globalWebSocketManager.updateGameSettings(patch);
  };

  const toggleVenus = () => {
    dispatch({ venusNextEnabled: !settings.venusNextEnabled });
  };

  const togglePack = (packId: string) => {
    const current = settings.cardPacks || [];
    const next = current.includes(packId)
      ? current.filter((p) => p !== packId)
      : [...current, packId];
    dispatch({ cardPacks: next });
  };

  const handleMapSelect = (mapId: string) => {
    if (mapId !== settings.mapId) {
      dispatch({ mapId });
    }
  };

  const handleSubmitToken = () => {
    if (!claudeToken.trim()) {
      return;
    }
    setTokenStatus("saving");
    dispatch({ claudeApiKey: claudeToken.trim() });
    setClaudeToken("");
    setTokenStatus("saved");
    window.setTimeout(() => setTokenStatus("idle"), 1500);
  };

  const handleClearToken = () => {
    dispatch({ claudeApiKey: "" });
  };

  const clamp = (value: number, min: number, max: number) => Math.max(min, Math.min(max, value));

  const handleMaxPlayersChange = (raw: number) => {
    const value = clamp(Math.round(raw), 1, 10);
    if (value === settings.maxPlayers) {
      return;
    }
    dispatch({ maxPlayers: value });
  };

  const tabClass = (tab: Tab) =>
    `text-left px-4 py-3 text-sm font-semibold uppercase tracking-wide transition-colors cursor-pointer ${
      activeTab === tab
        ? "text-white bg-space-blue-600/30 border-r-2 border-space-blue-400"
        : "text-white/50 hover:text-white/80 hover:bg-white/5"
    }`;

  const minPlayersAllowed = Math.max(1, playerCount);

  return (
    <div
      className="fixed inset-0 flex items-center justify-center bg-black/70 backdrop-blur-sm animate-[fadeIn_0.3s_ease]"
      style={{ zIndex: Z_INDEX.LOBBY_SETTINGS_MODAL }}
    >
      <div className={`${OVERLAY_CONTAINER_CLASS} max-w-[1100px] h-[80vh]`}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Lobby Settings</h2>
        </div>

        <div className="flex-1 flex min-h-0">
          <div className="w-44 shrink-0 bg-black/30 border-r border-space-blue-600/50 flex flex-col py-2">
            <button onClick={() => setActiveTab("general")} className={tabClass("general")}>
              General
            </button>
            <button onClick={() => setActiveTab("cardPacks")} className={tabClass("cardPacks")}>
              Card Packs
            </button>
            <button onClick={() => setActiveTab("ai")} className={tabClass("ai")}>
              AI
            </button>
          </div>

          <div className="flex-1 overflow-y-auto p-6 min-h-0">
            {activeTab === "general" && (
              <div className="max-w-xl mx-auto space-y-4">
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                    Map
                  </h3>
                  <LobbyPickerDropdown
                    items={mapItems}
                    selectedId={settings.mapId}
                    onSelect={handleMapSelect}
                    trigger={({ open, toggle }) => (
                      <button
                        type="button"
                        onClick={toggle}
                        className="w-full flex items-center justify-between py-2 px-3 bg-black/40 rounded-lg border border-space-blue-600/50 text-white text-sm font-medium transition-colors cursor-pointer hover:border-space-blue-400"
                      >
                        <span>{currentMapName}</span>
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

                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                    Lobby
                  </h3>
                  <label className="flex items-center justify-between gap-3 py-2 px-1">
                    <span className="text-white text-sm font-medium flex items-center gap-2">
                      Max Players
                      <InfoTooltip size="small">
                        Cannot be reduced below the number of players who have already joined.
                      </InfoTooltip>
                    </span>
                    <input
                      type="number"
                      min={minPlayersAllowed}
                      max={10}
                      value={settings.maxPlayers}
                      onChange={(e) => handleMaxPlayersChange(parseInt(e.target.value, 10) || 0)}
                      className="w-16 bg-black/50 border border-white/20 rounded-lg py-1 px-2 text-white text-sm text-center outline-none focus:border-white/60 transition-colors"
                    />
                  </label>
                </div>

                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                    Toggles
                  </h3>
                  <ToggleRow
                    label="Development Mode"
                    checked={settings.developmentMode}
                    onChange={(v) => dispatch({ developmentMode: v })}
                    tooltip="Enable admin commands for debugging and testing."
                  />
                  <ToggleRow
                    label="Demo Game"
                    checked={settings.demoGame}
                    onChange={(v) => dispatch({ demoGame: v })}
                    tooltip="Players configure their corporation, starting cards, resources and production in the lobby before starting."
                  />
                  <ToggleRow
                    label="Allow Random Buy"
                    checked={settings.allowRandomBuy}
                    onChange={(v) => dispatch({ allowRandomBuy: v })}
                    tooltip="When selecting 0 cards to buy at the end of a generation, players can choose to buy a single random card instead."
                  />
                </div>
              </div>
            )}

            {activeTab === "cardPacks" && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3 max-w-3xl mx-auto">
                {CARD_PACKS.map((pack) => {
                  const isSelected = (settings.cardPacks || []).includes(pack.id);
                  const isLocked = !!pack.lockedOn;
                  return (
                    <PackBox
                      key={pack.id}
                      label={pack.label}
                      description={pack.description}
                      cardCount={pack.cardCount}
                      wip={pack.wip}
                      selected={isSelected || isLocked}
                      locked={isLocked}
                      onToggle={() => togglePack(pack.id)}
                    />
                  );
                })}
                <PackBox
                  key={VENUS_PACK.id}
                  label={VENUS_PACK.label}
                  cardCount={VENUS_PACK.cardCount}
                  description={VENUS_PACK.description}
                  selected={settings.venusNextEnabled}
                  onToggle={toggleVenus}
                />
              </div>
            )}

            {activeTab === "ai" && (
              <div className="max-w-xl mx-auto space-y-4">
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                    Claude API Token
                  </h3>
                  <p className="text-white/70 text-sm mb-3">
                    {settings.hasClaudeApiKey
                      ? "A Claude token is configured. Paste a new one to replace it, or clear to disable bots."
                      : "Paste a Claude API token to enable bot players. Bots use this token for every action they take."}
                  </p>
                  <div className="flex gap-2">
                    <input
                      type="password"
                      value={claudeToken}
                      onChange={(e) => setClaudeToken(e.target.value)}
                      placeholder="sk-ant-..."
                      spellCheck={false}
                      autoComplete="off"
                      className="flex-1 bg-black/50 border border-white/20 rounded-lg py-2 px-3 text-white text-sm outline-none focus:border-white/60 placeholder:text-white/30"
                    />
                    <GameButton
                      buttonType="primary"
                      size="sm"
                      onClick={handleSubmitToken}
                      disabled={!claudeToken.trim() || tokenStatus === "saving"}
                    >
                      {tokenStatus === "saved" ? "Saved" : "Save"}
                    </GameButton>
                    {settings.hasClaudeApiKey && (
                      <GameButton
                        buttonType="secondary"
                        variant="error"
                        size="sm"
                        onClick={handleClearToken}
                      >
                        Clear
                      </GameButton>
                    )}
                  </div>
                </div>
                <div className="bg-yellow-900/30 border border-yellow-600/40 rounded-lg p-3 text-yellow-200/90 text-xs leading-relaxed">
                  Bots consume Anthropic API tokens that bill against your account. Bots will stop
                  acting if your usage limit is hit and may behave unpredictably. Terraforming Mars
                  CE is not responsible for charges incurred while running bots.
                </div>
              </div>
            )}
          </div>
        </div>

        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="text-white/60 text-sm">All changes save automatically.</div>
          <GameButton buttonType="primary" size="md" onClick={onClose}>
            Close
          </GameButton>
        </div>
      </div>
    </div>
  );
};

interface ToggleRowProps {
  label: string;
  checked: boolean;
  onChange: (next: boolean) => void;
  tooltip?: string;
}

const ToggleRow: React.FC<ToggleRowProps> = ({ label, checked, onChange, tooltip }) => (
  <label className="flex items-center gap-3 cursor-pointer py-2 px-1 rounded hover:bg-white/5 transition-all duration-200">
    <input
      type="checkbox"
      checked={checked}
      onChange={(e) => onChange(e.target.checked)}
      className="w-[18px] h-[18px] accent-space-blue-solid cursor-pointer m-0"
    />
    <span className="text-white text-sm font-medium leading-none flex items-center gap-2">
      {label}
      {tooltip ? <InfoTooltip size="small">{tooltip}</InfoTooltip> : null}
    </span>
  </label>
);

interface PackBoxProps {
  label: string;
  description: string;
  cardCount?: string;
  wip?: boolean;
  selected: boolean;
  locked?: boolean;
  onToggle: () => void;
}

const PackBox: React.FC<PackBoxProps> = ({
  label,
  description,
  cardCount,
  wip,
  selected,
  locked,
  onToggle,
}) => {
  const interactive = !locked;
  const containerClass = selected
    ? "border-space-blue-400 bg-space-blue-900/30 shadow-[0_0_24px_rgba(96,165,250,0.25)]"
    : "border-space-blue-600/50 bg-black/40 shadow-none";
  const titleClass = selected ? "text-white" : "text-white/60";
  const descClass = selected ? "text-white/80" : "text-white/50";
  const metaClass = selected ? "text-white/60" : "text-white/35";
  return (
    <button
      type="button"
      onClick={() => {
        if (interactive) {
          onToggle();
        }
      }}
      disabled={!interactive}
      className={`text-left rounded-xl p-4 border transition-[background-color,border-color,box-shadow,color] duration-300 ease-out ${containerClass} ${
        interactive ? "cursor-pointer hover:border-space-blue-400" : "cursor-default"
      }`}
    >
      <span
        className={`font-orbitron ${titleClass} text-base font-semibold tracking-wide flex items-center gap-2 mb-2 transition-colors duration-300 ease-out`}
      >
        {label}
        {wip ? (
          <span className="text-[10px] font-orbitron font-bold text-yellow-400/80 uppercase tracking-wider">
            WIP
          </span>
        ) : null}
      </span>
      <div className={`${descClass} text-xs leading-snug transition-colors duration-300 ease-out`}>
        {description}
      </div>
      <div
        className={`mt-2 flex gap-3 text-[11px] ${metaClass} transition-colors duration-300 ease-out`}
      >
        {cardCount ? <span>{cardCount}</span> : null}
        {locked ? <span>Required</span> : null}
      </div>
    </button>
  );
};

export default LobbySettingsOverlay;
