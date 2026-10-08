import { useState } from "react";
import type { PlayerCardDto } from "@/types/generated/api-types.ts";
import { useMobileUiStore } from "@/stores/mobileUiStore.ts";
import { useSpectateStore } from "@/stores/spectateStore.ts";
import MobileScreen, { type MobileScreenTab } from "../MobileScreen.tsx";
import { HandIcon } from "../screenIcons.tsx";
import { useMobileGame } from "../MobileGameContext.tsx";
import HandRow from "../hand/HandRow.tsx";
import HandCardDetail from "../hand/HandCardDetail.tsx";
import { useStableHandOrder } from "../hand/useStableHandOrder.ts";
import { useNewCardHighlight } from "../hand/useNewCardHighlight.ts";

type HandTab = "all" | "playable";

export default function MobileHandScreen() {
  const close = useMobileUiStore((s) => s.close);
  const spectatePlayerId = useSpectateStore((s) => s.spectatePlayerId);
  const { currentPlayer, isSpectator, isReplay, onPlayCard } = useMobileGame();
  const [tab, setTab] = useState<HandTab>("all");
  const [detailCardId, setDetailCardId] = useState<string | null>(null);

  const orderedCards = useStableHandOrder(currentPlayer?.cards ?? []);
  const highlightIds = useNewCardHighlight(orderedCards.map((card) => card.id));
  const playableCards = orderedCards.filter((card) => card.available);
  const visibleCards = tab === "playable" ? playableCards : orderedCards;
  const emptyMessage = tab === "playable" ? "No playable cards" : "No cards in hand";
  const showPlay = !isSpectator && !isReplay && !spectatePlayerId;

  const tabs: MobileScreenTab[] = [
    { id: "all", label: "All", badge: orderedCards.length },
    { id: "playable", label: "Playable", badge: playableCards.length },
  ];

  const detailCard = orderedCards.find((card) => card.id === detailCardId) ?? null;
  if (detailCardId !== null && !detailCard) {
    setDetailCardId(null);
  }
  const navigationCards = visibleCards.some((card) => card.id === detailCardId)
    ? visibleCards
    : orderedCards;
  const detailIndex = navigationCards.findIndex((card) => card.id === detailCardId);
  const previousCard = detailIndex > 0 ? navigationCards[detailIndex - 1] : null;
  const nextCard =
    detailIndex >= 0 && detailIndex < navigationCards.length - 1
      ? navigationCards[detailIndex + 1]
      : null;

  const handlePlay = (card: PlayerCardDto) => {
    setDetailCardId(null);
    close();
    onPlayCard(card.id);
  };

  return (
    <>
      <MobileScreen
        icon={HandIcon}
        title="Hand"
        tabs={tabs}
        activeTab={tab}
        onTabChange={(id) => setTab(id === "playable" ? "playable" : "all")}
        onClose={close}
      >
        <HandRow
          cards={visibleCards}
          emptyMessage={emptyMessage}
          highlightIds={highlightIds}
          focusCardId={detailCardId}
          onSelect={(card) => setDetailCardId(card.id)}
        />
      </MobileScreen>
      {detailCard && (
        <HandCardDetail
          card={detailCard}
          showPlay={showPlay}
          onPlay={handlePlay}
          onClose={() => setDetailCardId(null)}
          onPrevious={previousCard ? () => setDetailCardId(previousCard.id) : undefined}
          onNext={nextCard ? () => setDetailCardId(nextCard.id) : undefined}
        />
      )}
    </>
  );
}
