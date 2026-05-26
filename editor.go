package main

import (
	"errors"
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
