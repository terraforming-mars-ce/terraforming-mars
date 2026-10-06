import FloatingWindow from "./FloatingWindow.tsx";
import React, { useMemo, useState } from "react";
import Sparkline from "./Sparkline.tsx";
import { usePerformanceMetrics } from "@/hooks/usePerformanceMetrics.ts";
import { performanceStore } from "@/services/performanceStore.ts";
import { useWindowDrag, useWindowManager } from "./WindowManager.tsx";

interface PerformanceWindowProps {
  isVisible: boolean;
  onClose: () => void;
}

const WINDOW_ID = "performance";
const WINDOW_WIDTH = 380;
const EXCLUDE_SELECTORS = [".perf-content-area"];

const hasMemoryApi =
  typeof performance !== "undefined" &&
  "memory" in performance &&
  (performance as any).memory != null;

const PerformanceWindow: React.FC<PerformanceWindowProps> = ({ isVisible, onClose }) => {
  const { position, handleMouseDown } = useWindowDrag({
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
    <FloatingWindow
      title="Performance"
      onClose={onClose}
      onMouseDown={handleMouseDown}
      style={{
        top: position.y,
        left: position.x,
        width: WINDOW_WIDTH,
        maxHeight: "80vh",
        zIndex: getZIndex(WINDOW_ID),
      }}
    >
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

        <SceneBreakdown />
        <StatSections />
      </div>
    </FloatingWindow>
  );
};

const sectionStyle: React.CSSProperties = {
  borderTop: "1px solid #222",
  paddingTop: "8px",
  display: "flex",
  flexDirection: "column",
  gap: "3px",
  fontSize: "10px",
};
const headingStyle: React.CSSProperties = {
  color: "rgba(255,255,255,0.55)",
  fontSize: "11px",
  marginBottom: "2px",
};
const rowStyle: React.CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "8px",
};
const labelStyle: React.CSSProperties = { color: "rgba(255,255,255,0.4)" };
const valueStyle: React.CSSProperties = { color: "#ddd", textAlign: "right" };

// Estimated per-group submissions (upper bound: frustum culling is not applied). Clicking a row
// hides that group from the render, for A/B measurements.
const SceneBreakdown: React.FC = () => {
  const [, setRevision] = useState(0);
  const census = performanceStore.getCensus();
  const hidden = [...performanceStore.hiddenGroups].filter(
    (group) => !census.some((entry) => entry.group === group),
  );
  if (!census.length && !hidden.length) {
    return null;
  }
  const toggle = (group: string) => {
    performanceStore.toggleGroup(group);
    setRevision((revision) => revision + 1);
  };
  return (
    <div style={sectionStyle}>
      <div style={{ ...rowStyle, ...headingStyle }}>
        <span>Scene breakdown</span>
        <span>calls · triangles</span>
      </div>
      {census.map((group) => (
        <div
          key={group.group}
          className="cursor-pointer"
          style={rowStyle}
          onClick={() => toggle(group.group)}
        >
          <span style={labelStyle}>{group.group}</span>
          <span style={valueStyle} className="font-orbitron">
            {group.drawCalls} · {group.triangles.toLocaleString()}
          </span>
        </div>
      ))}
      {hidden.map((group) => (
        <div
          key={group}
          className="cursor-pointer"
          style={{ ...rowStyle, opacity: 0.35 }}
          onClick={() => toggle(group)}
        >
          <span style={labelStyle}>{group}</span>
          <span style={valueStyle}>hidden</span>
        </div>
      ))}
    </div>
  );
};

const StatSections: React.FC = () => {
  const sections = [...performanceStore.getSections()];
  return (
    <>
      {sections.map(([name, values]) => (
        <div key={name} style={sectionStyle}>
          <div style={headingStyle}>{name}</div>
          {Object.entries(values).map(([key, value]) => (
            <div key={key} style={rowStyle}>
              <span style={labelStyle}>{key}</span>
              <span style={valueStyle}>{value}</span>
            </div>
          ))}
        </div>
      ))}
    </>
  );
};

export default PerformanceWindow;
