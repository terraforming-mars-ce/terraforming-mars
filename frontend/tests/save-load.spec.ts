import { expect, test, type Page } from "@playwright/test";
declare global {
  interface Window {
    saveFileReads: number;
    finishSaveRead: () => void;
  }
}
const summary = {
  generation: 4,
  mapId: "amazonis",
  phase: "action",
  savedAt: "2026-10-10T10:00:00Z",
  historyEntries: 42,
  logEntries: 12,
  seats: ["Alice", "Bob", "Charlie", "Dana", "Eve"].map((name) => ({
    id: name.toLowerCase(),
    name,
    savedName: name,
    playerType: "human",
    exited: false,
  })),
};
async function openLoad(page: Page) {
  await page.route("**/api/v1/meta", (r) => r.fulfill({ json: { version: "localbuild" } }));
  await page.route("**/game-options", (r) =>
    r.fulfill({ json: { availableMaps: [{ id: "amazonis", name: "Amazonis Planitia" }] } }),
  );
  await page.goto("/load?quick");
  await expect(page.getByRole("heading", { name: "Load game" })).toBeVisible();
  await reachable(page, "Back");
}
async function reachable(page: Page, name: string) {
  const control = page.getByRole("button", { name, exact: true });
  await control.scrollIntoViewIfNeeded();
  await expect(control).toBeInViewport();
  expect((await control.boundingBox())?.height).toBeGreaterThanOrEqual(32);
  expect(
    await control.evaluate((element) => {
      const rect = element.getBoundingClientRect();
      return element.contains(
        document.elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2),
      );
    }),
  ).toBe(true);
  return control;
}
test("picker rejects, resets, reselects and displays valid filenames", async ({ page }, info) => {
  let requests = 0;
  await page.route("**/game-saves/validate", (r) => {
    requests++;
    return r.request().postData() === "[]"
      ? r.fulfill({
          status: 400,
          json: {
            code: "invalid_save",
            message: "json: cannot unmarshal array into Go value of type save.Document",
          },
        })
      : r.fulfill({ json: summary });
  });
  await openLoad(page);
  await page.screenshot({ animations: "disabled", path: info.outputPath("initial.png") });
  const input = page.locator("input[type=file]");
  for (let i = 0; i < 2; i++) {
    await input.setInputFiles({
      name: "awards.json",
      mimeType: "application/json",
      buffer: Buffer.from("[]"),
    });
    await expect(page.getByRole("alert")).toContainText("Invalid file");
    await expect(
      page.getByRole("button", { name: "Select an exported Game file", exact: true }),
    ).toBeVisible();
    await expect(page.getByText(/cannot unmarshal|save.Document/)).toHaveCount(0);
  }
  expect(requests).toBe(2);
  await page.screenshot({ animations: "disabled", path: info.outputPath("invalid.png") });
  const filename = "exported-game-".repeat(8) + ".json";
  await input.setInputFiles({
    name: filename,
    mimeType: "application/json",
    buffer: Buffer.from('{"seed":"18446744073709551615"}'),
  });
  const picker = page.getByRole("button", { name: filename, exact: true });
  await expect(picker).toBeVisible();
  await expect(picker).not.toHaveAttribute("title");
  expect(await picker.evaluate((e) => getComputedStyle(e).fontFamily)).toContain("Orbitron");
  await expect(page.getByLabel("Your name", { exact: true })).toHaveValue("Alice");
  await expect(page.getByText("Amazonis Planitia", { exact: true })).toBeVisible();
  await expect(page.getByText(/log entries|replay snapshots|^Saved /)).toHaveCount(0);
  await expect(page.getByText("Your name", { exact: true })).toHaveCount(0);
  await expect(page.getByLabel("Your name", { exact: true })).toHaveAttribute(
    "placeholder",
    "Your name",
  );
  const seats = page.getByRole("button", { name: "Choose your seat", exact: true });
  expect(await seats.evaluate((element) => getComputedStyle(element).fontFamily)).toContain(
    "Orbitron",
  );
  await seats.click();
  await page.getByRole("option", { name: "Eve", exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ animations: "disabled", path: info.outputPath("seats.png") });
  await page.getByRole("option", { name: "Eve", exact: true }).click();
  await expect(page.getByLabel("Your name", { exact: true })).toHaveValue("Eve");
  await reachable(page, "Load game");
  await page.screenshot({ animations: "disabled", path: info.outputPath("ready.png") });
});
test("bad metadata prevents file reads and uploads", async ({ page }) => {
  let requests = 0;
  await page.route("**/game-saves/validate", (r) => {
    requests++;
    return r.fulfill({ json: summary });
  });
  await page.addInitScript(() => {
    Object.assign(window, { saveFileReads: 0 });
    File.prototype.text = async function () {
      (window as Window & { saveFileReads: number }).saveFileReads++;
      return "{}";
    };
  });
  await openLoad(page);
  for (const [name, size, message] of [
    ["huge.json", 128 * 1024 * 1024 + 1, "128 MiB"],
    ["a".repeat(251) + ".json", 2, "Filename too long"],
    ["é".repeat(126) + ".json", 2, "Filename too long"],
    ["bad\u202ename.json", 2, "Invalid filename"],
    ["wrong.txt", 2, "Select a .json file"],
    ["empty.json", 0, "Empty file"],
  ] as const) {
    await page.locator("input[type=file]").evaluate(
      (element, data) => {
        const file = new File(["{}"], data.name, { type: "application/json" });
        Object.defineProperty(file, "size", { value: data.size });
        const transfer = new DataTransfer();
        transfer.items.add(file);
        (element as HTMLInputElement).files = transfer.files;
        element.dispatchEvent(new Event("change", { bubbles: true }));
      },
      { name, size },
    );
    await expect(page.getByRole("alert")).toContainText(message);
  }
  expect(requests).toBe(0);
  expect(
    await page.evaluate(() => (window as Window & { saveFileReads: number }).saveFileReads),
  ).toBe(0);
});
test("paste expands with safe errors and reachable compact controls", async ({
  page,
  isMobile,
}, info) => {
  await page.route("**/game-saves/validate", (r) =>
    r.fulfill({ status: 400, json: { code: "invalid_json", message: "$: invalid character 'a'" } }),
  );
  await openLoad(page);
  const initial = await page.locator("main").boundingBox();
  await page.getByRole("button", { name: "Paste JSON", exact: true }).click();
  const editor = page.getByRole("textbox", { name: "Game save JSON" });
  const checkCompactFit = async () => {
    if (!isMobile) return;
    await page.waitForFunction(() =>
      document
        .getAnimations()
        .every(
          (a) => a.playState !== "running" || a.effect?.getComputedTiming().iterations === Infinity,
        ),
    );
    const check = page.getByRole("button", { name: "Check save", exact: true });
    const useFile = page.getByRole("button", { name: "Use file", exact: true });
    for (const control of [editor, check, useFile])
      await expect(control).toBeInViewport({ ratio: 1 });
    expect((await check.boundingBox())!.y).toBeCloseTo((await useFile.boundingBox())!.y, 0);
    expect((await editor.boundingBox())!.height).toBeGreaterThan(24);
    expect(
      await page.locator("main > div").evaluate((body) => body.scrollHeight <= body.clientHeight),
    ).toBe(true);
  };
  if (isMobile) {
    for (const height of [300, 260]) {
      await page.setViewportSize({ width: page.viewportSize()!.width, height });
      await checkCompactFit();
    }
  }
  await editor.fill("asd");
  await (await reachable(page, "Check save")).click();
  await expect(page.getByRole("alert")).toContainText("Invalid JSON");
  await expect(page.getByRole("alert")).not.toContainText("invalid character");
  await checkCompactFit();
  if (isMobile) {
    await expect(page.locator("main > header")).toBeInViewport();
    const box = await page.locator("main").boundingBox();
    expect(box?.x).toBe(0);
    expect(box?.width).toBe(page.viewportSize()?.width);
    await editor.focus();
    await page.setViewportSize({ width: page.viewportSize()!.width, height: 260 });
    await checkCompactFit();
    await expect(page.getByRole("alert")).toBeInViewport({ ratio: 1 });
    await expect(page.locator("main > header")).toBeInViewport();
  } else {
    await expect
      .poll(async () => (await page.locator("main").boundingBox())!.width)
      .toBeGreaterThan(initial!.width);
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
    true,
  );
  await page.screenshot({ animations: "disabled", path: info.outputPath("paste-error.png") });
  await page.emulateMedia({ reducedMotion: "reduce" });
  await (await reachable(page, "Use file")).click();
  await expect(
    page.getByRole("button", { name: "Select an exported Game file", exact: true }),
  ).toBeVisible();
});
test("newer selection wins over a slow file read", async ({ page }) => {
  await page.route("**/game-saves/validate", (r) => r.fulfill({ json: summary }));
  await page.addInitScript(() => {
    const original = File.prototype.text;
    File.prototype.text = function () {
      if (this.name === "slow.json")
        return new Promise<string>((resolve) => {
          Object.assign(window, { finishSaveRead: () => resolve("[]") });
        });
      return original.call(this);
    };
  });
  await openLoad(page);
  const input = page.locator("input[type=file]");
  await input.setInputFiles({
    name: "slow.json",
    mimeType: "application/json",
    buffer: Buffer.from("[]"),
  });
  await expect(page.getByRole("status")).toContainText("Reading");
  await input.setInputFiles({
    name: "latest.json",
    mimeType: "application/json",
    buffer: Buffer.from("{}"),
  });
  await expect(page.getByRole("button", { name: "latest.json", exact: true })).toBeVisible();
  await page.evaluate(() => (window as Window & { finishSaveRead: () => void }).finishSaveRead());
  await expect(page.getByRole("button", { name: "latest.json", exact: true })).toBeVisible();
  await expect(page.getByRole("alert")).toHaveCount(0);
});

test("mode changes animate width and height together with and without errors", async ({
  page,
  isMobile,
}) => {
  test.skip(isMobile, "Compact layout remains full screen");
  await page.route("**/game-saves/validate", (r) =>
    r.fulfill({ status: 400, json: { code: "invalid_save" } }),
  );
  await openLoad(page);
  await page.evaluate(() => document.fonts.ready);
  for (const withError of [false, true]) {
    if (withError) {
      await page.locator("input[type=file]").setInputFiles({
        name: "awards.json",
        mimeType: "application/json",
        buffer: Buffer.from("[]"),
      });
      await expect(page.getByRole("alert")).toBeVisible();
    }
    for (const name of ["Paste JSON", "Use file"]) {
      if (withError && name === "Use file") {
        await page.getByRole("textbox", { name: "Game save JSON" }).fill("[]");
        await page.getByRole("button", { name: "Check save", exact: true }).click();
        await expect(page.getByRole("alert")).toBeVisible();
      }
      await page.waitForFunction(() =>
        document
          .getAnimations()
          .every(
            (a) =>
              a.playState !== "running" || a.effect?.getComputedTiming().iterations === Infinity,
          ),
      );
      const samples = await page
        .getByRole("button", { name, exact: true })
        .evaluate(async (element) => {
          const panel = document.querySelector("main .game-panel")!;
          const first = panel.getBoundingClientRect();
          const points = [{ time: 0, width: first.width, height: first.height }];
          const start = performance.now();
          (element as HTMLButtonElement).click();
          await new Promise<void>((resolve) => {
            const sample = () => {
              const rect = panel.getBoundingClientRect();
              const time = performance.now() - start;
              points.push({ time, width: rect.width, height: rect.height });
              if (time < 400) requestAnimationFrame(sample);
              else resolve();
            };
            requestAnimationFrame(sample);
          });
          return points;
        });
      const widthStart = samples.find((s) => Math.abs(s.width - samples[0].width) > 1)?.time;
      const heightStart = samples.find((s) => Math.abs(s.height - samples[0].height) > 1)?.time;
      expect(widthStart).toBeDefined();
      expect(heightStart).toBeDefined();
      expect(Math.abs(widthStart! - heightStart!)).toBeLessThan(50);
      const end = samples.at(-1)!;
      expect(
        samples
          .filter((s) => s.time > 300)
          .every((s) => Math.abs(s.width - end.width) < 1 && Math.abs(s.height - end.height) < 1),
      ).toBe(true);
    }
  }
});

test("landing alignment and resume lobby remain visible", async ({ page, isMobile }, info) => {
  await page.route("**/api/v1/meta", (r) => r.fulfill({ json: { version: "localbuild" } }));
  await page.goto("/?quick");
  const logo = page.getByRole("img", { name: "Open Mars", exact: true });
  const load = page.getByRole("link", { name: "Load game", exact: true });
  await expect(load).toBeVisible();
  const logoBox = (await logo.boundingBox())!;
  const buttonBox = (await load.boundingBox())!;
  if (isMobile) {
    const createBox = (await page
      .getByRole("link", { name: "New game", exact: true })
      .boundingBox())!;
    const browseBox = (await page
      .getByRole("link", { name: "Browse", exact: true })
      .boundingBox())!;
    expect(createBox.y).toBeGreaterThanOrEqual(logoBox.y + logoBox.height);
    expect(buttonBox.y).toBeGreaterThanOrEqual(createBox.y + createBox.height);
    expect(buttonBox.y).toBeCloseTo(browseBox.y, 0);
    expect(buttonBox.x).toBeGreaterThan(browseBox.x + browseBox.width);
  } else {
    expect(
      Math.abs(logoBox.x + logoBox.width / 2 - buttonBox.x - buttonBox.width / 2),
    ).toBeLessThan(2);
  }
  await page.screenshot({ animations: "disabled", path: info.outputPath("landing.png") });
  const game = {
    id: "saved-game",
    hostPlayerId: "alice",
    viewingPlayerId: "",
    settings: {},
    resumeLobby: {
      seats: [{ ...summary.seats[0], claimed: false, connected: false, color: "#42a5f5" }],
    },
  };
  await page.route("**/games/saved-game", (r) => r.fulfill({ json: { game } }));
  await page.routeWebSocket(/ws/, (socket) => {
    socket.onMessage(() =>
      socket.send(JSON.stringify({ type: "game-updated", payload: { game } })),
    );
  });
  await page.goto("/resume/saved-game?quick");
  await expect(page.getByRole("heading", { name: "Join game" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Join", exact: true })).toBeEnabled();
  await reachable(page, "Back");
  await reachable(page, "Choose your seat");
  await reachable(page, "Join");
  if (isMobile) await expect(page.locator("main > header")).toBeInViewport();
  await page.screenshot({ animations: "disabled", path: info.outputPath("resume.png") });
});

test("compact landing fits above a flat install bar, including a saved session", async ({
  page,
  isMobile,
}, info) => {
  test.skip(!isMobile, "Install bar is compact only");
  test.setTimeout(60_000);
  await page.route("**/api/v1/meta", (r) => r.fulfill({ json: { version: "localbuild" } }));
  const game = {
    id: "saved-game",
    currentPhase: "action",
    generation: 4,
    currentPlayer: {},
    otherPlayers: [{}],
    settings: { maxPlayers: 2 },
  };
  await page.route("**/games/saved-game**", (r) => r.fulfill({ json: { game } }));
  for (const saved of [false, true]) {
    await page.goto("/?quick");
    await page.evaluate((hasSession) => {
      if (hasSession)
        localStorage.setItem(
          "openmars.game",
          JSON.stringify({ gameId: "saved-game", playerId: "alice", playerName: "Alice" }),
        );
      else localStorage.removeItem("openmars.game");
      localStorage.removeItem("openmars.installBar.dismissedUntil");
    }, saved);
    await page.reload();
    await expect(page.getByRole("link", { name: "Load game", exact: true })).toBeVisible();
    await page.evaluate(() => {
      const event = new Event("beforeinstallprompt", { cancelable: true });
      Object.assign(event, {
        prompt: async () => {},
        userChoice: Promise.resolve({ outcome: "dismissed", platform: "web" }),
      });
      window.dispatchEvent(event);
    });
    const bar = page.getByRole("region", { name: "Install app", exact: true });
    await expect(bar).toBeVisible();
    await expect(bar).not.toHaveClass(/game-panel/);
    for (const viewport of [
      { width: 700, height: 300 },
      { width: 568, height: 320 },
    ]) {
      await page.setViewportSize(viewport);
      const bounds = (await bar.boundingBox())!;
      expect(bounds.x).toBe(0);
      expect(bounds.width).toBe(viewport.width);
      expect(bounds.y + bounds.height).toBeCloseTo(viewport.height, 0);
      for (const name of ["New game", "Browse", "Load game"]) {
        const button = page.getByRole("link", { name, exact: true });
        const box = (await button.boundingBox())!;
        expect(box.y).toBeGreaterThanOrEqual(0);
        expect(box.y + box.height).toBeLessThanOrEqual(bounds.y);
        await button.click({ trial: true });
      }
      if (saved) {
        await page.getByRole("button", { name: "RECONNECT", exact: true }).click({ trial: true });
        const card = (await page.locator(".game-panel").boundingBox())!;
        const menu = (await page
          .getByRole("link", { name: "New game", exact: true })
          .boundingBox())!;
        expect(card.x).toBeGreaterThanOrEqual(24);
        expect(card.x).toBeGreaterThan(menu.x + menu.width);
      }
      expect(
        await page.evaluate(
          () =>
            document.documentElement.scrollHeight <= innerHeight &&
            document.documentElement.scrollWidth <= innerWidth,
        ),
      ).toBe(true);
      await page.screenshot({
        animations: "disabled",
        path: info.outputPath(`landing-${viewport.width}-${saved ? "saved" : "new"}.png`),
      });
    }
    await page.getByRole("button", { name: "Not now", exact: true }).click();
    await expect(bar).toHaveCount(0);
  }
  // This layout also needs visual review against Mars: quick mode hides the
  // background and cannot reveal text disappearing over the planet.
  await page.evaluate(() => localStorage.removeItem("openmars.game"));
  await page.goto("/");
  await page.getByRole("link", { name: "New game", exact: true }).click({ trial: true });
  await page.screenshot({
    animations: "disabled",
    path: info.outputPath("landing-background.png"),
  });
});
