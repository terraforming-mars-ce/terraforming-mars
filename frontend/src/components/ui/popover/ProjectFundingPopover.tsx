import React, { useState, useCallback } from "react";
import {
  CardPaymentDto,
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ProjectFundingDto,
  ProjectSeatDto,
  ColonyOutputDto,
  ProjectGlobalOutputDto,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { webSocketService } from "@/services/webSocketService.ts";
import { canPerformActions } from "@/utils/actionUtils.ts";
import { GamePopover, GamePopoverItem } from "../GamePopover";
import { mapOutputTypeToIcon } from "./ColonySteps.tsx";
import PaymentSelectionPopover from "./PaymentSelectionPopover.tsx";
import type { GenericPaymentConfig } from "./PaymentSelectionPopover.tsx";

interface ProjectFundingPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

const ProjectFundingPopover: React.FC<ProjectFundingPopoverProps> = ({
  isVisible,
  onClose,
  gameState,
  anchorRef,
}) => {
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
  const canAct =
    isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);

  const projects = gameState?.projectFunding ?? [];

  const [pendingProject, setPendingProject] = useState<ProjectFundingDto | null>(null);
  const resources = gameState?.currentPlayer?.resources;

  const paymentConfig: GenericPaymentConfig | undefined = pendingProject
    ? {
        name: `${pendingProject.name} Seat`,
        cost: pendingProject.nextSeatCost,
        substitutes: pendingProject.paymentSubstitutes.map((s) => ({
          resourceType: s.resourceType,
          conversionRate: s.conversionRate,
        })),
      }
    : undefined;

  const handlePaymentConfirm = useCallback(
    (payment: CardPaymentDto) => {
      if (pendingProject) {
        void webSocketService.buyProjectSeat(
          pendingProject.id,
          payment.credits,
          payment.steel,
          payment.titanium,
        );
        setPendingProject(null);
      }
    },
    [pendingProject],
  );

  const handlePaymentCancel = useCallback(() => {
    setPendingProject(null);
  }, []);

  const handleBuySeat = useCallback((project: ProjectFundingDto) => {
    if (project.paymentSubstitutes.length > 0) {
      setPendingProject(project);
    } else {
      void webSocketService.buyProjectSeat(project.id, project.nextSeatCost, 0, 0);
    }
  }, []);

  return (
    <>
      <GamePopover
        isVisible={isVisible}
        onClose={onClose}
        position={{ type: "anchor", anchorRef, placement: "below" }}
        theme="colonies"
        excludeRef={anchorRef}
        header={undefined}
        width={560}
        maxHeight="80vh"
        animation="slideDown"
        className="!bg-space-black-darker"
      >
        <div className="p-2 space-y-2">
          {projects.map((project) => (
            <ProjectCard
              key={project.id}
              project={project}
              canAct={canAct}
              onBuySeat={handleBuySeat}
            />
          ))}
        </div>
      </GamePopover>

      {resources && paymentConfig && (
        <PaymentSelectionPopover
          genericPayment={paymentConfig}
          playerResources={resources}
          onConfirm={handlePaymentConfirm}
          onCancel={handlePaymentCancel}
          isVisible={!!pendingProject}
        />
      )}
    </>
  );
};

interface ProjectCardProps {
  project: ProjectFundingDto;
  canAct: boolean;
  onBuySeat: (project: ProjectFundingDto) => void;
}

const ProjectCard: React.FC<ProjectCardProps> = ({ project, canAct, onBuySeat }) => {
  const canBuy = canAct && project.canBuySeat;
  const dimmed = canAct && !canBuy && !project.isCompleted;

  const filledSeats = project.seats.filter((s) => s.isFilled).length;
  const totalSeats = project.seats.length;

  return (
    <GamePopoverItem
      state={dimmed ? "disabled" : "available"}
      borderColor={project.isCompleted ? "#10b981" : project.style.color}
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <h3 className="text-white text-sm font-bold font-orbitron m-0">{project.name}</h3>
          <span className="text-[10px] font-orbitron text-white/40">
            {filledSeats}/{totalSeats}
          </span>
          {project.isCompleted && (
            <span className="px-2 py-0.5 rounded-full text-[10px] font-orbitron font-bold text-emerald-400 bg-emerald-400/15">
              COMPLETED
            </span>
          )}
        </div>

        {canAct && !project.isCompleted && (
          <button
            className={`px-3 py-1 rounded text-xs font-semibold font-orbitron transition-all cursor-pointer ${
              canBuy ? "bg-white/15 hover:bg-white/25 text-white" : "bg-gray-600/30 text-gray-500"
            }`}
            onClick={(e) => {
              e.stopPropagation();
              if (canBuy) {
                onBuySeat(project);
              }
            }}
            disabled={!canBuy}
          >
            Buy Seat
          </button>
        )}
      </div>

      <div className="text-xs text-white/50 mb-2">{project.description}</div>

      <SeatBar seats={project.seats} styleColor={project.style.color} />

      {!project.isCompleted && project.nextSeatCost > 0 && (
        <div className="flex items-center gap-1.5 mt-2 text-xs text-white/60">
          <span className="font-orbitron text-[10px] uppercase tracking-wider">Next Seat</span>
          <GameIcon iconType="credit" amount={project.nextSeatCost} size="small" />
          {project.paymentSubstitutes.map((sub) => (
            <span key={sub.resourceType} className="text-[10px] text-white/40">
              ({sub.resourceType} {sub.conversionRate}:1)
            </span>
          ))}
        </div>
      )}

      <div style={{ height: "1px", background: "rgba(255,255,255,0.15)", margin: "10px 0" }} />

      {project.firstFunderBonus && project.firstFunderBonus.length > 0 && (
        <div className="flex items-center gap-1.5 mb-2">
          <span className="text-[10px] font-orbitron text-amber-400/70 uppercase tracking-wider">
            First Funder
          </span>
          <OutputDisplay outputs={project.firstFunderBonus} />
        </div>
      )}

      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-1">
          <span className="text-[10px] font-orbitron text-white/40 uppercase tracking-wider">
            Rewards
          </span>
          <div className="flex items-center gap-2">
            {project.rewardTiers.map((tier) => (
              <div key={tier.seatsOwned} className="flex items-center gap-1">
                <span className="text-[10px] text-white/50 font-orbitron">{tier.seatsOwned}x:</span>
                <OutputDisplay outputs={tier.rewards} />
              </div>
            ))}
          </div>
        </div>

        <div className="flex flex-col gap-1 items-end">
          <span className="text-[10px] font-orbitron text-white/40 uppercase tracking-wider">
            Completion
          </span>
          <div className="flex flex-col items-end gap-0.5">
            {project.completionEffect.rewards.length > 0 && (
              <OutputDisplay outputs={project.completionEffect.rewards} />
            )}
            {project.completionEffect.globalEffects &&
              project.completionEffect.globalEffects.length > 0 && (
                <GlobalEffectsDisplay effects={project.completionEffect.globalEffects} />
              )}
          </div>
        </div>
      </div>

      {project.currentPlayerSeats > 0 && (
        <div className="mt-2 flex items-center gap-2">
          <span className="text-[10px] font-orbitron text-white/40">
            Your seats: {project.currentPlayerSeats}
          </span>
          {project.currentPlayerTier && (
            <span className="text-[10px] font-orbitron text-emerald-400/80">
              Tier {project.currentPlayerTier.seatsOwned} reward
            </span>
          )}
        </div>
      )}
    </GamePopoverItem>
  );
};

interface SeatBarProps {
  seats: ProjectSeatDto[];
  styleColor: string;
}

const SeatBar: React.FC<SeatBarProps> = ({ seats, styleColor }) => {
  return (
    <div className="flex gap-1">
      {seats.map((seat, i) => (
        <div
          key={i}
          className="flex-1 h-5 rounded-sm flex items-center justify-center text-[9px] font-orbitron font-bold transition-colors"
          style={{
            backgroundColor: seat.isFilled
              ? (seat.ownerColor || styleColor) + "80"
              : "rgba(255,255,255,0.06)",
            border: `1px solid ${seat.isFilled ? (seat.ownerColor || styleColor) + "60" : "rgba(255,255,255,0.1)"}`,
          }}
        >
          {seat.isFilled ? (
            <span className="text-white/80 truncate px-0.5">{seat.ownerName}</span>
          ) : (
            <span className="text-white/20">{seat.cost}</span>
          )}
        </div>
      ))}
    </div>
  );
};

interface OutputDisplayProps {
  outputs: ColonyOutputDto[];
}

const OutputDisplay: React.FC<OutputDisplayProps> = ({ outputs }) => {
  return (
    <span className="inline-flex items-center gap-0.5">
      {outputs.map((output, i) => {
        const icon = mapOutputTypeToIcon(output.type);
        const useAmountProp = output.type === "credit" || output.type === "credit-production";
        return (
          <span key={i} className="inline-flex items-center gap-0.5">
            {!useAmountProp && output.amount > 1 && (
              <span className="text-xs text-white/70 font-orbitron font-bold">{output.amount}</span>
            )}
            <GameIcon
              iconType={icon}
              amount={useAmountProp ? output.amount : undefined}
              size="small"
            />
          </span>
        );
      })}
    </span>
  );
};

interface GlobalEffectsDisplayProps {
  effects: ProjectGlobalOutputDto[];
}

const globalEffectLabel = (effect: ProjectGlobalOutputDto): string => {
  switch (effect.type) {
    case "freeze-turn-order":
      return "Freeze turn order";
    case "production-choice":
      return `Each player: +${effect.amount} any production`;
    case "card-draw":
      return `Each player: +${effect.amount} cards`;
    case "temperature":
      return `+${effect.amount} temperature`;
    case "oxygen":
      return `+${effect.amount} oxygen`;
    default:
      return effect.type;
  }
};

const GlobalEffectsDisplay: React.FC<GlobalEffectsDisplayProps> = ({ effects }) => {
  return (
    <div className="flex flex-col items-end gap-0.5">
      {effects.map((effect, i) => (
        <span key={i} className="text-[10px] font-orbitron text-purple-300/80">
          {globalEffectLabel(effect)}
        </span>
      ))}
    </div>
  );
};

export default ProjectFundingPopover;
