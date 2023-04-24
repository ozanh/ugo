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
	"unicode/utf8"

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
		Name:    "Contains",
		Value:   stdlib.FuncPssRO(containsFunc),
		ValueEx: stdlib.FuncPssROEx(containsFunc),
	},
	// ugo:doc
	// ContainsAny(s string, chars string) -> bool
	// Reports whether any char in chars are within s.
	"ContainsAny": &ugo.Function{
		Name:    "ContainsAny",
		Value:   stdlib.FuncPssRO(containsAnyFunc),
		ValueEx: stdlib.FuncPssROEx(containsAnyFunc),
	},
	// ugo:doc
	// ContainsChar(s string, c char) -> bool
	// Reports whether the char c is within s.
	"ContainsChar": &ugo.Function{
		Name:    "ContainsChar",
		Value:   stdlib.FuncPsrRO(containsCharFunc),
		ValueEx: stdlib.FuncPsrROEx(containsCharFunc),
	},
	// ugo:doc
	// Count(s string, substr string) -> int
	// Counts the number of non-overlapping instances of substr in s.
	"Count": &ugo.Function{
		Name:    "Count",
		Value:   stdlib.FuncPssRO(countFunc),
		ValueEx: stdlib.FuncPssROEx(countFunc),
	},
	// ugo:doc
	// EqualFold(s string, t string) -> bool
	// EqualFold reports whether s and t, interpreted as UTF-8 strings,
	// are equal under Unicode case-folding, which is a more general form of
	// case-insensitivity.
	"EqualFold": &ugo.Function{
		Name:    "EqualFold",
		Value:   stdlib.FuncPssRO(equalFoldFunc),
		ValueEx: stdlib.FuncPssROEx(equalFoldFunc),
	},
	// ugo:doc
	// Fields(s string) -> array
	// Splits the string s around each instance of one or more consecutive white
	// space characters, returning an array of substrings of s or an empty array
	// if s contains only white space.
	"Fields": &ugo.Function{
		Name:    "Fields",
		Value:   stdlib.FuncPsRO(fieldsFunc),
		ValueEx: stdlib.FuncPsROEx(fieldsFunc),
	},
	// ugo:doc
	// FieldsFunc(s string, f func(char) bool) -> array
	// Splits the string s at each run of Unicode code points c satisfying f(c),
	// and returns an array of slices of s. If all code points in s satisfy
	// f(c) or the string is empty, an empty array is returned.
	"FieldsFunc": &ugo.Function{
		Name: "FieldsFunc",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return fieldsFuncInv(ugo.NewCall(nil, args))
		},
		ValueEx: fieldsFuncInv,
	},
	// ugo:doc
	// HasPrefix(s string, prefix string) -> bool
	// Reports whether the string s begins with prefix.
	"HasPrefix": &ugo.Function{
		Name:    "HasPrefix",
		Value:   stdlib.FuncPssRO(hasPrefixFunc),
		ValueEx: stdlib.FuncPssROEx(hasPrefixFunc),
	},
	// ugo:doc
	// HasSuffix(s string, suffix string) -> bool
	// Reports whether the string s ends with prefix.
	"HasSuffix": &ugo.Function{
		Name:    "HasSuffix",
		Value:   stdlib.FuncPssRO(hasSuffixFunc),
		ValueEx: stdlib.FuncPssROEx(hasSuffixFunc),
	},
	// ugo:doc
	// Index(s string, substr string) -> int
	// Returns the index of the first instance of substr in s, or -1 if substr
	// is not present in s.
	"Index": &ugo.Function{
		Name:    "Index",
		Value:   stdlib.FuncPssRO(indexFunc),
		ValueEx: stdlib.FuncPssROEx(indexFunc),
	},
	// ugo:doc
	// IndexAny(s string, chars string) -> int
	// Returns the index of the first instance of any char from chars in s, or
	// -1 if no char from chars is present in s.
	"IndexAny": &ugo.Function{
		Name:    "IndexAny",
		Value:   stdlib.FuncPssRO(indexAnyFunc),
		ValueEx: stdlib.FuncPssROEx(indexAnyFunc),
	},
	// ugo:doc
	// IndexByte(s string, c char|int) -> int
	// Returns the index of the first byte value of c in s, or -1 if byte value
	// of c is not present in s. c's integer value must be between 0 and 255.
	"IndexByte": &ugo.Function{
		Name:    "IndexByte",
		Value:   stdlib.FuncPsrRO(indexByteFunc),
		ValueEx: stdlib.FuncPsrROEx(indexByteFunc),
	},
	// ugo:doc
	// IndexChar(s string, c char) -> int
	// Returns the index of the first instance of the char c, or -1 if char is
	// not present in s.
	"IndexChar": &ugo.Function{
		Name:    "IndexChar",
		Value:   stdlib.FuncPsrRO(indexCharFunc),
		ValueEx: stdlib.FuncPsrROEx(indexCharFunc),
	},
	// ugo:doc
	// IndexFunc(s string, f func(char) bool) -> int
	// Returns the index into s of the first Unicode code point satisfying f(c),
	// or -1 if none do.
	"IndexFunc": &ugo.Function{
		Name: "IndexFunc",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newIndexFuncInv(strings.IndexFunc)(ugo.NewCall(nil, args))
		},
		ValueEx: newIndexFuncInv(strings.IndexFunc),
	},
	// ugo:doc
	// Join(arr array, sep string) -> string
	// Concatenates the string values of array arr elements to create a
	// single string. The separator string sep is placed between elements in the
	// resulting string.
	"Join": &ugo.Function{
		Name:    "Join",
		Value:   stdlib.FuncPAsRO(joinFunc),
		ValueEx: stdlib.FuncPAsROEx(joinFunc),
	},
	// ugo:doc
	// LastIndex(s string, substr string) -> int
	// Returns the index of the last instance of substr in s, or -1 if substr
	// is not present in s.
	"LastIndex": &ugo.Function{
		Name:    "LastIndex",
		Value:   stdlib.FuncPssRO(lastIndexFunc),
		ValueEx: stdlib.FuncPssROEx(lastIndexFunc),
	},
	// ugo:doc
	// LastIndexAny(s string, chars string) -> int
	// Returns the index of the last instance of any char from chars in s, or
	// -1 if no char from chars is present in s.
	"LastIndexAny": &ugo.Function{
		Name:    "LastIndexAny",
		Value:   stdlib.FuncPssRO(lastIndexAnyFunc),
		ValueEx: stdlib.FuncPssROEx(lastIndexAnyFunc),
	},
	// ugo:doc
	// LastIndexByte(s string, c char|int) -> int
	// Returns the index of byte value of the last instance of c in s, or -1
	// if c is not present in s. c's integer value must be between 0 and 255.
	"LastIndexByte": &ugo.Function{
		Name:    "LastIndexByte",
		Value:   stdlib.FuncPsrRO(lastIndexByteFunc),
		ValueEx: stdlib.FuncPsrROEx(lastIndexByteFunc),
	},
	// ugo:doc
	// LastIndexFunc(s string, f func(char) bool) -> int
	// Returns the index into s of the last Unicode code point satisfying f(c),
	// or -1 if none do.
	"LastIndexFunc": &ugo.Function{
		Name: "LastIndexFunc",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newIndexFuncInv(strings.LastIndexFunc)(ugo.NewCall(nil, args))
		},
		ValueEx: newIndexFuncInv(strings.LastIndexFunc),
	},
	// ugo:doc
	// Map(f func(char) char, s string) -> string
	// Returns a copy of the string s with all its characters modified
	// according to the mapping function f. If f returns a negative value, the
	// character is dropped from the string with no replacement.
	"Map": &ugo.Function{
		Name: "Map",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return mapFuncInv(ugo.NewCall(nil, args))
		},
		ValueEx: mapFuncInv,
	},
	// ugo:doc
	// PadLeft(s string, padLen int[, padWith any]) -> string
	// Returns a string that is padded on the left with the string `padWith` until
	// the `padLen` length is reached. If padWith is not given, a white space is
	// used as default padding.
	"PadLeft": &ugo.Function{
		Name: "PadLeft",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return pad(ugo.NewCall(nil, args), true)
		},
		ValueEx: func(c ugo.Call) (ugo.Object, error) {
			return pad(c, true)
		},
	},
	// ugo:doc
	// PadRight(s string, padLen int[, padWith any]) -> string
	// Returns a string that is padded on the right with the string `padWith` until
	// the `padLen` length is reached. If padWith is not given, a white space is
	// used as default padding.
	"PadRight": &ugo.Function{
		Name: "PadRight",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return pad(ugo.NewCall(nil, args), false)
		},
		ValueEx: func(c ugo.Call) (ugo.Object, error) {
			return pad(c, false)
		},
	},
	// ugo:doc
	// Repeat(s string, count int) -> string
	// Returns a new string consisting of count copies of the string s.
	//
	// - If count is a negative int, it returns empty string.
	// - If (len(s) * count) overflows, it panics.
	"Repeat": &ugo.Function{
		Name:    "Repeat",
		Value:   stdlib.FuncPsiRO(repeatFunc),
		ValueEx: stdlib.FuncPsiROEx(repeatFunc),
	},
	// ugo:doc
	// Replace(s string, old string, new string[, n int]) -> string
	// Returns a copy of the string s with the first n non-overlapping instances
	// of old replaced by new. If n is not provided or -1, it replaces all
	// instances.
	"Replace": &ugo.Function{
		Name: "Replace",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return replaceFunc(ugo.NewCall(nil, args))
		},
		ValueEx: replaceFunc,
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
		Name: "Split",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newSplitFunc(strings.SplitN)(ugo.NewCall(nil, args))
		},
		ValueEx: newSplitFunc(strings.SplitN),
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
		Name: "SplitAfter",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newSplitFunc(strings.SplitAfterN)(ugo.NewCall(nil, args))
		},
		ValueEx: newSplitFunc(strings.SplitAfterN),
	},
	// ugo:doc
	// Title(s string) -> string
	// Deprecated: Returns a copy of the string s with all Unicode letters that
	// begin words mapped to their Unicode title case.
	"Title": &ugo.Function{
		Name:    "Title",
		Value:   stdlib.FuncPsRO(titleFunc),
		ValueEx: stdlib.FuncPsROEx(titleFunc),
	},
	// ugo:doc
	// ToLower(s string) -> string
	// Returns s with all Unicode letters mapped to their lower case.
	"ToLower": &ugo.Function{
		Name:    "ToLower",
		Value:   stdlib.FuncPsRO(toLowerFunc),
		ValueEx: stdlib.FuncPsROEx(toLowerFunc),
	},
	// ugo:doc
	// ToTitle(s string) -> string
	// Returns a copy of the string s with all Unicode letters mapped to their
	// Unicode title case.
	"ToTitle": &ugo.Function{
		Name:    "ToTitle",
		Value:   stdlib.FuncPsRO(toTitleFunc),
		ValueEx: stdlib.FuncPsROEx(toTitleFunc),
	},
	// ugo:doc
	// ToUpper(s string) -> string
	// Returns s with all Unicode letters mapped to their upper case.
	"ToUpper": &ugo.Function{
		Name:    "ToUpper",
		Value:   stdlib.FuncPsRO(toUpperFunc),
		ValueEx: stdlib.FuncPsROEx(toUpperFunc),
	},
	// ugo:doc
	// ToValidUTF8(s string[, replacement string]) -> string
	// Returns a copy of the string s with each run of invalid UTF-8 byte
	// sequences replaced by the replacement string, which may be empty.
	"ToValidUTF8": &ugo.Function{
		Name: "ToValidUTF8",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return toValidUTF8Func(ugo.NewCall(nil, args))
		},
		ValueEx: toValidUTF8Func,
	},
	// ugo:doc
	// Trim(s string, cutset string) -> string
	// Returns a slice of the string s with all leading and trailing Unicode
	// code points contained in cutset removed.
	"Trim": &ugo.Function{
		Name:    "Trim",
		Value:   stdlib.FuncPssRO(trimFunc),
		ValueEx: stdlib.FuncPssROEx(trimFunc),
	},
	// ugo:doc
	// TrimFunc(s string, f func(char) bool) -> string
	// Returns a slice of the string s with all leading and trailing Unicode
	// code points satisfying f removed.
	"TrimFunc": &ugo.Function{
		Name: "TrimFunc",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newTrimFuncInv(strings.TrimFunc)(ugo.NewCall(nil, args))
		},
		ValueEx: newTrimFuncInv(strings.TrimFunc),
	},
	// ugo:doc
	// TrimLeft(s string, cutset string) -> string
	// Returns a slice of the string s with all leading Unicode code points
	// contained in cutset removed.
	"TrimLeft": &ugo.Function{
		Name:    "TrimLeft",
		Value:   stdlib.FuncPssRO(trimLeftFunc),
		ValueEx: stdlib.FuncPssROEx(trimLeftFunc),
	},
	// ugo:doc
	// TrimLeftFunc(s string, f func(char) bool) -> string
	// Returns a slice of the string s with all leading Unicode code points
	// c satisfying f(c) removed.
	"TrimLeftFunc": &ugo.Function{
		Name: "TrimLeftFunc",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newTrimFuncInv(strings.TrimLeftFunc)(ugo.NewCall(nil, args))
		},
		ValueEx: newTrimFuncInv(strings.TrimLeftFunc),
	},
	// ugo:doc
	// TrimPrefix(s string, prefix string) -> string
	// Returns s without the provided leading prefix string. If s doesn't start
	// with prefix, s is returned unchanged.
	"TrimPrefix": &ugo.Function{
		Name:    "TrimPrefix",
		Value:   stdlib.FuncPssRO(trimPrefixFunc),
		ValueEx: stdlib.FuncPssROEx(trimPrefixFunc),
	},
	// ugo:doc
	// TrimRight(s string, cutset string) -> string
	// Returns a slice of the string s with all trailing Unicode code points
	// contained in cutset removed.
	"TrimRight": &ugo.Function{
		Name:    "TrimRight",
		Value:   stdlib.FuncPssRO(trimRightFunc),
		ValueEx: stdlib.FuncPssROEx(trimRightFunc),
	},
	// ugo:doc
	// TrimRightFunc(s string, f func(char) bool) -> string
	// Returns a slice of the string s with all trailing Unicode code points
	// c satisfying f(c) removed.
	"TrimRightFunc": &ugo.Function{
		Name: "TrimRightFunc",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newTrimFuncInv(strings.TrimRightFunc)(ugo.NewCall(nil, args))
		},
		ValueEx: newTrimFuncInv(strings.TrimRightFunc),
	},
	// ugo:doc
	// TrimSpace(s string) -> string
	// Returns a slice of the string s, with all leading and trailing white
	// space removed, as defined by Unicode.
	"TrimSpace": &ugo.Function{
		Name:    "TrimSpace",
		Value:   stdlib.FuncPsRO(trimSpaceFunc),
		ValueEx: stdlib.FuncPsROEx(trimSpaceFunc),
	},
	// ugo:doc
	// TrimSuffix(s string, suffix string) -> string
	// Returns s without the provided trailing suffix string. If s doesn't end
	// with suffix, s is returned unchanged.
	"TrimSuffix": &ugo.Function{
		Name:    "TrimSuffix",
		Value:   stdlib.FuncPssRO(trimSuffixFunc),
		ValueEx: stdlib.FuncPssROEx(trimSuffixFunc),
	},
}

func containsFunc(s, substr string) ugo.Object {
	return ugo.Bool(strings.Contains(s, substr))
}

func containsAnyFunc(s, chars string) ugo.Object {
	return ugo.Bool(strings.ContainsAny(s, chars))
}

func containsCharFunc(s string, c rune) ugo.Object {
	return ugo.Bool(strings.ContainsRune(s, c))
}

func countFunc(s, substr string) ugo.Object {
	return ugo.Int(strings.Count(s, substr))
}

func equalFoldFunc(s, t string) ugo.Object {
	return ugo.Bool(strings.EqualFold(s, t))
}

func fieldsFunc(s string) ugo.Object {
	fields := strings.Fields(s)
	out := make(ugo.Array, 0, len(fields))
	for _, s := range fields {
		out = append(out, ugo.String(s))
	}
	return out
}

func fieldsFuncInv(c ugo.Call) (ugo.Object, error) {
	return stringInvoke(c, 0, 1,
		func(s string, inv *ugo.Invoker) (ugo.Object, error) {
			var err error
			fields := strings.FieldsFunc(s, func(r rune) bool {
				if err != nil {
					return false
				}
				var ret ugo.Object
				ret, err = inv.Invoke(ugo.Char(r))
				if err != nil {
					return false
				}
				return !ret.IsFalsy()
			})
			if err != nil {
				return ugo.Undefined, err
			}
			out := make(ugo.Array, 0, len(fields))
			for _, s := range fields {
				out = append(out, ugo.String(s))
			}
			return out, nil
		},
	)
}

func hasPrefixFunc(s, prefix string) ugo.Object {
	return ugo.Bool(strings.HasPrefix(s, prefix))
}

func hasSuffixFunc(s, suffix string) ugo.Object {
	return ugo.Bool(strings.HasSuffix(s, suffix))
}

func indexFunc(s, substr string) ugo.Object {
	return ugo.Int(strings.Index(s, substr))
}

func indexAnyFunc(s, chars string) ugo.Object {
	return ugo.Int(strings.IndexAny(s, chars))
}

func indexByteFunc(s string, c rune) ugo.Object {
	if c > 255 || c < 0 {
		return ugo.Int(-1)
	}
	return ugo.Int(strings.IndexByte(s, byte(c)))
}

func indexCharFunc(s string, c rune) ugo.Object {
	return ugo.Int(strings.IndexRune(s, c))
}

func joinFunc(arr ugo.Array, sep string) ugo.Object {
	elems := make([]string, len(arr))
	for i := range arr {
		elems[i] = arr[i].String()
	}
	return ugo.String(strings.Join(elems, sep))
}

func lastIndexFunc(s, substr string) ugo.Object {
	return ugo.Int(strings.LastIndex(s, substr))
}

func lastIndexAnyFunc(s, chars string) ugo.Object {
	return ugo.Int(strings.LastIndexAny(s, chars))
}

func lastIndexByteFunc(s string, c rune) ugo.Object {
	if c > 255 || c < 0 {
		return ugo.Int(-1)
	}
	return ugo.Int(strings.LastIndexByte(s, byte(c)))
}

func mapFuncInv(c ugo.Call) (ugo.Object, error) {
	return stringInvoke(c, 1, 0,
		func(s string, inv *ugo.Invoker) (ugo.Object, error) {
			var err error
			out := strings.Map(func(r rune) rune {
				if err != nil {
					return utf8.RuneError
				}
				var ret ugo.Object
				ret, err = inv.Invoke(ugo.Char(r))
				if err != nil {
					return 0
				}
				r, ok := ugo.ToGoRune(ret)
				if !ok {
					return utf8.RuneError
				}
				return r
			}, s)
			return ugo.String(out), err
		},
	)
}

func pad(c ugo.Call, left bool) (ugo.Object, error) {
	size := c.Len()
	if size != 2 && size != 3 {
		return ugo.Undefined,
			ugo.ErrWrongNumArguments.NewError("want=2..3 got=" + strconv.Itoa(size))
	}
	s := c.Get(0).String()
	padLen, ok := ugo.ToGoInt(c.Get(1))
	if !ok {
		return ugo.Undefined,
			ugo.NewArgumentTypeError("2nd", "int", c.Get(1).TypeName())
	}
	diff := padLen - len(s)
	if diff <= 0 {
		return ugo.String(s), nil
	}
	padWith := " "
	if size > 2 {
		if padWith = c.Get(2).String(); len(padWith) == 0 {
			return ugo.String(s), nil
		}
	}
	r := (diff-len(padWith))/len(padWith) + 2
	if r <= 0 {
		return ugo.String(s), nil
	}
	var sb strings.Builder
	sb.Grow(padLen)
	if left {
		sb.WriteString(strings.Repeat(padWith, r)[:diff])
		sb.WriteString(s)
	} else {
		sb.WriteString(s)
		sb.WriteString(strings.Repeat(padWith, r)[:diff])
	}
	return ugo.String(sb.String()), nil
}

func repeatFunc(s string, count int) ugo.Object {
	// if n is negative strings.Repeat function panics
	if count < 0 {
		return ugo.String("")
	}
	return ugo.String(strings.Repeat(s, count))
}

func replaceFunc(c ugo.Call) (ugo.Object, error) {
	size := c.Len()
	if size != 3 && size != 4 {
		return ugo.Undefined,
			ugo.ErrWrongNumArguments.NewError("want=3..4 got=" + strconv.Itoa(size))
	}
	s := c.Get(0).String()
	old := c.Get(1).String()
	news := c.Get(2).String()
	n := -1
	if size == 4 {
		v, ok := ugo.ToGoInt(c.Get(3))
		if !ok {
			return ugo.Undefined,
				ugo.NewArgumentTypeError("4th", "int", c.Get(3).TypeName())
		}
		n = v
	}
	return ugo.String(strings.Replace(s, old, news, n)), nil
}

func titleFunc(s string) ugo.Object {
	//lint:ignore SA1019 Keep it for backward compatibility.
	return ugo.String(strings.Title(s)) //nolint staticcheck Keep it for backward compatibility
}

func toLowerFunc(s string) ugo.Object { return ugo.String(strings.ToLower(s)) }

func toTitleFunc(s string) ugo.Object { return ugo.String(strings.ToTitle(s)) }

func toUpperFunc(s string) ugo.Object { return ugo.String(strings.ToUpper(s)) }

func toValidUTF8Func(c ugo.Call) (ugo.Object, error) {
	size := c.Len()
	if size != 1 && size != 2 {
		return ugo.Undefined,
			ugo.ErrWrongNumArguments.NewError("want=1..2 got=" + strconv.Itoa(size))
	}
	s := c.Get(0).String()
	var repl string
	if size == 2 {
		repl = c.Get(1).String()
	}
	return ugo.String(strings.ToValidUTF8(s, repl)), nil
}

func trimFunc(s, cutset string) ugo.Object {
	return ugo.String(strings.Trim(s, cutset))
}

func trimLeftFunc(s, cutset string) ugo.Object {
	return ugo.String(strings.TrimLeft(s, cutset))
}

func trimPrefixFunc(s, prefix string) ugo.Object {
	return ugo.String(strings.TrimPrefix(s, prefix))
}

func trimRightFunc(s, cutset string) ugo.Object {
	return ugo.String(strings.TrimRight(s, cutset))
}

func trimSpaceFunc(s string) ugo.Object {
	return ugo.String(strings.TrimSpace(s))
}

func trimSuffixFunc(s, suffix string) ugo.Object {
	return ugo.String(strings.TrimSuffix(s, suffix))
}

func newSplitFunc(fn func(string, string, int) []string) ugo.CallableExFunc {
	return func(c ugo.Call) (ugo.Object, error) {
		size := c.Len()
		if size != 2 && size != 3 {
			return ugo.Undefined,
				ugo.ErrWrongNumArguments.NewError("want=2..3 got=" + strconv.Itoa(size))
		}
		s := c.Get(0).String()
		sep := c.Get(1).String()
		n := -1
		if size == 3 {
			v, ok := ugo.ToGoInt(c.Get(2))
			if !ok {
				return ugo.Undefined,
					ugo.NewArgumentTypeError("3rd", "int", c.Get(2).TypeName())
			}
			n = v
		}
		strs := fn(s, sep, n)
		out := make(ugo.Array, 0, len(strs))
		for _, s := range strs {
			out = append(out, ugo.String(s))
		}
		return out, nil
	}
}

func newIndexFuncInv(fn func(string, func(rune) bool) int) ugo.CallableExFunc {
	return func(c ugo.Call) (ugo.Object, error) {
		return stringInvoke(c, 0, 1,
			func(s string, inv *ugo.Invoker) (ugo.Object, error) {
				var err error
				out := fn(s, func(r rune) bool {
					if err != nil {
						return false
					}
					var ret ugo.Object
					ret, err = inv.Invoke(ugo.Char(r))
					if err != nil {
						return false
					}
					return !ret.IsFalsy()
				})
				return ugo.Int(out), err
			},
		)
	}
}

func newTrimFuncInv(fn func(string, func(rune) bool) string) ugo.CallableExFunc {
	return func(c ugo.Call) (ugo.Object, error) {
		return stringInvoke(c, 0, 1,
			func(s string, inv *ugo.Invoker) (ugo.Object, error) {
				var err error
				out := fn(s, func(r rune) bool {
					if err != nil {
						return false
					}
					var ret ugo.Object
					ret, err = inv.Invoke(ugo.Char(r))
					if err != nil {
						return false
					}
					return !ret.IsFalsy()
				})
				return ugo.String(out), err
			},
		)
	}
}

func stringInvoke(
	c ugo.Call,
	sidx int,
	cidx int,
	fn func(string, *ugo.Invoker) (ugo.Object, error),
) (ugo.Object, error) {
	err := c.CheckLen(2)
	if err != nil {
		return ugo.Undefined, err
	}

	str := c.Get(sidx).String()
	callee := c.Get(cidx)
	if !callee.CanCall() {
		return ugo.Undefined, ugo.ErrNotCallable
	}
	if c.VM() == nil {
		if _, ok := callee.(*ugo.CompiledFunction); ok {
			return ugo.Undefined, ugo.ErrNotCallable
		}
	}

	inv := ugo.NewInvoker(c.VM(), callee)
	inv.Acquire()
	defer inv.Release()
	return fn(str, inv)
}
