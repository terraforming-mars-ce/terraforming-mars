import { FC, ReactNode } from "react";
import { Z_INDEX } from "@/constants/zIndex.ts";

type TooltipSize = "small" | "medium";

interface InfoTooltipProps {
  /** The tooltip content to display on hover */
  children: ReactNode;
  /** Size variant affects icon size and tooltip width */
  size?: TooltipSize;
}

const sizeConfig = {
  small: {
    icon: "text-[10px] w-[14px] h-[14px]",
    tooltip: "w-[260px] text-[12px]",
    shadow:
      "shadow-[0_0_8px_rgba(30,60,150,0.2)] group-hover:shadow-[0_0_12px_rgba(30,60,150,0.4)]",
  },
  medium: {
    icon: "text-[11px] w-[16px] h-[16px]",
    tooltip: "w-[280px] text-[13px]",
    shadow:
      "shadow-[0_0_10px_rgba(30,60,150,0.2)] group-hover:shadow-[0_0_15px_rgba(30,60,150,0.4)]",
  },
};

const InfoTooltip: FC<InfoTooltipProps> = ({ children, size = "medium" }) => {
  const config = sizeConfig[size];

  return (
    <div className="relative inline-block group">
      <span
        className={`text-space-blue-solid cursor-help flex items-center justify-center rounded-full bg-space-blue-100 border border-space-blue-400 transition-all duration-200 group-hover:bg-space-blue-200 ${config.icon} ${config.shadow}`}
      >
        <span className="font-serif italic">i</span>
      </span>
      <div
        className={`invisible opacity-0 bg-space-black/[0.98] text-white text-left rounded-lg p-3 absolute bottom-[125%] right-0 leading-normal border border-space-blue-400 shadow-glow transition-all duration-300 group-hover:visible group-hover:opacity-100 after:content-[''] after:absolute after:top-full after:right-3 after:border-8 after:border-solid after:border-t-space-black/[0.98] after:border-r-transparent after:border-b-transparent after:border-l-transparent ${config.tooltip}`}
        style={{ zIndex: Z_INDEX.MENU_DROPDOWN }}
      >
        {children}
      </div>
    </div>
  );
};

export default InfoTooltip;
