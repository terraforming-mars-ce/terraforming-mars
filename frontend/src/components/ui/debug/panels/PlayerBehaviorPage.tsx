import React, { useState, useEffect, useMemo, useRef } from "react";
import { createPortal } from "react-dom";
import { globalWebSocketManager } from "../../../../services/globalWebSocketManager.ts";
import { apiService } from "../../../../services/apiService.ts";
import {
  GameDto,
  AdminCommandRequest,
  AdminCommandTypeGiveCard,
  GiveCardAdminCommand,
  CardDto,
  CardTypeCorporation,
} from "../../../../types/generated/api-types.ts";
import PlayerSelector from "../PlayerSelector.tsx";
import GameCard from "../../cards/GameCard.tsx";
import CorporationCard from "../../cards/CorporationCard.tsx";

interface PlayerBehaviorPageProps {
  gameState: GameDto;
  selectedPlayerIds: string[];
  onPlayerChange: (ids: string[]) => void;
}

const buttonStyle = {
  padding: "6px 14px",
  background: "linear-gradient(135deg, rgba(59, 130, 246, 0.8), rgba(59, 130, 246, 0.6))",
  border: "1px solid rgba(59, 130, 246, 0.5)",
  borderRadius: "6px",
  color: "white",
  fontSize: "11px",
  cursor: "pointer" as const,
  fontWeight: "500" as const,
  minWidth: "50px",
  textAlign: "center" as const,
};

const PlayerBehaviorPage: React.FC<PlayerBehaviorPageProps> = ({
  gameState,
  selectedPlayerIds,
  onPlayerChange,
}) => {
  const allPlayers = [gameState.currentPlayer, ...gameState.otherPlayers];
  const playerId = selectedPlayerIds[0];

  const [allCards, setAllCards] = useState<CardDto[]>([]);
  const [cardsLoading, setCardsLoading] = useState(false);

  const [cardId, setCardId] = useState("");
  const [cardQuery, setCardQuery] = useState("");
  const [showCardDropdown, setShowCardDropdown] = useState(false);

  const [corporationId, setCorporationId] = useState("");
  const [corpQuery, setCorpQuery] = useState("");
  const [showCorpDropdown, setShowCorpDropdown] = useState(false);

  const [hoveredCard, setHoveredCard] = useState<CardDto | null>(null);

  const cardInputRef = useRef<HTMLDivElement>(null);
  const corpInputRef = useRef<HTMLDivElement>(null);
  const cardDropdownRef = useRef<HTMLDivElement>(null);
  const corpDropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const loadCards = async () => {
      if (allCards.length > 0 || cardsLoading) {
        return;
      }
      setCardsLoading(true);
      try {
        const response = await apiService.listCards(0, 10000);
        setAllCards(response.cards);
      } catch (error) {
        console.error("Failed to load cards:", error);
      } finally {
        setCardsLoading(false);
      }
    };
    void loadCards();
  }, []);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as Node;
      if (cardInputRef.current && !cardInputRef.current.contains(target)) {
        setShowCardDropdown(false);
      }
      if (corpInputRef.current && !corpInputRef.current.contains(target)) {
        setShowCorpDropdown(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const parseSearchQuery = (rawQuery: string) => {
    const query = rawQuery.toLowerCase().trim();
    if (query.startsWith("tag:")) {
      return { type: "tag" as const, value: query.slice(4).trim() };
    }
    if (query.startsWith("behavior:") || query.startsWith("b:")) {
      const prefix = query.startsWith("behavior:") ? 9 : 2;
      return { type: "behavior" as const, value: query.slice(prefix).trim() };
    }
    if (query.startsWith("type:") || query.startsWith("t:")) {
      const prefix = query.startsWith("type:") ? 5 : 2;
      return { type: "cardType" as const, value: query.slice(prefix).trim() };
    }
    return { type: "text" as const, value: query };
  };

  const filterCardsByQuery = (cards: CardDto[], rawQuery: string) => {
    const parsed = parseSearchQuery(rawQuery);
    if (!parsed.value) {
      return [];
    }
    return cards.filter((card) => {
      switch (parsed.type) {
        case "tag":
          return card.tags?.some((tag) => tag.toLowerCase().includes(parsed.value));
        case "behavior":
          return card.behaviors?.some((behavior) => {
            const inputMatch = behavior.inputs?.some((input) =>
              input.type.toLowerCase().includes(parsed.value),
            );
            const outputMatch = behavior.outputs?.some((output) =>
              output.type.toLowerCase().includes(parsed.value),
            );
            return inputMatch || outputMatch;
          });
        case "cardType":
          return card.type.toLowerCase().includes(parsed.value);
        case "text":
        default:
          return (
            card.id.toLowerCase().includes(parsed.value) ||
            card.name.toLowerCase().includes(parsed.value)
          );
      }
    });
  };

  const filteredCards = useMemo(() => {
    if (!cardQuery.trim()) {
      return [];
    }
    return filterCardsByQuery(allCards, cardQuery).slice(0, 10);
  }, [allCards, cardQuery]);

  const filteredCorporations = useMemo(() => {
    const corps = allCards.filter((card) => card.type === CardTypeCorporation);
    if (!corpQuery.trim()) {
      return [];
    }
    const q = corpQuery.toLowerCase();
    return corps
      .filter((card) => card.id.toLowerCase().includes(q) || card.name.toLowerCase().includes(q))
      .slice(0, 10);
  }, [allCards, corpQuery]);

  const sendCommand = async (commandType: string, payload: any) => {
    const req: AdminCommandRequest = { commandType: commandType as any, payload };
    try {
      await globalWebSocketManager.sendAdminCommand(req);
    } catch (error) {
      console.error("Failed to send admin command:", error);
    }
  };

  const handleGiveCard = async () => {
    if (selectedPlayerIds.length === 0 || !cardId) {
      return;
    }
    for (const pid of selectedPlayerIds) {
      const command: GiveCardAdminCommand = { playerId: pid, cardId };
      await sendCommand(AdminCommandTypeGiveCard, command);
    }
    setCardId("");
    setCardQuery("");
  };

  const handleSetCorporation = async () => {
    if (!playerId || !corporationId) {
      return;
    }
    await sendCommand("set-corporation" as any, { playerId, corporationId });
    setCorporationId("");
    setCorpQuery("");
  };

  const players = allPlayers.map((p) => ({ id: p.id, name: p.name }));

  const getDropdownPosition = (ref: React.RefObject<HTMLDivElement | null>) => {
    if (!ref.current) return { top: 0, left: 0, width: 0 };
    const rect = ref.current.getBoundingClientRect();
    return { top: rect.bottom + 2, left: rect.left, width: rect.width };
  };

  const dropdownBaseStyle = {
    position: "fixed" as const,
    background: "rgba(0, 0, 0, 0.98)",
    border: "1px solid rgba(59, 130, 246, 0.5)",
    borderRadius: "4px",
    zIndex: 99999,
    boxShadow: "0 4px 12px rgba(0, 0, 0, 0.8)",
    overflow: "hidden",
  };

  const dropdownItemStyle = {
    padding: "6px 12px",
    cursor: "pointer" as const,
    fontSize: "12px",
    borderBottom: "1px solid rgba(59, 130, 246, 0.2)",
    transition: "background 0.15s ease",
    textAlign: "left" as const,
  };

  return (
    <div>
      <PlayerSelector players={players} selectedIds={selectedPlayerIds} onChange={onPlayerChange} />

      <div style={{ marginTop: "12px" }}>
        <label
          style={{
            color: "#3b82f6",
            fontSize: "11px",
            fontWeight: "bold",
            display: "block",
            textAlign: "left",
          }}
        >
          Card
        </label>
        <div style={{ display: "flex", gap: "6px", alignItems: "flex-start" }}>
          <div ref={cardInputRef} style={{ flex: 1 }}>
            <input
              type="text"
              placeholder="Search: name, tag:space, b:discount..."
              spellCheck={false}
              autoComplete="off"
              value={cardQuery}
              onChange={(e) => {
                setCardQuery(e.target.value);
                setShowCardDropdown(true);
                if (cardId) {
                  setCardId("");
                }
              }}
              onFocus={() => setShowCardDropdown(true)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  if (filteredCards.length === 1) {
                    setCardId(filteredCards[0].id);
                    setCardQuery(`${filteredCards[0].id} - ${filteredCards[0].name}`);
                    setShowCardDropdown(false);
                  } else if (cardId) {
                    void handleGiveCard();
                  }
                }
              }}
              style={{
                width: "100%",
                padding: "6px 10px",
                background: "rgba(0, 0, 0, 0.8)",
                border: "1px solid rgba(59, 130, 246, 0.3)",
                borderRadius: "4px",
                color: "white",
                fontSize: "12px",
                outline: "none",
                boxSizing: "border-box",
              }}
            />
            {showCardDropdown &&
              cardQuery.trim() &&
              (() => {
                const pos = getDropdownPosition(cardInputRef);
                return createPortal(
                  <div
                    ref={cardDropdownRef}
                    style={{ ...dropdownBaseStyle, top: pos.top, left: pos.left, width: pos.width }}
                    onMouseLeave={() => setHoveredCard(null)}
                  >
                    {filteredCards.length === 0 ? (
                      <div style={{ padding: "8px 12px", color: "#666", fontSize: "12px" }}>
                        No results
                      </div>
                    ) : (
                      filteredCards.map((card) => (
                        <div
                          key={card.id}
                          onClick={() => {
                            setCardId(card.id);
                            setCardQuery(`${card.id} - ${card.name}`);
                            setShowCardDropdown(false);
                            setHoveredCard(null);
                          }}
                          style={dropdownItemStyle}
                          onMouseEnter={(e) => {
                            e.currentTarget.style.background = "rgba(59, 130, 246, 0.2)";
                            setHoveredCard(card);
                          }}
                          onMouseLeave={(e) => {
                            e.currentTarget.style.background = "transparent";
                          }}
                        >
                          <span style={{ color: "#3b82f6", fontWeight: "bold" }}>{card.id}</span>
                          <span style={{ color: "#abb2bf" }}> - {card.name}</span>
                        </div>
                      ))
                    )}
                  </div>,
                  document.body,
                );
              })()}
          </div>
          <button onClick={() => void handleGiveCard()} style={buttonStyle}>
            Give
          </button>
        </div>
      </div>

      <div style={{ marginTop: "12px" }}>
        <label
          style={{
            color: "#3b82f6",
            fontSize: "11px",
            fontWeight: "bold",
            display: "block",
            textAlign: "left",
          }}
        >
          Corporation
        </label>
        <div style={{ display: "flex", gap: "6px", alignItems: "flex-start" }}>
          <div ref={corpInputRef} style={{ flex: 1 }}>
            <input
              type="text"
              placeholder="Search corporation..."
              spellCheck={false}
              autoComplete="off"
              value={corpQuery}
              onChange={(e) => {
                setCorpQuery(e.target.value);
                setShowCorpDropdown(true);
                if (corporationId) {
                  setCorporationId("");
                }
              }}
              onFocus={() => setShowCorpDropdown(true)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  if (filteredCorporations.length === 1) {
                    setCorporationId(filteredCorporations[0].id);
                    setCorpQuery(`${filteredCorporations[0].id} - ${filteredCorporations[0].name}`);
                    setShowCorpDropdown(false);
                  } else if (corporationId) {
                    void handleSetCorporation();
                  }
                }
              }}
              style={{
                width: "100%",
                padding: "6px 10px",
                background: "rgba(0, 0, 0, 0.8)",
                border: "1px solid rgba(59, 130, 246, 0.3)",
                borderRadius: "4px",
                color: "white",
                fontSize: "12px",
                outline: "none",
                boxSizing: "border-box",
              }}
            />
            {showCorpDropdown &&
              corpQuery.trim() &&
              (() => {
                const pos = getDropdownPosition(corpInputRef);
                return createPortal(
                  <div
                    ref={corpDropdownRef}
                    style={{ ...dropdownBaseStyle, top: pos.top, left: pos.left, width: pos.width }}
                    onMouseLeave={() => setHoveredCard(null)}
                  >
                    {filteredCorporations.length === 0 ? (
                      <div style={{ padding: "8px 12px", color: "#666", fontSize: "12px" }}>
                        No results
                      </div>
                    ) : (
                      filteredCorporations.map((card) => (
                        <div
                          key={card.id}
                          onClick={() => {
                            setCorporationId(card.id);
                            setCorpQuery(`${card.id} - ${card.name}`);
                            setShowCorpDropdown(false);
                            setHoveredCard(null);
                          }}
                          style={dropdownItemStyle}
                          onMouseEnter={(e) => {
                            e.currentTarget.style.background = "rgba(59, 130, 246, 0.2)";
                            setHoveredCard(card);
                          }}
                          onMouseLeave={(e) => {
                            e.currentTarget.style.background = "transparent";
                          }}
                        >
                          <span style={{ color: "#ffc107", fontWeight: "bold" }}>{card.id}</span>
                          <span style={{ color: "#abb2bf" }}> - {card.name}</span>
                        </div>
                      ))
                    )}
                  </div>,
                  document.body,
                );
              })()}
          </div>
          <button onClick={() => void handleSetCorporation()} style={buttonStyle}>
            Set
          </button>
        </div>
      </div>
      {hoveredCard &&
        (() => {
          const activeDropdownRef = showCardDropdown ? cardDropdownRef : corpDropdownRef;
          const dropdownEl = activeDropdownRef.current;
          if (!dropdownEl) {
            return null;
          }
          const dropdownRect = dropdownEl.getBoundingClientRect();
          const isCorporation = hoveredCard.type === CardTypeCorporation;
          const corpScale = 0.85;
          const cardWidth = isCorporation ? 400 * corpScale : 200;
          const cardHeight = isCorporation ? 380 * corpScale : 280;
          const gap = 8;
          const previewLeft = dropdownRect.left - cardWidth - gap;
          const dropdownCenterY = dropdownRect.top + dropdownRect.height / 2;
          const previewTop = dropdownCenterY - cardHeight / 2;

          return createPortal(
            <div
              style={{
                position: "fixed",
                left: previewLeft,
                top: previewTop,
                zIndex: 99999,
                pointerEvents: "none",
              }}
            >
              {isCorporation ? (
                <div style={{ transform: `scale(${corpScale})`, transformOrigin: "top left" }}>
                  <CorporationCard
                    card={hoveredCard}
                    isSelected={false}
                    onSelect={() => {}}
                    showCheckbox={false}
                    disableInteraction={true}
                  />
                </div>
              ) : (
                <GameCard
                  card={hoveredCard}
                  isSelected={false}
                  onSelect={() => {}}
                  showCheckbox={false}
                />
              )}
            </div>,
            document.body,
          );
        })()}
    </div>
  );
};

export default PlayerBehaviorPage;
