//go:build !go1.20
// +build !go1.20

package compat

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

// FmtFormatString is a compatibility wrapper for fmt.FormatString, added in
// go1.20.
func FmtFormatString(state fmt.State, verb rune) string {
	var tmp [16]byte // Use a local buffer.
	b := append(tmp[:0], '%')
	for _, c := range " +-#0" {
		if state.Flag(int(c)) {
			b = append(b, byte(c))
		}
	}
	if w, ok := state.Width(); ok {
		b = strconv.AppendInt(b, int64(w), 10)
	}
	if p, ok := state.Precision(); ok {
		b = append(b, '.')
		b = strconv.AppendInt(b, int64(p), 10)
	}
	if verb < utf8.RuneSelf {
		b = append(b, byte(verb))
		return string(b)
	}
	var verbb [utf8.UTFMax]byte
	n := utf8.EncodeRune(verbb[:], verb)
	b = append(b, verbb[:n]...)
	return string(b)
}
