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

# Environment:

- `VISUAL` / `EDITOR` — editor command (VISUAL wins). Supports args, e.g. `EDITOR="nvim -p"`.
- `NO_COLOR` — set to any value to disable colored output.
