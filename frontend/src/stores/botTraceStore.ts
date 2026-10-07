import { create } from "zustand";
import type { BotCallDto, BotTraceDto, BotTraceEventDto } from "@/types/generated/api-types.ts";

const MAX_CALLS = 15;
const MAX_REACTIONS = 30;

interface BotTraceState {
  playerId: string | null;
  trace: BotTraceDto | null;
  error: string | null;
  begin: (playerId: string) => void;
  applySnapshot: (snapshot: BotTraceDto) => void;
  applyEvent: (event: BotTraceEventDto) => void;
  fail: (message: string) => void;
  reset: () => void;
}

function updateCall(
  calls: BotCallDto[],
  id: string,
  update: (call: BotCallDto) => BotCallDto,
): BotCallDto[] {
  const index = calls.findIndex((call) => call.id === id);
  if (index === -1) {
    return calls;
  }
  const next = calls.slice();
  next[index] = update(calls[index]);
  return next;
}

function reduceTrace(trace: BotTraceDto, event: BotTraceEventDto): BotTraceDto {
  switch (event.kind) {
    case "call-start": {
      if (!event.call) {
        return trace;
      }
      const call: BotCallDto = { ...event.call, steps: event.call.steps ?? [] };
      return { ...trace, calls: [...trace.calls, call].slice(-MAX_CALLS) };
    }
    case "call-step": {
      const step = event.step;
      if (!event.callId || !step) {
        return trace;
      }
      return {
        ...trace,
        calls: updateCall(trace.calls, event.callId, (call) => ({
          ...call,
          steps: [...call.steps, step],
        })),
      };
    }
    case "call-end": {
      const ended = event.call;
      if (!ended) {
        return trace;
      }
      return {
        ...trace,
        calls: updateCall(trace.calls, ended.id, (call) => ({
          ...call,
          ...ended,
          prompt: call.prompt,
          steps: call.steps,
        })),
      };
    }
    case "plan": {
      if (!event.plan) {
        return trace;
      }
      return { ...trace, plan: event.plan, planUpdatedAt: event.planUpdatedAt };
    }
    case "reaction": {
      if (!event.reaction) {
        return trace;
      }
      return { ...trace, reactions: [...trace.reactions, event.reaction].slice(-MAX_REACTIONS) };
    }
    case "grudges": {
      return { ...trace, grudges: event.grudges ?? [] };
    }
    default:
      return trace;
  }
}

export const useBotTraceStore = create<BotTraceState>((set) => ({
  playerId: null,
  trace: null,
  error: null,
  begin: (playerId) => set({ playerId, trace: null, error: null }),
  applySnapshot: (snapshot) =>
    set((state) => {
      if (snapshot.playerId !== state.playerId) {
        return state;
      }
      return {
        error: null,
        trace: {
          ...snapshot,
          calls: (snapshot.calls ?? []).slice(-MAX_CALLS),
          reactions: (snapshot.reactions ?? []).slice(-MAX_REACTIONS),
          grudges: snapshot.grudges ?? [],
        },
      };
    }),
  applyEvent: (event) =>
    set((state) => {
      if (!state.trace || event.playerId !== state.playerId) {
        return state;
      }
      return { trace: reduceTrace(state.trace, event) };
    }),
  fail: (message) =>
    set((state) => {
      if (!state.playerId || state.trace) {
        return state;
      }
      return { error: message };
    }),
  reset: () => set({ playerId: null, trace: null, error: null }),
}));
