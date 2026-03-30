import React, { useState, useEffect, useCallback, useRef } from "react";
import { Z_INDEX } from "@/constants/zIndex.ts";
import {
  GameDto,
  CardDto,
  ResourcesDto,
  ProductionDto,
  GlobalParametersDto,
  ResourceTypeCredit,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlant,
  ResourceTypeEnergy,
  ResourceTypeHeat,
} from "@/types/generated/api-types.ts";
import { apiService } from "../../../services/apiService.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import GameIcon from "../display/GameIcon.tsx";
import GameCard from "../cards/GameCard.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
} from "./overlayStyles.ts";
import GameButton from "../buttons/GameButton.tsx";

interface DemoSetupOverlayProps {
  game: GameDto;
  playerId: string;
  isOpen: boolean;
  onClose: () => void;
}

const TEMP_MIN = -30;
const TEMP_MAX = 8;
const OXYGEN_MIN = 0;
const OXYGEN_MAX = 14;
const OCEANS_MIN = 0;
const OCEANS_MAX = 9;
const GENERATION_MIN = 1;
const GENERATION_MAX = 14;

// Stepper component with optional icon
interface StepperProps {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  step?: number;
  defaultValue?: number; // If provided, shows grayed hint until modified
  icon?: string; // Optional icon to display
  iconLabel?: string; // Optional text label instead of icon
}

const Stepper: React.FC<StepperProps> = ({
  value,
  onChange,
  min = 0,
  max = 999,
  step = 1,
  defaultValue,
  icon,
  iconLabel,
}) => {
  const [inputValue, setInputValue] = useState(value.toString());

  useEffect(() => {
    setInputValue(value.toString());
  }, [value]);

  const canDecrease = value > min;
  const canIncrease = value < max;

  const handleDecrease = () => {
    if (canDecrease) {
      const newValue = Math.max(min, value - step);
      onChange(newValue);
    }
  };

  const handleIncrease = () => {
    if (canIncrease) {
      const newValue = Math.min(max, value + step);
      onChange(newValue);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value;
    setInputValue(raw);

    // Allow typing negative for negative mins
    const filtered = raw.replace(/[^0-9-]/g, "").replace(/(?!^)-/g, "");
    const num = parseInt(filtered, 10);
    if (!isNaN(num)) {
      const clamped = Math.max(min, Math.min(max, num));
      onChange(clamped);
    }
  };

  const handleInputBlur = () => {
    // On blur, reset input display to actual value
    setInputValue(value.toString());
  };

  const handleInputFocus = (e: React.FocusEvent<HTMLInputElement>) => {
    e.target.select();
  };

  const dragRef = useRef<{ startX: number; startVal: number; dragging: boolean } | null>(null);

  const handleMouseDown = useCallback(
    (e: React.MouseEvent<HTMLInputElement>) => {
      e.preventDefault();
      const input = e.currentTarget;
      dragRef.current = { startX: e.clientX, startVal: value, dragging: false };

      const preventSelect = (ev: Event) => ev.preventDefault();
      document.addEventListener("selectstart", preventSelect);

      const onMouseMove = (me: MouseEvent) => {
        if (!dragRef.current) {
          return;
        }
        const dx = me.clientX - dragRef.current.startX;
        if (!dragRef.current.dragging && Math.abs(dx) < 4) {
          return;
        }
        if (!dragRef.current.dragging) {
          dragRef.current.dragging = true;
          document.body.style.cursor = "ew-resize";
          document.body.style.userSelect = "none";
          input.blur();
        }
        const steps = Math.round(dx / 8) * step;
        const newVal = Math.max(min, Math.min(max, dragRef.current.startVal + steps));
        onChange(newVal);
      };

      const onMouseUp = () => {
        if (!dragRef.current?.dragging) {
          input.focus();
          input.select();
        }
        dragRef.current = null;
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
        document.removeEventListener("selectstart", preventSelect);
        document.body.style.cursor = "";
        document.body.style.userSelect = "";
      };

      document.addEventListener("mousemove", onMouseMove);
      document.addEventListener("mouseup", onMouseUp);
    },
    [value, min, max, step, onChange],
  );

  const isAtDefault = defaultValue !== undefined && value === defaultValue;

  return (
    <div className="flex items-center gap-2 bg-black/30 rounded-lg p-2">
      {icon && <GameIcon iconType={icon} size="small" />}
      {iconLabel && <span className="text-white/60 text-xs font-medium w-6">{iconLabel}</span>}
      <div className="flex items-center gap-1">
        <button
          onClick={handleDecrease}
          disabled={!canDecrease}
          className={`w-6 h-6 text-sm rounded transition-all ${
            canDecrease
              ? "bg-black/40 text-white hover:bg-black/60 cursor-pointer"
              : "bg-black/30 text-white/30"
          }`}
        >
          -
        </button>
        <input
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          onBlur={handleInputBlur}
          onFocus={handleInputFocus}
          onMouseDown={handleMouseDown}
          className={`w-10 h-6 text-center font-medium text-sm bg-black/60 border border-white/30 rounded outline-none focus:border-white/60 ${
            isAtDefault ? "text-white/40" : "text-white"
          }`}
          style={{ cursor: "ew-resize" }}
        />
        <button
          onClick={handleIncrease}
          disabled={!canIncrease}
          className={`w-6 h-6 text-sm rounded transition-all ${
            canIncrease
              ? "bg-black/40 text-white hover:bg-black/60 cursor-pointer"
              : "bg-black/30 text-white/30"
          }`}
        >
          +
        </button>
      </div>
    </div>
  );
};

interface ResourceStepperProps {
  icon: string;
  value: number;
  onChange: (value: number) => void;
  min?: number;
  isProduction?: boolean;
}

const ResourceStepper: React.FC<ResourceStepperProps> = ({
  icon,
  value,
  onChange,
  min = 0,
  isProduction = false,
}) => {
  return (
    <Stepper
      value={value}
      onChange={onChange}
      min={min}
      defaultValue={0}
      icon={isProduction ? `${icon}-production` : icon}
    />
  );
};

type SidebarTab = "cards" | "resources" | "global";

const DemoSetupOverlay: React.FC<DemoSetupOverlayProps> = ({ game, playerId, isOpen, onClose }) => {
  const isHost = game.hostPlayerId === playerId;
  const hasPrelude = game.settings.cardPacks?.includes("prelude") || false;
  const [activeTab, setActiveTab] = useState<SidebarTab>("cards");

  // Global parameters (host only)
  const [globalParams, setGlobalParams] = useState<GlobalParametersDto>({
    temperature: game.settings?.temperature ?? TEMP_MIN,
    oxygen: game.settings?.oxygen ?? OXYGEN_MIN,
    oceans: game.settings?.oceans ?? OCEANS_MIN,
    maxOceans: 9,
    venus: 0,
    bonuses: [],
  });
  const [generation, setGeneration] = useState(game.settings?.generation ?? 1);

  // Player setup
  const [availableCorporations, setAvailableCorporations] = useState<CardDto[]>([]);
  const [availablePreludes, setAvailablePreludes] = useState<CardDto[]>([]);
  const [availableCards, setAvailableCards] = useState<CardDto[]>([]);
  const [selectedCorporationId, setSelectedCorporationId] = useState<string>(
    game.currentPlayer?.pendingDemoChoices?.corporationId || "",
  );
  const [selectedPreludeIds, setSelectedPreludeIds] = useState<string[]>(
    game.currentPlayer?.pendingDemoChoices?.preludeIds || [],
  );
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>(
    game.currentPlayer?.pendingDemoChoices?.cardIds || [],
  );
  const [cardSearchTerm, setCardSearchTerm] = useState("");
  const [corpSearchTerm, setCorpSearchTerm] = useState("");
  const [preludeSearchTerm, setPreludeSearchTerm] = useState("");

  // Resources
  const [resources, setResources] = useState<ResourcesDto>(
    game.currentPlayer?.pendingDemoChoices?.resources || {
      credits: 0,
      steel: 0,
      titanium: 0,
      plants: 0,
      energy: 0,
      heat: 0,
    },
  );

  // Production
  const [production, setProduction] = useState<ProductionDto>(
    game.currentPlayer?.pendingDemoChoices?.production || {
      credits: 0,
      steel: 0,
      titanium: 0,
      plants: 0,
      energy: 0,
      heat: 0,
    },
  );

  // Terraform rating
  const [terraformRating, setTerraformRating] = useState(
    game.currentPlayer?.pendingDemoChoices?.terraformRating || 20,
  );

  // Loading state
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Load corporations and cards on mount
  useEffect(() => {
    const loadCardsData = async () => {
      try {
        const response = await apiService.listCards(0, 1000);
        const corps: CardDto[] = [];
        const prels: CardDto[] = [];
        const projectCards: CardDto[] = [];
        for (const card of response.cards) {
          if (card.type === "corporation") {
            corps.push(card);
          } else if (card.type === "prelude") {
            prels.push(card);
          } else {
            projectCards.push(card);
          }
        }
        setAvailableCorporations(corps);
        setAvailablePreludes(prels);
        setAvailableCards(projectCards);
      } catch (err) {
        console.error("Failed to load cards:", err);
      }
    };

    if (isOpen) {
      void loadCardsData();
    }
  }, [isOpen]);

  // When a corporation is selected, apply its starting resources and production
  useEffect(() => {
    if (!selectedCorporationId) {
      return;
    }

    const corp = availableCorporations.find((c) => c.id === selectedCorporationId);
    if (corp) {
      if (corp.startingResources) {
        setResources({
          credits: corp.startingResources.credits ?? 0,
          steel: corp.startingResources.steel ?? 0,
          titanium: corp.startingResources.titanium ?? 0,
          plants: corp.startingResources.plants ?? 0,
          energy: corp.startingResources.energy ?? 0,
          heat: corp.startingResources.heat ?? 0,
        });
      }
      if (corp.startingProduction) {
        setProduction({
          credits: corp.startingProduction.credits ?? 0,
          steel: corp.startingProduction.steel ?? 0,
          titanium: corp.startingProduction.titanium ?? 0,
          plants: corp.startingProduction.plants ?? 0,
          energy: corp.startingProduction.energy ?? 0,
          heat: corp.startingProduction.heat ?? 0,
        });
      }
    }
  }, [selectedCorporationId, availableCorporations]);

  const toggleCardSelection = (cardId: string) => {
    setSelectedCardIds((prev) =>
      prev.includes(cardId) ? prev.filter((id) => id !== cardId) : [...prev, cardId],
    );
  };

  const togglePreludeSelection = (preludeId: string) => {
    setSelectedPreludeIds((prev) => {
      if (prev.includes(preludeId)) {
        return prev.filter((id) => id !== preludeId);
      }
      if (prev.length >= 2) {
        return prev;
      }
      return [...prev, preludeId];
    });
  };

  const filteredCards = availableCards.filter(
    (card) =>
      card.name.toLowerCase().includes(cardSearchTerm.toLowerCase()) ||
      card.id.toLowerCase().includes(cardSearchTerm.toLowerCase()),
  );

  const filteredCorporations = availableCorporations.filter(
    (corp) =>
      corp.name.toLowerCase().includes(corpSearchTerm.toLowerCase()) ||
      corp.id.toLowerCase().includes(corpSearchTerm.toLowerCase()),
  );

  const filteredPreludes = availablePreludes.filter(
    (prelude) =>
      prelude.name.toLowerCase().includes(preludeSearchTerm.toLowerCase()) ||
      prelude.id.toLowerCase().includes(preludeSearchTerm.toLowerCase()),
  );

  const selectedCorporation = availableCorporations.find((c) => c.id === selectedCorporationId);

  const canConfirm =
    selectedCorporationId !== "" && (!hasPrelude || selectedPreludeIds.length === 2);

  const handleConfirm = async () => {
    if (isSubmitting || !canConfirm) return;

    setIsSubmitting(true);
    try {
      await globalWebSocketManager.selectDemoChoices({
        corporationId: selectedCorporationId,
        preludeIds: selectedPreludeIds,
        cardIds: selectedCardIds,
        resources,
        production,
        terraformRating,
        globalParameters: isHost ? globalParams : undefined,
        generation: isHost ? generation : undefined,
      });
      onClose();
    } catch (err) {
      console.error("Failed to select demo choices:", err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const resourceTypes = [
    { key: "credits" as const, icon: ResourceTypeCredit },
    { key: "steel" as const, icon: ResourceTypeSteel },
    { key: "titanium" as const, icon: ResourceTypeTitanium },
    { key: "plants" as const, icon: ResourceTypePlant },
    { key: "energy" as const, icon: ResourceTypeEnergy },
    { key: "heat" as const, icon: ResourceTypeHeat },
  ];

  if (!isOpen) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 flex items-center justify-center bg-black/70 backdrop-blur-sm animate-[fadeIn_0.3s_ease]"
      style={{ zIndex: Z_INDEX.STANDARD_MODAL }}
    >
      <div className={`${OVERLAY_CONTAINER_CLASS} max-w-[1400px] h-[90vh]`}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Demo Game Setup</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Configure your starting setup.{" "}
            {isHost ? "As host, you can also set global parameters." : ""}
          </p>
        </div>

        {/* Content: Sidebar + Main */}
        <div className="flex-1 flex min-h-0">
          {/* Sidebar */}
          <div className="w-40 shrink-0 bg-black/30 border-r border-space-blue-600/50 flex flex-col py-2">
            <button
              onClick={() => setActiveTab("cards")}
              className={`text-left px-4 py-3 text-sm font-semibold uppercase tracking-wide transition-colors cursor-pointer ${
                activeTab === "cards"
                  ? "text-white bg-space-blue-600/30 border-r-2 border-space-blue-400"
                  : "text-white/50 hover:text-white/80 hover:bg-white/5"
              }`}
            >
              Cards
            </button>
            <button
              onClick={() => setActiveTab("resources")}
              className={`text-left px-4 py-3 text-sm font-semibold uppercase tracking-wide transition-colors cursor-pointer ${
                activeTab === "resources"
                  ? "text-white bg-space-blue-600/30 border-r-2 border-space-blue-400"
                  : "text-white/50 hover:text-white/80 hover:bg-white/5"
              }`}
            >
              Resources
            </button>
            <button
              onClick={() => {
                if (isHost) {
                  setActiveTab("global");
                }
              }}
              className={`text-left px-4 py-3 text-sm font-semibold uppercase tracking-wide transition-colors ${
                !isHost
                  ? "text-white/20 cursor-default"
                  : activeTab === "global"
                    ? "text-white bg-space-blue-600/30 border-r-2 border-space-blue-400 cursor-pointer"
                    : "text-white/50 hover:text-white/80 hover:bg-white/5 cursor-pointer"
              }`}
            >
              Global
            </button>
          </div>

          {/* Main content area */}
          <div className="flex-1 overflow-y-auto p-4 min-h-0">
            {/* Cards tab */}
            {activeTab === "cards" && (
              <div
                className="grid grid-cols-1 gap-4 h-full"
                style={{
                  gridTemplateColumns: hasPrelude ? "1.1fr 0.7fr 1.15fr" : "1.1fr 1.15fr",
                }}
              >
                {/* Corporation Selection */}
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3 flex flex-col min-h-0">
                  <h3 className="text-white font-semibold mb-2 uppercase tracking-wide text-xs shrink-0">
                    Corporation{" "}
                    <span className="text-white/50 font-normal normal-case">
                      ({selectedCorporationId ? "1 selected" : "None"})
                    </span>
                  </h3>
                  <input
                    type="text"
                    placeholder="Search corporations..."
                    value={corpSearchTerm}
                    onChange={(e) => setCorpSearchTerm(e.target.value)}
                    className="w-full bg-black/60 border border-space-blue-400/30 rounded-lg py-2 px-3 text-white text-sm outline-none focus:border-space-blue-400 mb-3 shrink-0 cursor-text"
                  />
                  <div className="flex flex-col gap-2 items-center flex-1 min-h-0 overflow-y-auto">
                    {filteredCorporations.map((corp) => (
                      <div key={corp.id} className="scale-90">
                        <CorporationCard
                          card={corp}
                          isSelected={selectedCorporationId === corp.id}
                          onSelect={() =>
                            setSelectedCorporationId(
                              selectedCorporationId === corp.id ? "" : corp.id,
                            )
                          }
                          showCheckbox
                          borderColor={getCorporationBorderColor(corp.name)}
                        />
                      </div>
                    ))}
                  </div>
                </div>

                {/* Prelude Selection (if enabled) */}
                {hasPrelude && (
                  <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3 flex flex-col min-h-0">
                    <h3 className="text-white font-semibold mb-2 uppercase tracking-wide text-xs shrink-0">
                      Prelude Cards{" "}
                      <span className="text-white/50 font-normal normal-case">
                        ({selectedPreludeIds.length}/2 selected)
                      </span>
                    </h3>
                    <input
                      type="text"
                      placeholder="Search preludes..."
                      value={preludeSearchTerm}
                      onChange={(e) => setPreludeSearchTerm(e.target.value)}
                      className="w-full bg-black/60 border border-space-blue-400/30 rounded-lg py-2 px-3 text-white text-sm outline-none focus:border-space-blue-400 mb-3 shrink-0 cursor-text"
                    />
                    <div className="flex flex-wrap gap-x-1 gap-y-2 justify-center content-start flex-1 min-h-0 overflow-y-auto">
                      {filteredPreludes.map((prelude) => (
                        <div key={prelude.id} className="scale-80">
                          <GameCard
                            card={prelude}
                            isSelected={selectedPreludeIds.includes(prelude.id)}
                            onSelect={() => togglePreludeSelection(prelude.id)}
                            animationDelay={0}
                            showCheckbox
                          />
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Starting Cards */}
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3 flex flex-col min-h-0">
                  <h3 className="text-white font-semibold mb-2 uppercase tracking-wide text-xs shrink-0">
                    Starting Cards{" "}
                    <span className="text-white/50 font-normal normal-case">
                      ({selectedCardIds.length} selected)
                    </span>
                  </h3>
                  <input
                    type="text"
                    placeholder="Search cards..."
                    value={cardSearchTerm}
                    onChange={(e) => setCardSearchTerm(e.target.value)}
                    className="w-full bg-black/60 border border-space-blue-400/30 rounded-lg py-2 px-3 text-white text-sm outline-none focus:border-space-blue-400 mb-3 shrink-0 cursor-text"
                  />
                  <div className="flex flex-wrap gap-x-1 gap-y-2 justify-center content-start flex-1 min-h-0 overflow-y-auto">
                    {filteredCards.slice(0, 50).map((card) => (
                      <div key={card.id} className="scale-80">
                        <GameCard
                          card={card}
                          isSelected={selectedCardIds.includes(card.id)}
                          onSelect={() => toggleCardSelection(card.id)}
                          animationDelay={0}
                          showCheckbox
                        />
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Resources tab */}
            {activeTab === "resources" && (
              <div className="max-w-lg mx-auto space-y-4">
                {/* Terraform Rating */}
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs text-center">
                    Terraform Rating
                  </h3>
                  <div className="flex justify-center">
                    <Stepper
                      value={terraformRating}
                      onChange={setTerraformRating}
                      min={0}
                      max={100}
                      defaultValue={20}
                      icon="tr"
                    />
                  </div>
                </div>

                {/* Resources */}
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                    Resources
                  </h3>
                  <div className="grid grid-cols-2 gap-2">
                    {resourceTypes.map(({ key, icon }) => (
                      <ResourceStepper
                        key={key}
                        icon={icon}
                        value={resources[key]}
                        onChange={(v) => setResources((p) => ({ ...p, [key]: v }))}
                      />
                    ))}
                  </div>
                </div>

                {/* Production */}
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                    Production
                  </h3>
                  <div className="grid grid-cols-2 gap-2">
                    {resourceTypes.map(({ key, icon }) => (
                      <ResourceStepper
                        key={key}
                        icon={icon}
                        value={production[key]}
                        onChange={(v) => setProduction((p) => ({ ...p, [key]: v }))}
                        min={key === "credits" ? -5 : 0}
                        isProduction
                      />
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Global tab (host only) */}
            {activeTab === "global" && isHost && (
              <div className="max-w-lg mx-auto space-y-4">
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs text-center">
                    Global Parameters
                  </h3>
                  <div className="grid grid-cols-2 gap-2">
                    <Stepper
                      value={globalParams.temperature}
                      onChange={(v) => setGlobalParams((p) => ({ ...p, temperature: v }))}
                      min={TEMP_MIN}
                      max={TEMP_MAX}
                      step={2}
                      defaultValue={TEMP_MIN}
                      icon="temperature"
                    />
                    <Stepper
                      value={globalParams.oxygen}
                      onChange={(v) => setGlobalParams((p) => ({ ...p, oxygen: v }))}
                      min={OXYGEN_MIN}
                      max={OXYGEN_MAX}
                      defaultValue={OXYGEN_MIN}
                      icon="oxygen"
                    />
                    <Stepper
                      value={globalParams.oceans}
                      onChange={(v) => setGlobalParams((p) => ({ ...p, oceans: v }))}
                      min={OCEANS_MIN}
                      max={OCEANS_MAX}
                      defaultValue={OCEANS_MIN}
                      icon="ocean"
                    />
                    <Stepper
                      value={generation}
                      onChange={setGeneration}
                      min={GENERATION_MIN}
                      max={GENERATION_MAX}
                      defaultValue={GENERATION_MIN}
                      iconLabel="Gen"
                    />
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="text-white/60 text-sm">
            {selectedCorporationId ? (
              <span>Corporation: {selectedCorporation?.name}</span>
            ) : (
              <span>No corporation selected</span>
            )}
            {selectedPreludeIds.length > 0 && (
              <span className="ml-4">Preludes: {selectedPreludeIds.length}</span>
            )}
            {selectedCardIds.length > 0 && (
              <span className="ml-4">Cards: {selectedCardIds.length}</span>
            )}
          </div>
          <div className="flex gap-3">
            <GameButton
              buttonType="secondary"
              size="md"
              onClick={onClose}
              className="whitespace-nowrap"
            >
              Cancel
            </GameButton>
            <GameButton
              size="lg"
              onClick={() => void handleConfirm()}
              disabled={isSubmitting || !canConfirm}
              className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
            >
              {isSubmitting ? "Confirming..." : "Confirm Setup"}
            </GameButton>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DemoSetupOverlay;
