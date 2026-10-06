import React from "react";
import type { EmoteKind } from "@/stores/botPresenceStore.ts";

export const EMOTES: { kind: EmoteKind; label: string }[] = [
  { kind: "celebrate", label: "Celebrate" },
  { kind: "applause", label: "Applause" },
  { kind: "laugh", label: "Laugh" },
  { kind: "cool", label: "Cool" },
  { kind: "thinking", label: "Thinking" },
  { kind: "shock", label: "Shock" },
  { kind: "sad", label: "Sad" },
  { kind: "angry", label: "Angry" },
];

const emoteColors: Record<EmoteKind, string> = {
  angry: "#f87171",
  celebrate: "#facc15",
  thinking: "#7dd3fc",
  applause: "#fbbf24",
  shock: "#fb923c",
  laugh: "#fde047",
  sad: "#60a5fa",
  cool: "#22d3ee",
};

const Face: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <>
    <circle cx="12" cy="12" r="9.5" fill="currentColor" fillOpacity="0.18" />
    <circle cx="12" cy="12" r="9.5" />
    {children}
  </>
);

const emoteShapes: Record<EmoteKind, React.ReactNode> = {
  angry: (
    <Face>
      <path d="M7 8.5l3 1.5M17 8.5l-3 1.5" />
      <circle cx="9" cy="11.5" r="0.8" fill="currentColor" />
      <circle cx="15" cy="11.5" r="0.8" fill="currentColor" />
      <path d="M8.5 17c1.8-1.6 5.2-1.6 7 0" />
    </Face>
  ),
  celebrate: (
    <>
      <path d="M4 20l4.5-11 6.5 6.5z" fill="currentColor" fillOpacity="0.25" />
      <path d="M4 20l4.5-11 6.5 6.5z" />
      <path d="M14 3v2M19 5l-1.5 1.5M21 10h-2M16 8.5c1-1 2.5-1 3.5 0" />
      <circle cx="12" cy="6" r="0.8" fill="currentColor" />
      <circle cx="19.5" cy="14" r="0.8" fill="currentColor" />
    </>
  ),
  thinking: (
    <Face>
      <path d="M8 8.5l2-.8M14.5 8l2 .6" />
      <circle cx="9" cy="11" r="0.8" fill="currentColor" />
      <circle cx="15" cy="11" r="0.8" fill="currentColor" />
      <path d="M10 16h4.5" />
      <path d="M7 16.5c1 .3 1.8 1.2 2 2.5" />
    </Face>
  ),
  applause: (
    <>
      <path
        d="M7 13l4-6.5a1.4 1.4 0 0 1 2.4 1.4L11.5 11l3.5-5a1.4 1.4 0 0 1 2.3 1.6l-3.6 5.4 1.8-1.8a1.4 1.4 0 0 1 2 2L14 18.5a5 5 0 0 1-7.6-.6 4 4 0 0 1 .6-4.9z"
        fill="currentColor"
        fillOpacity="0.2"
      />
      <path d="M7 13l4-6.5a1.4 1.4 0 0 1 2.4 1.4L11.5 11l3.5-5a1.4 1.4 0 0 1 2.3 1.6l-3.6 5.4 1.8-1.8a1.4 1.4 0 0 1 2 2L14 18.5a5 5 0 0 1-7.6-.6 4 4 0 0 1 .6-4.9z" />
      <path d="M4 7l1.5 1M3 11h2M6 4l.8 1.6" />
    </>
  ),
  shock: (
    <Face>
      <circle cx="9" cy="10" r="1.5" />
      <circle cx="15" cy="10" r="1.5" />
      <ellipse cx="12" cy="16" rx="1.8" ry="2.2" />
    </Face>
  ),
  laugh: (
    <Face>
      <path d="M7.5 10.5c.6-1.2 2.4-1.2 3 0M13.5 10.5c.6-1.2 2.4-1.2 3 0" />
      <path d="M7.5 13.5h9a4.5 4.5 0 0 1-9 0z" fill="currentColor" fillOpacity="0.5" />
    </Face>
  ),
  sad: (
    <Face>
      <circle cx="9" cy="10.5" r="0.8" fill="currentColor" />
      <circle cx="15" cy="10.5" r="0.8" fill="currentColor" />
      <path d="M8.5 17c1.8-1.6 5.2-1.6 7 0" />
      <path d="M16 12.5c-.6 1-.9 1.6-.9 2a.9.9 0 0 0 1.8 0c0-.4-.3-1-.9-2z" fill="currentColor" />
    </Face>
  ),
  cool: (
    <Face>
      <path d="M5 9.5h14" />
      <path
        d="M6 9.5h5v1.5a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2zM13 9.5h5v1.5a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2z"
        fill="currentColor"
      />
      <path d="M9 16.5c2 1 4.5.8 6-.8" />
    </Face>
  ),
};

interface EmoteIconProps {
  emote: EmoteKind;
  size?: number;
}

const EmoteIcon: React.FC<EmoteIconProps> = ({ emote, size = 24 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.6"
    strokeLinecap="round"
    strokeLinejoin="round"
    style={{ color: emoteColors[emote] }}
    aria-hidden="true"
  >
    {emoteShapes[emote]}
  </svg>
);

export default EmoteIcon;
