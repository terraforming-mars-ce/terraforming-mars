# Shared Types - Behavior Condition System

Card behaviors use a **discriminated union** pattern. The `BehaviorCondition` interface has 10 typed category structs, each with only the fields that make sense for that category.

## Architecture

```
Card JSON (flat) → UnmarshalJSON → categorizeCondition() → Typed struct (e.g., *BasicResourceCondition)
                                                                    ↓
                                                        BehaviorCondition interface
                                                                    ↓
                                              behavior_applier type-switches per category
```

## Key Files

| File | Purpose |
|------|---------|
| `behavior_condition.go` | `BehaviorCondition` interface, `ConditionBase`, accessor helpers (`GetPerCondition`, `IsVariableAmount`, etc.) |
| `behavior_condition_categories.go` | 10 typed structs, `categorizeCondition`, `classifyResourceType` |
| `card_behavior_types.go` | `resourceConditionJSON` (flat JSON format), `Trigger`, `PerCondition`, `Choice` |
| `behavior.go` | `CardBehavior` with `[]BehaviorCondition` inputs/outputs, `UnmarshalJSON` |
| `resource_condition_rules.go` | Field validation profiles per resource type |
| `resource_type.go` | `ResourceType` constants |

## Categories

| Category Struct | Resource Types | Category-Specific Fields |
|----------------|----------------|------------------------|
| `BasicResourceCondition` | credit, steel, titanium, plant, energy, heat | Per, VariableAmount, MaxTrigger, Optional, TargetRestriction, PaymentAllowed |
| `ProductionCondition` | *-production, any-production | Per, VariableAmount |
| `TilePlacementCondition` | city/greenery/ocean/volcano/tile-placement, land-claim | TileRestrictions, TileType, Optional |
| `GlobalParameterCondition` | temperature, oxygen, ocean, venus, tr, global-parameter | Per |
| `CardOperationCondition` | card-draw/take/peek/buy/discard | Selectors, Optional, PaymentAllowed, VariableAmount |
| `CardStorageCondition` | microbe, animal, floater, science, asteroid, fighter, disease, card-resource | Selectors, Per, Optional, VariableAmount |
| `EffectCondition` | discount, payment-substitute, value-modifier, defense, etc. | Selectors, Temporary |
| `ColonyCondition` | colony-tile/count/bonus/track-step | AllowDuplicatePlayerColony |
| `TileModificationCondition` | tile-destruction/replacement | TileType |
| `MiscCondition` | extra-actions, bonus-tags, world-tree-tile, award-fund, trade | Per, Selectors |

All structs embed `ConditionBase` (ResourceType, Amount, Target).

## Adding a New Resource Type

1. Add constant to `resource_type.go` and add it to `AllResourceTypes`
2. Determine which category it belongs to (or create a new one)
3. Add to `classifyResourceType()` switch in `behavior_condition_categories.go`
4. Add to `categorizeCondition()` and `flattenCondition()` in the same file
5. Add field validation profile in `resource_condition_rules.go` (both input and output maps)
6. If new category: create the struct, implement `deepCopyCondition()` and `isBehaviorCondition()`
7. Add handler in `internal/game/cards/behavior_applier_*.go`
8. Add DTO struct in `internal/delivery/dto/game_dto.go` with `tstype:` literal union on `type` field
9. Update the `tygo:emit` union type alias and `toResourceConditionDto` mapper
10. Update frontend type guards in `frontend/src/types/resourceConditions.ts`
11. Run `make generate` and `make test`

## Adding a New Field to an Existing Category

1. Add field to the category struct in `behavior_condition_categories.go`
2. Add to `categorizeCondition()` (copy from JSON) and `flattenCondition()` (copy back)
3. Update `deepCopyCondition()` if the field is a pointer or slice
4. Add to the corresponding DTO struct in `game_dto.go`
5. Update `toResourceConditionDto` mapper in `mapper_card.go`
6. Update validation profile in `resource_condition_rules.go` if needed
7. Run `make generate` and `make test`

## Accessing Fields

Common fields via interface methods:
```go
bc.GetResourceType()  // ResourceType
bc.GetAmount()        // int
bc.GetTarget()        // string
```

Category-specific fields via accessor helpers:
```go
shared.GetPerCondition(bc)       // *PerCondition or nil
shared.IsVariableAmount(bc)      // bool
shared.GetSelectors(bc)          // []Selector or nil
shared.GetTileRestrictions(bc)   // *TileRestrictions or nil
shared.GetTemporary(bc)          // string
shared.IsOptional(bc)            // bool
shared.GetPaymentAllowed(bc)     // []ResourceType or nil
```

Or via type assertion when you know the category:
```go
if basic, ok := bc.(*shared.BasicResourceCondition); ok {
    basic.Per           // directly accessible
    basic.MaxTrigger    // directly accessible
}
```
