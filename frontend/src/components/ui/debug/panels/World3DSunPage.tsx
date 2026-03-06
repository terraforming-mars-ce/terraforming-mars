import React, { useRef, useEffect, useCallback } from "react";
import { useWorld3DSettings } from "../../../../contexts/World3DSettingsContext";
import { ColorSwatch } from "../HSVColorPicker.tsx";

const SUN_PREVIEW_SIZE = 90;

const CAM_PITCH = 0.3;
const COS_P = Math.cos(CAM_PITCH);
const SIN_P = Math.sin(CAM_PITCH);

function project3D(x: number, y: number, z: number, cx: number, cy: number, scale: number) {
  const ry = y * COS_P - z * SIN_P;
  const rz = y * SIN_P + z * COS_P;
  return { px: cx + x * scale, py: cy - ry * scale, depth: rz };
}

const SunPreview: React.FC<{ dirX: number; dirY: number; dirZ: number }> = ({
  dirX,
  dirY,
  dirZ,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const w = SUN_PREVIEW_SIZE;
    const h = SUN_PREVIEW_SIZE;
    const cx = w / 2;
    const cy = h / 2;
    const sphereRadius = 1.0;
    const scale = 26;
    const screenRadius = sphereRadius * scale;

    ctx.clearRect(0, 0, w, h);

    const len = Math.sqrt(dirX * dirX + dirY * dirY + dirZ * dirZ) || 1;
    const lx = dirX / len;
    const ly = dirY / len;
    const lz = dirZ / len;

    let upX = 0,
      upY = 1,
      upZ = 0;
    if (Math.abs(ly) > 0.9) {
      upX = 1;
      upY = 0;
      upZ = 0;
    }
    const t1x = ly * upZ - lz * upY;
    const t1y = lz * upX - lx * upZ;
    const t1z = lx * upY - ly * upX;
    const t1Len = Math.sqrt(t1x * t1x + t1y * t1y + t1z * t1z) || 1;
    const ax = t1x / t1Len,
      ay = t1y / t1Len,
      az = t1z / t1Len;
    const bx = ly * az - lz * ay;
    const by = lz * ax - lx * az;
    const bz = lx * ay - ly * ax;

    const arrowDist = 1.8;
    const arrowLen3D = 0.55;
    const gridSpacing = 0.55;
    const gridOffsets = [-1, 0, 1];

    type Arrow3D = { sx: number; sy: number; sz: number; ex: number; ey: number; ez: number };
    const arrows: Arrow3D[] = [];
    for (const gi of gridOffsets) {
      for (const gj of gridOffsets) {
        const ox = lx * arrowDist + ax * gi * gridSpacing + bx * gj * gridSpacing;
        const oy = ly * arrowDist + ay * gi * gridSpacing + by * gj * gridSpacing;
        const oz = lz * arrowDist + az * gi * gridSpacing + bz * gj * gridSpacing;
        arrows.push({
          sx: ox,
          sy: oy,
          sz: oz,
          ex: ox - lx * arrowLen3D,
          ey: oy - ly * arrowLen3D,
          ez: oz - lz * arrowLen3D,
        });
      }
    }

    const drawArrow = (a: Arrow3D) => {
      const s = project3D(a.sx, a.sy, a.sz, cx, cy, scale);
      const e = project3D(a.ex, a.ey, a.ez, cx, cy, scale);

      ctx.beginPath();
      ctx.moveTo(s.px, s.py);
      ctx.lineTo(e.px, e.py);
      ctx.stroke();

      const angle = Math.atan2(e.py - s.py, e.px - s.px);
      const headLen = 4;
      ctx.beginPath();
      ctx.moveTo(e.px, e.py);
      ctx.lineTo(e.px - headLen * Math.cos(angle - 0.5), e.py - headLen * Math.sin(angle - 0.5));
      ctx.lineTo(e.px - headLen * Math.cos(angle + 0.5), e.py - headLen * Math.sin(angle + 0.5));
      ctx.closePath();
      ctx.fill();
    };

    const behindArrows: Arrow3D[] = [];
    const frontArrows: Arrow3D[] = [];
    for (const a of arrows) {
      const midDepth = project3D(
        (a.sx + a.ex) / 2,
        (a.sy + a.ey) / 2,
        (a.sz + a.ez) / 2,
        cx,
        cy,
        scale,
      ).depth;
      if (midDepth < 0) {
        behindArrows.push(a);
      } else {
        frontArrows.push(a);
      }
    }

    ctx.strokeStyle = "rgba(251, 191, 36, 0.3)";
    ctx.fillStyle = "rgba(251, 191, 36, 0.3)";
    ctx.lineWidth = 1;
    for (const a of behindArrows) {
      drawArrow(a);
    }

    const imgData = ctx.createImageData(w, h);
    const aaMargin = 1.0;
    const outerR = screenRadius + aaMargin;
    for (let py = 0; py < h; py++) {
      for (let px = 0; px < w; px++) {
        const dx = px - cx;
        const dy = py - cy;
        const dist = Math.sqrt(dx * dx + dy * dy);
        if (dist > outerR) continue;

        const sx = dx / screenRadius;
        const sy = -dy / screenRadius;
        const r2 = sx * sx + sy * sy;

        let alpha = 1;
        if (r2 > 1) continue;
        if (r2 > (1 - aaMargin / screenRadius) ** 2) {
          alpha = Math.max(0, ((1 - Math.sqrt(r2)) * screenRadius) / aaMargin);
        }

        const sz = Math.sqrt(Math.max(0, 1 - r2));

        const wy = sy * COS_P + sz * SIN_P;
        const wz = -sy * SIN_P + sz * COS_P;

        const dot = Math.max(0, sx * lx + wy * ly + wz * lz);
        const ambient = 0.12;
        const brightness = Math.min(1, ambient + dot * 0.88);

        const idx = (py * w + px) * 4;
        imgData.data[idx] = Math.round(150 * brightness);
        imgData.data[idx + 1] = Math.round(140 * brightness);
        imgData.data[idx + 2] = Math.round(130 * brightness);
        imgData.data[idx + 3] = Math.round(255 * alpha);
      }
    }
    ctx.putImageData(imgData, 0, 0);

    ctx.strokeStyle = "#fbbf24";
    ctx.fillStyle = "#fbbf24";
    ctx.lineWidth = 1.5;
    for (const a of frontArrows) {
      drawArrow(a);
    }
  }, [dirX, dirY, dirZ]);

  useEffect(() => {
    draw();
  }, [draw]);

  return (
    <canvas
      ref={canvasRef}
      width={SUN_PREVIEW_SIZE}
      height={SUN_PREVIEW_SIZE}
      style={{ display: "block" }}
    />
  );
};

const World3DSunPage: React.FC = () => {
  const { settings, updateSettings, resetSettings } = useWorld3DSettings();

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
          Sun Direction
        </label>
        <div style={{ display: "flex", gap: "12px", alignItems: "center" }}>
          <div style={{ display: "flex", flexDirection: "column", gap: "8px", flex: 1 }}>
            {(["X", "Y", "Z"] as const).map((axis) => {
              const key = `sunDirection${axis}` as
                | "sunDirectionX"
                | "sunDirectionY"
                | "sunDirectionZ";
              return (
                <div key={axis} style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                  <span style={{ color: "#abb2bf", fontSize: "11px", width: "20px" }}>{axis}:</span>
                  <input
                    type="range"
                    min="-1"
                    max="1"
                    step="0.01"
                    value={settings[key]}
                    onChange={(e) => updateSettings({ [key]: parseFloat(e.target.value) })}
                    style={{ flex: 1, accentColor: "#3b82f6" }}
                  />
                  <span
                    style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}
                  >
                    {settings[key].toFixed(2)}
                  </span>
                </div>
              );
            })}
          </div>
          <div style={{ paddingRight: "8px" }}>
            <SunPreview
              dirX={settings.sunDirectionX}
              dirY={settings.sunDirectionY}
              dirZ={settings.sunDirectionZ}
            />
          </div>
        </div>
      </div>

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
          Sun Intensity
        </label>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <input
            type="range"
            min="0"
            max="3"
            step="0.1"
            value={settings.sunIntensity}
            onChange={(e) => updateSettings({ sunIntensity: parseFloat(e.target.value) })}
            style={{ flex: 1, accentColor: "#3b82f6" }}
          />
          <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
            {settings.sunIntensity.toFixed(1)}
          </span>
        </div>
      </div>

      <div style={{ marginBottom: "16px" }}>
        <ColorSwatch
          label="Sun Color"
          color={settings.sunColor}
          onChange={(c) => updateSettings({ sunColor: c })}
        />
      </div>

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
          Fresnel Reflectance (rf0)
        </label>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <input
            type="range"
            min="0"
            max="1"
            step="0.01"
            value={settings.reflectance}
            onChange={(e) => updateSettings({ reflectance: parseFloat(e.target.value) })}
            style={{ flex: 1, accentColor: "#3b82f6" }}
          />
          <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
            {settings.reflectance.toFixed(2)}
          </span>
        </div>
      </div>

      <div style={{ marginBottom: "16px" }}>
        <ColorSwatch
          label="Water Color"
          color={settings.waterColor}
          onChange={(c) => updateSettings({ waterColor: c })}
        />
      </div>

      <button
        onClick={resetSettings}
        style={{
          padding: "8px 16px",
          background: "linear-gradient(135deg, rgba(59, 130, 246, 0.8), rgba(59, 130, 246, 0.6))",
          border: "1px solid rgba(59, 130, 246, 0.5)",
          borderRadius: "6px",
          color: "white",
          fontSize: "12px",
          cursor: "pointer",
          fontWeight: "500",
        }}
      >
        Reset to Defaults
      </button>
    </div>
  );
};

export default World3DSunPage;
