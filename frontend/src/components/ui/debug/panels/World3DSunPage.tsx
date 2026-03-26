import React from "react";
import { useWorld3DSettings } from "../../../../contexts/World3DSettingsContext";
import { ColorSwatch } from "../HSVColorPicker.tsx";

const World3DSunPage: React.FC = () => {
  const { settings, updateSettings, resetSettings } = useWorld3DSettings();

  return (
    <div>
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
          Sun Intensity
        </label>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <input
            type="range"
            min="0"
            max="3"
            step="0.1"
            value={settings.sunIntensity}
            onChange={(e) => updateSettings({ sunIntensity: parseFloat(e.target.value) })}
            style={{ flex: 1, accentColor: "#3b82f6" }}
          />
          <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
            {settings.sunIntensity.toFixed(1)}
          </span>
        </div>
      </div>

      <div style={{ marginBottom: "16px" }}>
        <ColorSwatch
          label="Sun Color"
          color={settings.sunColor}
          onChange={(c) => updateSettings({ sunColor: c })}
        />
      </div>

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
          Fresnel Reflectance (rf0)
        </label>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <input
            type="range"
            min="0"
            max="1"
            step="0.01"
            value={settings.reflectance}
            onChange={(e) => updateSettings({ reflectance: parseFloat(e.target.value) })}
            style={{ flex: 1, accentColor: "#3b82f6" }}
          />
          <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
            {settings.reflectance.toFixed(2)}
          </span>
        </div>
      </div>

      <div style={{ marginBottom: "16px" }}>
        <ColorSwatch
          label="Water Color"
          color={settings.waterColor}
          onChange={(c) => updateSettings({ waterColor: c })}
        />
      </div>

      <button
        onClick={resetSettings}
        style={{
          padding: "8px 16px",
          background: "linear-gradient(135deg, rgba(59, 130, 246, 0.8), rgba(59, 130, 246, 0.6))",
          border: "1px solid rgba(59, 130, 246, 0.5)",
          borderRadius: "6px",
          color: "white",
          fontSize: "12px",
          cursor: "pointer",
          fontWeight: "500",
        }}
      >
        Reset to Defaults
      </button>
    </div>
  );
};

export default World3DSunPage;
