import { describe, expect, test } from "bun:test";
import { probeServer } from "./probe.ts";

const server = { alias: "saffronbun", url: "https://bun.example" };
const meta = { alias: "saffronbun", name: "Saffronbun", version: "v1.2.3" };

function respond(response: Response | Error): typeof fetch {
  return (async (input: RequestInfo | URL) => {
    expect(String(input)).toBe("https://bun.example/api/v1/meta");
    if (response instanceof Error) {
      throw response;
    }
    return response;
  }) as typeof fetch;
}

describe("probeServer", () => {
  test("returns the server's meta", async () => {
    const result = await probeServer(server, respond(Response.json(meta)));
    expect(result.ok).toBe(true);
    expect(result.ok && result.meta).toEqual(meta);
  });

  test("rejects a server that answers with another alias", async () => {
    const result = await probeServer(server, respond(Response.json({ ...meta, alias: "local" })));
    expect(result).toEqual({ ok: false, reason: 'identifies as "local"' });
  });

  test("reports an error status", async () => {
    const result = await probeServer(server, respond(new Response("", { status: 502 })));
    expect(result).toEqual({ ok: false, reason: "responded 502" });
  });

  test("reports a network failure as unreachable", async () => {
    const result = await probeServer(server, respond(new TypeError("Failed to fetch")));
    expect(result).toEqual({ ok: false, reason: "unreachable" });
  });

  test("gives up after the timeout", async () => {
    const hang = ((_input: RequestInfo | URL, init?: RequestInit) =>
      new Promise((_resolve, reject) => {
        init?.signal?.addEventListener("abort", () => reject(init.signal?.reason));
      })) as typeof fetch;
    const result = await probeServer(server, hang, 20);
    expect(result).toEqual({ ok: false, reason: "unreachable" });
  });
});
