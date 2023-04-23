// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package fmt

import (
	"fmt"
	"strconv"

	"github.com/ozanh/ugo"
)

// Module represents fmt module.
var Module = map[string]ugo.Object{
	// ugo:doc
	// # fmt Module
	//
	// ## Scan Examples
	//
	// ```go
	// arg1 := fmt.ScanArg("string")
	// arg2 := fmt.ScanArg("int")
	// ret := fmt.Sscanf("abc123", "%3s%d", arg1, arg2)
	// if isError(ret) {
	//   // handle error
	//   fmt.Println(err)
	// } else {
	//   fmt.Println(ret)            // 2, number of scanned items
	//   fmt.Println(arg1.Value)     // abc
	//   fmt.Println(bool(arg1))     // true, reports whether arg1 is scanned
	//   fmt.Println(arg2.Value)     // 123
	//   fmt.Println(bool(arg2))     // true, reports whether arg2 is scanned
	// }
	// ```
	//
	// ```go
	// arg1 = fmt.ScanArg("string")
	// arg2 = fmt.ScanArg("int")
	// arg3 = fmt.ScanArg("float")
	// ret = fmt.Sscanf("abc 123", "%s%d%f", arg1, arg2, arg3)
	// fmt.Println(ret)         // error: EOF
	// fmt.Println(arg1.Value)  // abc
	// fmt.Println(bool(arg1))  // true
	// fmt.Println(arg2.Value)  // 123
	// fmt.Println(bool(arg2))  // true
	// fmt.Println(arg3.Value)  // undefined
	// fmt.Println(bool(arg2))  // false, not scanned
	//
	// // Use if statement or a ternary expression to get the scanned value or a default value.
	// v := arg1 ? arg1.Value : "default value"
	// ```

	// ugo:doc
	// ## Functions
	// Print(...any) -> int
	// Formats using the default formats for its operands and writes to standard
	// output. Spaces are added between operands when neither is a string.
	// It returns the number of bytes written and any encountered write error
	// throws a runtime error.
	"Print": &ugo.Function{
		Name: "Print",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newPrint(fmt.Print)(ugo.NewCall(nil, args))
		},
		ValueEx: newPrint(fmt.Print),
	},
	// ugo:doc
	// Printf(format string, ...any) -> int
	// Formats according to a format specifier and writes to standard output.
	// It returns the number of bytes written and any encountered write error
	// throws a runtime error.
	"Printf": &ugo.Function{
		Name: "Printf",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newPrintf(fmt.Printf)(ugo.NewCall(nil, args))
		},
		ValueEx: newPrintf(fmt.Printf),
	},
	// ugo:doc
	// Println(...any) -> int
	// Formats using the default formats for its operands and writes to standard
	// output. Spaces are always added between operands and a newline
	// is appended. It returns the number of bytes written and any encountered
	// write error throws a runtime error.
	"Println": &ugo.Function{
		Name: "Println",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newPrint(fmt.Println)(ugo.NewCall(nil, args))
		},
		ValueEx: newPrint(fmt.Println),
	},
	// ugo:doc
	// Sprint(...any) -> string
	// Formats using the default formats for its operands and returns the
	// resulting string. Spaces are added between operands when neither is a
	// string.
	"Sprint": &ugo.Function{
		Name: "Sprint",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newSprint(fmt.Sprint)(ugo.NewCall(nil, args))
		},
		ValueEx: newSprint(fmt.Sprint),
	},
	// ugo:doc
	// Sprintf(format string, ...any) -> string
	// Formats according to a format specifier and returns the resulting string.
	"Sprintf": &ugo.Function{
		Name: "Sprintf",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newSprintf(fmt.Sprintf)(ugo.NewCall(nil, args))
		},
		ValueEx: newSprintf(fmt.Sprintf),
	},
	// ugo:doc
	// Sprintln(...any) -> string
	// Formats using the default formats for its operands and returns the
	// resulting string. Spaces are always added between operands and a newline
	// is appended.
	"Sprintln": &ugo.Function{
		Name: "Sprintln",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newSprint(fmt.Sprintln)(ugo.NewCall(nil, args))
		},
		ValueEx: newSprint(fmt.Sprintln),
	},
	// ugo:doc
	// Sscan(str string, ScanArg[, ...ScanArg]) -> int | error
	// Scans the argument string, storing successive space-separated values into
	// successive ScanArg arguments. Newlines count as space. If no error is
	// encountered, it returns the number of items successfully scanned. If that
	// is less than the number of arguments, error will report why.
	"Sscan": &ugo.Function{
		Name: "Sscan",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newSscan(fmt.Sscan)(ugo.NewCall(nil, args))
		},
		ValueEx: newSscan(fmt.Sscan),
	},
	// ugo:doc
	// Sscanf(str string, format string, ScanArg[, ...ScanArg]) -> int | error
	// Scans the argument string, storing successive space-separated values into
	// successive ScanArg arguments as determined by the format. It returns the
	// number of items successfully parsed or an error.
	// Newlines in the input must match newlines in the format.
	"Sscanf": &ugo.Function{
		Name: "Sscanf",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newSscanf(fmt.Sscanf)(ugo.NewCall(nil, args))
		},
		ValueEx: newSscanf(fmt.Sscanf),
	},
	// Sscanln(str string, ScanArg[, ...ScanArg]) -> int | error
	// Sscanln is similar to Sscan, but stops scanning at a newline and after
	// the final item there must be a newline or EOF. It returns the number of
	// items successfully parsed or an error.
	"Sscanln": &ugo.Function{
		Name: "Sscanln",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newSscan(fmt.Sscanln)(ugo.NewCall(nil, args))
		},
		ValueEx: newSscan(fmt.Sscanln),
	},
	// ugo:doc
	// ScanArg(typeName string) -> scanArg
	// Returns a `scanArg` object to scan a value of given type name in scan
	// functions.
	// Supported type names are `"string", "int", "uint", "float", "char",
	// "bool", "bytes"`.
	// It throws a runtime error if type name is not supported.
	// Alternatively, `string, int, uint, float, char, bool, bytes` builtin
	// functions can be provided to get the type name from the BuiltinFunction's
	// Name field.
	"ScanArg": &ugo.Function{
		Name: "ScanArg",
		Value: func(args ...ugo.Object) (ugo.Object, error) {
			return newScanArgFunc(ugo.NewCall(nil, args))
		},
		ValueEx: newScanArgFunc,
	},
}

func newPrint(fn func(...interface{}) (int, error)) ugo.CallableExFunc {
	return func(c ugo.Call) (ret ugo.Object, err error) {
		vargs := toPrintArgs(0, c)
		n, err := fn(vargs...)
		return ugo.Int(n), err
	}
}

func newPrintf(fn func(string, ...interface{}) (int, error)) ugo.CallableExFunc {
	return func(c ugo.Call) (ret ugo.Object, err error) {
		if c.Len() < 1 {
			return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
				"want>=1 got=" + strconv.Itoa(c.Len()))
		}
		vargs := toPrintArgs(1, c)
		n, err := fn(c.Get(0).String(), vargs...)
		return ugo.Int(n), err
	}
}

func newSprint(fn func(...interface{}) string) ugo.CallableExFunc {
	return func(c ugo.Call) (ret ugo.Object, err error) {
		vargs := toPrintArgs(0, c)
		return ugo.String(fn(vargs...)), nil
	}
}

func newSprintf(fn func(string, ...interface{}) string) ugo.CallableExFunc {
	return func(c ugo.Call) (ret ugo.Object, err error) {
		if c.Len() < 1 {
			return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
				"want>=1 got=" + strconv.Itoa(c.Len()))
		}
		vargs := toPrintArgs(1, c)
		return ugo.String(fn(c.Get(0).String(), vargs...)), nil
	}
}

func newSscan(fn func(string, ...interface{}) (int, error)) ugo.CallableExFunc {
	return func(c ugo.Call) (ret ugo.Object, err error) {
		if c.Len() < 2 {
			return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
				"want>=2 got=" + strconv.Itoa(c.Len()))
		}
		vargs, err := toScanArgs(1, c)
		if err != nil {
			return ugo.Undefined, err
		}
		n, err := fn(c.Get(0).String(), vargs...)
		return postScan(1, n, err, c), nil
	}
}

func newSscanf(
	fn func(string, string, ...interface{}) (int, error),
) ugo.CallableExFunc {
	return func(c ugo.Call) (ret ugo.Object, err error) {
		if c.Len() < 3 {
			return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
				"want>=3 got=" + strconv.Itoa(c.Len()))
		}
		vargs, err := toScanArgs(2, c)
		if err != nil {
			return ugo.Undefined, err
		}
		n, err := fn(c.Get(0).String(), c.Get(1).String(), vargs...)
		return postScan(2, n, err, c), nil
	}
}

func toScanArgs(offset int, c ugo.Call) ([]interface{}, error) {
	size := c.Len()
	vargs := make([]interface{}, 0, size-offset)
	for i := offset; i < size; i++ {
		v, ok := c.Get(i).(ScanArg)
		if !ok {
			return nil, ugo.NewArgumentTypeError(strconv.Itoa(i),
				"ScanArg interface", c.Get(i).TypeName())
		}
		v.Set(false)
		vargs = append(vargs, v.Arg())
	}
	return vargs, nil
}

func toPrintArgs(offset int, c ugo.Call) []interface{} {
	size := c.Len()
	vargs := make([]interface{}, 0, size-offset)
	for i := offset; i < size; i++ {
		vargs = append(vargs, c.Get(i))
	}
	return vargs
}

// args are always of ScanArg interface type.
func postScan(offset, n int, err error, c ugo.Call) ugo.Object {
	for i := offset; i < n+offset; i++ {
		c.Get(i).(ScanArg).Set(true)
	}
	if err != nil {
		return &ugo.Error{
			Message: err.Error(),
			Cause:   err,
		}
	}
	return ugo.Int(n)
}
