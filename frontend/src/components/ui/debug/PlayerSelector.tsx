import React from "react";

interface PlayerInfo {
  id: string;
  name: string;
}

interface PlayerSelectorProps {
  players: PlayerInfo[];
  selectedId: string;
  onChange: (id: string) => void;
}

const PlayerSelector: React.FC<PlayerSelectorProps> = ({ players, selectedId, onChange }) => {
  return (
    <div style={{ marginBottom: "8px" }}>
      <label
        style={{
          color: "#3b82f6",
          fontSize: "11px",
          fontWeight: "bold",
          display: "block",
          textAlign: "left",
          marginBottom: "4px",
        }}
      >
        Player
      </label>
      <select
        value={selectedId}
        onChange={(e) => onChange(e.target.value)}
        style={{
          width: "100%",
          padding: "6px 10px",
          background: "rgba(0, 0, 0, 0.8)",
          border: "1px solid rgba(59, 130, 246, 0.3)",
          borderRadius: "4px",
          color: "white",
          fontSize: "12px",
          outline: "none",
          boxSizing: "border-box",
          cursor: "pointer",
        }}
      >
        {players.map((p) => (
          <option key={p.id} value={p.id}>
            {p.name}
          </option>
        ))}
      </select>
    </div>
  );
};

export default PlayerSelector;
