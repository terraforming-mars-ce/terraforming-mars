import React, { useState, useEffect, useRef, useCallback } from "react";
import { createPortal } from "react-dom";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface TreeNodeProps {
  nodeKey: string;
  value: any;
  depth?: number;
  changedPaths?: Set<string>;
  currentPath?: string;
  expandAllSignal?: number;
  shouldExpandAll?: boolean;
}

interface ContextMenuState {
  x: number;
  y: number;
  nodeKey: string;
  value: any;
}

const ContextMenu: React.FC<{
  menu: ContextMenuState;
  onClose: () => void;
}> = ({ menu, onClose }) => {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleDismiss = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleDismiss, true);
    document.addEventListener("contextmenu", handleDismiss, true);
    return () => {
      document.removeEventListener("mousedown", handleDismiss, true);
      document.removeEventListener("contextmenu", handleDismiss, true);
    };
  }, [onClose]);

  const itemStyle = {
    padding: "6px 12px",
    fontSize: "12px",
    color: "#abb2bf",
    cursor: "pointer" as const,
    display: "flex",
    alignItems: "center",
    gap: "8px",
    transition: "background 0.15s ease",
  };

  const handleCopyObject = () => {
    void navigator.clipboard.writeText(JSON.stringify(menu.value, null, 2));
    onClose();
  };

  const handleCopyName = () => {
    void navigator.clipboard.writeText(menu.nodeKey);
    onClose();
  };

  return createPortal(
    <div
      ref={ref}
      style={{
        position: "fixed",
        top: menu.y,
        left: menu.x,
        background: "rgba(0, 0, 0, 0.98)",
        border: "1px solid rgba(59, 130, 246, 0.5)",
        borderRadius: "4px",
        zIndex: Z_INDEX.LOADING_OVERLAY,
        boxShadow: "0 4px 12px rgba(0, 0, 0, 0.8)",
        minWidth: "140px",
        overflow: "hidden",
      }}
    >
      <div
        onClick={handleCopyObject}
        style={itemStyle}
        onMouseEnter={(e) => (e.currentTarget.style.background = "rgba(59, 130, 246, 0.2)")}
        onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#abb2bf"
          strokeWidth="2"
        >
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
        </svg>
        Copy Object
      </div>
      <div
        onClick={handleCopyName}
        style={{ ...itemStyle, borderTop: "1px solid rgba(59, 130, 246, 0.2)" }}
        onMouseEnter={(e) => (e.currentTarget.style.background = "rgba(59, 130, 246, 0.2)")}
        onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#abb2bf"
          strokeWidth="2"
        >
          <path d="M4 7V4a2 2 0 012-2h8.5L20 7.5V20a2 2 0 01-2 2H6a2 2 0 01-2-2v-3" />
          <polyline points="14 2 14 8 20 8" />
          <line x1="2" y1="15" x2="12" y2="15" />
        </svg>
        Copy Name
      </div>
    </div>,
    document.body,
  );
};

const TreeNode: React.FC<TreeNodeProps> = ({
  nodeKey,
  value,
  depth = 0,
  changedPaths = new Set(),
  currentPath = "",
  expandAllSignal,
  shouldExpandAll,
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  useEffect(() => {
    if (expandAllSignal !== undefined) {
      setIsExpanded(shouldExpandAll || false);
    }
  }, [expandAllSignal, shouldExpandAll]);

  const nodePath = currentPath ? `${currentPath}.${nodeKey}` : nodeKey;
  const isChanged = changedPaths.has(nodePath);

  const getValueType = (val: any): string => {
    if (val === null) return "null";
    if (val === undefined) return "undefined";
    if (Array.isArray(val)) return "array";
    return typeof val;
  };

  const valueType = getValueType(value);

  const getTypeColor = (type: string) => {
    switch (type) {
      case "string":
        return "#98c379";
      case "number":
        return "#61afef";
      case "boolean":
        return "#d19a66";
      case "null":
      case "undefined":
        return "#abb2bf";
      case "object":
      case "array":
        return "#c678dd";
      default:
        return "#ffffff";
    }
  };

  const renderPrimitiveValue = () => {
    if (valueType === "string") {
      return <span style={{ color: getTypeColor(valueType) }}>"{value}"</span>;
    }
    if (valueType === "boolean" || valueType === "number") {
      return <span style={{ color: getTypeColor(valueType) }}>{String(value)}</span>;
    }
    if (valueType === "null" || valueType === "undefined") {
      return <span style={{ color: getTypeColor(valueType) }}>{valueType}</span>;
    }
    return null;
  };

  const renderObjectPreview = () => {
    if (valueType === "array") {
      return (
        <span style={{ color: "#abb2bf" }}>
          [{value.length} {value.length === 1 ? "item" : "items"}]
        </span>
      );
    }
    if (valueType === "object") {
      const keys = Object.keys(value);
      return (
        <span style={{ color: "#abb2bf" }}>
          {"{"}
          {keys.length} {keys.length === 1 ? "property" : "properties"}
          {"}"}
        </span>
      );
    }
    return null;
  };

  const renderExpandToggle = () => {
    if (valueType !== "object" && valueType !== "array") return null;
    if (
      (valueType === "array" && value.length === 0) ||
      (valueType === "object" && Object.keys(value).length === 0)
    ) {
      return null;
    }

    return (
      <span
        className="cursor-pointer mr-1 text-[#abb2bf] text-xs select-none"
        onClick={(e) => {
          e.stopPropagation();
          setIsExpanded(!isExpanded);
        }}
        onMouseDown={(e) => {
          e.stopPropagation();
        }}
      >
        {isExpanded ? "\u25BC" : "\u25B6"}
      </span>
    );
  };

  const renderKey = () => {
    return <span style={{ color: "#e06c75", marginRight: "4px" }}>{nodeKey}:</span>;
  };

  const renderExpandedContent = () => {
    if (!isExpanded) return null;

    if (valueType === "array") {
      return (
        <div className="ml-5">
          {value.map((item: any, index: number) => (
            <TreeNode
              key={index}
              nodeKey={String(index)}
              value={item}
              depth={depth + 1}
              changedPaths={changedPaths}
              currentPath={nodePath}
              expandAllSignal={expandAllSignal}
              shouldExpandAll={shouldExpandAll}
            />
          ))}
        </div>
      );
    }

    if (valueType === "object") {
      return (
        <div className="ml-5">
          {Object.entries(value).map(([key, val]) => (
            <TreeNode
              key={key}
              nodeKey={key}
              value={val}
              depth={depth + 1}
              changedPaths={changedPaths}
              currentPath={nodePath}
              expandAllSignal={expandAllSignal}
              shouldExpandAll={shouldExpandAll}
            />
          ))}
        </div>
      );
    }

    return null;
  };

  const handleContextMenu = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setContextMenu({
        x: e.clientX,
        y: e.clientY,
        nodeKey,
        value,
      });
    },
    [nodeKey, value],
  );

  const closeContextMenu = useCallback(() => setContextMenu(null), []);

  // Invisible indent text: included in copy/paste but takes no visual space
  const indentStr = "  ".repeat(depth);

  return (
    <div
      className={`font-mono text-[13px] leading-[1.5] relative ${isChanged ? "[&>.treeNodeContent]:animate-[blink_1.5s_ease-out]" : ""}`}
    >
      <div
        className="treeNodeContent flex items-start py-0.5 px-1 rounded relative select-text"
        onDoubleClick={() => void navigator.clipboard.writeText(JSON.stringify(value, null, 2))}
        onContextMenu={handleContextMenu}
      >
        {renderExpandToggle()}
        <span>
          <span aria-hidden="true" style={{ fontSize: 0, opacity: 0, whiteSpace: "pre" }}>
            {indentStr}
          </span>
          {renderKey()} {renderPrimitiveValue()}
          {renderObjectPreview()}
        </span>
      </div>
      {renderExpandedContent()}
      {contextMenu && <ContextMenu menu={contextMenu} onClose={closeContextMenu} />}
    </div>
  );
};

export default TreeNode;
