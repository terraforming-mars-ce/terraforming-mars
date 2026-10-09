import { describe, expect, test } from "bun:test";
import { rebaseImports } from "./boot.ts";

describe("rebaseImports", () => {
  test("points root-relative imports at the server", () => {
    const preamble = `import { injectIntoGlobalHook } from "/@react-refresh";
import "/src/setup.ts";
const late = await import('/src/late.ts');`;
    expect(rebaseImports(preamble, "http://localhost:3000/index.html")).toBe(
      `import { injectIntoGlobalHook } from "http://localhost:3000/@react-refresh";
import "http://localhost:3000/src/setup.ts";
const late = await import('http://localhost:3000/src/late.ts');`,
    );
  });

  test("leaves absolute, protocol-relative and bare imports alone", () => {
    const code = `import a from "https://cdn.example/a.js";
import b from "//cdn.example/b.js";
import c from "react";`;
    expect(rebaseImports(code, "http://localhost:3000")).toBe(code);
  });
});
