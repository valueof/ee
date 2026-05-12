# ee — Design

Date: 2026-05-11

## Purpose

A small terminal tool that replaces the workflow:

```
$ git status
$ hx .
$ (search for the file)
```

with a single command that shows the list of tracked changes in a TUI and
opens the selected file in `$EDITOR` on a keypress.

## Goals & Non-Goals

**Goals**

- Single binary, no runtime install on the user's machine.
- Minimal third-party dependencies.
- Fast, predictable keybindings: arrows, j/k, 1-9 numeric shortcuts, Enter, q.
- Loop back to the file list after the editor closes.

**Non-goals (v1)**

- Diff preview pane.
- Multi-select / batch open.
- In-list searching/filtering.
- Untracked file support.
- Watching `git status` for changes while displayed (no SIGWINCH/inotify).
- Cross-machine distribution beyond `deno task install` on the developer's box.

## Stack

- **Language:** TypeScript.
- **Runtime:** Deno.
- **Distribution:** `deno compile` → single static binary, installed to
  `$HOME/.local/bin/ee` via a `deno task install` task.
- **TUI:** rolled by hand using raw mode + ANSI escape sequences. No TUI
  framework dependency.

The rationale for rolling the TUI: the surface area is small (a list, six
keys, one redraw function). Pulling in a framework would add more code
to learn than to write.

## Architecture

Six modules, each with a single responsibility:

### `src/git.ts`

Wraps git invocations and parses output. Pure with respect to the rest of
the app — does not touch the terminal.

- `findRepoRoot(): Promise<string>` — runs `git rev-parse --show-toplevel`.
  Throws if not in a repo.
- `listChanges(repoRoot: string): Promise<Entry[]>` — runs
  `git status --porcelain=v1 -z` and returns parsed entries.
- `parsePorcelain(bytes: Uint8Array): Entry[]` — exported separately for
  unit testing. Handles NUL-delimited records and rename pairs.

`Entry` shape:

```ts
type Entry = {
  indexStatus: string;     // single char from porcelain (e.g. "M", " ", "A")
  worktreeStatus: string;  // single char
  path: string;            // post-rename path, repo-relative
  renamedFrom?: string;    // present only for rename entries
};
```

Tracked-changes filter applied here: keep entries where either status char
is in `{M, A, D, R, C, T, U}`. Drop `??` and `!!`.

### `src/tui.ts`

Owns terminal state. Knows nothing about git or app logic.

- `enter()` — switch to alternate screen, hide cursor, enable raw mode.
- `exit()` — invert of `enter()`. Idempotent and safe to call from signal
  handlers.
- `readKey(): Promise<Key>` — reads one logical keypress, parsing escape
  sequences. Returns a tagged union:
  ```ts
  type Key =
    | { kind: "up" } | { kind: "down" }
    | { kind: "enter" } | { kind: "quit" }
    | { kind: "digit"; value: number }   // 1-9
    | { kind: "char"; value: string }    // j, k, other
    | { kind: "unknown" };
  ```
- `write(s: string)` — write to stdout, no transformation.

Escape-sequence handling:

- Bare `\x1b` followed by nothing within ~10ms → bare Esc, treated as quit.
- `\x1b[A` / `\x1b[B` → up / down arrows.
- `\x03` (Ctrl-C) → quit.
- All other escape sequences are consumed and dropped.

### `src/render.ts`

Pure functions. Takes state and returns a string for `tui.write()`.

- `renderFrame(entries: Entry[], cursorIndex: number, width: number,
  colorEnabled: boolean): string`

Layout (example):

```
modified files (3)

  [1] M  src/git.ts
> [2] MM src/tui.ts
  [3]  D src/old.ts

↑/↓ or j/k  •  1-9 open  •  Enter open  •  q quit
```

Details:

- Header line shows total count.
- Cursor row prefixed with `> `; non-cursor rows prefixed with `  `.
- Number badge: `[N]` for entries 1-9, `[ ]` for entries 10+.
- Status chars colored:
  - index char colored green if non-space (staged change),
  - worktree char colored red if non-space (unstaged change),
  - both space (shouldn't happen post-filter) → no color.
- Rename display: `old → new`.
- Long paths truncated with `…` to fit `width`.
- `colorEnabled=false` → no SGR escapes, but cursor control escapes still
  emitted (color is suppressed independently of the rest of the TUI).

### `src/editor.ts`

Resolves and runs the user's editor.

- `resolveEditor(env: Record<string,string|undefined>): string` — returns
  `env.VISUAL ?? env.EDITOR`; throws a typed error if neither is set.
- `run(editor: string, filePath: string, tui: TuiHandle): Promise<void>` —
  calls `tui.exit()`, spawns the editor with stdio inherited, awaits exit,
  calls `tui.enter()`. Always restores TUI state, even if the child crashes.

The editor's exit code is discarded; ee's exit code reflects ee
itself, not the editor.

### `src/app.ts`

Main loop. Holds the only mutable state in the app.

```
state = { entries, cursorIndex }
loop:
  render(state)
  key = await tui.readKey()
  match key:
    quit       → break
    up / k     → cursorIndex = (cursorIndex - 1) mod len
    down / j   → cursorIndex = (cursorIndex + 1) mod len
    digit N    → if N <= len: openAt(N - 1)
    enter      → openAt(cursorIndex)
openAt(i):
  editor.run(resolveEditor(Deno.env), abs(entries[i].path))
  entries = await git.listChanges(repoRoot)
  if entries.empty: print "Nothing to edit"; exit 0
  cursorIndex = clamp(cursorIndex, 0, len - 1)
```

### `src/main.ts`

Entry point. Responsibilities:

1. Parse `--help` / `--version` (anything else → ignore for v1).
2. Install signal handlers (SIGINT, SIGTERM, `globalThis.addEventListener
   ("unhandledrejection", ...)`, `addEventListener("error", ...)`) that
   always run `tui.exit()` before re-throwing or exiting.
3. Resolve repo root. If absent, print to stderr and exit 1.
4. Run initial `git.listChanges()`. If empty, print `Nothing to edit` and
   exit 0 without entering the TUI.
5. Hand control to `app.run()`.

## Behavior Spec

### Startup

- If `git rev-parse --show-toplevel` fails: print error to stderr, exit 1.
- If no tracked changes: print `Nothing to edit` to stdout, exit 0. Do not
  enter the TUI.

### Display

- One file per line. Order: as returned by `git status` (do not re-sort).
- Numeric shortcut visible only on first 9 entries.
- Renames: `old → new`. Selecting opens `new`.

### Keys

| Key | Action |
|---|---|
| `↑`, `k` | Cursor up, wraps at top |
| `↓`, `j` | Cursor down, wraps at bottom |
| `1`-`9` | Open file at that 1-based position |
| `Enter` | Open file at cursor |
| `q`, `Esc`, `Ctrl-C` | Quit, exit 0 |

### Opening a file

1. Exit raw mode + alt screen.
2. Spawn `$VISUAL` or `$EDITOR` with the file's absolute path, stdio
   inherited.
3. Await exit. Discard exit code.
4. Re-enter raw mode + alt screen.
5. Reload `git.listChanges()`. If now empty, exit 0 with `Nothing to edit`.
6. Otherwise clamp cursor index and redraw.

### Edge cases

- **Deleted file selected:** pass path to editor anyway. Editor will open
  an empty buffer; user can recreate content if they want.
- **No `$EDITOR` and no `$VISUAL`:** error out with a clear message before
  attempting to spawn.
- **Terminal resize:** there is no continuous render loop — the next
  redraw happens when the user presses a key. Resize during idle will
  show stale layout until then. No SIGWINCH subscription in v1.
- **Terminal too narrow:** truncate paths with `…`. Do not wrap.
- **Crash or signal:** signal handlers and a top-level `try/finally` in
  `main.ts` guarantee `tui.exit()` runs.

### Colors

- 16-color ANSI: green = staged, red = unstaged, yellow if both index and
  worktree status chars are non-space.
- `NO_COLOR` env var: when set (to any value), suppress all SGR escapes.
  Cursor-control escapes (cursor move, clear screen, alt-screen toggle)
  are still emitted.

## Build & Install

### Project layout

```
ee/
├─ deno.json
├─ src/
│  ├─ main.ts
│  ├─ app.ts
│  ├─ git.ts
│  ├─ tui.ts
│  ├─ render.ts
│  └─ editor.ts
├─ tests/
│  ├─ git_test.ts
│  ├─ render_test.ts
│  └─ editor_test.ts
├─ README.md
└─ .gitignore
```

### `deno.json` tasks

- `dev` — `deno run --allow-run --allow-env --allow-read src/main.ts`
- `test` — `deno test`
- `build` — `deno compile --allow-run --allow-env --allow-read
  --output dist/ee src/main.ts`
- `install` — runs `build`, then copies `dist/ee` to
  `$HOME/.local/bin/ee` and `chmod +x`.

### Permissions

- `--allow-run` (unrestricted) — needed because `$EDITOR` is dynamic.
  Documented in README.
- `--allow-env` — reads `EDITOR`, `VISUAL`, `NO_COLOR`, `HOME`.
- `--allow-read` — Deno requires this to look up binaries on `PATH`
  before spawning.

### Distribution

- v1: clone repo, `deno task install`. Single user, single machine.
- Not published to JSR or npm in v1.
- Prebuilt cross-platform binaries via GitHub Releases is a possible
  follow-up; out of scope for v1.

### Binary size

Expect 80-100 MB. Accepted cost of "single binary, no runtime install."

### Versioning

A `VERSION` constant in `src/main.ts`. `--version` prints it. Bump
manually. No release machinery in v1.

## Testing Strategy

### Unit-tested

- **`git.ts` parser** — feed canned porcelain byte strings (including
  rename pairs with their NUL separators and paths containing spaces,
  newlines, and shell metacharacters — `-z` emits these raw, no quoting)
  and assert parsed `Entry[]`. Cover every status
  code in `{M, A, D, R, C, T, U, ??, !!}`. Confirm the tracked-changes
  filter drops `??` and `!!`.
- **`render.ts`** — snapshot-test `renderFrame()` for: empty list,
  single entry, 1-9 entries, 10+ entries (placeholder `[ ]`), each
  status-code color combination, cursor at top/middle/bottom,
  narrow-terminal truncation, `NO_COLOR` mode.
- **`editor.ts` resolver** — `resolveEditor()` is a pure function over
  env; cover present/absent/both cases.

### Integration-tested

- **`git.ts` end-to-end** — in each test: `Deno.makeTempDir`, `git init`,
  create/stage/modify/rename/delete files, run the real `git.listChanges`,
  assert. Catches porcelain drift across git versions.

### Not tested (deliberately)

- **`tui.ts`** raw-mode I/O — driving a real terminal in tests is
  high-effort, low-value. Manual smoke tests during development.
- **`app.ts`** main loop — thin once dependencies are tested. Revisit if
  bugs cluster there.

### Manual smoke checklist (in README)

- Empty state → `Nothing to edit`.
- Mix of staged + unstaged + renamed + deleted files renders correctly.
- Open a file, save, exit editor → list refreshes.
- Open a file, commit it from a separate shell, exit editor → list
  reflects updated state.
- Ctrl-C during list view → terminal restored cleanly.
- Narrow terminal (50 cols) → paths truncate, no wrapping.
- `NO_COLOR=1` → no color codes emitted; layout unchanged.

## Open Questions

None at design time. Anything that surfaces during implementation will
either fit one of the listed edge cases or be deferred to v2.
