//go:build go1.20
// +build go1.20

package compat

import "fmt"

// FmtFormatString is a compatibility wrapper for fmt.FormatString, added in go1.20.
func FmtFormatString(state fmt.State, verb rune) string {
	return fmt.FormatString(state, verb)
}
