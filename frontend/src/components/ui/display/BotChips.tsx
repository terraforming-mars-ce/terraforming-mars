import React from "react";

export const PlayerChip: React.FC<{ children: React.ReactNode; className?: string }> = ({
  children,
  className = "",
}) => <span className={`player-chip ${className}`}>{children}</span>;

const personaLabels: Record<string, string> = {
  rival: "Rival",
  commentator: "Commentator",
  veteran: "Grumpy veteran",
  scientist: "Scientist",
  corporate: "Corporate",
  optimist: "Optimist",
};

export function getBotPersonaLabel(persona?: string): string {
  if (!persona) {
    return "Bot";
  }
  return personaLabels[persona] ?? persona;
}

interface BotPersonaChipProps {
  persona?: string;
  botStatus?: string;
}

export const BotPersonaChip: React.FC<BotPersonaChipProps> = ({ persona, botStatus }) => {
  const failed = botStatus === "failed";
  const loading = botStatus === "loading";
  let color = "bg-indigo-600/80";
  if (failed) {
    color = "bg-red-700/80";
  } else if (loading) {
    color = "bg-purple-700/80";
  }

  return (
    <PlayerChip className={`${color} text-white`}>
      {getBotPersonaLabel(persona)}
      {loading && (
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
      {failed && (
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
    </PlayerChip>
  );
};
