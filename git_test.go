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
