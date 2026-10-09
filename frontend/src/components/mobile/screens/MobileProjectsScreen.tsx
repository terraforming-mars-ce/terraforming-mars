import type { ReactNode } from "react";
import type { GameDto } from "@/types/generated/api-types.ts";
import type { StandardProject } from "@/types/cards.tsx";
import { canPerformActions } from "@/utils/actionUtils.ts";
import { useMobileUiStore, type MobileProjectsTab } from "@/stores/mobileUiStore.ts";
import StandardProjectList from "../../ui/popover/content/StandardProjectList.tsx";
import MilestoneList from "../../ui/popover/content/MilestoneList.tsx";
import AwardList from "../../ui/popover/content/AwardList.tsx";
import ColonyList, { useColonyControls } from "../../ui/popover/content/ColonyList.tsx";
import ProjectFundingList, {
  useProjectSeatPurchase,
} from "../../ui/popover/content/ProjectFundingList.tsx";
import MobileScreen, { type MobileScreenTab } from "../MobileScreen.tsx";
import { ProjectsIcon } from "../screenIcons.tsx";
import { useMobileGame } from "../MobileGameContext.tsx";

function availabilityByTab(gameState: GameDto): Record<MobileProjectsTab, boolean> {
  const canAct = canPerformActions(gameState);
  const viewer = gameState.currentPlayer;
  return {
    standard: canAct && !!viewer?.standardProjects?.some((p) => p.available),
    milestones: canAct && !!viewer?.milestones?.some((m) => m.available),
    awards: canAct && !!viewer?.awards?.some((a) => a.available),
    colonies: canAct && !!gameState.colonies?.some((c) => c.tradeAvailable || c.buildAvailable),
    funding: canAct && !!gameState.projectFunding?.some((p) => p.canBuySeat),
  };
}

export default function MobileProjectsScreen() {
  const close = useMobileUiStore((s) => s.close);
  const projectsTab = useMobileUiStore((s) => s.projectsTab);
  const setProjectsTab = useMobileUiStore((s) => s.setProjectsTab);
  const { gameState, onStandardProjectSelect } = useMobileGame();

  const colonyControls = useColonyControls(gameState, close);
  const seatPurchase = useProjectSeatPurchase(gameState, close);

  const hasColonies = (gameState.colonies?.length ?? 0) > 0;
  const hasProjectFunding = (gameState.projectFunding?.length ?? 0) > 0;
  const available = availabilityByTab(gameState);

  const tabs: (MobileScreenTab & { id: MobileProjectsTab })[] = [
    { id: "standard", label: "Standard Projects", badge: available.standard },
    { id: "milestones", label: "Milestones", badge: available.milestones },
    { id: "awards", label: "Awards", badge: available.awards },
    ...(hasColonies
      ? [{ id: "colonies" as const, label: "Colonies", badge: available.colonies }]
      : []),
    ...(hasProjectFunding
      ? [{ id: "funding" as const, label: "Funding", badge: available.funding }]
      : []),
  ];

  const activeTab = tabs.some((tab) => tab.id === projectsTab) ? projectsTab : "standard";

  const handleStandardProjectSelect = (project: StandardProject) => {
    close();
    onStandardProjectSelect(project);
  };

  let content: ReactNode;
  switch (activeTab) {
    case "milestones":
      content = <MilestoneList gameState={gameState} density="screen" onFlowStart={close} />;
      break;
    case "awards":
      content = <AwardList gameState={gameState} density="screen" onFlowStart={close} />;
      break;
    case "colonies":
      content = <ColonyList gameState={gameState} controls={colonyControls} density="screen" />;
      break;
    case "funding":
      content = (
        <ProjectFundingList gameState={gameState} purchase={seatPurchase} density="screen" />
      );
      break;
    default:
      content = (
        <StandardProjectList
          gameState={gameState}
          onProjectSelect={handleStandardProjectSelect}
          density="screen"
        />
      );
  }

  return (
    <MobileScreen
      icon={ProjectsIcon}
      title="Projects"
      tabs={tabs}
      activeTab={activeTab}
      onTabChange={(tab) => setProjectsTab(tab as MobileProjectsTab)}
      onClose={close}
    >
      <div className="p-3">{content}</div>
    </MobileScreen>
  );
}
