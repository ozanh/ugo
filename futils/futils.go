package futils

//go:generate go run ../cmd/mkcallable -output zfutils.go futils.go

// builtin delete
//
//ugo:callable func(o ugo.Object, k string) (err error)

// builtin copy, len, error, typeName, bool, string, isInt, isUint
// isFloat, isChar, isBool, isString, isBytes, isMap, isSyncMap, isArray
// isUndefined, isFunction, isCallable, isIterable
// time module IsTime
// json module Marshal, Quote, NoQuote, NoEscape
//
//ugo:callable func(o ugo.Object) (ret ugo.Object)

// builtin repeat
//
//ugo:callable func(o ugo.Object, n int) (ret ugo.Object, err error)

// builtin contains
//
//ugo:callable func(o ugo.Object, v ugo.Object) (ret ugo.Object, err error)

// builtin sort, sortReverse, int, uint, float, char, chars
//
//ugo:callable func(o ugo.Object) (ret ugo.Object, err error)

// time module MountString, WeekdayString
//
//ugo:callable func(i1 int) (ret ugo.Object, err error)

// time module DurationString, DurationHours, DurationMinutes, DurationSeconds
// DurationMilliseconds, DurationMicroseconds, DurationNanoseconds, Sleep
//
//ugo:callable func(i1 int64) (ret ugo.Object, err error)

// time module ParseDuration, LoadLocation
//
//ugo:callable func(s string) (ret ugo.Object, err error)

// time module FixedZone
//
//ugo:callable func(s string, i1 int64) (ret ugo.Object, err error)

// time module Time, Now
//
//ugo:callable func() (ret ugo.Object)

// time module DurationRound, DurationTruncate
//
//ugo:callable func(i1 int64, i2 int64) (ret ugo.Object)

// json module Unmarshal, RawMessage, Valid
//
//ugo:callable func(b []byte) (ret ugo.Object)

// json module MarshalIndent
//
//ugo:callable func(o ugo.Object, s1 string, s2 string) (ret ugo.Object)

// json module Compact
//
//ugo:callable func(p []byte, b bool) (ret ugo.Object)

// json module Indent
//
//ugo:callable func(p []byte, s1 string, s2 string) (ret ugo.Object)
