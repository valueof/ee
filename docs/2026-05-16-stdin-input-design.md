# ee — Stdin Input Design

Date: 2026-05-16

## Purpose

Let `ee` read its file list from piped input instead of `git status`, so
common pipelines work:

```
rg -l "TODO" | ee
git diff --name-only main...HEAD | ee
find . -name '*.ts' | ee
```

The TUI and editor flow are unchanged. Only the source of the list and
the status-annotation step change.

## Goals & Non-Goals

**Goals**

- Accept newline-delimited file paths on stdin.
- Auto-detect piped input via `Deno.stdin.isTerminal()`. No flag.
- Constrain piped paths to the current git repo. External paths are
  dropped, the repo requirement stays.
- Annotate each piped path with its current git status so the existing
  display works unchanged for both modes.
- Refresh statuses after each edit. The path list itself is frozen.
- Refactor the data flow into two stages (paths → annotate) so both
  modes share the same pipeline.

**Non-goals**

- NUL-delimited input (`xargs -0` style). Filenames containing newlines
  are not supported.
- Grep-with-line-numbers parsing (`file:lineno:content`). Path-only.
- Honoring `--stdin` / `-` flags. Detection is implicit.
- Windows support. `/dev/tty` is Unix-only and the project already
  targets Unix.
- Capping stdin size, paginating, or filtering for very large lists.

## Architecture

Two-stage pipeline. Both modes flow through the same shape:

```
                         ┌─────────────────┐
git mode  ──────────►    │   path source   │
stdin mode ─────────►    │ (returns paths) │ ──┐
                         └─────────────────┘   │
                                               ▼
                              ┌──────────────────────────┐
                              │   annotate with status   │ ──► Entry[]
                              │  (git status, look up)   │
                              └──────────────────────────┘
```

Mode selection happens once in `main.ts` based on `Deno.stdin.isTerminal()`.
The chosen path source and the annotation step are composed into a
`loadEntries: () => Promise<Entry[]>` closure passed to `app.run()`.
`app.ts` calls it at startup and after each edit; it does not know which
mode is active.

### Module changes

`src/git.ts`

- `findRepoRoot()` — unchanged.
- `parsePorcelain(bytes)` — unchanged, still exported for tests.
- `listChanges()` — **removed**.
- `statusMap(repoRoot): Promise<Map<string, Entry>>` — **new**. Runs
  `git status --porcelain=v1 -z`, parses, returns a map keyed by
  `entry.path` (post-rename path). Iteration order matches git's
  output order.
- `filterEditable(entries)` — unchanged. Used only by the git path
  source.

`src/paths.ts` — **new module.**

- `readStdin(): Promise<string[]>` — reads `Deno.stdin.readable` to EOF,
  decodes as UTF-8, splits on `\n`, strips trailing `\r`, trims each
  line, drops empty lines.
- `resolveInRepo(lines, repoRoot, cwd): string[]` — for each line:
  resolve against `cwd` if not absolute, `Deno.realPathSync()` to
  canonicalize, `Deno.statSync()` to verify it is a regular file,
  check the canonical path is inside `repoRoot`. Return repo-relative
  paths, de-duplicated, first occurrence wins.
- `annotate(paths, statusMap): Entry[]` — for each path, return
  `statusMap.get(p)` if present, else
  `{ indexStatus: " ", worktreeStatus: " ", path: p }`.

`src/tui.ts`

- Add module-level `let input = Deno.stdin`.
- `useInput(handle: Deno.FsFile): void` — **new**. Swaps the stored
  input handle.
- `childStdin(): Deno.FsFile | "inherit"` — **new**. Returns
  `"inherit"` when `input === Deno.stdin`; otherwise opens a fresh
  `/dev/tty` handle for the spawned editor (Deno's `Command` takes
  ownership of the FsFile, so the TUI's own handle cannot be shared).
- `enter()` calls `input.setRaw(true, { cbreak: false })`.
- `readKey()` calls `input.read(buf)`.

`src/editor.ts`

- `TuiHandle` gains `childStdin(): Deno.FsFile | "inherit"`.
- `runEditor()` passes `stdin: tui.childStdin()` to `Deno.Command`
  instead of the hardcoded `"inherit"`.

`src/app.ts`

- `run(repoRoot, initial, loadEntries)` — gains a third parameter,
  `loadEntries: () => Promise<Entry[]>`.
- `openSelected` calls `await loadEntries()` instead of
  `git.listChanges(state.repoRoot)`.
- State and key dispatch unchanged.

`src/main.ts` — orchestrates mode selection:

```
parse --help / --version
resolve repo root (or exit 1)

if Deno.stdin.isTerminal():
    loadEntries = async () =>
        git.filterEditable([...(await git.statusMap(root)).values()])
else:
    lines = await paths.readStdin()
    resolved = paths.resolveInRepo(lines, root, Deno.cwd())
    tty = await Deno.open("/dev/tty", { read: true })
    tui.useInput(tty)
    loadEntries = async () =>
        paths.annotate(resolved, await git.statusMap(root))

initial = await loadEntries()
if initial.empty: print "Nothing to edit"; exit 0
installSignalHandlers()
try:
    tui.enter()
    await app.run(root, initial, loadEntries)
finally:
    tui.exit()
```

The order is load-critical: `readStdin()` must complete before
`tui.enter()` switches the terminal into raw mode.

`src/render.ts`

- `renderFrame(entries, cursorIndex, width, colorEnabled, header)` —
  gains a `header` string parameter. `app.ts` passes
  `"modified files"` in git mode and `"files"` in stdin mode. The
  count is appended by `renderFrame` as today.
- `statusCell` is unchanged. Tracked-clean entries (both status chars
  space) already fall through to plain `"  "` with no color.

## Behavior Spec

### Mode detection

- `Deno.stdin.isTerminal() === true` → git mode (current behavior).
- `Deno.stdin.isTerminal() === false` → stdin mode.

### Stdin parsing

- Input is read fully to EOF before the TUI starts.
- Split on `\n`. Trailing `\r` on each line is stripped. Each line is
  trimmed; empty lines are dropped.
- No interpretation of `#`, quoting, or escape sequences. Each non-empty
  trimmed line is a path.
- Filenames containing literal newlines are not supported.

### Path resolution

For each line, in order:

1. If absolute → use as-is. Else → resolve against `Deno.cwd()`.
2. `Deno.realPathSync()` canonicalizes (symlinks, `..`, trailing
   slashes). On `ENOENT` or any error, drop the entry.
3. `Deno.statSync()`: drop if not a regular file (directories, sockets,
   block/char devices).
4. If the canonical path is not equal to `repoRoot` or under
   `${repoRoot}/`, drop.
5. Compute `relative(repoRoot, canonicalPath)` for the repo-relative
   key used by display and status lookup.

De-duplicate by repo-relative path. Preserve first occurrence.

### Status annotation

- `git.statusMap(repoRoot)` runs once per refresh.
- For each resolved path: `map.get(path) ?? cleanEntry(path)`.
- `cleanEntry(p) = { indexStatus: " ", worktreeStatus: " ", path: p }`.

### Loop behavior

- Startup: build `loadEntries`, call it once, exit 0 with
  `Nothing to edit` if the result is empty.
- After each editor exit: call `loadEntries()` again. In stdin mode the
  resolved path list is frozen in the closure; only `statusMap` re-runs.
- Cursor clamp is the same as today.

### Display

- git mode header: `modified files (N)` — unchanged.
- stdin mode header: `files (N)`.
- Tracked-clean entries: status cell `"  "` with no color. Already
  correct in `statusCell()`.
- Rename display: not produced for stdin entries (the user piped a
  single path; we have no `renamedFrom` for it). Renames continue to
  appear in git mode unchanged.
- Truncation, numeric badges, cursor prefix: unchanged.

### CLI

- No new flags. `--help` text extended with a stdin-mode paragraph and
  two example pipelines:

  ```
  With piped input, ee reads file paths (one per line) from stdin
  instead of running `git status`. Paths are resolved against the
  current directory; entries outside the repo or that don't exist are
  skipped. The status column shows git status for the piped paths
  (blank if clean).

    rg -l TODO | ee
    git diff --name-only main...HEAD | ee
  ```

- `--version` unchanged.

### Edge cases

- **Empty or all-filtered stdin** → `Nothing to edit`, exit 0.
- **`/dev/tty` open fails** (detached process, no controlling
  terminal) → top-level `try/catch` in `main.ts` prints
  `ee: <error>`, exits 1.
- **Path that looks like a flag** in stdin (`--help`, `-h`) → treated
  as a path. Flag parsing only consults `Deno.args`.
- **Very large stdin** → buffered fully in memory. No size cap. Same
  semantics as `xargs`.
- **Gitignored file in stdin** → kept. It is inside `repoRoot`, just
  has no `git status` entry, so it renders with blank status.
- **Symlink inside repo pointing outside** → canonical path is
  outside → dropped.
- **SIGINT during `readStdin()`** → handled by the existing signal
  handler. `tui.exit()` is a no-op pre-`enter()`, process exits 130.

## Permissions

The compiled binary already declares unrestricted
`--allow-run --allow-env --allow-read`. `/dev/tty` reads are covered by
the existing `--allow-read`. No permission additions in `deno.json`.

## Testing Strategy

### Unit (new)

- **`paths.readStdin` parsing logic** — refactor so a buffer can be
  injected. Cover: CRLF, blank lines, leading/trailing whitespace,
  trailing newline present and absent.
- **`paths.resolveInRepo`** — in a `Deno.makeTempDir` git repo: existing
  file, missing file, directory, symlink to inside, symlink to outside,
  absolute path inside, absolute path outside, duplicates, paths from a
  subdirectory cwd.
- **`paths.annotate`** — paths in map, paths not in map, mixed.

### Unit (modified)

- **`git.statusMap`** — replaces the existing `listChanges` test.
  Same fixtures, asserts map contents and key set.
- **`render.ts`** — add a case for the stdin-mode header label.

### Not tested

- **`tui.useInput` / `tui.childStdin`** — manual smoke only, matching
  the existing "no TUI tests" stance.
- **End-to-end stdin → spawned binary → simulated keys** — high
  effort, low value relative to the unit coverage above.

### Manual smoke (added to README)

- `echo src/git.ts | ee` → one entry, opens, returns to list.
- `rg -l TODO | ee` → status column populated; modified files colored,
  clean files blank.
- `find /tmp -name '*.txt' | ee` → `Nothing to edit`.
- `cd src && ls *.ts | ee` → paths resolve correctly from subdir.
- Editor opens normally in stdin mode (exercises `/dev/tty` for both
  TUI and child stdin).
- `printf "src/git.ts\nsrc/git.ts\n" | ee` → one entry, de-duped.

## Open Questions

None at design time.
