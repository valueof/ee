package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
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

func TestLoadStatus_StagedUnstagedUntrackedSkipsIgnored(t *testing.T) {
	dir := setupRepo(t)

	mustWrite := func(rel, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mustWrite("committed.txt", "v1\n")
	runGit(t, dir, "add", "committed.txt")
	runGit(t, dir, "commit", "-m", "init")

	mustWrite("committed.txt", "v2\n")
	mustWrite("staged-new.txt", "hi\n")
	runGit(t, dir, "add", "staged-new.txt")
	mustWrite("untracked.txt", "u\n")
	mustWrite(".gitignore", "ignored.log\n")
	mustWrite("ignored.log", "should be excluded\n")

	raw, err := loadStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	entries := filterEditable(raw)

	var paths []string
	for _, e := range entries {
		paths = append(paths, e.Path)
	}
	sort.Strings(paths)
	want := []string{".gitignore", "committed.txt", "staged-new.txt", "untracked.txt"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("got %v, want %v", paths, want)
	}

	var untracked *Entry
	for i := range entries {
		if entries[i].Path == "untracked.txt" {
			untracked = &entries[i]
			break
		}
	}
	if untracked == nil {
		t.Fatal("untracked.txt missing")
	}
	if untracked.IndexStatus != '?' || untracked.WorktreeStatus != '?' {
		t.Fatalf("untracked status %c%c, want ??", untracked.IndexStatus, untracked.WorktreeStatus)
	}
}
