import { expect, test, type Page } from "@playwright/test";

const gameId = "563c07b4-2352-4776-8d93-8196a12e063a";
function fixture() {
  return {
    id: gameId,
    status: "active",
    hostPlayerId: "alice",
    viewingPlayerId: "",
    generation: 2,
    settings: { mapId: "tharsis", cardPacks: ["base-game", "prelude"] },
    otherPlayers: [],
    resumeLobby: {
      seats: ["Alice", "Bob", "Charlie", "Dana", "Eve"].map((name, index) => ({
        id: name.toLowerCase(),
        name,
        savedName: name,
        playerType: "human",
        exited: false,
        claimed: index === 0,
        connected: index === 0,
        color: index === 0 ? "#1e88e5" : "#e53935",
        corporationId: index === 0 ? "tharsis-republic" : "phobolog",
        corporationName: index === 0 ? "Tharsis Republic" : "PhoboLog",
        botStatus: "",
        botError: "",
      })),
    },
  };
}
async function server(page: Page, owner = false) {
  const game = fixture();
  const commands: {
    type: string;
    payload: { seatId?: string; playerName?: string; playerId?: string };
  }[] = [];
  let viewer = "";
  let send: (message: string) => void = () => {};
  const control = {
    game,
    commands,
    conflict: false,
    hold: false,
    broadcast: () =>
      send(
        JSON.stringify({
          type: "game-updated",
          payload: { game: { ...game, viewingPlayerId: viewer } },
        }),
      ),
    releaseViewer: () => {
      game.resumeLobby.seats.find((s) => s.id === viewer)!.claimed = false;
      viewer = "";
      control.broadcast();
    },
  };
  if (owner) {
    await page.addInitScript(
      ({ gameId }) =>
        localStorage.setItem(
          "openmars.game",
          JSON.stringify({ gameId, playerId: "alice", playerName: "Alice" }),
        ),
      { gameId },
    );
  }
  await page.route("**/api/v1/meta", (route) => route.fulfill({ json: { version: "localbuild" } }));
  await page.route("**/game-options", (route) =>
    route.fulfill({ json: { availableMaps: [{ id: "tharsis", name: "Tharsis" }] } }),
  );
  await page.route(`**/games/${gameId}*`, (route) => route.fulfill({ json: { game } }));
  await page.route("**/games", (route) => route.fulfill({ json: { games: [game] } }));
  await page.routeWebSocket(/ws/, (socket) => {
    send = (message) => socket.send(message);
    socket.onMessage((raw) => {
      const message = JSON.parse(String(raw));
      commands.push(message);
      if (message.type === "player-connect") {
        viewer = message.payload.playerId;
      }
      if (message.type === "claim-resume-seat") {
        if (control.hold) {
          return;
        }
        const seat = game.resumeLobby.seats.find((s) => s.id === message.payload.seatId)!;
        if (control.conflict) {
          seat.claimed = true;
          seat.connected = true;
          send(
            JSON.stringify({
              type: "error",
              payload: {
                code: "seat_taken",
                message: "Go: internal private failure must never be shown",
              },
            }),
          );
          control.conflict = false;
          return;
        }
        seat.claimed = true;
        seat.connected = true;
        seat.name = message.payload.playerName;
        viewer = seat.id;
      }
      if (message.type === "release-resume-seat") {
        const seat = game.resumeLobby.seats.find((s) => s.id === message.payload.seatId)!;
        seat.claimed = false;
        seat.connected = false;
      }
      control.broadcast();
    });
  });
  return control;
}
async function visibleControl(page: Page, name: string) {
  const control = page.getByRole("button", { name, exact: true });
  await control.scrollIntoViewIfNeeded();
  await expect(control).toBeInViewport();
  expect(
    await control.evaluate((element) => {
      const r = element.getBoundingClientRect();
      return element.contains(document.elementFromPoint(r.x + r.width / 2, r.y + r.height / 2));
    }),
  ).toBe(true);
  return control;
}

test("resume joining handles a taken seat and reconnects without identity controls", async ({
  page,
  isMobile,
}, info) => {
  const mock = await server(page);
  await page.goto(`/resume/${gameId}?quick`);
  await expect(page.getByRole("heading", { name: "Join game" })).toBeVisible();
  await expect(page.getByPlaceholder("Your name")).toHaveValue("Bob");
  await page.screenshot({ path: info.outputPath("join.png"), animations: "disabled" });
  await page.getByRole("button", { name: "Choose your seat" }).click();
  const lastSeat = page.getByRole("option", { name: "Eve", exact: true });
  await lastSeat.scrollIntoViewIfNeeded();
  await expect(lastSeat).toBeInViewport();
  if (isMobile) {
    expect((await lastSeat.boundingBox())!.height).toBeGreaterThanOrEqual(44);
  }
  await page.screenshot({ path: info.outputPath("seats.png"), animations: "disabled" });
  await page.getByRole("listbox").press("Escape");
  mock.conflict = true;
  await page.getByRole("button", { name: "Join", exact: true }).click();
  await expect(page.getByRole("alert")).toHaveText("Seat already taken");
  await expect(page.getByRole("alert")).toHaveCSS("text-align", "center");
  await expect(page.getByRole("button", { name: "Join", exact: true })).toBeDisabled();
  await expect(page.getByText("Go: internal", { exact: false })).toHaveCount(0);
  await page.screenshot({ path: info.outputPath("taken.png"), animations: "disabled" });
  await page.getByRole("button", { name: "Choose your seat" }).click();
  await expect(page.getByRole("option", { name: "Bob", exact: true })).toHaveCount(0);
  await page.getByRole("option", { name: "Charlie", exact: true }).click();
  await expect(page.getByRole("alert")).toHaveCount(0);
  if (isMobile) {
    await page.setViewportSize({ width: 700, height: 260 });
    await page.getByPlaceholder("Your name").focus();
    await visibleControl(page, "Join");
    await expect(page.locator("main > header")).toBeInViewport();
    await page.screenshot({ path: info.outputPath("keyboard.png"), animations: "disabled" });
    await page.setViewportSize({ width: 700, height: 300 });
  }
  await page.getByPlaceholder("Your name").fill("Saffron");
  await page.getByRole("button", { name: "Join", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Resume game" })).toBeVisible();
  await expect(page.getByTestId("resume-player").filter({ hasText: "Saffron" })).toContainText(
    "YOU",
  );
  await expect(page.getByPlaceholder("Your name")).toHaveCount(0);
  await expect(page.getByRole("button", { name: /Play as|Update name|Release seat/ })).toHaveCount(
    0,
  );
  await expect(page.getByText("Choose your saved seat.", { exact: false })).toHaveCount(0);
  await visibleControl(page, "Copy invite link");
  await page.screenshot({ path: info.outputPath("joined.png"), animations: "disabled" });
  await page.reload();
  await expect(page.getByRole("heading", { name: "Resume game" })).toBeVisible();
  await expect(page.getByPlaceholder("Your name")).toHaveCount(0);
  expect(mock.commands.filter((m) => m.type === "claim-resume-seat")).toHaveLength(2);
  mock.releaseViewer();
  await expect(page.getByRole("heading", { name: "Join game" })).toBeVisible();
  await expect(page.getByRole("alert")).toHaveText("Seat released");
  expect(await page.evaluate(() => localStorage.getItem("openmars.game"))).toBeNull();
});

test("resume selection tracks live claims and does not treat unrelated broadcasts as join success", async ({
  page,
  isMobile,
}, info) => {
  const mock = await server(page);
  await page.goto(`/resume/${gameId}?quick`);
  await expect(page.getByPlaceholder("Your name")).toHaveValue("Bob");
  mock.game.resumeLobby.seats[1].claimed = true;
  mock.broadcast();
  await expect(page.getByRole("alert")).toHaveText("Seat already taken");
  await page.getByRole("button", { name: "Choose your seat" }).click();
  await page.getByRole("option", { name: "Charlie" }).click();
  mock.hold = true;
  await page.getByRole("button", { name: "Join", exact: true }).click();
  mock.broadcast();
  await expect(page.getByRole("button", { name: "Join", exact: true })).toBeDisabled();
  await expect(page.getByRole("heading", { name: "Join game" })).toBeVisible();
  mock.hold = false;
  await page.reload();
  await expect(page.getByPlaceholder("Your name")).toHaveValue("Charlie");
  for (const seat of mock.game.resumeLobby.seats) {
    seat.claimed = true;
  }
  mock.broadcast();
  await expect(page.getByText("No seats available", { exact: true })).toBeVisible();
  await visibleControl(page, "Cancel");
  await page.screenshot({ path: info.outputPath("full.png"), animations: "disabled" });
  if (isMobile) {
    await page.setViewportSize({ width: 700, height: 260 });
    await visibleControl(page, "Cancel");
    await expect(page.locator("main > header")).toBeInViewport();
  }
});

test("host lobby stays static with horizontal identities and separate recovery", async ({
  page,
  isMobile,
}, info) => {
  const mock = await server(page, true);
  mock.game.resumeLobby.seats[1].claimed = true;
  mock.game.resumeLobby.seats[1].connected = true;
  mock.game.resumeLobby.seats[2].name = "LongPlayerNameTwenty";
  await page.goto(`/resume/${gameId}?quick`);
  await expect(page.getByRole("heading", { name: "Resume game" })).toBeVisible();
  await expect(page.getByText("YOU", { exact: true })).toBeVisible();
  await expect(page.getByText("Waiting for Bob", { exact: false })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Resume game", exact: true })).toBeDisabled();
  await expect(page.getByPlaceholder("Your name")).toHaveCount(0);
  const row = page.getByTestId("resume-player").first();
  const alignment = await row.evaluate((element) => {
    const name = element.querySelector("span.font-orbitron")!.getBoundingClientRect();
    const logo = element.querySelector("img")!.getBoundingClientRect();
    return {
      gap: logo.x - name.right,
      offset: Math.abs(name.y + name.height / 2 - logo.y - logo.height / 2),
    };
  });
  expect(alignment.gap).toBeGreaterThan(0);
  expect(alignment.offset).toBeLessThan(2);
  const copy = (await (await visibleControl(page, "Copy invite link")).boundingBox())!;
  const resume = (await (await visibleControl(page, "Resume game")).boundingBox())!;
  expect(copy.x + copy.width).toBeLessThan(resume.x);
  expect(Math.abs(copy.y + copy.height / 2 - resume.y - resume.height / 2)).toBeLessThan(2);
  if (isMobile) {
    await expect(page.locator("main > header")).toBeInViewport();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(
      true,
    );
  }
  await page.screenshot({ path: info.outputPath("host.png"), animations: "disabled" });
  await (await visibleControl(page, "Manage seats")).click();
  await expect(page.getByRole("heading", { name: "Manage seats" })).toBeVisible();
  await page.getByRole("button", { name: "Seat to release" }).click();
  await expect(page.getByRole("option", { name: "Alice" })).toHaveCount(0);
  await page.getByRole("option", { name: "Bob" }).click();
  await page.screenshot({ path: info.outputPath("recovery.png"), animations: "disabled" });
  await (await visibleControl(page, "Release seat")).click();
  await expect(page.getByRole("heading", { name: "Resume game" })).toBeVisible();
  await expect(page.getByTestId("resume-player").filter({ hasText: "Bob" })).toContainText(
    "Available",
  );
  for (const seat of mock.game.resumeLobby.seats) {
    seat.claimed = true;
    seat.connected = true;
  }
  mock.broadcast();
  await expect(page.getByRole("button", { name: "Resume game", exact: true })).toBeEnabled();
  await page.getByRole("button", { name: "Resume game", exact: true }).click();
  await expect.poll(() => mock.commands.some((m) => m.type === "resume-game")).toBe(true);
});

test("Browse shows resume owner, readiness and Join, and game codes use seat selection", async ({
  page,
}, info) => {
  const mock = await server(page);
  mock.game.hostPlayerId = "bob";
  mock.game.resumeLobby.seats[0].claimed = false;
  mock.game.resumeLobby.seats[0].connected = false;
  mock.game.resumeLobby.seats[1].claimed = true;
  mock.game.resumeLobby.seats[1].connected = true;
  mock.game.resumeLobby.seats[1].name = "Saffron";
  await page.goto("/join?quick");
  await expect(page.getByText("Resume game", { exact: true })).toBeVisible();
  await expect(page.getByText("Saffron", { exact: true })).toBeVisible();
  await expect(page.getByText("1/5 ready", { exact: false })).toBeVisible();
  await expect(page.getByText("Unknown", { exact: true })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Spectate" })).toHaveCount(0);
  await page.getByPlaceholder("Search games...").fill("saffron");
  await expect(page.getByRole("button", { name: "Join", exact: true })).toBeVisible();
  await page.screenshot({ path: info.outputPath("browse.png"), animations: "disabled" });
  await page.getByRole("button", { name: "Join", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Join game" })).toBeVisible();
  await expect(page.getByPlaceholder("Your name")).toHaveValue("Alice");
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  expect(mock.commands.some((m) => m.type === "claim-resume-seat")).toBe(false);
  await page.goto(`/join?quick&code=${gameId}`);
  await expect(page.getByRole("heading", { name: "Join game" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Choose your seat" })).toBeVisible();
  await page.goto(`/game/${gameId}?quick&type=join`);
  await expect(page.getByRole("heading", { name: "Join game" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Choose your seat" })).toBeVisible();
});
