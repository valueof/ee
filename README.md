# ee

A tiny TUI that lists modified and untracked files in a git repo and opens
the selected one in your editor — replacing the `git status` +
open-in-editor + find-file dance with one command. Ignored files (per
`.gitignore`) are excluded.

## Install

Requires [Go](https://go.dev) 1.22 or newer.

```
git clone <this repo>
cd ee
go build -o "$HOME/.local/bin/ee"
```

Make sure `~/.local/bin` is on your `PATH`.

## Usage

In a git repo:

```
ee
```

Keys:

| Key | Action |
|---|---|
| ↑ / k | Cursor up (wraps) |
| ↓ / j | Cursor down (wraps) |
| 1-9 | Open file at that position |
| Enter | Open file at cursor |
| q / Esc / Ctrl-C | Quit |

Environment:

- `VISUAL` / `EDITOR` — editor command (VISUAL wins). Supports args, e.g. `EDITOR="nvim -p"`.
- `NO_COLOR` — set to any value to disable colored output.

## Manual smoke checklist

- Empty state shows `Nothing to edit` and exits cleanly.
- Mixed staged + unstaged + renamed + deleted + untracked files render correctly.
- Untracked files show `??` in red and open as their existing content.
- Files matched by `.gitignore` do NOT appear in the list.
- Editor opens, edits saved, list refreshes after editor closes.
- Committing all changes from inside the editor causes the next list to be empty.
- Ctrl-C cleanly restores the terminal.
- 50-column terminal: paths truncate with `…`, no wrapping.
- `NO_COLOR=1` suppresses color, layout unchanged.

## Development

```
go test ./...                # run all tests
go run .                     # run from source
go build -o ee               # produce ./ee
```
