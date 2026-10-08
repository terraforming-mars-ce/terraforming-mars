import React from "react";
import { GameDto } from "@/types/generated/api-types.ts";
import { StandardProject } from "@/types/cards.tsx";
import { GamePopover } from "../GamePopover";
import StandardProjectList from "./content/StandardProjectList.tsx";

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
  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "below" }}
      theme="colonies"
      excludeRef={anchorRef}
      header={undefined}
      width={500}
      maxHeight="80dvh"
      animation="slideDown"
      className="game-popover-list"
    >
      <StandardProjectList
        gameState={gameState}
        onProjectSelect={onProjectSelect}
        density="popover"
      />
    </GamePopover>
  );
};

export default StandardProjectPopover;
