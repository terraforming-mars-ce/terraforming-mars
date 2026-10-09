import { describe, expect, test } from "bun:test";
import { selectServer } from "./select.ts";

const rack = { alias: "rackaracka", url: "https://rack.example" };
const bun = { alias: "saffronbun", url: "https://bun.example" };
const servers = [rack, bun];

function select(search: string, tabAlias: string | null = null, lastAlias: string | null = null) {
  return selectServer({ params: new URLSearchParams(search), tabAlias, lastAlias, servers });
}

describe("selectServer", () => {
  test("asks on a first visit", () => {
    expect(select("")).toEqual({ kind: "pick" });
  });

  test("always asks with ?pick, even with a link and a remembered server", () => {
    expect(select("?pick&s=saffronbun", "rackaracka", "rackaracka")).toEqual({ kind: "pick" });
  });

  test("a shared link wins over what this tab and browser used", () => {
    expect(select("?s=saffronbun", "rackaracka", "rackaracka")).toEqual({
      kind: "boot",
      server: bun,
    });
  });

  test("an unknown link alias asks, saying why", () => {
    expect(select("?s=gone", "rackaracka")).toEqual({
      kind: "pick",
      notice: 'There is no server called "gone".',
    });
  });

  test("this tab's server wins over the browser's last one", () => {
    expect(select("", "saffronbun", "rackaracka")).toEqual({ kind: "boot", server: bun });
  });

  test("falls back to the browser's last server", () => {
    expect(select("", null, "rackaracka")).toEqual({ kind: "boot", server: rack });
  });

  test("ignores remembered servers that are no longer listed", () => {
    expect(select("", "gone", "also-gone")).toEqual({ kind: "pick" });
  });

  test("boots the only server without asking", () => {
    expect(
      selectServer({
        params: new URLSearchParams(),
        tabAlias: null,
        lastAlias: null,
        servers: [rack],
      }),
    ).toEqual({ kind: "boot", server: rack });
  });
});
