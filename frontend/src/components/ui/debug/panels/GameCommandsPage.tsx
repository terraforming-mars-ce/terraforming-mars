import React, { useState, useEffect } from "react";
import { globalWebSocketManager } from "../../../../services/globalWebSocketManager.ts";
import {
  GameDto,
  AdminCommandRequest,
  AdminCommandTypeSetGlobalParams,
  AdminCommandTypeSetPhase,
  SetGlobalParamsAdminCommand,
  SetPhaseAdminCommand,
  GamePhaseWaitingForGameStart,
  GamePhaseStartingSelection,
  GamePhaseAction,
  GamePhaseProductionAndCardDraw,
  GamePhaseComplete,
} from "../../../../types/generated/api-types.ts";

interface GameCommandsPageProps {
  gameState: GameDto;
}

const PARAMS = [
  { key: "temperature", label: "Temperature", min: -30, max: 8, unit: "°C" },
  { key: "oxygen", label: "Oxygen", min: 0, max: 14, unit: "%" },
  { key: "oceans", label: "Oceans", min: 0, max: 9, unit: "" },
  { key: "venus", label: "Venus", min: 0, max: 30, unit: "%" },
] as const;

type ParamKey = (typeof PARAMS)[number]["key"];

const PHASE_OPTIONS = [
  { value: GamePhaseWaitingForGameStart, label: "Waiting for Game Start" },
  { value: GamePhaseStartingSelection, label: "Starting Card Selection" },
  { value: GamePhaseAction, label: "Action Phase" },
  { value: GamePhaseProductionAndCardDraw, label: "Production and Card Draw" },
  { value: GamePhaseComplete, label: "Game Complete" },
];

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

const smallButtonStyle = {
  padding: "3px 8px",
  background: "rgba(59, 130, 246, 0.3)",
  border: "1px solid rgba(59, 130, 246, 0.4)",
  borderRadius: "4px",
  color: "white",
  fontSize: "10px",
  cursor: "pointer" as const,
  fontWeight: "500" as const,
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

const GameCommandsPage: React.FC<GameCommandsPageProps> = ({ gameState }) => {
  const [form, setForm] = useState<Record<ParamKey, string>>({
    temperature: gameState.globalParameters.temperature.toString(),
    oxygen: gameState.globalParameters.oxygen.toString(),
    oceans: gameState.globalParameters.oceans.toString(),
    venus: gameState.globalParameters.venus.toString(),
  });
  const [phase, setPhase] = useState("");

  useEffect(() => {
    setForm({
      temperature: gameState.globalParameters.temperature.toString(),
      oxygen: gameState.globalParameters.oxygen.toString(),
      oceans: gameState.globalParameters.oceans.toString(),
      venus: gameState.globalParameters.venus.toString(),
    });
  }, [gameState.globalParameters]);

  const filterNumericInput = (value: string): string => {
    return value.replace(/[^0-9-]/g, "").replace(/(?!^)-/g, "");
  };

  const clampValue = (value: string, min: number, max: number, defaultVal: number): string => {
    if (value === "" || value === "-") {
      return defaultVal.toString();
    }
    const num = parseInt(value, 10);
    if (isNaN(num)) {
      return defaultVal.toString();
    }
    return Math.max(min, Math.min(max, num)).toString();
  };

  const sendCommand = async (commandType: string, payload: any) => {
    const req: AdminCommandRequest = { commandType: commandType as any, payload };
    try {
      await globalWebSocketManager.sendAdminCommand(req);
    } catch (error) {
      console.error("Failed to send admin command:", error);
    }
  };

  const handleSetGlobalParams = async () => {
    const command: SetGlobalParamsAdminCommand = {
      globalParameters: {
        temperature: isNaN(parseInt(form.temperature, 10)) ? -30 : parseInt(form.temperature, 10),
        oxygen: isNaN(parseInt(form.oxygen, 10)) ? 0 : parseInt(form.oxygen, 10),
        oceans: isNaN(parseInt(form.oceans, 10)) ? 0 : parseInt(form.oceans, 10),
        maxOceans: gameState.globalParameters.maxOceans,
        venus: isNaN(parseInt(form.venus, 10)) ? 0 : parseInt(form.venus, 10),
        bonuses: gameState.globalParameters.bonuses,
      },
    };
    await sendCommand(AdminCommandTypeSetGlobalParams, command);
  };

  const handleSetPhase = async () => {
    if (!phase) {
      return;
    }
    const command: SetPhaseAdminCommand = { phase };
    await sendCommand(AdminCommandTypeSetPhase, command);
  };

  const setAllMin = () => {
    const newForm: Record<string, string> = {};
    for (const p of PARAMS) {
      newForm[p.key] = p.min.toString();
    }
    setForm(newForm as Record<ParamKey, string>);
  };

  const setAllMax = () => {
    const newForm: Record<string, string> = {};
    for (const p of PARAMS) {
      newForm[p.key] = p.max.toString();
    }
    setForm(newForm as Record<ParamKey, string>);
  };

  return (
    <div>
      <label
        style={{
          color: "#3b82f6",
          fontSize: "11px",
          fontWeight: "bold",
          display: "block",
          textAlign: "left",
          marginBottom: "12px",
        }}
      >
        Global Parameters
      </label>

      <div style={{ display: "flex", flexDirection: "column", gap: "10px", marginBottom: "10px" }}>
        {PARAMS.map((param) => {
          const val = parseInt(form[param.key], 10);
          const sliderVal = isNaN(val) ? param.min : Math.max(param.min, Math.min(param.max, val));

          return (
            <div key={param.key}>
              <label
                style={{
                  color: "#abb2bf",
                  fontSize: "10px",
                  display: "block",
                  marginBottom: "2px",
                }}
              >
                {param.label} ({param.min} to {param.max}
                {param.unit}):
              </label>
              <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                <input
                  type="range"
                  min={param.min}
                  max={param.max}
                  step={param.key === "temperature" ? 2 : 1}
                  value={sliderVal}
                  onChange={(e) => setForm({ ...form, [param.key]: e.target.value })}
                  style={{ flex: 1, accentColor: "#3b82f6" }}
                />
                <input
                  type="text"
                  value={form[param.key]}
                  onChange={(e) =>
                    setForm({ ...form, [param.key]: filterNumericInput(e.target.value) })
                  }
                  onBlur={(e) =>
                    setForm({
                      ...form,
                      [param.key]: clampValue(e.target.value, param.min, param.max, param.min),
                    })
                  }
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      void handleSetGlobalParams();
                    }
                  }}
                  style={{
                    width: "50px",
                    padding: "4px 6px",
                    background: "rgba(0, 0, 0, 0.8)",
                    border: "1px solid rgba(59, 130, 246, 0.3)",
                    borderRadius: "4px",
                    color: "white",
                    fontSize: "12px",
                    outline: "none",
                    textAlign: "center",
                  }}
                />
              </div>
            </div>
          );
        })}
      </div>

      <div style={{ display: "flex", gap: "8px", marginBottom: "10px" }}>
        <button onClick={setAllMin} style={smallButtonStyle}>
          Min All
        </button>
        <button onClick={setAllMax} style={smallButtonStyle}>
          Max All
        </button>
        <div style={{ flex: 1 }} />
        <button
          onClick={() => setForm({ temperature: "-30", oxygen: "0", oceans: "0", venus: "0" })}
          style={secondaryButtonStyle}
        >
          Defaults
        </button>
        <button onClick={() => void handleSetGlobalParams()} style={buttonStyle}>
          Set
        </button>
      </div>

      <div style={{ borderTop: "1px solid #333", paddingTop: "12px", marginTop: "4px" }}>
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
          Game Phase
        </label>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <select
            value={phase}
            onChange={(e) => setPhase(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                void handleSetPhase();
              }
            }}
            style={{
              width: "220px",
              maxWidth: "100%",
              padding: "6px 10px",
              background: "rgba(0, 0, 0, 0.8)",
              border: "1px solid rgba(59, 130, 246, 0.3)",
              borderRadius: "4px",
              color: "white",
              fontSize: "12px",
              outline: "none",
              cursor: "pointer",
              appearance: "none",
              backgroundImage: `url("data:image/svg+xml;charset=US-ASCII,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 4 5'><path fill='%23abb2bf' d='M2 0L0 2h4zm0 5L0 3h4z'/></svg>")`,
              backgroundRepeat: "no-repeat",
              backgroundPosition: "right 6px center",
              backgroundSize: "10px",
              paddingRight: "28px",
            }}
          >
            <option value="">Select phase...</option>
            {PHASE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <button onClick={() => void handleSetPhase()} style={buttonStyle}>
            Set
          </button>
        </div>
      </div>
    </div>
  );
};

export default GameCommandsPage;
