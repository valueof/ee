# ee

A tiny TUI that lists modified and untracked files in a git repo and opens
the selected one in your editor — replacing the `git status` +
open-in-editor + find-file with one command. Ignored files (per `.gitignore`) are excluded.

## Install

Requires [Go](https://go.dev) 1.22 or newer.

```
git clone <this repo>
cd ee
make install
```

Make sure `~/.local/bin` is on your `PATH`.

## Stdin mode

If you pipe input, `ee` reads file paths (one per line) from stdin instead
of running `git status`. Paths outside the repo or that don't exist are
skipped. The status column reflects current git status for the piped paths
(blank if the file is tracked and clean).

```
rg -l TODO | ee
git diff --name-only main...HEAD | ee
find . -name '*.ts' | ee
```

# Environment:

- `VISUAL` / `EDITOR` — editor command (VISUAL wins). Supports args, e.g. `EDITOR="nvim -p"`.
- `NO_COLOR` — set to any value to disable colored output.
