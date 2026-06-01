package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

const version = "0.1.0"

func helpText() string {
	return `ee ` + version + `

Usage: ee
       ... | ee

A TUI that lists tracked git changes (staged + unstaged) and opens the
selected file in your editor.

With piped input, ee reads file paths (one per line) from stdin instead of
running ` + "`git status`" + `. Paths are resolved against the current directory;
entries outside the repo or that don't exist are skipped. The status column
shows git status for the piped paths (blank if clean).

  rg -l TODO | ee
  git diff --name-only main...HEAD | ee

Environment:
  VISUAL, EDITOR   Editor command (VISUAL takes precedence)
  NO_COLOR         Set to disable colored output

Keys:
  ↑/↓ or j/k       Move cursor (wraps)
  1-9              Open file at that position
  Enter            Open file at cursor
  q, Esc, Ctrl-C   Quit`
}

func installSignalHandlers() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		tuiExit()
		os.Exit(130)
	}()
}

func run() int {
	for _, a := range os.Args[1:] {
		if a == "--version" {
			fmt.Println(version)
			return 0
		}
		if a == "--help" || a == "-h" {
			fmt.Println(helpText())
			return 0
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ee:", err)
		return 1
	}
	repoRoot, err := findRepoRoot(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ee: not a git repository")
		return 1
	}

	var (
		loadEntries func() ([]Entry, error)
		header      string
	)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		header = "modified files"
		loadEntries = func() ([]Entry, error) {
			raw, err := loadStatus(repoRoot)
			if err != nil {
				return nil, err
			}
			return filterEditable(raw), nil
		}
	} else {
		lines, err := readStdin(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ee:", err)
			return 1
		}
		resolved := resolveInRepo(lines, repoRoot, cwd)
		if len(resolved) == 0 {
			fmt.Println("Nothing to edit")
			return 0
		}
		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Fprintln(os.Stderr, "ee:", err)
			return 1
		}
		tuiUseInput(tty)
		header = "files"
		loadEntries = func() ([]Entry, error) {
			raw, err := loadStatus(repoRoot)
			if err != nil {
				return nil, err
			}
			return annotate(resolved, statusMapFrom(raw)), nil
		}
	}

	initial, err := loadEntries()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ee:", err)
		return 1
	}
	if len(initial) == 0 {
		fmt.Println("Nothing to edit")
		return 0
	}

	installSignalHandlers()

	if err := tuiEnter(); err != nil {
		fmt.Fprintln(os.Stderr, "ee:", err)
		return 1
	}
	defer tuiExit()

	if err := appRun(repoRoot, initial, loadEntries, header); err != nil {
		tuiExit()
		fmt.Fprintln(os.Stderr, "ee:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
