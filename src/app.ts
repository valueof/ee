import * as tui from "./tui.ts";
import * as git from "./git.ts";
import { renderFrame } from "./render.ts";
import { resolveEditor, runEditor } from "./editor.ts";

type State = {
  repoRoot: string;
  entries: git.Entry[];
  cursor: number;
};

function colorEnabled(env: Record<string, string | undefined>): boolean {
  return env.NO_COLOR === undefined;
}

function clampCursor(cursor: number, len: number): number {
  if (len === 0) return 0;
  if (cursor < 0) return 0;
  if (cursor >= len) return len - 1;
  return cursor;
}

function pathToOpen(entry: git.Entry, repoRoot: string): string {
  return `${repoRoot}/${entry.path}`;
}

async function openSelected(state: State): Promise<State> {
  const env = Deno.env.toObject();
  const editor = resolveEditor(env);
  const entry = state.entries[state.cursor];
  await runEditor(editor, pathToOpen(entry, state.repoRoot), tui);

  const next = await git.listChanges(state.repoRoot);
  return {
    ...state,
    entries: next,
    cursor: clampCursor(state.cursor, next.length),
  };
}

export async function run(
  repoRoot: string,
  initial: git.Entry[],
): Promise<void> {
  let state: State = { repoRoot, entries: initial, cursor: 0 };
  const env = Deno.env.toObject();

  while (true) {
    if (state.entries.length === 0) {
      tui.exit();
      console.log("Nothing to edit");
      return;
    }

    tui.render(renderFrame(
      state.entries,
      state.cursor,
      tui.terminalWidth(),
      colorEnabled(env),
    ));

    const key = await tui.readKey();
    switch (key.kind) {
      case "quit":
        return;
      case "up":
        state = {
          ...state,
          cursor: (state.cursor - 1 + state.entries.length) % state.entries.length,
        };
        break;
      case "down":
        state = { ...state, cursor: (state.cursor + 1) % state.entries.length };
        break;
      case "char":
        if (key.value === "j") {
          state = { ...state, cursor: (state.cursor + 1) % state.entries.length };
        } else if (key.value === "k") {
          state = {
            ...state,
            cursor: (state.cursor - 1 + state.entries.length) % state.entries.length,
          };
        }
        break;
      case "digit":
        if (key.value >= 1 && key.value <= state.entries.length) {
          state = { ...state, cursor: key.value - 1 };
          state = await openSelected(state);
        }
        break;
      case "enter":
        state = await openSelected(state);
        break;
      case "unknown":
        break;
    }
  }
}
