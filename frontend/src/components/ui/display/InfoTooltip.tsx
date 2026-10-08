import { type FC, type ReactNode } from "react";
import RevealTrigger from "./RevealTrigger.tsx";

interface InfoTooltipProps {
  children: ReactNode;
  size?: "small" | "medium";
}

export const InfoGlyph: FC = () => (
  <svg width="14" height="14" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <path d="M1.5 1.5h10l3 3v10h-10l-3-3z" stroke="currentColor" />
    <path d="M8 7v4M8 4.5v.5" stroke="currentColor" strokeWidth="1.5" />
  </svg>
);

const InfoTooltip: FC<InfoTooltipProps> = ({ children, size = "medium" }) => (
  <span className="inline-flex align-middle">
    <RevealTrigger
      content={<span className="font-sans text-xs font-normal leading-relaxed">{children}</span>}
      maxWidth={size === "small" ? 260 : 280}
      cornerSize={8}
      tabIndex={0}
      aria-label="More information"
      className={`inline-flex shrink-0 items-center justify-center cursor-default text-white/50 hover:text-white focus-visible:text-white focus-visible:outline focus-visible:outline-1 focus-visible:outline-white/60 ${size === "small" ? "size-4" : "size-5"}`}
    >
      <InfoGlyph />
    </RevealTrigger>
  </span>
);

export default InfoTooltip;
