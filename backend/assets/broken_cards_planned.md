# Broken Cards Fix Plan

> **CRITICAL: Before applying ANY fix from this plan, verify the requested change against the card's `description` field in `terraforming_mars_cards.json`. The card description is the source of truth. It is NOT allowed to change a card's description to match the planned fix. If the planned fix contradicts the card description, mark the group task as `[x]` (done) but add a note to the "Description Mismatches" section at the bottom of this file explaining the discrepancy.**

66 broken cards grouped by the type of fix required. Groups are ordered from simplest (JSON-only) to most complex (new backend features + new tile types). Each group targets a **generic capability** so the fix applies to all similar cards.

---

## Feasibility Assessment

All groups are **confirmed implementable** within the current architecture, with one exception:

- **Groups 1-20, 22-25**: Fully feasible. The existing systems (BehaviorApplier, state calculator, event bus, tile selection queue, TileCompletionCallback, payment system, selector matching) provide all the extension points needed.
- **Group 21 (Mars University + Sponsored Academies)**: Feasible. Uses the same **pending selection pattern** as card draw and tile placement — when a behavior has `card-discard` inputs, create a `PendingCardDiscardSelection` instead of auto-applying outputs. The passive effect subscriber creates the selection, the player responds, and a confirmation handler completes the effect. Same infrastructure, just wired into the passive path.

New tile types use **purple placeholder hexes** with text labels instead of 3D models — easily upgradeable later.

---

## TIER 1: JSON-Only Data Fixes

No code changes needed. Fix values in `terraforming_mars_cards.json`.

---

### [x] Group 1: Wrong `per` Divisor Values

**Root cause:** `per.amount` is wrong, causing scaled outputs to give wrong amounts.

| # | Card | ID | Current | Correct | Notes |
|---|------|----|---------|---------|-------|
| 26 | Worms | 130 | `per.amount: 1` | `per.amount: 2` | Should be "plant production per **2** microbe tags" |
| 41 | Medical Lab | 207 | `per.amount: 1` | `per.amount: 2` | Should be "credit production per **2** building tags" |
| 16 | Physics Complex | 095 | VP: `amount: 2, per.amount: 2` | VP: `amount: 2, per.amount: 1` | Should be "2 VP per science resource", not "2 VP per 2 science resources" |

**Test:** Verify `countPerCondition` produces correct scaled amounts with new divisor values.

---

### [x] Group 2: Wrong Target on Per-Condition

**Root cause:** `per.target` points to wrong player, causing counts from wrong source.

| # | Card | ID | Current | Correct | Notes |
|---|------|----|---------|---------|-------|
| 19 | Toll Station | 099 | `per.target: "self-player"` | `per.target: "any-player"` | Should count **opponents'** space tags, not own |

**Test:** Verify `countPerCondition` counts opponent tags when `target: "any-player"`.

---

### [x] Group 3: Missing Trigger Conditions on Passive Effects

**Root cause:** Behaviors have `auto` trigger with no `condition`, so they fire immediately on play instead of as passive effects. Need to add the correct `condition` so the event system handles them.

| # | Card | ID | Trigger Condition Needed |
|---|------|----|------------------------|
| 20 | Media Group | 109 | `{type: "card-played", target: "self-player", selectors: [{cardTypes: ["event"]}]}` |
| 65 | Venusian Animals | 259 | `{type: "tag-played", target: "self-player", selectors: [{tags: ["science"]}]}` |

**Already supported:** `card-played` and `tag-played` triggers exist in `passive_effect_subscriber.go`. Standard Technology (#31) also needs this fix but `standard-project-played` trigger already exists.

| # | Card | ID | Trigger Condition Needed |
|---|------|----|------------------------|
| 31 | Standard Technology | 156 | `{type: "standard-project-played", target: "self-player"}` |

**Note on Media Group:** The first behavior (immediate `card-draw: 1`) should be split. The `auto` behavior should become the passive trigger, and the existing immediate draw needs to be a separate behavior or removed depending on the original card rules. The real card says "Effect: After you play an event card, you gain 3 M€." - so the entire behavior should be passive (no immediate effect on play).

**Note on Venusian Animals:** The first behavior currently gives 1 animal immediately on play. The real card says "Effect: When you play a science tag, including this, add 1 animal to this card." This should be a passive trigger that also fires on self-play (the `tag-played` event fires for each tag on the played card, including itself).

**Test:** Verify passive effects fire on the correct events and don't fire immediately on play.

---

### [x] Group 4: Missing Outputs in Card JSON

**Root cause:** Card JSON is missing outputs that the card description requires.

| # | Card | ID | Missing Output | Notes |
|---|------|----|---------------|-------|
| 45 | Air-Scrapping Expedition | 215 | `{type: "venus", amount: 1, target: "none"}` | Description says "Raise Venus 1 step" |
| 47 | Comet for Venus | 218 | `{type: "venus", amount: 1, target: "none"}` | Description says "Raise Venus 1 step" |
| 54 | Hydrogen To Venus | 231 | `{type: "venus", amount: 1, target: "none"}` | Description says "Raise Venus 1 step" — floater per-condition output exists but venus raise is missing |
| 240 | Neutralizer Factory | 240 | Full behavior: `{triggers: [{type: "auto"}], outputs: [{type: "venus", amount: 1}]}` | Entire behavior array is `null` |
| 53 | Gyropolis | 230 | Second per-condition for Venus tags | Only counts Earth tags, should count both Venus AND Earth |

**Gyropolis fix detail:** Currently has one per output for Earth tags. Need to add a second per output for Venus tags:
```json
{
  "type": "credit-production", "amount": 1, "target": "self-player",
  "per": {"type": "tag", "amount": 1, "target": "self-player", "tag": "venus"}
}
```

**Test:** Verify Venus global parameter increments, verify Gyropolis counts both tag types.

---

### [x] Group 5: Missing City Placement Outputs

**Root cause:** Cards with `city` tag describe placing a city tile but have no `city-placement` output.

| # | Card | ID | Placement Details |
|---|------|----|------------------|
| 3 | Phobos Space Haven | 021 | City on Phobos (off-Mars, uses boardTag) |
| 49 | Dawn City | 220 | City on reserved area |
| 57 | Maxwell Base | 238 | City (add to first/auto behavior) |
| 61 | Stratopolis | 248 | City (add to first/auto behavior) |

**Fix:** Add `{type: "city-placement", amount: 1, target: "none"}` to auto behaviors. For Phobos Space Haven, add `tileRestrictions: {boardTags: ["phobos-space-haven"]}` (requires adding that board tag). For other off-Mars cities, add appropriate board tags.

**Note:** This depends on the board having special reserved areas defined with board tags. If Phobos/Dawn City board tags don't exist yet, they need to be added to the board definition as well.

**Test:** Verify city placement is queued when these cards are played.

---

### [x] Group 6: Missing Requirements in Card JSON

| # | Card | ID | Missing Requirement |
|---|------|----|-------------------|
| 64 | Terraforming Contract | 252 | `{type: "tr", min: 25}` |

**Test:** Verify card is unplayable when TR < 25 and playable when TR >= 25.

---

### [x] Group 7: Fix Io Sulphur Research Card Draw

**Root cause:** Choices have wrong amounts and missing conditional logic.

| # | Card | ID | Issue |
|---|------|----|-------|
| 55 | Io Sulphur Research | 232 | Choice 1 draws 1, Choice 2 draws 2. Should be: draw 1 card normally, OR draw 3 cards if you have 3+ Venus tags |

**Fix:** Update choice amounts. Choice 2 needs `card-draw: 3`. The conditional "if you have 3+ Venus tags" for the second choice requires a new feature (see Group 20).

---

### [x] Group 8: Fix Sulphur-Eating Bacteria Missing Choice

| # | Card | ID | Issue |
|---|------|----|-------|
| 63 | Sulphur-Eating Bacteria | 251 | Only has "add 1 microbe" action. Missing second choice: "spend X microbes for 3X M€" |

**Fix:** Change manual behavior to have 2 choices:
- Choice 1: `outputs: [{type: "microbe", amount: 1, target: "self-card"}]`
- Choice 2: `inputs: [{type: "microbe", amount: X, target: "self-card"}], outputs: [{type: "credit", amount: 3X}]`

**Note:** This requires variable-amount input support (see Group 16). The player needs to choose HOW MANY microbes to spend.

---

### [x] Group 9: Fix Sponsored Academies

| # | Card | ID | Issue |
|---|------|----|-------|
| 60 | Sponsored Academies | 247 | Draw amount wrong (2 instead of 3). Missing: discard 1 card, opponents draw 1 |

**Fix:** Needs card discard support + opponent draw (see Group 21). Update `card-draw: 3`, add discard input, add opponent card-draw.

---

## TIER 2: Frontend Display Fixes

These require changes to the frontend BehaviorSection rendering system but no backend changes.

---

### [x] Group 10: Missing "Any" Red Tint on Per-Condition Icons

**Root cause:** When a `per` condition counts something across ALL players or from ALL cities (not just self), the icon should have a red/any-player glow to indicate it counts globally. The frontend `TriggeredEffectLayout` and `ImmediateResourceLayout` don't apply the "any" indicator to per-condition resource icons.

| # | Card | ID | Icon Needing Red Tint |
|---|------|----|----------------------|
| 17 | Greenhouses | 096 | city-tile (per counts all cities) |
| 34 | Energy Saving | 189 | city-tile (per counts all cities) |
| 25 | Zeppelins | 129 | city-tile (per counts Mars cities, location: mars) |
| 44 | Aerosport Tournament | 214 | city-tile (per counts all cities) |

**Generic fix:** In the BehaviorIcon/GameIcon rendering for `per` conditions, check if `per.target` is absent or not `"self-player"` (default counts all), and apply the red tint/glow CSS class. This also covers future cards with similar per-conditions.

**Also affected post-JSON-fix:**
- Toll Station (099) - after fixing target to "any-player"

**Test:** Visual - verify icons render with red tint when per counts globally.

---

### [x] Group 11: Minus/Negative Sign and Steal Display Issues

**Root cause:** Cards with negative `amount` values or `steal-*` targets don't clearly show the attack/removal nature. The frontend grouping of "plus" and "minus" rows has issues with how global parameters and tile placements are categorized.

| # | Card | ID | Issue |
|---|------|----|-------|
| 9 | Virus | 050 | Choice outputs with negative amounts look like giving resources instead of removing |
| 10 | Mining Expedition | 063 | `-2 plant` with `any-player` target needs clear minus indicator |
| 22 | Sabotage | 121 | Choice outputs with negative amounts need minus indicators |
| 23 | Hired Raiders | 124 | `steal-any-player` target doesn't convey "stealing" visually |
| 33 | Flooding | 188 | `-4 credit` needs minus sign; `ocean-placement` shouldn't be in the "plus" row |
| 42 | Small Asteroid | 209 | `temperature` shouldn't have "+" prefix when in same behavior as negative outputs |

**Generic fixes needed:**
1. **Global params and tile placements should NOT be grouped in plus/minus rows.** They are neutral effects. Split rendering: negative resources (attack row) → global params/tiles (neutral row) → positive resources (gain row).
2. **Choices with only negative outputs** should show minus signs on each choice option.
3. **`steal-*` targets** should have a distinct visual indicator (e.g., a steal arrow icon or different glow color).

**Test:** Visual - verify minus signs appear correctly, global params are not in plus/minus grouping.

---

### [x] Group 12: Layout and Formatting Issues

**Root cause:** Complex behavior configurations exceed the current layout handling.

| # | Card | ID | Issue | Fix |
|---|------|----|-------|-----|
| 5 | Optimal Aerobraking | 031 | Triggered effect layout broken | Verify triggered effect with `card-played` condition + selectors renders correctly. May need icon for "space event" trigger |
| 32 | Olympus Conference | 185 | Choice layout with card storage + card-draw is confusing | Need proper rendering of "add science resource" OR "spend science resource → draw card" choice layout |
| 35 | Invention Contest | 192 | 3 peek + 1 take icons not visually separated | Card draw consolidation should visually separate peek group from take group |
| 39 | Underground City | 032 | Shows "2" number instead of two separate heat icons | When amount=2 and total icons ≤ MAX_HORIZONTAL_ICONS, prefer showing 2 individual icons over "2 × icon" |
| 53 | Gyropolis | 230 | After JSON fix, needs 3-row layout (2 per-conditions + production) | BehaviorSection needs to handle 3+ production rows |
| 55 | Io Sulphur Research | 232 | OR condition with card icons renders badly, shows numbers for no reason | Card draw choice layout needs proper rendering without spurious number prefixes |

**Generic fixes:**
1. **Card draw consolidation** (`analyzeCardOutputs`): Visually separate different card operation types (peek vs take vs buy).
2. **Production row overflow**: Support 3+ production rows when multiple per-conditions exist.
3. **Icon count threshold**: When `amount` is small (2-3) and display space permits, show individual icons.

---

### [x] Group 13: Wrong or Missing Icons on Cards

| # | Card | ID | Issue |
|---|------|----|-------|
| 2 | Research Outpost | 020 | Missing city tile icon with asterisk (has city-placement with adjacency restriction) |
| 3 | Phobos Space Haven | 021 | Missing city tile icon with asterisk (after JSON fix adds city-placement) |
| 6 | Security Fleet | 028 | Wrong icon: shows titanium input + asteroid VP. Should show fighter resource (needs Group 15 first) |
| 28 | Herbivores | 147 | Missing greenery tile icon left of ":" in triggered effect box |
| 43 | Aerial Mappers | 213 | Extra "1" prefix on card icon when amount=1 |
| 46 | Atmoscoop | 217 | Shows venus-tag icon instead of venus global param icon |
| 51 | Extractor Balloons | 223 | Shows venus-tag icon instead of venus global param icon in choice |
| 56 | Local Shading | 235 | Missing number in credits production icon |
| 58 | Neutralizer Factory | 240 | Venus requirement should display as "X%" not just "X" |
| 66 | Venusian Plants | 261 | Missing venus "modifier" icon on animal/microbe resource icons (should show small venus badge) |

**Generic fixes:**
1. **Venus icon confusion**: Ensure `venus` ResourceType maps to the Venus global parameter icon, not the venus card tag icon. These are different: `venus` = global param (planet), `venus-tag` = card tag.
2. **Triggered effect trigger icon**: The `TriggeredEffectLayout` should render the trigger condition type as an icon before the `:` separator. E.g., `greenery-placed` → greenery tile icon.
3. **Card icon "1" prefix**: `CardIcon` should not show "1" when amount=1 for a single card operation.
4. **Requirement format**: Venus requirements should append "%" to the value.
5. **Resource modifier badges**: Support small overlay badges on resource icons (e.g., venus badge on animal icon to indicate "venus card animal").

---

### [x] Group 14: Missing Venus Global Param Outputs

Some cards reference `venus` as a tag icon in the frontend when it should be the Venus global parameter. This is partly a JSON issue (some cards are missing the Venus raise output entirely) and partly a frontend icon mapping issue.

Already covered in Group 4 (missing outputs) and Group 13 (wrong icons). Cross-reference:
- Air-Scrapping Expedition: Missing `venus: 1` output (Group 4)
- Comet for Venus: Missing `venus: 1` output (Group 4)
- Atmoscoop: Frontend uses wrong icon (Group 13)
- Extractor Balloons: Frontend uses wrong icon (Group 13)

---

## TIER 3: New Backend Features (Generic Capabilities)

Each feature enables multiple broken cards. Implementation should be generic enough to handle all current and future cards.

---

### [x] Group 15: New Resource Type — Fighter

**Cards affected:** Security Fleet (028)

**Problem:** The card stores "fighter" resources and gains VP per fighter. Currently uses `credit` for storage output and `asteroid` for VP. Neither is correct.

**Fix:**
1. Add `ResourceFighter = "fighter"` to `shared/resource_type.go`
2. Update Security Fleet JSON:
   - `resourceStorage: {type: "fighter", starting: 0}`
   - Manual output: `{type: "fighter", amount: 1, target: "self-card"}`
   - VP per: `{type: "fighter", amount: 1, target: "self-card"}`
3. Add `fighter` to `CARD_STORAGE_RESOURCE_TYPES` in frontend
4. Add fighter icon to frontend assets

**Backend test:** Play Security Fleet, use action (spend titanium → gain fighter on card), verify VP calculation.

---

### [x] Group 16: Variable-Amount User Selection

**Cards affected:** Insulation (152), Power Infrastructure (194)

**Problem:** These cards require the player to choose HOW MUCH of an effect to apply. Currently behaviors are `null`.

- **Insulation:** "Decrease heat production any number of steps. Increase M€ production the same number of steps." Player picks X (0 to current heat production).
- **Power Infrastructure:** "Spend any amount of energy to gain that amount of M€." Player picks X (0 to current energy).

**Generic capability needed:**
A new behavior input type or modifier that indicates "variable amount, player selects". Options:

**Option A: New `variableAmount` field on ResourceCondition**
```json
{
  "inputs": [{"type": "heat-production", "amount": 1, "variableAmount": true}],
  "outputs": [{"type": "credit-production", "amount": 1, "variableAmount": true}]
}
```
Backend validates the selected amount is within valid range (0 to max available).

**Option B: New `amountSelection` field on CardBehavior**
```json
{
  "amountSelection": {"min": 0, "maxResource": "heat-production"},
  "inputs": [{"type": "heat-production", "amount": 1}],
  "outputs": [{"type": "credit-production", "amount": 1}]
}
```
Amount multiplied by the selection value.

**Recommended: Option A** — simpler, works per-resource, and the frontend can detect `variableAmount: true` to show a slider/input.

**Backend changes:**
1. Add `VariableAmount bool` field to `ResourceCondition` in `shared/card_behavior_types.go`
2. In `BehaviorApplier.ApplyInputs/ApplyOutputs`, multiply `amount` by the user-selected value
3. In `PlayCardAction` and `UseCardActionAction`, accept `selectedAmount *int` parameter
4. In `state_calculator.go`, calculate max selectable amount for frontend metadata
5. In `play_card.go`/`use_card_action.go`, validate selected amount is in valid range

**Frontend changes:**
1. Detect `variableAmount` flag on inputs/outputs
2. Show amount selector UI (slider or number input)
3. Send selected amount in WebSocket message

**Backend test:** Play Insulation with heat production 5, select 3 → verify heat production -3, credit production +3. Edge cases: select 0, select max.

**Also enables:** Sulphur-Eating Bacteria's "spend any number of microbes" (Group 8).

---

### [x] Group 17: Temporary Effects (Next Card Only)

**Cards affected:** Indentured Workers (195), Special Design (206)

**Problem:** These cards apply an effect that only lasts for the NEXT card played, then auto-removes. Currently behaviors are `null`.

- **Indentured Workers:** "The next card you play this generation costs 8 M€ less."
- **Special Design:** "The next card you play this generation, its global requirements are +/- 2 steps."

**Generic capability needed:**
A "temporary effect" system that:
1. Registers a modifier on the player
2. Auto-removes after one card play (or at generation end)
3. Is calculated by `RequirementModifierCalculator` for costs/requirements

**Backend changes:**
1. Add `TemporaryEffect` struct to `player/`:
   ```go
   type TemporaryEffect struct {
       SourceCardID string
       EffectType   string        // "discount" or "global-parameter-lenience"
       Amount       int
       ExpiresAfter string        // "next-card" or "generation-end"
   }
   ```
2. Add `temporaryEffects []TemporaryEffect` to Player with Add/Remove/Get methods
3. In `RequirementModifierCalculator`: include temporary discount effects in cost calculations
4. In `state_calculator.go` `validateRequirements`: include temporary lenience in requirement checks
5. In `PlayCardAction.Execute()`: after successful card play, remove all "next-card" temporary effects
6. Subscribe to `GenerationAdvancedEvent` to clear "generation-end" effects

**JSON data:**
```json
// Indentured Workers
{"behaviors": [{"triggers": [{"type": "auto"}], "outputs": [{"type": "discount", "amount": 8, "temporary": "next-card"}]}]}

// Special Design
{"behaviors": [{"triggers": [{"type": "auto"}], "outputs": [{"type": "global-parameter-lenience", "amount": 2, "temporary": "next-card"}]}]}
```

**Backend test:** Play Indentured Workers, verify next card costs 8 less, verify effect removed after playing that card. Play two cards after Special Design, verify only first gets lenience.

---

### [x] Group 18: Resource Storage as Payment / Spending Storage Resources

**Cards affected:** Dirigibles (222), Rotator Impacts (243), Stratospheric Birds (249), Sulphur-Eating Bacteria (251)

**Problem:** These cards involve spending resources FROM card storage as part of:
- A) Payment for playing Venus cards (Dirigibles: floaters worth 3 M€ each)
- B) Input cost for a card action (Rotator Impacts: spend asteroid to raise Venus)
- C) Cost to play a card (Stratospheric Birds: spend 1 floater from any card as play cost)
- D) Action to convert storage resources to credits (Sulphur-Eating Bacteria: spend X microbes for 3X M€)

**Sub-problems:**

**18A: Storage resource inputs in card actions (B + D)**
Rotator Impacts already has the correct JSON: `inputs: [{type: "asteroid", amount: 1, target: "self-card"}]`. The backend `BehaviorApplier.ApplyInputs()` needs to handle `target: "self-card"` by deducting from card storage instead of player resources.

**Backend changes for 18A:**
1. In `BehaviorApplier.ApplyInputs()`: when `target == "self-card"`, deduct from `player.Resources().RemoveFromStorage(cardID, amount)` instead of `player.Resources().Deduct()`
2. In `state_calculator.go` `validateInputResources`: check storage has sufficient resources when `target == "self-card"`
3. For Sulphur-Eating Bacteria: combines with variable-amount (Group 16) - add second choice with variable microbe input from self-card

**18B: Storage resource as card payment (A + C)**
Dirigibles: floaters can be used as M€ payment (3 M€ per floater) when playing Venus-tagged cards.
Stratospheric Birds: requires spending 1 floater from any card when playing this card.

**Backend changes for 18B (Dirigibles-style):**
1. Add a new effect type `storage-payment-substitute` that functions like `payment-substitute` but draws from card storage
2. In `PaymentRequest`, add `StorageSubstitutes map[string]int` (cardID → amount of storage resources to use)
3. In `CardPayment.TotalValue()`, include storage substitute value
4. In `CardPayment.CanAfford()`, verify storage has resources
5. The Dirigibles auto behavior should register this as a persistent effect with selectors for Venus tags

**Backend changes for 18C (Stratospheric Birds-style):**
1. This is a one-time cost when playing the card, modeled as a card input: `inputs: [{type: "floater", amount: -1, target: "any-card"}]`
2. `ApplyInputs()` handles `any-card` target by requiring player to select which card to deduct from
3. Add `targetCardID` parameter to the play card flow for input targeting

**Backend test:** Use Rotator Impacts action spending asteroid, verify storage decremented. Play Venus card with Dirigibles floaters as payment. Play Stratospheric Birds, verify floater deducted from another card.

---

### [x] Group 19: Place Resource on Any Card's Storage (Generic "Card Resource")

**Cards affected:** CEO's Favorite Project (149), Corroder Suits (219), Freyja Biodomes (227), Maxwell Base (238), Venusian Plants (261)

**Problem:** These cards add resources to OTHER cards' storage. Currently modeled as `{type: "credit", amount: 1, target: "any-card"}` which doesn't make sense — `credit` isn't a storage resource type. Need a generic "add resource to target card's storage type" output.

**Generic capability needed:**
A new resource type `card-resource` that means "add 1 of whatever the target card stores."

**Backend changes:**
1. Add `ResourceCardResource = "card-resource"` to `shared/resource_type.go`
2. In `BehaviorApplier.applyOutput()` for `card-resource`:
   - Look up target card's `ResourceStorage.Type`
   - Add `amount` of that resource to the target card's storage
   - Requires `targetCardID` to be set (player selects which card)
3. In `state_calculator.go`: validate player has at least one card with storage (optionally matching selectors)
4. Frontend: show a generic "resource" icon with selector indicator; when player clicks, show eligible cards

**JSON fix for all affected cards:**
Change `type: "credit"` → `type: "card-resource"` in outputs with `target: "any-card"`.

**Backend test:** Play CEO's Favorite Project targeting a card with animal storage, verify animal count increased. Test with microbe and floater storage cards.

---

### [x] Group 20: Conditional Effects (If Player Has X, Then Do Y)

**Cards affected:** Comet for Venus (218), Io Sulphur Research (232), Flooding (188)

**Problem:** Some card effects should only apply conditionally based on the target player's state.

- **Comet for Venus:** Remove 4 M€ from a player **with a Venus tag in play**
- **Io Sulphur Research:** Draw 3 cards **if you have 3+ Venus tags**, otherwise draw 1
- **Flooding:** Remove 4 M€ from **owner of an adjacent tile** (depends on tile placement)

**Generic capability needed:**
A `condition` field on Choice that gates whether the choice is available.

**Backend changes:**
1. Add `Requirements *CardRequirements` field to `Choice` struct in `shared/card_behavior_types.go`
2. In `PlayCardAction`, when processing choices, validate requirements before allowing selection
3. In `state_calculator.go`, mark unavailable choices in metadata so frontend can disable them
4. For Flooding: this is a special case where the target depends on tile placement — needs post-tile-selection callback

**JSON data:**
```json
// Io Sulphur Research - Choice 2 requires 3 Venus tags
{
  "choices": [
    {"outputs": [{"type": "card-draw", "amount": 1}]},
    {"outputs": [{"type": "card-draw", "amount": 3}], "requirements": {"items": [{"type": "tags", "min": 3, "tag": "venus"}]}}
  ]
}
```

**Backend test:** Play Io Sulphur Research with 2 Venus tags → only choice 1 available. With 3+ → both available.

---

### [x] Group 21: Card Discard and Opponent Draw

**Cards affected:** Mars University (073), Sponsored Academies (247)

**Problem:** These cards involve discarding from hand and/or opponents drawing cards. Neither mechanic is currently supported.

- **Sponsored Academies:** "Discard 1 card from hand, then draw 3 cards. All opponents draw 1 card." (immediate effect)
- **Mars University:** "When you play a science tag, you may discard a card from hand to draw a card." (passive effect)

**Key insight — this follows the existing pending selection pattern:**
The codebase already handles "pause and wait for player input" via pending selections:
- `PendingCardDrawSelection` — card peek/take/buy pauses for player to pick cards
- `PendingTileSelection` — tile placement pauses for player to pick hex

Mars University works the same way: when the passive effect fires, instead of auto-applying outputs, create a `PendingCardDiscardSelection`. The player sees a prompt ("discard a card to draw one?"), picks a card or skips, and the backend completes the effect. Same infrastructure, just wired into the passive effect path.

**Flow for Mars University:**
1. Player plays card with science tag → `TagPlayedEvent` fires
2. Passive effect subscriber detects Mars University's behavior has a `card-discard` input
3. Instead of calling `ApplyOutputs()`, creates `PendingCardDiscardSelection` with optional=true
4. Frontend shows "Mars University: discard a card to draw one?" with skip option
5. Player discards → backend draws 1 card. Player skips → nothing happens.
6. Selection cleared, game continues.

**Flow for Sponsored Academies (immediate):**
1. Card played → auto behavior fires
2. `card-discard` input detected → creates `PendingCardDiscardSelection`
3. Player picks card to discard
4. On completion: draw 3 cards for player, draw 1 card for each opponent

**Generic capabilities needed:**
1. **Card discard input type**: `{type: "card-discard", amount: 1, target: "self-player"}` — player selects card(s) to discard from hand
2. **Opponent card draw**: `{type: "card-draw", amount: 1, target: "all-opponents"}` — all other players draw cards
3. **Optional inputs**: `optional: true` on ResourceCondition — player can skip

**Backend changes:**
1. Add `ResourceCardDiscard = "card-discard"` to resource types
2. Add `Optional bool` field to `ResourceCondition`
3. Create `PendingCardDiscardSelection` phase state (similar to `PendingCardDrawSelection`):
   ```go
   type PendingCardDiscardSelection struct {
       MinCards    int      // 0 if optional, 1 if required
       MaxCards    int      // max cards to discard
       Source      string   // "mars-university" or "sponsored-academies"
       SourceCardID string
       // After discard completes, apply these outputs:
       PendingOutputs []shared.ResourceCondition
   }
   ```
4. Add confirmation action + WebSocket handler: `ConfirmCardDiscardAction`
   - Validates selected cards are in hand
   - Removes selected cards from hand (discards them)
   - Applies `PendingOutputs` via BehaviorApplier (draws, opponent draws, etc.)
   - Clears the pending selection
5. Add `TargetAllOpponents = "all-opponents"` target type
6. In `BehaviorApplier.ApplyOutputs()`: handle `all-opponents` target by iterating non-active players
7. Extend `passive_effect_subscriber.go`: when a triggered behavior has `card-discard` inputs, create `PendingCardDiscardSelection` instead of calling `ApplyOutputs()`
8. Extend `PlayCardAction` behavior application: when an auto behavior has `card-discard` inputs, create `PendingCardDiscardSelection` and defer output application

**JSON data:**
```json
// Mars University - passive trigger on science tag, optional discard
{
  "triggers": [{"type": "auto", "condition": {"type": "tag-played", "target": "self-player", "selectors": [{"tags": ["science"]}]}}],
  "inputs": [{"type": "card-discard", "amount": 1, "target": "self-player", "optional": true}],
  "outputs": [{"type": "card-draw", "amount": 1, "target": "self-player"}]
}

// Sponsored Academies - immediate, mandatory discard
{
  "triggers": [{"type": "auto"}],
  "inputs": [{"type": "card-discard", "amount": 1, "target": "self-player"}],
  "outputs": [
    {"type": "card-draw", "amount": 3, "target": "self-player"},
    {"type": "card-draw", "amount": 1, "target": "all-opponents"}
  ]
}
```

**Backend test:** Play Mars University, play science card → get discard prompt → discard → draw new card. Also test skip. Play Sponsored Academies → discard 1 → draw 3 → verify opponents each draw 1.

---

### [x] Group 22: Tile Placement Restriction — Adjacent to N Tiles

**Cards affected:** Urbanized Area (120), Ecological Zone (128)

**Problem:** These cards require placing tiles with specific adjacency conditions that go beyond the current `TileRestrictions` system.

- **Urbanized Area:** City tile **adjacent to at least 2 other city tiles**
- **Ecological Zone:** Tile **adjacent to a greenery you own**

**Current system:** `TileRestrictions` supports:
- `Adjacency: "none"` (no adjacent occupied tiles)
- `OnTileType: "ocean"` (only on ocean spaces)
- `BoardTags: []` (specific reserved spaces)

**Generic capability needed:**
Extended `TileRestrictions` with:
- `MinAdjacentOfType` — minimum count of adjacent tiles of a specific type
- `AdjacentToOwned` — must be adjacent to a tile owned by the placing player
- `AdjacentToType` — must be adjacent to a specific tile type

**Backend changes:**
1. Extend `TileRestrictions` in `shared/`:
   ```go
   type TileRestrictions struct {
       BoardTags          []string `json:"boardTags,omitempty"`
       Adjacency          string   `json:"adjacency,omitempty"`        // "none"
       OnTileType         string   `json:"onTileType,omitempty"`       // "ocean"
       AdjacentToType     string   `json:"adjacentToType,omitempty"`   // "city", "greenery"
       MinAdjacentOfType  *int     `json:"minAdjacentOfType,omitempty"` // min count
       AdjacentToOwned    bool     `json:"adjacentToOwned,omitempty"`  // must own adjacent tile
   }
   ```
2. In `Board.GetAvailableHexesForTile()`: filter by new restriction fields
3. In `state_calculator.go`: validate tile outputs consider new restrictions

**JSON data:**
```json
// Urbanized Area
{"type": "city-placement", "amount": 1, "tileRestrictions": {"adjacentToType": "city", "minAdjacentOfType": 2}}

// Ecological Zone
{"type": "greenery-placement", "amount": 1, "tileRestrictions": {"adjacentToType": "greenery", "adjacentToOwned": true}}
```

**Note:** Ecological Zone also needs a custom tile type (eco-zone). See Group 23.

**Backend test:** Verify Urbanized Area only offers hexes with 2+ adjacent cities. Verify Ecological Zone only offers hexes next to player-owned greenery.

---

### [x] Group 23: New Tile Types

**Cards affected:** Natural Preserve (044), Mining Area (064), Mining Rights (067), Nuclear Zone (097), Ecological Zone (128), Mohole Area (142), Restricted Area (199)

**Problem:** These cards place tiles that don't exist as tile types in the system. Currently they either have no tile placement output or use a generic city/greenery type.

**Tiles needed:**

| Tile | Type Constant | Placement Rule | Placeholder Label |
|------|--------------|----------------|-------------------|
| Natural Preserve | `"natural-preserve"` | No adjacent tiles | "Nature Preserve" |
| Mining | `"mining"` | On hex with steel/titanium bonus | "Mining" |
| Nuclear Zone | `"nuclear-zone"` | Normal land | "Nuclear Zone" |
| Ecological Zone | `"ecological-zone"` | Adjacent to owned greenery | "Eco Zone" |
| Mohole | `"mohole"` | On ocean-reserved space | "Mohole" |
| Restricted Area | `"restricted"` | Normal land | "Restricted" |

**Backend changes:**
1. Add tile type constants to `board/`:
   ```go
   TileTypeNaturalPreserve = "natural-preserve"
   TileTypeMining          = "mining"
   TileTypeNuclearZone     = "nuclear-zone"
   TileTypeEcologicalZone  = "ecological-zone"
   TileTypeMohole          = "mohole"
   TileTypeRestricted      = "restricted"
   ```
2. Add a generic `tile-placement` output type (or card-specific types like `natural-preserve-placement`)
3. In `Board.UpdateTileOccupancy()`: support new tile types
4. Ensure new tile types don't grant VP like cities/greeneries unless specified

**Frontend changes — Purple placeholder tiles (no 3D models needed):**
All new tile types render as **purple hex tiles** with a **white text label** showing the tile name. This avoids needing 3D models and makes them visually distinct from cities/greeneries/oceans.

1. In the tile rendering system (`Tile.tsx` / `TileGrid.tsx`), detect unknown/special tile types
2. Render them as a **purple-colored hex** (e.g., `#9C27B0` / purple-600) with the tile's label text centered on it
3. Add a tile type → label mapping:
   ```ts
   const SPECIAL_TILE_LABELS: Record<string, string> = {
     "natural-preserve": "Nature Preserve",
     "mining": "Mining",
     "nuclear-zone": "Nuclear Zone",
     "ecological-zone": "Eco Zone",
     "mohole": "Mohole",
     "restricted": "Restricted",
   };
   ```
4. Placeholder tiles use the same hex geometry, owner color border, and emergence animation as existing tiles — just purple fill + text instead of a 3D model
5. Later these can be upgraded to proper 3D models without any backend changes

**JSON data updates:** Add tile placement outputs with appropriate restrictions to each affected card.

**Backend test:** Place each tile type, verify board state updates correctly. Verify tile restrictions are enforced.

---

### [x] Group 24: Water Import From Europa — Titanium as Payment for Actions

**Cards affected:** Water Import From Europa (012), Rotator Impacts (243)

**Problem:** Card actions have M€ costs where titanium may be used as payment. The current system only supports titanium/steel payment for playing cards, not for card action inputs.

- **Water Import From Europa:** Action costs 12 M€, "titanium may be used as if playing a space card."
- **Rotator Impacts:** Choice 1 costs 6 M€, "titanium may be used." (Choice 2 — spending asteroid from storage — is covered by Group 18.)

**Current system:** `PaymentRequest` with steel/titanium/substitutes is used for playing cards. Card actions use `ApplyInputs()` which directly deducts resources. There's no payment system for action inputs.

**Fix:**
1. Extend action inputs to support payment-like behavior with titanium/steel
2. Add `paymentAllowed` field to `ResourceCondition` indicating which alternative resources can be used
3. In `UseCardActionAction`, accept a `PaymentRequest` for action inputs
4. In `BehaviorApplier.ApplyInputs()`, support payment calculation with substitutes

**JSON fix:**
```json
// Water Import From Europa
{"inputs": [{"type": "credit", "amount": 12, "paymentAllowed": ["titanium"]}]}

// Rotator Impacts - Choice 1
{"inputs": [{"type": "credit", "amount": 6, "paymentAllowed": ["titanium"]}]}
```

**Backend test:** Use Water Import action paying 6 credits + 2 titanium (value 3 each = 6) = 12 total. Verify titanium deducted. Use Rotator Impacts choice 1 paying with titanium.

---

### [x] Group 25: Land Claim Visual Indicator

**Card:** Land Claim (066)

**Problem:** Land claim (`land-claim` type) already exists in the tile system, but there's no clear visual indicator on the 3D board showing that a hex is reserved.

**Fix (Frontend only) — Purple placeholder with "Claimed" label:**
1. In `Tile.tsx`, detect reserved/land-claim status
2. Render as a **purple hex** (same style as Group 23 special tiles) with the label **"Claimed"**
3. Show owner's player color as the hex border so it's clear who reserved it
4. The purple + "Claimed" text makes it immediately obvious the hex is taken but not yet built on

**No backend changes needed** — land-claim is already implemented.

---

## TIER 4: Test Plan

All backend changes must have corresponding tests in `backend/test/`.

### Test Categories:

**1. JSON Data Fix Tests** (`test/cards/card_data_test.go`)
- Load card registry and verify corrected field values
- Verify per-condition divisors
- Verify trigger conditions exist on passive effect cards
- Verify missing outputs are present
- Verify requirements exist

**2. Behavior Application Tests** (`test/action/`)
- Variable amount selection (Insulation, Power Infrastructure)
- Card storage inputs (Rotator Impacts, Stratospheric Birds)
- Card-resource output type (CEO's Favorite Project)
- Card discard + draw (Mars University, Sponsored Academies)
- Opponent effects (Sponsored Academies all-opponents draw)

**3. State Calculator Tests** (`test/action/state_calculator_test.go`)
- Temporary effects applied to next card cost/requirements
- Storage-as-payment validation
- Choice requirements gating
- Variable amount max calculation
- New tile restriction validation

**4. Passive Effect Tests** (`test/action/passive_effects_test.go`)
- Media Group fires on event card play
- Standard Technology fires on standard project
- Venusian Animals fires on science tag play
- Mars University triggers on science tag with optional discard

**5. Tile Placement Tests** (`test/action/tile/`)
- New tile types place correctly
- Adjacency restrictions (min 2 cities for Urbanized Area)
- Adjacent-to-owned (Ecological Zone)
- Ocean-space placement (Mohole Area)
- Board tag placement (Phobos Space Haven)

**6. Integration Tests** (`test/integration/`)
- Full card play flow for each fixed card
- Payment with storage resources (Dirigibles)
- Temporary effect lifecycle (play → next card → removed)
- Variable amount flows end-to-end

---

## Implementation Order (Recommended)

**Sprint 1 — Quick Wins (no code changes):**
- Groups 1-9: JSON data fixes
- Run existing tests to verify nothing breaks

**Sprint 2 — Frontend Fixes:**
- Group 10: "Any" red tint on per-conditions
- Group 11: Minus/negative display
- Group 12: Layout improvements
- Group 13: Icon fixes

**Sprint 3 — Core Backend Features:**
- Group 16: Variable amount selection (enables Insulation, Power Infrastructure, Sulphur-Eating Bacteria)
- Group 17: Temporary effects (enables Indentured Workers, Special Design)
- Group 19: Card-resource output type (enables CEO's Favorite, Corroder Suits, Freyja Biodomes, Maxwell Base, Venusian Plants)

**Sprint 4 — Advanced Backend Features:**
- Group 15: Fighter resource type
- Group 18: Storage-as-payment (enables Dirigibles, Rotator Impacts, Stratospheric Birds)
- Group 20: Conditional effects (enables Comet for Venus, Io Sulphur Research)
- Group 21: Card discard + opponent draw (enables Mars University, Sponsored Academies)

**Sprint 5 — Tile System:**
- Group 22: Tile restriction extensions
- Group 23: New tile types + 3D models
- Group 24: Titanium payment for actions
- Group 25: Land claim visual

---

## Card Index (sorted by card #)

| # | Card | ID | Groups | Complexity |
|---|------|----|--------|------------|
| 1 | Water Import From Europa | 012 | 24 | Hard |
| 2 | Research Outpost | 020 | 13 | Easy |
| 3 | Phobos Space Haven | 021 | 5, 13 | Medium |
| 4 | Predators | 024 | 13 | Easy (display only — data is correct) |
| 5 | Optimal Aerobraking | 031 | 12 | Medium |
| 6 | Security Fleet | 028 | 15 | Medium |
| 7 | Ants | 035 | 13 | Easy (display only — data is correct) |
| 8 | Natural Preserve | 044 | 23 | Hard |
| 9 | Virus | 050 | 11 | Medium |
| 10 | Mining Expedition | 063 | 11 | Easy |
| 11 | Mining Area | 064 | 23 | Hard |
| 12 | Mining Rights | 067 | 23 | Hard |
| 13 | Land Claim | 066 | 25 | Medium |
| 14 | Mars University | 073 | 3, 21 | Hard |
| 15 | Mass Converter | 094 | — | Already correct (verified data matches description) |
| 16 | Physics Complex | 095 | 1 | Easy |
| 17 | Greenhouses | 096 | 10 | Easy |
| 18 | Nuclear Zone | 097 | 23 | Hard |
| 19 | Toll Station | 099 | 2, 10 | Easy |
| 20 | Media Group | 109 | 3 | Easy |
| 21 | Urbanized Area | 120 | 22 | Medium |
| 22 | Sabotage | 121 | 11 | Medium |
| 23 | Hired Raiders | 124 | 11 | Medium |
| 24 | Ecological Zone | 128 | 22, 23 | Hard |
| 25 | Zeppelins | 129 | 10 | Easy |
| 26 | Worms | 130 | 1 | Easy |
| 27 | Mohole Area | 142 | 23 | Hard |
| 28 | Herbivores | 147 | 13 | Easy |
| 29 | CEO's Favorite Project | 149 | 19 | Medium |
| 30 | Insulation | 152 | 16 | Hard |
| 31 | Standard Technology | 156 | 3 | Easy |
| 32 | Olympus Conference | 185 | 12 | Medium |
| 33 | Flooding | 188 | 11, 20 | Hard |
| 34 | Energy Saving | 189 | 10 | Easy |
| 35 | Invention Contest | 192 | 12 | Easy |
| 36 | Power Infrastructure | 194 | 16 | Hard |
| 37 | Indentured Workers | 195 | 17 | Medium |
| 38 | Restricted Area | 199 | 23 | Hard |
| 39 | Underground City | 032 | 12 | Easy |
| 40 | Special Design | 206 | 17 | Medium |
| 41 | Medical Lab | 207 | 1 | Easy |
| 42 | Small Asteroid | 209 | 11 | Easy |
| 43 | Aerial Mappers | 213 | 13 | Easy |
| 44 | Aerosport Tournament | 214 | 10 | Easy |
| 45 | Air-Scrapping Expedition | 215 | 4 | Easy |
| 46 | Atmoscoop | 217 | 13 | Easy |
| 47 | Comet for Venus | 218 | 4, 20 | Medium |
| 48 | Corroder Suits | 219 | 19 | Medium |
| 49 | Dawn City | 220 | 5 | Easy |
| 50 | Dirigibles | 222 | 18 | Hard |
| 51 | Extractor Balloons | 223 | 13 | Easy |
| 52 | Freyja Biodomes | 227 | 19 | Medium |
| 53 | Gyropolis | 230 | 4, 12 | Medium |
| 54 | Hydrogen To Venus | 231 | 4 | Easy (missing venus raise output, floater per-condition exists) |
| 55 | Io Sulphur Research | 232 | 7, 12, 20 | Hard |
| 56 | Local Shading | 235 | 13 | Easy |
| 57 | Maxwell Base | 238 | 5, 19 | Medium |
| 58 | Neutralizer Factory | 240 | 4, 13 | Easy |
| 59 | Rotator Impacts | 243 | 18, 24 | Medium |
| 60 | Sponsored Academies | 247 | 9, 21 | Hard |
| 61 | Stratopolis | 248 | 5 | Easy |
| 62 | Stratospheric Birds | 249 | 18 | Medium |
| 63 | Sulphur-Eating Bacteria | 251 | 8, 16, 18 | Hard |
| 64 | Terraforming Contract | 252 | 6 | Easy |
| 65 | Venusian Animals | 259 | 3, 13 | Easy |
| 66 | Venusian Plants | 261 | 13, 19 | Medium |

**Summary:** 23 Easy, 21 Medium, 16 Hard, 1 Already Working (Mass Converter)

---

## Description Mismatches

If a planned fix contradicts the card's `description` field, log it here instead of applying the incorrect fix.

| Card | ID | Group | Mismatch Description |
|------|----|-------|---------------------|
| Ecological Zone | 128 | 22 | Plan says `adjacentToOwned: true` but card description says "adjacent to any greenery tile" (not "your greenery"). Applied `adjacentToType: "greenery"` without `adjacentToOwned: true` to match the description. |
