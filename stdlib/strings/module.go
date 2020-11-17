// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Package strings provides strings module implementing simple functions to
// manipulate UTF-8 encoded strings for uGO script language. It wraps Go's
// strings package functionalities.
//
package strings

import (
	"strconv"
	"strings"

	"github.com/ozanh/ugo"
)

// Module represents time module.
var Module = map[string]ugo.Object{
	// ugo:doc
	// # strings Module
	//
	// ## Functions
	// Contains(s string, substr string) -> bool
	// Reports whether substr is within s.
	"Contains": &ugo.Function{
		Name:  "Contains",
		Value: fnASSRB(strings.Contains),
	},
	// ugo:doc
	// ContainsAny(s string, chars string) -> bool
	// Reports whether any char in chars are within s.
	"ContainsAny": &ugo.Function{
		Name:  "ContainsAny",
		Value: fnASSRB(strings.ContainsAny),
	},
	// ugo:doc
	// ContainsChar(s string, c char) -> bool
	// Reports whether the char c is within s.
	"ContainsChar": &ugo.Function{
		Name:  "ContainsChar",
		Value: containsChar,
	},
	// ugo:doc
	// Count(s string, substr string) -> int
	// Counts the number of non-overlapping instances of substr in s.
	"Count": &ugo.Function{
		Name:  "Count",
		Value: fnASSRI(strings.Count),
	},
	// ugo:doc
	// EqualFold(s string, t string) -> bool
	// EqualFold reports whether s and t, interpreted as UTF-8 strings,
	// are equal under Unicode case-folding, which is a more general form of
	// case-insensitivity.
	"EqualFold": &ugo.Function{
		Name:  "EqualFold",
		Value: fnASSRB(strings.EqualFold),
	},
	// ugo:doc
	// Fields(s string) -> array
	// Splits the string s around each instance of one or more consecutive white
	// space characters, returning an array of substrings of s or an empty array
	// if s contains only white space.
	"Fields": &ugo.Function{
		Name:  "Fields",
		Value: fields,
	},
	// ugo:doc
	// HasPrefix(s string, prefix string) -> bool
	// Reports whether the string s begins with prefix.
	"HasPrefix": &ugo.Function{
		Name:  "HasPrefix",
		Value: fnASSRB(strings.HasPrefix),
	},
	// ugo:doc
	// HasSuffix(s string, suffix string) -> bool
	// Reports whether the string s ends with prefix.
	"HasSuffix": &ugo.Function{
		Name:  "HasSuffix",
		Value: fnASSRB(strings.HasSuffix),
	},
	// ugo:doc
	// Index(s string, substr string) -> int
	// Returns the index of the first instance of substr in s, or -1 if substr
	// is not present in s.
	"Index": &ugo.Function{
		Name:  "Index",
		Value: fnASSRI(strings.Index),
	},
	// ugo:doc
	// IndexAny(s string, chars string) -> int
	// Returns the index of the first instance of any char from chars in s, or
	// -1 if no char from chars is present in s.
	"IndexAny": &ugo.Function{
		Name:  "IndexAny",
		Value: fnASSRI(strings.IndexAny),
	},
	// ugo:doc
	// IndexByte(s string, c char|int) -> int
	// Returns the index of the first byte value of c in s, or -1 if byte value
	// of c is not present in s.
	"IndexByte": &ugo.Function{
		Name:  "IndexByte",
		Value: indexByte,
	},
	// ugo:doc
	// IndexChar(s string, c char) -> int
	// Returns the index of the first instance of the char c, or -1 if char is
	// not present in s.
	"IndexChar": &ugo.Function{
		Name:  "IndexChar",
		Value: indexChar,
	},
	// ugo:doc
	// Join(arr array, sep string) -> string
	// Concatenates the string values of array arr elements to create a
	// single string. The separator string sep is placed between elements in the
	// resulting string.
	"Join": &ugo.Function{
		Name:  "Join",
		Value: join,
	},
	// ugo:doc
	// LastIndex(s string, substr string) -> int
	// Returns the index of the last instance of substr in s, or -1 if substr
	// is not present in s.
	"LastIndex": &ugo.Function{
		Name:  "LastIndex",
		Value: fnASSRI(strings.LastIndex),
	},
	// ugo:doc
	// LastIndexAny(s string, chars string) -> int
	// Returns the index of the last instance of any char from chars in s, or
	// -1 if no char from chars is present in s.
	"LastIndexAny": &ugo.Function{
		Name:  "LastIndexAny",
		Value: fnASSRI(strings.LastIndexAny),
	},
	// ugo:doc
	// LastIndexByte(s string, c char|int) -> int
	// Returns the index of byte value of the last instance of c in s, or -1
	// if c is not present in s.
	"LastIndexByte": &ugo.Function{
		Name:  "LastIndexByte",
		Value: lastIndexByte,
	},
	// ugo:doc
	// PadLeft(s string, padLen int[, padWith any]) -> string
	// Returns a string that is padded on the left with the string `padWith` until
	// the `padLen` length is reached. If padWith is not given, a white space is
	// used as default padding.
	"PadLeft": &ugo.Function{
		Name:  "PadLeft",
		Value: padLeft,
	},
	// ugo:doc
	// PadRight(s string, padLen int[, padWith any]) -> string
	// Returns a string that is padded on the right with the string `padWith` until
	// the `padLen` length is reached. If padWith is not given, a white space is
	// used as default padding.
	"PadRight": &ugo.Function{
		Name:  "PadRight",
		Value: padRight,
	},
	// ugo:doc
	// Repeat(s string, count int) -> string
	// Returns a new string consisting of count copies of the string s.
	//
	// - If count is a negative int, it returns empty string.
	// - If (len(s) * count) overflows, it panics.
	"Repeat": &ugo.Function{
		Name:  "Repeat",
		Value: repeat,
	},
	// ugo:doc
	// Replace(s string, old string, new string[, n int]) -> string
	// Returns a copy of the string s with the first n non-overlapping instances
	// of old replaced by new. If n is not provided or -1, it replaces all
	// instances.
	"Replace": &ugo.Function{
		Name:  "Replace",
		Value: replace,
	},
	// ugo:doc
	// Split(s string, sep string[, n int]) -> [string]
	// Splits s into substrings separated by sep and returns an array of
	// the substrings between those separators.
	//
	// n determines the number of substrings to return:
	//
	// - n < 0: all substrings (default)
	// - n > 0: at most n substrings; the last substring will be the unsplit remainder.
	// - n == 0: the result is empty array
	"Split": &ugo.Function{
		Name:  "Split",
		Value: fnASSIRA(strings.SplitN),
	},
	// ugo:doc
	// SplitAfter(s string, sep string[, n int]) -> [string]
	// Slices s into substrings after each instance of sep and returns an array
	// of those substrings.
	//
	// n determines the number of substrings to return:
	//
	// - n < 0: all substrings (default)
	// - n > 0: at most n substrings; the last substring will be the unsplit remainder.
	// - n == 0: the result is empty array
	"SplitAfter": &ugo.Function{
		Name:  "SplitAfter",
		Value: fnASSIRA(strings.SplitAfterN),
	},
	// ugo:doc
	// Title(s string) -> string
	// Returns a copy of the string s with all Unicode letters that begin words
	// mapped to their Unicode title case.
	"Title": &ugo.Function{
		Name:  "Title",
		Value: fnASRS(strings.Title),
	},
	// ugo:doc
	// ToLower(s string) -> string
	// Returns s with all Unicode letters mapped to their lower case.
	"ToLower": &ugo.Function{
		Name:  "ToLower",
		Value: fnASRS(strings.ToLower),
	},
	// ugo:doc
	// ToTitle(s string) -> string
	// Returns a copy of the string s with all Unicode letters mapped to their
	// Unicode title case.
	"ToTitle": &ugo.Function{
		Name:  "ToTitle",
		Value: fnASRS(strings.ToTitle),
	},
	// ugo:doc
	// ToUpper(s string) -> string
	// Returns s with all Unicode letters mapped to their upper case.
	"ToUpper": &ugo.Function{
		Name:  "ToUpper",
		Value: fnASRS(strings.ToUpper),
	},
	// ugo:doc
	// Trim(s, cutset string) -> string
	// Returns a slice of the string s with all leading and trailing Unicode
	// code points contained in cutset removed.
	"Trim": &ugo.Function{
		Name:  "Trim",
		Value: fnASSRS(strings.Trim),
	},
	// ugo:doc
	// TrimLeft(s, cutset string) -> string
	// Returns a slice of the string s with all leading Unicode code points
	// contained in cutset removed.
	"TrimLeft": &ugo.Function{
		Name:  "TrimLeft",
		Value: fnASSRS(strings.TrimLeft),
	},
	// ugo:doc
	// TrimPrefix(s, prefix string) -> string
	// Returns s without the provided leading prefix string. If s doesn't start
	// with prefix, s is returned unchanged.
	"TrimPrefix": &ugo.Function{
		Name:  "TrimPrefix",
		Value: fnASSRS(strings.TrimPrefix),
	},
	// ugo:doc
	// TrimRight(s, cutset string) -> string
	// Returns a slice of the string s with all trailing Unicode code points
	// contained in cutset removed.
	"TrimRight": &ugo.Function{
		Name:  "TrimRight",
		Value: fnASSRS(strings.TrimRight),
	},
	// ugo:doc
	// TrimSpace(s) -> string
	// Returns a slice of the string s, with all leading and trailing white
	// space removed, as defined by Unicode.
	"TrimSpace": &ugo.Function{
		Name:  "TrimSpace",
		Value: fnASRS(strings.TrimSpace),
	},
	// ugo:doc
	// TrimSuffix(s, suffix string) -> string
	// Returns s without the provided trailing suffix string. If s doesn't end
	// with suffix, s is returned unchanged.
	"TrimSuffix": &ugo.Function{
		Name:  "TrimSuffix",
		Value: fnASSRS(strings.TrimSuffix),
	},
}

func fnASSRS(fn func(string, string) string) ugo.CallableFunc {
	return func(args ...ugo.Object) (ugo.Object, error) {
		if len(args) != 2 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				wantEqXGotY(2, len(args)))
		}
		s1, ok := args[0].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("first", "string",
				args[0].TypeName())
		}
		s2, ok := args[1].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("second", "string",
				args[1].TypeName())
		}
		return ugo.String(fn(string(s1), string(s2))), nil
	}
}

func fnASSRB(fn func(string, string) bool) ugo.CallableFunc {
	return func(args ...ugo.Object) (ugo.Object, error) {
		if len(args) != 2 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				wantEqXGotY(2, len(args)))
		}
		s1, ok := args[0].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("first", "string",
				args[0].TypeName())
		}
		s2, ok := args[1].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("second", "string",
				args[1].TypeName())
		}
		return ugo.Bool(fn(string(s1), string(s2))), nil
	}
}

func fnASSRI(fn func(string, string) int) ugo.CallableFunc {
	return func(args ...ugo.Object) (ugo.Object, error) {
		if len(args) != 2 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				wantEqXGotY(2, len(args)))
		}
		s1, ok := args[0].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("first", "string",
				args[0].TypeName())
		}
		s2, ok := args[1].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("second", "string",
				args[1].TypeName())
		}
		return ugo.Int(fn(string(s1), string(s2))), nil
	}
}

func fnASSIRA(fn func(string, string, int) []string) ugo.CallableFunc {
	return func(args ...ugo.Object) (ugo.Object, error) {
		if len(args) != 2 && len(args) != 3 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				"want=2..3 got=" + strconv.Itoa(len(args)))
		}
		s, ok := args[0].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("first", "string",
				args[0].TypeName())
		}
		sep, ok := args[1].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("second", "string",
				args[1].TypeName())
		}
		n := -1
		if len(args) == 3 {
			v, ok := args[2].(ugo.Int)
			if !ok {
				return nil, ugo.NewArgumentTypeError("third", "int",
					args[2].TypeName())
			}
			n = int(v)
		}
		strs := fn(string(s), string(sep), n)
		out := make(ugo.Array, len(strs))
		for i := range strs {
			out[i] = ugo.String(strs[i])
		}
		return out, nil
	}
}

func fnASRS(fn func(string) string) ugo.CallableFunc {
	return func(args ...ugo.Object) (ugo.Object, error) {
		if len(args) != 1 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				wantEqXGotY(1, len(args)))
		}
		s, ok := args[0].(ugo.String)
		if !ok {
			return nil, ugo.NewArgumentTypeError("first", "string",
				args[0].TypeName())
		}
		return ugo.String(fn(string(s))), nil
	}
}

func containsChar(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	c, ok := args[1].(ugo.Char)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "char",
			args[1].TypeName())
	}
	return ugo.Bool(strings.ContainsRune(string(s), rune(c))), nil
}

func fields(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 1 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	strs := strings.Fields(string(s))
	out := make(ugo.Array, len(strs))
	for i := range strs {
		out[i] = ugo.String(strs[i])
	}
	return out, nil
}

func indexByte(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	var b byte
	switch v := args[1].(type) {
	case ugo.Char:
		b = byte(v)
	case ugo.Int:
		b = byte(v)
	default:
		return nil, ugo.NewArgumentTypeError("second", "char|int",
			args[1].TypeName())
	}
	return ugo.Int(strings.IndexByte(string(s), b)), nil
}

func indexChar(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	c, ok := args[1].(ugo.Char)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "char",
			args[1].TypeName())
	}
	return ugo.Int(strings.IndexRune(string(s), rune(c))), nil
}

func join(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	arr, ok := args[0].(ugo.Array)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "array",
			args[0].TypeName())
	}
	sep, ok := args[1].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "string",
			args[1].TypeName())
	}
	elems := make([]string, len(arr))
	for i := range arr {
		elems[i] = arr[i].String()
	}
	return ugo.String(strings.Join(elems, string(sep))), nil
}

func lastIndexByte(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	var b byte
	switch v := args[1].(type) {
	case ugo.Char:
		b = byte(v)
	case ugo.Int:
		b = byte(v)
	default:
		return nil, ugo.NewArgumentTypeError("second", "char|int",
			args[1].TypeName())
	}
	return ugo.Int(strings.LastIndexByte(string(s), b)), nil
}

func padLeft(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, ugo.ErrWrongNumArguments.NewError(
			"want=2..3 got=" + strconv.Itoa(len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	padLen, ok := args[1].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
	diff := int(padLen) - len(s)
	if diff <= 0 {
		return s, nil
	}
	padWith := " "
	if len(args) > 2 {
		if padWith = args[2].String(); len(padWith) == 0 {
			return s, nil
		}
	}
	r := (diff-len(padWith))/len(padWith) + 2
	if r <= 0 {
		return s, nil
	}
	var sb strings.Builder
	sb.Grow(int(padLen))
	sb.WriteString(strings.Repeat(padWith, r)[:diff])
	sb.WriteString(string(s))
	return ugo.String(sb.String()), nil
}

func padRight(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, ugo.ErrWrongNumArguments.NewError(
			"want=2..3 got=" + strconv.Itoa(len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	padLen, ok := args[1].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
	diff := int(padLen) - len(s)
	if diff <= 0 {
		return s, nil
	}
	padWith := " "
	if len(args) > 2 {
		if padWith = args[2].String(); len(padWith) == 0 {
			return s, nil
		}
	}
	r := (diff-len(padWith))/len(padWith) + 2
	if r <= 0 {
		return s, nil
	}
	var sb strings.Builder
	sb.Grow(int(padLen))
	sb.WriteString(string(s))
	sb.WriteString(strings.Repeat(padWith, r)[:diff])
	return ugo.String(sb.String()), nil
}

func repeat(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	n, ok := args[1].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
	// if n is negative strings.Repeat function panics
	if n < 0 {
		return ugo.String(""), nil
	}
	return ugo.String(strings.Repeat(string(s), int(n))), nil
}

func replace(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 3 && len(args) != 4 {
		return nil, ugo.ErrWrongNumArguments.NewError(
			"want=3..4 got=" + strconv.Itoa(len(args)))
	}
	s, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	old, ok := args[1].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "string",
			args[1].TypeName())
	}
	new, ok := args[2].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("third", "string",
			args[2].TypeName())
	}
	n := -1
	if len(args) == 4 {
		v, ok := args[3].(ugo.Int)
		if !ok {
			return nil, ugo.NewArgumentTypeError("fourth", "int",
				args[3].TypeName())
		}
		n = int(v)
	}
	return ugo.String(
		strings.Replace(string(s), string(old), string(new), n),
	), nil
}

func wantEqXGotY(x, y int) string {
	buf := make([]byte, 0, 20)
	buf = append(buf, "want="...)
	buf = strconv.AppendInt(buf, int64(x), 10)
	buf = append(buf, " got="...)
	buf = strconv.AppendInt(buf, int64(y), 10)
	return string(buf)
}
