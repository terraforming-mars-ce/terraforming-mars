import React from "react";
import { GameDto } from "@/types/generated/api-types";
import { useJoinGame } from "@/hooks/useJoinGame";
import LoadingOverlay from "../../game/view/LoadingOverlay";
import GameMenuModal from "./GameMenuModal";

interface JoinGameOverlayProps {
  game: GameDto;
  onCancel: () => void;
  visible?: boolean;
  onExited?: () => void;
  title?: string;
  subtitle?: string;
}

const JoinGameOverlay: React.FC<JoinGameOverlayProps> = ({
  game,
  onCancel,
  visible,
  onExited,
  title = "Game Found",
  subtitle,
}) => {
  const { playerName, setPlayerName, isLoading, handleJoin, handleKeyDown, loadingMessage } =
    useJoinGame({ game });

  return (
    <>
      <GameMenuModal
        title={title}
        subtitle={subtitle}
        onBack={onCancel}
        visible={visible}
        onExited={onExited}
        showBackdrop={true}
        showSettings={false}
        onClose={onCancel}
        showCloseButton={true}
      >
        <div className="flex flex-row gap-3 items-center px-2">
          <input
            type="text"
            value={playerName}
            onChange={(e) => setPlayerName(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Enter your name"
            disabled={isLoading}
            spellCheck={false}
            autoComplete="off"
            autoCorrect="off"
            maxLength={50}
            autoFocus
            className="flex-1 bg-black/50 border border-white/20 rounded-lg py-3 px-4 text-white text-base outline-none placeholder:text-white/50 focus:border-white/60 focus:shadow-[0_0_20px_rgba(255,255,255,0.1)] transition-all duration-200 disabled:opacity-60"
          />
          <button
            onClick={() => void handleJoin()}
            disabled={isLoading || !playerName.trim()}
            className="font-orbitron bg-space-blue-600 border border-space-blue-500 rounded-lg py-3 px-6 text-white text-sm font-medium hover:bg-space-blue-500 transition-colors disabled:opacity-50 disabled:cursor-default"
          >
            {isLoading ? "Joining..." : "Join"}
          </button>
        </div>
      </GameMenuModal>

      {isLoading && <LoadingOverlay isLoaded={false} message={loadingMessage} />}
    </>
  );
};

export default JoinGameOverlay;
