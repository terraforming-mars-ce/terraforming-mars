import React, { useState, useEffect } from "react";

interface TreeNodeProps {
  nodeKey: string;
  value: any;
  depth?: number;
  changedPaths?: Set<string>;
  currentPath?: string;
  expandAllSignal?: number; // Signal to expand/collapse all (timestamp)
  shouldExpandAll?: boolean; // Whether the signal is expand (true) or collapse (false)
}

const TreeNode: React.FC<TreeNodeProps> = ({
  nodeKey,
  value,
  depth = 0,
  changedPaths = new Set(),
  currentPath = "",
  expandAllSignal,
  shouldExpandAll,
}) => {
  const [isExpanded, setIsExpanded] = useState(false); // Start collapsed by default

  // Handle expand/collapse all signal
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
        return "#98c379"; // Green
      case "number":
        return "#61afef"; // Blue
      case "boolean":
        return "#d19a66"; // Orange
      case "null":
      case "undefined":
        return "#abb2bf"; // Gray
      case "object":
      case "array":
        return "#c678dd"; // Purple
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
        {isExpanded ? "▼" : "▶"}
      </span>
    );
  };

  const renderKey = () => {
    return (
      <span
        style={{
          color: "#e06c75",
          marginRight: "4px",
        }}
      >
        {nodeKey}:
      </span>
    );
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

  const handleCopy = () => {
    navigator.clipboard.writeText(JSON.stringify(value, null, 2));
  };

  return (
    <div
      className={`font-mono text-[13px] leading-[1.5] relative ${isChanged ? "[&>.treeNodeContent]:animate-[blink_1.5s_ease-out]" : ""}`}
    >
      <div
        className="treeNodeContent flex items-center py-0.5 px-1 rounded relative select-text"
        onDoubleClick={handleCopy}
      >
        {renderExpandToggle()}
        {renderKey()}
        {renderPrimitiveValue()}
        {renderObjectPreview()}
      </div>
      {renderExpandedContent()}
    </div>
  );
};

export default TreeNode;
