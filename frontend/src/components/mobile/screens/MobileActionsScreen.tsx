import { useState } from "react";
import {
  ResourceTypeGreeneryTile,
  ResourceTypeTemperature,
  type GameDto,
  type PlayerActionDto,
  type PlayerStandardProjectDto,
} from "@/types/generated/api-types.ts";
import { useMobileUiStore } from "@/stores/mobileUiStore.ts";
import GameButton from "../../ui/buttons/GameButton.tsx";
import GameIcon from "../../ui/display/GameIcon.tsx";
import { GamePopoverItem, getThemeStyles } from "../../ui/GamePopover";
import ActionsList from "../../ui/popover/content/ActionsList.tsx";
import MobileScreen, { type MobileScreenTab } from "../MobileScreen.tsx";
import { ActionsIcon } from "../screenIcons.tsx";
import { useMobileGame } from "../MobileGameContext.tsx";
import { getConversionAvailability } from "../conversionAvailability.ts";

const MAX_TEMPERATURE = 8;

type ActionsTab = "card-actions" | "conversions";

interface ConversionRowProps {
  label: string;
  project: PlayerStandardProjectDto | undefined;
  resultIcon: string;
  available: boolean;
  unavailableReason: string;
  onUse: () => void;
}

function ConversionRow({
  label,
  project,
  resultIcon,
  available,
  unavailableReason,
  onUse,
}: ConversionRowProps) {
  const costs = Object.entries(project?.effectiveCost ?? {});
  const error = available ? undefined : { message: unavailableReason };
  const warningMessage = available ? project?.warnings?.[0]?.message : undefined;
  const warning = warningMessage ? { message: warningMessage } : undefined;
  const hasBadge = !!error || !!warning;

  return (
    <GamePopoverItem
      state={available ? "available" : "disabled"}
      borderColor={project?.style?.color ?? "#6b7280"}
      error={error}
      warning={warning}
      className="popover-list-item min-h-14"
    >
      <div className="flex items-end gap-3">
        <div className="flex-1 min-w-0 flex flex-col gap-1">
          <span className="font-orbitron text-[13px] font-bold text-white">{label}</span>
          <div className="flex items-center gap-1.5">
            {costs.map(([resource, amount]) => (
              <span key={resource} className="inline-flex items-center gap-0.5">
                <span className="font-orbitron text-[13px] font-bold text-white/80">{amount}</span>
                <GameIcon iconType={resource} size="small" />
              </span>
            ))}
            <span className="text-white/60 text-[13px]" aria-hidden="true">
              →
            </span>
            <GameIcon iconType={resultIcon} size="small" />
          </div>
        </div>
        <GameButton
          size="md"
          tone="success"
          className={`shrink-0 !min-h-11 min-w-[72px] ${hasBadge ? "mt-8" : ""}`}
          disabled={!available}
          onClick={onUse}
        >
          Use
        </GameButton>
      </div>
    </GamePopoverItem>
  );
}

function conversionReason(
  project: PlayerStandardProjectDto | undefined,
  gameState: GameDto,
  isHeat: boolean,
): string {
  const error = project?.errors?.[0];
  if (error) {
    return error.message;
  }
  if (isHeat && (gameState.globalParameters?.temperature ?? -30) >= MAX_TEMPERATURE) {
    return "Temperature is at maximum";
  }
  return "Not available right now";
}

export default function MobileActionsScreen() {
  const close = useMobileUiStore((s) => s.close);
  const {
    gameState,
    currentPlayer,
    onActionSelect,
    onConvertPlantsToGreenery,
    onConvertHeatToTemperature,
  } = useMobileGame();
  const [tab, setTab] = useState<ActionsTab>("card-actions");

  const conversions = getConversionAvailability(gameState, currentPlayer);
  const projects = gameState.currentPlayer?.standardProjects ?? [];
  const plantsProject = projects.find((p) => p.projectType === "convert-plants-to-greenery");
  const heatProject = projects.find((p) => p.projectType === "convert-heat-to-temperature");
  const actions = currentPlayer?.actions ?? [];

  const tabs: (MobileScreenTab & { id: ActionsTab })[] = [
    {
      id: "card-actions",
      label: "Card actions",
      badge: actions.filter((action) => action.available).length,
    },
    {
      id: "conversions",
      label: "Conversions",
      badge: Number(conversions.plants) + Number(conversions.heat),
    },
  ];

  const startFlow = (handler: () => void) => {
    close();
    handler();
  };

  const handleActionSelect = (action: PlayerActionDto) => {
    startFlow(() => onActionSelect(action));
  };

  return (
    <MobileScreen
      icon={ActionsIcon}
      title="Actions"
      tabs={tabs}
      activeTab={tab}
      onTabChange={(id) => setTab(id === "conversions" ? "conversions" : "card-actions")}
      onClose={close}
    >
      <div className="p-3">
        {tab === "card-actions" ? (
          <ActionsList
            actions={actions}
            gameState={gameState}
            onActionSelect={handleActionSelect}
            density="screen"
          />
        ) : (
          <div className="popover-list grid grid-cols-2 gap-2" style={getThemeStyles("actions")}>
            <ConversionRow
              label="Plants to greenery"
              project={plantsProject}
              resultIcon={ResourceTypeGreeneryTile}
              available={conversions.plants}
              unavailableReason={conversionReason(plantsProject, gameState, false)}
              onUse={() => startFlow(onConvertPlantsToGreenery)}
            />
            <ConversionRow
              label="Heat to temperature"
              project={heatProject}
              resultIcon={ResourceTypeTemperature}
              available={conversions.heat}
              unavailableReason={conversionReason(heatProject, gameState, true)}
              onUse={() => startFlow(onConvertHeatToTemperature)}
            />
          </div>
        )}
      </div>
    </MobileScreen>
  );
}
