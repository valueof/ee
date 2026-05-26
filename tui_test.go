package main

import "testing"

func TestParseKey_CtrlC(t *testing.T) {
	if k := parseKey([]byte{0x03}); k.Kind != KeyQuit {
		t.Fatalf("got %+v, want quit", k)
	}
}

func TestParseKey_q(t *testing.T) {
	if k := parseKey([]byte{'q'}); k.Kind != KeyQuit {
		t.Fatalf("got %+v, want quit", k)
	}
}

func TestParseKey_BareEsc(t *testing.T) {
	if k := parseKey([]byte{0x1b}); k.Kind != KeyQuit {
		t.Fatalf("got %+v, want quit", k)
	}
}

func TestParseKey_EnterCR(t *testing.T) {
	if k := parseKey([]byte{0x0d}); k.Kind != KeyEnter {
		t.Fatalf("got %+v, want enter", k)
	}
}

func TestParseKey_EnterLF(t *testing.T) {
	if k := parseKey([]byte{0x0a}); k.Kind != KeyEnter {
		t.Fatalf("got %+v, want enter", k)
	}
}

func TestParseKey_Up(t *testing.T) {
	if k := parseKey([]byte{0x1b, 0x5b, 0x41}); k.Kind != KeyUp {
		t.Fatalf("got %+v, want up", k)
	}
}

func TestParseKey_Down(t *testing.T) {
	if k := parseKey([]byte{0x1b, 0x5b, 0x42}); k.Kind != KeyDown {
		t.Fatalf("got %+v, want down", k)
	}
}

func TestParseKey_Digits1To9(t *testing.T) {
	for n := 1; n <= 9; n++ {
		k := parseKey([]byte{byte('0' + n)})
		if k.Kind != KeyDigit || k.Value != n {
			t.Fatalf("n=%d: got %+v", n, k)
		}
	}
}

func TestParseKey_jAndk(t *testing.T) {
	if k := parseKey([]byte{'j'}); k.Kind != KeyChar || k.Char != 'j' {
		t.Fatalf("j: got %+v", k)
	}
	if k := parseKey([]byte{'k'}); k.Kind != KeyChar || k.Char != 'k' {
		t.Fatalf("k: got %+v", k)
	}
}

func TestParseKey_Zero(t *testing.T) {
	k := parseKey([]byte{'0'})
	if k.Kind != KeyChar || k.Char != '0' {
		t.Fatalf("got %+v, want char '0'", k)
	}
}

func TestParseKey_UnknownEscSeq(t *testing.T) {
	if k := parseKey([]byte{0x1b, 0x5b, 0x5a}); k.Kind != KeyUnknown {
		t.Fatalf("got %+v, want unknown", k)
	}
}

func TestParseKey_Empty(t *testing.T) {
	if k := parseKey(nil); k.Kind != KeyUnknown {
		t.Fatalf("got %+v, want unknown", k)
	}
}
