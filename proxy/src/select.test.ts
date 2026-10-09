import { describe, expect, test } from "bun:test";
import { defaultOrder, selectServer } from "./select.ts";

const rack = { alias: "eu-1", url: "https://rack.example" };
const bun = { alias: "eu-2", url: "https://bun.example" };
const servers = [rack, bun];

function select(search: string, tabAlias: string | null = null, choiceAlias: string | null = null) {
  return selectServer({ params: new URLSearchParams(search), tabAlias, choiceAlias, servers });
}

describe("selectServer", () => {
  test("boots the first listed server on a first visit", () => {
    expect(select("")).toEqual({ kind: "boot", server: rack, source: "default" });
  });

  test("always asks with ?pick, even with a link and remembered servers", () => {
    expect(select("?pick&s=eu-2", "eu-1", "eu-1")).toEqual({ kind: "pick" });
  });

  test("a shared link wins over this tab and the remembered choice", () => {
    expect(select("?s=eu-2", "eu-1", "eu-1")).toEqual({
      kind: "boot",
      server: bun,
      source: "link",
    });
  });

  test("an unknown link alias asks, saying why", () => {
    expect(select("?s=gone", "eu-1")).toEqual({
      kind: "pick",
      notice: 'There is no server called "gone".',
    });
  });

  test("this tab's server wins over the remembered choice", () => {
    expect(select("", "eu-2", "eu-1")).toEqual({ kind: "boot", server: bun, source: "tab" });
  });

  test("uses the remembered choice before the default", () => {
    expect(select("", null, "eu-2")).toEqual({ kind: "boot", server: bun, source: "choice" });
  });

  test("falls back to the default when remembered servers are no longer listed", () => {
    expect(select("", "gone", "also-gone")).toEqual({
      kind: "boot",
      server: rack,
      source: "default",
    });
  });

  test("asks when no servers are configured", () => {
    expect(
      selectServer({
        params: new URLSearchParams(),
        tabAlias: null,
        choiceAlias: null,
        servers: [],
      }),
    ).toEqual({ kind: "pick" });
  });
});

describe("defaultOrder", () => {
  test("tries the default first, then the rest in listed order", () => {
    const third = { alias: "eu-3", url: "https://third.example" };
    expect(defaultOrder([rack, bun, third], bun)).toEqual([bun, rack, third]);
  });
});
