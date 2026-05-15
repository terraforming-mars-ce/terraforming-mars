import { create } from "zustand";
import type {
  CardDto,
  ChatMessageDto,
  GameDto,
  PlayerDto,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";

interface GameStoreState {
  game: GameDto | null;
  gameRef: GameDto | null;
  currentPlayer: PlayerDto | null;
  playerId: string | null;
  isConnected: boolean;
  isSpectator: boolean;
  isReconnecting: boolean;
  reconnectionStep: "game" | "environment" | null;
  gameForSelection: GameDto | null;
  changedPaths: Set<string>;
  triggeredEffects: TriggeredEffectDto[];
  chatMessages: ChatMessageDto[];
  corporationData: CardDto | null;
  showCorp: boolean;
  displayedInitPlayerId: string | null;

  setGame: (game: GameDto | null) => void;
  setGameRef: (game: GameDto | null) => void;
  setCurrentPlayer: (player: PlayerDto | null) => void;
  setPlayerId: (id: string | null) => void;
  setIsConnected: (connected: boolean) => void;
  setIsSpectator: (spectator: boolean) => void;
  setIsReconnecting: (reconnecting: boolean) => void;
  setReconnectionStep: (step: "game" | "environment" | null) => void;
  setGameForSelection: (game: GameDto | null) => void;
  setChangedPaths: (paths: Set<string>) => void;
  setTriggeredEffects: (effects: TriggeredEffectDto[]) => void;
  setChatMessages: (msgs: ChatMessageDto[]) => void;
  addChatMessage: (msg: ChatMessageDto) => void;
  setCorporationData: (card: CardDto | null) => void;
  setShowCorp: (show: boolean) => void;
  setDisplayedInitPlayerId: (id: string | null) => void;
  reset: () => void;
}

const initialState = {
  game: null,
  gameRef: null,
  currentPlayer: null,
  playerId: null,
  isConnected: false,
  isSpectator: false,
  isReconnecting: false,
  reconnectionStep: null,
  gameForSelection: null,
  changedPaths: new Set<string>(),
  triggeredEffects: [] as TriggeredEffectDto[],
  chatMessages: [] as ChatMessageDto[],
  corporationData: null,
  showCorp: false,
  displayedInitPlayerId: null,
};

export const useGameStore = create<GameStoreState>((set) => ({
  ...initialState,

  setGame: (game) => set({ game }),
  setGameRef: (gameRef) => set({ gameRef }),
  setCurrentPlayer: (currentPlayer) => set({ currentPlayer }),
  setPlayerId: (playerId) => set({ playerId }),
  setIsConnected: (isConnected) => set({ isConnected }),
  setIsSpectator: (isSpectator) => set({ isSpectator }),
  setIsReconnecting: (isReconnecting) => set({ isReconnecting }),
  setReconnectionStep: (reconnectionStep) => set({ reconnectionStep }),
  setGameForSelection: (gameForSelection) => set({ gameForSelection }),
  setChangedPaths: (changedPaths) => set({ changedPaths }),
  setTriggeredEffects: (triggeredEffects) => set({ triggeredEffects }),
  setChatMessages: (chatMessages) => set({ chatMessages }),
  addChatMessage: (msg) => set((state) => ({ chatMessages: [...state.chatMessages, msg] })),
  setCorporationData: (corporationData) => set({ corporationData }),
  setShowCorp: (showCorp) => set({ showCorp }),
  setDisplayedInitPlayerId: (displayedInitPlayerId) => set({ displayedInitPlayerId }),
  reset: () => set(initialState),
}));
