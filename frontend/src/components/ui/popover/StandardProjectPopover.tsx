import React from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  PlayerStandardProjectDto,
} from "@/types/generated/api-types.ts";
import { StandardProject } from "@/types/cards.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { canPerformActions } from "@/utils/actionUtils.ts";
import { GamePopover, GamePopoverItem } from "../GamePopover";
import { FormattedDescription } from "../display/FormattedDescription";
import GameButton from "../buttons/GameButton.tsx";
import BehaviorSection from "../cards/BehaviorSection/BehaviorSection.tsx";

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

  const playerProjects: PlayerStandardProjectDto[] =
    gameState?.currentPlayer?.standardProjects ?? [];
  const availableCount = playerProjects.filter((p) => p.available).length;

  const handleProjectClick = (project: PlayerStandardProjectDto) => {
    if (!canExecuteProjects || !project.available) return;
    onProjectSelect(project.projectType as StandardProject);
  };

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "fixed", top: 60, left: 20 }}
      theme="colonies"
      excludeRef={anchorRef}
      header={{
        title: "Standard Projects",
        badge: `${availableCount}/${playerProjects.length} Available`,
        rightContent: (
          <GameButton buttonType="textonly" size="xs" onClick={onClose}>
            ✕
          </GameButton>
        ),
      }}
      width={500}
      maxHeight="80vh"
      animation="slideDown"
    >
      <div className="p-2 flex flex-col gap-2">
        {playerProjects.map((project) => {
          const isExecutable = canExecuteProjects && project.available;
          const styleColor = project.style?.color ?? "#6b7280";

          return (
            <GamePopoverItem
              key={project.projectType}
              state={project.available ? "available" : "disabled"}
              onClick={isExecutable ? () => handleProjectClick(project) : undefined}
              borderColor={styleColor}
              error={
                !project.available && project.errors?.length
                  ? { message: project.errors[0].message, count: project.errors.length }
                  : undefined
              }
              warning={
                project.available && project.warnings?.length
                  ? { message: project.warnings[0].message }
                  : undefined
              }
            >
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  {project.style?.icon && (
                    <div className="opacity-70">
                      <GameIcon iconType={project.style.icon} size="small" />
                    </div>
                  )}
                  <h3 className="text-white text-sm font-bold font-orbitron m-0">{project.name}</h3>
                  {project.behaviors?.[0]?.outputs?.some((o) =>
                    (o.type as string).includes("-tile"),
                  ) && (
                    <span
                      className="text-[10px] text-white/60 px-1.5 py-0.5 rounded"
                      style={{ background: `${styleColor}33` }}
                    >
                      Tile
                    </span>
                  )}
                </div>

                {project.behaviors && project.behaviors.length > 0 && (
                  <div className="[&>div]:items-start [&_div]:justify-start mb-2">
                    <BehaviorSection behaviors={project.behaviors} noContainer />
                  </div>
                )}

                <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                  <FormattedDescription text={project.description ?? ""} />
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
