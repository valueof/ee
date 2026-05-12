# ee

A tiny TUI that lists tracked git changes and opens the selected file in
your editor — replacing the `git status` + open-in-editor + find-file dance
with one command.

## Install

Requires [Deno](https://deno.com).

```
git clone <this repo>
cd ee
deno task install
```

This compiles a static binary to `$HOME/.local/bin/ee`. Make sure
`~/.local/bin` is on your `PATH`.

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

## Permissions note

The compiled binary requires `--allow-run` unrestricted (rather than
`--allow-run=git,$EDITOR`) because `$EDITOR` is resolved at runtime.

## Manual smoke checklist

- Empty state shows `Nothing to edit` and exits cleanly.
- Mixed staged + unstaged + renamed + deleted files render correctly.
- Editor opens, edits saved, list refreshes after editor closes.
- Committing all changes from inside the editor causes the next list to be empty.
- Ctrl-C cleanly restores the terminal.
- 50-column terminal: paths truncate with `…`, no wrapping.
- `NO_COLOR=1` suppresses color, layout unchanged.

## Development

```
deno task test       # run all tests
deno task dev        # run from source
deno task build      # produce dist/ee
```
