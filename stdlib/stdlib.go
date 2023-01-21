package stdlib

//go:generate go run ../cmd/mkcallable -export -output zfuncs.go stdlib.go

// time module IsTime
// json module Marshal, Quote, NoQuote, NoEscape
//
//ugo:callable func(o ugo.Object) (ret ugo.Object)

// time module MountString, WeekdayString
//
//ugo:callable func(i1 int) (ret ugo.Object)

// time module DurationString, DurationHours, DurationMinutes, DurationSeconds
// DurationMilliseconds, DurationMicroseconds, DurationNanoseconds
//
//ugo:callable func(i1 int64) (ret ugo.Object)

// time module Sleep
//
//ugo:callable func(i1 int64)

// time module ParseDuration, LoadLocation
//
//ugo:callable func(s string) (ret ugo.Object, err error)

// time module FixedZone
// strings module Repeat
//
//ugo:callable func(s string, i1 int) (ret ugo.Object)

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

// strings module Contains, ContainsAny, Count, EqualFold, HasPrefix, HasSuffix
// Index, IndexAny, LastIndex, LastIndexAny, Trim, TrimLeft, TrimPrefix,
// TrimRight, TrimSuffix
//
//ugo:callable func(s1 string, s2 string) (ret ugo.Object)

// strings module Fields, Title, ToLower, ToTitle, ToUpper, TrimSpace
//
//ugo:callable func(s string) (ret ugo.Object)

// strings module ContainsChar, IndexByte, IndexChar, LastIndexByte
//
//ugo:callable func(s string, r rune) (ret ugo.Object)

// strings module Join
//
//ugo:callable func(arr ugo.Array, s string) (ret ugo.Object)

// misc. functions
//
//ugo:callable func(o ugo.Object, i int64) (ret ugo.Object, err error)
