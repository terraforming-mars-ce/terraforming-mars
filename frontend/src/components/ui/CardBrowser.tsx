import React, { useEffect, useState, useCallback, useMemo, useRef } from "react";
import { apiService } from "../../services/apiService";
import GameCard from "./cards/GameCard";
import CorporationCard from "./cards/CorporationCard";
import CopyLinkButton from "./buttons/CopyLinkButton";
import GameButton from "./buttons/GameButton.tsx";
import { CardDto, CardTypeCorporation } from "@/types/generated/api-types";
import GameIcon from "./display/GameIcon.tsx";
import BackButton from "./buttons/BackButton.tsx";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import { getCardTypeColor } from "@/utils/cardTypeColors.ts";

interface CardBrowserProps {
  onBack: () => void;
  backLabel?: string;
  title?: string;
  scrollContainerRef?: React.RefObject<HTMLDivElement | null>;
  filterCardIds?: string[];
}

const CardBrowser: React.FC<CardBrowserProps> = ({
  onBack,
  backLabel = "Back to Home",
  title = "Terraforming Mars Cards",
  scrollContainerRef,
  filterCardIds,
}) => {
  const isOverlay = !!scrollContainerRef;

  const [allCards, setAllCards] = useState<CardDto[]>([]);
  const [selectedCards, setSelectedCards] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [isScrolled, setIsScrolled] = useState(false);
  const [visibleRange, setVisibleRange] = useState({ start: 0, end: 100 });
  const [sortBy, setSortBy] = useState<
    "unsorted" | "name-asc" | "name-desc" | "type-asc" | "type-desc"
  >("unsorted");
  const [selectedTags, setSelectedTags] = useState<Set<string>>(new Set());
  const [selectedTypes, setSelectedTypes] = useState<Set<string>>(new Set());
  const [selectedPacks, setSelectedPacks] = useState<Set<string>>(new Set());
  const [showFilters, setShowFilters] = useState(false);
  const [showSortDropdown, setShowSortDropdown] = useState(false);
  const [isFadedIn, setIsFadedIn] = useState(false);
  const [windowWidth, setWindowWidth] = useState(
    typeof window !== "undefined" ? window.innerWidth : 1300,
  );

  const loadingRef = useRef(false);

  const toggleSetItem = useCallback(
    (setter: React.Dispatch<React.SetStateAction<Set<string>>>, item: string) => {
      setter((prev) => {
        const next = new Set(prev);
        if (next.has(item)) {
          next.delete(item);
        } else {
          next.add(item);
        }
        return next;
      });
    },
    [],
  );

  const getScrollValues = useCallback(() => {
    if (scrollContainerRef?.current) {
      const el = scrollContainerRef.current;
      return {
        scrollTop: el.scrollTop,
        viewportHeight: el.clientHeight,
        width: window.innerWidth,
      };
    }
    return {
      scrollTop: window.scrollY,
      viewportHeight: window.innerHeight,
      width: window.innerWidth,
    };
  }, [scrollContainerRef]);

  const loadAllCards = useCallback(async () => {
    if (loadingRef.current) {
      return;
    }

    loadingRef.current = true;
    setLoading(true);
    setError(null);

    try {
      const response = await apiService.listCards(0, 10000);
      setAllCards(response.cards);
    } catch (err: any) {
      setError(err.message || "Failed to load cards");
    } finally {
      setLoading(false);
      loadingRef.current = false;
    }
  }, []);

  const availableTags = useMemo(() => {
    const tags = new Set<string>();
    allCards.forEach((card) => {
      card.tags?.forEach((tag) => tags.add(tag));
    });
    return Array.from(tags).sort();
  }, [allCards]);

  const availableTypes = useMemo(() => {
    const types = new Set<string>();
    allCards.forEach((card) => {
      if (card.type) {
        types.add(card.type);
      }
    });
    return Array.from(types).sort();
  }, [allCards]);

  const availablePacks = useMemo(() => {
    const packs = new Set<string>();
    allCards.forEach((card) => {
      if (card.pack) {
        packs.add(card.pack);
      }
    });
    return Array.from(packs).sort();
  }, [allCards]);

  const formatPackName = (pack: string) => {
    return pack
      .split("-")
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" ");
  };

  const typePriority = (type: string) => {
    if (type === "corporation") {
      return 2;
    }
    if (type === "prelude") {
      return 1;
    }
    return 0;
  };

  const cards = useMemo(() => {
    let filtered = allCards;

    if (filterCardIds && filterCardIds.length > 0) {
      const idSet = new Set(filterCardIds);
      filtered = filtered.filter((card) => idSet.has(card.id));
    }

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (card) =>
          card.id.toLowerCase().includes(query) ||
          card.name.toLowerCase().includes(query) ||
          card.description.toLowerCase().includes(query) ||
          card.tags?.some((tag) => tag.toLowerCase().includes(query)),
      );
    }

    if (selectedTags.size > 0) {
      filtered = filtered.filter((card) => card.tags?.some((tag) => selectedTags.has(tag)));
    }

    if (selectedTypes.size > 0) {
      filtered = filtered.filter((card) => selectedTypes.has(card.type));
    }

    if (selectedPacks.size > 0) {
      filtered = filtered.filter((card) => selectedPacks.has(card.pack));
    }

    const sorted = [...filtered];
    if (sortBy === "unsorted") {
      sorted.sort((a, b) => {
        const priorityDiff = typePriority(a.type) - typePriority(b.type);
        if (priorityDiff !== 0) {
          return priorityDiff;
        }
        return a.id.localeCompare(b.id);
      });
    } else {
      sorted.sort((a, b) => {
        switch (sortBy) {
          case "name-asc":
            return a.name.localeCompare(b.name);
          case "name-desc":
            return b.name.localeCompare(a.name);
          case "type-asc":
            return a.type.localeCompare(b.type);
          case "type-desc":
            return b.type.localeCompare(a.type);
          default:
            return 0;
        }
      });
    }

    return sorted;
  }, [allCards, searchQuery, selectedTags, selectedTypes, selectedPacks, sortBy, filterCardIds]);

  const handleSortSelection = useCallback(
    (selectedSort: typeof sortBy) => {
      if (selectedSort === sortBy) {
        setSortBy("unsorted");
      } else {
        setSortBy(selectedSort);
      }
      setShowSortDropdown(false);
    },
    [sortBy],
  );

  const getSortDisplayText = useCallback(() => {
    switch (sortBy) {
      case "unsorted":
        return "No Sorting";
      case "name-asc":
        return "Name (A-Z)";
      case "name-desc":
        return "Name (Z-A)";
      case "type-asc":
        return "Type (A-Z)";
      case "type-desc":
        return "Type (Z-A)";
      default:
        return "No Sorting";
    }
  }, [sortBy]);

  const REGULAR_CARD_HEIGHT = 360;
  const CORP_CARD_HEIGHT = 480;
  const ROW_BUFFER = 10;

  const calculateVisibleRange = useCallback(
    (positions: Array<{ cardIndex: number; top: number; height: number }>) => {
      if (positions.length === 0) {
        return { start: 0, end: 100 };
      }

      const { scrollTop, viewportHeight, width } = getScrollValues();

      let headerOffset = 120;
      if (width <= 480) {
        headerOffset = showFilters ? 470 : 180;
      } else if (width <= 768) {
        headerOffset = showFilters ? 440 : 160;
      } else {
        headerOffset = showFilters ? 320 : 120;
      }

      const adjustedScrollTop = Math.max(0, scrollTop - headerOffset);
      const viewportBottom = adjustedScrollTop + viewportHeight;

      const bufferDistance = REGULAR_CARD_HEIGHT * ROW_BUFFER;

      let startIndex = 0;
      let endIndex = positions.length;

      for (let i = 0; i < positions.length; i++) {
        const cardBottom = positions[i].top + positions[i].height;
        if (cardBottom >= adjustedScrollTop - bufferDistance) {
          startIndex = i;
          break;
        }
      }

      for (let i = startIndex; i < positions.length; i++) {
        const cardTop = positions[i].top;
        if (cardTop > viewportBottom + bufferDistance) {
          endIndex = i;
          break;
        }
      }

      return { start: startIndex, end: endIndex };
    },
    [showFilters, getScrollValues],
  );

  const handleScroll = useCallback(() => {
    const { scrollTop } = getScrollValues();
    setIsScrolled(scrollTop > 50);
  }, [getScrollValues]);

  const handleResize = useCallback(() => {
    setWindowWidth(window.innerWidth);
  }, []);

  useEffect(() => {
    void loadAllCards();
    setTimeout(() => {
      setIsFadedIn(true);
    }, 10);
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest(".sort-dropdown-container")) {
        setShowSortDropdown(false);
      }
    };

    if (showSortDropdown) {
      document.addEventListener("click", handleClickOutside);
      return () => document.removeEventListener("click", handleClickOutside);
    }
    return undefined;
  }, [showSortDropdown]);

  const handleCardSelect = useCallback(
    (cardId: string) => toggleSetItem(setSelectedCards, cardId),
    [toggleSetItem],
  );

  const generatePermalinkUrl = useCallback(() => {
    const params = new URLSearchParams();
    Array.from(selectedCards).forEach((cardId) => {
      params.append("cId", cardId);
    });
    return `${window.location.origin}/cards?${params.toString()}`;
  }, [selectedCards]);

  const getContainerWidth = useCallback(() => {
    if (windowWidth <= 480) {
      return windowWidth - 16;
    }
    if (windowWidth <= 768) {
      return windowWidth - 20;
    }
    if (windowWidth <= 1100) {
      return windowWidth - 40;
    }
    return Math.min(1300, windowWidth - 40);
  }, [windowWidth]);

  const SECTION_BREAK_HEIGHT = 40;

  const { cardPositions, sectionBreaks } = useMemo(() => {
    const REGULAR_CARD_WIDTH = 220;
    const CORP_CARD_WIDTH = 420;
    const containerWidth = getContainerWidth();

    const positions: Array<{
      cardIndex: number;
      row: number;
      top: number;
      left: number;
      width: number;
      height: number;
    }> = [];

    const breaks: Array<{ top: number; label: string }> = [];

    let currentRow = 0;
    let currentRowCards: Array<{ index: number; left: number; width: number }> = [];
    let currentRowWidth = 0;
    let currentRowHasCorp = false;
    let cumulativeTop = 0;
    let lastTypePriority = -1;

    const sectionLabels: { [key: number]: string } = {
      0: "Cards",
      1: "Preludes",
      2: "Corporations",
    };

    cards.forEach((card, index) => {
      const cardWidth = card.type === CardTypeCorporation ? CORP_CARD_WIDTH : REGULAR_CARD_WIDTH;
      const isCorporation = card.type === CardTypeCorporation;
      const currentPriority = typePriority(card.type);

      if (currentPriority !== lastTypePriority && sectionLabels[currentPriority]) {
        if (currentRowCards.length > 0) {
          const rowOffset = (containerWidth - currentRowWidth) / 2;
          const rowHeight = currentRowHasCorp ? CORP_CARD_HEIGHT : REGULAR_CARD_HEIGHT;
          currentRowCards.forEach((cardInfo) => {
            positions[cardInfo.index].left = cardInfo.left + rowOffset;
          });
          cumulativeTop += rowHeight;
          currentRow++;
          currentRowCards = [];
          currentRowWidth = 0;
          currentRowHasCorp = false;
        }

        breaks.push({ top: cumulativeTop, label: sectionLabels[currentPriority] });
        cumulativeTop += SECTION_BREAK_HEIGHT;
      }
      lastTypePriority = currentPriority;

      if (currentRowWidth + cardWidth > containerWidth && currentRowWidth > 0) {
        const rowOffset = (containerWidth - currentRowWidth) / 2;
        const rowHeight = currentRowHasCorp ? CORP_CARD_HEIGHT : REGULAR_CARD_HEIGHT;

        currentRowCards.forEach((cardInfo) => {
          positions[cardInfo.index].left = cardInfo.left + rowOffset;
        });

        cumulativeTop += rowHeight;
        currentRow++;
        currentRowCards = [];
        currentRowWidth = 0;
        currentRowHasCorp = false;
      }

      const cardLeft = currentRowWidth;
      const rowHeight = isCorporation ? CORP_CARD_HEIGHT : REGULAR_CARD_HEIGHT;

      positions[index] = {
        cardIndex: index,
        row: currentRow,
        top: cumulativeTop,
        left: cardLeft,
        width: cardWidth,
        height: rowHeight,
      };

      currentRowCards.push({ index, left: cardLeft, width: cardWidth });
      currentRowWidth += cardWidth;
      if (isCorporation) {
        currentRowHasCorp = true;
      }
    });

    if (currentRowCards.length > 0) {
      const rowOffset = (containerWidth - currentRowWidth) / 2;
      currentRowCards.forEach((cardInfo) => {
        positions[cardInfo.index].left = cardInfo.left + rowOffset;
      });
    }

    return { cardPositions: positions, sectionBreaks: breaks };
  }, [cards, getContainerWidth]);

  const totalHeight = useMemo(() => {
    if (cardPositions.length === 0) {
      return 0;
    }
    const lastCard = cardPositions[cardPositions.length - 1];
    return lastCard.top + lastCard.height;
  }, [cardPositions]);

  const visibleCards = useMemo(() => {
    const result: Array<{
      card: CardDto;
      position: {
        row: number;
        col: number;
        top: number;
        left: number;
        width: number;
        height: number;
      };
    }> = [];

    cardPositions.forEach((pos) => {
      if (pos.cardIndex >= visibleRange.start && pos.cardIndex < visibleRange.end) {
        result.push({
          card: cards[pos.cardIndex],
          position: {
            row: pos.row,
            col: 0,
            top: pos.top,
            left: pos.left,
            width: pos.width,
            height: pos.height,
          },
        });
      }
    });

    return result;
  }, [cards, cardPositions, visibleRange]);

  useEffect(() => {
    if (cardPositions.length > 0) {
      const newRange = calculateVisibleRange(cardPositions);
      setVisibleRange(newRange);
    }
  }, [cardPositions, calculateVisibleRange, windowWidth, showFilters]);

  useEffect(() => {
    const handleScrollUpdate = () => {
      handleScroll();
      if (cardPositions.length > 0) {
        const newRange = calculateVisibleRange(cardPositions);
        setVisibleRange(newRange);
      }
    };

    const scrollTarget = scrollContainerRef?.current || window;
    scrollTarget.addEventListener("scroll", handleScrollUpdate);
    window.addEventListener("resize", handleResize);
    handleScrollUpdate();
    return () => {
      scrollTarget.removeEventListener("scroll", handleScrollUpdate);
      window.removeEventListener("resize", handleResize);
    };
  }, [handleScroll, handleResize, cardPositions, calculateVisibleRange, scrollContainerRef]);

  const hasActiveChipFilters =
    selectedTags.size > 0 || selectedTypes.size > 0 || selectedPacks.size > 0;
  const hasActiveFilters = searchQuery || hasActiveChipFilters;

  return (
    <div
      className="card-browser"
      style={{
        opacity: isFadedIn ? 1 : 0,
        transition: "opacity 0.3s ease-in",
      }}
    >
      <div
        className={`card-browser-header ${isScrolled ? "scrolled" : ""}`}
        style={
          isOverlay
            ? { position: "sticky" as const, top: 0 }
            : { position: "fixed" as const, top: 0, left: 0, right: 0 }
        }
      >
        <div className="sticky-content">
          <BackButton onClick={onBack}>{backLabel}</BackButton>
          <h1 className="card-browser-title">{title}</h1>
          <div className="right-section">
            <div className="cards-info-header">
              {cards.length > 0 && (
                <span>
                  {hasActiveFilters ? `${cards.length} of ${allCards.length}` : `${cards.length}`}{" "}
                  cards
                </span>
              )}
            </div>
            {selectedCards.size > 0 && (
              <div className="flex items-center gap-2 mr-2">
                <CopyLinkButton
                  textToCopy={generatePermalinkUrl()}
                  defaultText={`Link (${selectedCards.size})`}
                  buttonType="primary"
                  size="sm"
                />
                <GameButton
                  buttonType="secondary"
                  size="sm"
                  onClick={() => setSelectedCards(new Set())}
                >
                  Clear
                </GameButton>
              </div>
            )}
            <button className="filter-toggle-button" onClick={() => setShowFilters(!showFilters)}>
              <span className={`filter-arrow ${showFilters ? "open" : ""}`}>▶</span>
              <span>Filters</span>
              {hasActiveChipFilters && (
                <span
                  className="filter-clear-x"
                  onClick={(e) => {
                    e.stopPropagation();
                    setSelectedTags(new Set());
                    setSelectedTypes(new Set());
                    setSelectedPacks(new Set());
                  }}
                >
                  ✕
                </span>
              )}
            </button>
            <div className="sort-dropdown-container">
              <button
                className="sort-dropdown-button"
                onClick={() => setShowSortDropdown(!showSortDropdown)}
              >
                <span className="sort-display-text">{getSortDisplayText()}</span>
                <span className={`dropdown-arrow ${showSortDropdown ? "open" : ""}`}>▼</span>
              </button>
              {showSortDropdown && (
                <div className="sort-dropdown-menu">
                  {[
                    { key: "name-asc", label: "Name (A-Z)" },
                    { key: "name-desc", label: "Name (Z-A)" },
                    { key: "type-asc", label: "Type (A-Z)" },
                    { key: "type-desc", label: "Type (Z-A)" },
                  ].map((option) => (
                    <button
                      key={option.key}
                      className={`sort-option ${sortBy === option.key ? "active" : ""}`}
                      onClick={() => handleSortSelection(option.key as typeof sortBy)}
                    >
                      <span>{option.label}</span>
                      {sortBy === option.key && <span className="check-icon">✓</span>}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <div className="search-container">
              <input
                type="text"
                placeholder="Search cards (ID, name, tags)..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="search-input"
              />
            </div>
          </div>
        </div>
        {showFilters && (
          <div className="filters-panel">
            <div className="filter-section">
              <h3>Tags</h3>
              <div className="filter-chips">
                {availableTags.map((tag) => (
                  <button
                    key={tag}
                    className={`filter-chip tag-chip ${selectedTags.has(tag) ? "active" : ""}`}
                    onClick={() => toggleSetItem(setSelectedTags, tag)}
                  >
                    <GameIcon iconType={`${tag}-tag`} size="small" />
                  </button>
                ))}
              </div>
            </div>
            <div className="filter-section">
              <h3>Card Types</h3>
              <div className="filter-chips">
                {availableTypes.map((type) => (
                  <button
                    key={type}
                    className={`filter-chip type-chip ${selectedTypes.has(type) ? "active" : ""}`}
                    onClick={() => toggleSetItem(setSelectedTypes, type)}
                    style={{
                      borderColor: getCardTypeColor(type),
                      backgroundColor: selectedTypes.has(type)
                        ? getCardTypeColor(type)
                        : "rgba(255, 255, 255, 0.1)",
                      color: selectedTypes.has(type) ? "#ffffff" : getCardTypeColor(type),
                    }}
                  >
                    {type}
                  </button>
                ))}
              </div>
            </div>
            <div className="filter-section">
              <h3>Card Packs</h3>
              <div className="filter-chips">
                {availablePacks.map((pack) => (
                  <button
                    key={pack}
                    className={`filter-chip pack-chip ${selectedPacks.has(pack) ? "active" : ""}`}
                    onClick={() => toggleSetItem(setSelectedPacks, pack)}
                    style={{
                      borderColor: selectedPacks.has(pack)
                        ? "rgba(0, 188, 212, 0.8)"
                        : "rgba(0, 188, 212, 0.4)",
                      backgroundColor: selectedPacks.has(pack)
                        ? "rgba(0, 188, 212, 0.3)"
                        : "rgba(255, 255, 255, 0.1)",
                      color: selectedPacks.has(pack) ? "#ffffff" : "rgba(0, 188, 212, 0.9)",
                    }}
                  >
                    {formatPackName(pack)}
                  </button>
                ))}
              </div>
            </div>
            {(selectedTags.size > 0 || selectedTypes.size > 0 || selectedPacks.size > 0) && (
              <div className="flex justify-center mt-4">
                <GameButton
                  buttonType="secondary"
                  size="sm"
                  onClick={() => {
                    setSelectedTags(new Set());
                    setSelectedTypes(new Set());
                    setSelectedPacks(new Set());
                  }}
                >
                  Clear All Filters
                </GameButton>
              </div>
            )}
          </div>
        )}
      </div>

      <div className="container">
        {!isOverlay && <div className={`header-spacer ${showFilters ? "expanded" : ""}`}></div>}

        {error && <div className="error-message">{error}</div>}

        <div className="cards-virtual-container" style={{ position: "relative" }}>
          <div style={{ height: `${totalHeight}px`, pointerEvents: "none" }} />

          {sectionBreaks.map((breakInfo) => (
            <div
              key={breakInfo.label}
              className="section-break"
              style={{
                position: "absolute",
                top: `${breakInfo.top}px`,
                left: 0,
                right: 0,
                height: `${SECTION_BREAK_HEIGHT}px`,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                gap: "16px",
              }}
            >
              <div className="section-break-line" />
              <span className="section-break-label">{breakInfo.label}</span>
              <div className="section-break-line" />
            </div>
          ))}

          {visibleCards.map(({ card, position }) => (
            <div
              key={card.id}
              className="card-wrapper"
              style={{
                position: "absolute",
                top: `${position.top}px`,
                left: `${position.left}px`,
                width: `${position.width}px`,
                padding: "0 10px",
                height: `${position.height}px`,
              }}
            >
              {card.type === CardTypeCorporation ? (
                <CorporationCard
                  card={card}
                  isSelected={selectedCards.has(card.id)}
                  onSelect={handleCardSelect}
                  showCheckbox
                  borderColor={getCorporationBorderColor(card.name)}
                />
              ) : (
                <GameCard
                  card={card}
                  isSelected={selectedCards.has(card.id)}
                  onSelect={handleCardSelect}
                  animationDelay={0}
                  showCheckbox
                />
              )}
            </div>
          ))}
        </div>

        {loading && <div className="loading-message">Loading all cards...</div>}
      </div>

      <style>{`
        .card-browser {
          background: #000000;
          color: white;
          min-height: 100vh;
          font-family:
            -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen",
            "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue",
            sans-serif;
          position: relative;
          z-index: 10;
          overflow-x: hidden;
        }

        .card-browser-header {
          z-index: 1000;
          transition: all 0.3s ease;
          padding: 0;
          background: rgba(0, 0, 0, 0.95);
          backdrop-filter: blur(10px);
        }

        .card-browser-header.scrolled {
          background: rgba(5, 5, 10, 0.98);
          backdrop-filter: blur(10px);
          box-shadow: 0 2px 20px rgba(0, 0, 0, 0.5);
        }

        .sticky-content {
          max-width: 1400px;
          margin: 0 auto;
          padding: 12px 20px;
          display: flex;
          align-items: center;
          gap: 16px;
          position: relative;
        }


        .container {
          max-width: 1400px;
          margin: 0 auto;
          padding: 4px 20px;
        }

        .header-spacer {
          height: 60px;
          transition: height 0.3s ease;
        }

        .header-spacer.expanded {
          height: 260px;
        }

        .card-browser-title {
          font-family: 'Orbitron', sans-serif;
          font-size: 16px;
          color: #ffffff;
          margin: 0;
          text-shadow: 0 0 30px rgba(30, 60, 150, 0.6);
          font-weight: bold;
          letter-spacing: 2px;
          white-space: nowrap;
          flex-shrink: 0;
        }

        .right-section {
          display: flex;
          align-items: center;
          gap: 12px;
          flex: 1;
          justify-content: flex-end;
          min-width: 0;
        }

        .cards-info-header {
          color: rgba(255, 255, 255, 0.7);
          font-size: 13px;
          white-space: nowrap;
          flex-shrink: 0;
        }

        .search-container {
          display: flex;
          align-items: center;
        }

        .search-input {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          padding: 8px 14px;
          color: white;
          font-size: 13px;
          width: 240px;
          min-width: 140px;
          backdrop-filter: blur(10px);
          transition: all 0.3s ease;
        }

        .search-input::placeholder {
          color: rgba(255, 255, 255, 0.5);
        }

        .search-input:focus {
          outline: none;
          border-color: rgba(30, 60, 150, 0.8);
          background: rgba(10, 10, 15, 0.9);
          box-shadow: 0 0 25px rgba(30, 60, 150, 0.5);
        }

        .filter-toggle-button {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          padding: 8px 14px;
          color: white;
          font-size: 13px;
          cursor: pointer;
          transition: all 0.3s ease;
          backdrop-filter: blur(10px);
          display: flex;
          align-items: center;
          gap: 6px;
          white-space: nowrap;
          flex-shrink: 0;
        }

        .filter-toggle-button:hover {
          background: rgba(10, 10, 15, 0.9);
          border-color: rgba(30, 60, 150, 0.8);
          box-shadow: 0 0 20px rgba(30, 60, 150, 0.6);
        }

        .filter-arrow {
          transition: transform 0.3s ease;
          font-size: 12px;
        }

        .filter-arrow.open {
          transform: rotate(90deg);
        }

        .filter-clear-x {
          margin-left: 2px;
          padding: 0 4px;
          border-radius: 4px;
          font-size: 11px;
          color: rgba(255, 255, 255, 0.5);
          transition: all 0.2s ease;
          cursor: pointer;
        }

        .filter-clear-x:hover {
          color: #ff6b6b;
          background: rgba(255, 107, 107, 0.15);
        }

        .sort-dropdown-container {
          position: relative;
        }

        .sort-dropdown-button {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          padding: 8px 14px;
          color: white;
          font-size: 13px;
          cursor: pointer;
          transition: all 0.3s ease;
          backdrop-filter: blur(10px);
          display: flex;
          align-items: center;
          gap: 6px;
          min-width: 120px;
          justify-content: space-between;
          white-space: nowrap;
          flex-shrink: 0;
        }

        .sort-dropdown-button:hover {
          background: rgba(10, 10, 15, 0.9);
          border-color: rgba(30, 60, 150, 0.8);
          box-shadow: 0 0 20px rgba(30, 60, 150, 0.6);
        }

        .sort-display-text {
          font-weight: 500;
        }

        .dropdown-arrow {
          transition: transform 0.2s ease;
          font-size: 12px;
        }

        .dropdown-arrow.open {
          transform: rotate(180deg);
        }

        .sort-dropdown-menu {
          position: absolute;
          top: 100%;
          left: 0;
          right: 0;
          background: rgba(0, 0, 0, 0.98);
          border: 2px solid rgba(255, 255, 255, 0.1);
          border-radius: 8px;
          backdrop-filter: blur(10px);
          z-index: 1000;
          margin-top: 4px;
          box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
        }

        .sort-option {
          width: 100%;
          background: none;
          border: none;
          padding: 12px 16px;
          color: white;
          font-size: 14px;
          cursor: pointer;
          transition: all 0.2s ease;
          display: flex;
          align-items: center;
          justify-content: space-between;
          text-align: left;
        }

        .sort-option:hover {
          background: rgba(255, 255, 255, 0.1);
        }

        .sort-option.active {
          background: rgba(30, 60, 150, 0.2);
          color: rgba(30, 60, 150, 1);
        }

        .check-icon {
          color: rgba(30, 60, 150, 1);
          font-weight: bold;
        }

        .filters-panel {
          background: rgba(0, 0, 0, 0.85);
          backdrop-filter: blur(10px);
          padding: 20px;
          width: 100%;
          position: relative;
          animation: cardBrowserSlideDown 0.3s ease-out forwards;
          transform-origin: top;
        }

        @keyframes cardBrowserSlideDown {
          from {
            transform: translateY(-20px);
          }
          to {
            transform: translateY(0);
          }
        }

        .filters-panel::after {
          content: '';
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          height: 1px;
          background: linear-gradient(
            90deg,
            transparent 0%,
            rgba(255, 255, 255, 0.1) 20%,
            rgba(255, 255, 255, 0.2) 50%,
            rgba(255, 255, 255, 0.1) 80%,
            transparent 100%
          );
        }

        .filter-section {
          margin-bottom: 20px;
        }

        .filter-section:last-child {
          margin-bottom: 0;
        }

        .filter-section h3 {
          color: white;
          font-size: 16px;
          margin: 0 0 12px 0;
          font-weight: 600;
        }

        .filter-chips {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
          justify-content: center;
        }

        .filter-chip {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 20px;
          padding: 6px 12px;
          color: white;
          font-size: 12px;
          cursor: pointer;
          transition: all 0.3s ease;
          backdrop-filter: blur(5px);
        }

        .filter-chip:hover {
          background: rgba(10, 10, 15, 0.9);
          border-color: rgba(30, 60, 150, 0.8);
          box-shadow: 0 0 15px rgba(30, 60, 150, 0.4);
        }

        .filter-chip.active {
          background: rgba(30, 60, 150, 0.3);
          border-color: rgba(30, 60, 150, 0.8);
          color: white;
          box-shadow: 0 0 20px rgba(30, 60, 150, 0.5);
        }

        .tag-chip {
          display: flex;
          align-items: center;
          justify-content: center;
          min-width: 40px;
          min-height: 32px;
          padding: 6px 8px;
        }

        .type-chip {
          font-weight: 600;
          text-transform: capitalize;
          transition: all 0.3s ease;
        }

        .type-chip:hover {
          transform: scale(1.05);
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
        }

        .pack-chip {
          font-weight: 600;
          font-size: 13px;
          transition: all 0.3s ease;
        }

        .pack-chip:hover {
          transform: scale(1.05);
          box-shadow: 0 2px 8px rgba(0, 188, 212, 0.3);
        }

        .section-break-line {
          flex: 1;
          height: 1px;
          background: linear-gradient(
            90deg,
            transparent 0%,
            rgba(255, 255, 255, 0.3) 50%,
            transparent 100%
          );
        }

        .section-break-label {
          font-family: 'Orbitron', sans-serif;
          font-size: 14px;
          font-weight: 600;
          color: rgba(255, 255, 255, 0.5);
          letter-spacing: 2px;
          text-transform: uppercase;
          white-space: nowrap;
        }

        .error-message {
          color: #ff6b6b;
          margin-bottom: 40px;
          padding: 16px;
          background: rgba(255, 107, 107, 0.1);
          border: 1px solid rgba(255, 107, 107, 0.3);
          border-radius: 8px;
          font-size: 14px;
          text-align: center;
        }

        .cards-virtual-container {
          max-width: 1300px;
          margin: 0 auto 40px auto;
          position: relative;
          width: 100%;
          overflow: hidden;
        }

        .card-wrapper {
          box-sizing: border-box;
        }

        .loading-message {
          text-align: center;
          padding: 20px;
          color: rgba(30, 60, 150, 1);
          font-size: 16px;
          background: rgba(30, 60, 150, 0.15);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          animation: cardBrowserPulse 2s infinite;
          box-shadow: 0 0 15px rgba(30, 60, 150, 0.3);
        }

        @keyframes cardBrowserPulse {
          0%, 100% {
            opacity: 1;
          }
          50% {
            opacity: 0.7;
          }
        }

        @media (max-width: 1200px) {
          .sticky-content {
            padding: 15px;
          }

          .right-section {
            flex-wrap: wrap;
            gap: 10px;
          }

          .search-input {
            width: 250px;
          }
        }

        @media (max-width: 768px) {
          .sticky-content {
            flex-direction: column;
            text-align: center;
            gap: 15px;
            padding: 15px 10px;
          }

          .right-section {
            flex-direction: column;
            width: 100%;
            gap: 10px;
          }

          .card-browser-title {
            font-size: 20px;
            order: -1;
          }

          .cards-info-header {
            order: 1;
          }

          .search-input {
            width: 100%;
            max-width: none;
          }

          .filter-toggle-button,
          .sort-dropdown-button {
            width: 100%;
            justify-content: center;
          }

          .container {
            padding: 4px 10px;
          }

          .header-spacer {
            height: 100px;
          }

          .header-spacer.expanded {
            height: 380px;
          }

          .filters-panel {
            padding: 15px 10px;
          }

          .filter-chips {
            gap: 6px;
            justify-content: flex-start;
          }

          .filter-chip {
            font-size: 11px;
            padding: 5px 10px;
          }

          .cards-virtual-container {
            padding: 0 5px;
          }
        }

        @media (max-width: 480px) {
          .sticky-content {
            padding: 12px 8px;
            gap: 12px;
          }

          .card-browser-title {
            font-size: 18px;
            letter-spacing: 1px;
          }

          .filter-toggle-button,
          .sort-dropdown-button {
            padding: 10px 16px;
            font-size: 13px;
          }

          .search-input {
            padding: 10px 14px;
            font-size: 13px;
          }

          .header-spacer {
            height: 120px;
          }

          .header-spacer.expanded {
            height: 410px;
          }

          .container {
            padding: 4px 8px;
          }

          .cards-info-header {
            font-size: 13px;
          }
        }
      `}</style>
    </div>
  );
};

export default CardBrowser;
