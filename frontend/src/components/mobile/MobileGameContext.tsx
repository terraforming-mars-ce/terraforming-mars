import { createContext, useContext } from "react";
import type {
  ChatMessageDto,
  GameDto,
  OtherPlayerDto,
  PlayerActionDto,
  PlayerDto,
} from "@/types/generated/api-types.ts";
import type { StandardProject } from "@/types/cards.tsx";

export interface MobileGameContextValue {
  gameState: GameDto;
  currentPlayer: PlayerDto | null;
  isSpectator: boolean;
  isReplay: boolean;
  playerColorMap: Map<string, string>;
  onPlayCard: (cardId: string) => void;
  onInspectCard: (cardId: string, source: HTMLElement) => void;
  onStandardProjectSelect: (project: StandardProject) => void;
  onActionSelect: (action: PlayerActionDto) => void;
  onConvertPlantsToGreenery: () => void;
  onConvertHeatToTemperature: () => void;
  onPlayerClick: (player: PlayerDto | OtherPlayerDto) => void;
  onSendChatMessage: (message: string) => void;
  chatMessages: ChatMessageDto[];
  onLeaveGame: () => void;
  onEndGame: () => void;
}

export const MobileGameContext = createContext<MobileGameContextValue | null>(null);

export function useMobileGame(): MobileGameContextValue {
  const value = useContext(MobileGameContext);
  if (!value) {
    throw new Error("useMobileGame must be used inside MobileGameContext.Provider");
  }
  return value;
}
