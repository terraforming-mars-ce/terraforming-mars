# BehaviorSection - Card Behavior Display System

Transforms card behavior data from the backend into visually organized UI presentations. Handles card effects, resource costs/gains, triggered conditions, production chains, and player choices.

## Architecture

### Rendering Pipeline

1. **Classification**: Behaviors assigned semantic types (discount, manual-action, triggered-effect, immediate-production)
2. **Merging**: Related behaviors consolidated to reduce visual clutter
3. **Space Analysis**: Estimates space requirements, forces compact mode if needed
4. **Layout Routing**: Each classified behavior routed to specialized layout handler
5. **Rendering**: Layout handlers use primitive components for final display

### Component Hierarchy

**Container Level**

- BehaviorSection: Main orchestrator
- BehaviorContainer: Type-specific styling (gradients, borders)

**Layout Handlers**

- ManualActionLayout: Player-activated actions with choice support (OR alternatives)
- TriggeredEffectLayout: Passive effects with trigger conditions (trigger : effect)
- ImmediateResourceLayout: Production and immediate effects

**Rendering Primitives**

- ResourceDisplay: Individual resource renderer with affordability logic
- BehaviorIcon: Context-aware sizing and styling
- CardIcon: Card operations (draw, peek, take, buy)

## Key Concepts

### Space Optimization

- If behaviors exceed MAX_CARD_ROWS (4), forces compact mode
- Compact mode uses "NxIcon" format instead of individual icons
- Consistency rule: If ANY resource uses number mode, ALL use it (except amount=1)

### Context-Aware Rendering

Different contexts affect icon sizing:

- **Standalone**: Large icons for tiles/cards (36px)
- **Action**: Medium icons with inputs (26px)
- **Production**: Brown box context (26px)

### Tile Scaling

- 2x if completely alone
- 1.5x if single tile in behavior
- 1.25x if pair of tiles
- 1x otherwise

### Affordability

- Unaffordable actions rendered in grayscale
- Receives `isResourceAffordable` function that checks player resources

### Behavior Merging

- Multiple "auto-no-background" behaviors merged into single display
- Does NOT merge behaviors with trigger conditions (preserves clarity)

## Extending the System

### Adding New Trigger Types

1. Add detection in behavior classifier
2. Create layout handler following existing patterns
3. Register in BehaviorSection switch statement
4. Define visual styling in BehaviorContainer

### Adding New Resource Types

1. Add to icon store in `utils/`
2. Update classification if special display rules needed
3. Update ResourceDisplay if special affordability/styling needed

## Visual Conventions

### Separators

- **Arrow (→)**: Inputs to outputs in manual actions
- **Colon (:)**: Trigger condition to effect
- **Slash (/)**: Alternative resources (steel/titanium)
- **Minus/Plus (-/+)**: Negative/positive production

### Visual Hierarchy

- Gradients distinguish behavior types (blue=actions, brown=production, transparent=triggered)
- Color coding: affordability, attack indicators, production vs. immediate

## Related Documentation

- **frontend/CLAUDE.md**: Frontend architecture and component standards
- **backend/assets/CLAUDE.md**: Card behavior data structures
