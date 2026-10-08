import { useEffect, useState } from "react";
import { getZIndex } from "@/constants/zIndex.ts";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

const ENABLED = new URLSearchParams(window.location.search).has("layoutdebug");

function measure() {
  const media = (query: string) => (window.matchMedia(query).matches ? "yes" : "no");
  return [
    ["pointer coarse", media("(pointer: coarse)")],
    ["pointer fine", media("(pointer: fine)")],
    ["any-pointer coarse", media("(any-pointer: coarse)")],
    ["hover", media("(hover: hover)")],
    ["portrait", media("(orientation: portrait)")],
    ["screen", `${window.screen.width} x ${window.screen.height}`],
    ["inner", `${window.innerWidth} x ${window.innerHeight}`],
    [
      "client",
      `${document.documentElement.clientWidth} x ${document.documentElement.clientHeight}`,
    ],
    [
      "visualViewport",
      window.visualViewport
        ? `${Math.round(window.visualViewport.width)} x ${Math.round(window.visualViewport.height)} @${window.visualViewport.scale.toFixed(2)}`
        : "none",
    ],
    ["dpr", String(window.devicePixelRatio)],
    ["ua", navigator.userAgent],
  ];
}

export default function LayoutDebugReadout() {
  const layout = useLayoutMode();
  const [rows, setRows] = useState(measure);

  useEffect(() => {
    if (!ENABLED) {
      return;
    }
    const update = () => setRows(measure());
    window.addEventListener("resize", update);
    window.visualViewport?.addEventListener("resize", update);
    return () => {
      window.removeEventListener("resize", update);
      window.visualViewport?.removeEventListener("resize", update);
    };
  }, []);

  if (!ENABLED) {
    return null;
  }

  return (
    <div
      className="fixed top-2 left-2 max-w-[calc(100vw-16px)] bg-black/85 border border-white/40 p-2 font-mono text-[11px] leading-snug text-white pointer-events-none"
      style={{ zIndex: getZIndex("ROTATE_DEVICE_OVERLAY", 1) }}
    >
      <div>
        mode: {layout.mode} (coarse {String(layout.isCoarsePointer)}, phoneScreen{" "}
        {String(layout.isPhoneScreen)}, smallViewport {String(layout.isSmallViewport)})
      </div>
      {rows.map(([label, value]) => (
        <div key={label} className="break-all">
          {label}: {value}
        </div>
      ))}
    </div>
  );
}
