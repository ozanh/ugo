// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"fmt"
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
