package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func realPath(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestResolveInRepo_KeepsExistingFile(t *testing.T) {
	root := realPath(t, t.TempDir())
	writeFile(t, filepath.Join(root, "a.ts"), "x")
	got := resolveInRepo([]string{"a.ts"}, root, root)
	if !reflect.DeepEqual(got, []string{"a.ts"}) {
		t.Fatalf("got %#v", got)
	}
}

func TestResolveInRepo_DropsMissingFile(t *testing.T) {
	root := realPath(t, t.TempDir())
	got := resolveInRepo([]string{"missing.ts"}, root, root)
	if len(got) != 0 {
		t.Fatalf("got %#v, want empty", got)
	}
}

func TestResolveInRepo_DropsDirectory(t *testing.T) {
	root := realPath(t, t.TempDir())
	if err := os.Mkdir(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := resolveInRepo([]string{"sub"}, root, root)
	if len(got) != 0 {
		t.Fatalf("got %#v, want empty", got)
	}
}

func TestResolveInRepo_FollowsSymlinkInside(t *testing.T) {
	root := realPath(t, t.TempDir())
	writeFile(t, filepath.Join(root, "real.ts"), "x")
	if err := os.Symlink("real.ts", filepath.Join(root, "link.ts")); err != nil {
		t.Fatal(err)
	}
	got := resolveInRepo([]string{"link.ts"}, root, root)
	if !reflect.DeepEqual(got, []string{"real.ts"}) {
		t.Fatalf("got %#v, want [real.ts]", got)
	}
}

func TestResolveInRepo_DropsSymlinkPointingOutside(t *testing.T) {
	outside := realPath(t, t.TempDir())
	writeFile(t, filepath.Join(outside, "secret.ts"), "x")
	root := realPath(t, t.TempDir())
	if err := os.Symlink(filepath.Join(outside, "secret.ts"), filepath.Join(root, "leak.ts")); err != nil {
		t.Fatal(err)
	}
	got := resolveInRepo([]string{"leak.ts"}, root, root)
	if len(got) != 0 {
		t.Fatalf("got %#v, want empty", got)
	}
}

func TestResolveInRepo_AbsolutePathInside(t *testing.T) {
	root := realPath(t, t.TempDir())
	writeFile(t, filepath.Join(root, "a.ts"), "x")
	got := resolveInRepo([]string{filepath.Join(root, "a.ts")}, root, "/")
	if !reflect.DeepEqual(got, []string{"a.ts"}) {
		t.Fatalf("got %#v", got)
	}
}

func TestResolveInRepo_AbsolutePathOutside(t *testing.T) {
	root := realPath(t, t.TempDir())
	outside := realPath(t, t.TempDir())
	writeFile(t, filepath.Join(outside, "x.ts"), "x")
	got := resolveInRepo([]string{filepath.Join(outside, "x.ts")}, root, "/")
	if len(got) != 0 {
		t.Fatalf("got %#v, want empty", got)
	}
}

func TestResolveInRepo_Dedup(t *testing.T) {
	root := realPath(t, t.TempDir())
	writeFile(t, filepath.Join(root, "a.ts"), "x")
	got := resolveInRepo([]string{"a.ts", "a.ts", "./a.ts"}, root, root)
	if !reflect.DeepEqual(got, []string{"a.ts"}) {
		t.Fatalf("got %#v, want [a.ts]", got)
	}
}

func TestResolveInRepo_RelativeFromSubdir(t *testing.T) {
	root := realPath(t, t.TempDir())
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "a.ts"), "x")
	got := resolveInRepo([]string{"a.ts"}, root, sub)
	if !reflect.DeepEqual(got, []string{"sub/a.ts"}) {
		t.Fatalf("got %#v, want [sub/a.ts]", got)
	}
}
