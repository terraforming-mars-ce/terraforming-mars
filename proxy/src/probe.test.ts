import { describe, expect, test } from "bun:test";
import { probeServer } from "./probe.ts";

const server = { alias: "eu-2", url: "https://bun.example" };
const meta = { alias: "eu-2", name: "EU 2", version: "v7" };

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
  test("is online when the server answers as itself", async () => {
    const result = await probeServer(server, respond(Response.json(meta)));
    expect(result.status).toBe("online");
    expect(result.status === "online" && result.meta).toEqual(meta);
  });

  test("is down on a network error", async () => {
    const result = await probeServer(server, respond(new TypeError("Failed to fetch")));
    expect(result).toEqual({ status: "down", reason: "unreachable" });
  });

  test("is down on a server error", async () => {
    const result = await probeServer(server, respond(new Response("", { status: 502 })));
    expect(result).toEqual({ status: "down", reason: "responded 502" });
  });

  test("is down after the timeout", async () => {
    const hang = ((_input: RequestInfo | URL, init?: RequestInit) =>
      new Promise((_resolve, reject) => {
        init?.signal?.addEventListener("abort", () => reject(init.signal?.reason));
      })) as typeof fetch;
    const result = await probeServer(server, hang, 20);
    expect(result).toEqual({ status: "down", reason: "unreachable" });
  });

  test("is invalid when there is no meta route", async () => {
    const result = await probeServer(server, respond(new Response("", { status: 404 })));
    expect(result).toEqual({ status: "invalid", reason: "responded 404" });
  });

  test("is invalid when the answer is not JSON", async () => {
    const result = await probeServer(server, respond(new Response("<html>")));
    expect(result).toEqual({ status: "invalid", reason: "not a meta response" });
  });

  test("is invalid when the server answers as another alias", async () => {
    const result = await probeServer(server, respond(Response.json({ ...meta, alias: "local" })));
    expect(result).toEqual({ status: "invalid", reason: 'identifies as "local"' });
  });
});
