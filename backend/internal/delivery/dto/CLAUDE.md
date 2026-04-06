# Data Transfer Objects (DTOs) - Frontend Communication Layer

**CRITICAL**: DTOs in this directory are the ONLY data structures exposed to the frontend. They define the contract between backend and frontend.

## Core Principles

1. **Frontend Contract**: DTOs are "artifacts" created from internal models before sending to frontend
2. **Bidirectional**: Frontend sends DTOs to backend, backend sends DTOs to frontend
3. **Must Generate TypeScript**: ALL DTOs are exported to TypeScript types via `tygo generate`
4. **NEVER Export Models**: Only DTOs are allowed in `tygo.yaml` - NEVER add model objects

## Development Workflow

### When Adding/Modifying DTOs

1. **Create/Update DTO struct** in `game_dto.go` with `json:` tags
2. **Add `tstype:` tags only where needed** — tygo infers types from Go types automatically. Only use `tstype:` to override (e.g., string literal unions for discriminated unions)
3. **Add/Update mapper function** in `mapper_*.go` to convert model to DTO
4. **Run `make generate`** to create TypeScript types
5. **Verify TypeScript types** in `frontend/src/types/generated/api-types.ts`

### DTO Structure Template

```go
type PlayerDto struct {
    ID       string `json:"id"`
    Credits  int    `json:"credits"`
    IsActive bool   `json:"isActive"`
}
```

### Tag Reference

```go
`json:"fieldName"`              // Required: controls TS field name
`json:"fieldName,omitempty"`    // Makes field optional (?) in TS
`tstype:"'a' | 'b' | 'c'"`     // Override: generates literal union type
`tstype:"CustomType[]"`         // Override: use a custom TS type
```

**Do NOT use `ts:` tags** — they are ignored by tygo.

### Emitting Custom TypeScript

Use `//tygo:emit` directly above a struct to inject literal TypeScript (e.g., union type aliases):

```go
//tygo:emit export type MyUnion = TypeA | TypeB;
type TypeA struct { ... }
```

## Resource Condition DTOs

Card behavior conditions use a **discriminated union** pattern. 10 category-specific DTO structs generate TypeScript interfaces with literal `type` fields:

```go
type BasicResourceConditionDto struct {
    Type   string `json:"type" tstype:"'credit' | 'steel' | 'titanium' | 'plant' | 'energy' | 'heat'"`
    Amount int    `json:"amount"`
    Target TargetType `json:"target"`
    // ... category-specific fields only
}
```

The `//tygo:emit` directive generates the `ResourceCondition` union type. The mapper in `mapper_card.go` returns the right DTO struct per category.

## Synchronization Rules

**CRITICAL**: When model changes, DTO MUST be updated:

1. Model field added → Add to DTO with `json:` tag
2. Model field removed → Remove from DTO
3. Model field renamed → Update DTO field name
4. **ALWAYS** update mapper function in `mapper_*.go`
5. **ALWAYS** run `make generate` after DTO changes

## Common Mistakes to Avoid

- Adding `ts:` tags (tygo ignores them, only `tstype:` works for overrides)
- Adding models to `tygo.yaml` (only DTOs allowed)
- Using DTOs internally in backend (use domain models instead)
- Updating DTO without updating mapper function
- Not running `make generate` after DTO changes
- Adding `// DEPRECATED` comments — delete deprecated fields entirely

## File Organization

- `game_dto.go`: All DTO struct definitions (including resource condition category DTOs)
- `mapper_card.go`: Card and behavior condition mappers
- `mapper_game.go`: Game state mappers
- `mapper_player.go`: Player state mappers
- `mapper_helpers.go`: Shared mapping utilities
