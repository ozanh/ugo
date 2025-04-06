// Copyright (c) 2020-2025 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrSymbolLimit represents a symbol limit error which is returned by
	// Compiler when number of local symbols exceeds the symbo limit for
	// a function that is 256.
	ErrSymbolLimit = &Error{
		Name:    "SymbolLimitError",
		Message: "number of local symbols exceeds the limit",
	}

	// ErrStackOverflow represents a stack overflow error.
	ErrStackOverflow = &Error{Name: "StackOverflowError"}

	// ErrVMAborted represents a VM aborted error.
	ErrVMAborted = &Error{Name: "VMAbortedError"}

	// ErrWrongNumArguments represents a wrong number of arguments error.
	ErrWrongNumArguments = &Error{Name: "WrongNumberOfArgumentsError"}

	// ErrInvalidOperator represents an error for invalid operator usage.
	ErrInvalidOperator = &Error{Name: "InvalidOperatorError"}

	// ErrIndexOutOfBounds represents an out of bounds index error.
	ErrIndexOutOfBounds = &Error{Name: "IndexOutOfBoundsError"}

	// ErrInvalidIndex represents an invalid index error.
	ErrInvalidIndex = &Error{Name: "InvalidIndexError"}

	// ErrNotIterable is an error where an Object is not iterable.
	ErrNotIterable = &Error{Name: "NotIterableError"}

	// ErrNotIndexable is an error where an Object is not indexable.
	ErrNotIndexable = &Error{Name: "NotIndexableError"}

	// ErrNotIndexAssignable is an error where an Object is not index assignable.
	ErrNotIndexAssignable = &Error{Name: "NotIndexAssignableError"}

	// ErrNotCallable is an error where Object is not callable.
	ErrNotCallable = &Error{Name: "NotCallableError"}

	// ErrNotImplemented is an error where an Object has not implemented a required method.
	ErrNotImplemented = &Error{Name: "NotImplementedError"}

	// ErrZeroDivision is an error where divisor is zero.
	ErrZeroDivision = &Error{Name: "ZeroDivisionError"}

	// ErrType represents a type error.
	ErrType = &Error{Name: "TypeError"}
)

// NewOperandTypeError creates a new Error from ErrType.
func NewOperandTypeError(token, leftType, rightType string) *Error {
	return ErrType.NewError(
		fmt.Sprintf("unsupported operand types for '%s': '%s' and '%s'",
			token, leftType, rightType))
}

// NewArgumentTypeError creates a new Error from ErrType.
func NewArgumentTypeError(pos, expectType, foundType string) *Error {
	return ErrType.NewError(
		fmt.Sprintf("invalid type for argument '%s': expected %s, found %s",
			pos, expectType, foundType))
}

// NewIndexTypeError creates a new Error from ErrType.
func NewIndexTypeError(expectType, foundType string) *Error {
	return ErrType.NewError(
		fmt.Sprintf("index type expected %s, found %s", expectType, foundType))
}

// NewIndexValueTypeError creates a new Error from ErrType.
func NewIndexValueTypeError(expectType, foundType string) *Error {
	return ErrType.NewError(
		fmt.Sprintf("index value type expected %s, found %s", expectType, foundType))
}

// Compiler and Optimizer error values, types and functions.

var (
	errEmptyDeclNotAllowed                  error = plainError("empty declaration not allowed")
	errMultipleVariadicParamDecl            error = plainError("multiple variadic param declaration")
	errMultipleExprRhsNotSupported          error = plainError("multiple expressions on the right side not supported")
	errShortVarDeclOpNotAllowedWithSelector error = plainError("operator ':=' not allowed with selector")
	errNoNewVariableOnLhs                   error = plainError("no new variable on the left side")
	errBreakOutsideOfLoop                   error = plainError("break not allowed outside of loop")
	errContinueOutsideOfLoop                error = plainError("continue not allowed outside of loop")
)

type plainError string

func (e plainError) Error() string { return string(e) }

type unresolvedRefError string

func (e unresolvedRefError) Error() string {
	return "unresolved reference \"" + string(e) + "\""
}

type multipleErr []error

func (m multipleErr) errorOrNil() error {
	if len(m) == 0 {
		return nil
	}
	if len(m) == 1 {
		return m[0]
	}
	return m
}

func (m multipleErr) Errors() []error {
	return m
}

func (m multipleErr) Error() string {
	if len(m) == 0 {
		return ""
	}
	if len(m) == 1 {
		return m[0].Error()
	}

	lines := make([]string, len(m))
	for i, err := range m {
		lines[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf("%d errors occurred:\n\t%s\n\n",
		len(m), strings.Join(lines, "\n\t"))
}

func (m multipleErr) Unwrap() error {
	if len(m) == 0 {
		return nil
	}
	if len(m) == 1 {
		return m[0]
	}
	errs := make([]error, len(m))
	copy(errs, m)
	return chainErr(errs)
}

type chainErr []error

func (c chainErr) Errors() []error {
	return c
}

func (c chainErr) Error() string {
	return c[0].Error()
}

func (c chainErr) Unwrap() error {
	if len(c) == 1 {
		return nil
	}
	return c[1:]
}

func (c chainErr) As(target interface{}) bool {
	return errors.As(c[0], target)
}

func (c chainErr) Is(target error) bool {
	return errors.Is(c[0], target)
}
