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
		Name:  "Print",
		Value: fnPrint(fmt.Print),
	},
	// ugo:doc
	// Printf(format string, ...any) -> int
	// Formats according to a format specifier and writes to standard output.
	// It returns the number of bytes written and any encountered write error
	// throws a runtime error.
	"Printf": &ugo.Function{
		Name:  "Printf",
		Value: fnPrintf(fmt.Printf),
	},
	// ugo:doc
	// Println(...any) -> int
	// Formats using the default formats for its operands and writes to standard
	// output. Spaces are always added between operands and a newline
	// is appended. It returns the number of bytes written and any encountered
	// write error throws a runtime error.
	"Println": &ugo.Function{
		Name:  "Println",
		Value: fnPrint(fmt.Println),
	},
	// ugo:doc
	// Sprint(...any) -> string
	// Formats using the default formats for its operands and returns the
	// resulting string. Spaces are added between operands when neither is a
	// string.
	"Sprint": &ugo.Function{
		Name:  "Sprint",
		Value: fnSprint(fmt.Sprint),
	},
	// ugo:doc
	// Sprintf(format string, ...any) -> string
	// Formats according to a format specifier and returns the resulting string.
	"Sprintf": &ugo.Function{
		Name:  "Sprintf",
		Value: fnSprintf(fmt.Sprintf),
	},
	// ugo:doc
	// Sprintln(...any) -> string
	// Formats using the default formats for its operands and returns the
	// resulting string. Spaces are always added between operands and a newline
	// is appended.
	"Sprintln": &ugo.Function{
		Name:  "Sprintln",
		Value: fnSprint(fmt.Sprintln),
	},
	// ugo:doc
	// Sscan(str string, ScanArg[, ...ScanArg]) -> int | error
	// Scans the argument string, storing successive space-separated values into
	// successive ScanArg arguments. Newlines count as space. If no error is
	// encountered, it returns the number of items successfully scanned. If that
	// is less than the number of arguments, error will report why.
	"Sscan": &ugo.Function{
		Name:  "Sscan",
		Value: fnSscan(fmt.Sscan),
	},
	// ugo:doc
	// Sscanf(str string, format string, ScanArg[, ...ScanArg]) -> int | error
	// Scans the argument string, storing successive space-separated values into
	// successive ScanArg arguments as determined by the format. It returns the
	// number of items successfully parsed or an error.
	// Newlines in the input must match newlines in the format.
	"Sscanf": &ugo.Function{
		Name:  "Sscanf",
		Value: fnSscanf(fmt.Sscanf),
	},
	// Sscanln(str string, ScanArg[, ...ScanArg]) -> int | error
	// Sscanln is similar to Sscan, but stops scanning at a newline and after
	// the final item there must be a newline or EOF. It returns the number of
	// items successfully parsed or an error.
	"Sscanln": &ugo.Function{
		Name:  "Sscanln",
		Value: fnSscan(fmt.Sscanln),
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
		Name:  "ScanArg",
		Value: newScanArg,
	},
}

func fnPrint(fn func(...interface{}) (int, error)) ugo.CallableFunc {

	return func(args ...ugo.Object) (ret ugo.Object, err error) {
		vargs := argsToPrintArgs(0, args)
		n, err := fn(vargs...)
		if err != nil {
			return nil, err
		}
		return ugo.Int(n), nil
	}
}

func fnPrintf(fn func(string, ...interface{}) (int, error)) ugo.CallableFunc {

	return func(args ...ugo.Object) (ret ugo.Object, err error) {
		if len(args) < 1 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				"want>=1 got=" + strconv.Itoa(len(args)))
		}
		vargs := argsToPrintArgs(1, args)
		n, err := fn(args[0].String(), vargs...)
		if err != nil {
			return nil, err
		}
		return ugo.Int(n), nil
	}
}

func fnSprint(fn func(...interface{}) string) ugo.CallableFunc {

	return func(args ...ugo.Object) (ret ugo.Object, err error) {
		vargs := argsToPrintArgs(0, args)
		return ugo.String(fn(vargs...)), nil
	}
}

func fnSprintf(fn func(string, ...interface{}) string) ugo.CallableFunc {

	return func(args ...ugo.Object) (ret ugo.Object, err error) {
		if len(args) < 1 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				"want>=1 got=" + strconv.Itoa(len(args)))
		}
		vargs := argsToPrintArgs(1, args)
		return ugo.String(fn(args[0].String(), vargs...)), nil
	}
}

func fnSscan(fn func(string, ...interface{}) (int, error)) ugo.CallableFunc {

	return func(args ...ugo.Object) (ret ugo.Object, err error) {
		if len(args) < 2 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				"want>=2 got=" + strconv.Itoa(len(args)))
		}
		offset := 1
		vargs, err := argsToScanArgs(offset, args)
		if err != nil {
			return nil, err
		}
		n, err := fn(args[0].String(), vargs...)
		return postScan(n, err, args[offset:]), nil
	}
}

func fnSscanf(
	fn func(string, string, ...interface{}) (int, error),
) ugo.CallableFunc {

	return func(args ...ugo.Object) (ret ugo.Object, err error) {
		if len(args) < 3 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				"want>=3 got=" + strconv.Itoa(len(args)))
		}
		offset := 2
		vargs, err := argsToScanArgs(offset, args)
		if err != nil {
			return nil, err
		}
		n, err := fn(args[0].String(), args[1].String(), vargs...)
		return postScan(n, err, args[offset:]), nil
	}
}

func argsToScanArgs(offset int, args []ugo.Object) ([]interface{}, error) {
	vargs := make([]interface{}, 0, len(args)-offset)
	for i := offset; i < len(args); i++ {
		v, ok := args[i].(ScanArg)
		if !ok {
			return nil, ugo.NewArgumentTypeError(strconv.Itoa(i),
				"ScanArg interface", args[i].TypeName())
		}
		v.Set(false)
		vargs = append(vargs, v.Arg())
	}
	return vargs, nil
}

func argsToPrintArgs(offset int, args []ugo.Object) []interface{} {
	vargs := make([]interface{}, 0, len(args)-offset)
	for i := offset; i < len(args); i++ {
		vargs = append(vargs, args[i])
	}
	return vargs
}

// args are always of ScanArg interface type.
func postScan(n int, err error, args []ugo.Object) ugo.Object {
	for i := 0; i < n; i++ {
		args[i].(ScanArg).Set(true)
	}
	if err != nil {
		return &ugo.Error{
			Message: err.Error(),
			Cause:   err,
		}
	}
	return ugo.Int(n)
}
