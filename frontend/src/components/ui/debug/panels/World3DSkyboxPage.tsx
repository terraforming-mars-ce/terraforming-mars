import React from "react";
import { useWorld3DSettings, SKYBOX_OPTIONS } from "../../../../contexts/World3DSettingsContext";

const World3DSkyboxPage: React.FC = () => {
  const { settings, updateSettings } = useWorld3DSettings();

  const currentSkybox = SKYBOX_OPTIONS[0];

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
        Skybox
      </label>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "10px",
          padding: "8px 12px",
          background: "rgba(59, 130, 246, 0.2)",
          border: "1px solid rgba(59, 130, 246, 0.5)",
          borderRadius: "6px",
        }}
      >
        <div>
          <div style={{ color: "#fff", fontSize: "12px", fontWeight: "500" }}>
            {currentSkybox.label}
          </div>
          <div style={{ color: "#666", fontSize: "10px", marginTop: "2px" }}>
            {currentSkybox.path.split("/").pop()}
          </div>
        </div>
      </div>

      <div style={{ marginTop: "16px" }}>
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
          Brightness
        </label>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <input
            type="range"
            min="0"
            max="1"
            step="0.01"
            value={settings.skyboxBrightness}
            onChange={(e) => updateSettings({ skyboxBrightness: parseFloat(e.target.value) })}
            style={{ flex: 1, accentColor: "#3b82f6" }}
          />
          <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
            {settings.skyboxBrightness.toFixed(2)}
          </span>
        </div>
      </div>
    </div>
  );
};

export default World3DSkyboxPage;
