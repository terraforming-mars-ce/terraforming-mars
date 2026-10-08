---
name: web-nav
description: Navigate and interact with the Terraforming Mars frontend using Playwright CLI. This skill should be used when the user wants to view, test, or interact with the web application in a browser. Supports headless browser navigation to known pages (cards, create, join, game) and custom URLs. Defaults to localhost:3000.
---

# Web Navigation Skill

Navigate and interact with the Terraforming Mars web frontend using the `playwright-cli` command-line tool.

## Prerequisites

Install playwright-cli globally:

```bash
npm install -g @playwright/cli@latest
```

If browser navigation fails with an error about the browser not being installed, run:

```bash
playwright-cli install-browser
```

## Default Configuration

- **Base URL**: `http://localhost:3000` (do not attempt to start the frontend if not running)
- **Mode**: Headless by default (use `--headed` to see the browser)
- **Session**: Persistent profile preserved between calls

## Known Pages

When the user requests navigation to a page by name (not a full URL), use these routes:

| Page Name | Route | Description |
|-----------|-------|-------------|
| home, landing, main | `/` | Main landing page with "New Game" and "Browse" buttons |
| create, new game | `/create` | Game creation page with settings |
| join, browse, lobby | `/join` | Browse available games |
| cards, card browser | `/cards` | Card browser/explorer with filtering and search |
| reconnecting | `/reconnecting` | Reconnection page for in-progress games |
| game | `/game` or `/game/:gameId` | Main game interface |

## Navigation Workflow

1. **Check if frontend is running**: Before navigating, the frontend should already be running at `localhost:3000`. Do not attempt to start it automatically.

2. **Navigate to the page**: Use `playwright-cli open <url>` or `playwright-cli goto <url>`

3. **Take a snapshot**: After navigation, use `playwright-cli snapshot` to capture the page state and get element references for interaction.

## CLI Commands Reference

### Browser Management
```bash
playwright-cli open [url]              # Launch browser, optionally navigate to URL
playwright-cli goto <url>              # Navigate to a URL
playwright-cli close                   # Close the page
playwright-cli resize <width> <height> # Resize browser window
```

### Page Interaction
```bash
playwright-cli snapshot                # Capture snapshot (get element refs)
playwright-cli screenshot [ref]        # Capture screenshot
playwright-cli click <ref> [button]    # Click element (button: left|right|middle)
playwright-cli dblclick <ref>          # Double-click element
playwright-cli fill <ref> <text>       # Fill text into input field
playwright-cli type <text>             # Type text into focused element
playwright-cli hover <ref>             # Hover over element
playwright-cli select <ref> <value>    # Select dropdown option
playwright-cli check <ref>             # Check checkbox/radio
playwright-cli uncheck <ref>           # Uncheck checkbox/radio
playwright-cli upload <file>           # Upload file(s)
playwright-cli drag <startRef> <endRef> # Drag and drop
```

### Navigation
```bash
playwright-cli go-back                 # Go back in history
playwright-cli go-forward              # Go forward in history
playwright-cli reload                  # Reload page
```

### Keyboard
```bash
playwright-cli press <key>             # Press key (e.g., Enter, Tab, Escape)
```

### Debugging
```bash
playwright-cli console [level]         # View console messages (error|warning|info|debug)
playwright-cli network                 # View network requests
playwright-cli eval <function> [ref]   # Run JavaScript on page/element
```

### Tabs
```bash
playwright-cli tab-list                # List all tabs
playwright-cli tab-new [url]           # Create new tab
playwright-cli tab-select <index>      # Select tab by index
playwright-cli tab-close [index]       # Close tab
```

### Session Management
```bash
playwright-cli list                    # List all sessions
playwright-cli close-all               # Close all browsers
playwright-cli kill-all                # Force-kill all processes
```

### Headed Mode
```bash
playwright-cli open <url> --headed     # Open with visible browser window
```

## Example Usage

### Navigate to cards page
```bash
playwright-cli open http://localhost:3000/cards
playwright-cli snapshot
```

### Navigate to create game
```bash
playwright-cli goto http://localhost:3000/create
playwright-cli snapshot
```

### Navigate to specific game
```bash
playwright-cli goto http://localhost:3000/game/ABC123
playwright-cli snapshot
```

### Click a button (after getting ref from snapshot)
```bash
playwright-cli click "New Game button"
```

### Fill a form field
```bash
playwright-cli fill "Player name input" "Alice"
```

### Take a screenshot
```bash
playwright-cli screenshot
```

### View in headed mode (visible browser)
```bash
playwright-cli open http://localhost:3000 --headed
```

## Troubleshooting

- **Browser not installed**: Run `playwright-cli install-browser`
- **Connection refused**: Frontend is not running - inform the user they need to start it with `just frontend dev` or `just dev`
- **Page not found**: Check if the route exists in the known pages table
- **Session issues**: Use `playwright-cli kill-all` to reset all sessions

## Sources

- [microsoft/playwright-cli](https://github.com/microsoft/playwright-cli)
