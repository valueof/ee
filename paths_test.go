package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadStdin_PlainLines(t *testing.T) {
	got, err := readStdin(strings.NewReader("a.ts\nb.ts\nc.ts\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.ts", "b.ts", "c.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestReadStdin_NoTrailingNewline(t *testing.T) {
	got, err := readStdin(strings.NewReader("a.ts\nb.ts"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.ts", "b.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestReadStdin_CRLF(t *testing.T) {
	got, err := readStdin(strings.NewReader("a.ts\r\nb.ts\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.ts", "b.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestReadStdin_BlankAndWhitespace(t *testing.T) {
	got, err := readStdin(strings.NewReader("\na.ts\n   \n  b.ts  \n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.ts", "b.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestReadStdin_Empty(t *testing.T) {
	got, err := readStdin(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %#v, want empty", got)
	}
}

func TestAnnotate_PathInMap(t *testing.T) {
	m := map[string]Entry{
		"a.ts": {IndexStatus: ' ', WorktreeStatus: 'M', Path: "a.ts"},
	}
	got := annotate([]string{"a.ts"}, m)
	want := []Entry{{IndexStatus: ' ', WorktreeStatus: 'M', Path: "a.ts"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestAnnotate_PathMissingFromMap(t *testing.T) {
	got := annotate([]string{"a.ts"}, map[string]Entry{})
	want := []Entry{{IndexStatus: ' ', WorktreeStatus: ' ', Path: "a.ts"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestAnnotate_PreservesInputOrder(t *testing.T) {
	m := map[string]Entry{
		"a.ts": {IndexStatus: ' ', WorktreeStatus: 'M', Path: "a.ts"},
		"b.ts": {IndexStatus: 'M', WorktreeStatus: ' ', Path: "b.ts"},
	}
	got := annotate([]string{"b.ts", "a.ts", "c.ts"}, m)
	if got[0].Path != "b.ts" || got[1].Path != "a.ts" || got[2].Path != "c.ts" {
		t.Fatalf("order: got %#v", got)
	}
	if got[2].WorktreeStatus != ' ' || got[2].IndexStatus != ' ' {
		t.Fatalf("c.ts should be clean: %+v", got[2])
	}
}
