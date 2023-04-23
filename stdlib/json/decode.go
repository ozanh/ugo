// A modified version of Go's json implementation.

// Copyright (c) 2022-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.golang file.

package json

import (
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/ozanh/ugo"
)

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
func Unmarshal(data []byte) (ugo.Object, error) {
	// Check for well-formedness.
	// Avoids filling out half a data structure
	// before discovering a JSON syntax error.
	var d decodeState
	err := checkValid(data, &d.scan)
	if err != nil {
		return ugo.Undefined, err
	}

	return d.init(data).unmarshal()
}

func (d *decodeState) unmarshal() (ugo.Object, error) {
	d.scan.reset()
	d.scanWhile(scanSkipSpace)

	v, err := d.value()
	if v == nil {
		v = ugo.Undefined
	}
	return v, err
}

// decodeState represents the state while decoding a JSON value.
type decodeState struct {
	data   []byte
	off    int // next read offset in data
	opcode int // last read result
	scan   scanner
}

// readIndex returns the position of the last byte read.
func (d *decodeState) readIndex() int {
	return d.off - 1
}

// phasePanicMsg is used as a panic message when we end up with something that
// shouldn't happen. It can indicate a bug in the JSON decoder, or that
// something is editing the data slice while the decoder executes.
const phasePanicMsg = "JSON decoder out of sync - data changing underfoot?"

func (d *decodeState) init(data []byte) *decodeState {
	d.data = data
	d.off = 0
	return d
}

// scanNext processes the byte at d.data[d.off].
func (d *decodeState) scanNext() {
	if d.off < len(d.data) {
		d.opcode = d.scan.step(&d.scan, d.data[d.off])
		d.off++
	} else {
		d.opcode = d.scan.eof()
		d.off = len(d.data) + 1 // mark processed EOF with len+1
	}
}

// scanWhile processes bytes in d.data[d.off:] until it
// receives a scan code not equal to op.
func (d *decodeState) scanWhile(op int) {
	s, data, i := &d.scan, d.data, d.off
	for i < len(data) {
		newOp := s.step(s, data[i])
		i++
		if newOp != op {
			d.opcode = newOp
			d.off = i
			return
		}
	}

	d.off = len(data) + 1 // mark processed EOF with len+1
	d.opcode = d.scan.eof()
}

func (d *decodeState) value() (ugo.Object, error) {
	switch d.opcode {
	default:
		panic(phasePanicMsg)
	case scanBeginArray:
		val, err := d.array()
		if err != nil {
			return nil, err
		}
		d.scanNext()
		return val, nil

	case scanBeginObject:
		val, err := d.object()
		if err != nil {
			return nil, err
		}
		d.scanNext()
		return val, nil

	case scanBeginLiteral:
		val, err := d.literal()
		if err != nil {
			return nil, err
		}
		return val, nil
	}
}

func (d *decodeState) array() (ugo.Object, error) {
	object := ugo.Array{}
	for {
		// Look ahead for ] - can only happen on first iteration.
		d.scanWhile(scanSkipSpace)
		if d.opcode == scanEndArray {
			break
		}

		v, err := d.value()
		if err != nil {
			return nil, err
		}
		object = append(object, v)

		// Next token must be , or ].
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		if d.opcode == scanEndArray {
			break
		}
		if d.opcode != scanArrayValue {
			panic(phasePanicMsg)
		}
	}
	return object, nil
}

func (d *decodeState) object() (ugo.Object, error) {
	object := ugo.Map{}
	for {
		// Read opening " of string key or closing }.
		d.scanWhile(scanSkipSpace)
		if d.opcode == scanEndObject {
			// closing } - can only happen on first iteration.
			break
		}
		if d.opcode != scanBeginLiteral {
			panic(phasePanicMsg)
		}

		// Read string key.
		start := d.readIndex()
		d.scanWhile(scanContinue)
		item := d.data[start:d.readIndex()]
		key, ok := unquote(item)
		if !ok {
			panic(phasePanicMsg)
		}

		// Read : before value.
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		if d.opcode != scanObjectKey {
			panic(phasePanicMsg)
		}
		d.scanWhile(scanSkipSpace)

		// Read value.
		v, err := d.value()
		if err != nil {
			return object, err
		}
		object[key] = v

		// Next token must be , or }.
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		if d.opcode == scanEndObject {
			break
		}
		if d.opcode != scanObjectValue {
			panic(phasePanicMsg)
		}
	}
	return object, nil
}

func (d *decodeState) literal() (ugo.Object, error) {
	// All bytes inside literal return scanContinue op code.
	start := d.readIndex()
	d.scanWhile(scanContinue)

	item := d.data[start:d.readIndex()]

	switch c := item[0]; c {
	case 'n': // null
		return ugo.Undefined, nil

	case 't', 'f': // true, false
		if c == 't' {
			return ugo.True, nil
		}
		return ugo.False, nil

	case '"': // string
		s, ok := unquote(item)
		if !ok {
			panic(phasePanicMsg)
		}
		return ugo.String(s), nil

	default: // number
		if c != '-' && (c < '0' || c > '9') {
			panic(phasePanicMsg)
		}

		n, err := strconv.ParseFloat(string(item), 64)
		return ugo.Float(n), err
	}
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range s[2:6] {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}

// unquote converts a quoted JSON string literal s into an actual string t.
// The rules are different than for Go, so cannot use strconv.Unquote.
func unquote(s []byte) (t string, ok bool) {
	s, ok = unquoteBytes(s)
	t = string(s)
	return
}

func unquoteBytes(s []byte) (t []byte, ok bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return
	}
	s = s[1 : len(s)-1]

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room? Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}
