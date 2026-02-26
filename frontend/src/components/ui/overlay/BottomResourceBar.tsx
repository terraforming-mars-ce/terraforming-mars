import React, { useRef, useState, useEffect } from "react";
import {
  PlayerDto,
  OtherPlayerDto,
  PlayerActionDto,
  GameDto,
  CardDto,
  ResourceTypeCredit,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlant,
  ResourceTypeEnergy,
  ResourceTypeHeat,
  ResourceTypeGreeneryTile,
  ResourceTypeTemperature,
} from "@/types/generated/api-types.ts";
import ActionsPopover from "../popover/ActionsPopover.tsx";
import EffectsPopover from "../popover/EffectsPopover.tsx";
import TagsPopover from "../popover/TagsPopover.tsx";
import StoragesPopover from "../popover/StoragesPopover.tsx";
import LogPopover from "../popover/LogPopover.tsx";
import VictoryPointsPopover from "../popover/VictoryPointsPopover.tsx";
import GameIcon from "../display/GameIcon.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import {
  calculatePlantsForGreenery,
  calculateHeatForTemperature,
} from "@/utils/resourceConversionUtils.ts";
import { useHoverSound } from "@/hooks/useHoverSound.ts";

interface AngledPanelProps {
  side: "left" | "right";
  corpColor: string;
  width: number;
  height: number;
  children: React.ReactNode;
}

const ANGLE_INDENT = 42;

const BORDER_COLOR = "rgba(60,60,70,0.7)";

const AngledPanel: React.FC<AngledPanelProps> = ({
  side,
  corpColor,
  width,
  height,
  children,
}) => {
  const fillPoints =
    side === "left"
      ? `0,0 ${width - ANGLE_INDENT},0 ${width},${height} 0,${height}`
      : `${ANGLE_INDENT},0 ${width},0 ${width},${height} 0,${height}`;

  const topEdge =
    side === "left"
      ? { x1: 0, y1: 0, x2: width - ANGLE_INDENT, y2: 0 }
      : { x1: ANGLE_INDENT, y1: 0, x2: width, y2: 0 };

  const angledEdge =
    side === "left"
      ? { x1: width - ANGLE_INDENT, y1: 0, x2: width, y2: height }
      : { x1: 0, y1: height, x2: ANGLE_INDENT, y2: 0 };

  const gradientId = `corpGradient-${side}`;
  const whiteGlowId = `whiteGlow-${side}`;

  return (
    <div className="relative pointer-events-auto" style={{ width, height }}>
      <svg
        className="absolute inset-0 w-full h-full"
        viewBox={`0 0 ${width} ${height}`}
        preserveAspectRatio="none"
      >
        <defs>
          {side === "left" && (
            <linearGradient id={gradientId} x1="0%" y1="0%" x2="30%" y2="0%">
              <stop offset="0%" stopColor={corpColor} stopOpacity="0.17" />
              <stop offset="100%" stopColor={corpColor} stopOpacity="0" />
            </linearGradient>
          )}
          {side === "right" && (
            <linearGradient id={whiteGlowId} x1="70%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#ffffff" stopOpacity="0" />
              <stop offset="100%" stopColor="#ffffff" stopOpacity="0.17" />
            </linearGradient>
          )}
        </defs>
        <polygon points={fillPoints} fill="rgba(10,10,15,0.95)" />
        {side === "left" && (
          <polygon points={fillPoints} fill={`url(#${gradientId})`} />
        )}
        {side === "right" && (
          <polygon points={fillPoints} fill={`url(#${whiteGlowId})`} />
        )}
        <line
          x1={topEdge.x1}
          y1={topEdge.y1}
          x2={topEdge.x2}
          y2={topEdge.y2}
          stroke={BORDER_COLOR}
          strokeWidth="4"
        />
        <line
          x1={angledEdge.x1}
          y1={angledEdge.y1}
          x2={angledEdge.x2}
          y2={angledEdge.y2}
          stroke={BORDER_COLOR}
          strokeWidth="3"
        />
      </svg>
      <div className="relative z-10 h-full">{children}</div>
    </div>
  );
};

interface ResourceData {
  id: string;
  name: string;
  current: number;
  production: number;
}

export interface BottomResourceBarCallbacks {
  onOpenCardEffectsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenActionsModal?: () => void;
  onActionSelect?: (action: PlayerActionDto) => void;
  onConvertPlantsToGreenery?: () => void;
  onConvertHeatToTemperature?: () => void;
}

interface BottomResourceBarProps {
  currentPlayer?: PlayerDto | null;
  gameState?: GameDto;
  playedCards?: CardDto[];
  changedPaths?: Set<string>;
  callbacks?: BottomResourceBarCallbacks;
  gameId?: string;
  corporation?: CardDto | null;
  spectatingPlayer?: PlayerDto | OtherPlayerDto | null;
  spectatingCorporation?: CardDto | null;
  spectatePlayerColor?: string;
  onStopSpectating?: () => void;
}

const BottomResourceBar: React.FC<BottomResourceBarProps> = ({
  currentPlayer,
  gameState,
  playedCards = [],
  changedPaths = new Set(),
  callbacks = {},
  gameId,
  corporation,
  spectatingPlayer,
  spectatingCorporation,
  spectatePlayerColor,
  onStopSpectating,
}) => {
  const isSpectating = !!spectatingPlayer;
  const displayPlayer: PlayerDto | OtherPlayerDto | null | undefined =
    isSpectating ? spectatingPlayer : currentPlayer;
  const displayCorporation = isSpectating ? spectatingCorporation : corporation;
  const displayPlayedCards = isSpectating
    ? (spectatingPlayer?.playedCards ?? [])
    : playedCards;

  const {
    onOpenCardEffectsModal,
    onOpenCardsPlayedModal,
    onOpenActionsModal,
    onActionSelect,
    onConvertPlantsToGreenery,
    onConvertHeatToTemperature,
  } = callbacks;
  const [showActionsPopover, setShowActionsPopover] = useState(false);
  const [showEffectsPopover, setShowEffectsPopover] = useState(false);
  const [showTagsPopover, setShowTagsPopover] = useState(false);
  const [showStoragesPopover, setShowStoragesPopover] = useState(false);
  const [showLogPopover, setShowLogPopover] = useState(false);
  const [showVPPopover, setShowVPPopover] = useState(false);
  const [isCorpExpanded, setIsCorpExpanded] = useState(false);
  const [showCorpExpanded, setShowCorpExpanded] = useState(false);
  const hoverSound = useHoverSound();
  const actionsButtonRef = useRef<HTMLButtonElement>(null);
  const effectsButtonRef = useRef<HTMLButtonElement>(null);
  const tagsButtonRef = useRef<HTMLButtonElement>(null);
  const storagesButtonRef = useRef<HTMLButtonElement>(null);
  const logButtonRef = useRef<HTMLButtonElement>(null);
  const vpButtonRef = useRef<HTMLButtonElement>(null);
  const corpContainerRef = useRef<HTMLDivElement>(null);

  const corpColor = isSpectating
    ? (spectatePlayerColor ?? "#ffc107")
    : corporation
      ? getCorporationBorderColor(corporation.name)
      : "#ffc107";

  const handleCorpToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isCorpExpanded) {
      handleCorpClose();
    } else {
      setShowCorpExpanded(true);
      requestAnimationFrame(() => {
        setIsCorpExpanded(true);
      });
    }
  };

  const handleCorpClose = () => {
    setIsCorpExpanded(false);
    setTimeout(() => {
      setShowCorpExpanded(false);
    }, 200);
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        corpContainerRef.current &&
        !corpContainerRef.current.contains(event.target as Node)
      ) {
        handleCorpClose();
      }
    };

    if (isCorpExpanded) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isCorpExpanded]);

  const hasPathChanged = (path: string): boolean => {
    return changedPaths.has(path);
  };

  const getResourceType = (resourceId: string): string => {
    const resourceTypeMap: Record<string, string> = {
      credits: ResourceTypeCredit,
      steel: ResourceTypeSteel,
      titanium: ResourceTypeTitanium,
      plants: ResourceTypePlant,
      energy: ResourceTypeEnergy,
      heat: ResourceTypeHeat,
    };
    return resourceTypeMap[resourceId] || resourceId;
  };

  const tagCounts = React.useMemo(() => {
    if (!displayPlayedCards || displayPlayedCards.length === 0) return [];

    const counts: { [key: string]: number } = {};

    displayPlayedCards.forEach((card) => {
      if (card.tags) {
        card.tags.forEach((tag) => {
          const tagKey = tag.toLowerCase();
          counts[tagKey] = (counts[tagKey] || 0) + 1;
        });
      }
    });

    const allTags = [
      "space",
      "earth",
      "science",
      "power",
      "building",
      "microbe",
      "animal",
      "plant",
      "event",
      "city",
      "venus",
      "jovian",
      "wild",
      "mars",
      "moon",
      "clone",
      "crime",
    ];

    return allTags.map((tag) => ({
      tag,
      count: counts[tag] || 0,
    }));
  }, [displayPlayedCards]);

  const storageCardsCount = React.useMemo(() => {
    if (!displayPlayer?.resourceStorage) return 0;
    return Object.keys(displayPlayer.resourceStorage).length;
  }, [displayPlayer?.resourceStorage]);

  if (!displayPlayer?.resources || !displayPlayer?.production) {
    return null;
  }

  const playerResources: ResourceData[] = [
    {
      id: "credit",
      name: "Credits",
      current: displayPlayer.resources.credits,
      production: displayPlayer.production.credits,
    },
    {
      id: "steel",
      name: "Steel",
      current: displayPlayer.resources.steel,
      production: displayPlayer.production.steel,
    },
    {
      id: "titanium",
      name: "Titanium",
      current: displayPlayer.resources.titanium,
      production: displayPlayer.production.titanium,
    },
    {
      id: "plant",
      name: "Plants",
      current: displayPlayer.resources.plants,
      production: displayPlayer.production.plants,
    },
    {
      id: "energy",
      name: "Energy",
      current: displayPlayer.resources.energy,
      production: displayPlayer.production.energy,
    },
    {
      id: "heat",
      name: "Heat",
      current: displayPlayer.resources.heat,
      production: displayPlayer.production.heat,
    },
  ];

  const playedCardsCount = displayPlayer?.playedCards?.length || 0;

  const requiredPlants = calculatePlantsForGreenery(displayPlayer?.effects);
  const requiredHeat = calculateHeatForTemperature(displayPlayer?.effects);

  const canConvertPlants =
    !isSpectating && (displayPlayer?.resources.plants ?? 0) >= requiredPlants;
  const canConvertHeat =
    !isSpectating &&
    (displayPlayer?.resources.heat ?? 0) >= requiredHeat &&
    (gameState?.globalParameters?.temperature ?? -30) < 8;

  const handleOpenCardsModal = () => {
    onOpenCardsPlayedModal?.();
  };

  const handleOpenActionsPopover = () => {
    setShowActionsPopover(!showActionsPopover);
  };

  const handleOpenEffectsPopover = () => {
    setShowEffectsPopover(!showEffectsPopover);
  };

  const handleOpenStoragesPopover = () => {
    setShowStoragesPopover(!showStoragesPopover);
  };

  const handleOpenTagsPopover = () => {
    setShowTagsPopover(!showTagsPopover);
  };

  const vpGranters =
    displayPlayer && "vpGranters" in displayPlayer
      ? displayPlayer.vpGranters
      : [];
  const totalVP = (vpGranters || []).reduce(
    (sum, g) => sum + g.computedValue,
    0,
  );

  const handleOpenVPPopover = () => {
    setShowVPPopover(!showVPPopover);
  };

  const isTilePlacementActive = !!currentPlayer?.pendingTileSelection;
  const isConversionDisabled = isTilePlacementActive;

  const BAR_HEIGHT = 90;

  const MAX_PANEL_WIDTH = 640;

  const calcPanelWidth = () =>
    Math.min(MAX_PANEL_WIDTH, Math.max(480, (window.innerWidth - 700) / 2));

  const [panelWidth, setPanelWidth] = useState(calcPanelWidth);

  useEffect(() => {
    const handleResize = () => setPanelWidth(calcPanelWidth());
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const LEFT_PANEL_WIDTH = panelWidth;
  const RIGHT_PANEL_WIDTH = panelWidth;
  const contentScale = panelWidth / MAX_PANEL_WIDTH;

  return (
    <div className="fixed bottom-0 left-0 right-0 flex justify-between pointer-events-none">
      {/* Spectating banner */}
      {isSpectating && (
        <div
          className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 flex items-center gap-3 px-4 py-2 bg-[rgba(10,10,15,0.95)] border border-[rgba(60,60,70,0.7)] pointer-events-auto cursor-pointer hover:bg-[rgba(20,20,25,0.95)] transition-colors"
          style={{ borderLeftColor: spectatePlayerColor, borderLeftWidth: 3 }}
          onClick={onStopSpectating}
        >
          <span className="text-white/90 text-xs font-orbitron">
            Viewing{" "}
            <span className="font-bold" style={{ color: spectatePlayerColor }}>
              {displayPlayer?.name}
            </span>
            {displayPlayer && "handCardCount" in displayPlayer && (
              <span className="text-white/50 ml-2">
                ({displayPlayer.handCardCount} cards in hand)
              </span>
            )}
          </span>
          <span className="text-white/50 text-[10px] font-orbitron">
            ESC to close
          </span>
        </div>
      )}

      {/* LEFT PANEL: Corporation + Resources */}
      <AngledPanel
        side="left"
        corpColor={corpColor}
        width={LEFT_PANEL_WIDTH}
        height={BAR_HEIGHT}
      >
        <div
          className="flex items-center h-full origin-left"
          style={{
            paddingRight: ANGLE_INDENT + 16,
            transform: `scale(${contentScale})`,
            width: MAX_PANEL_WIDTH,
          }}
        >
          {/* Corporation Section */}
          <div
            ref={corpContainerRef}
            className="flex items-center relative w-[120px] justify-center"
          >
            {displayCorporation && (
              <>
                {/* Corporation Logo Button */}
                <div
                  className="cursor-pointer p-2 transition-all duration-200 hover:brightness-110"
                  onClick={(e) => {
                    hoverSound.onClick?.();
                    handleCorpToggle(e);
                  }}
                  onMouseEnter={hoverSound.onMouseEnter}
                  style={{
                    filter: `drop-shadow(0 0 8px ${corpColor}50)`,
                  }}
                >
                  <div className="flex items-center justify-center min-h-[50px] [&>*]:scale-65 [&>*]:origin-center">
                    {getCorporationLogo(displayCorporation.name.toLowerCase())}
                  </div>
                </div>

                {/* Expanded Corporation Card */}
                {showCorpExpanded && (
                  <div
                    className={`absolute bottom-[100%] left-4 mb-2 origin-bottom-left transition-all duration-200 ${
                      isCorpExpanded
                        ? "opacity-100 scale-100"
                        : "opacity-0 scale-90"
                    }`}
                  >
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCorpClose();
                      }}
                      className="absolute top-4 right-4 text-white/70 hover:text-white text-xl leading-none transition-colors z-10 cursor-pointer"
                    >
                      ×
                    </button>
                    <CorporationCard
                      corporation={{
                        id: displayCorporation.id,
                        name: displayCorporation.name,
                        description: displayCorporation.description,
                        startingMegaCredits:
                          displayCorporation.startingResources?.credits || 0,
                        startingProduction:
                          displayCorporation.startingProduction,
                        startingResources: displayCorporation.startingResources,
                        behaviors: displayCorporation.behaviors,
                      }}
                      isSelected={false}
                      onSelect={() => {}}
                      showCheckbox={false}
                      borderColor={corpColor}
                      disableInteraction={true}
                    />
                  </div>
                )}
              </>
            )}
          </div>

          {/* Divider between corp and resources */}
          <div
            className="w-[1px] h-[50px] self-center mx-2"
            style={{
              background: `linear-gradient(transparent, ${corpColor}40, transparent)`,
            }}
          />

          {/* Resources Section */}
          <div className="flex items-center justify-evenly flex-1">
            {playerResources.map((resource, index) => {
              const resourceChanged = hasPathChanged(
                `currentPlayer.resources.${resource.id}`,
              );
              const productionChanged = hasPathChanged(
                `currentPlayer.production.${resource.id}`,
              );

              const showConversionButton =
                (resource.id === "plant" && canConvertPlants) ||
                (resource.id === "heat" && canConvertHeat);

              return (
                <React.Fragment key={resource.id}>
                  {/* Resource Item */}
                  <div className="flex flex-col items-center gap-0.5 px-2 py-1.5 relative">
                    {/* Conversion button - positioned absolutely above production box */}
                    {(resource.id === "plant" || resource.id === "heat") && (
                      <div
                        className={`absolute left-1/2 -translate-x-1/2 bottom-full mb-1 transition-all duration-300 ${
                          showConversionButton
                            ? "opacity-100 scale-100"
                            : "opacity-0 scale-90 pointer-events-none"
                        }`}
                      >
                        <button
                          disabled={
                            isConversionDisabled || !showConversionButton
                          }
                          className={`flex items-center justify-center gap-0.5 px-1.5 py-0.5 bg-black/80 border border-white/20 transition-all duration-200 ${
                            isConversionDisabled || !showConversionButton
                              ? "opacity-40 cursor-not-allowed"
                              : "cursor-pointer hover:bg-white/10 hover:border-white/40"
                          }`}
                          style={{
                            borderColor: showConversionButton
                              ? `${corpColor}60`
                              : undefined,
                          }}
                          onClick={(e) => {
                            e.stopPropagation();
                            if (isConversionDisabled || !showConversionButton)
                              return;
                            hoverSound.onClick?.();
                            if (resource.id === "plant") {
                              void onConvertPlantsToGreenery?.();
                            } else if (resource.id === "heat") {
                              void onConvertHeatToTemperature?.();
                            }
                          }}
                          onMouseEnter={
                            isConversionDisabled || !showConversionButton
                              ? undefined
                              : hoverSound.onMouseEnter
                          }
                        >
                          <span className="text-[10px] font-bold text-white/90">
                            +
                          </span>
                          <GameIcon
                            iconType={
                              resource.id === "plant"
                                ? ResourceTypeGreeneryTile
                                : ResourceTypeTemperature
                            }
                            size="small"
                          />
                        </button>
                      </div>
                    )}

                    {/* Production badge */}
                    <div className="inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.5)_0%,rgba(139,89,42,0.45)_100%)] border border-[rgba(160,110,60,0.6)] px-3 py-0.5 w-[32px] mb-1">
                      <span
                        className={`text-[10px] font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none tabular-nums ${
                          productionChanged
                            ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                            : ""
                        }`}
                      >
                        {resource.production}
                      </span>
                    </div>

                    {/* Resource icon and value */}
                    {resource.id === "credit" ? (
                      <div className="flex items-center gap-1 w-[48px] justify-center scale-110">
                        <GameIcon
                          iconType={ResourceTypeCredit}
                          amount={resource.current}
                          size="small"
                        />
                      </div>
                    ) : (
                      <div className="flex items-center gap-1 w-[48px] justify-center scale-[0.85]">
                        <GameIcon
                          iconType={getResourceType(resource.id)}
                          size="small"
                        />
                        <span
                          className={`text-sm font-bold font-orbitron text-white [text-shadow:0_1px_3px_rgba(0,0,0,0.8)] tabular-nums w-[24px] text-left ${
                            resourceChanged
                              ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                              : ""
                          }`}
                        >
                          {resource.current}
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Divider between resources */}
                  {index < playerResources.length - 1 && (
                    <div
                      className="w-[1px] h-[40px] self-center"
                      style={{
                        background: `linear-gradient(transparent, rgba(60,60,70,0.8), transparent)`,
                      }}
                    />
                  )}
                </React.Fragment>
              );
            })}
          </div>
        </div>
      </AngledPanel>

      {/* RIGHT PANEL: Action Buttons */}
      <AngledPanel
        side="right"
        corpColor={corpColor}
        width={RIGHT_PANEL_WIDTH}
        height={BAR_HEIGHT}
      >
        <div
          className="flex items-center h-full justify-evenly"
          style={{
            paddingLeft: ANGLE_INDENT + 16,
            paddingRight: ANGLE_INDENT + 16,
          }}
        >
          {/* Actions Button */}
          <button
            ref={actionsButtonRef}
            className="group flex flex-col items-center gap-1.5 p-1.5 cursor-pointer transition-all duration-200 w-[52px] hover:bg-white/5"
            onClick={() => {
              hoverSound.onClick?.();
              handleOpenActionsPopover();
            }}
            onMouseEnter={hoverSound.onMouseEnter}
          >
            <div className="font-bold flex items-center gap-[2px] h-[24px] w-[24px] justify-center text-[rgb(140,140,150)] group-hover:text-[rgb(100,160,220)] transition-colors duration-200">
              <span className="text-[7px] leading-none translate-y-[1px]">
                ●
              </span>
              <span className="text-[7px] leading-none translate-y-[1px]">
                ●
              </span>
              <span className="text-[18px] leading-none">→</span>
            </div>
            <div
              className={`text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${
                hasPathChanged("currentPlayer.actions")
                  ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                  : ""
              }`}
            >
              {displayPlayer?.actions?.length || 0}
            </div>
            <div className="text-[8px] font-medium font-orbitron text-white/70 uppercase tracking-[0.5px]">
              Actions
            </div>
          </button>

          {/* Effects Button */}
          <button
            ref={effectsButtonRef}
            className="group flex flex-col items-center gap-1.5 p-1.5 cursor-pointer transition-all duration-200 w-[52px] hover:bg-white/5"
            onClick={() => {
              hoverSound.onClick?.();
              handleOpenEffectsPopover();
            }}
            onMouseEnter={hoverSound.onMouseEnter}
          >
            <div className="font-bold flex items-center justify-center h-[24px] w-[24px] relative text-[rgb(140,140,150)] group-hover:text-[rgb(100,160,220)] transition-colors duration-200">
              <div className="absolute w-[20px] h-[20px] rounded-full border-2 border-current" />
              <div className="flex flex-col items-center justify-center relative">
                <span className="text-[7px] leading-none">●</span>
                <span className="text-[7px] leading-none">●</span>
              </div>
            </div>
            <div
              className={`text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${
                hasPathChanged("currentPlayer.effects")
                  ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                  : ""
              }`}
            >
              {displayPlayer?.effects?.length || 0}
            </div>
            <div className="text-[8px] font-medium font-orbitron text-white/70 uppercase tracking-[0.5px]">
              Effects
            </div>
          </button>

          {/* Tags Button */}
          <button
            ref={tagsButtonRef}
            className="group flex flex-col items-center gap-1.5 p-1.5 cursor-pointer transition-all duration-200 w-[52px] hover:bg-white/5"
            onClick={() => {
              hoverSound.onClick?.();
              handleOpenTagsPopover();
            }}
            onMouseEnter={hoverSound.onMouseEnter}
          >
            <div className="font-bold flex items-center justify-center h-[24px] w-[24px] relative text-[rgb(140,140,150)] group-hover:text-[rgb(100,160,220)] transition-colors duration-200">
              <div className="absolute w-[20px] h-[20px] rounded-full border-2 border-current" />
              <div className="flex items-center gap-[2px] relative text-[6px] leading-none">
                <span>●</span>
                <span>●</span>
                <span>●</span>
              </div>
            </div>
            <div
              className={`text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${
                hasPathChanged("currentPlayer.playedCards")
                  ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                  : ""
              }`}
            >
              {tagCounts.reduce((sum, tag) => sum + tag.count, 0)}
            </div>
            <div className="text-[8px] font-medium font-orbitron text-white/70 uppercase tracking-[0.5px]">
              Tags
            </div>
          </button>

          {/* Storages Button */}
          <button
            ref={storagesButtonRef}
            className="group flex flex-col items-center gap-1.5 p-1.5 cursor-pointer transition-all duration-200 w-[52px] hover:bg-white/5"
            onClick={() => {
              hoverSound.onClick?.();
              handleOpenStoragesPopover();
            }}
            onMouseEnter={hoverSound.onMouseEnter}
          >
            <div className="font-bold flex items-center justify-center h-[24px] w-[24px] relative text-[rgb(140,140,150)] group-hover:text-[rgb(100,160,220)] transition-colors duration-200">
              <div className="absolute w-[20px] h-[20px] border-2 border-current" />
              <div className="flex items-center gap-[2px] relative text-[6px] leading-none">
                <span>●</span>
                <span>●</span>
                <span>●</span>
              </div>
            </div>
            <div
              className={`text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${
                hasPathChanged("currentPlayer.resourceStorage")
                  ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                  : ""
              }`}
            >
              {storageCardsCount}
            </div>
            <div className="text-[8px] font-medium font-orbitron text-white/70 uppercase tracking-[0.5px]">
              Storages
            </div>
          </button>

          {/* Played Cards Button */}
          <button
            className="group flex flex-col items-center gap-1.5 p-1.5 cursor-pointer transition-all duration-200 w-[52px] hover:bg-white/5"
            onClick={() => {
              hoverSound.onClick?.();
              handleOpenCardsModal();
            }}
            onMouseEnter={hoverSound.onMouseEnter}
          >
            <div className="text-lg font-bold flex items-center justify-center h-[24px] w-[24px] text-[rgb(140,140,150)] group-hover:text-[rgb(100,160,220)] transition-colors duration-200">
              ↓
            </div>
            <div
              className={`text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${
                hasPathChanged("currentPlayer.playedCards")
                  ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                  : ""
              }`}
            >
              {playedCardsCount}
            </div>
            <div className="text-[8px] font-medium font-orbitron text-white/70 uppercase tracking-[0.5px]">
              Played
            </div>
          </button>

          {/* VP Button */}
          <button
            ref={vpButtonRef}
            className="group flex flex-col items-center gap-1.5 p-1.5 cursor-pointer transition-all duration-200 w-[52px] hover:bg-white/5"
            onClick={() => {
              hoverSound.onClick?.();
              handleOpenVPPopover();
            }}
            onMouseEnter={hoverSound.onMouseEnter}
          >
            <div className="font-bold flex items-center justify-center h-[24px] w-[24px] relative text-[rgb(140,140,150)] group-hover:text-[rgb(100,160,220)] transition-colors duration-200">
              <span className="text-xl absolute">○</span>
              <span className="text-sm absolute">●</span>
            </div>
            <div
              className={`text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${
                hasPathChanged("currentPlayer.vpGranters") ||
                hasPathChanged("currentPlayer.terraformRating")
                  ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                  : ""
              }`}
            >
              {totalVP}
            </div>
            <div className="text-[8px] font-medium font-orbitron text-white/70 uppercase tracking-[0.5px]">
              VP
            </div>
          </button>

          {/* Log Button */}
          {gameId && (
            <button
              ref={logButtonRef}
              className="group flex flex-col items-center gap-1.5 p-1.5 cursor-pointer transition-all duration-200 w-[52px] hover:bg-white/5"
              onClick={() => {
                hoverSound.onClick?.();
                setShowLogPopover(!showLogPopover);
              }}
              onMouseEnter={hoverSound.onMouseEnter}
            >
              <div className="font-bold flex items-center justify-center h-[24px] w-[24px] relative text-base text-[rgb(140,140,150)] group-hover:text-[rgb(100,160,220)] transition-colors duration-200">
                ☰
              </div>
              <div className="text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none invisible">
                0
              </div>
              <div className="text-[8px] font-medium font-orbitron text-white/70 uppercase tracking-[0.5px]">
                Log
              </div>
            </button>
          )}
        </div>
      </AngledPanel>

      {/* Popovers */}
      <ActionsPopover
        isVisible={showActionsPopover && !isSpectating}
        onClose={() => setShowActionsPopover(false)}
        actions={displayPlayer?.actions || []}
        playerName={displayPlayer?.name}
        onActionSelect={(action) => {
          onActionSelect?.(action);
          setShowActionsPopover(false);
        }}
        onOpenDetails={onOpenActionsModal}
        anchorRef={actionsButtonRef as React.RefObject<HTMLElement>}
        gameState={gameState}
      />

      <EffectsPopover
        isVisible={showEffectsPopover}
        onClose={() => setShowEffectsPopover(false)}
        effects={displayPlayer?.effects || []}
        playerName={displayPlayer?.name}
        onOpenDetails={isSpectating ? undefined : onOpenCardEffectsModal}
        anchorRef={effectsButtonRef as React.RefObject<HTMLElement>}
      />

      <TagsPopover
        isVisible={showTagsPopover}
        onClose={() => setShowTagsPopover(false)}
        tagCounts={tagCounts}
        anchorRef={tagsButtonRef as React.RefObject<HTMLElement>}
      />

      {currentPlayer && (
        <StoragesPopover
          isVisible={showStoragesPopover && !isSpectating}
          onClose={() => setShowStoragesPopover(false)}
          player={currentPlayer}
          anchorRef={storagesButtonRef as React.RefObject<HTMLElement>}
        />
      )}

      <VictoryPointsPopover
        isVisible={showVPPopover && !isSpectating}
        onClose={() => setShowVPPopover(false)}
        vpGranters={vpGranters || []}
        totalVP={totalVP}
        anchorRef={vpButtonRef as React.RefObject<HTMLElement>}
      />

      {gameId && (
        <LogPopover
          isVisible={showLogPopover}
          onClose={() => setShowLogPopover(false)}
          anchorRef={logButtonRef as React.RefObject<HTMLElement>}
          gameId={gameId}
          gameState={gameState}
        />
      )}
    </div>
  );
};

export default BottomResourceBar;
