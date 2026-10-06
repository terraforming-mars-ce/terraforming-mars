import React from "react";
import { useWorld3DSettings } from "../../../../contexts/World3DSettingsContext";
import { GameDto } from "../../../../types/generated/api-types.ts";
import {
  BARREN_PARAMETERS,
  OCEANS_MAX,
  OXYGEN_MAX,
  TEMPERATURE_MAX,
  TEMPERATURE_MIN,
  TERRAFORMED_PARAMETERS,
  type ClimateParameters,
} from "../../../game/board/climate";

const labelStyle: React.CSSProperties = {
  color: "#3b82f6",
  fontSize: "11px",
  fontWeight: "bold",
  display: "block",
  textAlign: "left",
  marginBottom: "8px",
};

const buttonStyle: React.CSSProperties = {
  flex: 1,
  padding: "6px 8px",
  background: "rgba(59, 130, 246, 0.2)",
  border: "1px solid rgba(59, 130, 246, 0.5)",
  borderRadius: "6px",
  color: "#fff",
  fontSize: "11px",
  cursor: "pointer",
};

const SLIDERS = [
  {
    key: "temperature",
    label: "Temperature",
    min: TEMPERATURE_MIN,
    max: TEMPERATURE_MAX,
    step: 2,
    unit: "°C",
  },
  { key: "oxygen", label: "Oxygen", min: 0, max: OXYGEN_MAX, step: 1, unit: "%" },
  { key: "oceans", label: "Oceans", min: 0, max: OCEANS_MAX, step: 1, unit: "" },
] as const;

interface World3DClimatePageProps {
  gameState: GameDto | null;
}

const World3DClimatePage: React.FC<World3DClimatePageProps> = ({ gameState }) => {
  const { settings, updateSettings } = useWorld3DSettings();
  const game = gameState?.globalParameters;
  const gameParameters: ClimateParameters = game
    ? { temperature: game.temperature, oxygen: game.oxygen, oceans: game.oceans }
    : BARREN_PARAMETERS;
  const override = settings.climateOverride;
  const shown = override ?? gameParameters;

  return (
    <div>
      <label style={labelStyle}>Climate preview</label>
      <div style={{ color: "#888", fontSize: "11px", textAlign: "left", marginBottom: "12px" }}>
        Client-only override of the parameters driving the 3D climate. The game state is unchanged.
      </div>
      <label
        style={{
          display: "flex",
          alignItems: "center",
          gap: "8px",
          color: "#fff",
          fontSize: "12px",
        }}
      >
        <input
          type="checkbox"
          checked={override !== null}
          onChange={(e) =>
            updateSettings({ climateOverride: e.target.checked ? { ...gameParameters } : null })
          }
          style={{ accentColor: "#3b82f6" }}
        />
        Override game values
      </label>

      {SLIDERS.map(({ key, label, min, max, step, unit }) => (
        <div key={key} style={{ marginTop: "16px", opacity: override ? 1 : 0.5 }}>
          <label style={labelStyle}>{label}</label>
          <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <input
              type="range"
              min={min}
              max={max}
              step={step}
              disabled={!override}
              value={shown[key]}
              onChange={(e) =>
                override &&
                updateSettings({
                  climateOverride: { ...override, [key]: parseFloat(e.target.value) },
                })
              }
              style={{ flex: 1, accentColor: "#3b82f6" }}
            />
            <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
              {shown[key]}
              {unit}
            </span>
          </div>
        </div>
      ))}

      <div style={{ display: "flex", gap: "8px", marginTop: "16px" }}>
        <button
          style={buttonStyle}
          onClick={() => updateSettings({ climateOverride: { ...BARREN_PARAMETERS } })}
        >
          Barren
        </button>
        <button
          style={buttonStyle}
          onClick={() => updateSettings({ climateOverride: { ...TERRAFORMED_PARAMETERS } })}
        >
          Terraformed
        </button>
      </div>
    </div>
  );
};

export default World3DClimatePage;
