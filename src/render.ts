import type { Entry } from "./git.ts";

const ESC = "\x1b";
const RESET = `${ESC}[0m`;
const FG_RED = `${ESC}[31m`;
const FG_GREEN = `${ESC}[32m`;
const FG_YELLOW = `${ESC}[33m`;

function paint(code: string, s: string, enabled: boolean): string {
  return enabled ? `${code}${s}${RESET}` : s;
}

function statusCell(entry: Entry, colorEnabled: boolean): string {
  const idx = entry.indexStatus;
  const wt = entry.worktreeStatus;
  if (idx !== " " && wt !== " ") {
    return paint(FG_YELLOW, idx + wt, colorEnabled);
  }
  if (idx !== " ") return paint(FG_GREEN, idx, colorEnabled) + wt;
  if (wt !== " ") return idx + paint(FG_RED, wt, colorEnabled);
  return idx + wt;
}

function truncate(s: string, max: number): string {
  if (max <= 0) return "";
  if (s.length <= max) return s;
  if (max === 1) return "…";
  return s.slice(0, max - 1) + "…";
}

export function renderFrame(
  entries: Entry[],
  cursorIndex: number,
  width: number,
  colorEnabled: boolean,
): string {
  const lines: string[] = [];
  lines.push(`modified files (${entries.length})`);
  lines.push("");

  const prefixCols = 9;

  for (let i = 0; i < entries.length; i++) {
    const entry = entries[i];
    const cursor = i === cursorIndex ? "> " : "  ";
    const badge = i < 9 ? `[${i + 1}]` : "[ ]";
    const status = statusCell(entry, colorEnabled);
    const rawPath = entry.renamedFrom
      ? `${entry.renamedFrom} → ${entry.path}`
      : entry.path;
    const pathBudget = Math.max(0, width - prefixCols);
    const path = truncate(rawPath, pathBudget);
    lines.push(`${cursor}${badge} ${status} ${path}`);
  }

  lines.push("");
  lines.push("↑/↓ or j/k  •  1-9 open  •  Enter open  •  q quit");
  return lines.join("\n");
}
