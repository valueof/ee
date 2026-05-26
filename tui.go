package main

import (
	"os"

	"golang.org/x/term"
)

type KeyKind int

const (
	KeyUnknown KeyKind = iota
	KeyUp
	KeyDown
	KeyEnter
	KeyQuit
	KeyDigit
	KeyChar
)

type Key struct {
	Kind  KeyKind
	Value int  // populated for KeyDigit
	Char  byte // populated for KeyChar
}

func parseKey(b []byte) Key {
	if len(b) == 0 {
		return Key{Kind: KeyUnknown}
	}
	b0 := b[0]
	if b0 == 0x03 {
		return Key{Kind: KeyQuit}
	}
	if b0 == 0x0d || b0 == 0x0a {
		return Key{Kind: KeyEnter}
	}
	if b0 == 0x1b {
		if len(b) == 1 {
			return Key{Kind: KeyQuit}
		}
		if len(b) >= 3 && b[1] == 0x5b {
			if b[2] == 0x41 {
				return Key{Kind: KeyUp}
			}
			if b[2] == 0x42 {
				return Key{Kind: KeyDown}
			}
		}
		return Key{Kind: KeyUnknown}
	}
	if b0 == 'q' {
		return Key{Kind: KeyQuit}
	}
	if b0 >= '1' && b0 <= '9' {
		return Key{Kind: KeyDigit, Value: int(b0 - '0')}
	}
	return Key{Kind: KeyChar, Char: b0}
}

const (
	tuiEsc        = "\x1b"
	tuiEnterAlt   = tuiEsc + "[?1049h"
	tuiExitAlt    = tuiEsc + "[?1049l"
	tuiHideCursor = tuiEsc + "[?25l"
	tuiShowCursor = tuiEsc + "[?25h"
	tuiClear      = tuiEsc + "[2J"
	tuiHome       = tuiEsc + "[H"
)

var (
	rawEntered bool
	oldState   *term.State
)

func tuiEnter() error {
	if rawEntered {
		return nil
	}
	st, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	oldState = st
	_, _ = os.Stdout.WriteString(tuiEnterAlt + tuiHideCursor)
	rawEntered = true
	return nil
}

func tuiExit() {
	if !rawEntered {
		return
	}
	_, _ = os.Stdout.WriteString(tuiShowCursor + tuiExitAlt)
	if oldState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}
	rawEntered = false
}

func tuiRender(s string) {
	_, _ = os.Stdout.WriteString(tuiHome + tuiClear + s)
}

func terminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80
	}
	return w
}

func readKey() Key {
	buf := make([]byte, 8)
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return Key{Kind: KeyQuit}
	}
	return parseKey(buf[:n])
}
