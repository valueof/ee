import { assertEquals } from "@std/assert";
import { filterEditable, parsePorcelain } from "../src/git.ts";
import type { Entry } from "../src/git.ts";

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

Deno.test("parsePorcelain: rename pair (R in index)", () => {
  const out = parsePorcelain(enc("R  src/new.ts\0src/old.ts\0"));
  assertEquals(out, [
    {
      indexStatus: "R",
      worktreeStatus: " ",
      path: "src/new.ts",
      renamedFrom: "src/old.ts",
    },
  ]);
});

Deno.test("parsePorcelain: copy pair (C in index)", () => {
  const out = parsePorcelain(enc("C  src/copy.ts\0src/orig.ts\0"));
  assertEquals(out, [
    {
      indexStatus: "C",
      worktreeStatus: " ",
      path: "src/copy.ts",
      renamedFrom: "src/orig.ts",
    },
  ]);
});

Deno.test("parsePorcelain: rename followed by another entry", () => {
  const out = parsePorcelain(
    enc("R  src/new.ts\0src/old.ts\0 M src/other.ts\0"),
  );
  assertEquals(out, [
    {
      indexStatus: "R",
      worktreeStatus: " ",
      path: "src/new.ts",
      renamedFrom: "src/old.ts",
    },
    { indexStatus: " ", worktreeStatus: "M", path: "src/other.ts" },
  ]);
});

Deno.test("filterEditable: keeps untracked (??)", () => {
  const entries = parsePorcelain(enc("?? new.txt\0 M tracked.ts\0"));
  assertEquals(filterEditable(entries), [
    { indexStatus: "?", worktreeStatus: "?", path: "new.txt" },
    { indexStatus: " ", worktreeStatus: "M", path: "tracked.ts" },
  ]);
});

Deno.test("filterEditable: drops ignored (!!)", () => {
  const entries = parsePorcelain(enc("!! ignored.log\0M  staged.ts\0"));
  assertEquals(filterEditable(entries), [
    { indexStatus: "M", worktreeStatus: " ", path: "staged.ts" },
  ]);
});

Deno.test("filterEditable: keeps every editable status code", () => {
  const codes = ["M", "A", "D", "R", "C", "T", "U", "?"];
  for (const code of codes) {
    const entries: Entry[] = [
      { indexStatus: code, worktreeStatus: " ", path: "x" },
    ];
    assertEquals(filterEditable(entries).length, 1, `kept index=${code}`);
    const entries2: Entry[] = [
      { indexStatus: " ", worktreeStatus: code, path: "x" },
    ];
    assertEquals(filterEditable(entries2).length, 1, `kept worktree=${code}`);
  }
});
