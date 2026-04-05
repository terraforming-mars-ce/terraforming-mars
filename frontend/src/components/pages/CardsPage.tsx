import React from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import CardBrowser from "../ui/CardBrowser.tsx";

const CardsPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const filterCardIds = searchParams.getAll("cId");

  return (
    <CardBrowser
      onBack={() => navigate("/")}
      backLabel="Back to Home"
      filterCardIds={filterCardIds.length > 0 ? filterCardIds : undefined}
    />
  );
};

export default CardsPage;
