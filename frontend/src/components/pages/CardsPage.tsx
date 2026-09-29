import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import CardBrowser from "../ui/CardBrowser.tsx";
import { parseFamily } from "../ui/cardBrowser/cardCatalog.ts";
import type { CardBrowserView } from "../ui/cardBrowser/cardCatalog.ts";

export default function CardsPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const [query, setQuery] = useState(() => searchParams.get("q") ?? "");
  const debounce = useRef<ReturnType<typeof setTimeout> | null>(null);
  const writtenSearch = useRef<string | null>(null);
  const idsKey = JSON.stringify(searchParams.getAll("cId"));
  const ids = useMemo<string[]>(() => JSON.parse(idsKey), [idsKey]);
  const family = parseFamily(searchParams.get("family"));
  const view = useMemo(() => ({ query, family, ids }), [query, family, ids]);

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    if (writtenSearch.current === params.toString()) {
      writtenSearch.current = null;
      return;
    }
    if (debounce.current) {
      clearTimeout(debounce.current);
    }
    setQuery(params.get("q") ?? "");
  }, [location.key, location.search]);

  useEffect(
    () => () => {
      if (debounce.current) {
        clearTimeout(debounce.current);
      }
    },
    [],
  );

  const handleViewChange = useCallback(
    (patch: Partial<CardBrowserView>) => {
      if (debounce.current) {
        clearTimeout(debounce.current);
      }
      const nextQuery = patch.query ?? query;
      setQuery(nextQuery);
      const write = () => {
        setSearchParams(
          (previous) => {
            const next = new URLSearchParams(previous);
            if (nextQuery) {
              next.set("q", nextQuery);
            } else {
              next.delete("q");
            }
            if (patch.family) {
              next.set("family", patch.family);
            }
            if (patch.ids) {
              next.delete("cId");
              patch.ids.forEach((id) => next.append("cId", id));
            }
            writtenSearch.current = next.toString();
            return next;
          },
          { replace: true },
        );
      };
      if (patch.query !== undefined && patch.family === undefined && patch.ids === undefined) {
        debounce.current = setTimeout(write, 300);
      } else {
        write();
      }
    },
    [query, setSearchParams],
  );

  return <CardBrowser onBack={() => navigate("/")} view={view} onViewChange={handleViewChange} />;
}
