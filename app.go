package main

import "os"

type state struct {
	repoRoot    string
	entries     []Entry
	cursor      int
	header      string
	loadEntries func() ([]Entry, error)
}

func colorEnabled() bool {
	_, set := os.LookupEnv("NO_COLOR")
	return !set
}

func clampCursor(cursor, length int) int {
	if length == 0 {
		return 0
	}
	if cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}

func envMap() map[string]string {
	out := make(map[string]string, len(os.Environ()))
	for _, kv := range os.Environ() {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				out[kv[:i]] = kv[i+1:]
				break
			}
		}
	}
	return out
}

func openSelected(s state) (state, error) {
	editor, err := resolveEditor(envMap())
	if err != nil {
		return s, err
	}
	entry := s.entries[s.cursor]
	full := s.repoRoot + "/" + entry.Path
	if err := runEditor(editor, full); err != nil {
		return s, err
	}
	next, err := s.loadEntries()
	if err != nil {
		return s, err
	}
	s.entries = next
	s.cursor = clampCursor(s.cursor, len(next))
	return s, nil
}

func appRun(repoRoot string, initial []Entry, loadEntries func() ([]Entry, error), header string) error {
	s := state{
		repoRoot:    repoRoot,
		entries:     initial,
		cursor:      0,
		header:      header,
		loadEntries: loadEntries,
	}
	color := colorEnabled()

	for {
		if len(s.entries) == 0 {
			tuiExit()
			os.Stdout.WriteString("Nothing to edit\n")
			return nil
		}

		tuiRender(renderFrame(s.entries, s.cursor, terminalWidth(), color, s.header))

		key := readKey()
		switch key.Kind {
		case KeyQuit:
			return nil
		case KeyUp:
			s.cursor = (s.cursor - 1 + len(s.entries)) % len(s.entries)
		case KeyDown:
			s.cursor = (s.cursor + 1) % len(s.entries)
		case KeyChar:
			switch key.Char {
			case 'j':
				s.cursor = (s.cursor + 1) % len(s.entries)
			case 'k':
				s.cursor = (s.cursor - 1 + len(s.entries)) % len(s.entries)
			}
		case KeyDigit:
			if key.Value >= 1 && key.Value <= len(s.entries) {
				s.cursor = key.Value - 1
				next, err := openSelected(s)
				if err != nil {
					return err
				}
				s = next
			}
		case KeyEnter:
			next, err := openSelected(s)
			if err != nil {
				return err
			}
			s = next
		case KeyUnknown:
		}
	}
}
