package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const version = "0.1.0"

func helpText() string {
	return `ee ` + version + `

Usage: ee

A TUI that lists tracked git changes (staged + unstaged) and opens the
selected file in your editor.

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

	initial, err := listChanges(repoRoot)
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

	if err := appRun(repoRoot, initial); err != nil {
		tuiExit()
		fmt.Fprintln(os.Stderr, "ee:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
