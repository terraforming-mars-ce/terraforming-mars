import React from "react";
import { GameDto } from "@/types/generated/api-types.ts";
import { GamePopover } from "../GamePopover";
import MilestoneList from "./content/MilestoneList.tsx";

interface MilestonePopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

const MilestonePopover: React.FC<MilestonePopoverProps> = ({
  isVisible,
  onClose,
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
      <MilestoneList gameState={gameState} density="popover" />
    </GamePopover>
  );
};

export default MilestonePopover;
