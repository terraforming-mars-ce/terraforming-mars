import React from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ResourceTypeCredit,
  PlayerStandardProjectDto,
} from "@/types/generated/api-types.ts";
import { StandardProject, STANDARD_PROJECTS, STANDARD_PROJECT_COSTS } from "@/types/cards.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { canPerformActions } from "@/utils/actionUtils.ts";
import { GamePopover, GamePopoverItem } from "../GamePopover";
import { FormattedDescription } from "../display/FormattedDescription";

interface StandardProjectsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  onProjectSelect: (project: StandardProject) => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

const StandardProjectPopover: React.FC<StandardProjectsPopoverProps> = ({
  isVisible,
  onClose,
  onProjectSelect,
  gameState,
  anchorRef,
}) => {
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
  const canExecuteProjects =
    isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);

  const serverProjects = gameState?.currentPlayer?.standardProjects ?? [];

  // For spectators, derive items from static data
  const playerProjects: PlayerStandardProjectDto[] =
    serverProjects.length > 0
      ? serverProjects
      : Object.entries(STANDARD_PROJECTS).map(([key]) => {
          const cost = STANDARD_PROJECT_COSTS[key as StandardProject];
          return {
            projectType: key,
            baseCost: { credit: cost },
            effectiveCost: { credit: cost },
            available: false,
            errors: [],
          };
        });
  const availableCount = playerProjects.filter((p) => p.available).length;

  const handleProjectClick = (project: PlayerStandardProjectDto) => {
    if (!canExecuteProjects || !project.available) return;
    onProjectSelect(project.projectType as StandardProject);
  };

  const renderEffects = (projectType: string) => {
    const effects: React.ReactElement[] = [];
    const staticProject = STANDARD_PROJECTS[projectType as StandardProject];

    if (!staticProject?.behaviors || staticProject.behaviors.length === 0) {
      return effects;
    }

    const outputs = staticProject.behaviors[0].outputs || [];

    outputs.forEach((output, idx) => {
      const outputType = output.type as string;
      effects.push(
        <GameIcon
          key={`output-${idx}`}
          iconType={outputType}
          amount={output.amount}
          size="small"
        />,
      );
    });

    return effects;
  };

  const getStaticProjectInfo = (projectType: string) => {
    return STANDARD_PROJECTS[projectType as StandardProject];
  };

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "fixed", top: 60, left: 20 }}
      theme="standardProjects"
      excludeRef={anchorRef}
      header={{
        title: "Standard Projects",
        badge: `${availableCount}/${playerProjects.length} Available`,
        showCloseButton: true,
      }}
      width={500}
      maxHeight="calc(100vh - 80px)"
      animation="slideDown"
    >
      <div className="p-2">
        {playerProjects.map((project) => {
          const staticInfo = getStaticProjectInfo(project.projectType);
          const isExecutable = canExecuteProjects && project.available;
          const effects = renderEffects(project.projectType);

          const effectiveCreditCost = project.effectiveCost["credit"] ?? 0;
          const baseCreditCost = project.baseCost["credit"] ?? 0;
          const hasDiscount = effectiveCreditCost < baseCreditCost;

          return (
            <GamePopoverItem
              key={project.projectType}
              state={project.available ? "available" : "disabled"}
              onClick={isExecutable ? () => handleProjectClick(project) : undefined}
              error={
                !project.available && project.errors && project.errors.length > 0
                  ? { message: project.errors[0].message, count: project.errors.length }
                  : undefined
              }
              hoverEffect="background"
              className="mb-2 last:mb-0"
            >
              <div className="flex-1">
                <div className="flex items-start justify-between gap-3 mb-2">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      {staticInfo?.icon && <div className="opacity-70">{staticInfo.icon}</div>}
                      <h3 className="text-white text-sm font-bold font-orbitron m-0">
                        {staticInfo?.name ?? project.projectType}
                      </h3>
                      {staticInfo?.behaviors?.[0]?.outputs?.some((o) =>
                        (o.type as string).includes("-tile"),
                      ) && (
                        <span className="text-[10px] text-white/60 bg-[rgba(var(--popover-accent-rgb),0.3)] px-1.5 py-0.5 rounded">
                          Tile
                        </span>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      {hasDiscount ? (
                        <div className="flex items-center gap-1">
                          <div className="grayscale-[0.7] flex items-center">
                            <GameIcon
                              iconType={ResourceTypeCredit}
                              amount={baseCreditCost}
                              size="small"
                            />
                          </div>
                          <svg
                            width="10"
                            height="8"
                            viewBox="0 0 10 8"
                            className="opacity-70 mx-0.5 flex-shrink-0"
                          >
                            <path
                              d="M10 4 L4 0 L4 2 L0 2 L0 6 L4 6 L4 8 Z"
                              fill="rgba(76, 175, 80, 0.9)"
                            />
                          </svg>
                          <div className="flex items-center">
                            <GameIcon
                              iconType={ResourceTypeCredit}
                              amount={effectiveCreditCost}
                              size="small"
                            />
                          </div>
                        </div>
                      ) : (
                        <GameIcon
                          iconType={ResourceTypeCredit}
                          amount={effectiveCreditCost}
                          size="small"
                        />
                      )}
                      {effects.length > 0 && (
                        <>
                          <span className="text-white/60 text-xs">→</span>
                          {effects}
                        </>
                      )}
                    </div>
                  </div>

                  {canExecuteProjects && (
                    <button
                      className={`flex-shrink-0 px-3 py-1.5 rounded text-xs font-semibold transition-all ${
                        project.available
                          ? "bg-green-600/80 hover:bg-green-600 text-white shadow-sm hover:shadow-md cursor-pointer"
                          : "bg-gray-600/50 text-gray-400"
                      }`}
                      onClick={(e) => {
                        e.stopPropagation();
                        if (isExecutable) handleProjectClick(project);
                      }}
                      disabled={!project.available}
                    >
                      Execute
                    </button>
                  )}
                </div>

                <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                  <FormattedDescription text={staticInfo?.description ?? ""} />
                </p>
              </div>
            </GamePopoverItem>
          );
        })}
      </div>
    </GamePopover>
  );
};

export default StandardProjectPopover;
