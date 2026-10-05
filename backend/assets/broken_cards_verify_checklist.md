# Broken Cards Progress Tracker

The 66 unique entries in the [original card index](broken_cards_planned.md#card-index-sorted-by-card-) are the fixed baseline. Promo and Turmoil are excluded from active work. Reconciled against the working tree on 2026-10-04.

## Current count

- **64 in-scope fix candidates:** all 64 recorded as implemented; 13 have fresh backend verification (Group 10 plus nine follow-ups). Group 10 also has browser verification; the other 51 retain their historical implementation status.
- **1 already correct:** Mass Converter (094), excluded from the fixed-card count.
- **1 excluded:** Small Asteroid (209), which the database classifies as Promo.
- **66 total baseline entries.** Cards discovered later do not change this denominator.

“Recorded implemented” preserves the completed implementation status in the original plan; it is not a claim that every original issue was independently retested or visually approved. The latest full backend run passed (`go test ./...`, `/tmp/tm_followups_full.log`). Group 10's four cards also passed focused real-database tests and browser checks in Small and Large views on 2026-10-04.

The evidence column points to existing named tests as review starting points, not a certification that they cover every issue. Browser verification remains pending for the other historical issues. Earlier checkmarks are retained below as historical evidence, not promoted to fresh verification.

## One row per baseline card

| # | ID | Card | Original groups | Status | Existing test reference / review note |
|---|---|---|---|---|---|
| 1 | 012 | Water Import From Europa | 24 | Recorded implemented | [TestWaterImportFromEuropa_PayWithCreditsOnly](../test/action/behavior/basic_resource_test.go) |
| 2 | 020 | Research Outpost | 13 | Implemented; follow-up backend verified | [TestResearchOutpost_AdjacencyAndPermanentDiscount](../test/action/card_packs/play_card_base_test.go) Corrected “no other tile” wording; browser approval pending. |
| 3 | 021 | Phobos Space Haven | 5, 13 | Recorded implemented | [TestPhobosSpaceHaven_TitaniumProdAndCityPlacement](../test/action/card_packs/play_card_base_test.go) |
| 4 | 024 | Predators | 13 | Recorded implemented | [TestPredatorsStealAnimalFromOtherPlayer](../test/action/behavior/card_storage_test.go) |
| 5 | 031 | Optimal Aerobraking | 12 | Recorded implemented | [TestOptimalAerobrakingScenario](../test/cards/selector_matcher_test.go) |
| 6 | 028 | Security Fleet | 15 | Implemented; follow-up backend verified | [TestStorageScoringCards_RealActionsAndVP](../test/action/card_packs/play_card_corporate_era_test.go) Actual fighter action and scoring; stale fixture corrected. |
| 7 | 035 | Ants | 13 | Recorded implemented | [TestAntsStealMicrobeFromOtherPlayer](../test/action/behavior/card_storage_test.go) |
| 8 | 044 | Natural Preserve | 23 | Recorded implemented | [TestNaturalPreserve_TilePlacementWithNoAdjacency](../test/action/behavior/tile_placement_test.go) |
| 9 | 050 | Virus | 11 | Recorded implemented | [TestVirus_RemovesFromActualOwnerAndCanSkip](../test/action/card_packs/play_card_base_test.go) |
| 10 | 063 | Mining Expedition | 11 | Recorded implemented | [TestMiningExpedition_OxygenAndSteelAndRemovePlants](../test/action/card_packs/play_card_base_test.go) |
| 11 | 064 | Mining Area | 23 | Recorded implemented | [TestMiningArea_PlaceOnSteelBonus_IncreaseSteelProduction](../test/action/card_packs/play_card_corporate_era_test.go) |
| 12 | 067 | Mining Rights | 23 | Recorded implemented | [TestMiningRights_PlaceOnSteelBonus_IncreaseSteelProduction](../test/action/card_packs/play_card_base_test.go) |
| 13 | 066 | Land Claim | 25 | Implemented; follow-up backend verified | [TestLandClaim_ReservationLifecycle](../test/action/card_packs/play_card_corporate_era_test.go) Claim, opponent exclusion, owner construction, and bonuses verified. Flag uses player colour; browser approval pending. |
| 14 | 073 | Mars University | 3, 21 | Recorded implemented | [TestMarsUniversity_PlaysSuccessfully](../test/action/card_packs/play_card_corporate_era_test.go) |
| 15 | 094 | Mass Converter | — | Already correct | [TestMassConverter_EnergyProductionWithScienceRequirement](../test/action/card_packs/play_card_base_test.go) |
| 16 | 095 | Physics Complex | 1 | Implemented; follow-up backend verified | [TestStorageScoringCards_RealActionsAndVP](../test/action/card_packs/play_card_corporate_era_test.go) Actual science action and 2 VP per resource; stale fixture corrected. |
| 17 | 096 | Greenhouses | 10 | Implemented; Group 10 verified | [TestGlobalCityCountCards_PreviewAndPlay](../test/action/card_packs/play_card_base_test.go); Small/Large tint checked in browser. |
| 18 | 097 | Nuclear Zone | 23 | Recorded implemented | [TestNuclearZone_TilePlacementOnNormalLand](../test/action/behavior/tile_placement_test.go) |
| 19 | 099 | Toll Station | 2, 10 | Recorded implemented | [TestTollStation_CreditProductionPerOpponentSpaceTag](../test/action/card_packs/play_card_base_test.go) |
| 20 | 109 | Media Group | 3 | Recorded implemented | [TestMediaGroup_Gain3Credits](../test/action/card_packs/play_card_base_test.go) |
| 21 | 120 | Urbanized Area | 22 | Recorded implemented | [TestUrbanizedArea_CityAdjacentTo2Cities](../test/action/behavior/tile_placement_test.go) |
| 22 | 121 | Sabotage | 11 | Recorded implemented | [TestSabotage_RemoveTitaniumFromOpponent](../test/action/card_packs/play_card_corporate_era_test.go) |
| 23 | 124 | Hired Raiders | 11 | Recorded implemented | [TestPlayCardAction_HiredRaidersEmptyTargetID_SkipsSteal](../test/action/play_card/play_card_any_player_test.go) |
| 24 | 128 | Ecological Zone | 22, 23 | Recorded implemented | [TestEcologicalZone_RewardsEachTagExactlyOnce](../test/action/card_packs/play_card_base_test.go) |
| 25 | 129 | Zeppelins | 10 | Implemented; Group 10 verified | [TestGlobalCityCountCards_PreviewAndPlay](../test/action/card_packs/play_card_base_test.go); Small/Large tint checked in browser. |
| 26 | 130 | Worms | 1 | Recorded implemented | [TestWorms_OneMicrobeTagBefore_GainsOnePlantProduction](../test/action/behavior/production_test.go) |
| 27 | 142 | Mohole Area | 23 | Recorded implemented | [TestMoholeArea_TilePlacementOnOceanSpace](../test/action/behavior/tile_placement_test.go) |
| 28 | 147 | Herbivores | 13 | Recorded implemented | [TestHerbivores_AddAnimalAndDecreaseTargetPlantProd](../test/action/card_packs/play_card_base_test.go) |
| 29 | 149 | CEO's Favorite Project | 19 | Recorded implemented | [TestCEOsFavoriteProject_AddsMicrobeToTargetCard](../test/action/card_packs/play_card_corporate_era_test.go) |
| 30 | 152 | Insulation | 16 | Recorded implemented | [TestInsulation_DecreaseHeatProductionIncreaseCreditProduction](../test/action/core/variable_amount_test.go) |
| 31 | 156 | Standard Technology | 3 | Recorded implemented | [TestStandardTechnology_Gain3Credits](../test/action/card_packs/play_card_base_test.go) |
| 32 | 185 | Olympus Conference | 12 | Recorded implemented | [TestOlympusConference_AddScienceOnScienceTagPlayed](../test/action/card_effects/passive_card_effects_test.go) |
| 33 | 188 | Flooding | 11, 20 | Recorded implemented | [TestFlooding_AdjacentOpponentTile_RemovalOffered](../test/action/card_packs/play_card_base_test.go) |
| 34 | 189 | Energy Saving | 10 | Implemented; Group 10 verified | [TestGlobalCityCountCards_PreviewAndPlay](../test/action/card_packs/play_card_base_test.go); Small/Large tint checked in browser. |
| 35 | 192 | Invention Contest | 12 | Recorded implemented | [TestInventionContest_PlaysSuccessfully](../test/action/card_packs/play_card_corporate_era_test.go) |
| 36 | 194 | Power Infrastructure | 16 | Recorded implemented | [TestPowerInfrastructure_PlayAndUseAction](../test/action/card_packs/play_card_corporate_era_test.go) |
| 37 | 195 | Indentured Workers | 17 | Recorded implemented | [TestIndenturedWorkers_DiscountAppliedToNextCard](../test/action/card_effects/temporary_effects_test.go) |
| 38 | 199 | Restricted Area | 23 | Recorded implemented | [TestRestrictedArea_SpendCreditsToDrawCard](../test/action/card_packs/play_card_base_test.go) |
| 39 | 032 | Underground City | 12 | Recorded implemented | [TestUndergroundCity_ProductionAndCityPlacement](../test/action/card_packs/play_card_base_test.go) |
| 40 | 206 | Special Design | 17 | Recorded implemented | [TestSpecialDesign_LenienceAppliedToNextCard](../test/action/card_effects/temporary_effects_test.go) |
| 41 | 207 | Medical Lab | 1 | Recorded implemented | [TestMedicalLab_CreditProductionPerTwoBuildingTags](../test/action/card_packs/play_card_base_test.go) |
| 42 | 209 | Small Asteroid | 11 | Excluded (Promo) | Card-specific test coverage needs review. |
| 43 | 213 | Aerial Mappers | 13 | Recorded implemented | [TestAerialMappers_PlayAndStorageCreated](../test/action/card_packs/play_card_venus_test.go) |
| 44 | 214 | Aerosport Tournament | 10 | Implemented; Group 10 verified | [TestAerosportTournament_GainCreditsForAllCities / StoredFloaterRequirement](../test/action/card_packs/play_card_venus_test.go); Small/Large tint checked in browser. |
| 45 | 215 | Air-Scrapping Expedition | 4 | Implemented; follow-up backend verified | [TestVenusStorageCards_RealDefinitions / InvalidTargetsAreAtomic](../test/action/card_packs/play_card_venus_test.go) Actual Venus raise and storage targeting verified. |
| 46 | 217 | Atmoscoop | 13 | Implemented; follow-up backend verified | [TestVenusStorageCards_RealDefinitions / InvalidTargetsAreAtomic / RequirementsAndChoices](../test/action/card_packs/play_card_venus_test.go) Both choices, shared floaters, caps, and atomic validation verified. |
| 47 | 218 | Comet for Venus | 4, 20 | Recorded implemented | [TestCometForVenus_OptionalRestrictedRemoval](../test/action/card_packs/play_card_venus_test.go) |
| 48 | 219 | Corroder Suits | 19 | Recorded implemented | [TestCorroderSuits_CreditProductionIncrease](../test/action/card_packs/play_card_venus_test.go) |
| 49 | 220 | Dawn City | 5 | Implemented; follow-up backend verified | [TestDawnCity_RealPlacementAndRequirements](../test/action/card_packs/play_card_venus_test.go) Corrected description; named-space placement, requirements, and VP verified. Occupied/claimed named spaces no longer fall back elsewhere. |
| 50 | 222 | Dirigibles | 18 | Recorded implemented | [TestDirigibles_StoragePaymentSubstitute_VenusCard](../test/action/card_packs/play_card_venus_test.go) |
| 51 | 223 | Extractor Balloons | 13 | Recorded implemented | [TestExtractorBalloons_PlayAdds3Floaters](../test/action/card_packs/play_card_venus_test.go) |
| 52 | 227 | Freyja Biodomes | 19 | Implemented; follow-up backend verified | [TestVenusStorageCards_RealDefinitions / InvalidTargetsAreAtomic / RequirementsAndChoices](../test/action/card_packs/play_card_venus_test.go) Both storage choices, shared production, requirements, and VP verified. |
| 53 | 230 | Gyropolis | 4, 12 | Recorded implemented | [TestGyropolis_ProductionAndCityPlacement](../test/action/card_packs/play_card_venus_test.go) |
| 54 | 231 | Hydrogen To Venus | 4 | Implemented; follow-up backend verified | [TestHydrogenToVenus_JovianScaling / TestVenusStorageCards_RealDefinitions / InvalidTargetsAreAtomic](../test/action/card_packs/play_card_venus_test.go) Actual Venus output and own Jovian scaling verified; misleading synthetic fixture renamed. |
| 55 | 232 | Io Sulphur Research | 7, 12, 20 | Recorded implemented | [TestIoSulphurResearch_DrawOneCardWithoutVenusTags](../test/action/card_packs/play_card_venus_test.go) |
| 56 | 235 | Local Shading | 13 | Recorded implemented | [TestLocalShading_Action_AddFloater](../test/action/card_packs/play_card_venus_test.go) |
| 57 | 238 | Maxwell Base | 5, 19 | Recorded implemented | [TestMaxwellBase_DecreaseEnergyProductionAndCityPlacement](../test/action/card_packs/play_card_venus_test.go) |
| 58 | 240 | Neutralizer Factory | 4, 13 | Recorded implemented | [TestNeutralizerFactory_IncreaseVenus](../test/action/card_packs/play_card_venus_test.go) |
| 59 | 243 | Rotator Impacts | 18, 24 | Recorded implemented | [TestRotatorImpacts_AddFloater](../test/action/card_packs/play_card_venus_test.go) |
| 60 | 247 | Sponsored Academies | 9, 21 | Recorded implemented | [TestSponsoredAcademies_DiscardDrawAndOpponentDraw](../test/action/card_packs/play_card_venus_test.go) |
| 61 | 248 | Stratopolis | 5 | Recorded implemented | [TestStratopolis_CityPlacement](../test/action/card_packs/play_card_venus_test.go) |
| 62 | 249 | Stratospheric Birds | 18 | Recorded implemented | [TestStratosphericBirds_ActionAddAnimal](../test/action/card_packs/play_card_venus_test.go) |
| 63 | 251 | Sulphur-Eating Bacteria | 8, 16, 18 | Recorded implemented | [TestSulphurEatingBacteria_Choice0_AddMicrobe](../test/action/card_packs/play_card_venus_test.go) |
| 64 | 252 | Terraforming Contract | 6 | Recorded implemented | [TestTerraformingContract_CreditProduction](../test/action/card_packs/play_card_venus_test.go) |
| 65 | 259 | Venusian Animals | 3, 13 | Recorded implemented | [TestVenusianAnimals_PlaysAndRegistersEffect](../test/action/card_packs/play_card_venus_test.go) |
| 66 | 261 | Venusian Plants | 13, 19 | Recorded implemented | [TestVenusianPlants_RaiseVenusWithAnimalChoice](../test/action/card_packs/play_card_venus_test.go) |

## Open verification and follow-ups

- [x] **Group 10:** Greenhouses (096), Zeppelins (129), Energy Saving (189), Aerosport Tournament (214). Omitted city-count targets now use the global-count tint. Confirmed all four in Small and Large browser views, with Toll Station and Medical Lab as opponent/self-only comparisons. The shared city counter now respects Mars-only location filters, including owner filters; previews match execution. Aerosport Tournament's resource requirement now counts stored floaters across owned project cards and corporation, with 4/5/6 boundary tests, unrelated resource/opponent exclusions, and no resource consumption. No card JSON changes.
- [x] **Nine follow-ups:** Dawn City, Physics Complex, Security Fleet, Research Outpost, Air-Scrapping Expedition, Atmoscoop, Freyja Biodomes, Hydrogen To Venus, and Land Claim. Actual database tests cover the scenarios listed above; descriptions and stale fixtures corrected.
- [x] **Shared validation:** positive storage gains validate ownership, storage type, and selectors before payment; mandatory destinations cannot be omitted when eligible storage exists. Empty destinations remain valid for zero gains or no eligible storage. Tests cover entering cards, corporations, positional zero outputs, and manual action input preservation. Missing/out-of-range card choices are rejected before mutation.
- [x] **Named placement restrictions:** maps without the named area retain the existing placement fallback, but an occupied or claimed named area cannot fall back to an unrelated space.
- [ ] **Follow-up browser approval:** corrected Dawn City/Research Outpost descriptions in Small and Large views; Land Claim marker matches each player's actual colour and disappears after building. User owns visual testing.
- [ ] Verify remaining original rendering issues in Small and Large views, and placement/claim markers on the board (Group 10 is complete). Preserve the distinction between code implementation and visual approval.

## Additional work outside the baseline

**Martian Rails (007):** fixed Mars-only city counting through the same shared counter; `TestMartianRails_ActionCountsOnlyMarsCities` verifies the action excludes off-Mars cities and spends energy once. This additional fix does not change the 66-card denominator.

Later work includes Search For Life, Robotic Workforce, Viral Enhancers, payment substitutes, wild tags, corporation first actions, received-card presentation, and Viron action reuse. These are **not added to the 66-card count**. This is not an exhaustive count of additional cards fixed.

Structured descriptions are implemented in card data and [CardDescriptionSections](../../frontend/src/components/ui/display/CardDescriptionSections.tsx), with generic/effect/action/requirement sections and generated labels. The old “Later” checklist below is historical; Small/Large visual verification remains pending.

## Latest verification limits

Full backend tests passed (`/tmp/tm_followups_full.log`); focused follow-up regressions passed, including the final Dawn City occupied/claimed-space and preview checks. Frontend typecheck, lint, and production build passed (existing lint/build warnings). Both changed board components pass formatting; `git diff --check` passed.

The full frontend formatting check still reports the unrelated generated `api-types.ts`. `make lint` remains blocked by the existing errcheck package-loading error for `encoding/json` in `backend/tools` (`/tmp/tm_followups_lint.log`). No browser testing was performed for these follow-ups, at the user's request.

## Historical verification checklist

The original checkbox state is preserved below. Cards repeat across groups, so do not use these checkbox totals to count cards. The completed group headings in the original plan likewise do not certify fresh verification.


Verify each card renders and behaves correctly after fixes.

## TIER 1: JSON-Only Data Fixes

### Group 1: Wrong `per` Divisor Values
- [x] #16 Physics Complex (095)
- [x] #26 Worms (130)
- [x] #41 Medical Lab (207)

### Group 2: Wrong Target on Per-Condition
- [x] #19 Toll Station (099)

### Group 3: Missing Trigger Conditions on Passive Effects
- [ ] #14 Mars University (073)
- [x] #20 Media Group (109)
- [x] #31 Standard Technology (156)
- [x] #65 Venusian Animals (259)

### Group 4: Missing Outputs in Card JSON
- [x] #45 Air-Scrapping Expedition (215)
- [x] #47 Comet for Venus (218)
- [x] #53 Gyropolis (230)
- [x] #54 Hydrogen To Venus (231)
- [x] #58 Neutralizer Factory (240)

### Group 5: Missing City Placement Outputs
- [x] #3 Phobos Space Haven (021)
- [x] #49 Dawn City (220)
- [x] #57 Maxwell Base (238)
- [ ] #61 Stratopolis (248)

### Group 6: Missing Requirements in Card JSON
- [x] #64 Terraforming Contract (252)

### Group 7: Fix Io Sulphur Research Card Draw
- [ ] #55 Io Sulphur Research (232)

### Group 8: Fix Sulphur-Eating Bacteria Missing Choice
- [ ] #63 Sulphur-Eating Bacteria (251)

### Group 9: Fix Sponsored Academies
- [ ] #60 Sponsored Academies (247)

## TIER 2: Frontend Display Fixes

### Group 10: Missing "Any" Red Tint on Per-Condition Icons
- [ ] #17 Greenhouses (096)
- [ ] #25 Zeppelins (129)
- [ ] #34 Energy Saving (189)
- [ ] #44 Aerosport Tournament (214)

### Group 11: Minus/Negative Sign and Steal Display Issues
- [ ] #9 Virus (050)
- [ ] #10 Mining Expedition (063)
- [ ] #22 Sabotage (121)
- [ ] #23 Hired Raiders (124)
- [ ] #33 Flooding (188)
- [ ] #42 Small Asteroid (209)

### Group 12: Layout and Formatting Issues
- [ ] #5 Optimal Aerobraking (031)
- [ ] #32 Olympus Conference (185)
- [ ] #35 Invention Contest (192)
- [ ] #39 Underground City (032)
- [ ] #53 Gyropolis (230)
- [ ] #55 Io Sulphur Research (232)

### Group 13: Wrong or Missing Icons on Cards
- [ ] #2 Research Outpost (020)
- [ ] #3 Phobos Space Haven (021)
- [ ] #4 Predators (024)
- [ ] #6 Security Fleet (028)
- [ ] #7 Ants (035)
- [ ] #28 Herbivores (147)
- [ ] #43 Aerial Mappers (213)
- [ ] #46 Atmoscoop (217)
- [ ] #51 Extractor Balloons (223)
- [ ] #56 Local Shading (235)
- [ ] #58 Neutralizer Factory (240)
- [ ] #66 Venusian Plants (261)

### Group 14: Missing Venus Global Param Outputs
_(Cross-reference with Groups 4 and 13 — no additional cards)_

## TIER 3: New Backend Features

### Group 15: New Resource Type — Fighter
- [ ] #6 Security Fleet (028)

### Group 16: Variable-Amount User Selection
- [ ] #30 Insulation (152)
- [ ] #36 Power Infrastructure (194)

### Group 17: Temporary Effects (Next Card Only)
- [ ] #37 Indentured Workers (195)
- [ ] #40 Special Design (206)

### Group 18: Resource Storage as Payment
- [ ] #50 Dirigibles (222)
- [ ] #59 Rotator Impacts (243)
- [ ] #62 Stratospheric Birds (249)
- [ ] #63 Sulphur-Eating Bacteria (251)

### Group 19: Place Resource on Any Card's Storage
- [ ] #29 CEO's Favorite Project (149)
- [ ] #48 Corroder Suits (219)
- [ ] #52 Freyja Biodomes (227)
- [ ] #57 Maxwell Base (238)
- [ ] #66 Venusian Plants (261)

### Group 20: Conditional Effects
- [ ] #33 Flooding (188)
- [ ] #47 Comet for Venus (218)
- [ ] #55 Io Sulphur Research (232)

### Group 21: Card Discard and Opponent Draw
- [ ] #14 Mars University (073)
- [ ] #60 Sponsored Academies (247)

### Group 22: Tile Placement Restriction — Adjacent to N Tiles
- [ ] #21 Urbanized Area (120)
- [ ] #24 Ecological Zone (128)

### Group 23: New Tile Types
- [ ] #8 Natural Preserve (044)
- [ ] #11 Mining Area (064)
- [ ] #12 Mining Rights (067)
- [ ] #18 Nuclear Zone (097)
- [ ] #24 Ecological Zone (128)
- [ ] #27 Mohole Area (142)
- [ ] #38 Restricted Area (199)

### Group 24: Titanium as Payment for Actions
- [ ] #1 Water Import From Europa (012)
- [ ] #59 Rotator Impacts (243)

### Group 25: Land Claim Visual Indicator
- [ ] #13 Land Claim (066)

## Already Correct
- [x] #15 Mass Converter (094) — verified correct

## Later: Structured Card Descriptions

- [ ] Replace the card description string with an ordered list of `{ type, text }` sections; supported types: `generic`, `effect`, `action`, `requirement`.
- [ ] Render each section as its own paragraph. Generate bold **Effect:**, **Action:**, and **Requirement:** labels from the type; generic sections have no prefix.
- [ ] Separate trailing requirements from action/effect text (for example, Martian Zoo's two-city requirement). Preserve wording and order; do not infer gameplay rules from description sections.
- [ ] Migrate card data, backend types, generated DTOs, and description consumers together, replacing the old string representation outright. Verify Small/Large layouts and existing inline emphasis.
