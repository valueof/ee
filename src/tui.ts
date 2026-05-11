const ESC = "\x1b";
const ENTER_ALT = `${ESC}[?1049h`;
const EXIT_ALT = `${ESC}[?1049l`;
const HIDE_CURSOR = `${ESC}[?25l`;
const SHOW_CURSOR = `${ESC}[?25h`;
const CLEAR = `${ESC}[2J`;
const HOME = `${ESC}[H`;

export type Key =
  | { kind: "up" }
  | { kind: "down" }
  | { kind: "enter" }
  | { kind: "quit" }
  | { kind: "digit"; value: number }
  | { kind: "char"; value: string }
  | { kind: "unknown" };

const enc = new TextEncoder();
let rawEntered = false;

export function enter(): void {
  if (rawEntered) return;
  Deno.stdin.setRaw(true, { cbreak: false });
  Deno.stdout.writeSync(enc.encode(ENTER_ALT + HIDE_CURSOR));
  rawEntered = true;
}

export function exit(): void {
  if (!rawEntered) return;
  try {
    Deno.stdout.writeSync(enc.encode(SHOW_CURSOR + EXIT_ALT));
  } catch (_) { /* best-effort */ }
  try {
    Deno.stdin.setRaw(false);
  } catch (_) { /* best-effort */ }
  rawEntered = false;
}

export function render(s: string): void {
  Deno.stdout.writeSync(enc.encode(HOME + CLEAR + s));
}

export function terminalWidth(): number {
  try {
    return Deno.consoleSize().columns;
  } catch {
    return 80;
  }
}

export async function readKey(): Promise<Key> {
  const buf = new Uint8Array(8);
  const n = await Deno.stdin.read(buf);
  if (n === null) return { kind: "quit" };
  return parseKey(buf.subarray(0, n));
}

export function parseKey(bytes: Uint8Array): Key {
  if (bytes.length === 0) return { kind: "unknown" };
  const b0 = bytes[0];

  if (b0 === 0x03) return { kind: "quit" };
  if (b0 === 0x0d || b0 === 0x0a) return { kind: "enter" };

  if (b0 === 0x1b) {
    if (bytes.length === 1) return { kind: "quit" };
    if (bytes.length >= 3 && bytes[1] === 0x5b) {
      if (bytes[2] === 0x41) return { kind: "up" };
      if (bytes[2] === 0x42) return { kind: "down" };
    }
    return { kind: "unknown" };
  }

  const c = String.fromCharCode(b0);
  if (c === "q") return { kind: "quit" };
  if (c >= "1" && c <= "9") {
    return { kind: "digit", value: c.charCodeAt(0) - "0".charCodeAt(0) };
  }
  return { kind: "char", value: c };
}
