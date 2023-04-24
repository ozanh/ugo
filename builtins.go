// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
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

	BuiltinMakeArray
	BuiltinCap
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

	":makeArray": BuiltinMakeArray,
	"cap":        BuiltinCap,
}

// BuiltinObjects is list of builtins, exported for REPL.
var BuiltinObjects = [...]Object{
	// :makeArray is a private builtin function to help destructuring array assignments
	BuiltinMakeArray: &BuiltinFunction{
		Name:    ":makeArray",
		Value:   funcPiOROe(builtinMakeArrayFunc),
		ValueEx: funcPiOROeEx(builtinMakeArrayFunc),
	},
	BuiltinAppend: &BuiltinFunction{
		Name:    "append",
		Value:   callExAdapter(builtinAppendFunc),
		ValueEx: builtinAppendFunc,
	},
	BuiltinDelete: &BuiltinFunction{
		Name:    "delete",
		Value:   funcPOsRe(builtinDeleteFunc),
		ValueEx: funcPOsReEx(builtinDeleteFunc),
	},
	BuiltinCopy: &BuiltinFunction{
		Name:    "copy",
		Value:   funcPORO(builtinCopyFunc),
		ValueEx: funcPOROEx(builtinCopyFunc),
	},
	BuiltinRepeat: &BuiltinFunction{
		Name:    "repeat",
		Value:   funcPOiROe(builtinRepeatFunc),
		ValueEx: funcPOiROeEx(builtinRepeatFunc),
	},
	BuiltinContains: &BuiltinFunction{
		Name:    "contains",
		Value:   funcPOOROe(builtinContainsFunc),
		ValueEx: funcPOOROeEx(builtinContainsFunc),
	},
	BuiltinLen: &BuiltinFunction{
		Name:    "len",
		Value:   funcPORO(builtinLenFunc),
		ValueEx: funcPOROEx(builtinLenFunc),
	},
	BuiltinCap: &BuiltinFunction{
		Name:    "cap",
		Value:   funcPORO(builtinCapFunc),
		ValueEx: funcPOROEx(builtinCapFunc),
	},
	BuiltinSort: &BuiltinFunction{
		Name:    "sort",
		Value:   funcPOROe(builtinSortFunc),
		ValueEx: funcPOROeEx(builtinSortFunc),
	},
	BuiltinSortReverse: &BuiltinFunction{
		Name:    "sortReverse",
		Value:   funcPOROe(builtinSortReverseFunc),
		ValueEx: funcPOROeEx(builtinSortReverseFunc),
	},
	BuiltinError: &BuiltinFunction{
		Name:    "error",
		Value:   funcPORO(builtinErrorFunc),
		ValueEx: funcPOROEx(builtinErrorFunc),
	},
	BuiltinTypeName: &BuiltinFunction{
		Name:    "typeName",
		Value:   funcPORO(builtinTypeNameFunc),
		ValueEx: funcPOROEx(builtinTypeNameFunc),
	},
	BuiltinBool: &BuiltinFunction{
		Name:    "bool",
		Value:   funcPORO(builtinBoolFunc),
		ValueEx: funcPOROEx(builtinBoolFunc),
	},
	BuiltinInt: &BuiltinFunction{
		Name:    "int",
		Value:   funcPi64RO(builtinIntFunc),
		ValueEx: funcPi64ROEx(builtinIntFunc),
	},
	BuiltinUint: &BuiltinFunction{
		Name:    "uint",
		Value:   funcPu64RO(builtinUintFunc),
		ValueEx: funcPu64ROEx(builtinUintFunc),
	},
	BuiltinFloat: &BuiltinFunction{
		Name:    "float",
		Value:   funcPf64RO(builtinFloatFunc),
		ValueEx: funcPf64ROEx(builtinFloatFunc),
	},
	BuiltinChar: &BuiltinFunction{
		Name:    "char",
		Value:   funcPOROe(builtinCharFunc),
		ValueEx: funcPOROeEx(builtinCharFunc),
	},
	BuiltinString: &BuiltinFunction{
		Name:    "string",
		Value:   funcPORO(builtinStringFunc),
		ValueEx: funcPOROEx(builtinStringFunc),
	},
	BuiltinBytes: &BuiltinFunction{
		Name:    "bytes",
		Value:   callExAdapter(builtinBytesFunc),
		ValueEx: builtinBytesFunc,
	},
	BuiltinChars: &BuiltinFunction{
		Name:    "chars",
		Value:   funcPOROe(builtinCharsFunc),
		ValueEx: funcPOROeEx(builtinCharsFunc),
	},
	BuiltinPrintf: &BuiltinFunction{
		Name:    "printf",
		Value:   callExAdapter(builtinPrintfFunc),
		ValueEx: builtinPrintfFunc,
	},
	BuiltinPrintln: &BuiltinFunction{
		Name:    "println",
		Value:   callExAdapter(builtinPrintlnFunc),
		ValueEx: builtinPrintlnFunc,
	},
	BuiltinSprintf: &BuiltinFunction{
		Name:    "sprintf",
		Value:   callExAdapter(builtinSprintfFunc),
		ValueEx: builtinSprintfFunc,
	},
	BuiltinGlobals: &BuiltinFunction{
		Name:    "globals",
		Value:   callExAdapter(builtinGlobalsFunc),
		ValueEx: builtinGlobalsFunc,
	},
	BuiltinIsError: &BuiltinFunction{
		Name:    "isError",
		Value:   callExAdapter(builtinIsErrorFunc),
		ValueEx: builtinIsErrorFunc,
	},
	BuiltinIsInt: &BuiltinFunction{
		Name:    "isInt",
		Value:   funcPORO(builtinIsIntFunc),
		ValueEx: funcPOROEx(builtinIsIntFunc),
	},
	BuiltinIsUint: &BuiltinFunction{
		Name:    "isUint",
		Value:   funcPORO(builtinIsUintFunc),
		ValueEx: funcPOROEx(builtinIsUintFunc),
	},
	BuiltinIsFloat: &BuiltinFunction{
		Name:    "isFloat",
		Value:   funcPORO(builtinIsFloatFunc),
		ValueEx: funcPOROEx(builtinIsFloatFunc),
	},
	BuiltinIsChar: &BuiltinFunction{
		Name:    "isChar",
		Value:   funcPORO(builtinIsCharFunc),
		ValueEx: funcPOROEx(builtinIsCharFunc),
	},
	BuiltinIsBool: &BuiltinFunction{
		Name:    "isBool",
		Value:   funcPORO(builtinIsBoolFunc),
		ValueEx: funcPOROEx(builtinIsBoolFunc),
	},
	BuiltinIsString: &BuiltinFunction{
		Name:    "isString",
		Value:   funcPORO(builtinIsStringFunc),
		ValueEx: funcPOROEx(builtinIsStringFunc),
	},
	BuiltinIsBytes: &BuiltinFunction{
		Name:    "isBytes",
		Value:   funcPORO(builtinIsBytesFunc),
		ValueEx: funcPOROEx(builtinIsBytesFunc),
	},
	BuiltinIsMap: &BuiltinFunction{
		Name:    "isMap",
		Value:   funcPORO(builtinIsMapFunc),
		ValueEx: funcPOROEx(builtinIsMapFunc),
	},
	BuiltinIsSyncMap: &BuiltinFunction{
		Name:    "isSyncMap",
		Value:   funcPORO(builtinIsSyncMapFunc),
		ValueEx: funcPOROEx(builtinIsSyncMapFunc),
	},
	BuiltinIsArray: &BuiltinFunction{
		Name:    "isArray",
		Value:   funcPORO(builtinIsArrayFunc),
		ValueEx: funcPOROEx(builtinIsArrayFunc),
	},
	BuiltinIsUndefined: &BuiltinFunction{
		Name:    "isUndefined",
		Value:   funcPORO(builtinIsUndefinedFunc),
		ValueEx: funcPOROEx(builtinIsUndefinedFunc),
	},
	BuiltinIsFunction: &BuiltinFunction{
		Name:    "isFunction",
		Value:   funcPORO(builtinIsFunctionFunc),
		ValueEx: funcPOROEx(builtinIsFunctionFunc),
	},
	BuiltinIsCallable: &BuiltinFunction{
		Name:  "isCallable",
		Value: funcPORO(builtinIsCallableFunc),
		//ValueEx: funcPOROEx(builtinIsCallableFunc),
	},
	BuiltinIsIterable: &BuiltinFunction{
		Name:    "isIterable",
		Value:   funcPORO(builtinIsIterableFunc),
		ValueEx: funcPOROEx(builtinIsIterableFunc),
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

func builtinMakeArrayFunc(n int, arg Object) (Object, error) {
	if n <= 0 {
		return arg, nil
	}

	arr, ok := arg.(Array)
	if !ok {
		ret := make(Array, n)
		for i := 1; i < n; i++ {
			ret[i] = Undefined
		}
		ret[0] = arg
		return ret, nil
	}

	length := len(arr)
	if n <= length {
		return arr[:n], nil
	}

	ret := make(Array, n)
	x := copy(ret, arr)
	for i := x; i < n; i++ {
		ret[i] = Undefined
	}
	return ret, nil
}

func builtinAppendFunc(c Call) (Object, error) {
	target, ok := c.shift()
	if !ok {
		return Undefined, ErrWrongNumArguments.NewError("want>=1 got=0")
	}

	switch obj := target.(type) {
	case Array:
		obj = append(obj, c.args...)
		obj = append(obj, c.vargs...)
		return obj, nil
	case Bytes:
		n := 0
		for _, args := range [][]Object{c.args, c.vargs} {
			for _, v := range args {
				n++
				switch vv := v.(type) {
				case Int:
					obj = append(obj, byte(vv))
				case Uint:
					obj = append(obj, byte(vv))
				case Char:
					obj = append(obj, byte(vv))
				default:
					return Undefined, NewArgumentTypeError(
						strconv.Itoa(n),
						"int|uint|char",
						vv.TypeName(),
					)
				}
			}
		}
		return obj, nil
	case *UndefinedType:
		ret := make(Array, 0, c.Len())
		ret = append(ret, c.args...)
		ret = append(ret, c.vargs...)
		return ret, nil
	default:
		return Undefined, NewArgumentTypeError(
			"1st",
			"array",
			obj.TypeName(),
		)
	}
}

func builtinDeleteFunc(arg Object, key string) (err error) {
	if v, ok := arg.(IndexDeleter); ok {
		err = v.IndexDelete(String(key))
	} else {
		err = NewArgumentTypeError(
			"1st",
			"map|syncMap|IndexDeleter",
			arg.TypeName(),
		)
	}
	return
}

func builtinCopyFunc(arg Object) Object {
	if v, ok := arg.(Copier); ok {
		return v.Copy()
	}
	return arg
}

func builtinRepeatFunc(arg Object, count int) (ret Object, err error) {
	if count < 0 {
		return nil, NewArgumentTypeError(
			"2nd",
			"non-negative integer",
			"negative integer",
		)
	}

	switch v := arg.(type) {
	case Array:
		out := make(Array, 0, len(v)*count)
		for i := 0; i < count; i++ {
			out = append(out, v...)
		}
		ret = out
	case String:
		ret = String(strings.Repeat(string(v), count))
	case Bytes:
		ret = Bytes(bytes.Repeat(v, count))
	default:
		err = NewArgumentTypeError(
			"1st",
			"array|string|bytes",
			arg.TypeName(),
		)
	}
	return
}

func builtinContainsFunc(arg0, arg1 Object) (Object, error) {
	var ok bool
	switch obj := arg0.(type) {
	case Map:
		_, ok = obj[arg1.String()]
	case *SyncMap:
		_, ok = obj.Get(arg1.String())
	case Array:
		for _, item := range obj {
			if item.Equal(arg1) {
				ok = true
				break
			}
		}
	case String:
		ok = strings.Contains(string(obj), arg1.String())
	case Bytes:
		switch v := arg1.(type) {
		case Int:
			ok = bytes.Contains(obj, []byte{byte(v)})
		case Uint:
			ok = bytes.Contains(obj, []byte{byte(v)})
		case Char:
			ok = bytes.Contains(obj, []byte{byte(v)})
		case String:
			ok = bytes.Contains(obj, []byte(v))
		case Bytes:
			ok = bytes.Contains(obj, v)
		default:
			return Undefined, NewArgumentTypeError(
				"2nd",
				"int|uint|string|char|bytes",
				arg1.TypeName(),
			)
		}
	case *UndefinedType:
	default:
		return Undefined, NewArgumentTypeError(
			"1st",
			"map|array|string|bytes",
			arg0.TypeName(),
		)
	}
	return Bool(ok), nil
}

func builtinLenFunc(arg Object) Object {
	var n int
	if v, ok := arg.(LengthGetter); ok {
		n = v.Len()
	}
	return Int(n)
}

func builtinCapFunc(arg Object) Object {
	var n int
	switch v := arg.(type) {
	case Array:
		n = cap(v)
	case Bytes:
		n = cap(v)
	}
	return Int(n)
}

func builtinSortFunc(arg Object) (ret Object, err error) {
	switch obj := arg.(type) {
	case Array:
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
		ret = arg
	case String:
		s := []rune(obj)
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
		ret = String(s)
	case Bytes:
		sort.Slice(obj, func(i, j int) bool {
			return obj[i] < obj[j]
		})
		ret = arg
	case *UndefinedType:
		ret = Undefined
	default:
		ret = Undefined
		err = NewArgumentTypeError(
			"1st",
			"array|string|bytes",
			arg.TypeName(),
		)
	}
	return
}

func builtinSortReverseFunc(arg Object) (Object, error) {
	switch obj := arg.(type) {
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
	case *UndefinedType:
		return Undefined, nil
	}

	return Undefined, NewArgumentTypeError(
		"1st",
		"array|string|bytes",
		arg.TypeName(),
	)
}

func builtinErrorFunc(arg Object) Object {
	return &Error{Name: "error", Message: arg.String()}
}

func builtinTypeNameFunc(arg Object) Object { return String(arg.TypeName()) }

func builtinBoolFunc(arg Object) Object { return Bool(!arg.IsFalsy()) }

func builtinIntFunc(v int64) Object { return Int(v) }

func builtinUintFunc(v uint64) Object { return Uint(v) }

func builtinFloatFunc(v float64) Object { return Float(v) }

func builtinCharFunc(arg Object) (Object, error) {
	v, ok := ToChar(arg)
	if ok && v != utf8.RuneError {
		return v, nil
	}
	if v == utf8.RuneError || arg == Undefined {
		return Undefined, nil
	}
	return Undefined, NewArgumentTypeError(
		"1st",
		"numeric|string|bool",
		arg.TypeName(),
	)
}

func builtinStringFunc(arg Object) Object { return String(arg.String()) }

func builtinBytesFunc(c Call) (Object, error) {
	size := c.Len()

	switch size {
	case 0:
		return Bytes{}, nil
	case 1:
		if v, ok := ToBytes(c.Get(0)); ok {
			return v, nil
		}
	}

	out := make(Bytes, 0, size)
	for _, args := range [][]Object{c.args, c.vargs} {
		for i, obj := range args {
			switch v := obj.(type) {
			case Int:
				out = append(out, byte(v))
			case Uint:
				out = append(out, byte(v))
			case Char:
				out = append(out, byte(v))
			default:
				return Undefined, NewArgumentTypeError(
					strconv.Itoa(i+1),
					"int|uint|char",
					args[i].TypeName(),
				)
			}
		}
	}
	return out, nil
}

func builtinCharsFunc(arg Object) (ret Object, err error) {
	switch obj := arg.(type) {
	case String:
		s := string(obj)
		ret = make(Array, 0, utf8.RuneCountInString(s))
		sz := len(obj)
		i := 0

		for i < sz {
			r, w := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError {
				return Undefined, nil
			}
			ret = append(ret.(Array), Char(r))
			i += w
		}
	case Bytes:
		ret = make(Array, 0, utf8.RuneCount(obj))
		sz := len(obj)
		i := 0

		for i < sz {
			r, w := utf8.DecodeRune(obj[i:])
			if r == utf8.RuneError {
				return Undefined, nil
			}
			ret = append(ret.(Array), Char(r))
			i += w
		}
	default:
		ret = Undefined
		err = NewArgumentTypeError(
			"1st",
			"string|bytes",
			arg.TypeName(),
		)
	}
	return
}

func builtinPrintfFunc(c Call) (ret Object, err error) {
	ret = Undefined
	switch size := c.Len(); size {
	case 0:
		err = ErrWrongNumArguments.NewError("want>=1 got=0")
	case 1:
		_, err = fmt.Fprint(PrintWriter, c.Get(0).String())
	default:
		format, _ := c.shift()
		vargs := make([]interface{}, 0, size-1)
		for i := 0; i < size-1; i++ {
			vargs = append(vargs, c.Get(i))
		}
		_, err = fmt.Fprintf(PrintWriter, format.String(), vargs...)
	}
	return
}

func builtinPrintlnFunc(c Call) (ret Object, err error) {
	ret = Undefined
	switch size := c.Len(); size {
	case 0:
		_, err = fmt.Fprintln(PrintWriter)
	case 1:
		_, err = fmt.Fprintln(PrintWriter, c.Get(0))
	default:
		vargs := make([]interface{}, 0, size)
		for i := 0; i < size; i++ {
			vargs = append(vargs, c.Get(i))
		}
		_, err = fmt.Fprintln(PrintWriter, vargs...)
	}
	return
}

func builtinSprintfFunc(c Call) (ret Object, err error) {
	ret = Undefined
	switch size := c.Len(); size {
	case 0:
		err = ErrWrongNumArguments.NewError("want>=1 got=0")
	case 1:
		ret = String(c.Get(0).String())
	default:
		format, _ := c.shift()
		vargs := make([]interface{}, 0, size-1)
		for i := 0; i < size-1; i++ {
			vargs = append(vargs, c.Get(i))
		}
		ret = String(fmt.Sprintf(format.String(), vargs...))
	}
	return
}

func builtinGlobalsFunc(c Call) (Object, error) {
	return c.VM().GetGlobals(), nil
}

func builtinIsErrorFunc(c Call) (ret Object, err error) {
	ret = False
	switch c.Len() {
	case 1:
		// We have Error, BuiltinError and also user defined error types.
		if _, ok := c.Get(0).(error); ok {
			ret = True
		}
	case 2:
		if err, ok := c.Get(0).(error); ok {
			if target, ok := c.Get(1).(error); ok {
				ret = Bool(errors.Is(err, target))
			}
		}
	default:
		err = ErrWrongNumArguments.NewError(
			"want=1..2 got=", strconv.Itoa(c.Len()))
	}
	return
}

func builtinIsIntFunc(arg Object) Object {
	_, ok := arg.(Int)
	return Bool(ok)
}

func builtinIsUintFunc(arg Object) Object {
	_, ok := arg.(Uint)
	return Bool(ok)
}

func builtinIsFloatFunc(arg Object) Object {
	_, ok := arg.(Float)
	return Bool(ok)
}

func builtinIsCharFunc(arg Object) Object {
	_, ok := arg.(Char)
	return Bool(ok)
}

func builtinIsBoolFunc(arg Object) Object {
	_, ok := arg.(Bool)
	return Bool(ok)
}

func builtinIsStringFunc(arg Object) Object {
	_, ok := arg.(String)
	return Bool(ok)
}

func builtinIsBytesFunc(arg Object) Object {
	_, ok := arg.(Bytes)
	return Bool(ok)
}

func builtinIsMapFunc(arg Object) Object {
	_, ok := arg.(Map)
	return Bool(ok)
}

func builtinIsSyncMapFunc(arg Object) Object {
	_, ok := arg.(*SyncMap)
	return Bool(ok)
}

func builtinIsArrayFunc(arg Object) Object {
	_, ok := arg.(Array)
	return Bool(ok)
}

func builtinIsUndefinedFunc(arg Object) Object {
	_, ok := arg.(*UndefinedType)
	return Bool(ok)
}

func builtinIsFunctionFunc(arg Object) Object {
	_, ok := arg.(*CompiledFunction)
	if ok {
		return True
	}

	_, ok = arg.(*BuiltinFunction)
	if ok {
		return True
	}

	_, ok = arg.(*Function)
	return Bool(ok)
}

func builtinIsCallableFunc(arg Object) Object { return Bool(arg.CanCall()) }

func builtinIsIterableFunc(arg Object) Object { return Bool(arg.CanIterate()) }

func callExAdapter(fn CallableExFunc) CallableFunc {
	return func(args ...Object) (Object, error) {
		return fn(Call{args: args})
	}
}
