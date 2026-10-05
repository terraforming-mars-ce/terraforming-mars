import type {
  GameDto,
  PlayerDto,
  OtherPlayerDto,
  ResourcesDto,
  ProductionDto,
  GlobalParametersDto,
  PlayerCardDto,
  PlayerActionDto,
  PlayerStandardProjectDto,
  PlayerMilestoneDto,
  PlayerAwardDto,
  CardDto,
  CardDescriptionSectionDto,
  TileDto,
  PendingTileSelectionDto,
  PendingCardSelectionDto,
  PendingCardDrawSelectionDto,
  PendingBehaviorResolutionDto,
  ForcedFirstActionDto,
  CardBehaviorDto,
  ResourceCondition,
} from "./types.js";
import { GameState } from "./state.js";

const DESCRIPTION_LABELS: Record<CardDescriptionSectionDto["type"], string | null> = {
  generic: null,
  effect: "Effect",
  action: "Action",
  requirement: "Requirement",
};

// Flatten a card description into labelled lines, indenting continuation lines for bullet output
function formatCardDescription(sections: CardDescriptionSectionDto[], indent = "    "): string {
  return sections
    .map((s) => (DESCRIPTION_LABELS[s.type] ? `${DESCRIPTION_LABELS[s.type]}: ${s.text}` : s.text))
    .join(`\n${indent}`);
}

export function summarizeGameState(
  state: GameState,
  verbose = false,
): string {
  const game = state.game;
  if (!game) return "No game state available. Use connect_to_game first.";

  const lines: string[] = [];

  lines.push(formatGameInfo(game, state));
  lines.push(formatGlobalParams(game.globalParameters));

  const player = game.currentPlayer;
  if (player) {
    lines.push(formatPendingActions(player));
    lines.push(formatPlayerStatus(player));
    lines.push(formatHand(player.cards, verbose));
    lines.push(formatCardActions(player.actions, verbose));
    lines.push(formatStandardProjects(player.standardProjects));
    lines.push(formatMilestones(player.milestones));
    lines.push(formatAwards(player.awards));
  }

  lines.push(formatOpponents(game.otherPlayers));
  lines.push(formatBoard(game.board.tiles, verbose));

  if (game.finalScores?.length) {
    lines.push(formatFinalScores(game));
  }

  return lines.filter((l) => l.length > 0).join("\n\n");
}

function formatGameInfo(game: GameDto, state: GameState): string {
  const turnInfo = game.currentTurn
    ? game.currentTurn === state.myPlayerId
      ? "YOUR TURN"
      : `Waiting for ${findPlayerName(game, game.currentTurn)}`
    : "N/A";

  return [
    "=== GAME INFO ===",
    `Game ID: ${game.id}`,
    `Phase: ${game.currentPhase}`,
    `Status: ${game.status}`,
    `Generation: ${game.generation}`,
    `Turn: ${turnInfo}`,
    `Players: ${game.turnOrder.length}`,
    `Turn order: ${game.turnOrder.map((id) => findPlayerName(game, id)).join(" → ")}`,
  ].join("\n");
}

function formatGlobalParams(gp: GlobalParametersDto): string {
  return [
    "=== GLOBAL PARAMETERS ===",
    `Temperature: ${gp.temperature}°C (target: 8°C)`,
    `Oxygen: ${gp.oxygen}% (target: 14%)`,
    `Oceans: ${gp.oceans}/${gp.maxOceans}`,
    `Venus: ${gp.venus}% (target: 30%)`,
  ].join("\n");
}

function formatPendingActions(player: PlayerDto): string {
  const parts: string[] = [];

  if (player.pendingTileSelection) {
    parts.push(formatPendingTile(player.pendingTileSelection));
  }
  if (player.pendingCardSelection) {
    parts.push(formatPendingCardSelection(player.pendingCardSelection));
  }
  if (player.pendingCardDrawSelection) {
    parts.push(formatPendingCardDraw(player.pendingCardDrawSelection));
  }
  for (const resolution of player.pendingBehaviorResolutions ?? []) {
    parts.push(resolution.kind === "card-discard" ? formatPendingCardDiscard(resolution) : formatPendingBehaviorChoice(resolution));
  }
  if (player.pendingCardReveal) {
    parts.push(`Resolved public reveal: ${JSON.stringify(player.pendingCardReveal)}. Acknowledge with confirm_cards action reveal.`);
  }
  if (player.pendingResourceRemovalSelection) {
    parts.push(`Pending removal: ${JSON.stringify(player.pendingResourceRemovalSelection)}. Use confirm_resource_removal; empty target and amount 0 skips.`);
  }
  if (player.pendingEffectSelection) {
    parts.push(`Pending effect selection: ${JSON.stringify(player.pendingEffectSelection)}. Use confirm_cards action effect with choiceIndex.`);
  }

  for (const receipt of player.cardReceipts ?? []) parts.push(`Cards already received from ${receipt.source}: ${receipt.cards.map(c => c.name).join(", ")}. Optional acknowledgement: acknowledge_cards_received with receiptId ${receipt.id}.`);
  if (player.pendingColonySelection) parts.push(`Pending colony choice: ${JSON.stringify(player.pendingColonySelection)}. Use confirm_colony_selection.`);
  if (player.pendingColonyResourceSelection) parts.push(`Pending colony resource: ${JSON.stringify(player.pendingColonyResourceSelection)}. Use confirm_colony_resource.`);
  if (player.pendingAwardFundSelection) parts.push(`Pending free award: ${JSON.stringify(player.pendingAwardFundSelection)}. Use confirm_free_award.`);
  if (player.forcedFirstAction?.state === "resolving") {
    parts.push(formatForcedAction(player.forcedFirstAction));
  }

  if (
    player.selectCorporationPhase ||
    player.selectStartingCardsPhase ||
    player.selectPreludeCardsPhase
  ) {
    parts.push(formatStartingSelection(player));
  }

  if (
    player.productionPhase &&
    !player.productionPhase.selectionComplete
  ) {
    parts.push(formatProductionPhase(player));
  }

  if (parts.length === 0) return "";

  return ">>> PENDING ACTIONS (must resolve) <<<\n" + parts.join("\n\n");
}

function formatPendingTile(sel: PendingTileSelectionDto): string {
  return [
    `TILE PLACEMENT REQUIRED: Place a ${sel.tileType} tile`,
    `Source: ${sel.source}`,
    `Available hexes: ${sel.availableHexes.join(", ")}`,
    `Use select_tile tool with q, r, s coordinates.`,
  ].join("\n");
}

function formatPendingCardSelection(sel: PendingCardSelectionDto): string {
  const cardList = sel.availableCards
    .map((c) => {
      const cost = sel.cardCosts[c.id] ?? 0;
      const reward = sel.cardRewards[c.id] ?? 0;
      const info = cost > 0 ? ` (cost: ${cost}M€)` : reward > 0 ? ` (reward: ${reward}M€)` : "";
      return `  - ${c.name} [${c.id}]${info}`;
    })
    .join("\n");

  return [
    `CARD SELECTION REQUIRED (source: ${sel.source})`,
    `Select ${sel.minCards}-${sel.maxCards} cards:`,
    cardList,
    `Use confirm_cards tool with action="select" and cardIds.`,
  ].join("\n");
}

function formatPendingCardDraw(sel: PendingCardDrawSelectionDto): string {
  const cardList = sel.availableCards
    .map((c) => `  - ${c.name} [${c.id}]: ${formatCardDescription(c.description)}`)
    .join("\n");

  const parts = [
    `CARD DRAW SELECTION (source: ${sel.source})`,
    `Free takes: ${sel.minFreeTakeCount} to ${sel.freeTakeCount}, Max buy: ${sel.maxBuyCount}${sel.cardBuyCost > 0 ? ` (${sel.cardBuyCost}M€ each)` : ""}`,
    cardList,
    `Use confirm_cards tool with action="draw" and cardsToTake/cardsToBuy.`,
  ];

  return parts.join("\n");
}

function formatPendingCardDiscard(sel: PendingBehaviorResolutionDto): string {
  return [
    `CARD DISCARD REQUIRED (source: ${sel.source})`,
    `Discard ${sel.minCards}-${sel.maxCards} cards from hand.`,
    `Use confirm_cards tool with action="discard" and cardsToDiscard, with resolutionId="${sel.id}".`,
  ].join("\n");
}

function formatPendingBehaviorChoice(
  sel: PendingBehaviorResolutionDto,
): string {
  const choiceList = (sel.choices ?? [])
    .map((c, i) => {
      const desc = formatBehaviorBrief(c.inputs, c.outputs);
      const avail = c.available ? "" : " [UNAVAILABLE]";
      return `  ${i}: ${desc}${avail}`;
    })
    .join("\n");

  return [
    `BEHAVIOR CHOICE REQUIRED (source: ${sel.source})`,
    choiceList,
    `Use confirm_cards tool with action="behavior-choice" and choiceIndex, with resolutionId="${sel.id}". Triggering card: ${sel.triggeringCardName ?? "none"}; triggering-card targets are fixed.`,
  ].join("\n");
}

function formatForcedAction(fa: ForcedFirstActionDto): string {
  return [
    `FORCED FIRST ACTION: ${fa.description}`,
    `State: ${fa.state}`,
    `Corporation: ${fa.corporationId}`,
  ].join("\n");
}

function formatStartingSelection(player: PlayerDto): string {
  const parts: string[] = ["STARTING SELECTION REQUIRED"];

  if (player.selectCorporationPhase) {
    const corps = player.selectCorporationPhase.availableCorporations
      .map((c) => `  - ${c.name} [${c.id}]: ${formatCardDescription(c.description)}`)
      .join("\n");
    parts.push("Corporations:\n" + corps);
  }

  if (player.selectPreludeCardsPhase) {
    const preludes = player.selectPreludeCardsPhase.availablePreludes
      .map((c) => `  - ${c.name} [${c.id}]: ${formatCardDescription(c.description)}`)
      .join("\n");
    parts.push(
      `Preludes (pick ${player.selectPreludeCardsPhase.maxSelectable}):\n` +
        preludes,
    );
  }

  if (player.selectStartingCardsPhase) {
    const cards = player.selectStartingCardsPhase.availableCards
      .map(
        (c) =>
          `  - ${c.name} [${c.id}] (${c.cost}M€) [${c.type}] ${c.tags?.join(", ") || ""}: ${formatCardDescription(c.description)}`,
      )
      .join("\n");
    parts.push("Starting cards (pick any to buy at 3M€ each):\n" + cards);
  }

  parts.push(
    "Use select_starting_choices tool with corporationId, preludeIds, and cardIds.",
  );

  return parts.join("\n");
}

function formatProductionPhase(player: PlayerDto): string {
  const pp = player.productionPhase!;
  const cards = pp.availableCards
    .map((c) => `  - ${c.name} [${c.id}]`)
    .join("\n");

  return [
    "PRODUCTION PHASE - Select cards to buy:",
    cards || "  (no cards available)",
    `Use confirm_cards tool with action="production" and cardIds.`,
  ].join("\n");
}

function formatPlayerStatus(player: PlayerDto): string {
  const r = player.resources;
  const p = player.production;

  const lines = [
    "=== YOUR STATUS ===",
    `Name: ${player.name} | Corporation: ${player.corporation?.name || "None"} | TR: ${player.terraformRating}`,
    `Status: ${player.status} | Actions remaining: ${player.availableActions} | Passed: ${player.passed}`,
    "",
    "Resources (amount / production):",
    `  Credits:  ${r.credits} / ${formatProd(p.credits)}`,
    `  Steel:    ${r.steel} / ${formatProd(p.steel)}`,
    `  Titanium: ${r.titanium} / ${formatProd(p.titanium)}`,
    `  Plants:   ${r.plants} / ${formatProd(p.plants)}`,
    `  Energy:   ${r.energy} / ${formatProd(p.energy)}`,
    `  Heat:     ${r.heat} / ${formatProd(p.heat)}`,
  ];

  lines.push("", "Legal resource removal targets (subject to the card's target restrictions):");
  for (const target of player.resourceRemovalTargets) {
    lines.push(`  player=${target.playerId} card=${target.cardId || ""} resource=${target.resourceType} amount=${target.amount}`);
  }

  if (player.playedCards.length > 0) {
    lines.push(
      "",
      `Played cards (${player.playedCards.length}): ${player.playedCards.map((c) => c.name).join(", ")}`,
    );
  }

  const storage = Object.entries(player.resourceStorage).filter(
    ([, v]) => v > 0,
  );
  if (storage.length > 0) {
    lines.push(
      `Resource storage: ${storage.map(([k, v]) => `${k}: ${v}`).join(", ")}`,
    );
  }

  if (player.paymentSubstitutes.length > 0) {
    lines.push(
      `Payment substitutes: ${player.paymentSubstitutes.map((s) => `${s.source.resource} (${s.conversionRate}:1)`).join(", ")}`,
    );
  }

  if (player.effects.length > 0) {
    lines.push(
      `Active effects: ${player.effects.map((e) => e.cardName).join(", ")}`,
    );
  }

  return lines.join("\n");
}

function formatHand(cards: PlayerCardDto[], verbose: boolean): string {
  if (!cards || cards.length === 0) return "=== HAND (0 cards) ===\n(empty)";

  const header = `=== HAND (${cards.length} cards) ===`;
  const cardLines = cards.map((c) => {
    const avail = c.available ? "PLAYABLE" : "BLOCKED";
    const errInfo =
      !c.available && c.errors.length > 0
        ? ` (${c.errors.map((e) => e.message).join("; ")})`
        : "";
    const tags = c.tags?.length ? ` [${c.tags.join(", ")}]` : "";
    const discount =
      c.effectiveCost < c.cost
        ? ` (discounted from ${c.cost})`
        : "";

    let line = `  - ${c.name} [${c.id}] | ${c.effectiveCost}M€${discount} | ${c.type}${tags} | ${avail}${errInfo}`;

    if (verbose && c.description.length > 0) {
      line += `\n    ${formatCardDescription(c.description)}`;
    }

    return line;
  });

  return header + "\n" + cardLines.join("\n");
}

function formatCardActions(
  actions: PlayerActionDto[],
  verbose: boolean,
): string {
  if (!actions || actions.length === 0) return "";

  const header = "=== CARD ACTIONS ===";
  const actionLines = actions.map((a) => {
    const avail = a.available ? "AVAILABLE" : "BLOCKED";
    const errInfo =
      !a.available && a.errors.length > 0
        ? ` (${a.errors.map((e) => e.message).join("; ")})`
        : "";
    const usedInfo =
      a.timesUsedThisTurn > 0 ? ` [used ${a.timesUsedThisTurn}x this turn]` : "";
    const desc = formatBehaviorBrief(
      a.behavior.inputs,
      a.behavior.outputs,
    );

    let line = `  - ${a.cardName} [${a.cardId}] behavior#${a.behaviorIndex} | ${avail}${usedInfo}${errInfo}`;
    if (desc) line += `\n    ${desc}`;

    if (verbose && a.behavior.description) {
      line += `\n    Description: ${a.behavior.description}`;
    }

    for (const option of a.reuseOptions ?? []) {
      if (option.available) {
        line += `\n    Reuse target cardId=${option.cardId} behaviorIndex=${option.behaviorIndex} with reuseSourceCardId=${a.cardId}`;
      }
    }

    return line;
  });

  return header + "\n" + actionLines.join("\n");
}

function formatStandardProjects(
  projects: PlayerStandardProjectDto[],
): string {
  if (!projects || projects.length === 0) return "";

  const header = "=== STANDARD PROJECTS ===";
  const lines = projects.map((p) => {
    const avail = p.available ? "AVAILABLE" : "BLOCKED";
    const errInfo =
      !p.available && p.errors.length > 0
        ? ` (${p.errors[0].message})`
        : "";
    const costStr = Object.entries(p.effectiveCost)
      .map(([k, v]) => `${v} ${k}`)
      .join(", ");

    return `  - ${p.projectType} | ${costStr} | ${avail}${errInfo}`;
  });

  return header + "\n" + lines.join("\n");
}

function formatMilestones(milestones: PlayerMilestoneDto[]): string {
  if (!milestones || milestones.length === 0) return "";

  const header = "=== MILESTONES ===";
  const lines = milestones.map((m) => {
    const claimed = m.isClaimed
      ? `CLAIMED by ${m.claimedBy}`
      : m.available
        ? `CLAIMABLE (${m.claimCost}M€)`
        : "NOT YET";
    return `  - ${m.name}: ${m.description} | Progress: ${m.progress}/${m.required} | ${claimed}`;
  });

  return header + "\n" + lines.join("\n");
}

function formatAwards(awards: PlayerAwardDto[]): string {
  if (!awards || awards.length === 0) return "";

  const header = "=== AWARDS ===";
  const lines = awards.map((a) => {
    const funded = a.isFunded
      ? `FUNDED by ${a.fundedBy}`
      : a.available
        ? `FUNDABLE (${a.fundingCost}M€)`
        : "NOT AVAILABLE";
    return `  - ${a.name}: ${a.description} | ${funded}`;
  });

  return header + "\n" + lines.join("\n");
}

function formatOpponents(others: OtherPlayerDto[]): string {
  if (!others || others.length === 0) return "";

  const header = "=== OPPONENTS ===";
  const lines = others.map((o) => {
    const r = o.resources;
    const corp = o.corporation?.name || "None";
    return [
      `  ${o.name} (${corp}) | TR: ${o.terraformRating} | Status: ${o.status} | Passed: ${o.passed}`,
      `    Resources: ${r.credits}M€, ${r.steel} steel, ${r.titanium} ti, ${r.plants} plants, ${r.energy} energy, ${r.heat} heat`,
      `    Cards in hand: ${o.handCardCount} | Played: ${o.playedCards.length} cards`,
    ].join("\n");
  });

  return header + "\n" + lines.join("\n");
}

function formatBoard(tiles: TileDto[], verbose: boolean): string {
  const occupied = tiles.filter((t) => t.occupiedBy);
  if (occupied.length === 0 && !verbose) return "";

  const header = "=== BOARD ===";
  const tilesToShow = verbose ? tiles : occupied;

  if (tilesToShow.length === 0) return header + "\n(no occupied tiles)";

  const lines = tilesToShow.map((t) => {
    const coord = `(${t.coordinates.q},${t.coordinates.r},${t.coordinates.s})`;
    const occ = t.occupiedBy
      ? ` | ${t.occupiedBy.type}${t.ownerId ? ` (owner: ${t.ownerId})` : ""}`
      : "";
    const name = t.displayName ? ` ${t.displayName}` : "";
    const bonuses =
      t.bonuses.length > 0
        ? ` | bonuses: ${t.bonuses.map((b) => `${b.amount}x ${b.type}`).join(", ")}`
        : "";
    return `  ${coord}${name}${occ}${bonuses}`;
  });

  return header + "\n" + lines.join("\n");
}

function formatFinalScores(game: GameDto): string {
  if (!game.finalScores?.length) return "";

  const header = "=== FINAL SCORES ===";
  const lines = game.finalScores.map((s) => {
    const vp = s.vpBreakdown;
    return [
      `  #${s.placement} ${s.playerName}${s.isWinner ? " (WINNER)" : ""}: ${vp.totalVP} VP`,
      `    TR: ${vp.terraformRating}, Cards: ${vp.cardVP}, Greenery: ${vp.greeneryVP}, City: ${vp.cityVP}, Milestones: ${vp.milestoneVP}, Awards: ${vp.awardVP}`,
    ].join("\n");
  });

  return header + "\n" + lines.join("\n");
}

function formatProd(val: number): string {
  return val >= 0 ? `+${val}` : `${val}`;
}

function formatBehaviorBrief(
  inputs?: ResourceCondition[],
  outputs?: ResourceCondition[],
): string {
  const parts: string[] = [];

  if (inputs?.length) {
    parts.push(
      "Costs: " +
        inputs.map((i) => `${i.amount} ${i.type}`).join(", "),
    );
  }
  if (outputs?.length) {
    parts.push(
      "Gives: " +
        outputs.map((o) => `${o.amount} ${o.type}`).join(", "),
    );
  }

  return parts.join(" → ");
}

function findPlayerName(game: GameDto, playerId: string): string {
  if (game.currentPlayer?.id === playerId) return game.currentPlayer.name;
  const other = game.otherPlayers.find((p) => p.id === playerId);
  return other?.name || playerId;
}
