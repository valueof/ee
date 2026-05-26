package main

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
