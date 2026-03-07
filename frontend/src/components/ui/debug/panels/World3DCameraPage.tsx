import React, { useState, useEffect, useCallback, useRef } from "react";
import { useWorld3DSettings } from "../../../../contexts/World3DSettingsContext";

const DRAG_PIXELS_PER_STEP = 20;
const POLL_INTERVAL = 100;

const inputStyle = {
  width: "100%",
  padding: "5px 8px",
  background: "rgba(0, 0, 0, 0.8)",
  border: "1px solid rgba(59, 130, 246, 0.3)",
  borderRadius: "4px",
  color: "white",
  fontSize: "11px",
  outline: "none",
  boxSizing: "border-box" as const,
  cursor: "text" as const,
};

const DraggableFloatInput: React.FC<{
  value: string;
  onChange: (v: string) => void;
  onCommit: (v: number) => void;
  step?: number;
}> = ({ value, onChange, onCommit, step = 0.1 }) => {
  const dragRef = useRef<{ startX: number; startVal: number; dragging: boolean } | null>(null);
  const valueRef = useRef(value);
  valueRef.current = value;

  const handleMouseDown = useCallback(
    (e: React.MouseEvent<HTMLInputElement>) => {
      const input = e.currentTarget;
      const currentVal = parseFloat(valueRef.current) || 0;
      dragRef.current = { startX: e.clientX, startVal: currentVal, dragging: false };

      const onMouseMove = (me: MouseEvent) => {
        if (!dragRef.current) return;
        const dx = me.clientX - dragRef.current.startX;
        if (!dragRef.current.dragging && Math.abs(dx) < 4) return;
        if (!dragRef.current.dragging) {
          dragRef.current.dragging = true;
          document.body.style.cursor = "ew-resize";
          document.body.style.userSelect = "none";
          input.blur();
        }
        const steps = Math.round(dx / DRAG_PIXELS_PER_STEP);
        const newVal = parseFloat((dragRef.current.startVal + steps * step).toFixed(3));
        onChange(newVal.toString());
        onCommit(newVal);
      };

      const onMouseUp = () => {
        dragRef.current = null;
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
        document.body.style.cursor = "";
        document.body.style.userSelect = "";
      };

      document.addEventListener("mousemove", onMouseMove);
      document.addEventListener("mouseup", onMouseUp);
    },
    [onChange, onCommit, step],
  );

  const handleBlur = () => {
    const n = parseFloat(value);
    if (!isNaN(n)) {
      onCommit(n);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      const n = parseFloat(value);
      if (!isNaN(n)) {
        onCommit(n);
      }
    }
  };

  return (
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      onBlur={handleBlur}
      onKeyDown={handleKeyDown}
      onMouseDown={handleMouseDown}
      spellCheck={false}
      autoComplete="off"
      style={inputStyle}
    />
  );
};

const World3DCameraPage: React.FC = () => {
  const { settings, updateSettings, cameraStateRef, pendingCameraTransformRef } =
    useWorld3DSettings();

  const [posX, setPosX] = useState("0.00");
  const [posY, setPosY] = useState("0.00");
  const [posZ, setPosZ] = useState("8.00");

  const isEditingRef = useRef(false);
  const editTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const markEditing = () => {
    isEditingRef.current = true;
    if (editTimeoutRef.current) {
      clearTimeout(editTimeoutRef.current);
    }
    editTimeoutRef.current = setTimeout(() => {
      isEditingRef.current = false;
    }, 500);
  };

  useEffect(() => {
    const interval = setInterval(() => {
      if (isEditingRef.current) return;
      const cam = cameraStateRef.current;
      setPosX(cam.position.x.toFixed(2));
      setPosY(cam.position.y.toFixed(2));
      setPosZ(cam.position.z.toFixed(2));
    }, POLL_INTERVAL);
    return () => clearInterval(interval);
  }, [cameraStateRef]);

  const commitPosition = (axis: "x" | "y" | "z", val: number) => {
    markEditing();
    const pos = { ...cameraStateRef.current.position, [axis]: val };
    pendingCameraTransformRef.current = { position: pos };
  };

  const posSetters = { x: setPosX, y: setPosY, z: setPosZ };
  const posValues = { x: posX, y: posY, z: posZ };

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
          Camera Controls
        </label>
        <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          <label style={{ display: "flex", alignItems: "center", gap: "8px", cursor: "pointer" }}>
            <input
              type="checkbox"
              checked={settings.freeCameraEnabled}
              onChange={(e) => updateSettings({ freeCameraEnabled: e.target.checked })}
              style={{ accentColor: "#3b82f6" }}
            />
            <span style={{ color: "#fff", fontSize: "12px" }}>Mars Mode</span>
          </label>
          {settings.freeCameraEnabled && (
            <label
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                cursor: "pointer",
                marginLeft: "16px",
              }}
            >
              <input
                type="checkbox"
                checked={settings.showCameraFrustum}
                onChange={(e) => updateSettings({ showCameraFrustum: e.target.checked })}
                style={{ accentColor: "#3b82f6" }}
              />
              <span style={{ color: "#fff", fontSize: "12px" }}>Show Game Camera Frustum</span>
            </label>
          )}
        </div>
      </div>

      <div style={{ marginBottom: "16px", borderTop: "1px solid #333", paddingTop: "16px" }}>
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
          Position
        </label>
        <div style={{ display: "flex", gap: "8px" }}>
          {(["x", "y", "z"] as const).map((axis) => (
            <div key={axis} style={{ flex: 1 }}>
              <span style={{ color: "#abb2bf", fontSize: "10px", textTransform: "uppercase" }}>
                {axis}
              </span>
              <DraggableFloatInput
                value={posValues[axis]}
                onChange={(v) => {
                  markEditing();
                  posSetters[axis](v);
                }}
                onCommit={(val) => commitPosition(axis, val)}
                step={0.05}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default World3DCameraPage;
