import React, { useRef, useEffect, useState, useCallback } from "react";
import { createPortal } from "react-dom";
import { Z_INDEX } from "../../../constants/zIndex.ts";

interface RGBColor {
  r: number;
  g: number;
  b: number;
}

interface HSVColor {
  h: number;
  s: number;
  v: number;
}

function rgbToHsv(r: number, g: number, b: number): HSVColor {
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const d = max - min;
  let h = 0;
  const s = max === 0 ? 0 : d / max;
  const v = max;

  if (d !== 0) {
    if (max === r) {
      h = ((g - b) / d + (g < b ? 6 : 0)) / 6;
    } else if (max === g) {
      h = ((b - r) / d + 2) / 6;
    } else {
      h = ((r - g) / d + 4) / 6;
    }
  }

  return { h: h * 360, s, v };
}

function hsvToRgb(h: number, s: number, v: number): RGBColor {
  const c = v * s;
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
  const m = v - c;
  let r = 0,
    g = 0,
    b = 0;

  if (h < 60) {
    r = c;
    g = x;
  } else if (h < 120) {
    r = x;
    g = c;
  } else if (h < 180) {
    g = c;
    b = x;
  } else if (h < 240) {
    g = x;
    b = c;
  } else if (h < 300) {
    r = x;
    b = c;
  } else {
    r = c;
    b = x;
  }

  return { r: r + m, g: g + m, b: b + m };
}

const SV_WIDTH = 200;
const SV_HEIGHT = 150;

interface HSVColorPickerProps {
  color: RGBColor;
  onChange: (color: RGBColor) => void;
  isOpen: boolean;
  onClose: () => void;
  anchorRect: DOMRect | null;
}

const HSVColorPicker: React.FC<HSVColorPickerProps> = ({
  color,
  onChange,
  isOpen,
  onClose,
  anchorRect,
}) => {
  const svCanvasRef = useRef<HTMLCanvasElement>(null);
  const hueCanvasRef = useRef<HTMLCanvasElement>(null);
  const popoverRef = useRef<HTMLDivElement>(null);
  const [hsv, setHsv] = useState<HSVColor>(() => rgbToHsv(color.r, color.g, color.b));
  const [isDraggingSV, setIsDraggingSV] = useState(false);
  const [isDraggingHue, setIsDraggingHue] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setHsv(rgbToHsv(color.r, color.g, color.b));
    }
  }, [isOpen]);

  const drawSVCanvas = useCallback(() => {
    const canvas = svCanvasRef.current;
    if (!canvas) {
      return;
    }
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      return;
    }

    for (let x = 0; x < SV_WIDTH; x++) {
      for (let y = 0; y < SV_HEIGHT; y++) {
        const s = x / SV_WIDTH;
        const v = 1 - y / SV_HEIGHT;
        const rgb = hsvToRgb(hsv.h, s, v);
        ctx.fillStyle = `rgb(${Math.round(rgb.r * 255)},${Math.round(rgb.g * 255)},${Math.round(rgb.b * 255)})`;
        ctx.fillRect(x, y, 1, 1);
      }
    }
  }, [hsv.h]);

  const drawHueCanvas = useCallback(() => {
    const canvas = hueCanvasRef.current;
    if (!canvas) {
      return;
    }
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      return;
    }

    const gradient = ctx.createLinearGradient(0, 0, SV_WIDTH, 0);
    for (let i = 0; i <= 6; i++) {
      const rgb = hsvToRgb((i / 6) * 360, 1, 1);
      gradient.addColorStop(
        i / 6,
        `rgb(${Math.round(rgb.r * 255)},${Math.round(rgb.g * 255)},${Math.round(rgb.b * 255)})`,
      );
    }
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, SV_WIDTH, 16);
  }, []);

  useEffect(() => {
    if (isOpen) {
      drawSVCanvas();
      drawHueCanvas();
    }
  }, [isOpen, drawSVCanvas, drawHueCanvas]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (popoverRef.current && !popoverRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside, true);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside, true);
  }, [isOpen, onClose]);

  useEffect(() => {
    const handleMouseUp = () => {
      setIsDraggingSV(false);
      setIsDraggingHue(false);
    };

    const handleMouseMove = (e: MouseEvent) => {
      if (isDraggingSV) {
        updateSV(e);
      }
      if (isDraggingHue) {
        updateHue(e);
      }
    };

    if (isDraggingSV || isDraggingHue) {
      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
    }
    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isDraggingSV, isDraggingHue]);

  const updateSV = (e: MouseEvent) => {
    const canvas = svCanvasRef.current;
    if (!canvas) {
      return;
    }
    const rect = canvas.getBoundingClientRect();
    const x = Math.max(0, Math.min(SV_WIDTH, e.clientX - rect.left));
    const y = Math.max(0, Math.min(SV_HEIGHT, e.clientY - rect.top));
    const s = x / SV_WIDTH;
    const v = 1 - y / SV_HEIGHT;
    const newHsv = { ...hsv, s, v };
    setHsv(newHsv);
    onChange(hsvToRgb(newHsv.h, newHsv.s, newHsv.v));
  };

  const updateHue = (e: MouseEvent) => {
    const canvas = hueCanvasRef.current;
    if (!canvas) {
      return;
    }
    const rect = canvas.getBoundingClientRect();
    const x = Math.max(0, Math.min(SV_WIDTH, e.clientX - rect.left));
    const h = (x / SV_WIDTH) * 360;
    const newHsv = { ...hsv, h };
    setHsv(newHsv);
    onChange(hsvToRgb(newHsv.h, newHsv.s, newHsv.v));
  };

  if (!isOpen || !anchorRect) {
    return null;
  }

  const popoverStyle: React.CSSProperties = {
    position: "fixed",
    top: anchorRect.bottom + 4,
    left: anchorRect.left,
    zIndex: Z_INDEX.DEBUG_WINDOWS + 10,
    background: "rgba(20, 20, 20, 0.98)",
    border: "1px solid rgba(59, 130, 246, 0.5)",
    borderRadius: "8px",
    padding: "12px",
    boxShadow: "0 8px 24px rgba(0, 0, 0, 0.8)",
  };

  const svMarkerX = hsv.s * SV_WIDTH;
  const svMarkerY = (1 - hsv.v) * SV_HEIGHT;
  const hueMarkerX = (hsv.h / 360) * SV_WIDTH;

  return createPortal(
    <div ref={popoverRef} style={popoverStyle} onMouseDown={(e) => e.stopPropagation()}>
      <div style={{ position: "relative", marginBottom: "8px" }}>
        <canvas
          ref={svCanvasRef}
          width={SV_WIDTH}
          height={SV_HEIGHT}
          style={{ cursor: "pointer", borderRadius: "4px", display: "block" }}
          onMouseDown={(e) => {
            setIsDraggingSV(true);
            updateSV(e.nativeEvent);
          }}
        />
        <div
          style={{
            position: "absolute",
            left: svMarkerX - 6,
            top: svMarkerY - 6,
            width: "12px",
            height: "12px",
            borderRadius: "50%",
            border: "2px solid white",
            boxShadow: "0 0 2px rgba(0,0,0,0.8)",
            pointerEvents: "none",
          }}
        />
      </div>

      <div style={{ position: "relative" }}>
        <canvas
          ref={hueCanvasRef}
          width={SV_WIDTH}
          height={16}
          style={{ cursor: "pointer", borderRadius: "4px", display: "block" }}
          onMouseDown={(e) => {
            setIsDraggingHue(true);
            updateHue(e.nativeEvent);
          }}
        />
        <div
          style={{
            position: "absolute",
            left: hueMarkerX - 3,
            top: -2,
            width: "6px",
            height: "20px",
            borderRadius: "3px",
            border: "2px solid white",
            boxShadow: "0 0 2px rgba(0,0,0,0.8)",
            pointerEvents: "none",
          }}
        />
      </div>
    </div>,
    document.body,
  );
};

interface ColorSwatchProps {
  color: RGBColor;
  onChange: (color: RGBColor) => void;
  label: string;
}

export const ColorSwatch: React.FC<ColorSwatchProps> = ({ color, onChange, label }) => {
  const [isOpen, setIsOpen] = useState(false);
  const swatchRef = useRef<HTMLDivElement>(null);
  const [anchorRect, setAnchorRect] = useState<DOMRect | null>(null);

  const handleClick = () => {
    if (swatchRef.current) {
      setAnchorRect(swatchRef.current.getBoundingClientRect());
    }
    setIsOpen(!isOpen);
  };

  return (
    <>
      <div style={{ display: "flex", alignItems: "center", gap: "8px", marginBottom: "4px" }}>
        <label style={{ color: "#abb2bf", fontSize: "12px", minWidth: "90px" }}>{label}</label>
        <div
          ref={swatchRef}
          onClick={handleClick}
          style={{
            width: "32px",
            height: "20px",
            borderRadius: "4px",
            background: `rgb(${Math.round(color.r * 255)}, ${Math.round(color.g * 255)}, ${Math.round(color.b * 255)})`,
            border: "1px solid rgba(59, 130, 246, 0.4)",
            cursor: "pointer",
          }}
        />
        <span style={{ color: "#666", fontSize: "10px" }}>
          {Math.round(color.r * 255)}, {Math.round(color.g * 255)}, {Math.round(color.b * 255)}
        </span>
      </div>
      <HSVColorPicker
        color={color}
        onChange={onChange}
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        anchorRect={anchorRect}
      />
    </>
  );
};

export default HSVColorPicker;
