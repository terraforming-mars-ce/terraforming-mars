import React, { useState, useEffect, useMemo } from "react";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import { apiService } from "../../../services/apiService.ts";
import {
  GameDto,
  AdminCommandRequest,
  AdminCommandTypeGiveCard,
  AdminCommandTypeSetPhase,
  AdminCommandTypeSetResources,
  AdminCommandTypeSetProduction,
  AdminCommandTypeSetGlobalParams,
  AdminCommandTypeStartTileSelection,
  AdminCommandTypeSetTR,
  GiveCardAdminCommand,
  SetPhaseAdminCommand,
  SetResourcesAdminCommand,
  SetProductionAdminCommand,
  SetGlobalParamsAdminCommand,
  StartTileSelectionAdminCommand,
  SetTRAdminCommand,
  GamePhaseWaitingForGameStart,
  GamePhaseStartingCardSelection,
  GamePhaseAction,
  GamePhaseProductionAndCardDraw,
  GamePhaseComplete,
  CardDto,
  CardTypeCorporation,
} from "../../../types/generated/api-types.ts";

interface AdminCommandPanelProps {
  gameState: GameDto;
  onClose?: () => void;
  onOpenTilePlacer?: (playerId: string) => void;
}

// Global parameters min/max bounds
const GLOBAL_PARAM_BOUNDS = {
  temperature: { min: -30, max: 8 },
  oxygen: { min: 0, max: 14 },
  oceans: { min: 0, max: 9 },
} as const;

const AdminCommandPanel: React.FC<AdminCommandPanelProps> = ({
  gameState,
  onClose,
  onOpenTilePlacer,
}) => {
  const [selectedCommand, setSelectedCommand] = useState<string>("");
  const [validationErrors, setValidationErrors] = useState<Record<string, boolean>>({});

  // Shared styling functions

  const getInputStyle = (hasError: boolean = false, disabled: boolean = false) => ({
    width: "100%",
    padding: "6px 10px",
    marginTop: "2px",
    background: disabled ? "rgba(0, 0, 0, 0.4)" : "rgba(0, 0, 0, 0.8)",
    border: hasError ? "1px solid #ff4444" : "1px solid rgba(155, 89, 182, 0.3)",
    borderRadius: "4px",
    color: disabled ? "#666" : "white",
    fontSize: "12px",
    outline: "none",
    boxShadow: hasError ? "0 0 0 2px rgba(255, 68, 68, 0.2)" : "none",
    cursor: disabled ? "not-allowed" : "text",
  });

  const getSelectStyle = (hasError: boolean = false, customWidth?: string) => ({
    width: customWidth || "200px",
    maxWidth: "100%",
    padding: "6px 10px",
    marginTop: "2px",
    background: "rgba(0, 0, 0, 0.8)",
    border: hasError ? "1px solid #ff4444" : "1px solid rgba(155, 89, 182, 0.3)",
    borderRadius: "4px",
    color: "white",
    fontSize: "12px",
    outline: "none",
    cursor: "pointer",
    appearance: "none" as const,
    backgroundImage: `url("data:image/svg+xml;charset=US-ASCII,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 4 5'><path fill='%23abb2bf' d='M2 0L0 2h4zm0 5L0 3h4z'/></svg>")`,
    backgroundRepeat: "no-repeat",
    backgroundPosition: "right 6px center",
    backgroundSize: "10px",
    paddingRight: "28px",
    boxShadow: hasError ? "0 0 0 2px rgba(255, 68, 68, 0.2)" : "none",
  });

  const buttonStyle = {
    padding: "8px 16px",
    background: "linear-gradient(135deg, rgba(155, 89, 182, 0.8), rgba(155, 89, 182, 0.6))",
    border: "1px solid rgba(155, 89, 182, 0.5)",
    borderRadius: "6px",
    color: "white",
    fontSize: "12px",
    cursor: "pointer",
    transition: "all 0.2s ease",
    fontWeight: "500" as const,
  };

  const smallButtonStyle = {
    padding: "3px 8px",
    background: "rgba(155, 89, 182, 0.3)",
    border: "1px solid rgba(155, 89, 182, 0.4)",
    borderRadius: "4px",
    color: "white",
    fontSize: "10px",
    cursor: "pointer",
    transition: "all 0.2s ease",
    fontWeight: "500" as const,
  };
  const [giveCardForm, setGiveCardForm] = useState({
    playerId: "",
    cardId: "",
  });
  const [setPhaseForm, setSetPhaseForm] = useState({ phase: "" });
  const [resourcesForm, setResourcesForm] = useState({
    playerId: "",
    credit: "",
    steel: "",
    titanium: "",
    plant: "",
    energy: "",
    heat: "",
  });
  const [productionForm, setProductionForm] = useState({
    playerId: "",
    credit: "",
    steel: "",
    titanium: "",
    plant: "",
    energy: "",
    heat: "",
  });
  const [globalParamsForm, setGlobalParamsForm] = useState({
    temperature: gameState.globalParameters.temperature.toString(),
    oxygen: gameState.globalParameters.oxygen.toString(),
    oceans: gameState.globalParameters.oceans.toString(),
  });
  const [tileSelectionForm, setTileSelectionForm] = useState({
    playerId: "",
    tileType: "",
  });
  const [setCorporationForm, setSetCorporationForm] = useState({
    playerId: "",
    corporationId: "",
  });
  const [setTRForm, setSetTRForm] = useState({
    playerId: "",
    terraformRating: "",
  });

  // Card data cache for autocomplete
  const [allCards, setAllCards] = useState<CardDto[]>([]);
  const [cardsLoading, setCardsLoading] = useState(false);

  // Autocomplete state
  const [giveCardQuery, setGiveCardQuery] = useState("");
  const [corporationQuery, setCorporationQuery] = useState("");
  const [showGiveCardDropdown, setShowGiveCardDropdown] = useState(false);
  const [showCorporationDropdown, setShowCorporationDropdown] = useState(false);

  const allPlayers = [gameState.currentPlayer, ...gameState.otherPlayers];
  const defaultPlayerId = allPlayers[0]?.id || "";

  // Update forms when player selection changes
  useEffect(() => {
    if (resourcesForm.playerId) {
      const selectedPlayer = allPlayers.find((p) => p.id === resourcesForm.playerId);
      if (selectedPlayer && selectedPlayer.resources) {
        setResourcesForm((prev) => ({
          ...prev,
          credit: (selectedPlayer.resources.credits || 0).toString(),
          steel: (selectedPlayer.resources.steel || 0).toString(),
          titanium: (selectedPlayer.resources.titanium || 0).toString(),
          plant: (selectedPlayer.resources.plants || 0).toString(),
          energy: (selectedPlayer.resources.energy || 0).toString(),
          heat: (selectedPlayer.resources.heat || 0).toString(),
        }));
      }
    } else {
      // Reset all fields when no player is selected
      setResourcesForm((prev) => ({
        ...prev,
        credit: "",
        steel: "",
        titanium: "",
        plant: "",
        energy: "",
        heat: "",
      }));
    }
  }, [resourcesForm.playerId]);

  useEffect(() => {
    if (productionForm.playerId) {
      const selectedPlayer = allPlayers.find((p) => p.id === productionForm.playerId);
      if (selectedPlayer && selectedPlayer.production) {
        setProductionForm((prev) => ({
          ...prev,
          credit: (selectedPlayer.production.credits || 0).toString(),
          steel: (selectedPlayer.production.steel || 0).toString(),
          titanium: (selectedPlayer.production.titanium || 0).toString(),
          plant: (selectedPlayer.production.plants || 0).toString(),
          energy: (selectedPlayer.production.energy || 0).toString(),
          heat: (selectedPlayer.production.heat || 0).toString(),
        }));
      }
    } else {
      // Reset all fields when no player is selected
      setProductionForm((prev) => ({
        ...prev,
        credit: "",
        steel: "",
        titanium: "",
        plant: "",
        energy: "",
        heat: "",
      }));
    }
  }, [productionForm.playerId]);

  // Update TR form when player selection changes
  useEffect(() => {
    if (setTRForm.playerId) {
      const selectedPlayer = allPlayers.find((p) => p.id === setTRForm.playerId);
      if (selectedPlayer) {
        setSetTRForm((prev) => ({
          ...prev,
          terraformRating: (selectedPlayer.terraformRating || 20).toString(),
        }));
      }
    } else {
      setSetTRForm((prev) => ({
        ...prev,
        terraformRating: "",
      }));
    }
  }, [setTRForm.playerId]);

  // Update global parameters form when game state changes
  useEffect(() => {
    setGlobalParamsForm({
      temperature: gameState.globalParameters.temperature.toString(),
      oxygen: gameState.globalParameters.oxygen.toString(),
      oceans: gameState.globalParameters.oceans.toString(),
    });
  }, [gameState.globalParameters]);

  // Load all cards on mount for autocomplete
  useEffect(() => {
    const loadCards = async () => {
      if (allCards.length > 0 || cardsLoading) return;
      setCardsLoading(true);
      try {
        const response = await apiService.listCards(0, 10000);
        setAllCards(response.cards);
      } catch (error) {
        console.error("Failed to load cards for autocomplete:", error);
      } finally {
        setCardsLoading(false);
      }
    };
    void loadCards();
  }, []);

  // Initialize player selection to first player for player-related forms
  useEffect(() => {
    if (defaultPlayerId) {
      if (!giveCardForm.playerId) {
        setGiveCardForm((prev) => ({ ...prev, playerId: defaultPlayerId }));
      }
      if (!resourcesForm.playerId) {
        setResourcesForm((prev) => ({ ...prev, playerId: defaultPlayerId }));
      }
      if (!productionForm.playerId) {
        setProductionForm((prev) => ({ ...prev, playerId: defaultPlayerId }));
      }
      if (!tileSelectionForm.playerId) {
        setTileSelectionForm((prev) => ({
          ...prev,
          playerId: defaultPlayerId,
        }));
      }
      if (!setCorporationForm.playerId) {
        setSetCorporationForm((prev) => ({
          ...prev,
          playerId: defaultPlayerId,
        }));
      }
      if (!setTRForm.playerId) {
        setSetTRForm((prev) => ({
          ...prev,
          playerId: defaultPlayerId,
        }));
      }
    }
  }, [defaultPlayerId]);

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as Element;
      if (!target.closest(".card-autocomplete-container")) {
        setShowGiveCardDropdown(false);
        setShowCorporationDropdown(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Parse search query for prefixes like tag:space or behavior:discount
  const parseSearchQuery = (rawQuery: string) => {
    const query = rawQuery.toLowerCase().trim();
    if (query.startsWith("tag:")) {
      return { type: "tag" as const, value: query.slice(4).trim() };
    }
    if (query.startsWith("behavior:") || query.startsWith("b:")) {
      const prefix = query.startsWith("behavior:") ? 9 : 2;
      return { type: "behavior" as const, value: query.slice(prefix).trim() };
    }
    if (query.startsWith("type:") || query.startsWith("t:")) {
      const prefix = query.startsWith("type:") ? 5 : 2;
      return { type: "cardType" as const, value: query.slice(prefix).trim() };
    }
    return { type: "text" as const, value: query };
  };

  // Filter cards based on parsed query
  const filterCardsByQuery = (cards: CardDto[], rawQuery: string) => {
    const parsed = parseSearchQuery(rawQuery);
    if (!parsed.value) return [];

    return cards.filter((card) => {
      switch (parsed.type) {
        case "tag":
          return card.tags?.some((tag) => tag.toLowerCase().includes(parsed.value));
        case "behavior":
          return card.behaviors?.some((behavior) => {
            const inputMatch = behavior.inputs?.some((input) =>
              input.type.toLowerCase().includes(parsed.value),
            );
            const outputMatch = behavior.outputs?.some((output) =>
              output.type.toLowerCase().includes(parsed.value),
            );
            return inputMatch || outputMatch;
          });
        case "cardType":
          return card.type.toLowerCase().includes(parsed.value);
        case "text":
        default:
          return (
            card.id.toLowerCase().includes(parsed.value) ||
            card.name.toLowerCase().includes(parsed.value)
          );
      }
    });
  };

  // Filter cards for autocomplete
  const filteredCards = useMemo(() => {
    if (!giveCardQuery.trim()) return [];
    return filterCardsByQuery(allCards, giveCardQuery).slice(0, 3);
  }, [allCards, giveCardQuery]);

  const filteredCorporations = useMemo(() => {
    const corporations = allCards.filter((card) => card.type === CardTypeCorporation);
    if (!corporationQuery.trim()) return [];
    return filterCardsByQuery(corporations, corporationQuery).slice(0, 3);
  }, [allCards, corporationQuery]);

  // Helper to filter numeric input - allows digits, minus sign, and empty string
  const filterNumericInput = (value: string): string => {
    // Allow empty, minus sign, or digits with optional leading minus
    return value.replace(/[^0-9-]/g, "").replace(/(?!^)-/g, "");
  };

  // Helper to clamp and validate numeric value on blur
  const clampValue = (value: string, min: number, max: number, defaultVal: number): string => {
    if (value === "" || value === "-") return defaultVal.toString();
    const num = parseInt(value, 10);
    if (isNaN(num)) return defaultVal.toString();
    return Math.max(min, Math.min(max, num)).toString();
  };

  // Keyboard event handlers for Enter key support
  const handleSetPhaseKeyDown = (e: React.KeyboardEvent<HTMLSelectElement>) => {
    if (e.key === "Enter") {
      void handleSetPhase();
    }
  };

  const handleResourcesKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleSetResources();
    }
  };

  const handleProductionKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleSetProduction();
    }
  };

  const handleGlobalParamsKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleSetGlobalParams();
    }
  };

  const sendAdminCommand = async (commandType: string, payload: any) => {
    const adminRequest: AdminCommandRequest = {
      commandType: commandType as any,
      payload: payload,
    };

    try {
      await globalWebSocketManager.sendAdminCommand(adminRequest);
    } catch (error) {
      console.error("❌ Failed to send admin command:", error);
    }
  };

  const handleGiveCard = async () => {
    const errors: Record<string, boolean> = {};

    if (!giveCardForm.playerId) errors.giveCardPlayerId = true;
    if (!giveCardForm.cardId) errors.giveCardCardId = true;

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      // Clear errors after 3 seconds
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: GiveCardAdminCommand = {
      playerId: giveCardForm.playerId,
      cardId: giveCardForm.cardId,
    };

    await sendAdminCommand(AdminCommandTypeGiveCard, command);
    // Keep player selected, only clear card ID for next card
    setGiveCardForm({ ...giveCardForm, cardId: "" });
    setGiveCardQuery("");
  };

  const handleSetPhase = async () => {
    const errors: Record<string, boolean> = {};

    if (!setPhaseForm.phase) errors.setPhase = true;

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: SetPhaseAdminCommand = {
      phase: setPhaseForm.phase,
    };

    await sendAdminCommand(AdminCommandTypeSetPhase, command);
  };

  const parseStringToNumber = (value: string): number => {
    if (value === "" || value === undefined || value === null) {
      return 0;
    }
    const parsed = parseInt(value, 10);
    return isNaN(parsed) ? 0 : Math.max(0, parsed); // Ensure non-negative
  };

  const handleSetResources = async () => {
    const errors: Record<string, boolean> = {};

    if (!resourcesForm.playerId) errors.setResourcesPlayerId = true;

    // Validate that all resource values are valid numbers
    const resourceFields = ["credit", "steel", "titanium", "plant", "energy", "heat"];
    for (const field of resourceFields) {
      const value = resourcesForm[field as keyof typeof resourcesForm] as string;
      if (value !== "" && (isNaN(parseInt(value, 10)) || parseInt(value, 10) < 0)) {
        errors[`setResources${field.charAt(0).toUpperCase() + field.slice(1)}`] = true;
      }
    }

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: SetResourcesAdminCommand = {
      playerId: resourcesForm.playerId,
      resources: {
        credits: parseStringToNumber(resourcesForm.credit),
        steel: parseStringToNumber(resourcesForm.steel),
        titanium: parseStringToNumber(resourcesForm.titanium),
        plants: parseStringToNumber(resourcesForm.plant),
        energy: parseStringToNumber(resourcesForm.energy),
        heat: parseStringToNumber(resourcesForm.heat),
      },
    };

    await sendAdminCommand(AdminCommandTypeSetResources, command);
  };

  const handleSetProduction = async () => {
    const errors: Record<string, boolean> = {};

    if (!productionForm.playerId) errors.setProductionPlayerId = true;

    // Validate that all production values are valid numbers
    const productionFields = ["credit", "steel", "titanium", "plant", "energy", "heat"];
    for (const field of productionFields) {
      const value = productionForm[field as keyof typeof productionForm] as string;
      if (value !== "" && (isNaN(parseInt(value, 10)) || parseInt(value, 10) < 0)) {
        errors[`setProduction${field.charAt(0).toUpperCase() + field.slice(1)}`] = true;
      }
    }

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: SetProductionAdminCommand = {
      playerId: productionForm.playerId,
      production: {
        credits: parseStringToNumber(productionForm.credit),
        steel: parseStringToNumber(productionForm.steel),
        titanium: parseStringToNumber(productionForm.titanium),
        plants: parseStringToNumber(productionForm.plant),
        energy: parseStringToNumber(productionForm.energy),
        heat: parseStringToNumber(productionForm.heat),
      },
    };

    await sendAdminCommand(AdminCommandTypeSetProduction, command);
  };

  const handleSetGlobalParams = async () => {
    const command: SetGlobalParamsAdminCommand = {
      globalParameters: {
        temperature: parseInt(globalParamsForm.temperature, 10) || -30,
        oxygen: parseInt(globalParamsForm.oxygen, 10) || 0,
        oceans: parseInt(globalParamsForm.oceans, 10) || 0,
      },
    };

    await sendAdminCommand(AdminCommandTypeSetGlobalParams, command);
  };

  // Global params min/max helpers
  const setGlobalParamMin = (param: keyof typeof GLOBAL_PARAM_BOUNDS) => {
    setGlobalParamsForm((prev) => ({
      ...prev,
      [param]: GLOBAL_PARAM_BOUNDS[param].min.toString(),
    }));
  };

  const setGlobalParamMax = (param: keyof typeof GLOBAL_PARAM_BOUNDS) => {
    setGlobalParamsForm((prev) => ({
      ...prev,
      [param]: GLOBAL_PARAM_BOUNDS[param].max.toString(),
    }));
  };

  const setAllGlobalParamsMin = () => {
    setGlobalParamsForm({
      temperature: GLOBAL_PARAM_BOUNDS.temperature.min.toString(),
      oxygen: GLOBAL_PARAM_BOUNDS.oxygen.min.toString(),
      oceans: GLOBAL_PARAM_BOUNDS.oceans.min.toString(),
    });
  };

  const setAllGlobalParamsMax = () => {
    setGlobalParamsForm({
      temperature: GLOBAL_PARAM_BOUNDS.temperature.max.toString(),
      oxygen: GLOBAL_PARAM_BOUNDS.oxygen.max.toString(),
      oceans: GLOBAL_PARAM_BOUNDS.oceans.max.toString(),
    });
  };

  const handleStartTileSelection = async () => {
    const errors: Record<string, boolean> = {};

    if (!tileSelectionForm.playerId) errors.tileSelectionPlayerId = true;
    if (!tileSelectionForm.tileType) errors.tileSelectionTileType = true;

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: StartTileSelectionAdminCommand = {
      playerId: tileSelectionForm.playerId,
      tileType: tileSelectionForm.tileType,
    };

    await sendAdminCommand(AdminCommandTypeStartTileSelection, command);
    // Keep player selected, only clear tile type
    setTileSelectionForm({ ...tileSelectionForm, tileType: "" });

    // Close the admin panel after starting tile selection
    if (onClose) {
      onClose();
    }
  };

  const handleSetCorporation = async () => {
    const errors: Record<string, boolean> = {};

    if (!setCorporationForm.playerId) errors.setCorporationPlayerId = true;
    if (!setCorporationForm.corporationId) errors.setCorporationCorporationId = true;

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command = {
      playerId: setCorporationForm.playerId,
      corporationId: setCorporationForm.corporationId,
    };

    await sendAdminCommand("set-corporation" as any, command);
    // Keep player selected, only clear corporation
    setSetCorporationForm({ ...setCorporationForm, corporationId: "" });
    setCorporationQuery("");
  };

  const handleTRKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleSetTR();
    }
  };

  const handleSetTR = async () => {
    const errors: Record<string, boolean> = {};

    if (!setTRForm.playerId) errors.setTRPlayerId = true;
    if (!setTRForm.terraformRating) errors.setTRValue = true;

    // Validate TR is a valid number in range 1-70
    const trValue = parseInt(setTRForm.terraformRating, 10);
    if (isNaN(trValue) || trValue < 1 || trValue > 70) {
      errors.setTRValue = true;
    }

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: SetTRAdminCommand = {
      playerId: setTRForm.playerId,
      terraformRating: trValue,
    };

    await sendAdminCommand(AdminCommandTypeSetTR, command);
  };

  const commandOptions = [
    { value: "give-card", label: "Give Card to Player" },
    { value: "set-phase", label: "Set Game Phase" },
    { value: "set-resources", label: "Set Player Resources" },
    { value: "set-production", label: "Set Player Production" },
    { value: "set-tr", label: "Set Player Terraform Rating" },
    { value: "set-global-params", label: "Set Global Parameters" },
    { value: "start-tile-selection", label: "Start Tile Selection (Demo)" },
    { value: "set-corporation", label: "Set Player Corporation" },
  ];

  const phaseOptions = [
    { value: GamePhaseWaitingForGameStart, label: "Waiting for Game Start" },
    { value: GamePhaseStartingCardSelection, label: "Starting Card Selection" },
    { value: GamePhaseAction, label: "Action Phase" },
    {
      value: GamePhaseProductionAndCardDraw,
      label: "Production and Card Draw",
    },
    { value: GamePhaseComplete, label: "Game Complete" },
  ];

  return (
    <div
      className="debug-content-area"
      style={{
        flex: 1,
        overflow: "visible",
        background: "rgba(0, 0, 0, 0.5)",
        padding: "12px",
        borderRadius: "4px",
        border: "1px solid #222",
      }}
    >
      <div style={{ marginBottom: "16px" }}>
        <label style={{ color: "#9b59b6", fontSize: "12px", fontWeight: "bold" }}>
          Select Admin Command:
        </label>
        <select
          value={selectedCommand}
          onChange={(e) => setSelectedCommand(e.target.value)}
          style={{
            ...getSelectStyle(false, "100%"),
            fontSize: "13px",
            padding: "8px 12px",
            borderRadius: "6px",
            backgroundPosition: "right 8px center",
            backgroundSize: "12px",
            paddingRight: "32px",
          }}
        >
          <option value="">Choose a command...</option>
          {commandOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      {selectedCommand === "give-card" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>Give Card to Player</h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={giveCardForm.playerId}
              onChange={(e) => setGiveCardForm({ ...giveCardForm, playerId: e.target.value })}
              style={getSelectStyle(validationErrors.giveCardPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div
            className="card-autocomplete-container"
            style={{ marginBottom: "8px", position: "relative" }}
          >
            <input
              type="text"
              placeholder="Search: name, tag:space, b:discount..."
              value={giveCardQuery}
              onChange={(e) => {
                setGiveCardQuery(e.target.value);
                setShowGiveCardDropdown(true);
                // Clear the actual cardId when user is typing
                if (giveCardForm.cardId) {
                  setGiveCardForm({ ...giveCardForm, cardId: "" });
                }
              }}
              onFocus={() => setShowGiveCardDropdown(true)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  // If there's exactly one match, select it
                  if (filteredCards.length === 1) {
                    setGiveCardForm({
                      ...giveCardForm,
                      cardId: filteredCards[0].id,
                    });
                    setGiveCardQuery(`${filteredCards[0].id} - ${filteredCards[0].name}`);
                    setShowGiveCardDropdown(false);
                  } else if (giveCardForm.cardId) {
                    void handleGiveCard();
                  }
                }
              }}
              style={{
                ...getInputStyle(validationErrors.giveCardCardId),
                width: "100%",
              }}
            />
            {showGiveCardDropdown && giveCardQuery.trim() && (
              <div
                style={{
                  position: "absolute",
                  top: "100%",
                  left: 0,
                  right: 0,
                  background: "rgba(0, 0, 0, 0.98)",
                  border: "1px solid rgba(155, 89, 182, 0.5)",
                  borderRadius: "4px",
                  marginTop: "2px",
                  zIndex: 9999,
                  boxShadow: "0 4px 12px rgba(0, 0, 0, 0.8)",
                }}
              >
                {filteredCards.length === 0 ? (
                  <div
                    style={{
                      padding: "8px 12px",
                      color: "#666",
                      fontSize: "12px",
                    }}
                  >
                    No results
                  </div>
                ) : (
                  filteredCards.map((card) => (
                    <div
                      key={card.id}
                      onClick={() => {
                        setGiveCardForm({ ...giveCardForm, cardId: card.id });
                        setGiveCardQuery(`${card.id} - ${card.name}`);
                        setShowGiveCardDropdown(false);
                      }}
                      style={{
                        padding: "8px 12px",
                        cursor: "pointer",
                        fontSize: "12px",
                        borderBottom: "1px solid rgba(155, 89, 182, 0.2)",
                        transition: "background 0.15s ease",
                      }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.background = "rgba(155, 89, 182, 0.2)")
                      }
                      onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                    >
                      <span style={{ color: "#9b59b6", fontWeight: "bold" }}>{card.id}</span>
                      <span style={{ color: "#abb2bf" }}> - {card.name}</span>
                    </div>
                  ))
                )}
              </div>
            )}
            {giveCardForm.cardId && (
              <div
                style={{
                  marginTop: "4px",
                  fontSize: "11px",
                  color: "#4caf50",
                }}
              >
                Selected: {giveCardForm.cardId}
              </div>
            )}
          </div>
          <button onClick={handleGiveCard} style={buttonStyle}>
            Give Card
          </button>
        </div>
      )}

      {selectedCommand === "set-phase" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>Set Game Phase</h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={setPhaseForm.phase}
              onChange={(e) => setSetPhaseForm({ phase: e.target.value })}
              onKeyDown={handleSetPhaseKeyDown}
              style={getSelectStyle(validationErrors.setPhase)}
            >
              <option value="">Select phase...</option>
              {phaseOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <button onClick={handleSetPhase} style={buttonStyle}>
            Set Phase
          </button>
        </div>
      )}

      {selectedCommand === "set-resources" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>Set Player Resources</h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={resourcesForm.playerId}
              onChange={(e) => setResourcesForm({ ...resourcesForm, playerId: e.target.value })}
              style={getSelectStyle(validationErrors.setResourcesPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "8px 16px",
              marginBottom: "8px",
              padding: "0 4px",
            }}
          >
            {["credit", "steel", "titanium", "plant", "energy", "heat"].map((resource) => (
              <div key={resource} style={{ minWidth: 0 }}>
                <label
                  style={{
                    color: "#abb2bf",
                    fontSize: "11px",
                    textTransform: "capitalize",
                    display: "block",
                    marginBottom: "4px",
                  }}
                >
                  {resource}:
                </label>
                <input
                  type="text"
                  value={resourcesForm[resource as keyof typeof resourcesForm] as string}
                  onChange={(e) =>
                    setResourcesForm({
                      ...resourcesForm,
                      [resource]: filterNumericInput(e.target.value),
                    })
                  }
                  onBlur={(e) => {
                    const val = e.target.value;
                    if (val === "" || val === "-") return;
                    const num = parseInt(val, 10);
                    if (isNaN(num) || num < 0) {
                      setResourcesForm({
                        ...resourcesForm,
                        [resource]: "0",
                      });
                    }
                  }}
                  onKeyDown={handleResourcesKeyDown}
                  disabled={!resourcesForm.playerId}
                  style={{
                    ...getInputStyle(false, !resourcesForm.playerId),
                    fontSize: "11px",
                    width: "100%",
                    minWidth: 0,
                    boxSizing: "border-box" as const,
                  }}
                />
              </div>
            ))}
          </div>
          <button onClick={handleSetResources} style={buttonStyle}>
            Set Resources
          </button>
        </div>
      )}

      {selectedCommand === "set-production" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>Set Player Production</h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={productionForm.playerId}
              onChange={(e) =>
                setProductionForm({
                  ...productionForm,
                  playerId: e.target.value,
                })
              }
              style={getSelectStyle(validationErrors.setProductionPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "8px 16px",
              marginBottom: "8px",
              padding: "0 4px",
            }}
          >
            {["credit", "steel", "titanium", "plant", "energy", "heat"].map((resource) => (
              <div key={resource} style={{ minWidth: 0 }}>
                <label
                  style={{
                    color: "#abb2bf",
                    fontSize: "11px",
                    textTransform: "capitalize",
                    display: "block",
                    marginBottom: "4px",
                  }}
                >
                  {resource}:
                </label>
                <input
                  type="text"
                  value={productionForm[resource as keyof typeof productionForm] as string}
                  onChange={(e) =>
                    setProductionForm({
                      ...productionForm,
                      [resource]: filterNumericInput(e.target.value),
                    })
                  }
                  onBlur={(e) => {
                    const val = e.target.value;
                    if (val === "" || val === "-") return;
                    const num = parseInt(val, 10);
                    // Production can be negative (credits can go to -5)
                    if (isNaN(num)) {
                      setProductionForm({
                        ...productionForm,
                        [resource]: "0",
                      });
                    }
                  }}
                  onKeyDown={handleProductionKeyDown}
                  disabled={!productionForm.playerId}
                  style={{
                    ...getInputStyle(false, !productionForm.playerId),
                    fontSize: "11px",
                    width: "100%",
                    minWidth: 0,
                    boxSizing: "border-box" as const,
                  }}
                />
              </div>
            ))}
          </div>
          <button onClick={handleSetProduction} style={buttonStyle}>
            Set Production
          </button>
        </div>
      )}

      {selectedCommand === "set-tr" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>Set Player Terraform Rating</h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={setTRForm.playerId}
              onChange={(e) => setSetTRForm({ ...setTRForm, playerId: e.target.value })}
              style={getSelectStyle(validationErrors.setTRPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div style={{ marginBottom: "8px" }}>
            <label
              style={{
                color: "#abb2bf",
                fontSize: "11px",
                display: "block",
                marginBottom: "4px",
              }}
            >
              Terraform Rating (1-70):
            </label>
            <input
              type="text"
              value={setTRForm.terraformRating}
              onChange={(e) =>
                setSetTRForm({
                  ...setTRForm,
                  terraformRating: filterNumericInput(e.target.value),
                })
              }
              onBlur={(e) => {
                const val = e.target.value;
                if (val === "" || val === "-") return;
                const num = parseInt(val, 10);
                if (isNaN(num) || num < 1) {
                  setSetTRForm({ ...setTRForm, terraformRating: "1" });
                } else if (num > 70) {
                  setSetTRForm({ ...setTRForm, terraformRating: "70" });
                }
              }}
              onKeyDown={handleTRKeyDown}
              disabled={!setTRForm.playerId}
              style={{
                ...getInputStyle(validationErrors.setTRValue, !setTRForm.playerId),
                width: "100px",
              }}
            />
          </div>
          <button onClick={handleSetTR} style={buttonStyle}>
            Set TR
          </button>
        </div>
      )}

      {selectedCommand === "set-global-params" && (
        <div style={{ marginBottom: "16px" }}>
          <h4
            style={{
              color: "#9b59b6",
              margin: "0 0 12px 0",
              textAlign: "center",
            }}
          >
            Set Global Parameters
          </h4>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              gap: "8px",
              marginBottom: "8px",
            }}
          >
            <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <label
                style={{
                  color: "#abb2bf",
                  fontSize: "12px",
                  minWidth: "140px",
                }}
              >
                Temperature (-30 to +8°C):
              </label>
              <input
                type="text"
                value={globalParamsForm.temperature}
                onChange={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    temperature: filterNumericInput(e.target.value),
                  })
                }
                onBlur={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    temperature: clampValue(e.target.value, -30, 8, -30),
                  })
                }
                onKeyDown={handleGlobalParamsKeyDown}
                style={{
                  ...getInputStyle(),
                  width: "70px",
                }}
              />
              <button onClick={() => setGlobalParamMin("temperature")} style={smallButtonStyle}>
                Min
              </button>
              <button onClick={() => setGlobalParamMax("temperature")} style={smallButtonStyle}>
                Max
              </button>
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <label
                style={{
                  color: "#abb2bf",
                  fontSize: "12px",
                  minWidth: "140px",
                }}
              >
                Oxygen (0-14%):
              </label>
              <input
                type="text"
                value={globalParamsForm.oxygen}
                onChange={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    oxygen: filterNumericInput(e.target.value),
                  })
                }
                onBlur={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    oxygen: clampValue(e.target.value, 0, 14, 0),
                  })
                }
                onKeyDown={handleGlobalParamsKeyDown}
                style={{
                  ...getInputStyle(),
                  width: "70px",
                }}
              />
              <button onClick={() => setGlobalParamMin("oxygen")} style={smallButtonStyle}>
                Min
              </button>
              <button onClick={() => setGlobalParamMax("oxygen")} style={smallButtonStyle}>
                Max
              </button>
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <label
                style={{
                  color: "#abb2bf",
                  fontSize: "12px",
                  minWidth: "140px",
                }}
              >
                Oceans (0-9):
              </label>
              <input
                type="text"
                value={globalParamsForm.oceans}
                onChange={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    oceans: filterNumericInput(e.target.value),
                  })
                }
                onBlur={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    oceans: clampValue(e.target.value, 0, 9, 0),
                  })
                }
                onKeyDown={handleGlobalParamsKeyDown}
                style={{
                  ...getInputStyle(),
                  width: "70px",
                }}
              />
              <button onClick={() => setGlobalParamMin("oceans")} style={smallButtonStyle}>
                Min
              </button>
              <button onClick={() => setGlobalParamMax("oceans")} style={smallButtonStyle}>
                Max
              </button>
            </div>
          </div>
          <div
            style={{
              display: "flex",
              gap: "8px",
              marginBottom: "12px",
              justifyContent: "center",
            }}
          >
            <button onClick={setAllGlobalParamsMin} style={smallButtonStyle}>
              Min All
            </button>
            <button onClick={setAllGlobalParamsMax} style={smallButtonStyle}>
              Max All
            </button>
          </div>
          <div style={{ display: "flex", justifyContent: "center" }}>
            <button onClick={handleSetGlobalParams} style={buttonStyle}>
              Set Global Parameters
            </button>
          </div>
        </div>
      )}

      {selectedCommand === "start-tile-selection" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>Start Tile Selection (Demo)</h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={tileSelectionForm.playerId}
              onChange={(e) =>
                setTileSelectionForm({
                  ...tileSelectionForm,
                  playerId: e.target.value,
                })
              }
              style={getSelectStyle(validationErrors.tileSelectionPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={tileSelectionForm.tileType}
              onChange={(e) =>
                setTileSelectionForm({
                  ...tileSelectionForm,
                  tileType: e.target.value,
                })
              }
              style={getSelectStyle(validationErrors.tileSelectionTileType)}
            >
              <option value="">Select tile type...</option>
              <option value="city">City</option>
              <option value="greenery">Greenery</option>
              <option value="ocean">Ocean</option>
              <option value="volcano">Volcano</option>
            </select>
          </div>
          <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
            <button onClick={handleStartTileSelection} style={buttonStyle}>
              Start Tile Selection
            </button>
            <button
              onClick={() => {
                if (!tileSelectionForm.playerId) {
                  setValidationErrors({ tileSelectionPlayerId: true });
                  setTimeout(() => setValidationErrors({}), 3000);
                  return;
                }
                onOpenTilePlacer?.(tileSelectionForm.playerId);
              }}
              style={{
                ...buttonStyle,
                background:
                  "linear-gradient(135deg, rgba(200, 50, 50, 0.8), rgba(200, 50, 50, 0.6))",
                border: "1px solid rgba(200, 50, 50, 0.5)",
              }}
            >
              Tile Placer
            </button>
          </div>
          <div
            style={{
              marginTop: "8px",
              padding: "8px",
              background: "rgba(255, 193, 7, 0.1)",
              border: "1px solid rgba(255, 193, 7, 0.3)",
              borderRadius: "4px",
              fontSize: "11px",
              color: "#ffc107",
            }}
          >
            <strong>Demo:</strong> This will trigger tile selection for the chosen player. Available
            hexes will be highlighted on the Mars board. Click a highlighted hex to complete the
            tile placement.
          </div>
        </div>
      )}

      {selectedCommand === "set-corporation" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>Set Player Corporation</h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={setCorporationForm.playerId}
              onChange={(e) =>
                setSetCorporationForm({
                  ...setCorporationForm,
                  playerId: e.target.value,
                })
              }
              style={getSelectStyle(validationErrors.setCorporationPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div
            className="card-autocomplete-container"
            style={{ marginBottom: "8px", position: "relative" }}
          >
            <input
              type="text"
              placeholder="Search: name, b:discount..."
              value={corporationQuery}
              onChange={(e) => {
                setCorporationQuery(e.target.value);
                setShowCorporationDropdown(true);
                // Clear the actual corporationId when user is typing
                if (setCorporationForm.corporationId) {
                  setSetCorporationForm({
                    ...setCorporationForm,
                    corporationId: "",
                  });
                }
              }}
              onFocus={() => setShowCorporationDropdown(true)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  // If there's exactly one match, select it
                  if (filteredCorporations.length === 1) {
                    setSetCorporationForm({
                      ...setCorporationForm,
                      corporationId: filteredCorporations[0].id,
                    });
                    setCorporationQuery(
                      `${filteredCorporations[0].id} - ${filteredCorporations[0].name}`,
                    );
                    setShowCorporationDropdown(false);
                  } else if (setCorporationForm.corporationId) {
                    void handleSetCorporation();
                  }
                }
              }}
              style={{
                ...getInputStyle(validationErrors.setCorporationCorporationId),
                width: "100%",
              }}
            />
            {showCorporationDropdown && corporationQuery.trim() && (
              <div
                style={{
                  position: "absolute",
                  top: "100%",
                  left: 0,
                  right: 0,
                  background: "rgba(0, 0, 0, 0.98)",
                  border: "1px solid rgba(155, 89, 182, 0.5)",
                  borderRadius: "4px",
                  marginTop: "2px",
                  zIndex: 9999,
                  boxShadow: "0 4px 12px rgba(0, 0, 0, 0.8)",
                }}
              >
                {filteredCorporations.length === 0 ? (
                  <div
                    style={{
                      padding: "8px 12px",
                      color: "#666",
                      fontSize: "12px",
                    }}
                  >
                    No results
                  </div>
                ) : (
                  filteredCorporations.map((card) => (
                    <div
                      key={card.id}
                      onClick={() => {
                        setSetCorporationForm({
                          ...setCorporationForm,
                          corporationId: card.id,
                        });
                        setCorporationQuery(`${card.id} - ${card.name}`);
                        setShowCorporationDropdown(false);
                      }}
                      style={{
                        padding: "8px 12px",
                        cursor: "pointer",
                        fontSize: "12px",
                        borderBottom: "1px solid rgba(155, 89, 182, 0.2)",
                        transition: "background 0.15s ease",
                      }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.background = "rgba(155, 89, 182, 0.2)")
                      }
                      onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                    >
                      <span style={{ color: "#ffc107", fontWeight: "bold" }}>{card.id}</span>
                      <span style={{ color: "#abb2bf" }}> - {card.name}</span>
                    </div>
                  ))
                )}
              </div>
            )}
            {setCorporationForm.corporationId && (
              <div
                style={{
                  marginTop: "4px",
                  fontSize: "11px",
                  color: "#4caf50",
                }}
              >
                Selected: {setCorporationForm.corporationId}
              </div>
            )}
          </div>
          <button onClick={handleSetCorporation} style={buttonStyle}>
            Set Corporation
          </button>
          <div
            style={{
              marginTop: "8px",
              padding: "10px",
              background: "rgba(100, 100, 100, 0.15)",
              border: "1px solid rgba(150, 150, 150, 0.3)",
              borderRadius: "4px",
              fontSize: "11px",
              color: "#bbb",
              lineHeight: "1.5",
            }}
          >
            <div style={{ marginBottom: "6px", color: "#fff", fontWeight: "bold" }}>
              What happens:
            </div>
            <div style={{ color: "#ff6b6b", marginBottom: "4px" }}>
              <strong>Cleared:</strong> Old corp effects, actions, card storage, payment
              substitutes, value modifiers
            </div>
            <div style={{ color: "#4caf50", marginBottom: "4px" }}>
              <strong>Applied:</strong> New corp starting resources, production, effects, actions,
              payment bonuses
            </div>
            <div style={{ color: "#888" }}>
              <strong>Kept:</strong> Current resources, production, played cards, terraform rating
            </div>
          </div>
        </div>
      )}

      {!selectedCommand && (
        <div
          style={{
            color: "#666",
            textAlign: "center",
            padding: "20px",
            fontSize: "12px",
          }}
        >
          Select an admin command above to get started.
          <br />
          <br />
          Available commands:
          <ul style={{ textAlign: "left", marginTop: "12px" }}>
            <li>Give cards to players</li>
            <li>Change game phase</li>
            <li>Set player resources</li>
            <li>Set player production</li>
            <li>Set player terraform rating</li>
            <li>Modify global parameters</li>
            <li>Start tile selection (demo)</li>
            <li>Set player corporation</li>
          </ul>
        </div>
      )}
    </div>
  );
};

export default AdminCommandPanel;
