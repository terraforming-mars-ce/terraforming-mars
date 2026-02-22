import React, { useRef, useState } from "react";
import TreeNode from "./TreeNode.tsx";
import AdminCommandPanel from "./AdminCommandPanel.tsx";
import World3DSettingsPanel from "./World3DSettingsPanel.tsx";
import { useWindowDrag, useWindowManager } from "./WindowManager.tsx";
import { GameDto } from "../../../types/generated/api-types.ts";

const WINDOW_ID = "admin-tools";
const WINDOW_WIDTH = 600;
const EXCLUDE_SELECTORS = [".tree-expand-toggle", ".tree-node-content", ".debug-content-area"];

interface DebugDropdownProps {
  isVisible: boolean;
  onClose: () => void;
  gameState: GameDto | null;
  changedPaths?: Set<string>;
  onOpenTilePlacer?: (playerId: string) => void;
}

const DebugDropdown: React.FC<DebugDropdownProps> = ({
  isVisible,
  onClose,
  gameState,
  changedPaths = new Set(),
  onOpenTilePlacer,
}) => {
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [expandAll, setExpandAll] = useState(false);
  const [expandAllSignal, setExpandAllSignal] = useState<number>(0);
  const [activeTab, setActiveTab] = useState<"state" | "admin" | "3d-world">("state");

  const { position, isDragging, handleMouseDown } = useWindowDrag({
    windowId: WINDOW_ID,
    width: WINDOW_WIDTH,
    height: () => window.innerHeight * 0.7,
    excludeSelectors: EXCLUDE_SELECTORS,
    isVisible,
  });

  const { getZIndex } = useWindowManager();

  if (!isVisible) return null;

  const handleCopyAll = () => {
    if (gameState) {
      navigator.clipboard.writeText(JSON.stringify(gameState, null, 2));
    }
  };

  const handleExportJSON = () => {
    if (gameState) {
      const dataStr = JSON.stringify(gameState, null, 2);
      const dataUri = "data:application/json;charset=utf-8," + encodeURIComponent(dataStr);
      const exportFileDefaultName = `game_state_${Date.now()}.json`;

      const linkElement = document.createElement("a");
      linkElement.setAttribute("href", dataUri);
      linkElement.setAttribute("download", exportFileDefaultName);
      linkElement.click();
    }
  };

  const filterGameState = (obj: any, search: string): any => {
    if (!search) return obj;

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
    <div
      ref={dropdownRef}
      className="debug-dropdown"
      onMouseDown={handleMouseDown}
      style={{
        position: "fixed",
        top: `${position.y}px`,
        left: `${position.x}px`,
        width: "600px",
        maxHeight: "70vh",
        background: "rgba(0, 0, 0, 0.95)",
        border: "2px solid #9b59b6",
        borderRadius: "8px",
        padding: "16px",
        zIndex: getZIndex(WINDOW_ID),
        overflow: activeTab === "admin" ? "visible" : "hidden",
        display: "flex",
        flexDirection: "column",
        boxShadow: "0 4px 20px rgba(155, 89, 182, 0.3)",
        cursor: isDragging ? "grabbing" : "default",
        transition: isDragging ? "none" : "top 0.2s ease-out, left 0.2s ease-out",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "12px",
          paddingBottom: "12px",
          borderBottom: "1px solid #333",
          userSelect: "none",
          cursor: "grab",
        }}
      >
        <h3
          style={{
            margin: 0,
            color: "#9b59b6",
            fontSize: "16px",
            display: "flex",
            alignItems: "center",
            gap: "8px",
          }}
        >
          <span style={{ opacity: 0.7, fontSize: "12px" }}>⋮⋮</span>
          Admin Tools
        </h3>
        <button
          onClick={onClose}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            background: "none",
            border: "none",
            color: "#abb2bf",
            fontSize: "20px",
            cursor: "pointer",
            padding: "0 4px",
          }}
        >
          ×
        </button>
      </div>

      <div
        style={{
          display: "flex",
          marginBottom: "12px",
          borderBottom: "1px solid #333",
        }}
      >
        <button
          onClick={() => setActiveTab("state")}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            flex: 1,
            padding: "8px",
            background: activeTab === "state" ? "rgba(155, 89, 182, 0.3)" : "transparent",
            border: "none",
            borderBottom: activeTab === "state" ? "2px solid #9b59b6" : "2px solid transparent",
            color: activeTab === "state" ? "#9b59b6" : "#abb2bf",
            fontSize: "12px",
            cursor: "pointer",
            transition: "all 0.2s",
          }}
        >
          Game State
        </button>
        {gameState?.settings.developmentMode && (
          <button
            onClick={() => setActiveTab("admin")}
            onMouseDown={(e) => e.stopPropagation()}
            style={{
              flex: 1,
              padding: "8px",
              background: activeTab === "admin" ? "rgba(155, 89, 182, 0.3)" : "transparent",
              border: "none",
              borderBottom: activeTab === "admin" ? "2px solid #9b59b6" : "2px solid transparent",
              color: activeTab === "admin" ? "#9b59b6" : "#abb2bf",
              fontSize: "12px",
              cursor: "pointer",
              transition: "all 0.2s",
            }}
          >
            Commands
          </button>
        )}
        <button
          onClick={() => setActiveTab("3d-world")}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            flex: 1,
            padding: "8px",
            background: activeTab === "3d-world" ? "rgba(155, 89, 182, 0.3)" : "transparent",
            border: "none",
            borderBottom: activeTab === "3d-world" ? "2px solid #9b59b6" : "2px solid transparent",
            color: activeTab === "3d-world" ? "#9b59b6" : "#abb2bf",
            fontSize: "12px",
            cursor: "pointer",
            transition: "all 0.2s",
          }}
        >
          3D World
        </button>
      </div>

      {activeTab === "state" && (
        <>
          <div
            style={{
              marginBottom: "12px",
              display: "flex",
              gap: "8px",
              cursor: "grab",
            }}
          >
            <input
              type="text"
              placeholder="Search keys or values..."
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
                background: "rgba(155, 89, 182, 0.2)",
                border: "1px solid #9b59b6",
                borderRadius: "4px",
                color: "#9b59b6",
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
                background: "rgba(155, 89, 182, 0.2)",
                border: "1px solid #9b59b6",
                borderRadius: "4px",
                color: "#9b59b6",
                fontSize: "12px",
                cursor: "pointer",
              }}
            >
              Copy JSON
            </button>
            <button
              onClick={handleExportJSON}
              onMouseDown={(e) => e.stopPropagation()}
              style={{
                padding: "6px 12px",
                background: "rgba(155, 89, 182, 0.2)",
                border: "1px solid #9b59b6",
                borderRadius: "4px",
                color: "#9b59b6",
                fontSize: "12px",
                cursor: "pointer",
              }}
            >
              Export
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
                No matches found for "{searchTerm}"
              </div>
            ) : (
              <div style={{ color: "#666", textAlign: "center", padding: "20px" }}>
                No game state available
              </div>
            )}
          </div>

          <div
            style={{
              marginTop: "12px",
              paddingTop: "12px",
              borderTop: "1px solid #333",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              cursor: "grab",
            }}
          >
            <span style={{ color: "#666", fontSize: "11px" }}>
              {changedPaths.size > 0 && (
                <span style={{ color: "#ffdf00" }}>
                  {changedPaths.size} change{changedPaths.size === 1 ? "" : "s"} detected
                </span>
              )}
            </span>
            <span style={{ color: "#666", fontSize: "11px" }}>
              Press Ctrl+D to toggle • Double-click values to copy
            </span>
          </div>
        </>
      )}

      {activeTab === "admin" && gameState?.settings.developmentMode && (
        <AdminCommandPanel
          gameState={gameState}
          onClose={onClose}
          onOpenTilePlacer={onOpenTilePlacer}
        />
      )}

      {activeTab === "3d-world" && <World3DSettingsPanel />}
    </div>
  );
};

export default DebugDropdown;
