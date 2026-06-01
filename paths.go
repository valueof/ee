package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func resolveInRepo(lines []string, repoRoot, cwd string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		abs := line
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(cwd, abs)
		}
		canon, err := filepath.EvalSymlinks(abs)
		if err != nil {
			continue
		}
		info, err := os.Stat(canon)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		rel, err := filepath.Rel(repoRoot, canon)
		if err != nil {
			continue
		}
		if rel == ".." || rel == "." || filepath.IsAbs(rel) ||
			(len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator)) {
			continue
		}
		if seen[rel] {
			continue
		}
		seen[rel] = true
		out = append(out, rel)
	}
	return out
}

func readStdin(r io.Reader) ([]string, error) {
	br := bufio.NewReader(r)
	var out []string
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\n")
			line = strings.TrimRight(line, "\r")
			line = strings.TrimSpace(line)
			if line != "" {
				out = append(out, line)
			}
		}
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

func annotate(paths []string, statusMap map[string]Entry) []Entry {
	out := make([]Entry, 0, len(paths))
	for _, p := range paths {
		if e, ok := statusMap[p]; ok {
			out = append(out, e)
			continue
		}
		out = append(out, Entry{IndexStatus: ' ', WorktreeStatus: ' ', Path: p})
	}
	return out
}
