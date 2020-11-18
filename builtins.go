// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ozanh/ugo/token"
)

var (
	// PrintWriter is the default writer for printf and println builtins.
	PrintWriter io.Writer = os.Stdout
)

// BuiltinType represents a builtin type
type BuiltinType byte

// Builtins
const (
	BuiltinAppend BuiltinType = iota
	BuiltinDelete
	BuiltinCopy
	BuiltinRepeat
	BuiltinContains
	BuiltinLen
	BuiltinSort
	BuiltinSortReverse
	BuiltinError
	BuiltinTypeName
	BuiltinBool
	BuiltinInt
	BuiltinUint
	BuiltinFloat
	BuiltinChar
	BuiltinString
	BuiltinBytes
	BuiltinChars
	BuiltinPrintf
	BuiltinPrintln
	BuiltinSprintf
	BuiltinGlobals

	BuiltinIsError
	BuiltinIsInt
	BuiltinIsUint
	BuiltinIsFloat
	BuiltinIsChar
	BuiltinIsBool
	BuiltinIsString
	BuiltinIsBytes
	BuiltinIsMap
	BuiltinIsSyncMap
	BuiltinIsArray
	BuiltinIsUndefined
	BuiltinIsFunction
	BuiltinIsCallable
	BuiltinIsIterable

	BuiltinWrongNumArgumentsError
	BuiltinInvalidOperatorError
	BuiltinIndexOutOfBoundsError
	BuiltinNotIterableError
	BuiltinNotIndexableError
	BuiltinNotIndexAssignableError
	BuiltinNotCallableError
	BuiltinNotImplementedError
	BuiltinZeroDivisionError
	BuiltinTypeError

	pBuiltinArrayDestruct
)

// BuiltinsMap is list of builtin types, exported for REPL.
var BuiltinsMap = map[string]BuiltinType{
	"append":      BuiltinAppend,
	"delete":      BuiltinDelete,
	"copy":        BuiltinCopy,
	"repeat":      BuiltinRepeat,
	"contains":    BuiltinContains,
	"len":         BuiltinLen,
	"sort":        BuiltinSort,
	"sortReverse": BuiltinSortReverse,
	"error":       BuiltinError,
	"typeName":    BuiltinTypeName,
	"bool":        BuiltinBool,
	"int":         BuiltinInt,
	"uint":        BuiltinUint,
	"float":       BuiltinFloat,
	"char":        BuiltinChar,
	"string":      BuiltinString,
	"bytes":       BuiltinBytes,
	"chars":       BuiltinChars,
	"printf":      BuiltinPrintf,
	"println":     BuiltinPrintln,
	"sprintf":     BuiltinSprintf,
	"globals":     BuiltinGlobals,

	"isError":     BuiltinIsError,
	"isInt":       BuiltinIsInt,
	"isUint":      BuiltinIsUint,
	"isFloat":     BuiltinIsFloat,
	"isChar":      BuiltinIsChar,
	"isBool":      BuiltinIsBool,
	"isString":    BuiltinIsString,
	"isBytes":     BuiltinIsBytes,
	"isMap":       BuiltinIsMap,
	"isSyncMap":   BuiltinIsSyncMap,
	"isArray":     BuiltinIsArray,
	"isUndefined": BuiltinIsUndefined,
	"isFunction":  BuiltinIsFunction,
	"isCallable":  BuiltinIsCallable,
	"isIterable":  BuiltinIsIterable,

	"WrongNumArgumentsError":  BuiltinWrongNumArgumentsError,
	"InvalidOperatorError":    BuiltinInvalidOperatorError,
	"IndexOutOfBoundsError":   BuiltinIndexOutOfBoundsError,
	"NotIterableError":        BuiltinNotIterableError,
	"NotIndexableError":       BuiltinNotIndexableError,
	"NotIndexAssignableError": BuiltinNotIndexAssignableError,
	"NotCallableError":        BuiltinNotCallableError,
	"NotImplementedError":     BuiltinNotImplementedError,
	"ZeroDivisionError":       BuiltinZeroDivisionError,
	"TypeError":               BuiltinTypeError,

	":arrayDestruct": pBuiltinArrayDestruct,
}

// BuiltinObjects is list of builtins, exported for REPL.
var BuiltinObjects = [...]Object{
	pBuiltinArrayDestruct: &BuiltinFunction{
		Name:  ":arrayDestruct",
		Value: pBuiltinArrayDestructFunc,
	},
	BuiltinAppend: &BuiltinFunction{
		Name:  "append",
		Value: builtinAppendFunc,
	},
	BuiltinDelete: &BuiltinFunction{
		Name:  "delete",
		Value: builtinDeleteFunc,
	},
	BuiltinCopy: &BuiltinFunction{
		Name:  "copy",
		Value: builtinCopyFunc,
	},
	BuiltinRepeat: &BuiltinFunction{
		Name:  "repeat",
		Value: builtinRepeatFunc,
	},
	BuiltinContains: &BuiltinFunction{
		Name:  "contains",
		Value: builtinContainsFunc,
	},
	BuiltinLen: &BuiltinFunction{
		Name:  "len",
		Value: builtinLenFunc,
	},
	BuiltinSort: &BuiltinFunction{
		Name:  "sort",
		Value: builtinSortFunc,
	},
	BuiltinSortReverse: &BuiltinFunction{
		Name:  "sortReverse",
		Value: builtinSortReverseFunc,
	},
	BuiltinError: &BuiltinFunction{
		Name:  "error",
		Value: builtinErrorFunc,
	},
	BuiltinTypeName: &BuiltinFunction{
		Name:  "typeName",
		Value: builtinTypeNameFunc,
	},
	BuiltinBool: &BuiltinFunction{
		Name:  "bool",
		Value: builtinBoolFunc,
	},
	BuiltinInt: &BuiltinFunction{
		Name:  "int",
		Value: builtinIntFunc,
	},
	BuiltinUint: &BuiltinFunction{
		Name:  "uint",
		Value: builtinUintFunc,
	},
	BuiltinFloat: &BuiltinFunction{
		Name:  "float",
		Value: builtinFloatFunc,
	},
	BuiltinChar: &BuiltinFunction{
		Name:  "char",
		Value: builtinCharFunc,
	},
	BuiltinString: &BuiltinFunction{
		Name:  "string",
		Value: builtinStringFunc,
	},
	BuiltinBytes: &BuiltinFunction{
		Name:  "bytes",
		Value: builtinBytesFunc,
	},
	BuiltinChars: &BuiltinFunction{
		Name:  "chars",
		Value: builtinCharsFunc,
	},
	BuiltinPrintf: &BuiltinFunction{
		Name:  "printf",
		Value: builtinPrintfFunc,
	},
	BuiltinPrintln: &BuiltinFunction{
		Name:  "println",
		Value: builtinPrintlnFunc,
	},
	BuiltinSprintf: &BuiltinFunction{
		Name:  "sprintf",
		Value: builtinSprintfFunc,
	},
	BuiltinGlobals: &BuiltinFunction{
		Name:  "globals",
		Value: noopFunc, // Value is set at runtime
	},

	BuiltinIsError: &BuiltinFunction{
		Name:  "isError",
		Value: builtinIsErrorFunc,
	},
	BuiltinIsInt: &BuiltinFunction{
		Name:  "isInt",
		Value: builtinIsIntFunc,
	},
	BuiltinIsUint: &BuiltinFunction{
		Name:  "isUint",
		Value: builtinIsUintFunc,
	},
	BuiltinIsFloat: &BuiltinFunction{
		Name:  "isFloat",
		Value: builtinIsFloatFunc,
	},
	BuiltinIsChar: &BuiltinFunction{
		Name:  "isChar",
		Value: builtinIsCharFunc,
	},
	BuiltinIsBool: &BuiltinFunction{
		Name:  "isBool",
		Value: builtinIsBoolFunc,
	},
	BuiltinIsString: &BuiltinFunction{
		Name:  "isString",
		Value: builtinIsStringFunc,
	},
	BuiltinIsBytes: &BuiltinFunction{
		Name:  "isBytes",
		Value: builtinIsBytesFunc,
	},
	BuiltinIsMap: &BuiltinFunction{
		Name:  "isMap",
		Value: builtinIsMapFunc,
	},
	BuiltinIsSyncMap: &BuiltinFunction{
		Name:  "isSyncMap",
		Value: builtinIsSyncMapFunc,
	},
	BuiltinIsArray: &BuiltinFunction{
		Name:  "isArray",
		Value: builtinIsArrayFunc,
	},
	BuiltinIsUndefined: &BuiltinFunction{
		Name:  "isUndefined",
		Value: builtinIsUndefinedFunc,
	},
	BuiltinIsFunction: &BuiltinFunction{
		Name:  "isFunction",
		Value: builtinIsFunctionFunc,
	},
	BuiltinIsCallable: &BuiltinFunction{
		Name:  "isCallable",
		Value: builtinIsCallableFunc,
	},
	BuiltinIsIterable: &BuiltinFunction{
		Name:  "isIterable",
		Value: builtinIsIterableFunc,
	},

	BuiltinWrongNumArgumentsError:  ErrWrongNumArguments,
	BuiltinInvalidOperatorError:    ErrInvalidOperator,
	BuiltinIndexOutOfBoundsError:   ErrIndexOutOfBounds,
	BuiltinNotIterableError:        ErrNotIterable,
	BuiltinNotIndexableError:       ErrNotIndexable,
	BuiltinNotIndexAssignableError: ErrNotIndexAssignable,
	BuiltinNotCallableError:        ErrNotCallable,
	BuiltinNotImplementedError:     ErrNotImplemented,
	BuiltinZeroDivisionError:       ErrZeroDivision,
	BuiltinTypeError:               ErrType,
}

func noopFunc(args ...Object) (Object, error) {
	return Undefined, nil
}

func pBuiltinArrayDestructFunc(args ...Object) (Object, error) {
	if len(args) != 2 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}

	n, ok := args[0].(Int)
	if !ok {
		return nil, NewArgumentTypeError(
			"first",
			"int",
			args[0].TypeName(),
		)
	}
	nn := int(n)
	if nn <= 0 {
		return args[1], nil
	}

	arr, ok := args[1].(Array)
	if !ok {
		ret := make(Array, nn)
		for i := 1; i < nn; i++ {
			ret[i] = Undefined
		}
		ret[0] = args[1]
		return ret, nil
	}
	length := len(arr)
	if nn <= length {
		return arr[:nn], nil
	}
	ret := make(Array, nn)
	x := copy(ret, arr)
	for i := x; i < nn; i++ {
		ret[i] = Undefined
	}
	return ret, nil
}

func builtinAppendFunc(args ...Object) (Object, error) {
	if len(args) < 1 {
		return nil, ErrWrongNumArguments.NewError("want>=1 got=0")
	}
	switch obj := args[0].(type) {
	case Array:
		if len(args) > 1 {
			return append(obj, args[1:]...), nil
		}
		return obj, nil
	case Bytes:
		if len(args) > 1 {
			var rest []byte
			for i, v := range args[1:] {
				switch vv := v.(type) {
				case Int:
					rest = append(rest, byte(vv))
				case Uint:
					rest = append(rest, byte(vv))
				case Char:
					rest = append(rest, byte(vv))
				default:
					return nil, NewArgumentTypeError(
						strconv.Itoa(i+1),
						"int|uint|char",
						args[i+1].TypeName(),
					)
				}
			}
			return append(obj, rest...), nil
		}
		return obj, nil
	case undefined:
		if len(args) > 1 {
			return append(Array{}, args[1:]...), nil
		}
		return Array{}, nil
	default:
		return nil, NewArgumentTypeError(
			"first",
			"array",
			args[0].TypeName(),
		)
	}
}

func builtinDeleteFunc(args ...Object) (Object, error) {
	if len(args) != 2 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	switch arg := args[0].(type) {
	case Map:
		if key, ok := args[1].(String); ok {
			delete(arg, string(key))
			return Undefined, nil
		}
		return nil, NewArgumentTypeError(
			"second",
			"string",
			args[1].TypeName(),
		)
	case *SyncMap:
		if key, ok := args[1].(String); ok {
			if arg.Map == nil {
				return Undefined, nil
			}
			delete(arg.Map, string(key))
			return Undefined, nil
		}
		return nil, NewArgumentTypeError(
			"second",
			"string",
			args[1].TypeName(),
		)
	default:
		return nil, NewArgumentTypeError(
			"first",
			"map|sync-map",
			args[0].TypeName(),
		)
	}
}

func builtinCopyFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	if v, ok := args[0].(Copier); ok {
		return v.Copy(), nil
	}
	return args[0], nil
}

func builtinRepeatFunc(args ...Object) (Object, error) {
	if len(args) != 2 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	var count int
	switch v := args[1].(type) {
	case Int:
		count = int(v)
	case Uint:
		count = int(v)
	default:
		return nil, NewArgumentTypeError(
			"second",
			"int|uint",
			args[1].TypeName(),
		)
	}
	if count < 0 {
		return nil, NewArgumentTypeError(
			"second",
			"non-negative integer",
			args[1].TypeName(),
		)
	}
	switch v := args[0].(type) {
	case Array:
		out := make(Array, 0, len(v)*count)
		for i := 0; i < count; i++ {
			out = append(out, v...)
		}
		return out, nil
	case String:
		return String(strings.Repeat(string(v), count)), nil
	case Bytes:
		return Bytes(bytes.Repeat(v, count)), nil
	}
	return nil, NewArgumentTypeError(
		"first",
		"array|string|bytes",
		args[0].TypeName(),
	)
}

func builtinContainsFunc(args ...Object) (Object, error) {
	if len(args) != 2 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(2, len(args)))
	}
	switch obj := args[0].(type) {
	case Map:
		_, ok := obj[args[1].String()]
		return Bool(ok), nil
	case *SyncMap:
		_, ok := obj.Get(args[1].String())
		return Bool(ok), nil
	case Array:
		search := args[1]
		for _, item := range obj {
			if item.Equal(search) {
				return True, nil
			}
		}
		return False, nil
	case String:
		return Bool(strings.Contains(string(obj), args[1].String())), nil
	case Bytes:
		switch v := args[1].(type) {
		case Int:
			return Bool(bytes.Contains(obj, []byte{byte(v)})), nil
		case Uint:
			return Bool(bytes.Contains(obj, []byte{byte(v)})), nil
		case Char:
			return Bool(bytes.Contains(obj, []byte{byte(v)})), nil
		case String:
			return Bool(bytes.Contains(obj, []byte(v))), nil
		case Bytes:
			return Bool(bytes.Contains(obj, v)), nil
		default:
			return nil, NewArgumentTypeError(
				"second",
				"int|uint|string|char|bytes",
				args[1].TypeName(),
			)
		}
	case undefined:
		return False, nil
	default:
		return nil, NewArgumentTypeError(
			"first",
			"map|array|string|bytes",
			args[0].TypeName(),
		)
	}
}

func builtinLenFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	switch v := args[0].(type) {
	case String:
		return Int(len(v)), nil
	case Array:
		return Int(len(v)), nil
	case Map:
		return Int(len(v)), nil
	case *SyncMap:
		return Int(len(v.Map)), nil
	case Bytes:
		return Int(len(v)), nil
	}
	return Int(0), nil
}

func builtinSortFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	switch obj := args[0].(type) {
	case Array:
		var err error
		sort.Slice(obj, func(i, j int) bool {
			v, e := obj[i].BinaryOp(token.Less, obj[j])
			if e != nil && err == nil {
				err = e
				return false
			}
			if v != nil {
				return !v.IsFalsy()
			}
			return false
		})
		if err != nil {
			return nil, err
		}
		return obj, nil
	case String:
		s := []rune(obj)
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
		return String(s), nil
	case Bytes:
		sort.Slice(obj, func(i, j int) bool {
			return obj[i] < obj[j]
		})
		return obj, nil
	case undefined:
		return Undefined, nil
	}
	return nil, NewArgumentTypeError(
		"first",
		"array|string|bytes",
		args[0].TypeName(),
	)
}

func builtinSortReverseFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	switch obj := args[0].(type) {
	case Array:
		var err error
		sort.Slice(obj, func(i, j int) bool {
			v, e := obj[j].BinaryOp(token.Less, obj[i])
			if e != nil && err == nil {
				err = e
				return false
			}
			if v != nil {
				return !v.IsFalsy()
			}
			return false
		})
		if err != nil {
			return nil, err
		}
		return obj, nil
	case String:
		s := []rune(obj)
		sort.Slice(s, func(i, j int) bool {
			return s[j] < s[i]
		})
		return String(s), nil
	case Bytes:
		sort.Slice(obj, func(i, j int) bool {
			return obj[j] < obj[i]
		})
		return obj, nil
	case undefined:
		return Undefined, nil
	}
	return nil, NewArgumentTypeError(
		"first",
		"array|string|bytes",
		args[0].TypeName(),
	)
}

func builtinErrorFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	return &Error{Name: "error", Message: args[0].String()}, nil
}

func builtinTypeNameFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	return String(args[0].TypeName()), nil
}

func builtinBoolFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	return Bool(!args[0].IsFalsy()), nil
}

func builtinIntFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	switch obj := args[0].(type) {
	case Uint:
		return Int(obj), nil
	case Float:
		return Int(obj), nil
	case Int:
		return obj, nil
	case Char:
		return Int(obj), nil
	case String:
		v, err := strconv.ParseInt(string(obj), 0, 64)
		if err != nil {
			return nil, err
		}
		return Int(v), nil
	case Bool:
		if obj {
			return Int(1), nil
		}
		return Int(0), nil
	default:
		return nil, NewArgumentTypeError(
			"first",
			"numeric|string|bool",
			args[0].TypeName(),
		)
	}
}

func builtinUintFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	switch obj := args[0].(type) {
	case Int:
		return Uint(obj), nil
	case Float:
		return Uint(obj), nil
	case Char:
		return Uint(obj), nil
	case Uint:
		return obj, nil
	case String:
		v, err := strconv.ParseUint(string(obj), 0, 64)
		if err != nil {
			return nil, err
		}
		return Uint(v), nil
	case Bool:
		if obj {
			return Uint(1), nil
		}
		return Uint(0), nil
	default:
		return nil, NewArgumentTypeError(
			"first",
			"numeric|string|bool",
			args[0].TypeName(),
		)
	}
}

func builtinCharFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	switch obj := args[0].(type) {
	case Int:
		return Char(obj), nil
	case Uint:
		return Char(obj), nil
	case Float:
		return Char(obj), nil
	case Char:
		return obj, nil
	case String:
		r, _ := utf8.DecodeRuneInString(string(obj))
		if r == utf8.RuneError {
			return Undefined, nil
		}
		return Char(r), nil
	case Bool:
		if obj {
			return Char(1), nil
		}
		return Char(0), nil
	default:
		return nil, NewArgumentTypeError(
			"first",
			"numeric|string|bool",
			args[0].TypeName(),
		)
	}
}

func builtinFloatFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	switch obj := args[0].(type) {
	case Int:
		return Float(obj), nil
	case Uint:
		return Float(obj), nil
	case Char:
		return Float(obj), nil
	case Float:
		return obj, nil
	case String:
		v, err := strconv.ParseFloat(string(obj), 64)
		if err != nil {
			return nil, err
		}
		return Float(v), nil
	case Bool:
		if obj {
			return Float(1), nil
		}
		return Float(0), nil
	default:
		return nil, NewArgumentTypeError(
			"first",
			"numeric|string|bool",
			args[0].TypeName(),
		)
	}
}

func builtinStringFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	return String(args[0].String()), nil
}

func builtinBytesFunc(args ...Object) (Object, error) {
	if len(args) == 0 {
		return Bytes{}, nil
	}
	switch obj := args[0].(type) {
	case String:
		return Bytes([]byte(obj)), nil
	case Bytes:
		return obj, nil
	default:
		var out Bytes
		for i, obj := range args {
			switch v := obj.(type) {
			case Int:
				out = append(out, byte(v))
			case Uint:
				out = append(out, byte(v))
			case Char:
				out = append(out, byte(v))
			default:
				return nil, NewArgumentTypeError(
					strconv.Itoa(i+1),
					"int|uint|char",
					args[i].TypeName(),
				)
			}
		}
		return out, nil
	}
}

func builtinCharsFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	switch obj := args[0].(type) {
	case String:
		s := string(obj)
		var out = make(Array, 0, utf8.RuneCountInString(s))
		sz := len(obj)
		i := 0
		for i < sz {
			r, w := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError {
				return Undefined, nil
			}
			out = append(out, Char(r))
			i += w
		}
		return out, nil
	case Bytes:
		var out = make(Array, 0, utf8.RuneCount(obj))
		sz := len(obj)
		i := 0
		for i < sz {
			r, w := utf8.DecodeRune(obj[i:])
			if r == utf8.RuneError {
				return Undefined, nil
			}
			out = append(out, Char(r))
			i += w
		}
		return out, nil
	}
	return nil, NewArgumentTypeError(
		"first",
		"string|bytes",
		args[0].TypeName(),
	)
}

func builtinPrintfFunc(args ...Object) (Object, error) {
	if len(args) < 1 {
		return nil, ErrWrongNumArguments.NewError("want>=1 got=0")
	}
	switch len(args) {
	case 1:
		if _, err := fmt.Fprint(PrintWriter, args[0].String()); err != nil {
			return nil, err
		}
	default:
		vargs := make([]interface{}, len(args)-1)
		for i := range args[1:] {
			vargs[i] = args[i+1]
		}
		_, err := fmt.Fprintf(PrintWriter, args[0].String(), vargs...)
		if err != nil {
			return nil, err
		}
	}
	return Undefined, nil
}

func builtinPrintlnFunc(args ...Object) (Object, error) {
	switch len(args) {
	case 0:
		if _, err := fmt.Fprintln(PrintWriter); err != nil {
			return nil, err
		}
	case 1:
		if _, err := fmt.Fprintln(PrintWriter, args[0]); err != nil {
			return nil, err
		}
	default:
		vargs := make([]interface{}, len(args))
		for i := range args {
			vargs[i] = args[i]
		}
		if _, err := fmt.Fprintln(PrintWriter, vargs...); err != nil {
			return nil, err
		}
	}
	return Undefined, nil
}

func builtinSprintfFunc(args ...Object) (Object, error) {
	if len(args) < 1 {
		return nil, ErrWrongNumArguments.NewError("want>=1 got=0")
	}
	vargs := make([]interface{}, len(args)-1)
	for i := range args[1:] {
		vargs[i] = args[i+1]
	}
	return String(fmt.Sprintf(args[0].String(), vargs...)), nil
}

func builtinIsErrorFunc(args ...Object) (Object, error) {
	switch len(args) {
	case 1:
		switch args[0].(type) {
		case *Error, *RuntimeError:
			return True, nil
		}
	case 2:
		if err, ok := args[0].(error); ok {
			if target, ok := args[1].(error); ok {
				return Bool(errors.Is(err, target)), nil
			}
		}
	default:
		return nil, ErrWrongNumArguments.NewError(
			fmt.Sprint("want=1..2 got=", len(args)))
	}
	return False, nil
}

func builtinIsIntFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(Int)
	return Bool(ok), nil
}

func builtinIsUintFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(Uint)
	return Bool(ok), nil
}

func builtinIsFloatFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(Float)
	return Bool(ok), nil
}

func builtinIsCharFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(Char)
	return Bool(ok), nil
}

func builtinIsBoolFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(Bool)
	return Bool(ok), nil
}

func builtinIsStringFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(String)
	return Bool(ok), nil
}

func builtinIsBytesFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(Bytes)
	return Bool(ok), nil
}

func builtinIsMapFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(Map)
	return Bool(ok), nil
}

func builtinIsSyncMapFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(*SyncMap)
	return Bool(ok), nil
}

func builtinIsArrayFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(Array)
	return Bool(ok), nil
}

func builtinIsUndefinedFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	return Bool(args[0] == Undefined), nil
}

func builtinIsFunctionFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	_, ok := args[0].(*CompiledFunction)
	if ok {
		return True, nil
	}
	_, ok = args[0].(*BuiltinFunction)
	if ok {
		return True, nil
	}
	_, ok = args[0].(*Function)
	return Bool(ok), nil
}

func builtinIsCallableFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	return Bool(args[0].CanCall()), nil
}

func builtinIsIterableFunc(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, ErrWrongNumArguments.NewError(wantEqXGotY(1, len(args)))
	}
	return Bool(args[0].CanIterate()), nil
}

func wantEqXGotY(x, y int) string {
	buf := make([]byte, 0, 20)
	buf = append(buf, "want="...)
	buf = strconv.AppendInt(buf, int64(x), 10)
	buf = append(buf, " got="...)
	buf = strconv.AppendInt(buf, int64(y), 10)
	return string(buf)
}

func wantGEqXGotY(x, y int) string {
	buf := make([]byte, 0, 20)
	buf = append(buf, "want>="...)
	buf = strconv.AppendInt(buf, int64(x), 10)
	buf = append(buf, " got="...)
	buf = strconv.AppendInt(buf, int64(y), 10)
	return string(buf)
}
