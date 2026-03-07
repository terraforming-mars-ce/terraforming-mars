import React, { useState, useRef, useEffect } from "react";

interface PlayerInfo {
  id: string;
  name: string;
}

interface PlayerSelectorProps {
  players: PlayerInfo[];
  selectedIds: string[];
  onChange: (ids: string[]) => void;
}

const PlayerSelector: React.FC<PlayerSelectorProps> = ({ players, selectedIds, onChange }) => {
  const [query, setQuery] = useState("");
  const [showDropdown, setShowDropdown] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setShowDropdown(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const unselectedPlayers = players.filter((p) => !selectedIds.includes(p.id));
  const filteredPlayers = query.trim()
    ? unselectedPlayers.filter((p) => p.name.toLowerCase().includes(query.toLowerCase()))
    : unselectedPlayers;

  const selectedPlayers = players.filter((p) => selectedIds.includes(p.id));

  const togglePlayer = (id: string) => {
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter((sid) => sid !== id));
    } else {
      onChange([...selectedIds, id]);
      setQuery("");
    }
  };

  const removePlayer = (id: string) => {
    onChange(selectedIds.filter((sid) => sid !== id));
  };

  return (
    <div ref={containerRef} style={{ marginBottom: "8px" }}>
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
        Players
      </label>

      {selectedPlayers.length > 0 && (
        <div style={{ display: "flex", flexWrap: "wrap", gap: "4px", marginBottom: "4px" }}>
          {selectedPlayers.map((p) => (
            <span
              key={p.id}
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: "4px",
                padding: "2px 8px",
                background: "rgba(59, 130, 246, 0.3)",
                border: "1px solid rgba(59, 130, 246, 0.5)",
                borderRadius: "12px",
                fontSize: "11px",
                color: "#fff",
              }}
            >
              {p.name}
              <span
                onClick={() => removePlayer(p.id)}
                style={{ cursor: "pointer", color: "#abb2bf", fontSize: "13px", lineHeight: 1 }}
              >
                ×
              </span>
            </span>
          ))}
        </div>
      )}

      <div style={{ position: "relative" }}>
        <input
          type="text"
          placeholder={selectedPlayers.length > 0 ? "Add another player..." : "Search players..."}
          spellCheck={false}
          autoComplete="off"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setShowDropdown(true);
          }}
          onFocus={() => setShowDropdown(true)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && filteredPlayers.length === 1) {
              togglePlayer(filteredPlayers[0].id);
              setShowDropdown(false);
            }
            if (e.key === "Backspace" && query === "" && selectedIds.length > 0) {
              removePlayer(selectedIds[selectedIds.length - 1]);
            }
          }}
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
          }}
        />
        {showDropdown && filteredPlayers.length > 0 && (
          <div
            style={{
              position: "absolute",
              top: "100%",
              left: 0,
              right: 0,
              background: "rgba(0, 0, 0, 0.98)",
              border: "1px solid rgba(59, 130, 246, 0.5)",
              borderRadius: "4px",
              marginTop: "2px",
              zIndex: 9999,
              boxShadow: "0 4px 12px rgba(0, 0, 0, 0.8)",
            }}
          >
            {filteredPlayers.map((player) => (
              <div
                key={player.id}
                onClick={() => {
                  togglePlayer(player.id);
                  setShowDropdown(false);
                }}
                style={{
                  padding: "6px 12px",
                  cursor: "pointer",
                  fontSize: "12px",
                  color: "#abb2bf",
                  borderBottom: "1px solid rgba(59, 130, 246, 0.2)",
                  transition: "background 0.15s ease",
                }}
                onMouseEnter={(e) => (e.currentTarget.style.background = "rgba(59, 130, 246, 0.2)")}
                onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
              >
                {player.name}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default PlayerSelector;
