// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Package time provides time module for measuring and displaying time for uGO
// script language. It wraps Go's time package functionalities.
// Note that: uGO's int values are converted to Go's time.Duration values.
package time

import (
	"strconv"
	"time"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/stdlib"
)

var utcLoc ugo.Object = &Location{Value: time.UTC}
var localLoc ugo.Object = &Location{Value: time.Local}

// Module represents time module.
var Module = map[string]ugo.Object{
	// ugo:doc
	// # time Module
	//
	// ## Constants
	// ### Months
	//
	// January
	// February
	// March
	// April
	// May
	// June
	// July
	// August
	// September
	// October
	// November
	// December
	"January":   ugo.Int(time.January),
	"February":  ugo.Int(time.February),
	"March":     ugo.Int(time.March),
	"April":     ugo.Int(time.April),
	"May":       ugo.Int(time.May),
	"June":      ugo.Int(time.June),
	"July":      ugo.Int(time.July),
	"August":    ugo.Int(time.August),
	"September": ugo.Int(time.September),
	"October":   ugo.Int(time.October),
	"November":  ugo.Int(time.November),
	"December":  ugo.Int(time.December),

	// ugo:doc
	// ### Weekdays
	//
	// Sunday
	// Monday
	// Tuesday
	// Wednesday
	// Thursday
	// Friday
	// Saturday
	"Sunday":    ugo.Int(time.Sunday),
	"Monday":    ugo.Int(time.Monday),
	"Tuesday":   ugo.Int(time.Tuesday),
	"Wednesday": ugo.Int(time.Wednesday),
	"Thursday":  ugo.Int(time.Thursday),
	"Friday":    ugo.Int(time.Friday),
	"Saturday":  ugo.Int(time.Saturday),

	// ugo:doc
	// ### Layouts
	//
	// ANSIC
	// UnixDate
	// RubyDate
	// RFC822
	// RFC822Z
	// RFC850
	// RFC1123
	// RFC1123Z
	// RFC3339
	// RFC3339Nano
	// Kitchen
	// Stamp
	// StampMilli
	// StampMicro
	// StampNano
	"ANSIC":       ugo.String(time.ANSIC),
	"UnixDate":    ugo.String(time.UnixDate),
	"RubyDate":    ugo.String(time.RubyDate),
	"RFC822":      ugo.String(time.RFC822),
	"RFC822Z":     ugo.String(time.RFC822Z),
	"RFC850":      ugo.String(time.RFC850),
	"RFC1123":     ugo.String(time.RFC1123),
	"RFC1123Z":    ugo.String(time.RFC1123Z),
	"RFC3339":     ugo.String(time.RFC3339),
	"RFC3339Nano": ugo.String(time.RFC3339Nano),
	"Kitchen":     ugo.String(time.Kitchen),
	"Stamp":       ugo.String(time.Stamp),
	"StampMilli":  ugo.String(time.StampMilli),
	"StampMicro":  ugo.String(time.StampMicro),
	"StampNano":   ugo.String(time.StampNano),

	// ugo:doc
	// ### Durations
	//
	// Nanosecond
	// Microsecond
	// Millisecond
	// Second
	// Minute
	// Hour
	"Nanosecond":  ugo.Int(time.Nanosecond),
	"Microsecond": ugo.Int(time.Microsecond),
	"Millisecond": ugo.Int(time.Millisecond),
	"Second":      ugo.Int(time.Second),
	"Minute":      ugo.Int(time.Minute),
	"Hour":        ugo.Int(time.Hour),

	// ugo:doc
	// ## Functions
	// UTC() -> location
	// Returns Universal Coordinated Time (UTC) location.
	"UTC": &ugo.Function{
		Name: "UTC",
		Value: stdlib.FuncPRO(func() ugo.Object {
			return utcLoc
		}),
	},

	// ugo:doc
	// Local() -> location
	// Returns Local return the system's local time zone location.
	"Local": &ugo.Function{
		Name: "Local",
		Value: stdlib.FuncPRO(func() ugo.Object {
			return localLoc
		}),
	},

	// ugo:doc
	// MonthString(m int) -> month string
	// Returns English name of the month m ("January", "February", ...).
	"MonthString": &ugo.Function{
		Name: "MonthString",
		Value: stdlib.FuncPiRO(func(m int) ugo.Object {
			return ugo.String(time.Month(m).String())
		}),
	},

	// ugo:doc
	// WeekdayString(w int) -> weekday string
	// Returns English name of the int weekday w, note that 0 is Sunday.
	"WeekdayString": &ugo.Function{
		Name: "WeekdayString",
		Value: stdlib.FuncPiRO(func(w int) ugo.Object {
			return ugo.String(time.Weekday(w).String())
		}),
	},

	// ugo:doc
	// DurationString(d int) -> string
	// Returns a string representing the duration d in the form "72h3m0.5s".
	"DurationString": &ugo.Function{
		Name: "DurationString",
		Value: stdlib.FuncPi64RO(func(d int64) ugo.Object {
			return ugo.String(time.Duration(d).String())
		}),
	},
	// ugo:doc
	// DurationNanoseconds(d int) -> int
	// Returns the duration d as an int nanosecond count.
	"DurationNanoseconds": &ugo.Function{
		Name: "DurationNanoseconds",
		Value: stdlib.FuncPi64RO(func(d int64) ugo.Object {
			return ugo.Int(time.Duration(d).Nanoseconds())
		}),
	},
	// ugo:doc
	// DurationMicroseconds(d int) -> int
	// Returns the duration d as an int microsecond count.
	"DurationMicroseconds": &ugo.Function{
		Name: "DurationMicroseconds",
		Value: stdlib.FuncPi64RO(func(d int64) ugo.Object {
			return ugo.Int(time.Duration(d).Microseconds())
		}),
	},
	// ugo:doc
	// DurationMilliseconds(d int) -> int
	// Returns the duration d as an int millisecond count.
	"DurationMilliseconds": &ugo.Function{
		Name: "DurationMilliseconds",
		Value: stdlib.FuncPi64RO(func(d int64) ugo.Object {
			return ugo.Int(time.Duration(d).Milliseconds())
		}),
	},
	// ugo:doc
	// DurationSeconds(d int) -> float
	// Returns the duration d as a floating point number of seconds.
	"DurationSeconds": &ugo.Function{
		Name: "DurationSeconds",
		Value: stdlib.FuncPi64RO(func(d int64) ugo.Object {
			return ugo.Float(time.Duration(d).Seconds())
		}),
	},
	// ugo:doc
	// DurationMinutes(d int) -> float
	// Returns the duration d as a floating point number of minutes.
	"DurationMinutes": &ugo.Function{
		Name: "DurationMinutes",
		Value: stdlib.FuncPi64RO(func(d int64) ugo.Object {
			return ugo.Float(time.Duration(d).Minutes())
		}),
	},
	// ugo:doc
	// DurationHours(d int) -> float
	// Returns the duration d as a floating point number of hours.
	"DurationHours": &ugo.Function{
		Name: "DurationHours",
		Value: stdlib.FuncPi64RO(func(d int64) ugo.Object {
			return ugo.Float(time.Duration(d).Hours())
		}),
	},
	// ugo:doc
	// Sleep(duration int) -> undefined
	// Pauses the current goroutine for at least the duration.
	"Sleep": &ugo.Function{
		Name: "Sleep",
		Value: stdlib.FuncPi64R(func(duration int64) {
			time.Sleep(time.Duration(duration))
		}),
	},
	// ugo:doc
	// ParseDuration(s string) -> duration int
	// Parses duration s and returns duration as int or error.
	"ParseDuration": &ugo.Function{
		Name: "ParseDuration",
		Value: stdlib.FuncPsROe(func(s string) (ugo.Object, error) {
			d, err := time.ParseDuration(s)
			if err != nil {
				return nil, err
			}
			return ugo.Int(d), nil
		}),
	},
	// ugo:doc
	// DurationRound(duration int, m int) -> duration int
	// Returns the result of rounding duration to the nearest multiple of m.
	"DurationRound": &ugo.Function{
		Name: "DurationRound",
		Value: stdlib.FuncPi64i64RO(func(d, m int64) ugo.Object {
			return ugo.Int(time.Duration(d).Round(time.Duration(m)))
		}),
	},
	// ugo:doc
	// DurationTruncate(duration int, m int) -> duration int
	// Returns the result of rounding duration toward zero to a multiple of m.
	"DurationTruncate": &ugo.Function{
		Name: "DurationTruncate",
		Value: stdlib.FuncPi64i64RO(func(d, m int64) ugo.Object {
			return ugo.Int(time.Duration(d).Truncate(time.Duration(m)))
		}),
	},
	// ugo:doc
	// FixedZone(name string, sec int) -> location
	// Returns a Location that always uses the given zone name and offset
	// (seconds east of UTC).
	"FixedZone": &ugo.Function{
		Name: "FixedZone",
		Value: stdlib.FuncPsiRO(func(name string, sec int) ugo.Object {
			return &Location{Value: time.FixedZone(name, sec)}
		}),
	},
	// ugo:doc
	// LoadLocation(name string) -> location
	// Returns the Location with the given name.
	"LoadLocation": &ugo.Function{
		Name: "LoadLocation",
		Value: stdlib.FuncPsROe(func(name string) (ugo.Object, error) {
			l, err := time.LoadLocation(name)
			if err != nil {
				return nil, err
			}
			return &Location{Value: l}, nil
		}),
	},
	// ugo:doc
	// IsLocation(any) -> bool
	// Reports whether any value is of location type.
	"IsLocation": &ugo.Function{
		Name: "IsLocation",
		Value: stdlib.FuncPORO(func(o ugo.Object) ugo.Object {
			_, ok := o.(*Location)
			return ugo.Bool(ok)
		}),
	},
	// ugo:doc
	// Time() -> time
	// Returns zero time.
	"Time": &ugo.Function{
		Name: "Time",
		Value: stdlib.FuncPRO(func() ugo.Object {
			return &Time{Value: time.Time{}}
		}),
	},
	// ugo:doc
	// Since(t time) -> duration int
	// Returns the time elapsed since t.
	"Since": &ugo.Function{
		Name: "Since",
		Value: funcPTRO(func(t *Time) ugo.Object {
			return ugo.Int(time.Since(t.Value))
		}),
	},
	// ugo:doc
	// Until(t time) -> duration int
	// Returns the duration until t.
	"Until": &ugo.Function{
		Name: "Until",
		Value: funcPTRO(func(t *Time) ugo.Object {
			return ugo.Int(time.Until(t.Value))
		}),
	},
	// ugo:doc
	// Date(year int, month int, day int[, hour int, min int, sec int, nsec int, loc location]) -> time
	// Returns the Time corresponding to yyyy-mm-dd hh:mm:ss + nsec nanoseconds
	// in the appropriate zone for that time in the given location. Zero values
	// of optional arguments are used if not provided.
	"Date": &ugo.Function{
		Name:  "Date",
		Value: date,
	},
	// ugo:doc
	// Now() -> time
	// Returns the current local time.
	"Now": &ugo.Function{
		Name: "Now",
		Value: stdlib.FuncPRO(func() ugo.Object {
			return &Time{Value: time.Now()}
		}),
	},
	// ugo:doc
	// Parse(layout string, value string[, loc location]) -> time
	// Parses a formatted string and returns the time value it represents.
	// If location is not provided, Go's `time.Parse` function is called
	// otherwise `time.ParseInLocation` is called.
	"Parse": &ugo.Function{
		Name:  "Parse",
		Value: parse,
	},
	// ugo:doc
	// Unix(sec int[, nsec int]) -> time
	// Returns the local time corresponding to the given Unix time,
	// sec seconds and nsec nanoseconds since January 1, 1970 UTC.
	// Zero values of optional arguments are used if not provided.
	"Unix": &ugo.Function{
		Name:  "Unix",
		Value: unix,
	},
	// ugo:doc
	// Deprecated: Use .Add method of time object.
	// Add(t time, duration int) -> time
	// Returns the time of t+duration.
	"Add": &ugo.Function{
		Name:  "Add",
		Value: funcPTi64RO(timeAdd),
	},
	// ugo:doc
	// Deprecated: Use .Sub method of time object.
	// Sub(t1 time, t2 time) -> int
	// Returns the duration of t1-t2.
	"Sub": &ugo.Function{
		Name:  "Sub",
		Value: funcPTTRO(timeSub),
	},
	// ugo:doc
	// Deprecated: Use .AddDate method of time object.
	// AddDate(t time, years int, months int, days int) -> time
	// Returns the time corresponding to adding the given number of
	// years, months, and days to t.
	"AddDate": &ugo.Function{
		Name:  "AddDate",
		Value: funcPTiiiRO(timeAddDate),
	},
	// ugo:doc
	// Deprecated: Use .After method of time object.
	// After(t1 time, t2 time) -> bool
	// Reports whether the time t1 is after t2.
	"After": &ugo.Function{
		Name:  "After",
		Value: funcPTTRO(timeAfter),
	},
	// ugo:doc
	// Deprecated: Use .Before method of time object.
	// Before(t1 time, t2 time) -> bool
	// Reports whether the time t1 is before t2.
	"Before": &ugo.Function{
		Name:  "Before",
		Value: funcPTTRO(timeBefore),
	},
	// ugo:doc
	// Deprecated: Use .Format method of time object.
	// Format(t time, layout string) -> string
	// Returns a textual representation of the time value formatted according
	// to layout.
	"Format": &ugo.Function{
		Name:  "Format",
		Value: funcPTsRO(timeFormat),
	},
	// ugo:doc
	// Deprecated: Use .AppendFormat method of time object.
	// AppendFormat(t time, b bytes, layout string) -> bytes
	// It is like `Format` but appends the textual representation to b and
	// returns the extended buffer.
	"AppendFormat": &ugo.Function{
		Name:  "AppendFormat", // funcPTb2sRO
		Value: funcPTb2sRO(timeAppendFormat),
	},
	// ugo:doc
	// Deprecated: Use .In method of time object.
	// In(t time, loc location) -> time
	// Returns a copy of t representing the same time t, but with the copy's
	// location information set to loc for display purposes.
	"In": &ugo.Function{
		Name:  "In",
		Value: funcPTLRO(timeIn),
	},
	// ugo:doc
	// Deprecated: Use .Round method of time object.
	// Round(t time, duration int) -> time
	// Round returns the result of rounding t to the nearest multiple of
	// duration.
	"Round": &ugo.Function{
		Name:  "Round",
		Value: funcPTi64RO(timeRound),
	},
	// ugo:doc
	// Deprecated: Use .Truncate method of time object.
	// Truncate(t time, duration int) -> time
	// Truncate returns the result of rounding t down to a multiple of duration.
	"Truncate": &ugo.Function{
		Name:  "Truncate",
		Value: funcPTi64RO(timeTruncate),
	},
	// ugo:doc
	// IsTime(any) -> bool
	// Reports whether any value is of time type.
	"IsTime": &ugo.Function{
		Name: "IsTime",
		Value: stdlib.FuncPORO(func(o ugo.Object) ugo.Object {
			_, ok := o.(*Time)
			return ugo.Bool(ok)
		}),
	},
}

func loadLocation(args ...ugo.Object) (ugo.Object, error) {
	name, ok := args[0].(ugo.String)
	if !ok {
		return newArgTypeErr("1st", "string", args[0].TypeName())
	}
	l, err := time.LoadLocation(string(name))
	if err != nil {
		return ugo.Undefined, err
	}
	return &Location{Value: l}, nil
}

func date(args ...ugo.Object) (ugo.Object, error) {
	if len(args) < 3 || len(args) > 8 {
		return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
			"want=3..8 got=" + strconv.Itoa(len(args)))
	}
	ymdHmsn := [7]int{}
	var loc = &Location{Value: time.Local}
	var ok bool
	for i := 0; i < len(args); i++ {
		if i < 7 {
			v, ok := args[i].(ugo.Int)
			if !ok {
				return nil, ugo.NewArgumentTypeError(
					strconv.Itoa(i+1), "int", args[i].TypeName())
			}
			ymdHmsn[i] = int(v)
			continue
		}
		loc, ok = args[i].(*Location)
		if !ok {
			return nil, ugo.NewArgumentTypeError(
				strconv.Itoa(i+1), "location", args[i].TypeName())
		}
	}

	tm := time.Date(ymdHmsn[0], time.Month(ymdHmsn[1]), ymdHmsn[2],
		ymdHmsn[3], ymdHmsn[4], ymdHmsn[5], ymdHmsn[6], loc.Value)
	return &Time{Value: tm}, nil
}

func parse(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 && len(args) != 3 {
		return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
			"want=2..3 got=" + strconv.Itoa(len(args)))
	}
	layout, ok := ugo.ToGoString(args[0])
	if !ok {
		return newArgTypeErr("1st", "string", args[0].TypeName())
	}
	value, ok := ugo.ToGoString(args[1])
	if !ok {
		return newArgTypeErr("2nd", "string", args[1].TypeName())
	}
	if len(args) == 2 {
		tm, err := time.Parse(layout, value)
		if err != nil {
			return ugo.Undefined, err
		}
		return &Time{Value: tm}, nil
	}
	loc, ok := ToLocation(args[2])
	if !ok {
		return newArgTypeErr("3rd", "location", args[2].TypeName())
	}
	tm, err := time.ParseInLocation(layout, value, loc.Value)
	if err != nil {
		return ugo.Undefined, err
	}
	return &Time{Value: tm}, nil
}

func unix(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 1 && len(args) != 2 {
		return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
			"want=1..2 got=" + strconv.Itoa(len(args)))
	}

	sec, ok := ugo.ToGoInt64(args[0])
	if !ok {
		return newArgTypeErr("1st", "int", args[0].TypeName())
	}

	var nsec int64
	if len(args) > 1 {
		nsec, ok = ugo.ToGoInt64(args[1])
		if !ok {
			return newArgTypeErr("2nd", "int", args[1].TypeName())
		}
	}
	return &Time{Value: time.Unix(sec, nsec)}, nil
}
