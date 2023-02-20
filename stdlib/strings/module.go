// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Package strings provides strings module implementing simple functions to
// manipulate UTF-8 encoded strings for uGO script language. It wraps Go's
// strings package functionalities.
package strings

import (
	"strconv"
	"strings"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/stdlib"
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
		Name: "Contains",
		Value: stdlib.FuncPssRO(func(s, substr string) ugo.Object {
			return ugo.Bool(strings.Contains(s, substr))
		}),
	},
	// ugo:doc
	// ContainsAny(s string, chars string) -> bool
	// Reports whether any char in chars are within s.
	"ContainsAny": &ugo.Function{
		Name: "ContainsAny",
		Value: stdlib.FuncPssRO(func(s, chars string) ugo.Object {
			return ugo.Bool(strings.ContainsAny(s, chars))
		}),
	},
	// ugo:doc
	// ContainsChar(s string, c char) -> bool
	// Reports whether the char c is within s.
	"ContainsChar": &ugo.Function{
		Name: "ContainsChar",
		Value: stdlib.FuncPsrRO(func(s string, c rune) ugo.Object {
			return ugo.Bool(strings.ContainsRune(s, c))
		}),
	},
	// ugo:doc
	// Count(s string, substr string) -> int
	// Counts the number of non-overlapping instances of substr in s.
	"Count": &ugo.Function{
		Name: "Count",
		Value: stdlib.FuncPssRO(func(s, substr string) ugo.Object {
			return ugo.Int(strings.Count(s, substr))
		}),
	},
	// ugo:doc
	// EqualFold(s string, t string) -> bool
	// EqualFold reports whether s and t, interpreted as UTF-8 strings,
	// are equal under Unicode case-folding, which is a more general form of
	// case-insensitivity.
	"EqualFold": &ugo.Function{
		Name: "EqualFold",
		Value: stdlib.FuncPssRO(func(s, t string) ugo.Object {
			return ugo.Bool(strings.EqualFold(s, t))
		}),
	},
	// ugo:doc
	// Fields(s string) -> array
	// Splits the string s around each instance of one or more consecutive white
	// space characters, returning an array of substrings of s or an empty array
	// if s contains only white space.
	"Fields": &ugo.Function{
		Name:  "Fields",
		Value: stdlib.FuncPsRO(fields),
	},
	// ugo:doc
	// HasPrefix(s string, prefix string) -> bool
	// Reports whether the string s begins with prefix.
	"HasPrefix": &ugo.Function{
		Name: "HasPrefix",
		Value: stdlib.FuncPssRO(func(s, prefix string) ugo.Object {
			return ugo.Bool(strings.HasPrefix(s, prefix))
		}),
	},
	// ugo:doc
	// HasSuffix(s string, suffix string) -> bool
	// Reports whether the string s ends with prefix.
	"HasSuffix": &ugo.Function{
		Name: "HasSuffix",
		Value: stdlib.FuncPssRO(func(s, suffix string) ugo.Object {
			return ugo.Bool(strings.HasSuffix(s, suffix))
		}),
	},
	// ugo:doc
	// Index(s string, substr string) -> int
	// Returns the index of the first instance of substr in s, or -1 if substr
	// is not present in s.
	"Index": &ugo.Function{
		Name: "Index",
		Value: stdlib.FuncPssRO(func(s, substr string) ugo.Object {
			return ugo.Int(strings.Index(s, substr))
		}),
	},
	// ugo:doc
	// IndexAny(s string, chars string) -> int
	// Returns the index of the first instance of any char from chars in s, or
	// -1 if no char from chars is present in s.
	"IndexAny": &ugo.Function{
		Name: "IndexAny",
		Value: stdlib.FuncPssRO(func(s, chars string) ugo.Object {
			return ugo.Int(strings.IndexAny(s, chars))
		}),
	},
	// ugo:doc
	// IndexByte(s string, c char|int) -> int
	// Returns the index of the first byte value of c in s, or -1 if byte value
	// of c is not present in s. c's integer value must be between 0 and 255.
	"IndexByte": &ugo.Function{
		Name: "IndexByte",
		Value: stdlib.FuncPsrRO(func(s string, c rune) ugo.Object {
			if c > 255 || c < 0 {
				return ugo.Int(-1)
			}
			return ugo.Int(strings.IndexByte(s, byte(c)))
		}),
	},
	// ugo:doc
	// IndexChar(s string, c char) -> int
	// Returns the index of the first instance of the char c, or -1 if char is
	// not present in s.
	"IndexChar": &ugo.Function{
		Name: "IndexChar",
		Value: stdlib.FuncPsrRO(func(s string, c rune) ugo.Object {
			return ugo.Int(strings.IndexRune(s, c))
		}),
	},
	// ugo:doc
	// Join(arr array, sep string) -> string
	// Concatenates the string values of array arr elements to create a
	// single string. The separator string sep is placed between elements in the
	// resulting string.
	"Join": &ugo.Function{
		Name:  "Join",
		Value: stdlib.FuncPAsRO(join),
	},
	// ugo:doc
	// LastIndex(s string, substr string) -> int
	// Returns the index of the last instance of substr in s, or -1 if substr
	// is not present in s.
	"LastIndex": &ugo.Function{
		Name: "LastIndex",
		Value: stdlib.FuncPssRO(func(s, substr string) ugo.Object {
			return ugo.Int(strings.LastIndex(s, substr))
		}),
	},
	// ugo:doc
	// LastIndexAny(s string, chars string) -> int
	// Returns the index of the last instance of any char from chars in s, or
	// -1 if no char from chars is present in s.
	"LastIndexAny": &ugo.Function{
		Name: "LastIndexAny",
		Value: stdlib.FuncPssRO(func(s, chars string) ugo.Object {
			return ugo.Int(strings.LastIndexAny(s, chars))
		}),
	},
	// ugo:doc
	// LastIndexByte(s string, c char|int) -> int
	// Returns the index of byte value of the last instance of c in s, or -1
	// if c is not present in s. c's integer value must be between 0 and 255.
	"LastIndexByte": &ugo.Function{
		Name: "LastIndexByte",
		Value: stdlib.FuncPsrRO(func(s string, c rune) ugo.Object {
			if c > 255 || c < 0 {
				return ugo.Int(-1)
			}
			return ugo.Int(strings.LastIndexByte(s, byte(c)))
		}),
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
		Value: stdlib.FuncPsiRO(repeat),
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
		Value: fnSplit(strings.SplitN),
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
		Value: fnSplit(strings.SplitAfterN),
	},
	// ugo:doc
	// Title(s string) -> string
	// Deprecated: Returns a copy of the string s with all Unicode letters that
	// begin words mapped to their Unicode title case.
	"Title": &ugo.Function{
		Name: "Title",
		Value: stdlib.FuncPsRO(func(s string) ugo.Object {
			//lint:ignore SA1019 Keep it for backward compatibility.
			return ugo.String(strings.Title(s))
		}),
	},
	// ugo:doc
	// ToLower(s string) -> string
	// Returns s with all Unicode letters mapped to their lower case.
	"ToLower": &ugo.Function{
		Name: "ToLower",
		Value: stdlib.FuncPsRO(func(s string) ugo.Object {
			return ugo.String(strings.ToLower(s))
		}),
	},
	// ugo:doc
	// ToTitle(s string) -> string
	// Returns a copy of the string s with all Unicode letters mapped to their
	// Unicode title case.
	"ToTitle": &ugo.Function{
		Name: "ToTitle",
		Value: stdlib.FuncPsRO(func(s string) ugo.Object {
			return ugo.String(strings.ToTitle(s))
		}),
	},
	// ugo:doc
	// ToUpper(s string) -> string
	// Returns s with all Unicode letters mapped to their upper case.
	"ToUpper": &ugo.Function{
		Name: "ToUpper",
		Value: stdlib.FuncPsRO(func(s string) ugo.Object {
			return ugo.String(strings.ToUpper(s))
		}),
	},
	// ugo:doc
	// Trim(s string, cutset string) -> string
	// Returns a slice of the string s with all leading and trailing Unicode
	// code points contained in cutset removed.
	"Trim": &ugo.Function{
		Name: "Trim",
		Value: stdlib.FuncPssRO(func(s, cutset string) ugo.Object {
			return ugo.String(strings.Trim(s, cutset))
		}),
	},
	// ugo:doc
	// TrimLeft(s string, cutset string) -> string
	// Returns a slice of the string s with all leading Unicode code points
	// contained in cutset removed.
	"TrimLeft": &ugo.Function{
		Name: "TrimLeft",
		Value: stdlib.FuncPssRO(func(s, cutset string) ugo.Object {
			return ugo.String(strings.TrimLeft(s, cutset))
		}),
	},
	// ugo:doc
	// TrimPrefix(s string, prefix string) -> string
	// Returns s without the provided leading prefix string. If s doesn't start
	// with prefix, s is returned unchanged.
	"TrimPrefix": &ugo.Function{
		Name: "TrimPrefix",
		Value: stdlib.FuncPssRO(func(s, prefix string) ugo.Object {
			return ugo.String(strings.TrimPrefix(s, prefix))
		}),
	},
	// ugo:doc
	// TrimRight(s string, cutset string) -> string
	// Returns a slice of the string s with all trailing Unicode code points
	// contained in cutset removed.
	"TrimRight": &ugo.Function{
		Name: "TrimRight",
		Value: stdlib.FuncPssRO(func(s, cutset string) ugo.Object {
			return ugo.String(strings.TrimRight(s, cutset))
		}),
	},
	// ugo:doc
	// TrimSpace(s string) -> string
	// Returns a slice of the string s, with all leading and trailing white
	// space removed, as defined by Unicode.
	"TrimSpace": &ugo.Function{
		Name: "TrimSpace",
		Value: stdlib.FuncPsRO(func(s string) ugo.Object {
			return ugo.String(strings.TrimSpace(s))
		}),
	},
	// ugo:doc
	// TrimSuffix(s string, suffix string) -> string
	// Returns s without the provided trailing suffix string. If s doesn't end
	// with suffix, s is returned unchanged.
	"TrimSuffix": &ugo.Function{
		Name: "TrimSuffix",
		Value: stdlib.FuncPssRO(func(s, suffix string) ugo.Object {
			return ugo.String(strings.TrimSuffix(s, suffix))
		}),
	},
}

func fnSplit(fn func(string, string, int) []string) ugo.CallableFunc {
	return func(args ...ugo.Object) (ugo.Object, error) {
		if len(args) != 2 && len(args) != 3 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				"want=2..3 got=" + strconv.Itoa(len(args)))
		}
		s, ok := ugo.ToGoString(args[0])
		if !ok {
			return nil, ugo.NewArgumentTypeError("1st", "string",
				args[0].TypeName())
		}
		sep, ok := ugo.ToGoString(args[1])
		if !ok {
			return nil, ugo.NewArgumentTypeError("2nd", "string",
				args[1].TypeName())
		}
		n := -1
		if len(args) == 3 {
			v, ok := ugo.ToGoInt(args[2])
			if !ok {
				return nil, ugo.NewArgumentTypeError("3rd", "int",
					args[2].TypeName())
			}
			n = v
		}
		strs := fn(s, sep, n)
		out := make(ugo.Array, len(strs))
		for i := range strs {
			out[i] = ugo.String(strs[i])
		}
		return out, nil
	}
}

func fields(s string) ugo.Object {
	strs := strings.Fields(s)
	out := make(ugo.Array, len(strs))
	for i := range strs {
		out[i] = ugo.String(strs[i])
	}
	return out
}

func join(arr ugo.Array, sep string) ugo.Object {
	elems := make([]string, len(arr))
	for i := range arr {
		elems[i] = arr[i].String()
	}
	return ugo.String(strings.Join(elems, sep))
}

func padLeft(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, ugo.ErrWrongNumArguments.NewError(
			"want=2..3 got=" + strconv.Itoa(len(args)))
	}
	s, ok := ugo.ToGoString(args[0])
	if !ok {
		return nil, ugo.NewArgumentTypeError("1st", "string",
			args[0].TypeName())
	}
	padLen, ok := ugo.ToGoInt(args[1])
	if !ok {
		return nil, ugo.NewArgumentTypeError("2nd", "int",
			args[1].TypeName())
	}
	diff := padLen - len(s)
	if diff <= 0 {
		return ugo.String(s), nil
	}
	padWith := " "
	if len(args) > 2 {
		if padWith = args[2].String(); len(padWith) == 0 {
			return ugo.String(s), nil
		}
	}
	r := (diff-len(padWith))/len(padWith) + 2
	if r <= 0 {
		return ugo.String(s), nil
	}
	var sb strings.Builder
	sb.Grow(padLen)
	sb.WriteString(strings.Repeat(padWith, r)[:diff])
	sb.WriteString(s)
	return ugo.String(sb.String()), nil
}

func padRight(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, ugo.ErrWrongNumArguments.NewError(
			"want=2..3 got=" + strconv.Itoa(len(args)))
	}
	s, ok := ugo.ToGoString(args[0])
	if !ok {
		return nil, ugo.NewArgumentTypeError("1st", "string",
			args[0].TypeName())
	}
	padLen, ok := ugo.ToGoInt(args[1])
	if !ok {
		return nil, ugo.NewArgumentTypeError("2nd", "int",
			args[1].TypeName())
	}
	diff := padLen - len(s)
	if diff <= 0 {
		return ugo.String(s), nil
	}
	padWith := " "
	if len(args) > 2 {
		if padWith = args[2].String(); len(padWith) == 0 {
			return ugo.String(s), nil
		}
	}
	r := (diff-len(padWith))/len(padWith) + 2
	if r <= 0 {
		return ugo.String(s), nil
	}
	var sb strings.Builder
	sb.Grow(padLen)
	sb.WriteString(s)
	sb.WriteString(strings.Repeat(padWith, r)[:diff])
	return ugo.String(sb.String()), nil
}

func repeat(s string, count int) ugo.Object {
	// if n is negative strings.Repeat function panics
	if count < 0 {
		return ugo.String("")
	}
	return ugo.String(strings.Repeat(s, count))
}

func replace(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 3 && len(args) != 4 {
		return nil, ugo.ErrWrongNumArguments.NewError(
			"want=3..4 got=" + strconv.Itoa(len(args)))
	}
	s, ok := ugo.ToGoString(args[0])
	if !ok {
		return nil, ugo.NewArgumentTypeError("1st", "string",
			args[0].TypeName())
	}
	old, ok := ugo.ToGoString(args[1])
	if !ok {
		return nil, ugo.NewArgumentTypeError("2nd", "string",
			args[1].TypeName())
	}
	news, ok := ugo.ToGoString(args[2])
	if !ok {
		return nil, ugo.NewArgumentTypeError("3rd", "string",
			args[2].TypeName())
	}
	n := -1
	if len(args) == 4 {
		v, ok := ugo.ToGoInt(args[3])
		if !ok {
			return nil, ugo.NewArgumentTypeError(
				"4th", "int", args[3].TypeName())
		}
		n = v
	}
	return ugo.String(strings.Replace(s, old, news, n)), nil
}
