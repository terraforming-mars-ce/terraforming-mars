import { memo, useEffect, useRef, useState } from "react";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import GameIcon from "../display/GameIcon.tsx";
import BackButton from "../buttons/BackButton.tsx";
import GameButton from "../buttons/GameButton.tsx";
import CopyLinkButton from "../buttons/CopyLinkButton.tsx";
import { CARD_FAMILIES, FAMILY_LABELS, labelFor } from "./cardCatalog.ts";
import type { CardFamily, CardSort } from "./cardCatalog.ts";

export type FilterKind = "tags" | "types" | "packs";
export type BrowserFilters = Record<FilterKind, ReadonlySet<string>>;

interface SidebarProps {
  family: CardFamily;
  query: string;
  sort: CardSort;
  filters: BrowserFilters;
  tags: string[];
  packs: string[];
  counts: Record<CardFamily, number>;
  selectedCounts: Record<CardFamily, number>;
  selectedCount: number;
  shareUrl: string;
  shared: boolean;
  open: boolean;
  backLabel: string;
  onBack: () => void;
  onClose: () => void;
  onFamily: (family: CardFamily) => void;
  onQuery: (query: string) => void;
  onSort: (sort: CardSort) => void;
  onToggle: (kind: FilterKind, value: string) => void;
  onClearFilters: () => void;
  onClearSelection: () => void;
  onExitShared: () => void;
}

const fieldClass =
  "w-full rounded-none border border-space-blue-500 bg-black px-3 py-2.5 text-sm text-white outline-none focus-visible:border-blue-400 focus-visible:ring-1 focus-visible:ring-blue-400";

export default memo(function CardBrowserSidebar(props: SidebarProps) {
  const { playButtonClickSound } = useSoundEffects();
  const panel = useRef<HTMLElement>(null);
  const [copyError, setCopyError] = useState(false);
  const { open, onClose } = props;
  useEffect(() => {
    if (!open) {
      return;
    }
    const previouslyFocused = document.activeElement as HTMLElement | null;
    panel.current?.querySelector<HTMLElement>("[data-drawer-close]")?.focus();
    const keydown = (event: globalThis.KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        event.stopPropagation();
        onClose();
      } else if (event.key === "Tab") {
        const focusable = Array.from(
          panel.current?.querySelectorAll<HTMLElement>(
            'button:not(:disabled), input, select, a[href], summary, [tabindex="0"]',
          ) ?? [],
        ).filter((element) => element.getClientRects().length > 0);
        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        if (event.shiftKey && document.activeElement === first) {
          event.preventDefault();
          last?.focus();
        } else if (!event.shiftKey && document.activeElement === last) {
          event.preventDefault();
          first?.focus();
        }
      }
    };
    document.addEventListener("keydown", keydown, true);
    return () => {
      document.removeEventListener("keydown", keydown, true);
      previouslyFocused?.focus();
    };
  }, [open, onClose]);

  return (
    <>
      {open && (
        <div
          className="fixed inset-0 bg-black/75 lg:hidden"
          style={{ zIndex: Z_INDEX.LEFT_SIDEBAR }}
          onClick={onClose}
        />
      )}
      <aside
        ref={panel}
        aria-label="Card browser controls"
        role={open ? "dialog" : undefined}
        aria-modal={open ? true : undefined}
        className={`${open ? "flex" : "hidden lg:flex"} fixed inset-y-0 left-0 w-[304px] max-w-[calc(100vw-32px)] shrink-0 flex-col border-r border-space-blue-400 bg-space-black-darker text-left shadow-[12px_0_32px_rgba(10,25,65,0.45)] lg:static lg:max-w-none`}
        style={{ zIndex: Z_INDEX.LEFT_SIDEBAR }}
      >
        <div className="shrink-0 space-y-5 p-5 pb-4">
          <div className="flex items-center justify-between gap-2">
            <BackButton onClick={props.onBack}>{props.backLabel}</BackButton>
            <GameButton
              emphasis="quiet"
              data-drawer-close
              type="button"
              onClick={onClose}
              aria-label="Close card controls"
              className="cursor-pointer rounded-none p-2 text-white/70 hover:text-white focus-visible:outline-2 focus-visible:outline-blue-400 lg:hidden"
            >
              ✕
            </GameButton>
          </div>
          <h1 className="m-0 pb-3 font-orbitron text-xl font-semibold leading-relaxed text-white">
            Card browser
          </h1>
          <fieldset className="m-0 space-y-1 border-0 p-0">
            <legend className="sr-only">Card family</legend>
            {CARD_FAMILIES.map((family) => (
              <label
                key={family}
                className={`flex cursor-pointer items-center gap-2 rounded-none border px-3 py-2.5 text-sm transition-colors has-focus-visible:ring-2 has-focus-visible:ring-blue-400 ${props.family === family ? "border-space-blue-600 bg-space-blue-400 text-white" : "border-transparent text-white/60 hover:bg-white/5 hover:text-white"}`}
              >
                <input
                  className="sr-only"
                  type="radio"
                  name="card-family"
                  value={family}
                  checked={props.family === family}
                  onChange={() => {
                    void playButtonClickSound();
                    props.onFamily(family);
                  }}
                />
                <span className="flex-1 font-orbitron text-xs">{FAMILY_LABELS[family]}</span>
                {props.selectedCounts[family] > 0 && (
                  <span className="text-[10px] text-blue-200">
                    {props.selectedCounts[family]} selected
                  </span>
                )}
                <span className="font-orbitron text-xs tabular-nums">{props.counts[family]}</span>
              </label>
            ))}
          </fieldset>
          <label className="block space-y-2">
            <span className="font-orbitron text-xs text-white/70">Search cards</span>
            <input
              type="search"
              value={props.query}
              onChange={(event) => props.onQuery(event.target.value)}
              placeholder="ID, name, text, or tag"
              className={`${fieldClass} cursor-text`}
            />
          </label>
        </div>
        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto overscroll-contain border-y border-white/10 px-5 py-4">
          {props.shared && (
            <div className="space-y-2 text-sm text-blue-100">
              <p className="m-0">Viewing a shared selection</p>
              <GameButton size="sm" emphasis="secondary" onClick={props.onExitShared}>
                Exit shared selection
              </GameButton>
            </div>
          )}
          <label className="block space-y-2">
            <span className="font-orbitron text-xs text-white/70">Sort by</span>
            <select
              className={`${fieldClass} cursor-pointer`}
              value={
                props.family !== "project" && props.sort.startsWith("type") ? "id" : props.sort
              }
              onChange={(event) => {
                void playButtonClickSound();
                props.onSort(event.target.value as CardSort);
              }}
            >
              <option value="id">{props.shared ? "Link order" : "Card ID"}</option>
              <option value="name-asc">Name (A–Z)</option>
              <option value="name-desc">Name (Z–A)</option>
              {props.family === "project" && (
                <>
                  <option value="type-asc">Type (A–Z)</option>
                  <option value="type-desc">Type (Z–A)</option>
                </>
              )}
            </select>
          </label>
          {props.family === "project" && (
            <FilterGroup
              label="Project type"
              kind="types"
              values={["automated", "active", "event"]}
              selected={props.filters.types}
              onToggle={props.onToggle}
            />
          )}
          <FilterGroup
            label="Tags"
            kind="tags"
            values={props.tags}
            selected={props.filters.tags}
            onToggle={props.onToggle}
          />
          <FilterGroup
            label="Card packs"
            kind="packs"
            values={props.packs}
            selected={props.filters.packs}
            onToggle={props.onToggle}
          />
          <GameButton emphasis="secondary" size="sm" onClick={props.onClearFilters}>
            Clear filters
          </GameButton>
        </div>
        <div className="shrink-0 space-y-3 p-5">
          <p aria-live="polite" className="m-0 text-xs text-white/60">
            {props.counts[props.family]} results · {props.selectedCount} selected
          </p>
          {props.selectedCount > 0 && (
            <div className="space-y-2">
              <CopyLinkButton
                textToCopy={props.shareUrl}
                defaultText={`Copy link (${props.selectedCount})`}
                size="sm"
                className="w-full"
                onCopyError={() => setCopyError(true)}
                onCopySuccess={() => setCopyError(false)}
              />
              <GameButton emphasis="quiet" size="sm" onClick={props.onClearSelection}>
                Clear selection
              </GameButton>
              {copyError && (
                <p role="alert" className="text-xs text-error-red">
                  Could not copy the link. Please try again.
                </p>
              )}
            </div>
          )}
        </div>
      </aside>
    </>
  );
});

function FilterGroup({
  label,
  kind,
  values,
  selected,
  onToggle,
}: {
  label: string;
  kind: FilterKind;
  values: string[];
  selected: ReadonlySet<string>;
  onToggle: SidebarProps["onToggle"];
}) {
  const { playButtonClickSound } = useSoundEffects();
  return (
    <details className="group">
      <summary
        onClick={() => void playButtonClickSound()}
        className="cursor-pointer font-orbitron text-xs text-white/80 focus-visible:outline-blue-400"
      >
        {label}
        {selected.size > 0 ? ` (${selected.size})` : ""}
      </summary>
      <div className="mt-3 flex flex-wrap gap-2">
        {values.map((value) => (
          <label
            key={value}
            className={`flex cursor-pointer items-center gap-1.5 rounded-none border px-2 py-1.5 text-xs has-focus-visible:ring-2 has-focus-visible:ring-blue-400 ${selected.has(value) ? "border-blue-400/70 bg-blue-950 text-white" : "border-white/15 text-white/60 hover:border-white/40 hover:text-white"}`}
          >
            <input
              type="checkbox"
              className="sr-only"
              checked={selected.has(value)}
              onChange={() => {
                void playButtonClickSound();
                onToggle(kind, value);
              }}
            />
            {kind === "tags" && <GameIcon iconType={`${value}-tag`} size="small" />}
            {labelFor(value)}
          </label>
        ))}
      </div>
    </details>
  );
}
