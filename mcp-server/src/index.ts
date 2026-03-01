import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { WsConnection } from "./connection.js";
import { GameState } from "./state.js";
import { registerTools } from "./tools.js";

const server = new McpServer({
  name: "terraforming-mars",
  version: "1.0.0",
});

const conn = new WsConnection();
const state = new GameState();

conn.onGameUpdated = (game) => {
  state.update(game);
};

conn.onFullState = (payload) => {
  state.myPlayerId = payload.playerId;
  if (payload.game) {
    state.update(payload.game);
  }
};

conn.onError = (payload) => {
  process.stderr.write(
    `[terraforming-mars-mcp] Error: ${payload.message}\n`,
  );
};

conn.onDisconnect = () => {
  process.stderr.write("[terraforming-mars-mcp] WebSocket disconnected\n");
};

registerTools(server, conn, state);

const transport = new StdioServerTransport();
await server.connect(transport);
