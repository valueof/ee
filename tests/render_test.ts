import { assertEquals, assertStringIncludes } from "@std/assert";
import { renderFrame } from "../src/render.ts";
import type { Entry } from "../src/git.ts";

const e = (i: string, w: string, p: string): Entry => ({
  indexStatus: i,
  worktreeStatus: w,
  path: p,
});

Deno.test("renderFrame: header shows count", () => {
  const out = renderFrame([e(" ", "M", "a.ts"), e("M", " ", "b.ts")], 0, 80, false);
  assertStringIncludes(out, "modified files (2)");
});

Deno.test("renderFrame: cursor row prefixed with '> '", () => {
  const out = renderFrame([e(" ", "M", "a.ts"), e(" ", "M", "b.ts")], 1, 80, false);
  const lines = out.split("\n");
  const entryLines = lines.filter((l) => /^[ >] {1,2}\[/.test(l));
  assertEquals(entryLines.length, 2);
  assertEquals(entryLines[0].startsWith("  "), true);
  assertEquals(entryLines[1].startsWith("> "), true);
});

Deno.test("renderFrame: numeric badge 1-9 then '[ ]' for 10+", () => {
  const entries = Array.from({ length: 11 }, (_, k) => e(" ", "M", `f${k}.ts`));
  const out = renderFrame(entries, 0, 80, false);
  assertStringIncludes(out, "[1]");
  assertStringIncludes(out, "[9]");
  assertStringIncludes(out, "[ ] ");
});

Deno.test("renderFrame: empty list still shows header and footer", () => {
  const out = renderFrame([], 0, 80, false);
  assertStringIncludes(out, "modified files (0)");
  assertStringIncludes(out, "1-9 open");
});

Deno.test("renderFrame: footer present", () => {
  const out = renderFrame([e(" ", "M", "a.ts")], 0, 80, false);
  assertStringIncludes(out, "↑/↓ or j/k");
  assertStringIncludes(out, "q quit");
});
