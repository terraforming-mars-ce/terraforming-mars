import React, { useCallback, useRef } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import CardBrowser from "../ui/CardBrowser.tsx";

const CardsPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const filterCardIds = searchParams.getAll("cId");
  const initialSearchQuery = searchParams.get("q") || "";
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(null);

  const handleSearchQueryChange = useCallback(
    (query: string) => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
      debounceRef.current = setTimeout(() => {
        setSearchParams(
          (prev) => {
            if (query) {
              prev.set("q", query);
            } else {
              prev.delete("q");
            }
            return prev;
          },
          { replace: true },
        );
      }, 300);
    },
    [setSearchParams],
  );

  return (
    <CardBrowser
      onBack={() => navigate("/")}
      backLabel="Back to Home"
      filterCardIds={filterCardIds.length > 0 ? filterCardIds : undefined}
      initialSearchQuery={initialSearchQuery}
      onSearchQueryChange={handleSearchQueryChange}
    />
  );
};

export default CardsPage;
