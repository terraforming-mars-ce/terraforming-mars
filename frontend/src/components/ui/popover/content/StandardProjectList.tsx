import {
  GameDto,
  GameStatusActive,
  PlayerStandardProjectDto,
} from "@/types/generated/api-types.ts";
import { StandardProject } from "@/types/cards.tsx";
import GameIcon from "../../display/GameIcon.tsx";
import { canPerformActions, isPlayerActionPhase } from "@/utils/actionUtils.ts";
import { GamePopoverItem, getThemeStyles } from "../../GamePopover";
import { FormattedDescription } from "../../display/FormattedDescription";
import BehaviorSection from "../../cards/BehaviorSection/BehaviorSection.tsx";
import type { ContentDensity } from "./density.ts";

interface StandardProjectListProps {
  gameState?: GameDto;
  onProjectSelect: (project: StandardProject) => void;
  density: ContentDensity;
}

function canExecuteStandardProjects(gameState: GameDto | undefined): boolean {
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = isPlayerActionPhase(gameState?.currentPhase);
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
  return isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);
}

export default function StandardProjectList({
  gameState,
  onProjectSelect,
  density,
}: StandardProjectListProps) {
  const isScreen = density === "screen";
  const canExecuteProjects = canExecuteStandardProjects(gameState);

  const playerProjects: PlayerStandardProjectDto[] = [
    ...(gameState?.currentPlayer?.standardProjects ?? []),
  ].sort((a, b) => (a.effectiveCost["credit"] ?? 0) - (b.effectiveCost["credit"] ?? 0));

  const handleProjectClick = (project: PlayerStandardProjectDto) => {
    if (!canExecuteProjects || !project.available) {
      return;
    }
    onProjectSelect(project.projectType as StandardProject);
  };

  return (
    <div
      className={`popover-list flex flex-col gap-2 ${isScreen ? "w-full" : "p-2"}`}
      style={isScreen ? getThemeStyles("colonies") : undefined}
    >
      {playerProjects.map((project) => {
        const isExecutable = canExecuteProjects && project.available;
        const styleColor = project.style?.color ?? "#6b7280";
        const error =
          !project.available && project.errors?.length
            ? { message: project.errors[0].message, count: project.errors.length }
            : undefined;
        const warning =
          project.available && project.warnings?.length
            ? { message: project.warnings[0].message }
            : undefined;

        return (
          <GamePopoverItem
            key={project.projectType}
            className={`popover-list-item ${isScreen ? "min-h-14" : ""}`}
            state={project.available ? "available" : "disabled"}
            onClick={isExecutable ? () => handleProjectClick(project) : undefined}
            clickSound={false}
            borderColor={styleColor}
            error={error}
            warning={warning}
          >
            <div className={`flex-1 ${isScreen ? "min-w-0" : ""}`}>
              <div className="flex items-center gap-2 mb-2">
                {project.style?.icon && (
                  <div className="opacity-70 flex items-center">
                    <GameIcon iconType={project.style.icon} size="small" />
                  </div>
                )}
                <h3 className="text-white text-sm font-bold font-orbitron m-0">{project.name}</h3>
                {project.behaviors?.[0]?.outputs?.some((o) =>
                  (o.type as string).includes("-tile"),
                ) && (
                  <span
                    className={`text-white/60 px-1.5 py-0.5 rounded ${isScreen ? "text-[11px]" : "text-[10px]"}`}
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

              <p
                className={`text-white/70 leading-relaxed m-0 text-left ${isScreen ? "text-[13px]" : "text-xs"}`}
              >
                <FormattedDescription text={project.description ?? ""} />
              </p>
            </div>
          </GamePopoverItem>
        );
      })}
    </div>
  );
}
