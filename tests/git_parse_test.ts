import { assertEquals } from "@std/assert";
import { parsePorcelain } from "../src/git.ts";

const enc = (s: string) => new TextEncoder().encode(s);

Deno.test("parsePorcelain: single modified file", () => {
  const out = parsePorcelain(enc(" M src/foo.ts\0"));
  assertEquals(out, [
    { indexStatus: " ", worktreeStatus: "M", path: "src/foo.ts" },
  ]);
});

Deno.test("parsePorcelain: staged add + unstaged modify", () => {
  const out = parsePorcelain(enc("A  src/added.ts\0 M src/changed.ts\0"));
  assertEquals(out, [
    { indexStatus: "A", worktreeStatus: " ", path: "src/added.ts" },
    { indexStatus: " ", worktreeStatus: "M", path: "src/changed.ts" },
  ]);
});

Deno.test("parsePorcelain: both staged and unstaged", () => {
  const out = parsePorcelain(enc("MM src/dual.ts\0"));
  assertEquals(out, [
    { indexStatus: "M", worktreeStatus: "M", path: "src/dual.ts" },
  ]);
});

Deno.test("parsePorcelain: empty input", () => {
  assertEquals(parsePorcelain(enc("")), []);
});

Deno.test("parsePorcelain: path with embedded space", () => {
  const out = parsePorcelain(enc(" M src/a b.ts\0"));
  assertEquals(out, [
    { indexStatus: " ", worktreeStatus: "M", path: "src/a b.ts" },
  ]);
});
