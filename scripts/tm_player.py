#!/usr/bin/env python3
"""Terraforming Mars autonomous bot - plays the game via WebSocket using Claude CLI.

Usage:
    python tm_player.py                     # auto-join active game on prod
    python tm_player.py --local             # auto-join on localhost:3001
    python tm_player.py --game-id <id>      # join specific game
    python tm_player.py --model sonnet      # use specific Claude model
"""

import argparse
import asyncio
import json
import os
import re
import subprocess
import sys
import tempfile
import urllib.request
import urllib.error

import websockets

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

ACTION_MAP = {
    "play_card": "action.card.play-card",
    "use_card_action": "action.card.card-action",
    "skip_action": "action.game-management.skip-action",
    "select_tile": "action.tile-selection.tile-selected",
    "select_starting_choices": "action.card.select-starting-choices",
    "claim_milestone": "action.milestone.claim-milestone",
    "fund_award": "action.award.fund-award",
}

STANDARD_PROJECT_WS_TYPES = {
    "sell-patents": "action.standard-project.sell-patents",
    "power-plant": "action.standard-project.build-power-plant",
    "asteroid": "action.standard-project.launch-asteroid",
    "aquifer": "action.standard-project.build-aquifer",
    "greenery": "action.standard-project.plant-greenery",
    "city": "action.standard-project.build-city",
}

CONVERSION_WS_TYPES = {
    "plants-to-greenery": "action.resource-conversion.convert-plants-to-greenery",
    "heat-to-temperature": "action.resource-conversion.convert-heat-to-temperature",
}

CONFIRM_CARDS_WS_TYPES = {
    "select": "action.standard-project.confirm-sell-patents",
    "production": "action.card.confirm-production-cards",
    "draw": "action.card.card-draw-confirmed",
    "discard": "action.card.card-discard-confirmed",
    "behavior-choice": "action.card.behavior-choice-confirmed",
}

GAME_PHASE_ACTION = "action"
GAME_PHASE_STARTING_SELECTION = "starting_selection"
GAME_PHASE_PRODUCTION = "production_and_card_draw"
GAME_PHASE_COMPLETE = "complete"

PLAYER_STATUS_ACTIVE = "active"
PLAYER_STATUS_SELECTING_STARTING = "selecting-starting-cards"
PLAYER_STATUS_SELECTING_PRODUCTION = "selecting-production-cards"

SYSTEM_PROMPT = """You are an expert Terraforming Mars player. You are given the current game state and must decide what action to take.

Respond with ONLY a JSON object (no markdown, no explanation). The JSON must have an "action" key.

Available actions:

1. play_card - Play a card from your hand
   {"action": "play_card", "cardId": "...", "credits": N, "steel": N, "titanium": N, "heat": N, "choiceIndex": N, "targetPlayerId": "...", "selectedAmount": N}
   - credits is REQUIRED (amount of credits to spend)
   - steel/titanium/heat are optional (default 0)
   - choiceIndex only for cards with multiple options
   - targetPlayerId only for cards targeting opponents

2. use_card_action - Activate a played card's action
   {"action": "use_card_action", "cardId": "...", "behaviorIndex": N}
   - Optional: choiceIndex, cardStorageTargets, targetPlayerId, sourceCardForInput, selectedAmount, credits, steel, titanium

3. standard_project - Execute a standard project
   {"action": "standard_project", "project": "sell-patents|power-plant|asteroid|aquifer|greenery|city"}

4. convert_resources - Convert resources
   {"action": "convert_resources", "conversion": "plants-to-greenery|heat-to-temperature"}

5. skip_action - Pass/skip your turn
   {"action": "skip_action"}

6. select_tile - Place a tile (when TILE PLACEMENT REQUIRED)
   {"action": "select_tile", "q": N, "r": N, "s": N}

7. select_starting_choices - Select corporation and starting cards
   {"action": "select_starting_choices", "corporationId": "...", "preludeIds": [...], "cardIds": [...]}

8. confirm_cards - Confirm card selections
   {"action": "confirm_cards", "confirmAction": "select|production|draw|discard|behavior-choice", ...}
   - For "select": include "cardIds": [...]
   - For "production": include "cardIds": [...]
   - For "draw": include "cardsToTake": [...] and/or "cardsToBuy": [...]
   - For "discard": include "cardsToDiscard": [...]
   - For "behavior-choice": include "choiceIndex": N

9. claim_milestone - Claim a milestone
   {"action": "claim_milestone", "milestoneType": "..."}

10. fund_award - Fund an award
    {"action": "fund_award", "awardType": "..."}

STRATEGY TIPS:
- Resolve PENDING ACTIONS first (tile placement, card selection, etc.)
- Look at card costs vs your resources when playing cards
- Use steel for building tags (2M per steel) and titanium for space tags (3M per titanium)
- Try to play cards that synergize with your corporation and existing cards
- Convert plants to greenery when you have 8+ plants
- Convert heat to temperature when you have 8+ heat
- Claim milestones early if you qualify (they cost 8M)
- Fund awards that benefit your strategy
- Skip only when no good actions are available

IMPORTANT: Respond with ONLY the JSON object. No other text."""


# ---------------------------------------------------------------------------
# Action Builder
# ---------------------------------------------------------------------------

def build_ws_message(action_dict):
    """Convert an action JSON to (ws_type, payload) for the WebSocket."""
    action = action_dict.get("action")

    if action == "standard_project":
        project = action_dict.get("project", "")
        ws_type = STANDARD_PROJECT_WS_TYPES.get(project)
        if not ws_type:
            raise ValueError(f"Unknown standard project: {project}")
        return ws_type, {}

    if action == "convert_resources":
        conversion = action_dict.get("conversion", "")
        ws_type = CONVERSION_WS_TYPES.get(conversion)
        if not ws_type:
            raise ValueError(f"Unknown conversion: {conversion}")
        return ws_type, {}

    if action == "confirm_cards":
        confirm_action = action_dict.get("confirmAction", "")
        ws_type = CONFIRM_CARDS_WS_TYPES.get(confirm_action)
        if not ws_type:
            raise ValueError(f"Unknown confirm action: {confirm_action}")
        return ws_type, _build_confirm_payload(confirm_action, action_dict)

    if action == "play_card":
        payment = {
            "credits": action_dict.get("credits", 0),
            "steel": action_dict.get("steel", 0),
            "titanium": action_dict.get("titanium", 0),
        }
        heat = action_dict.get("heat", 0)
        if heat and heat > 0:
            payment["substitutes"] = {"heat": heat}
        payload = {"type": "play-card", "cardId": action_dict["cardId"], "payment": payment}
        for key in ("choiceIndex", "cardStorageTargets", "targetPlayerId", "selectedAmount"):
            if key in action_dict and action_dict[key] is not None:
                payload[key] = action_dict[key]
        return ACTION_MAP["play_card"], payload

    if action == "use_card_action":
        payload = {
            "type": "card-action",
            "cardId": action_dict["cardId"],
            "behaviorIndex": action_dict["behaviorIndex"],
        }
        for key in ("choiceIndex", "cardStorageTargets", "targetPlayerId",
                     "sourceCardForInput", "selectedAmount"):
            if key in action_dict and action_dict[key] is not None:
                payload[key] = action_dict[key]
        credits = action_dict.get("credits")
        steel = action_dict.get("steel")
        titanium = action_dict.get("titanium")
        if credits is not None or steel is not None or titanium is not None:
            payload["payment"] = {
                "credits": credits or 0,
                "steel": steel or 0,
                "titanium": titanium or 0,
            }
        return ACTION_MAP["use_card_action"], payload

    if action == "select_tile":
        q, r, s = action_dict["q"], action_dict["r"], action_dict["s"]
        return ACTION_MAP["select_tile"], {"hex": f"{q},{r},{s}"}

    if action == "select_starting_choices":
        return ACTION_MAP["select_starting_choices"], {
            "corporationId": action_dict["corporationId"],
            "preludeIds": action_dict.get("preludeIds", []),
            "cardIds": action_dict.get("cardIds", []),
        }

    if action == "claim_milestone":
        return ACTION_MAP["claim_milestone"], {"milestoneType": action_dict["milestoneType"]}

    if action == "fund_award":
        return ACTION_MAP["fund_award"], {"awardType": action_dict["awardType"]}

    if action == "skip_action":
        return ACTION_MAP["skip_action"], {}

    raise ValueError(f"Unknown action: {action}")


def _build_confirm_payload(confirm_action, action_dict):
    if confirm_action == "select":
        return {"selectedCardIds": action_dict.get("cardIds", [])}
    if confirm_action == "production":
        return {"cardIds": action_dict.get("cardIds", [])}
    if confirm_action == "draw":
        return {
            "cardsToTake": action_dict.get("cardsToTake", []),
            "cardsToBuy": action_dict.get("cardsToBuy", []),
        }
    if confirm_action == "discard":
        return {"cardsToDiscard": action_dict.get("cardsToDiscard", [])}
    if confirm_action == "behavior-choice":
        payload = {"choiceIndex": action_dict.get("choiceIndex", 0)}
        if "cardStorageTargets" in action_dict:
            payload["cardStorageTargets"] = action_dict["cardStorageTargets"]
        return payload
    return {}


# ---------------------------------------------------------------------------
# State Summarizer
# ---------------------------------------------------------------------------

def summarize_game_state(game, my_player_id):
    if not game:
        return "No game state available."

    sections = []
    sections.append(_format_game_info(game, my_player_id))
    sections.append(_format_global_params(game.get("globalParameters", {})))

    player = game.get("currentPlayer")
    if player:
        sections.append(_format_pending_actions(player))
        sections.append(_format_player_status(player))
        sections.append(_format_hand(player.get("cards", [])))
        sections.append(_format_card_actions(player.get("actions", [])))
        sections.append(_format_standard_projects(player.get("standardProjects", [])))
        sections.append(_format_milestones(player.get("milestones", [])))
        sections.append(_format_awards(player.get("awards", [])))

    sections.append(_format_opponents(game.get("otherPlayers", [])))
    sections.append(_format_board(game.get("board", {}).get("tiles", [])))

    final_scores = game.get("finalScores")
    if final_scores:
        sections.append(_format_final_scores(game))

    return "\n\n".join(s for s in sections if s)


def _find_player_name(game, player_id):
    cp = game.get("currentPlayer")
    if cp and cp.get("id") == player_id:
        return cp.get("name", player_id)
    for o in game.get("otherPlayers", []):
        if o.get("id") == player_id:
            return o.get("name", player_id)
    return player_id


def _format_game_info(game, my_player_id):
    current_turn = game.get("currentTurn")
    if current_turn:
        turn_info = "YOUR TURN" if current_turn == my_player_id else f"Waiting for {_find_player_name(game, current_turn)}"
    else:
        turn_info = "N/A"
    turn_order = game.get("turnOrder", [])
    order_names = " -> ".join(_find_player_name(game, pid) for pid in turn_order)
    return "\n".join([
        "=== GAME INFO ===",
        f"Game ID: {game.get('id', '?')}",
        f"Phase: {game.get('currentPhase', '?')}",
        f"Status: {game.get('status', '?')}",
        f"Generation: {game.get('generation', '?')}",
        f"Turn: {turn_info}",
        f"Players: {len(turn_order)}",
        f"Turn order: {order_names}",
    ])


def _format_global_params(gp):
    return "\n".join([
        "=== GLOBAL PARAMETERS ===",
        f"Temperature: {gp.get('temperature', '?')}C (target: 8C)",
        f"Oxygen: {gp.get('oxygen', '?')}% (target: 14%)",
        f"Oceans: {gp.get('oceans', '?')}/9",
        f"Venus: {gp.get('venus', '?')}% (target: 30%)",
    ])


def _format_pending_actions(player):
    parts = []

    pts = player.get("pendingTileSelection")
    if pts:
        parts.append("\n".join([
            f"TILE PLACEMENT REQUIRED: Place a {pts.get('tileType', '?')} tile",
            f"Source: {pts.get('source', '?')}",
            f"Available hexes: {', '.join(pts.get('availableHexes', []))}",
            "Use select_tile action with q, r, s coordinates.",
        ]))

    pcs = player.get("pendingCardSelection")
    if pcs:
        card_list = []
        for c in pcs.get("availableCards", []):
            cost = pcs.get("cardCosts", {}).get(c["id"], 0)
            reward = pcs.get("cardRewards", {}).get(c["id"], 0)
            info = f" (cost: {cost}M)" if cost > 0 else (f" (reward: {reward}M)" if reward > 0 else "")
            card_list.append(f"  - {c.get('name', '?')} [{c['id']}]{info}")
        parts.append("\n".join([
            f"CARD SELECTION REQUIRED (source: {pcs.get('source', '?')})",
            f"Select {pcs.get('minCards', '?')}-{pcs.get('maxCards', '?')} cards:",
            *card_list,
            'Use confirm_cards action with confirmAction="select" and cardIds.',
        ]))

    pcd = player.get("pendingCardDrawSelection")
    if pcd:
        card_list = [f"  - {c.get('name', '?')} [{c['id']}]: {c.get('description', '')}"
                     for c in pcd.get("availableCards", [])]
        buy_cost = pcd.get("cardBuyCost", 0)
        cost_str = f" ({buy_cost}M each)" if buy_cost > 0 else ""
        parts.append("\n".join([
            f"CARD DRAW SELECTION (source: {pcd.get('source', '?')})",
            f"Free takes: {pcd.get('freeTakeCount', 0)}, Max buy: {pcd.get('maxBuyCount', 0)}{cost_str}",
            *card_list,
            'Use confirm_cards action with confirmAction="draw" and cardsToTake/cardsToBuy.',
        ]))

    pcdisc = player.get("pendingCardDiscardSelection")
    if pcdisc:
        parts.append("\n".join([
            f"CARD DISCARD REQUIRED (source: {pcdisc.get('source', '?')})",
            f"Discard {pcdisc.get('minCards', '?')}-{pcdisc.get('maxCards', '?')} cards from hand.",
            'Use confirm_cards action with confirmAction="discard" and cardsToDiscard.',
        ]))

    pbc = player.get("pendingBehaviorChoiceSelection")
    if pbc:
        choice_list = []
        for i, c in enumerate(pbc.get("choices", [])):
            desc = _format_behavior_brief(c.get("inputs"), c.get("outputs"))
            avail = "" if c.get("available", True) else " [UNAVAILABLE]"
            choice_list.append(f"  {i}: {desc}{avail}")
        parts.append("\n".join([
            f"BEHAVIOR CHOICE REQUIRED (source: {pbc.get('source', '?')})",
            *choice_list,
            'Use confirm_cards action with confirmAction="behavior-choice" and choiceIndex.',
        ]))

    ffa = player.get("forcedFirstAction")
    if ffa and not ffa.get("completed", False):
        parts.append("\n".join([
            f"FORCED FIRST ACTION: {ffa.get('description', '?')}",
            f"Action type: {ffa.get('actionType', '?')}",
            f"Corporation: {ffa.get('corporationId', '?')}",
        ]))

    if player.get("selectCorporationPhase") or player.get("selectStartingCardsPhase") or player.get("selectPreludeCardsPhase"):
        parts.append(_format_starting_selection(player))

    pp = player.get("productionPhase")
    if pp and not pp.get("selectionComplete", False):
        cards = [f"  - {c.get('name', '?')} [{c['id']}]" for c in pp.get("availableCards", [])]
        parts.append("\n".join([
            "PRODUCTION PHASE - Select cards to buy:",
            *(cards or ["  (no cards available)"]),
            'Use confirm_cards action with confirmAction="production" and cardIds.',
        ]))

    if not parts:
        return ""
    return ">>> PENDING ACTIONS (must resolve) <<<\n" + "\n\n".join(parts)


def _format_starting_selection(player):
    parts = ["STARTING SELECTION REQUIRED"]
    scp = player.get("selectCorporationPhase")
    if scp:
        corps = [f"  - {c.get('name', '?')} [{c['id']}]: {c.get('description', '')}"
                 for c in scp.get("availableCorporations", [])]
        parts.append("Corporations:\n" + "\n".join(corps))
    spp = player.get("selectPreludeCardsPhase")
    if spp:
        preludes = [f"  - {c.get('name', '?')} [{c['id']}]: {c.get('description', '')}"
                    for c in spp.get("availablePreludes", [])]
        parts.append(f"Preludes (pick {spp.get('maxSelectable', '?')}):\n" + "\n".join(preludes))
    sscp = player.get("selectStartingCardsPhase")
    if sscp:
        cards = []
        for c in sscp.get("availableCards", []):
            tags = ", ".join(c.get("tags", []) or [])
            cards.append(f"  - {c.get('name', '?')} [{c['id']}] ({c.get('cost', '?')}M) [{c.get('type', '?')}] {tags}: {c.get('description', '')}")
        parts.append("Starting cards (pick any to buy at 3M each):\n" + "\n".join(cards))
    parts.append("Use select_starting_choices action with corporationId, preludeIds, and cardIds.")
    return "\n".join(parts)


def _format_player_status(player):
    r = player.get("resources", {})
    p = player.get("production", {})

    def fmt_prod(val):
        return f"+{val}" if val >= 0 else str(val)

    lines = [
        "=== YOUR STATUS ===",
        f"Name: {player.get('name', '?')} | Corporation: {(player.get('corporation') or {}).get('name', 'None')} | TR: {player.get('terraformRating', '?')}",
        f"Status: {player.get('status', '?')} | Actions remaining: {player.get('availableActions', '?')} | Passed: {player.get('passed', '?')}",
        "",
        "Resources (amount / production):",
        f"  Credits:  {r.get('credits', 0)} / {fmt_prod(p.get('credits', 0))}",
        f"  Steel:    {r.get('steel', 0)} / {fmt_prod(p.get('steel', 0))}",
        f"  Titanium: {r.get('titanium', 0)} / {fmt_prod(p.get('titanium', 0))}",
        f"  Plants:   {r.get('plants', 0)} / {fmt_prod(p.get('plants', 0))}",
        f"  Energy:   {r.get('energy', 0)} / {fmt_prod(p.get('energy', 0))}",
        f"  Heat:     {r.get('heat', 0)} / {fmt_prod(p.get('heat', 0))}",
    ]

    played = player.get("playedCards", [])
    if played:
        names = ", ".join(c.get("name", "?") for c in played)
        lines.extend(["", f"Played cards ({len(played)}): {names}"])

    storage = {k: v for k, v in player.get("resourceStorage", {}).items() if v > 0}
    if storage:
        lines.append(f"Resource storage: {', '.join(f'{k}: {v}' for k, v in storage.items())}")

    subs = player.get("paymentSubstitutes", [])
    if subs:
        sub_strs = [f"{s.get('resourceType', '?')} ({s.get('conversionRate', '?')}:1)" for s in subs]
        lines.append(f"Payment substitutes: {', '.join(sub_strs)}")

    effects = player.get("effects", [])
    if effects:
        lines.append(f"Active effects: {', '.join(e.get('cardName', '?') for e in effects)}")

    return "\n".join(lines)


def _format_hand(cards):
    if not cards:
        return "=== HAND (0 cards) ===\n(empty)"
    header = f"=== HAND ({len(cards)} cards) ==="
    card_lines = []
    for c in cards:
        avail = "PLAYABLE" if c.get("available") else "BLOCKED"
        err_info = ""
        if not c.get("available") and c.get("errors"):
            err_info = f" ({'; '.join(e.get('message', '') for e in c['errors'])})"
        tags = f" [{', '.join(c.get('tags', []))}]" if c.get("tags") else ""
        eff_cost = c.get("effectiveCost", c.get("cost", 0))
        cost = c.get("cost", 0)
        discount = f" (discounted from {cost})" if eff_cost < cost else ""
        line = f"  - {c.get('name', '?')} [{c['id']}] | {eff_cost}M{discount} | {c.get('type', '?')}{tags} | {avail}{err_info}"
        if c.get("description"):
            line += f"\n    {c['description']}"
        card_lines.append(line)
    return header + "\n" + "\n".join(card_lines)


def _format_card_actions(actions):
    if not actions:
        return ""
    header = "=== CARD ACTIONS ==="
    lines = []
    for a in actions:
        avail = "AVAILABLE" if a.get("available") else "BLOCKED"
        err_info = ""
        if not a.get("available") and a.get("errors"):
            err_info = f" ({'; '.join(e.get('message', '') for e in a['errors'])})"
        used = f" [used {a['timesUsedThisTurn']}x this turn]" if a.get("timesUsedThisTurn", 0) > 0 else ""
        beh = a.get("behavior", {})
        desc = _format_behavior_brief(beh.get("inputs"), beh.get("outputs"))
        line = f"  - {a.get('cardName', '?')} [{a.get('cardId', '?')}] behavior#{a.get('behaviorIndex', '?')} | {avail}{used}{err_info}"
        if desc:
            line += f"\n    {desc}"
        lines.append(line)
    return header + "\n" + "\n".join(lines)


def _format_standard_projects(projects):
    if not projects:
        return ""
    header = "=== STANDARD PROJECTS ==="
    lines = []
    for p in projects:
        avail = "AVAILABLE" if p.get("available") else "BLOCKED"
        err_info = ""
        if not p.get("available") and p.get("errors"):
            err_info = f" ({p['errors'][0].get('message', '')})"
        cost_parts = [f"{v} {k}" for k, v in p.get("effectiveCost", {}).items()]
        cost_str = ", ".join(cost_parts)
        lines.append(f"  - {p.get('projectType', '?')} | {cost_str} | {avail}{err_info}")
    return header + "\n" + "\n".join(lines)


def _format_milestones(milestones):
    if not milestones:
        return ""
    header = "=== MILESTONES ==="
    lines = []
    for m in milestones:
        if m.get("isClaimed"):
            status = f"CLAIMED by {m.get('claimedBy', '?')}"
        elif m.get("available"):
            status = f"CLAIMABLE ({m.get('claimCost', '?')}M)"
        else:
            status = "NOT YET"
        lines.append(f"  - {m.get('name', '?')}: {m.get('description', '')} | Progress: {m.get('progress', '?')}/{m.get('required', '?')} | {status}")
    return header + "\n" + "\n".join(lines)


def _format_awards(awards):
    if not awards:
        return ""
    header = "=== AWARDS ==="
    lines = []
    for a in awards:
        if a.get("isFunded"):
            status = f"FUNDED by {a.get('fundedBy', '?')}"
        elif a.get("available"):
            status = f"FUNDABLE ({a.get('fundingCost', '?')}M)"
        else:
            status = "NOT AVAILABLE"
        lines.append(f"  - {a.get('name', '?')}: {a.get('description', '')} | {status}")
    return header + "\n" + "\n".join(lines)


def _format_opponents(others):
    if not others:
        return ""
    header = "=== OPPONENTS ==="
    lines = []
    for o in others:
        r = o.get("resources", {})
        corp = (o.get("corporation") or {}).get("name", "None")
        lines.append("\n".join([
            f"  {o.get('name', '?')} ({corp}) | TR: {o.get('terraformRating', '?')} | Status: {o.get('status', '?')} | Passed: {o.get('passed', '?')}",
            f"    Resources: {r.get('credits', 0)}M, {r.get('steel', 0)} steel, {r.get('titanium', 0)} ti, {r.get('plants', 0)} plants, {r.get('energy', 0)} energy, {r.get('heat', 0)} heat",
            f"    Cards in hand: {o.get('handCardCount', 0)} | Played: {len(o.get('playedCards', []))} cards",
        ]))
    return header + "\n" + "\n".join(lines)


def _format_board(tiles):
    occupied = [t for t in tiles if t.get("occupiedBy")]
    if not occupied:
        return ""
    header = "=== BOARD ==="
    lines = []
    for t in occupied:
        c = t.get("coordinates", {})
        coord = f"({c.get('q', '?')},{c.get('r', '?')},{c.get('s', '?')})"
        occ = t.get("occupiedBy", {})
        occ_str = f" | {occ.get('type', '?')}"
        if t.get("ownerId"):
            occ_str += f" (owner: {t['ownerId']})"
        name = f" {t['displayName']}" if t.get("displayName") else ""
        bonuses = t.get("bonuses", [])
        bonus_parts = [f"{b.get('amount', 0)}x {b.get('type', '?')}" for b in bonuses] if bonuses else []
        bonus_str = f" | bonuses: {', '.join(bonus_parts)}" if bonus_parts else ""
        lines.append(f"  {coord}{name}{occ_str}{bonus_str}")
    return header + "\n" + "\n".join(lines)


def _format_final_scores(game):
    scores = game.get("finalScores", [])
    if not scores:
        return ""
    header = "=== FINAL SCORES ==="
    lines = []
    for s in scores:
        vp = s.get("vpBreakdown", {})
        winner = " (WINNER)" if s.get("isWinner") else ""
        lines.append("\n".join([
            f"  #{s.get('placement', '?')} {s.get('playerName', '?')}{winner}: {vp.get('totalVP', '?')} VP",
            f"    TR: {vp.get('terraformRating', 0)}, Cards: {vp.get('cardVP', 0)}, Greenery: {vp.get('greeneryVP', 0)}, City: {vp.get('cityVP', 0)}, Milestones: {vp.get('milestoneVP', 0)}, Awards: {vp.get('awardVP', 0)}",
        ]))
    return header + "\n" + "\n".join(lines)


def _format_behavior_brief(inputs, outputs):
    parts = []
    if inputs:
        parts.append("Costs: " + ", ".join(f"{i.get('amount', '?')} {i.get('type', '?')}" for i in inputs))
    if outputs:
        parts.append("Gives: " + ", ".join(f"{o.get('amount', '?')} {o.get('type', '?')}" for o in outputs))
    return " -> ".join(parts)


# ---------------------------------------------------------------------------
# GameConnection (WebSocket client)
# ---------------------------------------------------------------------------

class GameConnection:
    def __init__(self, server_url, game_id, player_name, player_id=None):
        self.server_url = server_url
        self.game_id = game_id
        self.player_name = player_name
        self.player_id = player_id
        self.my_player_id = player_id
        self.game_state = None
        self.ws = None
        self._turn_event = asyncio.Event()
        self._action_result = None
        self._action_event = asyncio.Event()

    async def connect(self):
        self.ws = await websockets.connect(self.server_url)
        payload = {"playerName": self.player_name, "gameId": self.game_id}
        if self.player_id:
            payload["playerId"] = self.player_id
        msg = json.dumps({"type": "player-connect", "payload": payload, "gameId": self.game_id})
        await self.ws.send(msg)
        print(f"[tm] Sent player-connect for '{self.player_name}' to game {self.game_id}")

    async def listen(self):
        try:
            async for raw in self.ws:
                try:
                    message = json.loads(raw)
                except json.JSONDecodeError:
                    continue

                msg_type = message.get("type", "")
                payload = message.get("payload", {})

                if msg_type == "game-updated":
                    game = payload.get("game") or payload
                    if not self.my_player_id:
                        cp = game.get("currentPlayer")
                        if cp and cp.get("id"):
                            self.my_player_id = cp["id"]
                            print(f"[tm] Resolved player ID: {self.my_player_id}")
                    self._update_state(game)
                    self._action_result = (True, game)
                    self._action_event.set()

                elif msg_type == "player-connected":
                    pid = payload.get("playerId") or payload.get("playerID")
                    if pid:
                        self.my_player_id = pid
                    game = payload.get("game")
                    if game:
                        self._update_state(game)
                    self._action_result = (True, game)
                    self._action_event.set()
                    print(f"[tm] Connected as player {self.my_player_id}")

                elif msg_type == "full-state":
                    pid = payload.get("playerId") or payload.get("playerID")
                    if pid:
                        self.my_player_id = pid
                    game = payload.get("game")
                    if game:
                        self._update_state(game)
                    self._action_result = (True, game)
                    self._action_event.set()

                elif msg_type == "error":
                    err_msg = payload.get("message") or payload.get("error") or "Unknown error"
                    print(f"[tm] Server error: {err_msg}")
                    self._action_result = (False, err_msg)
                    self._action_event.set()

        except websockets.exceptions.ConnectionClosed:
            print("[tm] WebSocket connection closed")

    def _update_state(self, game):
        self.game_state = game
        if self._is_my_turn():
            self._turn_event.set()

    def _is_my_turn(self):
        if not self.game_state or not self.my_player_id:
            return False
        player = self.game_state.get("currentPlayer")
        if not player:
            return False

        if (self.game_state.get("currentPhase") == GAME_PHASE_ACTION and
                self.game_state.get("currentTurn") == self.my_player_id):
            return True

        status = player.get("status")
        if status in (PLAYER_STATUS_ACTIVE, PLAYER_STATUS_SELECTING_STARTING, PLAYER_STATUS_SELECTING_PRODUCTION):
            return True

        for key in ("pendingTileSelection", "pendingCardSelection", "pendingCardDrawSelection",
                     "pendingCardDiscardSelection", "pendingBehaviorChoiceSelection"):
            if player.get(key):
                return True

        ffa = player.get("forcedFirstAction")
        if ffa and not ffa.get("completed", False):
            return True

        if self.game_state.get("currentPhase") == GAME_PHASE_STARTING_SELECTION:
            return True

        if (self.game_state.get("currentPhase") == GAME_PHASE_PRODUCTION and
                player.get("productionPhase") and
                not player["productionPhase"].get("selectionComplete", False)):
            return True

        return False

    async def wait_for_turn(self, timeout=300):
        if self._is_my_turn():
            return
        self._turn_event.clear()
        await asyncio.wait_for(self._turn_event.wait(), timeout=timeout)

    async def send_action(self, ws_type, payload):
        self._action_event.clear()
        self._action_result = None
        msg = json.dumps({"type": ws_type, "payload": payload, "gameId": self.game_id})
        await self.ws.send(msg)
        try:
            await asyncio.wait_for(self._action_event.wait(), timeout=15)
        except asyncio.TimeoutError:
            return False, "Timeout waiting for server response"
        return self._action_result or (False, "No result")

    def is_game_complete(self):
        return self.game_state and self.game_state.get("currentPhase") == GAME_PHASE_COMPLETE

    def clear_turn_signal(self):
        """Clear the turn event to prevent spin loops after processing a turn."""
        self._turn_event.clear()


# ---------------------------------------------------------------------------
# Claude Decision Engine
# ---------------------------------------------------------------------------

class ClaudeDecisionEngine:
    def __init__(self, model=None):
        self.model = model

    async def decide(self, state_summary, error_context=None):
        prompt = SYSTEM_PROMPT + "\n\n--- CURRENT GAME STATE ---\n" + state_summary
        if error_context:
            prompt += f"\n\n--- PREVIOUS ACTION FAILED ---\n{error_context}\nPlease try a different action or fix the parameters."
        return await asyncio.to_thread(self._call_claude, prompt)

    def _call_claude(self, prompt):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False) as f:
            f.write(prompt)
            f.flush()
            tmp_path = f.name
        try:
            cmd = ["claude", "-p", "--output-format", "json"]
            if self.model:
                cmd.extend(["--model", self.model])
            # Filter out CLAUDECODE env var to allow nested CLI invocation
            env = {k: v for k, v in os.environ.items() if k != "CLAUDECODE"}
            with open(tmp_path) as stdin_file:
                result = subprocess.run(
                    cmd, stdin=stdin_file, capture_output=True, text=True,
                    timeout=120, env=env,
                )
            if result.returncode != 0:
                print(f"[tm] Claude CLI error (rc={result.returncode}): {result.stderr[:500]}")
                return {"action": "skip_action"}
            return self._parse_response(result.stdout)
        except subprocess.TimeoutExpired:
            print("[tm] Claude CLI timed out")
            return {"action": "skip_action"}
        finally:
            os.unlink(tmp_path)

    def _parse_response(self, text):
        # Parse wrapper from --output-format json
        try:
            wrapper = json.loads(text)
            if isinstance(wrapper, dict) and "result" in wrapper:
                text = wrapper["result"]
        except (json.JSONDecodeError, TypeError):
            pass
        if isinstance(text, dict):
            return text
        # Try direct JSON parse
        try:
            return json.loads(text)
        except (json.JSONDecodeError, TypeError):
            pass
        # Try markdown code block
        match = re.search(r"```(?:json)?\s*\n?(.*?)\n?```", text, re.DOTALL)
        if match:
            try:
                return json.loads(match.group(1).strip())
            except json.JSONDecodeError:
                pass
        # Try finding JSON object in text
        match = re.search(r"\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}", text, re.DOTALL)
        if match:
            try:
                return json.loads(match.group(0))
            except json.JSONDecodeError:
                pass
        print(f"[tm] Failed to parse Claude response: {text[:300]}")
        return {"action": "skip_action"}


# ---------------------------------------------------------------------------
# Game listing & auto-join
# ---------------------------------------------------------------------------

PROD_WS_URL = "wss://terraforming-mars.rackaracka.net/ws"
LOCAL_WS_URL = "ws://localhost:3001/ws"


def _ws_url_to_http(ws_url):
    return re.sub(r"^ws(s?)://", r"http\1://", ws_url).rstrip("/ws").rstrip("/")


def fetch_games(server_url):
    base = _ws_url_to_http(server_url)
    url = f"{base}/api/v1/games"
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "tm-bot/1.0"})
        with urllib.request.urlopen(req, timeout=5) as resp:
            data = json.loads(resp.read())
            return data.get("games", [])
    except (urllib.error.URLError, json.JSONDecodeError, OSError) as e:
        print(f"[tm] Failed to fetch games from {url}: {e}")
        return None


def _find_player_in_game(game, player_name):
    game_id = game.get("id")
    cp = game.get("currentPlayer")
    if cp and cp.get("name", "").lower() == player_name.lower():
        return game_id, cp.get("id")
    for op in game.get("otherPlayers", []):
        if op.get("name", "").lower() == player_name.lower():
            return game_id, op.get("id")
    return game_id, None


def pick_game(server_url, player_name):
    games = fetch_games(server_url)
    if games is None:
        print("[tm] Could not reach server. Is the backend running?")
        sys.exit(1)
    if not games:
        print("[tm] No games found on server.")
        sys.exit(1)

    active = [g for g in games if g.get("status") == "active"]
    if len(active) == 1:
        gid, pid = _find_player_in_game(active[0], player_name)
        print(f"[tm] Auto-joining the only active game: {gid}")
        if pid:
            print(f"[tm] Found existing player '{player_name}' -> reconnecting as {pid}")
        return gid, pid

    status_order = {"active": 0, "lobby": 1, "completed": 2}
    games.sort(key=lambda g: status_order.get(g.get("status", ""), 9))

    print(f"\n{'#':<4} {'Status':<12} {'Gen':<5} {'Phase':<28} {'Players':<40} {'Game ID'}")
    print("-" * 120)

    for i, g in enumerate(games, 1):
        players = []
        cp = g.get("currentPlayer")
        if cp:
            players.append(cp.get("name", "?"))
        for op in g.get("otherPlayers", []):
            players.append(op.get("name", "?"))
        player_str = ", ".join(players) if players else "(none)"
        print(f"{i:<4} {g.get('status', '?'):<12} {g.get('generation', '?'):<5} {g.get('currentPhase', '?'):<28} {player_str:<40} {g.get('id', '?')}")

    print()
    while True:
        try:
            choice = input("Select game number (or 'q' to quit): ").strip()
        except (EOFError, KeyboardInterrupt):
            print()
            sys.exit(0)
        if choice.lower() == "q":
            sys.exit(0)
        try:
            idx = int(choice) - 1
            if 0 <= idx < len(games):
                return _find_player_in_game(games[idx], player_name)
            print(f"  Enter a number between 1 and {len(games)}")
        except ValueError:
            print("  Enter a number or 'q'")


# ---------------------------------------------------------------------------
# Main Loop
# ---------------------------------------------------------------------------

async def main(game_id, player_name, server_url, player_id=None, model=None):
    conn = GameConnection(server_url, game_id, player_name, player_id)
    engine = ClaudeDecisionEngine(model)

    await conn.connect()
    listener_task = asyncio.create_task(conn.listen())

    # Wait for initial state
    try:
        await asyncio.wait_for(conn.wait_for_turn(timeout=10), timeout=10)
    except (asyncio.TimeoutError, Exception):
        pass

    print(f"[tm] Player ID: {conn.my_player_id}")
    phase = conn.game_state.get("currentPhase", "?") if conn.game_state else "?"
    gen = conn.game_state.get("generation", "?") if conn.game_state else "?"
    print(f"[tm] Phase: {phase} | Generation: {gen}")
    print("[tm] Autonomous bot running. Press Ctrl+C to stop.")

    try:
        while not conn.is_game_complete():
            # Wait for our turn
            print("[tm] Waiting for turn...")
            conn.clear_turn_signal()
            try:
                await conn.wait_for_turn(timeout=600)
            except asyncio.TimeoutError:
                print("[tm] Timeout waiting for turn, retrying...")
                continue

            if conn.is_game_complete():
                break

            summary = summarize_game_state(conn.game_state, conn.my_player_id)
            gen = conn.game_state.get("generation", "?")
            phase = conn.game_state.get("currentPhase", "?")
            print(f"[tm] === MY TURN === Gen {gen} | Phase: {phase}")

            # Retry loop
            error_context = None
            for attempt in range(3):
                if attempt > 0:
                    print(f"[tm] Retry {attempt}/3...")
                    # Re-summarize in case state changed
                    summary = summarize_game_state(conn.game_state, conn.my_player_id)

                print("[tm] Asking Claude for decision...")
                action = await engine.decide(summary, error_context)
                print(f"[tm] Claude says: {json.dumps(action)}")

                try:
                    ws_type, payload = build_ws_message(action)
                except (ValueError, KeyError) as e:
                    print(f"[tm] Invalid action format: {e}")
                    error_context = f"Invalid action format: {e}. The action JSON was: {json.dumps(action)}"
                    continue

                success, result = await conn.send_action(ws_type, payload)

                if success:
                    print(f"[tm] Action succeeded")
                    break
                else:
                    err_msg = str(result)
                    print(f"[tm] Action failed: {err_msg}")
                    # If "not your turn", don't retry - wait for next turn
                    if "not your turn" in err_msg.lower() or "not active" in err_msg.lower():
                        print("[tm] Not our turn anymore, waiting...")
                        break
                    error_context = f"Action {json.dumps(action)} failed with error: {err_msg}"
            else:
                # All retries exhausted - skip
                print("[tm] All retries exhausted, skipping turn")
                await conn.send_action(ACTION_MAP["skip_action"], {})

    except KeyboardInterrupt:
        print("\n[tm] Shutting down")
    finally:
        listener_task.cancel()
        if conn.ws:
            await conn.ws.close()

    # Print final scores
    if conn.is_game_complete():
        print("\n[tm] === GAME COMPLETE ===")
        print(summarize_game_state(conn.game_state, conn.my_player_id))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Terraforming Mars Autonomous Bot")
    parser.add_argument("--game-id", help="Game ID to join (omit to auto-join or pick)")
    parser.add_argument("--player-name", default="Claude", help="Display name (default: Claude)")
    parser.add_argument("--local", action="store_true", help="Use local server (default: prod)")
    parser.add_argument("--server-url", help="Custom WebSocket URL (overrides --local)")
    parser.add_argument("--player-id", help="Existing player ID for reconnection")
    parser.add_argument("--model", help="Claude model to use (e.g. sonnet, opus)")
    args = parser.parse_args()

    server_url = args.server_url or (LOCAL_WS_URL if args.local else PROD_WS_URL)

    game_id = args.game_id
    player_id = args.player_id
    if not game_id:
        game_id, auto_pid = pick_game(server_url, args.player_name)
        if not player_id and auto_pid:
            player_id = auto_pid

    print(f"[tm] Game: {game_id} | Player: {args.player_name} | Server: {server_url}")
    if player_id:
        print(f"[tm] Reconnecting as: {player_id}")
    if args.model:
        print(f"[tm] Model: {args.model}")

    asyncio.run(main(game_id, args.player_name, server_url, player_id, args.model))
