import React from "react";
import { GameDto } from "@/types/generated/api-types.ts";
import { GamePopover } from "../GamePopover";
import AwardList from "./content/AwardList.tsx";

interface AwardPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

const AwardPopover: React.FC<AwardPopoverProps> = ({
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
      <AwardList gameState={gameState} density="popover" />
    </GamePopover>
  );
};

export default AwardPopover;
