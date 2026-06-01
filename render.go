package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	renderReset    = "\x1b[0m"
	renderFgRed    = "\x1b[31m"
	renderFgGreen  = "\x1b[32m"
	renderFgYellow = "\x1b[33m"
)

func paint(code, s string, enabled bool) string {
	if !enabled {
		return s
	}
	return code + s + renderReset
}

func statusCell(e Entry, color bool) string {
	idx, wt := e.IndexStatus, e.WorktreeStatus
	if idx == '?' && wt == '?' {
		return paint(renderFgRed, "??", color)
	}
	if idx != ' ' && wt != ' ' {
		return paint(renderFgYellow, string([]byte{idx, wt}), color)
	}
	if idx != ' ' {
		return paint(renderFgGreen, string(idx), color) + string(wt)
	}
	if wt != ' ' {
		return string(idx) + paint(renderFgRed, string(wt), color)
	}
	return string([]byte{idx, wt})
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

func renderFrame(entries []Entry, cursor int, width int, color bool, header string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%d)\n\n", header, len(entries))

	// prefixCols: 2 (cursor) + 3 (badge) + 1 (space) + 2 (status) + 1 (space) = 9
	const prefixCols = 9
	for i, e := range entries {
		cur := "  "
		if i == cursor {
			cur = "> "
		}
		badge := "[ ]"
		if i < 9 {
			badge = fmt.Sprintf("[%d]", i+1)
		}
		status := statusCell(e, color)
		raw := e.Path
		if e.RenamedFrom != "" {
			raw = e.RenamedFrom + " → " + e.Path
		}
		budget := width - prefixCols
		if budget < 0 {
			budget = 0
		}
		path := truncate(raw, budget)
		fmt.Fprintf(&b, "%s%s %s %s\n", cur, badge, status, path)
	}

	b.WriteString("\n↑/↓ or j/k  •  1-9 open  •  Enter open  •  q quit")
	return b.String()
}
