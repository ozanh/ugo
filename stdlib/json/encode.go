// A modified version of Go's json implementation.

// Copyright (c) 2022-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.golang file.

package json

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"unicode/utf8"

	"github.com/ozanh/ugo"
)

// Marshal returns the JSON encoding of v.
func Marshal(v ugo.Object) ([]byte, error) {
	e := newEncodeState()

	err := e.marshal(v, encOpts{escapeHTML: true})
	if err != nil {
		return nil, err
	}
	buf := append([]byte(nil), e.Bytes()...)

	return buf, nil
}

// MarshalIndent is like Marshal but applies Indent to format the output.
// Each JSON element in the output will begin on a new line beginning with prefix
// followed by one or more copies of indent according to the indentation nesting.
func MarshalIndent(v ugo.Object, prefix, indent string) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = indentBuffer(&buf, b, prefix, indent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Marshaler is the interface implemented by types that
// can marshal themselves into valid JSON.
type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

// An UnsupportedValueError is returned by Marshal when attempting
// to encode an unsupported value.
type UnsupportedValueError struct {
	Object ugo.Object
	Str    string
}

func (e *UnsupportedValueError) Error() string {
	return "json: unsupported value: " + e.Str
}

// A MarshalerError represents an error from calling a MarshalJSON or MarshalText method.
type MarshalerError struct {
	Object     ugo.Object
	Err        error
	sourceFunc string
}

func (e *MarshalerError) Error() string {
	srcFunc := e.sourceFunc
	if srcFunc == "" {
		srcFunc = "MarshalJSON"
	}
	return "json: error calling " + srcFunc +
		" for type " + e.Object.TypeName() +
		": " + e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *MarshalerError) Unwrap() error { return e.Err }

const hex = "0123456789abcdef"
const startDetectingCyclesAfter = 1000

// An encodeState encodes JSON into a bytes.Buffer.
type encodeState struct {
	bytes.Buffer // accumulated output
	scratch      [64]byte

	// Keep track of what pointers we've seen in the current recursive call
	// path, to avoid cycles that could lead to a stack overflow. Only do
	// the relatively expensive map operations if ptrLevel is larger than
	// startDetectingCyclesAfter, so that we skip the work if we're within a
	// reasonable amount of nested pointers deep.
	ptrLevel uint
	ptrSeen  map[interface{}]struct{}
}

func newEncodeState() *encodeState {
	return &encodeState{ptrSeen: make(map[interface{}]struct{})}
}

// jsonError is an error wrapper type for internal use only.
// Panics with errors are wrapped in jsonError so that the top-level recover
// can distinguish intentional panics from this package.
type jsonError struct{ error }

func (e *encodeState) marshal(v ugo.Object, opts encOpts) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(jsonError); ok {
				err = je.error
			} else {
				panic(r)
			}
		}
	}()
	e.encode(v, opts)
	return nil
}

// error aborts the encoding by panicking with err wrapped in jsonError.
func (e *encodeState) error(err error) {
	panic(jsonError{err})
}

func (e *encodeState) encode(v ugo.Object, opts encOpts) {
	objectEncoder(v)(e, v, opts)
}

type encOpts struct {
	// quoted causes primitive fields to be encoded inside JSON strings.
	quoted bool
	// escapeHTML causes '<', '>', and '&' to be escaped in JSON strings.
	escapeHTML bool
}

type encoderFunc func(e *encodeState, v ugo.Object, opts encOpts)

// objectEncoder constructs an encoderFunc for a ugo.Object.
func objectEncoder(v ugo.Object) encoderFunc {
	switch v.(type) {
	case ugo.Bool:
		return boolEncoder
	case ugo.Int:
		return intEncoder
	case ugo.Uint:
		return uintEncoder
	case ugo.Float:
		return floatEncoder
	case ugo.String:
		return stringEncoder
	case ugo.Bytes:
		return bytesEncoder
	case ugo.Map, *ugo.SyncMap:
		return mapEncoder
	case ugo.Array:
		return arrayEncoder
	case ugo.Char:
		return charEncoder
	case *EncoderOptions:
		return optionsEncoder
	case *ugo.ObjectPtr:
		return objectPtrEncoder
	case *ugo.UndefinedType:
		return invalidValueEncoder
	case encoding.TextMarshaler:
		return textMarshalerEncoder
	case Marshaler:
		return marshalerEncoder
	default:
		return noopEncoder
	}
}

func invalidValueEncoder(e *encodeState, _ ugo.Object, _ encOpts) {
	e.WriteString("null")
}

func noopEncoder(_ *encodeState, _ ugo.Object, _ encOpts) {}

func optionsEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	opts.quoted = v.(*EncoderOptions).Quote
	opts.escapeHTML = v.(*EncoderOptions).EscapeHTML
	e.encode(v.(*EncoderOptions).Value, opts)
}

func boolEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	if opts.quoted {
		e.WriteByte('"')
	}
	if v.(ugo.Bool) {
		e.WriteString("true")
	} else {
		e.WriteString("false")
	}
	if opts.quoted {
		e.WriteByte('"')
	}
}

func intEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	b := strconv.AppendInt(e.scratch[:0], int64(v.(ugo.Int)), 10)
	if opts.quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if opts.quoted {
		e.WriteByte('"')
	}
}

func uintEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	b := strconv.AppendUint(e.scratch[:0], uint64(v.(ugo.Uint)), 10)
	if opts.quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if opts.quoted {
		e.WriteByte('"')
	}
}

func floatEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	f := float64(v.(ugo.Float))
	if math.IsInf(f, 0) || math.IsNaN(f) {
		e.error(&UnsupportedValueError{v, strconv.FormatFloat(f, 'g', -1, 64)})
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	b := e.scratch[:0]
	abs := math.Abs(f)
	fmt := byte('f')

	if abs != 0 {
		if abs < 1e-6 || abs >= 1e21 {
			fmt = 'e'
		}
	}
	b = strconv.AppendFloat(b, f, fmt, -1, 64)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}

	if opts.quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if opts.quoted {
		e.WriteByte('"')
	}
}

func charEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	b := strconv.AppendInt(e.scratch[:0], int64(v.(ugo.Char)), 10)
	if opts.quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if opts.quoted {
		e.WriteByte('"')
	}
}

func stringEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	if opts.quoted {
		e2 := newEncodeState()
		// Since we encode the string twice, we only need to escape HTML
		// the first time.
		e2.string(v.String(), opts.escapeHTML)
		e.stringBytes(e2.Bytes(), false)
	} else {
		e.string(v.String(), opts.escapeHTML)
	}
}

func mapEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	if v == nil {
		e.WriteString("null")
		return
	}
	if e.ptrLevel++; e.ptrLevel > startDetectingCyclesAfter {
		// Start checking if we've run into a pointer cycle.
		var ptr interface{}
		if _, ok := v.(ugo.Map); ok {
			ptr = reflect.ValueOf(v).Pointer()
		} else { // *SyncMap
			ptr = v
		}
		if _, ok := e.ptrSeen[ptr]; ok {
			e.error(&UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.TypeName())})
		}
		e.ptrSeen[ptr] = struct{}{}
		defer delete(e.ptrSeen, ptr)
	}

	var m ugo.Map
	var ok bool
	if m, ok = v.(ugo.Map); !ok {
		sm := v.(*ugo.SyncMap)
		if sm == nil {
			e.WriteString("null")
			e.ptrLevel--
			return
		}
		sm.RLock()
		defer sm.RUnlock()
		m = sm.Value
	}
	if m == nil {
		e.WriteString("null")
		e.ptrLevel--
		return
	}
	e.WriteByte('{')
	// Extract and sort the keys.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, kv := range keys {
		if i > 0 {
			e.WriteByte(',')
		}
		e.string(kv, opts.escapeHTML)
		e.WriteByte(':')
		e.encode(m[kv], opts)
	}
	e.WriteByte('}')
	e.ptrLevel--
}

func bytesEncoder(e *encodeState, v ugo.Object, _ encOpts) {
	if v == nil {
		e.WriteString("null")
		return
	}
	s := v.(ugo.Bytes)
	e.WriteByte('"')
	encodedLen := base64.StdEncoding.EncodedLen(len(s))
	if encodedLen <= len(e.scratch) {
		// If the encoded bytes fit in e.scratch, avoid an extra
		// allocation and use the cheaper Encoding.Encode.
		dst := e.scratch[:encodedLen]
		base64.StdEncoding.Encode(dst, s)
		e.Write(dst)
	} else if encodedLen <= 1024 {
		// The encoded bytes are short enough to allocate for, and
		// Encoding.Encode is still cheaper.
		dst := make([]byte, encodedLen)
		base64.StdEncoding.Encode(dst, s)
		e.Write(dst)
	} else {
		// The encoded bytes are too long to cheaply allocate, and
		// Encoding.Encode is no longer noticeably cheaper.
		enc := base64.NewEncoder(base64.StdEncoding, e)
		_, _ = enc.Write(s)
		_ = enc.Close()
	}
	e.WriteByte('"')
}

func arrayEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	if v == nil {
		e.WriteString("null")
		return
	}
	if e.ptrLevel++; e.ptrLevel > startDetectingCyclesAfter {
		// Start checking if we've run into a pointer cycle.
		// Here we use a struct to memorize the pointer to the first element of the slice
		// and its length.
		rval := reflect.ValueOf(v)
		ptr := struct {
			ptr uintptr
			len int
		}{rval.Pointer(), rval.Len()}
		if _, ok := e.ptrSeen[ptr]; ok {
			e.error(&UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.TypeName())})
		}
		e.ptrSeen[ptr] = struct{}{}
		defer delete(e.ptrSeen, ptr)
	}
	arr := v.(ugo.Array)
	if arr == nil {
		e.WriteString("null")
		e.ptrLevel--
		return
	}
	e.WriteByte('[')
	n := len(arr)
	for i := 0; i < n; i++ {
		if i > 0 {
			e.WriteByte(',')
		}
		e.encode(arr[i], opts)
	}
	e.WriteByte(']')
	e.ptrLevel--
}

func objectPtrEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	if v == nil {
		e.WriteString("null")
		return
	}
	if e.ptrLevel++; e.ptrLevel > startDetectingCyclesAfter {
		// Start checking if we've run into a pointer cycle.
		if _, ok := e.ptrSeen[v]; ok {
			e.error(&UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.TypeName())})
		}
		e.ptrSeen[v] = struct{}{}
		defer delete(e.ptrSeen, v)
	}
	vv := v.(*ugo.ObjectPtr).Value
	if vv == nil {
		e.WriteString("null")
	} else {
		e.encode(*vv, opts)
	}
	e.ptrLevel--
}

func textMarshalerEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	if v == nil {
		e.WriteString("null")
		return
	}
	m, ok := v.(encoding.TextMarshaler)
	if !ok || m == nil {
		e.WriteString("null")
		return
	}
	b, err := m.MarshalText()
	if err != nil {
		e.error(&MarshalerError{v, err, "MarshalText"})
	}
	e.stringBytes(b, opts.escapeHTML)
}

func marshalerEncoder(e *encodeState, v ugo.Object, opts encOpts) {
	if v == nil {
		e.WriteString("null")
		return
	}
	m, ok := v.(Marshaler)
	if !ok || m == nil {
		e.WriteString("null")
		return
	}
	b, err := m.MarshalJSON()
	if err == nil {
		// copy JSON into buffer, checking validity.
		err = compact(&e.Buffer, b, opts.escapeHTML)
	}
	if err != nil {
		e.error(&MarshalerError{v, err, "MarshalJSON"})
	}
}

// NOTE: keep in sync with stringBytes below.
func (e *encodeState) string(s string, escapeHTML bool) {
	e.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
				i++
				continue
			}
			if start < i {
				e.WriteString(s[start:i])
			}
			e.WriteByte('\\')
			switch b {
			case '\\', '"':
				e.WriteByte(b)
			case '\n':
				e.WriteByte('n')
			case '\r':
				e.WriteByte('r')
			case '\t':
				e.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				e.WriteString(`u00`)
				e.WriteByte(hex[b>>4])
				e.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e.WriteString(s[start:i])
			}
			e.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				e.WriteString(s[start:i])
			}
			e.WriteString(`\u202`)
			e.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.WriteString(s[start:])
	}
	e.WriteByte('"')
}

// NOTE: keep in sync with string above.
func (e *encodeState) stringBytes(s []byte, escapeHTML bool) {
	e.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
				i++
				continue
			}
			if start < i {
				e.Write(s[start:i])
			}
			e.WriteByte('\\')
			switch b {
			case '\\', '"':
				e.WriteByte(b)
			case '\n':
				e.WriteByte('n')
			case '\r':
				e.WriteByte('r')
			case '\t':
				e.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				e.WriteString(`u00`)
				e.WriteByte(hex[b>>4])
				e.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRune(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e.Write(s[start:i])
			}
			e.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				e.Write(s[start:i])
			}
			e.WriteString(`\u202`)
			e.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.Write(s[start:])
	}
	e.WriteByte('"')
}
