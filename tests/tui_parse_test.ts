import { assertEquals } from "@std/assert";
import { parseKey } from "../src/tui.ts";

const b = (...nums: number[]) => new Uint8Array(nums);

Deno.test("parseKey: Ctrl-C → quit", () => {
  assertEquals(parseKey(b(0x03)), { kind: "quit" });
});

Deno.test("parseKey: q → quit", () => {
  assertEquals(parseKey(b("q".charCodeAt(0))), { kind: "quit" });
});

Deno.test("parseKey: bare Esc → quit", () => {
  assertEquals(parseKey(b(0x1b)), { kind: "quit" });
});

Deno.test("parseKey: Enter (CR) → enter", () => {
  assertEquals(parseKey(b(0x0d)), { kind: "enter" });
});

Deno.test("parseKey: Enter (LF) → enter", () => {
  assertEquals(parseKey(b(0x0a)), { kind: "enter" });
});

Deno.test("parseKey: ESC [ A → up", () => {
  assertEquals(parseKey(b(0x1b, 0x5b, 0x41)), { kind: "up" });
});

Deno.test("parseKey: ESC [ B → down", () => {
  assertEquals(parseKey(b(0x1b, 0x5b, 0x42)), { kind: "down" });
});

Deno.test("parseKey: '1'..'9' → digit", () => {
  for (let n = 1; n <= 9; n++) {
    assertEquals(parseKey(b(("" + n).charCodeAt(0))), { kind: "digit", value: n });
  }
});

Deno.test("parseKey: 'j' and 'k' → char", () => {
  assertEquals(parseKey(b("j".charCodeAt(0))), { kind: "char", value: "j" });
  assertEquals(parseKey(b("k".charCodeAt(0))), { kind: "char", value: "k" });
});

Deno.test("parseKey: '0' → char (not a 1-9 shortcut)", () => {
  assertEquals(parseKey(b("0".charCodeAt(0))), { kind: "char", value: "0" });
});

Deno.test("parseKey: unknown ESC seq → unknown", () => {
  assertEquals(parseKey(b(0x1b, 0x5b, 0x5a)), { kind: "unknown" });
});

Deno.test("parseKey: empty input → unknown", () => {
  assertEquals(parseKey(b()), { kind: "unknown" });
});
