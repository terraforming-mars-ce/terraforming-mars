import { useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { Z_INDEX } from "@/constants/zIndex.ts";
import GameButton from "./buttons/GameButton.tsx";
import CardBrowserSidebar from "./cardBrowser/CardBrowserSidebar.tsx";
import type { BrowserFilters, FilterKind } from "./cardBrowser/CardBrowserSidebar.tsx";
import CachedCardFamily from "./cardBrowser/VirtualCardGrid.tsx";
import {
  CARD_FAMILIES,
  EMPTY_VIEW,
  filterCatalog,
  loadCardCatalog,
  useCardCatalog,
} from "./cardBrowser/cardCatalog.ts";
import type { CardBrowserView, CardFamily, CardSort } from "./cardBrowser/cardCatalog.ts";

interface CardBrowserProps {
  onBack: () => void;
  backLabel?: string;
  view?: CardBrowserView;
  onViewChange?: (patch: Partial<CardBrowserView>) => void;
}

const emptyFilters = (): BrowserFilters => ({
  tags: new Set(),
  types: new Set(),
  packs: new Set(),
});

export default function CardBrowser({
  onBack,
  backLabel = "Back to Home",
  view: controlledView,
  onViewChange,
}: CardBrowserProps) {
  const [localView, setLocalView] = useState(EMPTY_VIEW);
  const view = controlledView ?? localView;
  const updateView = useCallback(
    (patch: Partial<CardBrowserView>) => {
      if (onViewChange) {
        onViewChange(patch);
      } else {
        setLocalView((previous) => ({ ...previous, ...patch }));
      }
    },
    [onViewChange],
  );
  const { entries, status, error } = useCardCatalog();
  const [filters, setFilters] = useState(emptyFilters);
  const [sort, setSort] = useState<CardSort>("id");
  const [selected, setSelected] = useState<Set<string>>(() => new Set());
  const [visitedFamilies, setVisitedFamilies] = useState<ReadonlySet<CardFamily>>(() => new Set());
  const [drawerOpen, setDrawerOpen] = useState(false);
  const query = useDeferredValue(view.query);
  const closeDrawer = useCallback(() => setDrawerOpen(false), []);
  useEffect(() => {
    void loadCardCatalog();
  }, []);
  useEffect(() => {
    const breakpoint = window.matchMedia("(min-width: 1024px)");
    const closeOnDesktop = () => {
      if (breakpoint.matches) {
        closeDrawer();
      }
    };
    breakpoint.addEventListener("change", closeOnDesktop);
    return () => breakpoint.removeEventListener("change", closeOnDesktop);
  }, [closeDrawer]);

  const baseFamily = useMemo(() => {
    if (!view.ids.length) {
      return "project";
    }
    const ids = new Set(view.ids);
    return (
      CARD_FAMILIES.find((family) =>
        entries.some((entry) => entry.family === family && ids.has(entry.card.id)),
      ) ?? "project"
    );
  }, [entries, view.ids]);
  const family = view.family ?? baseFamily;
  useEffect(() => {
    setVisitedFamilies((previous) => {
      if (previous.has(family)) {
        return previous;
      }
      return new Set([...previous, family]);
    });
  }, [family]);
  const groups = useMemo(
    () => filterCatalog(entries, { query, ids: view.ids, ...filters, sort }),
    [entries, query, view.ids, filters, sort],
  );
  const counts = useMemo(
    () => ({
      project: groups.project.length,
      prelude: groups.prelude.length,
      corporation: groups.corporation.length,
    }),
    [groups],
  );
  const { tags, packs } = useMemo(
    () => ({
      tags: [...new Set(entries.flatMap(({ card }) => card.tags ?? []))].sort(),
      packs: [...new Set(entries.map(({ card }) => card.pack))].sort(),
    }),
    [entries],
  );
  const selectedCounts = useMemo(() => {
    const counts = { project: 0, prelude: 0, corporation: 0 };
    for (const entry of entries) {
      if (selected.has(entry.card.id)) {
        counts[entry.family]++;
      }
    }
    return counts;
  }, [entries, selected]);
  const shareUrl = useMemo(() => {
    const params = new URLSearchParams();
    const sharedFamily =
      selectedCounts[family] > 0
        ? family
        : (CARD_FAMILIES.find((item) => selectedCounts[item] > 0) ?? "project");
    params.set("family", sharedFamily);
    [...selected].sort().forEach((id) => params.append("cId", id));
    return `${window.location.origin}/cards?${params}`;
  }, [selected, selectedCounts, family]);
  const toggleSelection = useCallback((id: string) => {
    setSelected((previous) => {
      const next = new Set(previous);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);
  const toggleFilter = useCallback((kind: FilterKind, value: string) => {
    setFilters((previous) => {
      const next = new Set(previous[kind]);
      if (next.has(value)) {
        next.delete(value);
      } else {
        next.add(value);
      }
      return { ...previous, [kind]: next };
    });
  }, []);
  const clearFilters = () => {
    setFilters(emptyFilters());
    updateView({ query: "" });
  };

  return (
    <div
      className="relative flex h-dvh min-h-0 w-full overflow-hidden bg-black text-white"
      style={{ zIndex: Z_INDEX.UI_BASE }}
      onKeyDown={(event) => {
        if (event.key !== "Escape") {
          event.stopPropagation();
        }
      }}
    >
      <CardBrowserSidebar
        family={family}
        query={view.query}
        sort={sort}
        filters={filters}
        tags={tags}
        packs={packs}
        counts={counts}
        selectedCounts={selectedCounts}
        selectedCount={selected.size}
        shareUrl={shareUrl}
        shared={view.ids.length > 0}
        open={drawerOpen}
        backLabel={backLabel}
        onBack={onBack}
        onClose={closeDrawer}
        onFamily={(family: CardFamily) => updateView({ family })}
        onQuery={(query) => updateView({ query })}
        onSort={setSort}
        onToggle={toggleFilter}
        onClearFilters={clearFilters}
        onClearSelection={() => setSelected(new Set())}
        onExitShared={() => updateView({ ids: [] })}
      />
      <main className="relative min-h-0 min-w-0 flex-1 pt-16 lg:pt-0" inert={drawerOpen}>
        <div className="absolute left-4 top-3 lg:hidden">
          <GameButton size="sm" emphasis="secondary" onClick={() => setDrawerOpen(true)}>
            Browse cards
          </GameButton>
        </div>
        {(status === "idle" || status === "loading") && (
          <div role="status" className="p-8 text-left text-white/60">
            Loading cards…
          </div>
        )}
        {status === "error" && (
          <div role="alert" className="space-y-4 p-8 text-left">
            <p>{error}</p>
            <GameButton emphasis="secondary" onClick={() => void loadCardCatalog()}>
              Retry
            </GameButton>
          </div>
        )}
        {status === "ready" && groups[family].length === 0 && (
          <div className="space-y-4 p-8 text-left">
            <h2 className="font-orbitron text-lg">No matching cards</h2>
            <p className="text-white/60">Try another family or clear your search and filters.</p>
            <GameButton emphasis="secondary" onClick={clearFilters}>
              Clear filters
            </GameButton>
            {view.ids.length > 0 && (
              <GameButton emphasis="quiet" onClick={() => updateView({ ids: [] })}>
                Exit shared selection
              </GameButton>
            )}
          </div>
        )}
        {status === "ready" && (
          <div className="relative h-full min-h-0" hidden={groups[family].length === 0}>
            {CARD_FAMILIES.filter((item) => item === family || visitedFamilies.has(item)).map(
              (item) => (
                <CachedCardFamily
                  key={item}
                  active={item === family}
                  cards={groups[item]}
                  family={item}
                  selected={selected}
                  onSelect={toggleSelection}
                />
              ),
            )}
          </div>
        )}
      </main>
    </div>
  );
}
