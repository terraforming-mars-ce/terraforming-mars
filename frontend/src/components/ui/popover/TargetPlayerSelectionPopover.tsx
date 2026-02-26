import React, { useEffect, useRef, useState } from "react";
import { ResourceType, ResourcesDto, ProductionDto } from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { Z_INDEX } from "@/constants/zIndex";

interface TargetPlayer {
  id: string;
  name: string;
  resources: ResourcesDto;
  production: ProductionDto;
}

interface TargetPlayerSelectionPopoverProps {
  resourceType: ResourceType;
  amount: number;
  isSteal?: boolean;
  players: TargetPlayer[];
  onPlayerSelect: (playerId: string) => void;
  onCancel: () => void;
  isVisible: boolean;
}

function getPlayerResourceAmount(player: TargetPlayer, resourceType: ResourceType): number {
  switch (resourceType) {
    case "credit":
      return player.resources.credits;
    case "steel":
      return player.resources.steel;
    case "titanium":
      return player.resources.titanium;
    case "plant":
      return player.resources.plants;
    case "energy":
      return player.resources.energy;
    case "heat":
      return player.resources.heat;
    case "credit-production":
      return player.production.credits;
    case "steel-production":
      return player.production.steel;
    case "titanium-production":
      return player.production.titanium;
    case "plant-production":
      return player.production.plants;
    case "energy-production":
      return player.production.energy;
    case "heat-production":
      return player.production.heat;
    default:
      return 0;
  }
}

const TargetPlayerSelectionPopover: React.FC<TargetPlayerSelectionPopoverProps> = ({
  resourceType,
  amount,
  isSteal,
  players,
  onPlayerSelect,
  onCancel,
  isVisible,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);

  const handleCancelClick = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onCancel();
    }, 200);
  };

  const handlePlayerClick = (playerId: string) => {
    onPlayerSelect(playerId);
  };

  const handleContinueAnyway = () => {
    onPlayerSelect("");
  };

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        handleCancelClick();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        event.button === 0 &&
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node)
      ) {
        handleCancelClick();
      }
    };

    const preventScroll = (event: WheelEvent | TouchEvent) => {
      event.preventDefault();
      event.stopPropagation();
    };

    if (isVisible) {
      document.body.style.overflow = "hidden";
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("wheel", preventScroll, { passive: false });
      document.addEventListener("touchmove", preventScroll, { passive: false });
    }

    return () => {
      document.body.style.overflow = "";
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("wheel", preventScroll);
      document.removeEventListener("touchmove", preventScroll);
    };
  }, [isVisible, onCancel]);

  if (!isVisible) return null;

  const isProduction = resourceType.endsWith("-production");
  const displayIconType = resourceType;
  const eligiblePlayers = players.filter(
    (player) => getPlayerResourceAmount(player, resourceType) > 0,
  );
  const hasNoTargets = eligiblePlayers.length === 0;

  return (
    <div
      className="fixed top-0 left-0 right-0 bottom-0 flex items-center justify-center pointer-events-auto overflow-hidden"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div
        className={`
          min-w-[280px] w-fit max-w-[90vw] max-h-[500px]
          bg-space-black-darker/95
          border-2 border-space-blue-500
          rounded-xl
          shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_rgba(30,60,150,1)]
          backdrop-blur-space
          flex flex-col overflow-hidden isolate
          pointer-events-auto
          ${isClosing ? "animate-fadeOut" : "animate-popIn"}
        `}
        ref={popoverRef}
      >
        {/* Header */}
        <div className="py-[15px] px-5 bg-black/40 border-b border-b-space-blue-500/60">
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            {hasNoTargets ? "No Valid Targets" : "Select Target Player"}
          </h3>
          {!hasNoTargets && (
            <div className="text-white/60 text-xs text-shadow-glow mt-1 flex items-center justify-center gap-1.5">
              <span>
                {isSteal ? "Steal" : "Remove"} up to {amount}
              </span>
              <GameIcon iconType={displayIconType} size="small" />
              <span>{isProduction ? "production" : ""}</span>
            </div>
          )}
        </div>

        {/* Players Container */}
        <div className="flex-1 overflow-y-auto p-2.5 scrollbar-thin scrollbar-thumb-space-blue-500/50 scrollbar-track-space-blue-900/30">
          {hasNoTargets ? (
            <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
              <div className="flex items-center justify-center gap-3 mb-4">
                <GameIcon iconType={displayIconType} size="large" />
              </div>
              <div className="text-white/70 text-xs mb-4 max-w-[280px]">
                No other players have {isProduction ? `${resourceType} production` : resourceType}{" "}
                available. You can continue without targeting anyone.
              </div>
            </div>
          ) : (
            eligiblePlayers.map((player, index) => {
              const currentAmount = getPlayerResourceAmount(player, resourceType);
              const delay = index * 0.05;

              return (
                <div
                  key={player.id}
                  className="
                    bg-black/30
                    border-2 border-space-blue-500/40
                    rounded-[10px] px-3.5 py-3
                    mb-2
                    transition-all duration-[250ms] ease-out
                    animate-choiceSlideIn
                    flex items-center justify-between gap-3
                    cursor-pointer hover:border-space-blue-500/80 hover:bg-black/50 hover:shadow-[0_4px_16px_rgba(30,60,150,0.5)]
                  "
                  style={{ animationDelay: `${delay}s` }}
                  onClick={() => handlePlayerClick(player.id)}
                >
                  <div className="text-white font-semibold text-sm">{player.name}</div>
                  <div className="flex items-center gap-1.5 flex-shrink-0">
                    <span className="text-white/60 text-xs font-medium">{currentAmount}</span>
                    <GameIcon iconType={displayIconType} size="small" />
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className="px-4 py-3 bg-black/40 border-t border-space-blue-500/60 flex justify-center gap-3">
          {hasNoTargets ? (
            <>
              <button
                className="
                  bg-yellow-600/50
                  border-2 border-yellow-500/60
                  rounded-md text-white text-xs font-semibold
                  px-6 py-2 cursor-pointer
                  transition-all duration-200
                  text-shadow-glow font-orbitron
                  shadow-[0_0_8px_rgba(180,120,0,0.4)]
                  hover:bg-yellow-500/60
                  hover:border-yellow-500/80
                                   hover:shadow-[0_0_12px_rgba(180,120,0,0.6)]
                "
                onClick={handleContinueAnyway}
              >
                Continue Anyway
              </button>
              <button
                className="
                  bg-space-blue-600/50
                  border-2 border-space-blue-500/60
                  rounded-md text-white text-xs font-semibold
                  px-6 py-2 cursor-pointer
                  transition-all duration-200
                  text-shadow-glow font-orbitron
                  shadow-[0_0_8px_rgba(30,60,150,0.4)]
                  hover:bg-space-blue-500/60
                  hover:border-space-blue-500/80
                                   hover:shadow-[0_0_12px_rgba(30,60,150,0.6)]
                "
                onClick={handleCancelClick}
              >
                Cancel
              </button>
            </>
          ) : (
            <button
              className="
                bg-space-blue-600/50
                border-2 border-space-blue-500/60
                rounded-md text-white text-xs font-semibold
                px-6 py-2 cursor-pointer
                transition-all duration-200
                text-shadow-glow font-orbitron
                shadow-[0_0_8px_rgba(30,60,150,0.4)]
                hover:bg-space-blue-500/60
                hover:border-space-blue-500/80
                               hover:shadow-[0_0_12px_rgba(30,60,150,0.6)]
              "
              onClick={handleCancelClick}
            >
              Cancel
            </button>
          )}
        </div>
      </div>

      <style>{`
        @keyframes popIn {
          from {
            opacity: 0;
            transform: scale(0.9) translateY(-20px);
          }
          to {
            opacity: 1;
            transform: scale(1) translateY(0);
          }
        }

        @keyframes fadeOut {
          from {
            opacity: 1;
          }
          to {
            opacity: 0;
          }
        }

        @keyframes choiceSlideIn {
          from {
            opacity: 0;
            transform: translateX(-20px);
          }
          to {
            opacity: 1;
            transform: translateX(0);
          }
        }

        .animate-popIn {
          animation: popIn 0.25s ease-out;
        }

        .animate-fadeOut {
          animation: fadeOut 0.2s ease-out forwards;
        }

        .animate-choiceSlideIn {
          animation: choiceSlideIn 0.3s ease-out both;
        }
      `}</style>
    </div>
  );
};

export default TargetPlayerSelectionPopover;
