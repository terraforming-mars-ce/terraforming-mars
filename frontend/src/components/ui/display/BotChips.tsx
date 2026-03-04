import React from "react";

const difficultyConfig: Record<string, { label: string; color: string }> = {
  hard: { label: "Hard", color: "bg-amber-600/80" },
  extreme: { label: "Actual Bot", color: "bg-red-700/80" },
  normal: { label: "Normal", color: "bg-green-700/80" },
};

const speedConfig: Record<string, { label: string; color: string }> = {
  fast: { label: "Haiku", color: "bg-sky-600/80" },
  thinker: { label: "Opus", color: "bg-violet-600/80" },
};

interface BotDifficultyChipProps {
  difficulty?: string;
  botStatus?: string;
  showStatusIcon?: boolean;
}

export const BotDifficultyChip: React.FC<BotDifficultyChipProps> = ({
  difficulty = "normal",
  botStatus,
  showStatusIcon = false,
}) => {
  const failed = botStatus === "failed";
  const loading = botStatus === "loading";
  const config = difficultyConfig[difficulty] || difficultyConfig.normal;
  const color = failed ? "bg-red-700/80" : loading ? "bg-purple-700/80" : config.color;

  return (
    <span
      className={`${color} text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase flex items-center gap-1`}
    >
      {config.label}
      {showStatusIcon && loading && (
        <svg
          className="animate-spin"
          width="10"
          height="10"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
        >
          <path d="M12 2a10 10 0 0 1 10 10" strokeLinecap="round" />
        </svg>
      )}
      {showStatusIcon && failed && (
        <svg
          width="10"
          height="10"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      )}
    </span>
  );
};

interface BotSpeedChipProps {
  speed?: string;
}

export const BotSpeedChip: React.FC<BotSpeedChipProps> = ({ speed = "fast" }) => {
  const config = speedConfig[speed] || speedConfig.fast;

  return (
    <span
      className={`${config.color} text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase`}
    >
      {config.label}
    </span>
  );
};
