import type { Entry } from "./git.ts";

export function renderFrame(
  entries: Entry[],
  cursorIndex: number,
  _width: number,
  _colorEnabled: boolean,
): string {
  const lines: string[] = [];
  lines.push(`modified files (${entries.length})`);
  lines.push("");

  for (let i = 0; i < entries.length; i++) {
    const entry = entries[i];
    const cursor = i === cursorIndex ? "> " : "  ";
    const badge = i < 9 ? `[${i + 1}]` : "[ ]";
    const status = `${entry.indexStatus}${entry.worktreeStatus}`;
    const path = entry.renamedFrom
      ? `${entry.renamedFrom} → ${entry.path}`
      : entry.path;
    lines.push(`${cursor}${badge} ${status} ${path}`);
  }

  lines.push("");
  lines.push("↑/↓ or j/k  •  1-9 open  •  Enter open  •  q quit");
  return lines.join("\n");
}
