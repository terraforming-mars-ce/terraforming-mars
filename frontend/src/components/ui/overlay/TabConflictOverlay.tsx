import React from "react";
import { Z_INDEX, getZIndex } from "@/constants/zIndex.ts";

interface TabConflictOverlayProps {
  activeGameInfo: {
    gameId: string;
    playerName: string;
  };
  onTakeOver: () => void;
  onCancel: () => void;
}

const TabConflictOverlay: React.FC<TabConflictOverlayProps> = ({
  activeGameInfo,
  onTakeOver,
  onCancel,
}) => {
  return (
    <>
      {/* Overlay backdrop */}
      <div
        className="fixed top-0 left-0 right-0 bottom-0 bg-[rgba(0,0,17,0.8)] [backdrop-filter:blur(4px)]"
        style={{ zIndex: Z_INDEX.ERROR_OVERLAYS }}
      />

      {/* Warning modal */}
      <div
        className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 min-w-[400px] max-w-[600px] w-[90%] max-[768px]:min-w-[320px] max-[768px]:mx-5"
        style={{ zIndex: getZIndex("ERROR_OVERLAYS", 1) }}
      >
        <div className="bg-[linear-gradient(135deg,rgba(30,60,90,0.95)_0%,rgba(20,40,70,0.9)_50%,rgba(10,30,60,0.95)_100%)] border-2 border-[rgba(255,193,7,0.6)] rounded-[20px] p-10 [backdrop-filter:blur(10px)] shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(255,193,7,0.2)] text-center max-[768px]:p-6 max-[480px]:p-5">
          <div className="text-[64px] mb-5 max-[768px]:text-5xl max-[480px]:text-[40px] max-[480px]:mb-4">
            ⚠️
          </div>

          <h2 className="text-[#ffc107] text-[28px] m-0 mb-6 [text-shadow:0_2px_4px_rgba(0,0,0,0.8)] font-bold max-[768px]:text-[22px] max-[480px]:text-lg">
            Game Already Active
          </h2>

          <div className="text-white/90 text-base leading-[1.6] mb-8 text-left max-[768px]:text-sm [&_p]:m-0 [&_p]:mb-4">
            <p>A game is already running in another tab or window:</p>
            <div className="bg-black/30 rounded-lg p-4 my-4 border border-white/20">
              <div className="flex justify-between mb-2 last:mb-0">
                <span className="text-white/70 font-medium">Player:</span>
                <span className="text-white font-bold font-[monospace]">
                  {activeGameInfo.playerName}
                </span>
              </div>
              <div className="flex justify-between mb-2 last:mb-0">
                <span className="text-white/70 font-medium">Game ID:</span>
                <span className="text-white font-bold font-[monospace]">
                  {activeGameInfo.gameId}
                </span>
              </div>
            </div>
            <p className="!text-[#ffc107] !text-sm italic !mt-4 !text-center">
              Opening the game here will close it in the other tab. This may interrupt your gameplay
              if you're actively playing there.
            </p>
          </div>

          <div className="flex gap-4 justify-center max-[768px]:flex-col max-[768px]:gap-3">
            <button
              className="bg-[rgba(108,117,125,0.8)] border-2 border-[rgba(108,117,125,0.5)] rounded-xl py-3 px-6 text-base font-bold text-white cursor-pointer transition-all duration-300 [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] hover:bg-[rgba(108,117,125,1)] hover:border-[rgba(108,117,125,0.8)] hover:-translate-y-px max-[768px]:w-full max-[768px]:py-[14px] max-[768px]:px-5 max-[768px]:text-sm"
              onClick={onCancel}
            >
              Cancel
            </button>
            <button
              className="bg-[linear-gradient(135deg,#dc3545_0%,#e55564_100%)] border-2 border-[rgba(220,53,69,0.5)] rounded-xl py-3 px-6 text-base font-bold text-white cursor-pointer transition-all duration-300 [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] shadow-[0_4px_20px_rgba(220,53,69,0.3)] hover:bg-[linear-gradient(135deg,#c82333_0%,#dc3545_100%)] hover:border-[rgba(220,53,69,0.8)] hover:-translate-y-px hover:shadow-[0_6px_25px_rgba(220,53,69,0.4)] max-[768px]:w-full max-[768px]:py-[14px] max-[768px]:px-5 max-[768px]:text-sm"
              onClick={onTakeOver}
            >
              Continue Here
            </button>
          </div>
        </div>
      </div>
    </>
  );
};

export default TabConflictOverlay;
