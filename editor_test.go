package main

import (
	"strings"
	"testing"
)

func TestResolveEditor_PrefersVISUAL(t *testing.T) {
	got, err := resolveEditor(map[string]string{"VISUAL": "nvim", "EDITOR": "vi"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "nvim" {
		t.Fatalf("got %q, want nvim", got)
	}
}

func TestResolveEditor_FallsBackToEDITOR(t *testing.T) {
	got, err := resolveEditor(map[string]string{"EDITOR": "hx"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "hx" {
		t.Fatalf("got %q, want hx", got)
	}
}

func TestResolveEditor_ErrorsWhenNeitherSet(t *testing.T) {
	_, err := resolveEditor(map[string]string{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "EDITOR") {
		t.Fatalf("error %q should mention EDITOR", err)
	}
}

func TestResolveEditor_ErrorsWhenBothEmptyOrWhitespace(t *testing.T) {
	_, err := resolveEditor(map[string]string{"VISUAL": "", "EDITOR": "   "})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveEditor_TrimsWhitespaceOnlyAsMissing(t *testing.T) {
	got, err := resolveEditor(map[string]string{"VISUAL": "   ", "EDITOR": "hx"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "hx" {
		t.Fatalf("got %q, want hx", got)
	}
}
