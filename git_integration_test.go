package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	return dir
}

func TestFindRepoRoot_FromSubdirectory(t *testing.T) {
	dir := setupRepo(t)
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := findRepoRoot(sub)
	if err != nil {
		t.Fatal(err)
	}
	gotReal, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", got, err)
	}
	wantReal, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", dir, err)
	}
	if gotReal != wantReal {
		t.Fatalf("got %q, want %q", gotReal, wantReal)
	}
}

func TestFindRepoRoot_RejectsNonRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := findRepoRoot(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
