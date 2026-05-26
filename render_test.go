package main

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

func entry(idx, wt byte, path string) Entry {
	return Entry{IndexStatus: idx, WorktreeStatus: wt, Path: path}
}

const (
	fgGreen  = "\x1b[32m"
	fgRed    = "\x1b[31m"
	fgYellow = "\x1b[33m"
	reset    = "\x1b[0m"
)

func TestRenderFrame_HeaderShowsCount(t *testing.T) {
	out := renderFrame([]Entry{entry(' ', 'M', "a.ts"), entry('M', ' ', "b.ts")}, 0, 80, false)
	if !strings.Contains(out, "modified files (2)") {
		t.Fatalf("missing header in %q", out)
	}
}

func TestRenderFrame_CursorRowPrefix(t *testing.T) {
	out := renderFrame([]Entry{entry(' ', 'M', "a.ts"), entry(' ', 'M', "b.ts")}, 1, 80, false)
	re := regexp.MustCompile(`(?m)^[ >] {1,2}\[`)
	matches := re.FindAllString(out, -1)
	if len(matches) != 2 {
		t.Fatalf("expected 2 entry lines, got %d in %q", len(matches), out)
	}
	lines := strings.Split(out, "\n")
	var entryLines []string
	for _, l := range lines {
		if re.MatchString(l) {
			entryLines = append(entryLines, l)
		}
	}
	if !strings.HasPrefix(entryLines[0], "  ") {
		t.Fatalf("first entry should start with '  ', got %q", entryLines[0])
	}
	if !strings.HasPrefix(entryLines[1], "> ") {
		t.Fatalf("second entry should start with '> ', got %q", entryLines[1])
	}
}

func TestRenderFrame_NumericBadges(t *testing.T) {
	entries := make([]Entry, 11)
	for i := range entries {
		entries[i] = entry(' ', 'M', "f.ts")
	}
	out := renderFrame(entries, 0, 80, false)
	for _, want := range []string{"[1]", "[9]", "[ ] "} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}

func TestRenderFrame_EmptyList(t *testing.T) {
	out := renderFrame(nil, 0, 80, false)
	if !strings.Contains(out, "modified files (0)") {
		t.Fatalf("missing header in %q", out)
	}
	if !strings.Contains(out, "1-9 open") {
		t.Fatalf("missing footer in %q", out)
	}
}

func TestRenderFrame_Footer(t *testing.T) {
	out := renderFrame([]Entry{entry(' ', 'M', "a.ts")}, 0, 80, false)
	for _, want := range []string{"↑/↓ or j/k", "q quit"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}

func TestRenderFrame_StagedGreen(t *testing.T) {
	out := renderFrame([]Entry{entry('M', ' ', "a.ts")}, 0, 80, true)
	if !strings.Contains(out, fgGreen+"M"+reset) {
		t.Fatalf("missing green M in %q", out)
	}
}

func TestRenderFrame_UnstagedRed(t *testing.T) {
	out := renderFrame([]Entry{entry(' ', 'M', "a.ts")}, 0, 80, true)
	if !strings.Contains(out, fgRed+"M"+reset) {
		t.Fatalf("missing red M in %q", out)
	}
}

func TestRenderFrame_BothYellow(t *testing.T) {
	out := renderFrame([]Entry{entry('M', 'M', "a.ts")}, 0, 80, true)
	if !strings.Contains(out, fgYellow+"MM"+reset) {
		t.Fatalf("missing yellow MM in %q", out)
	}
}

func TestRenderFrame_UntrackedRed(t *testing.T) {
	out := renderFrame([]Entry{entry('?', '?', "new.txt")}, 0, 80, true)
	if !strings.Contains(out, fgRed+"??"+reset) {
		t.Fatalf("missing red ?? in %q", out)
	}
}

func TestRenderFrame_NoEscapesWhenColorDisabled(t *testing.T) {
	out := renderFrame(
		[]Entry{entry('M', 'M', "a.ts"), entry(' ', 'M', "b.ts")},
		0, 80, false,
	)
	if regexp.MustCompile(`\x1b\[\d+m`).MatchString(out) {
		t.Fatalf("found SGR escape in colorless output %q", out)
	}
}

func TestRenderFrame_LongPathTruncated(t *testing.T) {
	long := "a/very/deeply/nested/file/path/that/is/quite/long.ts"
	out := renderFrame([]Entry{entry(' ', 'M', long)}, 0, 30, false)
	var line string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "[1]") {
			line = l
			break
		}
	}
	if utf8.RuneCountInString(line) > 30 {
		t.Fatalf("line %q has rune count %d, want <= 30", line, utf8.RuneCountInString(line))
	}
	if !strings.Contains(line, "…") {
		t.Fatalf("line %q missing ellipsis", line)
	}
}

func TestRenderFrame_ShortPathNotTruncated(t *testing.T) {
	out := renderFrame([]Entry{entry(' ', 'M', "x.ts")}, 0, 30, false)
	var line string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "[1]") {
			line = l
			break
		}
	}
	if strings.Contains(line, "…") {
		t.Fatalf("line %q should not be truncated", line)
	}
	if !strings.Contains(line, "x.ts") {
		t.Fatalf("line %q missing path", line)
	}
}
