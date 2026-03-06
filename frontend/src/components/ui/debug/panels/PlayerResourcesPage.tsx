import React, { useState, useEffect, useCallback, useRef } from "react";
import { globalWebSocketManager } from "../../../../services/globalWebSocketManager.ts";
import {
  GameDto,
  AdminCommandRequest,
  AdminCommandTypeSetResources,
  AdminCommandTypeSetProduction,
  AdminCommandTypeSetTR,
  SetResourcesAdminCommand,
  SetProductionAdminCommand,
  SetTRAdminCommand,
} from "../../../../types/generated/api-types.ts";
import PlayerSelector from "../PlayerSelector.tsx";
import GameIcon, { GameIconType } from "../../display/GameIcon.tsx";

interface PlayerResourcesPageProps {
  gameState: GameDto;
  selectedPlayerIds: string[];
  onPlayerChange: (ids: string[]) => void;
}

const RESOURCE_FIELDS = ["credit", "steel", "titanium", "plant", "energy", "heat"] as const;

type ResourceForm = Record<(typeof RESOURCE_FIELDS)[number], string>;

const RESOURCE_ICON_MAP: Record<string, GameIconType> = {
  credit: "credit",
  steel: "steel",
  titanium: "titanium",
  plant: "plant",
  energy: "energy",
  heat: "heat",
};

const secondaryButtonStyle = {
  padding: "6px 14px",
  background: "transparent",
  border: "1px solid rgba(59, 130, 246, 0.5)",
  borderRadius: "6px",
  color: "#93c5fd",
  fontSize: "11px",
  cursor: "pointer" as const,
  fontWeight: "500" as const,
};

const buttonStyle = {
  padding: "6px 14px",
  background: "linear-gradient(135deg, rgba(59, 130, 246, 0.8), rgba(59, 130, 246, 0.6))",
  border: "1px solid rgba(59, 130, 246, 0.5)",
  borderRadius: "6px",
  color: "white",
  fontSize: "11px",
  cursor: "pointer" as const,
  fontWeight: "500" as const,
};

const inputStyle = (disabled: boolean) => ({
  width: "100%",
  padding: "5px 8px",
  background: disabled ? "rgba(0, 0, 0, 0.4)" : "rgba(0, 0, 0, 0.8)",
  border: "1px solid rgba(59, 130, 246, 0.3)",
  borderRadius: "4px",
  color: disabled ? "#666" : "white",
  fontSize: "11px",
  outline: "none",
  boxSizing: "border-box" as const,
  cursor: disabled ? ("default" as const) : ("text" as const),
});

const filterNumericInput = (value: string): string => {
  return value.replace(/[^0-9-]/g, "").replace(/(?!^)-/g, "");
};

const parseVal = (v: string): number => {
  if (v === "" || v === undefined || v === null) {
    return 0;
  }
  const n = parseInt(v, 10);
  return isNaN(n) ? 0 : Math.max(0, n);
};

const DRAG_PIXELS_PER_STEP = 8;

const DraggableInput: React.FC<{
  value: string;
  onChange: (v: string) => void;
  onBlur: (e: React.FocusEvent<HTMLInputElement>) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  disabled: boolean;
  style: React.CSSProperties;
  allowNegative?: boolean;
}> = ({ value, onChange, onBlur, onKeyDown, disabled, style, allowNegative }) => {
  const dragRef = useRef<{ startX: number; startVal: number; dragging: boolean } | null>(null);
  const valueRef = useRef(value);
  valueRef.current = value;

  const handleMouseDown = useCallback(
    (e: React.MouseEvent<HTMLInputElement>) => {
      if (disabled) return;
      const input = e.currentTarget;
      const currentVal = parseInt(valueRef.current, 10) || 0;
      dragRef.current = { startX: e.clientX, startVal: currentVal, dragging: false };

      const onMouseMove = (me: MouseEvent) => {
        if (!dragRef.current) return;
        const dx = me.clientX - dragRef.current.startX;
        if (!dragRef.current.dragging && Math.abs(dx) < 4) return;
        if (!dragRef.current.dragging) {
          dragRef.current.dragging = true;
          document.body.style.cursor = "ew-resize";
          document.body.style.userSelect = "none";
          input.blur();
        }
        const steps = Math.round(dx / DRAG_PIXELS_PER_STEP);
        let newVal = dragRef.current.startVal + steps;
        if (!allowNegative && newVal < 0) {
          newVal = 0;
        }
        onChange(newVal.toString());
      };

      const onMouseUp = () => {
        dragRef.current = null;
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
        document.body.style.cursor = "";
        document.body.style.userSelect = "";
      };

      document.addEventListener("mousemove", onMouseMove);
      document.addEventListener("mouseup", onMouseUp);
    },
    [disabled, allowNegative, onChange],
  );

  return (
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(filterNumericInput(e.target.value))}
      onBlur={onBlur}
      onKeyDown={onKeyDown}
      onMouseDown={handleMouseDown}
      disabled={disabled}
      style={{ ...style, cursor: disabled ? "default" : "ew-resize" }}
    />
  );
};

const PlayerResourcesPage: React.FC<PlayerResourcesPageProps> = ({
  gameState,
  selectedPlayerIds,
  onPlayerChange,
}) => {
  const allPlayers = [gameState.currentPlayer, ...gameState.otherPlayers];
  const playerId = selectedPlayerIds[0];

  const [resourcesForm, setResourcesForm] = useState<ResourceForm>({
    credit: "",
    steel: "",
    titanium: "",
    plant: "",
    energy: "",
    heat: "",
  });

  const [productionForm, setProductionForm] = useState<ResourceForm>({
    credit: "",
    steel: "",
    titanium: "",
    plant: "",
    energy: "",
    heat: "",
  });

  const [trValue, setTRValue] = useState("");

  useEffect(() => {
    if (playerId) {
      const player = allPlayers.find((p) => p.id === playerId);
      if (player?.resources) {
        setResourcesForm({
          credit: (player.resources.credits || 0).toString(),
          steel: (player.resources.steel || 0).toString(),
          titanium: (player.resources.titanium || 0).toString(),
          plant: (player.resources.plants || 0).toString(),
          energy: (player.resources.energy || 0).toString(),
          heat: (player.resources.heat || 0).toString(),
        });
      }
      if (player?.production) {
        setProductionForm({
          credit: (player.production.credits || 0).toString(),
          steel: (player.production.steel || 0).toString(),
          titanium: (player.production.titanium || 0).toString(),
          plant: (player.production.plants || 0).toString(),
          energy: (player.production.energy || 0).toString(),
          heat: (player.production.heat || 0).toString(),
        });
      }
      if (player) {
        setTRValue((player.terraformRating || 20).toString());
      }
    } else {
      const empty = { credit: "", steel: "", titanium: "", plant: "", energy: "", heat: "" };
      setResourcesForm(empty);
      setProductionForm(empty);
      setTRValue("");
    }
  }, [playerId]);

  const sendCommand = async (commandType: string, payload: any) => {
    const req: AdminCommandRequest = { commandType: commandType as any, payload };
    try {
      await globalWebSocketManager.sendAdminCommand(req);
    } catch (error) {
      console.error("Failed to send admin command:", error);
    }
  };

  const handleSetResources = async () => {
    if (!playerId) {
      return;
    }
    const command: SetResourcesAdminCommand = {
      playerId,
      resources: {
        credits: parseVal(resourcesForm.credit),
        steel: parseVal(resourcesForm.steel),
        titanium: parseVal(resourcesForm.titanium),
        plants: parseVal(resourcesForm.plant),
        energy: parseVal(resourcesForm.energy),
        heat: parseVal(resourcesForm.heat),
      },
    };
    await sendCommand(AdminCommandTypeSetResources, command);
  };

  const handleSetProduction = async () => {
    if (!playerId) {
      return;
    }
    const command: SetProductionAdminCommand = {
      playerId,
      production: {
        credits: parseVal(productionForm.credit),
        steel: parseVal(productionForm.steel),
        titanium: parseVal(productionForm.titanium),
        plants: parseVal(productionForm.plant),
        energy: parseVal(productionForm.energy),
        heat: parseVal(productionForm.heat),
      },
    };
    await sendCommand(AdminCommandTypeSetProduction, command);
  };

  const handleSetTR = async () => {
    if (!playerId || !trValue) {
      return;
    }
    const val = parseInt(trValue, 10);
    if (isNaN(val) || val < 1 || val > 70) {
      return;
    }
    const command: SetTRAdminCommand = { playerId, terraformRating: val };
    await sendCommand(AdminCommandTypeSetTR, command);
  };

  const players = allPlayers.map((p) => ({ id: p.id, name: p.name }));

  const zeroForm: ResourceForm = {
    credit: "0",
    steel: "0",
    titanium: "0",
    plant: "0",
    energy: "0",
    heat: "0",
  };

  const renderResourceGrid = (
    form: ResourceForm,
    setForm: React.Dispatch<React.SetStateAction<ResourceForm>>,
    onSubmit: () => Promise<void>,
    label: string,
    allowNegativeBlur: boolean,
    isProduction: boolean,
  ) => (
    <div style={{ marginBottom: "16px" }}>
      <label
        style={{
          color: "#3b82f6",
          fontSize: "11px",
          fontWeight: "bold",
          display: "block",
          textAlign: "left",
          marginBottom: "8px",
        }}
      >
        {label}
      </label>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr 1fr",
          gap: "6px 12px",
          marginBottom: "8px",
        }}
      >
        {RESOURCE_FIELDS.map((resource) => {
          const baseIcon = RESOURCE_ICON_MAP[resource];
          const iconType =
            isProduction && baseIcon ? (`${baseIcon}-production` as GameIconType) : baseIcon;
          return (
            <div
              key={resource}
              style={{ display: "flex", alignItems: "center", gap: "4px", minWidth: 0 }}
            >
              {iconType && <GameIcon iconType={iconType} size="small" />}
              <DraggableInput
                value={form[resource]}
                onChange={(v) => setForm((prev) => ({ ...prev, [resource]: v }))}
                onBlur={(e) => {
                  const val = e.target.value;
                  if (val === "" || val === "-") {
                    return;
                  }
                  const num = parseInt(val, 10);
                  if (allowNegativeBlur) {
                    if (isNaN(num)) {
                      setForm((prev) => ({ ...prev, [resource]: "0" }));
                    }
                  } else {
                    if (isNaN(num) || num < 0) {
                      setForm((prev) => ({ ...prev, [resource]: "0" }));
                    }
                  }
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    void onSubmit();
                  }
                }}
                disabled={!playerId}
                style={inputStyle(!playerId)}
                allowNegative={allowNegativeBlur}
              />
            </div>
          );
        })}
      </div>
      <div style={{ display: "flex", justifyContent: "flex-end", gap: "8px" }}>
        <button onClick={() => setForm(zeroForm)} style={secondaryButtonStyle}>
          Defaults
        </button>
        <button onClick={() => void onSubmit()} style={buttonStyle}>
          Set
        </button>
      </div>
    </div>
  );

  return (
    <div>
      <PlayerSelector players={players} selectedIds={selectedPlayerIds} onChange={onPlayerChange} />

      <div style={{ borderTop: "1px solid #333", paddingTop: "12px" }}>
        {renderResourceGrid(
          resourcesForm,
          setResourcesForm,
          handleSetResources,
          "Resources",
          false,
          false,
        )}

        <div style={{ borderTop: "1px solid #222", paddingTop: "12px" }}>
          {renderResourceGrid(
            productionForm,
            setProductionForm,
            handleSetProduction,
            "Production",
            true,
            true,
          )}
        </div>

        <div style={{ borderTop: "1px solid #222", paddingTop: "12px" }}>
          <label
            style={{
              color: "#3b82f6",
              fontSize: "11px",
              fontWeight: "bold",
              display: "block",
              textAlign: "left",
              marginBottom: "8px",
            }}
          >
            Terraform Rating
          </label>
          <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <GameIcon iconType={"terraform-rating" as GameIconType} size="small" />
            <DraggableInput
              value={trValue}
              onChange={setTRValue}
              onBlur={(e) => {
                const val = e.target.value;
                if (val === "" || val === "-") {
                  return;
                }
                const num = parseInt(val, 10);
                if (isNaN(num) || num < 1) {
                  setTRValue("1");
                } else if (num > 70) {
                  setTRValue("70");
                }
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  void handleSetTR();
                }
              }}
              disabled={!playerId}
              style={{ ...inputStyle(!playerId), width: "80px" }}
            />
            <span style={{ color: "#666", fontSize: "10px" }}>(1-70)</span>
            <div style={{ flex: 1 }} />
            <button onClick={() => setTRValue("20")} style={secondaryButtonStyle}>
              Defaults
            </button>
            <button onClick={() => void handleSetTR()} style={buttonStyle}>
              Set
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PlayerResourcesPage;
