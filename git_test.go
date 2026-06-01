package main

import (
	"reflect"
	"testing"
)

func TestParsePorcelain_SingleModified(t *testing.T) {
	got := parsePorcelain([]byte(" M src/foo.ts\x00"))
	want := []Entry{{IndexStatus: ' ', WorktreeStatus: 'M', Path: "src/foo.ts"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParsePorcelain_StagedAddPlusUnstagedModify(t *testing.T) {
	got := parsePorcelain([]byte("A  src/added.ts\x00 M src/changed.ts\x00"))
	want := []Entry{
		{IndexStatus: 'A', WorktreeStatus: ' ', Path: "src/added.ts"},
		{IndexStatus: ' ', WorktreeStatus: 'M', Path: "src/changed.ts"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParsePorcelain_BothStagedAndUnstaged(t *testing.T) {
	got := parsePorcelain([]byte("MM src/dual.ts\x00"))
	want := []Entry{{IndexStatus: 'M', WorktreeStatus: 'M', Path: "src/dual.ts"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParsePorcelain_EmptyInput(t *testing.T) {
	got := parsePorcelain([]byte(""))
	if len(got) != 0 {
		t.Fatalf("expected empty, got %#v", got)
	}
}

func TestParsePorcelain_PathWithEmbeddedSpace(t *testing.T) {
	got := parsePorcelain([]byte(" M src/a b.ts\x00"))
	want := []Entry{{IndexStatus: ' ', WorktreeStatus: 'M', Path: "src/a b.ts"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParsePorcelain_RenameInIndex(t *testing.T) {
	got := parsePorcelain([]byte("R  src/new.ts\x00src/old.ts\x00"))
	want := []Entry{{
		IndexStatus: 'R', WorktreeStatus: ' ',
		Path: "src/new.ts", RenamedFrom: "src/old.ts",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParsePorcelain_CopyInIndex(t *testing.T) {
	got := parsePorcelain([]byte("C  src/copy.ts\x00src/orig.ts\x00"))
	want := []Entry{{
		IndexStatus: 'C', WorktreeStatus: ' ',
		Path: "src/copy.ts", RenamedFrom: "src/orig.ts",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParsePorcelain_RenameFollowedByAnotherEntry(t *testing.T) {
	got := parsePorcelain([]byte("R  src/new.ts\x00src/old.ts\x00 M src/other.ts\x00"))
	want := []Entry{
		{IndexStatus: 'R', WorktreeStatus: ' ', Path: "src/new.ts", RenamedFrom: "src/old.ts"},
		{IndexStatus: ' ', WorktreeStatus: 'M', Path: "src/other.ts"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestFilterEditable_KeepsUntracked(t *testing.T) {
	in := parsePorcelain([]byte("?? new.txt\x00 M tracked.ts\x00"))
	got := filterEditable(in)
	want := []Entry{
		{IndexStatus: '?', WorktreeStatus: '?', Path: "new.txt"},
		{IndexStatus: ' ', WorktreeStatus: 'M', Path: "tracked.ts"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestFilterEditable_DropsIgnored(t *testing.T) {
	in := parsePorcelain([]byte("!! ignored.log\x00M  staged.ts\x00"))
	got := filterEditable(in)
	want := []Entry{{IndexStatus: 'M', WorktreeStatus: ' ', Path: "staged.ts"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestStatusMapFrom_KeysByPath(t *testing.T) {
	entries := []Entry{
		{IndexStatus: ' ', WorktreeStatus: 'M', Path: "a.ts"},
		{IndexStatus: 'M', WorktreeStatus: ' ', Path: "b.ts"},
	}
	got := statusMapFrom(entries)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got["a.ts"].WorktreeStatus != 'M' {
		t.Fatalf("a.ts: %+v", got["a.ts"])
	}
	if got["b.ts"].IndexStatus != 'M' {
		t.Fatalf("b.ts: %+v", got["b.ts"])
	}
}

func TestStatusMapFrom_EmptyInput(t *testing.T) {
	got := statusMapFrom(nil)
	if got == nil {
		t.Fatalf("want non-nil empty map")
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestFilterEditable_KeepsEveryEditableCode(t *testing.T) {
	codes := []byte{'M', 'A', 'D', 'R', 'C', 'T', 'U', '?'}
	for _, c := range codes {
		got := filterEditable([]Entry{{IndexStatus: c, WorktreeStatus: ' ', Path: "x"}})
		if len(got) != 1 {
			t.Fatalf("dropped index=%c", c)
		}
		got = filterEditable([]Entry{{IndexStatus: ' ', WorktreeStatus: c, Path: "x"}})
		if len(got) != 1 {
			t.Fatalf("dropped worktree=%c", c)
		}
	}
}
