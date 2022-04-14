// Copyright (c) 2020-2022 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Package time provides time module for measuring and displaying time for uGO
// script language. It wraps Go's time package functionalities.
// Note that: uGO's int values are converted to Go's time.Duration values.
//
package time

import (
	"encoding/gob"
	"strconv"
	"time"

	"github.com/ozanh/ugo"
)

func init() {
	gob.Register(&Time{})
	gob.Register(&Location{})
}

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
	// MonthString(m int) -> month string
	// Returns English name of the month m ("January", "February", ...).
	"MonthString": &ugo.Function{
		Name:  "MonthString",
		Value: want1(monthString),
	},

	// ugo:doc
	// WeekdayString(w int) -> weekday string
	// Returns English name of the int weekday w, note that 0 is Sunday.
	"WeekdayString": &ugo.Function{
		Name:  "WeekdayString",
		Value: want1(weekdayString),
	},

	// ugo:doc
	// DurationString(d int) -> string
	// Returns a string representing the duration d in the form "72h3m0.5s".
	"DurationString": &ugo.Function{
		Name:  "DurationString",
		Value: want1(durationString),
	},
	// ugo:doc
	// DurationNanoseconds(d int) -> int
	// Returns the duration d as an int nanosecond count.
	"DurationNanoseconds": &ugo.Function{
		Name:  "DurationNanoseconds",
		Value: want1(durationNanoseconds),
	},
	// ugo:doc
	// DurationMicroseconds(d int) -> int
	// Returns the duration d as an int microsecond count.
	"DurationMicroseconds": &ugo.Function{
		Name:  "DurationMicroseconds",
		Value: want1(durationMicroseconds),
	},
	// ugo:doc
	// DurationMilliseconds(d int) -> int
	// Returns the duration d as an int millisecond count.
	"DurationMilliseconds": &ugo.Function{
		Name:  "DurationMilliseconds",
		Value: want1(durationMilliseconds),
	},
	// ugo:doc
	// DurationSeconds(d int) -> float
	// Returns the duration d as a floating point number of seconds.
	"DurationSeconds": &ugo.Function{
		Name:  "DurationSeconds",
		Value: want1(durationSeconds),
	},
	// ugo:doc
	// DurationMinutes(d int) -> float
	// Returns the duration d as a floating point number of minutes.
	"DurationMinutes": &ugo.Function{
		Name:  "DurationMinutes",
		Value: want1(durationMinutes),
	},
	// ugo:doc
	// DurationHours(d int) -> float
	// Returns the duration d as a floating point number of hours.
	"DurationHours": &ugo.Function{
		Name:  "DurationHours",
		Value: want1(durationHours),
	},
	// ugo:doc
	// Sleep(duration int) -> undefined
	// Pauses the current goroutine for at least the duration.
	"Sleep": &ugo.Function{
		Name:  "Sleep",
		Value: want1(sleep),
	},
	// ugo:doc
	// ParseDuration(s string) -> duration int
	// Parses duration s and returns duration as int.
	"ParseDuration": &ugo.Function{
		Name:  "ParseDuration",
		Value: want1(parseDuration),
	},
	// ugo:doc
	// DurationRound(duration int, m int) -> duration int
	// Returns the result of rounding duration to the nearest multiple of m.
	"DurationRound": &ugo.Function{
		Name:  "DurationRound",
		Value: want2(durationRound),
	},
	// ugo:doc
	// DurationTruncate(duration int, m int) -> duration int
	// Returns the result of rounding duration toward zero to a multiple of m.
	"DurationTruncate": &ugo.Function{
		Name:  "DurationTruncate",
		Value: want2(durationTruncate),
	},
	// ugo:doc
	// FixedZone(name string, sec int) -> location
	// Returns a Location that always uses the given zone name and offset
	// (seconds east of UTC).
	"FixedZone": &ugo.Function{
		Name:  "FixedZone",
		Value: want2(fixedZone),
	},
	// ugo:doc
	// LoadLocation(name string) -> location
	// Returns the Location with the given name.
	"LoadLocation": &ugo.Function{
		Name:  "LoadLocation",
		Value: want1(loadLocation),
	},
	// ugo:doc
	// IsLocation(any) -> bool
	// Reports whether any value is of location type.
	"IsLocation": &ugo.Function{
		Name:  "IsLocation",
		Value: want1(isLocation),
	},
	// ugo:doc
	// Time() -> time
	// Returns zero time.
	"Time": &ugo.Function{
		Name:  "Time",
		Value: timeTime,
	},
	// ugo:doc
	// Since(t time) -> duration int
	// Returns the time elapsed since t.
	"Since": &ugo.Function{
		Name:  "Since",
		Value: want1(timeSince),
	},
	// ugo:doc
	// Until(t time) -> duration int
	// Returns the duration until t.
	"Until": &ugo.Function{
		Name:  "Until",
		Value: want1(timeUntil),
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
		Name:  "Now",
		Value: now,
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
	// Add(t time, duration int) -> time
	// Returns the time of t+duration.
	"Add": &ugo.Function{
		Name:  "Add",
		Value: want2(add),
	},
	// ugo:doc
	// Sub(t1 time, t2 time) -> int
	// Returns the duration of t1-t2.
	"Sub": &ugo.Function{
		Name:  "Sub",
		Value: want2(sub),
	},
	// ugo:doc
	// AddDate(t time, years int, months int, days int) -> time
	// Returns the time corresponding to adding the given number of
	// years, months, and days to t.
	"AddDate": &ugo.Function{
		Name:  "AddDate",
		Value: addDate,
	},
	// ugo:doc
	// After(t1 time, t2 time) -> bool
	// Reports whether the time t1 is after t2.
	"After": &ugo.Function{
		Name:  "After",
		Value: want2(after),
	},
	// ugo:doc
	// Before(t1 time, t2 time) -> bool
	// Reports whether the time t1 is before t2.
	"Before": &ugo.Function{
		Name:  "Before",
		Value: want2(before),
	},
	// ugo:doc
	// Format(t time, layout string) -> string
	// Returns a textual representation of the time value formatted according
	// to layout.
	"Format": &ugo.Function{
		Name:  "Format",
		Value: want2(format),
	},
	// ugo:doc
	// AppendFormat(t time, b bytes, layout string) -> bytes
	// It is like `Format` but appends the textual representation to b and
	// returns the extended buffer.
	"AppendFormat": &ugo.Function{
		Name:  "AppendFormat",
		Value: appendFormat,
	},
	// ugo:doc
	// In(t time, loc location) -> time
	// Returns a copy of t representing the same time t, but with the copy's
	// location information set to loc for display purposes.
	"In": &ugo.Function{
		Name:  "In",
		Value: want2(timeIn),
	},
	// ugo:doc
	// Round(t time, duration int) -> time
	// Round returns the result of rounding t to the nearest multiple of
	// duration.
	"Round": &ugo.Function{
		Name:  "Round",
		Value: want2(timeRound),
	},
	// ugo:doc
	// Truncate(t time, duration int) -> time
	// Truncate returns the result of rounding t down to a multiple of duration.
	"Truncate": &ugo.Function{
		Name:  "Truncate",
		Value: want2(timeTruncate),
	},
	// ugo:doc
	// IsTime(any) -> bool
	// Reports whether any value is of time type.
	"IsTime": &ugo.Function{
		Name:  "IsTime",
		Value: want1(isTime),
	},
}

func want1(fn ugo.CallableFunc) ugo.CallableFunc {
	return func(args ...ugo.Object) (ret ugo.Object, err error) {
		if len(args) != 1 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				wantEqXGotY(1, len(args)))
		}
		return fn(args...)
	}
}

func want2(fn ugo.CallableFunc) ugo.CallableFunc {
	return func(args ...ugo.Object) (ret ugo.Object, err error) {
		if len(args) != 2 {
			return nil, ugo.ErrWrongNumArguments.NewError(
				wantEqXGotY(2, len(args)))
		}
		return fn(args...)
	}
}

func durationString(args ...ugo.Object) (ugo.Object, error) {
	d, ok := args[0].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}
	return ugo.String(time.Duration(d).String()), nil
}

func durationNanoseconds(args ...ugo.Object) (ugo.Object, error) {
	d, ok := args[0].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}
	return ugo.Int(time.Duration(d).Nanoseconds()), nil
}

func durationMicroseconds(args ...ugo.Object) (ugo.Object, error) {
	d, ok := args[0].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}
	return ugo.Int(time.Duration(d).Microseconds()), nil
}

func durationMilliseconds(args ...ugo.Object) (ugo.Object, error) {
	d, ok := args[0].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}
	return ugo.Int(time.Duration(d).Milliseconds()), nil
}

func durationSeconds(args ...ugo.Object) (ugo.Object, error) {
	d, ok := args[0].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}
	return ugo.Float(time.Duration(d).Seconds()), nil
}

func durationMinutes(args ...ugo.Object) (ugo.Object, error) {
	d, ok := args[0].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}
	return ugo.Float(time.Duration(d).Minutes()), nil
}

func durationHours(args ...ugo.Object) (ugo.Object, error) {
	d, ok := args[0].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}
	return ugo.Float(time.Duration(d).Hours()), nil
}

func sleep(args ...ugo.Object) (ugo.Object, error) {
	switch v := args[0].(type) {
	case ugo.Int:
		time.Sleep(time.Duration(v))
	default:
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}
	return ugo.Undefined, nil
}

func parseDuration(args ...ugo.Object) (ugo.Object, error) {
	v, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	d, err := time.ParseDuration(string(v))
	if err != nil {
		return nil, err
	}
	return ugo.Int(d), nil
}

func timeTime(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 0 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(0, len(args)))
	}
	return &Time{Value: time.Time{}}, nil
}

func timeSince(args ...ugo.Object) (ugo.Object, error) {
	v, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	return ugo.Int(time.Since(v.Value)), nil
}

func timeUntil(args ...ugo.Object) (ugo.Object, error) {
	v, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	return ugo.Int(time.Until(v.Value)), nil
}

func durationRound(args ...ugo.Object) (ugo.Object, error) {
	switch d := args[0].(type) {
	case ugo.Int:
		switch m := args[1].(type) {
		case ugo.Int:
			return ugo.Int(time.Duration(d).Round(time.Duration(m))), nil
		}
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
	return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
}

func durationTruncate(args ...ugo.Object) (ugo.Object, error) {
	switch d := args[0].(type) {
	case ugo.Int:
		switch m := args[1].(type) {
		case ugo.Int:
			return ugo.Int(time.Duration(d).Truncate(time.Duration(m))), nil
		}
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
	return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
}

func monthString(args ...ugo.Object) (ugo.Object, error) {
	if v, ok := args[0].(ugo.Int); ok {
		return ugo.String(time.Month(v).String()), nil
	}
	return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
}

func weekdayString(args ...ugo.Object) (ugo.Object, error) {
	if v, ok := args[0].(ugo.Int); ok {
		return ugo.String(time.Weekday(v).String()), nil
	}
	return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
}

func fixedZone(args ...ugo.Object) (ugo.Object, error) {
	name, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	var offset int
	switch v := args[1].(type) {
	case ugo.Int:
		offset = int(v)
	default:
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
	l := time.FixedZone(string(name), offset)
	return &Location{Location: l}, nil
}

func loadLocation(args ...ugo.Object) (ugo.Object, error) {
	name, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	l, err := time.LoadLocation(string(name))
	if err != nil {
		return nil, err
	}
	return &Location{Location: l}, nil
}

func isLocation(args ...ugo.Object) (ugo.Object, error) {
	_, ok := args[0].(*Location)
	return ugo.Bool(ok), nil
}

func date(args ...ugo.Object) (ugo.Object, error) {
	if len(args) < 3 || len(args) > 8 {
		return nil, ugo.ErrWrongNumArguments.NewError(
			"want=3..8 got=" + strconv.Itoa(len(args)))
	}
	ymdHmsn := [7]int{}
	var loc = &Location{Location: time.Local}
	var ok bool
	for i := 0; i < len(args); i++ {
		if i < 7 {
			v, ok := args[i].(ugo.Int)
			if !ok {
				return nil, ugo.NewArgumentTypeError(
					strconv.Itoa(i+1),
					"int",
					args[i].TypeName(),
				)
			}
			ymdHmsn[i] = int(v)
			continue
		}
		loc, ok = args[i].(*Location)
		if !ok {
			return nil, ugo.NewArgumentTypeError(
				strconv.Itoa(i+1),
				"location",
				args[i].TypeName(),
			)
		}
	}

	tm := time.Date(ymdHmsn[0], time.Month(ymdHmsn[1]), ymdHmsn[2],
		ymdHmsn[3], ymdHmsn[4], ymdHmsn[5], ymdHmsn[6], loc.Location)
	return &Time{Value: tm}, nil
}

func now(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 0 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(0, len(args)))
	}
	return &Time{Value: time.Now()}, nil
}

func parse(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, ugo.ErrWrongNumArguments.NewError("want=2..3 got=" +
			strconv.Itoa(len(args)))
	}
	layout, ok := args[0].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "string",
			args[0].TypeName())
	}
	value, ok := args[1].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "string",
			args[1].TypeName())
	}
	if len(args) == 2 {
		tm, err := time.Parse(string(layout), string(value))
		if err != nil {
			return nil, err
		}
		return &Time{Value: tm}, nil
	}
	loc, ok := args[2].(*Location)
	if !ok {
		return nil, ugo.NewArgumentTypeError("third", "location",
			args[2].TypeName())
	}
	tm, err := time.ParseInLocation(string(layout), string(value), loc.Location)
	if err != nil {
		return nil, err
	}
	return &Time{Value: tm}, nil
}

func unix(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, ugo.ErrWrongNumArguments.NewError("want=1..2 got=" +
			strconv.Itoa(len(args)))
	}

	var sec int64
	switch v := args[0].(type) {
	case ugo.Int:
		sec = int64(v)
	default:
		return nil, ugo.NewArgumentTypeError("first", "int", args[0].TypeName())
	}

	var nsec int64
	if len(args) > 1 {
		switch v := args[1].(type) {
		case ugo.Int:
			nsec = int64(v)
		default:
			return nil, ugo.NewArgumentTypeError("second", "int",
				args[1].TypeName())
		}
	}
	return &Time{Value: time.Unix(sec, nsec)}, nil
}

func add(args ...ugo.Object) (ugo.Object, error) {
	tm, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	switch v := args[1].(type) {
	case ugo.Int:
		return &Time{Value: tm.Value.Add(time.Duration(v))}, nil
	}
	return nil, ugo.NewArgumentTypeError("second", "int",
		args[1].TypeName())
}

func sub(args ...ugo.Object) (ugo.Object, error) {
	tm1, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	tm2, ok := args[1].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "time",
			args[1].TypeName())
	}
	return ugo.Int(tm1.Value.Sub(tm2.Value)), nil
}

func addDate(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 4 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(4, len(args)))
	}
	tm, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	years, ok := args[1].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
	months, ok := args[2].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("third", "int",
			args[2].TypeName())
	}
	days, ok := args[3].(ugo.Int)
	if !ok {
		return nil, ugo.NewArgumentTypeError("fourth", "int",
			args[3].TypeName())
	}
	return &Time{
		Value: tm.Value.AddDate(int(years), int(months), int(days)),
	}, nil
}

func after(args ...ugo.Object) (ugo.Object, error) {
	tm1, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	tm2, ok := args[1].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "time",
			args[1].TypeName())
	}
	return ugo.Bool(tm1.Value.After(tm2.Value)), nil
}

func before(args ...ugo.Object) (ugo.Object, error) {
	tm1, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	tm2, ok := args[1].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "time",
			args[1].TypeName())
	}
	return ugo.Bool(tm1.Value.Before(tm2.Value)), nil
}

func appendFormat(args ...ugo.Object) (ugo.Object, error) {
	if len(args) != 3 {
		return nil, ugo.ErrWrongNumArguments.NewError(wantEqXGotY(3, len(args)))
	}
	tm, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	b, ok := args[1].(ugo.Bytes)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "bytes",
			args[1].TypeName())
	}
	layout, ok := args[2].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("third", "string",
			args[2].TypeName())
	}
	return ugo.Bytes(tm.Value.AppendFormat(b, string(layout))), nil
}

func format(args ...ugo.Object) (ugo.Object, error) {
	tm, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	layout, ok := args[1].(ugo.String)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "string",
			args[1].TypeName())
	}
	return ugo.String(tm.Value.Format(string(layout))), nil
}

func timeIn(args ...ugo.Object) (ugo.Object, error) {
	tm, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	loc, ok := args[1].(*Location)
	if !ok {
		return nil, ugo.NewArgumentTypeError("second", "location",
			args[1].TypeName())
	}
	return &Time{Value: tm.Value.In(loc.Location)}, nil
}

func timeRound(args ...ugo.Object) (ugo.Object, error) {
	tm, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	switch v := args[1].(type) {
	case ugo.Int:
		return &Time{Value: tm.Value.Round(time.Duration(v))}, nil
	default:
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
}

func timeTruncate(args ...ugo.Object) (ugo.Object, error) {
	tm, ok := args[0].(*Time)
	if !ok {
		return nil, ugo.NewArgumentTypeError("first", "time",
			args[0].TypeName())
	}
	switch v := args[1].(type) {
	case ugo.Int:
		return &Time{Value: tm.Value.Truncate(time.Duration(v))}, nil
	default:
		return nil, ugo.NewArgumentTypeError("second", "int",
			args[1].TypeName())
	}
}

func isTime(args ...ugo.Object) (ugo.Object, error) {
	_, ok := args[0].(*Time)
	return ugo.Bool(ok), nil
}

func wantEqXGotY(x, y int) string {
	buf := make([]byte, 0, 20)
	buf = append(buf, "want="...)
	buf = strconv.AppendInt(buf, int64(x), 10)
	buf = append(buf, " got="...)
	buf = strconv.AppendInt(buf, int64(y), 10)
	return string(buf)
}
