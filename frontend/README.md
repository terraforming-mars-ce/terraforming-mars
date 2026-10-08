# Terraforming Mars Frontend

React frontend for the digital implementation of Terraforming Mars board game with 3D game view and real-time multiplayer functionality.

## Features

- **3D Mars Visualization**: Interactive 3D Mars sphere using React Three Fiber
- **Hexagonal Grid System**: 42-hex Mars board with tile placement mechanics
- **Real-time Multiplayer**: WebSocket integration with Go backend
- **Create/Join Game Flow**: Complete game lobby system with localStorage persistence
- **Corporation Selection**: Choose from multiple corporations with unique abilities
- **Resource Management**: Visual resource tracking with production displays
- **Card System**: Comprehensive card engine with effects and actions
- **Responsive UI**: Adaptive design for different screen sizes

## Quick Start

### Development Mode

```bash
just frontend dev
```

Runs the app in development mode. Opens [http://localhost:3000](http://localhost:3000) to view it in your browser.

### Production Build

```bash
just frontend build
```

Builds the app for production to the `build` folder with optimizations.

### Code Quality

```bash
just frontend check   # format-check, oxlint, typecheck and asset checks
just frontend format  # Format code with oxfmt
```

## Architecture

### Component Structure

```
src/
├── components/
│   ├── game/              # 3D game view components
│   │   ├── board/         # Mars board, hex tiles, sphere
│   │   └── view/          # Game3DView with Three.js
│   ├── layout/            # Main layout components
│   │   ├── main/          # GameInterface, GameLayout
│   │   └── panels/        # Sidebars, menu bar
│   ├── pages/             # Route pages (Create, Landing)
│   └── ui/                # Reusable UI components
│       ├── display/       # Cost, production displays
│       ├── modals/        # Game modals and popups
│       └── overlay/       # Resource bar, card hand
├── services/              # API and WebSocket services
├── types/
│   ├── generated/         # Auto-generated from Go backend
│   └── ...               # Frontend-specific types
└── contexts/             # React contexts for state
```

### Key Technologies

- **React 18** with TypeScript for robust component development
- **React Three Fiber** for 3D Mars visualization and hex grid
- **React Router** for client-side routing (Create/Join/Game)
- **WebSocket** for real-time multiplayer communication
- **Generated Types** from Go backend for type safety

### Services Architecture

#### API Service (`apiService.ts`)

- **REST API**: Game creation, retrieval, joining
- **Type Safety**: Uses generated `Game` and `GameSettings` types
- **Error Handling**: Comprehensive error handling and logging

#### WebSocket Service (`webSocketService.ts`)

- **Real-time Communication**: Bidirectional game state updates
- **Event System**: Flexible event listener pattern
- **Auto-reconnection**: Handles connection drops gracefully
- **Message Types**: Uses generated WebSocket message types

### State Management

- **Generated Types**: Full type safety via Go struct generation
- **React Context**: MainContentContext for UI state
- **localStorage**: Game persistence for reconnection
- **WebSocket State**: Real-time game state synchronization

## Development Workflow

### Type Generation

The frontend uses types automatically generated from the Go backend:

```bash
# In backend directory
tygo generate
```

This updates `src/types/generated/` with latest backend types.

### Code Quality Standards

- **oxlint**: Fast linting for code quality and consistency
- **Prettier**: Automatic code formatting
- **TypeScript**: Strict typing with generated backend types
- **No Console Logs**: Use `console.warn` or `console.error` only

### 3D Development

- **React Three Fiber**: Declarative Three.js in React
- **Hex Coordinates**: Cube coordinate system (q, r, s)
- **Custom Controls**: Pan/zoom only (no orbit rotation)
- **Performance**: Optimized for smooth 60fps gameplay

## Game Features

### Lobby System

- **Create Game**: Name input with validation and error handling
- **Join Game**: Game code entry with validation
- **Auto-reconnect**: localStorage persistence across page reloads
- **Landing Page**: Clean interface for game creation/joining

### 3D Game View

- **Mars Sphere**: Realistic Mars terrain with hex overlay
- **Hex Grid**: 42 interactive hexagonal tiles
- **Pan Controls**: Mouse/touch controls for view manipulation
- **Visual Feedback**: Tile highlighting and selection states

### Real-time Multiplayer

- **WebSocket Integration**: Instant game state updates
- **Player Synchronization**: All players see consistent game state
- **Corporation Selection**: Real-time corporation choice updates
- **Turn Management**: Synchronized turn-based gameplay

### UI Components

#### Resource Displays

- **CostDisplay**: Megacredit amounts with official assets
- **ProductionDisplay**: Production values with resource icons
- **Resource Bar**: Live resource tracking with click interactions

#### Card System

- **Hand Overlay**: Hearthstone-style card fan with animations
- **Card Effects**: Visual effects system for card abilities
- **Drag & Drop**: Card interaction with smooth animations

## Configuration

### Environment Setup

- **Development**: Auto-connects to `ws://localhost:3001`
- **Backend Integration**: Go server on port 3001
- **Type Sync**: Automatic TypeScript generation from Go structs

### Build Configuration

- **Vite**: Fast development server and build tool
- **TypeScript**: Strict mode with generated type checking
- **Asset Optimization**: Automatic image and bundle optimization

## Contributing

1. **Follow TypeScript patterns**: Use generated types from backend
2. **Maintain 3D performance**: Keep 60fps in 3D views
3. **Test WebSocket integration**: Verify real-time functionality
4. **Use existing UI components**: Leverage CostDisplay, ProductionDisplay
5. **Run quality checks**: `just frontend check` (or `just prepare-for-commit` from the root)

## Troubleshooting

### WebSocket Connection Issues

- Ensure Go backend is running on port 3001
- Check browser console for connection errors
- Verify WebSocket URL in service configuration

### Type Errors

- Run `tygo generate` in backend directory
- Check that generated types are imported correctly
- Ensure backend and frontend type versions match

### 3D Performance Issues

- Check browser WebGL support
- Monitor frame rate in development tools
- Optimize Three.js scene complexity if needed
