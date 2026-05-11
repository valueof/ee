import { assertEquals, assertThrows } from "@std/assert";
import { resolveEditor } from "../src/editor.ts";

Deno.test("resolveEditor: prefers VISUAL", () => {
  const out = resolveEditor({ VISUAL: "nvim", EDITOR: "vi" });
  assertEquals(out, "nvim");
});

Deno.test("resolveEditor: falls back to EDITOR", () => {
  const out = resolveEditor({ EDITOR: "hx" });
  assertEquals(out, "hx");
});

Deno.test("resolveEditor: throws when neither is set", () => {
  assertThrows(() => resolveEditor({}), Error, "EDITOR");
});

Deno.test("resolveEditor: throws when both are empty strings", () => {
  assertThrows(() => resolveEditor({ VISUAL: "", EDITOR: "   " }), Error);
});

Deno.test("resolveEditor: trims whitespace-only as missing", () => {
  const out = resolveEditor({ VISUAL: "   ", EDITOR: "hx" });
  assertEquals(out, "hx");
});
