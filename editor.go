package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func resolveEditor(env map[string]string) (string, error) {
	if v := strings.TrimSpace(env["VISUAL"]); v != "" {
		return v, nil
	}
	if v := strings.TrimSpace(env["EDITOR"]); v != "" {
		return v, nil
	}
	return "", errors.New("Set $EDITOR or $VISUAL to use ee")
}

func runEditor(editorCmd, filePath string) error {
	parts := strings.Fields(editorCmd)
	if len(parts) == 0 {
		return errors.New("empty editor command")
	}
	bin := parts[0]
	args := append(parts[1:], filePath)

	tuiExit()
	defer func() { _ = tuiEnter() }()

	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
