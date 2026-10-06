import { create } from "zustand";
import type { EmotePayload } from "@/types/generated/api-types.ts";

export type EmoteKind = EmotePayload["emote"];

export interface BotThought {
  text: string;
  seq: number;
}

export interface PlayerEmote {
  emote: EmoteKind;
  seq: number;
}

interface BotPresenceState {
  thoughts: Record<string, BotThought>;
  emotes: Record<string, PlayerEmote>;
  typing: Record<string, boolean>;
  setThought: (playerId: string, text: string) => void;
  clearThought: (playerId: string, seq: number) => void;
  setEmote: (playerId: string, emote: EmoteKind) => void;
  clearEmote: (playerId: string, seq: number) => void;
  setTyping: (playerId: string, typing: boolean) => void;
  reset: () => void;
}

let nextSeq = 0;

function without<T>(record: Record<string, T>, key: string): Record<string, T> {
  const next = { ...record };
  delete next[key];
  return next;
}

export const useBotPresenceStore = create<BotPresenceState>((set) => ({
  thoughts: {},
  emotes: {},
  typing: {},
  setThought: (playerId, text) =>
    set((state) => ({
      thoughts: { ...state.thoughts, [playerId]: { text, seq: nextSeq++ } },
    })),
  clearThought: (playerId, seq) =>
    set((state) => {
      if (state.thoughts[playerId]?.seq !== seq) {
        return state;
      }
      return { thoughts: without(state.thoughts, playerId) };
    }),
  setEmote: (playerId, emote) =>
    set((state) => ({
      emotes: { ...state.emotes, [playerId]: { emote, seq: nextSeq++ } },
    })),
  clearEmote: (playerId, seq) =>
    set((state) => {
      if (state.emotes[playerId]?.seq !== seq) {
        return state;
      }
      return { emotes: without(state.emotes, playerId) };
    }),
  setTyping: (playerId, typing) =>
    set((state) => {
      if (typing) {
        return { typing: { ...state.typing, [playerId]: true } };
      }
      return { typing: without(state.typing, playerId) };
    }),
  reset: () => set({ thoughts: {}, emotes: {}, typing: {} }),
}));

const SHOW_BOT_THOUGHTS_KEY = "terraforming-mars-show-bot-thoughts";

function readShowBotThoughts(): boolean {
  try {
    const stored = localStorage.getItem(SHOW_BOT_THOUGHTS_KEY);
    if (stored === null) {
      return true;
    }
    return stored === "true";
  } catch {
    return true;
  }
}

interface BotThoughtsPreferenceState {
  showBotThoughts: boolean;
  setShowBotThoughts: (show: boolean) => void;
}

export const useBotThoughtsPreferenceStore = create<BotThoughtsPreferenceState>((set) => ({
  showBotThoughts: readShowBotThoughts(),
  setShowBotThoughts: (show) => {
    try {
      localStorage.setItem(SHOW_BOT_THOUGHTS_KEY, String(show));
    } catch {
      console.warn("Failed to save bot thoughts preference to localStorage");
    }
    set({ showBotThoughts: show });
  },
}));
