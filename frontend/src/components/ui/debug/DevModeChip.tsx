import React, { useState } from "react";
import { Z_INDEX } from "../../../constants/zIndex";

const DevModeChip: React.FC = () => {
  const [isHovered, setIsHovered] = useState(false);

  const handleClick = () => {
    window.dispatchEvent(new CustomEvent("toggle-debug-dropdown"));
  };

  return (
    <button
      onClick={handleClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className="fixed top-0 left-1/2 -translate-x-1/2 bg-[#ff6b35] text-white text-[10px] font-bold py-1 px-3 rounded-b-lg border border-[#e55a2e] border-t-0 whitespace-nowrap shadow-[0_2px_4px_rgba(0,0,0,0.3)] cursor-pointer hover:bg-[#ff7f4d] transition-all duration-200"
      style={{
        minWidth: "120px",
        textAlign: "center",
        zIndex: Z_INDEX.TOP_MENU_ALWAYS_ON_TOP,
      }}
    >
      <span className="transition-opacity duration-200" style={{ opacity: isHovered ? 0 : 1 }}>
        DEV MODE
      </span>
      <span
        className="absolute inset-0 flex items-center justify-center transition-opacity duration-200"
        style={{ opacity: isHovered ? 1 : 0 }}
      >
        OPEN ADMIN TOOLS
      </span>
    </button>
  );
};

export default DevModeChip;
