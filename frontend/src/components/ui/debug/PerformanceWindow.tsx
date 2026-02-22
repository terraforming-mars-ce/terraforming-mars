import React, { useRef, useMemo } from "react";
import Sparkline from "./Sparkline.tsx";
import { usePerformanceMetrics } from "@/hooks/usePerformanceMetrics.ts";
import { useWindowDrag, useWindowManager } from "./WindowManager.tsx";

interface PerformanceWindowProps {
  isVisible: boolean;
  onClose: () => void;
}

const WINDOW_ID = "performance";
const WINDOW_WIDTH = 320;
const ACCENT = "#00d4aa";
const ACCENT_SHADOW = "rgba(0, 212, 170, 0.3)";
const EXCLUDE_SELECTORS = [".perf-content-area"];

const hasMemoryApi =
  typeof performance !== "undefined" &&
  "memory" in performance &&
  (performance as any).memory != null;

const PerformanceWindow: React.FC<PerformanceWindowProps> = ({ isVisible, onClose }) => {
  const dropdownRef = useRef<HTMLDivElement>(null);

  const { position, isDragging, handleMouseDown } = useWindowDrag({
    windowId: WINDOW_ID,
    width: WINDOW_WIDTH,
    height: () => window.innerHeight * 0.5,
    defaultPosition:
      typeof window !== "undefined"
        ? { x: window.innerWidth - WINDOW_WIDTH - 20, y: 60 }
        : undefined,
    excludeSelectors: EXCLUDE_SELECTORS,
    isVisible,
  });

  const { getZIndex } = useWindowManager();
  const { snapshots, latest } = usePerformanceMetrics();

  const fpsData = useMemo(() => snapshots.map((s) => s.fps), [snapshots]);
  const frameTimeData = useMemo(() => snapshots.map((s) => s.frameTimeMs), [snapshots]);
  const memoryData = useMemo(() => snapshots.map((s) => s.jsHeapUsedMB), [snapshots]);
  const drawCallData = useMemo(() => snapshots.map((s) => s.drawCalls), [snapshots]);

  if (!isVisible) return null;

  const sparklineWidth = WINDOW_WIDTH - 32;

  return (
    <div
      ref={dropdownRef}
      onMouseDown={handleMouseDown}
      style={{
        position: "fixed",
        top: `${position.y}px`,
        left: `${position.x}px`,
        width: `${WINDOW_WIDTH}px`,
        maxHeight: "50vh",
        background: "rgba(0, 0, 0, 0.95)",
        border: `2px solid ${ACCENT}`,
        borderRadius: "8px",
        padding: "12px 16px",
        zIndex: getZIndex(WINDOW_ID),
        overflow: "hidden",
        display: "flex",
        flexDirection: "column",
        boxShadow: `0 4px 20px ${ACCENT_SHADOW}`,
        cursor: isDragging ? "grabbing" : "default",
        transition: isDragging ? "none" : "top 0.2s ease-out, left 0.2s ease-out",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "10px",
          paddingBottom: "8px",
          borderBottom: "1px solid #333",
          userSelect: "none",
          cursor: "grab",
        }}
      >
        <h3
          style={{
            margin: 0,
            color: ACCENT,
            fontSize: "14px",
            display: "flex",
            alignItems: "center",
            gap: "8px",
          }}
          className="font-orbitron"
        >
          <svg
            width="10"
            height="14"
            viewBox="0 0 10 14"
            fill="currentColor"
            style={{ opacity: 0.5 }}
          >
            <circle cx="2" cy="2" r="1.5" />
            <circle cx="8" cy="2" r="1.5" />
            <circle cx="2" cy="7" r="1.5" />
            <circle cx="8" cy="7" r="1.5" />
            <circle cx="2" cy="12" r="1.5" />
            <circle cx="8" cy="12" r="1.5" />
          </svg>
          Performance
        </h3>
        <button
          onClick={onClose}
          onMouseDown={(e) => e.stopPropagation()}
          style={{
            background: "none",
            border: "none",
            color: "#abb2bf",
            fontSize: "18px",
            cursor: "pointer",
            padding: "0 4px",
            lineHeight: 1,
          }}
        >
          ×
        </button>
      </div>

      <div
        className="perf-content-area"
        style={{
          flex: 1,
          overflowY: "auto",
          overflowX: "hidden",
          display: "flex",
          flexDirection: "column",
          gap: "10px",
        }}
      >
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
          <div>
            <span className="font-orbitron" style={{ color: "#00ff88", fontSize: "24px" }}>
              {latest ? Math.round(latest.fps) : "—"}
            </span>
            <span style={{ color: "rgba(255,255,255,0.4)", fontSize: "11px", marginLeft: "4px" }}>
              FPS
            </span>
          </div>
          <div>
            <span className="font-orbitron" style={{ color: "#ffd700", fontSize: "24px" }}>
              {latest ? latest.frameTimeMs.toFixed(1) : "—"}
            </span>
            <span style={{ color: "rgba(255,255,255,0.4)", fontSize: "11px", marginLeft: "4px" }}>
              ms
            </span>
          </div>
        </div>

        <Sparkline
          data={fpsData}
          width={sparklineWidth}
          height={48}
          color="#00ff88"
          fillColor="rgba(0, 255, 136, 0.1)"
          min={0}
          label="FPS"
          currentValue={latest ? `${Math.round(latest.fps)}` : ""}
        />

        <Sparkline
          data={frameTimeData}
          width={sparklineWidth}
          height={48}
          color="#ffd700"
          fillColor="rgba(255, 215, 0, 0.1)"
          min={0}
          max={33}
          label="Frame Time"
          currentValue={latest ? `${latest.frameTimeMs.toFixed(1)}ms` : ""}
        />

        {hasMemoryApi && (
          <>
            <div
              style={{
                borderTop: "1px solid #222",
                paddingTop: "8px",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "baseline",
              }}
            >
              <span style={{ color: "rgba(255,255,255,0.4)", fontSize: "11px" }}>JS Heap</span>
              <span style={{ color: "#00ccff", fontSize: "12px" }} className="font-orbitron">
                {latest
                  ? `${latest.jsHeapUsedMB.toFixed(0)} / ${latest.jsHeapTotalMB.toFixed(0)} MB`
                  : "—"}
              </span>
            </div>
            <Sparkline
              data={memoryData}
              width={sparklineWidth}
              height={40}
              color="#00ccff"
              fillColor="rgba(0, 204, 255, 0.1)"
              label="Memory"
              currentValue={latest ? `${latest.jsHeapUsedMB.toFixed(0)}MB` : ""}
            />
          </>
        )}

        <div
          style={{
            borderTop: "1px solid #222",
            paddingTop: "8px",
            display: "flex",
            flexDirection: "column",
            gap: "6px",
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              fontSize: "11px",
            }}
          >
            <span style={{ color: "rgba(255,255,255,0.4)" }}>Draw Calls</span>
            <span style={{ color: "#ff8844" }} className="font-orbitron">
              {latest?.drawCalls ?? "—"}
            </span>
          </div>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              fontSize: "11px",
            }}
          >
            <span style={{ color: "rgba(255,255,255,0.4)" }}>Triangles</span>
            <span style={{ color: "#ff8844" }} className="font-orbitron">
              {latest ? latest.triangles.toLocaleString() : "—"}
            </span>
          </div>
          <Sparkline
            data={drawCallData}
            width={sparklineWidth}
            height={40}
            color="#ff8844"
            fillColor="rgba(255, 136, 68, 0.1)"
          />
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              fontSize: "10px",
              color: "rgba(255,255,255,0.3)",
            }}
          >
            <span>Textures: {latest?.textureCount ?? "—"}</span>
            <span>Geometries: {latest?.geometryCount ?? "—"}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PerformanceWindow;
