import { create } from "zustand";
import { apiService } from "@/services/apiService.ts";
import type { CardDto } from "@/types/generated/api-types.ts";

export const CARD_FAMILIES = ["project", "prelude", "corporation"] as const;
export type CardFamily = (typeof CARD_FAMILIES)[number];
export type CardSort = "id" | "name-asc" | "name-desc" | "type-asc" | "type-desc";
export interface CardBrowserView {
  query: string;
  family?: CardFamily;
  ids: readonly string[];
}
export const EMPTY_VIEW: CardBrowserView = { query: "", ids: [] };
export const FAMILY_LABELS: Record<CardFamily, string> = {
  project: "Project cards",
  prelude: "Preludes",
  corporation: "Corporations",
};

export function cardFamily(card: CardDto): CardFamily {
  if (card.type === "corporation" || card.type === "prelude") {
    return card.type;
  }
  return "project";
}

export function parseFamily(value: string | null): CardFamily | undefined {
  return CARD_FAMILIES.find((family) => family === value);
}

interface CatalogEntry {
  card: CardDto;
  family: CardFamily;
  search: string;
}
interface CatalogState {
  entries: CatalogEntry[];
  status: "idle" | "loading" | "ready" | "error";
  error: string | null;
}

export const useCardCatalog = create<CatalogState>(() => ({
  entries: [],
  status: "idle",
  error: null,
}));

let pending: Promise<void> | null = null;
export function loadCardCatalog(): Promise<void> {
  if (pending) {
    return pending;
  }
  if (useCardCatalog.getState().status === "ready") {
    return Promise.resolve();
  }
  useCardCatalog.setState({ status: "loading", error: null });
  pending = (async () => {
    try {
      const cards: CardDto[] = [];
      let total = Infinity;
      while (cards.length < total) {
        const response = await apiService.listCards(cards.length, 1000);
        total = response.totalCount;
        if (response.cards.length === 0 && cards.length < total) {
          throw new Error("The card catalog was incomplete. Please retry.");
        }
        cards.push(...response.cards);
      }
      useCardCatalog.setState({
        status: "ready",
        entries: cards.map((card) => ({
          card,
          family: cardFamily(card),
          search: [card.id, card.name, card.description, ...(card.tags ?? [])]
            .join("\n")
            .toLowerCase(),
        })),
      });
    } catch (error) {
      useCardCatalog.setState({
        status: "error",
        error: error instanceof Error ? error.message : "Unable to load cards.",
      });
    } finally {
      pending = null;
    }
  })();
  return pending;
}

export interface CatalogFilters {
  query: string;
  ids: readonly string[];
  tags: ReadonlySet<string>;
  types: ReadonlySet<string>;
  packs: ReadonlySet<string>;
  sort: CardSort;
}

export function filterCatalog(entries: CatalogEntry[], filters: CatalogFilters) {
  const query = filters.query.trim().toLowerCase();
  const ids = new Set(filters.ids);
  const idOrder = new Map([...ids].map((id, index) => [id, index]));
  const groups: Record<CardFamily, CardDto[]> = { project: [], prelude: [], corporation: [] };
  for (const entry of entries) {
    const { card, family } = entry;
    if (ids.size && !ids.has(card.id)) {
      continue;
    }
    if (query && !entry.search.includes(query)) {
      continue;
    }
    if (filters.tags.size && !card.tags?.some((tag) => filters.tags.has(tag))) {
      continue;
    }
    if (filters.packs.size && !filters.packs.has(card.pack)) {
      continue;
    }
    if (family === "project" && filters.types.size && !filters.types.has(card.type)) {
      continue;
    }
    groups[family].push(card);
  }
  for (const family of CARD_FAMILIES) {
    groups[family].sort((a, b) => {
      let order = 0;
      if (filters.sort.startsWith("name")) {
        order = a.name.localeCompare(b.name);
      } else if (family === "project" && filters.sort.startsWith("type")) {
        order = a.type.localeCompare(b.type);
      }
      if (filters.sort.endsWith("desc")) {
        order *= -1;
      }
      return (
        order || (ids.size ? idOrder.get(a.id)! - idOrder.get(b.id)! : a.id.localeCompare(b.id))
      );
    });
  }
  return groups;
}

export function labelFor(value: string) {
  return value.replace(
    /(^|-)(\w)/g,
    (_, separator: string, letter: string) => `${separator ? " " : ""}${letter.toUpperCase()}`,
  );
}
