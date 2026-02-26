# Frontend - Terraforming Mars Game UI

React frontend with 3D Mars visualization using Three.js/React Three Fiber. Real-time multiplayer via WebSocket state synchronization from Go backend.

## Development Server Policy

**CRITICAL**: NEVER restart the frontend dev server yourself. ALWAYS ask the user if you think a restart is needed.

- Vite hot reload handles all code changes automatically
- Hot Module Replacement reloads React components, CSS, and TypeScript instantly
- Your role is to write code, not manage processes

## Architecture

### Directory Structure

```
frontend/
├── src/
│   ├── components/        # React components
│   │   ├── 3d/            # Three.js background components
│   │   ├── game/          # Core game UI (board, view, controls)
│   │   ├── layout/        # Layout components (panels, main)
│   │   ├── pages/         # Top-level page components
│   │   └── ui/            # Reusable UI components (cards, display, modals)
│   ├── contexts/          # React contexts for state management
│   ├── hooks/             # Custom React hooks
│   ├── services/          # API and WebSocket services
│   ├── types/             # TypeScript types (includes generated/ from backend)
│   └── utils/             # Utility functions and helpers
├── public/
│   ├── assets/            # Static game assets (images, icons)
│   └── models/            # 3D models for Three.js
```

### State Management

- **No local game state**: Backend is source of truth
- **Real-time sync**: All state changes via WebSocket events
- **Unidirectional flow**: Server → WebSocket → React state → UI
- **localStorage**: Game/player ID persistence for reconnection

## Key Development Patterns

### Code Style

- **CRITICAL**: NO unnecessary comments - code should be self-documenting. Avoid logic comments that just restate what the code does. Only add comments when truly necessary (complex algorithms, non-obvious business rules).

### Component Development

1. **Inspect existing design language** before creating new components
2. **Reuse over creation**: Check for existing components first
3. **No emojis in UI**: Use GameIcon component or assets instead
4. Use `void <function>()` to explicitly discard promises in event handlers

### Displaying Resources, Costs, and Gains

**CRITICAL**: ALWAYS reuse existing components for displaying game resources.

**GameIcon** - Single icon with optional value:

```tsx
<GameIcon iconType="credits" amount={25} size="medium" />
<GameIcon iconType="energy-production" amount={3} size="medium" />  // Auto brown background
```

Sizes: 'small' (24px), 'medium' (32px), 'large' (40px). Add new icons to `utils/` icon store.

**BehaviorSection** - Card behavior layouts:

```tsx
<BehaviorSection behavior={cardBehavior} />
```

Use for any display with multiple resources, costs/gains, or input/output relationships. See `components/ui/cards/BehaviorSection/CLAUDE.md` for details.

### Styling with Tailwind CSS v4

**CRITICAL**: Uses Tailwind CSS v4 with CSS-based configuration.

- **Configuration**: `@theme {}` block in main CSS file
- **NO CSS Modules**: NEVER create `.module.css` files
- **NO JavaScript Config**: `tailwind.config.js` is ignored

Custom theme values (colors, fonts, shadows) defined in `@theme {}` block.

### Typography & Fonts

The project uses two font styles:

**Game Font (Orbitron)** - `font-orbitron`
- Use for: titles, headers, numbers, labels, button text, UI elements
- Futuristic, bold, sci-fi aesthetic
- Available weights: 400-900
- Example: `className="font-orbitron font-bold"`

**Default Font (System)**
- Use for: long descriptions, card text, tooltips, explanatory content
- More readable for extended text
- Default if no font class specified

**Guidelines:**
- Numbers and statistics: Always use `font-orbitron`
- Short labels (1-3 words): Use `font-orbitron`
- Descriptions (sentences/paragraphs): Use default font
- Interactive elements (buttons): Use `font-orbitron`

### 3D Rendering

Uses React Three Fiber for Three.js integration. Hex coordinates use cube system (q, r, s) where q + r + s = 0, with utilities in `utils/`.

**Controls**: Custom pan/zoom (no orbital rotation), parallax background for depth.


### Sound System

```tsx
import { useSoundEffects } from '../hooks/useSoundEffects';

const { playSound, playProductionSound } = useSoundEffects();
void playProductionSound();
```

Add new sounds to `public/assets/audio/` and register in the audio service preload list.

### Notifications

Use `useNotifications()` from `contexts/NotificationContext.tsx` for error and warning messages. Notifications are displayed by `NotificationContainer` at the bottom-left of the screen.

- **Only shown outside game pages**: `NotificationContainer` returns `null` on `/game/*` routes, so notifications fired during gameplay are silently ignored. Use in-game UI (modals, overlays) for in-game feedback instead.
- **Types**: `"error"` (red) and `"warning"` (yellow).
- **Auto-dismiss**: Default 3000ms. Pass `duration: 0` for persistent notifications that require manual dismiss (e.g. "Server is down").

```tsx
const { showNotification } = useNotifications();

showNotification({ message: "Name too short", type: "error" });
showNotification({ message: "Server is down", type: "error", duration: 0 });
```

### WebSocket Communication

WebSocket service in `services/` handles real-time game state sync.

**Outbound**: `join-game`, `player-reconnect`, `select-corporation`, `skip-action`, `start-game`
**Inbound**: `game-updated`, `player-connected`, `player-reconnected`, `player-disconnected`

## Testing with Playwright

Use Playwright MCP tools for live debugging:

- `mcp__playwright__browser_navigate`: Navigate to URLs
- `mcp__playwright__browser_snapshot`: Capture page state
- `mcp__playwright__browser_click`: Click elements
- `mcp__playwright__browser_take_screenshot`: Capture visuals

## Important Notes

### State Management Rules

**CRITICAL**: No timeouts or arbitrary delays for state synchronization. Use proper WebSocket event handling.

### Design Principles

- **No emojis in UI**: Use GameIcon or assets
- **GameIcon first**: Never use direct `<img>` tags for game icons
- **Tailwind CSS only**: No CSS Modules
- **Type safety**: Always use generated types from backend

### Menu and Modal Components

**CRITICAL**: Reuse existing components for main menu and game modals.

**GameMenuButton** - All buttons in main menu and game modals:

```tsx
import GameMenuButton from "../buttons/GameMenuButton.tsx";

<GameMenuButton variant="primary" size="lg" onClick={handleClick}>START GAME</GameMenuButton>
<GameMenuButton variant="secondary" size="sm">Cancel</GameMenuButton>
<GameMenuButton variant="action" size="sm">Confirm</GameMenuButton>
<GameMenuButton variant="text" size="sm">Icon-only button</GameMenuButton>
```

Variants: `primary`, `secondary`, `action`, `text`, `toolbar`, `error`. Sizes: `sm`, `md`, `lg`.

**GameMenuModal** - All main menu modals and overlays:

```tsx
import GameMenuModal from "./GameMenuModal.tsx";

<GameMenuModal
  title="Modal Title"
  subtitle="Optional subtitle"
  onBack={() => navigate("/")}
  visible={isVisible}
>
  {/* Modal content */}
</GameMenuModal>
```

Provides consistent styling, animations, Back button (top-left), and Settings button (top-right).

## Related Documentation

- **Root CLAUDE.md**: Project overview and commands
- **backend/CLAUDE.md**: Backend architecture and event system
- **backend/assets/CLAUDE.md**: Card database documentation
