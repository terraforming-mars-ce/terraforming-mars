import React, { useState } from "react";
import TreeNode from "../TreeNode.tsx";
import { GameDto } from "../../../../types/generated/api-types.ts";

interface GameStatePanelProps {
  gameState: GameDto | null;
  changedPaths: Set<string>;
}

const GameStatePanel: React.FC<GameStatePanelProps> = ({ gameState, changedPaths }) => {
  const [searchTerm, setSearchTerm] = useState("");
  const [expandAll, setExpandAll] = useState(false);
  const [expandAllSignal, setExpandAllSignal] = useState<number>(0);

  const handleCopyAll = () => {
    if (gameState) {
      void navigator.clipboard.writeText(JSON.stringify(gameState, null, 2));
    }
  };

  const filterGameState = (obj: any, search: string): any => {
    if (!search) {
      return obj;
    }

    const searchLower = search.toLowerCase();

    if (typeof obj !== "object" || obj === null) {
      return String(obj).toLowerCase().includes(searchLower) ? obj : undefined;
    }

    if (Array.isArray(obj)) {
      const filtered = obj
        .map((item) => filterGameState(item, search))
        .filter((item) => item !== undefined);
      return filtered.length > 0 ? filtered : undefined;
    }

    const filtered: any = {};
    let hasMatch = false;

    for (const [key, value] of Object.entries(obj)) {
      if (key.toLowerCase().includes(searchLower)) {
        filtered[key] = value;
        hasMatch = true;
      } else {
        const filteredValue = filterGameState(value, search);
        if (filteredValue !== undefined) {
          filtered[key] = filteredValue;
          hasMatch = true;
        }
      }
    }

    return hasMatch ? filtered : undefined;
  };

  const displayState = searchTerm ? filterGameState(gameState, searchTerm) : gameState;

  return (
    <>
      <div
        style={{
          marginBottom: "12px",
          display: "flex",
          gap: "8px",
        }}
      >
        <input
          type="text"
          placeholder="Search..."
          spellCheck={false}
          autoComplete="off"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            flex: 1,
            padding: "6px 10px",
            background: "rgba(255, 255, 255, 0.1)",
            border: "1px solid #333",
            borderRadius: "4px",
            color: "white",
            fontSize: "13px",
          }}
        />
        <button
          onClick={() => {
            const newExpandAll = !expandAll;
            setExpandAll(newExpandAll);
            setExpandAllSignal(Date.now());
          }}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            padding: "6px 12px",
            background: "transparent",
            border: "1px solid rgba(59, 130, 246, 0.5)",
            borderRadius: "4px",
            color: "#93c5fd",
            fontSize: "12px",
            cursor: "pointer",
          }}
        >
          {expandAll ? "Collapse" : "Expand"} All
        </button>
        <button
          onClick={handleCopyAll}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            padding: "6px 12px",
            background: "transparent",
            border: "1px solid rgba(59, 130, 246, 0.5)",
            borderRadius: "4px",
            color: "#93c5fd",
            fontSize: "12px",
            cursor: "pointer",
          }}
        >
          Copy JSON
        </button>
      </div>

      <div
        className="debug-content-area"
        style={{
          flex: 1,
          overflow: "auto",
          background: "rgba(0, 0, 0, 0.5)",
          padding: "12px",
          borderRadius: "4px",
          border: "1px solid #222",
        }}
      >
        {displayState ? (
          <div>
            {Object.entries(displayState).map(([key, value]) => (
              <TreeNode
                key={key}
                nodeKey={key}
                value={value}
                changedPaths={changedPaths}
                expandAllSignal={expandAllSignal}
                shouldExpandAll={expandAll}
              />
            ))}
          </div>
        ) : gameState ? (
          <div style={{ color: "#666", textAlign: "center", padding: "20px" }}>
            No matches found for &quot;{searchTerm}&quot;
          </div>
        ) : (
          <div style={{ color: "#666", textAlign: "center", padding: "20px" }}>
            No game state available
          </div>
        )}
      </div>

      {changedPaths.size > 0 && (
        <div
          style={{
            marginTop: "12px",
            paddingTop: "12px",
            borderTop: "1px solid #333",
          }}
        >
          <span style={{ color: "#ffdf00", fontSize: "11px" }}>
            {changedPaths.size} change{changedPaths.size === 1 ? "" : "s"} detected
          </span>
        </div>
      )}
    </>
  );
};

export default GameStatePanel;
