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
var zeroTime ugo.Object = &Time{}

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
		Name:    "UTC",
		Value:   stdlib.FuncPRO(utcFunc),
		ValueEx: stdlib.FuncPROEx(utcFunc),
	},

	// ugo:doc
	// Local() -> location
	// Returns the system's local time zone location.
	"Local": &ugo.Function{
		Name:    "Local",
		Value:   stdlib.FuncPRO(localFunc),
		ValueEx: stdlib.FuncPROEx(localFunc),
	},

	// ugo:doc
	// MonthString(m int) -> month string
	// Returns English name of the month m ("January", "February", ...).
	"MonthString": &ugo.Function{
		Name:    "MonthString",
		Value:   stdlib.FuncPiRO(monthStringFunc),
		ValueEx: stdlib.FuncPiROEx(monthStringFunc),
	},

	// ugo:doc
	// WeekdayString(w int) -> weekday string
	// Returns English name of the int weekday w, note that 0 is Sunday.
	"WeekdayString": &ugo.Function{
		Name:    "WeekdayString",
		Value:   stdlib.FuncPiRO(weekdayStringFunc),
		ValueEx: stdlib.FuncPiROEx(weekdayStringFunc),
	},

	// ugo:doc
	// DurationString(d int) -> string
	// Returns a string representing the duration d in the form "72h3m0.5s".
	"DurationString": &ugo.Function{
		Name:    "DurationString",
		Value:   stdlib.FuncPi64RO(durationStringFunc),
		ValueEx: stdlib.FuncPi64ROEx(durationStringFunc),
	},
	// ugo:doc
	// DurationNanoseconds(d int) -> int
	// Returns the duration d as an int nanosecond count.
	"DurationNanoseconds": &ugo.Function{
		Name:    "DurationNanoseconds",
		Value:   stdlib.FuncPi64RO(durationNanosecondsFunc),
		ValueEx: stdlib.FuncPi64ROEx(durationNanosecondsFunc),
	},
	// ugo:doc
	// DurationMicroseconds(d int) -> int
	// Returns the duration d as an int microsecond count.
	"DurationMicroseconds": &ugo.Function{
		Name:    "DurationMicroseconds",
		Value:   stdlib.FuncPi64RO(durationMicrosecondsFunc),
		ValueEx: stdlib.FuncPi64ROEx(durationMicrosecondsFunc),
	},
	// ugo:doc
	// DurationMilliseconds(d int) -> int
	// Returns the duration d as an int millisecond count.
	"DurationMilliseconds": &ugo.Function{
		Name:    "DurationMilliseconds",
		Value:   stdlib.FuncPi64RO(durationMillisecondsFunc),
		ValueEx: stdlib.FuncPi64ROEx(durationMillisecondsFunc),
	},
	// ugo:doc
	// DurationSeconds(d int) -> float
	// Returns the duration d as a floating point number of seconds.
	"DurationSeconds": &ugo.Function{
		Name:    "DurationSeconds",
		Value:   stdlib.FuncPi64RO(durationSecondsFunc),
		ValueEx: stdlib.FuncPi64ROEx(durationSecondsFunc),
	},
	// ugo:doc
	// DurationMinutes(d int) -> float
	// Returns the duration d as a floating point number of minutes.
	"DurationMinutes": &ugo.Function{
		Name:    "DurationMinutes",
		Value:   stdlib.FuncPi64RO(durationMinutesFunc),
		ValueEx: stdlib.FuncPi64ROEx(durationMinutesFunc),
	},
	// ugo:doc
	// DurationHours(d int) -> float
	// Returns the duration d as a floating point number of hours.
	"DurationHours": &ugo.Function{
		Name:    "DurationHours",
		Value:   stdlib.FuncPi64RO(durationHoursFunc),
		ValueEx: stdlib.FuncPi64ROEx(durationHoursFunc),
	},
	// ugo:doc
	// Sleep(duration int) -> undefined
	// Pauses the current goroutine for at least the duration.
	"Sleep": &ugo.Function{
		Name: "Sleep",
		Value: stdlib.FuncPi64R(func(duration int64) {
			time.Sleep(time.Duration(duration))
		}),
		ValueEx: sleepFunc,
	},
	// ugo:doc
	// ParseDuration(s string) -> duration int
	// Parses duration s and returns duration as int or error.
	"ParseDuration": &ugo.Function{
		Name:    "ParseDuration",
		Value:   stdlib.FuncPsROe(parseDurationFunc),
		ValueEx: stdlib.FuncPsROeEx(parseDurationFunc),
	},
	// ugo:doc
	// DurationRound(duration int, m int) -> duration int
	// Returns the result of rounding duration to the nearest multiple of m.
	"DurationRound": &ugo.Function{
		Name:    "DurationRound",
		Value:   stdlib.FuncPi64i64RO(durationRoundFunc),
		ValueEx: stdlib.FuncPi64i64ROEx(durationRoundFunc),
	},
	// ugo:doc
	// DurationTruncate(duration int, m int) -> duration int
	// Returns the result of rounding duration toward zero to a multiple of m.
	"DurationTruncate": &ugo.Function{
		Name:    "DurationTruncate",
		Value:   stdlib.FuncPi64i64RO(durationTruncateFunc),
		ValueEx: stdlib.FuncPi64i64ROEx(durationTruncateFunc),
	},
	// ugo:doc
	// FixedZone(name string, sec int) -> location
	// Returns a Location that always uses the given zone name and offset
	// (seconds east of UTC).
	"FixedZone": &ugo.Function{
		Name:    "FixedZone",
		Value:   stdlib.FuncPsiRO(fixedZoneFunc),
		ValueEx: stdlib.FuncPsiROEx(fixedZoneFunc),
	},
	// ugo:doc
	// LoadLocation(name string) -> location
	// Returns the Location with the given name.
	"LoadLocation": &ugo.Function{
		Name:    "LoadLocation",
		Value:   stdlib.FuncPsROe(loadLocationFunc),
		ValueEx: stdlib.FuncPsROeEx(loadLocationFunc),
	},
	// ugo:doc
	// IsLocation(any) -> bool
	// Reports whether any value is of location type.
	"IsLocation": &ugo.Function{
		Name:    "IsLocation",
		Value:   stdlib.FuncPORO(isLocationFunc),
		ValueEx: stdlib.FuncPOROEx(isLocationFunc),
	},
	// ugo:doc
	// Time() -> time
	// Returns zero time.
	"Time": &ugo.Function{
		Name:    "Time",
		Value:   stdlib.FuncPRO(zerotimeFunc),
		ValueEx: stdlib.FuncPROEx(zerotimeFunc),
	},
	// ugo:doc
	// Since(t time) -> duration int
	// Returns the time elapsed since t.
	"Since": &ugo.Function{
		Name:    "Since",
		Value:   funcPTRO(sinceFunc),
		ValueEx: funcPTROEx(sinceFunc),
	},
	// ugo:doc
	// Until(t time) -> duration int
	// Returns the duration until t.
	"Until": &ugo.Function{
		Name:    "Until",
		Value:   funcPTRO(untilFunc),
		ValueEx: funcPTROEx(untilFunc),
	},
	// ugo:doc
	// Date(year int, month int, day int[, hour int, min int, sec int, nsec int, loc location]) -> time
	// Returns the Time corresponding to yyyy-mm-dd hh:mm:ss + nsec nanoseconds
	// in the appropriate zone for that time in the given location. Zero values
	// of optional arguments are used if not provided.
	"Date": &ugo.Function{
		Name:    "Date",
		Value:   dateFunc,
		ValueEx: dateFuncEx,
	},
	// ugo:doc
	// Now() -> time
	// Returns the current local time.
	"Now": &ugo.Function{
		Name:    "Now",
		Value:   stdlib.FuncPRO(nowFunc),
		ValueEx: stdlib.FuncPROEx(nowFunc),
	},
	// ugo:doc
	// Parse(layout string, value string[, loc location]) -> time
	// Parses a formatted string and returns the time value it represents.
	// If location is not provided, Go's `time.Parse` function is called
	// otherwise `time.ParseInLocation` is called.
	"Parse": &ugo.Function{
		Name:    "Parse",
		Value:   parseFunc,
		ValueEx: parseFuncEx,
	},
	// ugo:doc
	// Unix(sec int[, nsec int]) -> time
	// Returns the local time corresponding to the given Unix time,
	// sec seconds and nsec nanoseconds since January 1, 1970 UTC.
	// Zero values of optional arguments are used if not provided.
	"Unix": &ugo.Function{
		Name:    "Unix",
		Value:   unixFunc,
		ValueEx: unixFuncEx,
	},
	// ugo:doc
	// Add(t time, duration int) -> time
	// Deprecated: Use .Add method of time object.
	// Returns the time of t+duration.
	"Add": &ugo.Function{
		Name:  "Add",
		Value: funcPTi64RO(timeAdd),
	},
	// ugo:doc
	// Sub(t1 time, t2 time) -> int
	// Deprecated: Use .Sub method of time object.
	// Returns the duration of t1-t2.
	"Sub": &ugo.Function{
		Name:  "Sub",
		Value: funcPTTRO(timeSub),
	},
	// ugo:doc
	// AddDate(t time, years int, months int, days int) -> time
	// Deprecated: Use .AddDate method of time object.
	// Returns the time corresponding to adding the given number of
	// years, months, and days to t.
	"AddDate": &ugo.Function{
		Name:  "AddDate",
		Value: funcPTiiiRO(timeAddDate),
	},
	// ugo:doc
	// After(t1 time, t2 time) -> bool
	// Deprecated: Use .After method of time object.
	// Reports whether the time t1 is after t2.
	"After": &ugo.Function{
		Name:  "After",
		Value: funcPTTRO(timeAfter),
	},
	// ugo:doc
	// Before(t1 time, t2 time) -> bool
	// Deprecated: Use .Before method of time object.
	// Reports whether the time t1 is before t2.
	"Before": &ugo.Function{
		Name:  "Before",
		Value: funcPTTRO(timeBefore),
	},
	// ugo:doc
	// Format(t time, layout string) -> string
	// Deprecated: Use .Format method of time object.
	// Returns a textual representation of the time value formatted according
	// to layout.
	"Format": &ugo.Function{
		Name:  "Format",
		Value: funcPTsRO(timeFormat),
	},
	// ugo:doc
	// AppendFormat(t time, b bytes, layout string) -> bytes
	// Deprecated: Use .AppendFormat method of time object.
	// It is like `Format` but appends the textual representation to b and
	// returns the extended buffer.
	"AppendFormat": &ugo.Function{
		Name:  "AppendFormat", // funcPTb2sRO
		Value: funcPTb2sRO(timeAppendFormat),
	},
	// ugo:doc
	// In(t time, loc location) -> time
	// Deprecated: Use .In method of time object.
	// Returns a copy of t representing the same time t, but with the copy's
	// location information set to loc for display purposes.
	"In": &ugo.Function{
		Name:  "In",
		Value: funcPTLRO(timeIn),
	},
	// ugo:doc
	// Round(t time, duration int) -> time
	// Deprecated: Use .Round method of time object.
	// Round returns the result of rounding t to the nearest multiple of
	// duration.
	"Round": &ugo.Function{
		Name:  "Round",
		Value: funcPTi64RO(timeRound),
	},
	// ugo:doc
	// Truncate(t time, duration int) -> time
	// Deprecated: Use .Truncate method of time object.
	// Truncate returns the result of rounding t down to a multiple of duration.
	"Truncate": &ugo.Function{
		Name:  "Truncate",
		Value: funcPTi64RO(timeTruncate),
	},
	// ugo:doc
	// IsTime(any) -> bool
	// Reports whether any value is of time type.
	"IsTime": &ugo.Function{
		Name:    "IsTime",
		Value:   stdlib.FuncPORO(isTimeFunc),
		ValueEx: stdlib.FuncPOROEx(isTimeFunc),
	},
}

func utcFunc() ugo.Object { return utcLoc }

func localFunc() ugo.Object { return localLoc }

func monthStringFunc(m int) ugo.Object {
	return ugo.String(time.Month(m).String())
}

func weekdayStringFunc(w int) ugo.Object {
	return ugo.String(time.Weekday(w).String())
}

func durationStringFunc(d int64) ugo.Object {
	return ugo.String(time.Duration(d).String())
}

func durationNanosecondsFunc(d int64) ugo.Object {
	return ugo.Int(time.Duration(d).Nanoseconds())
}

func durationMicrosecondsFunc(d int64) ugo.Object {
	return ugo.Int(time.Duration(d).Microseconds())
}

func durationMillisecondsFunc(d int64) ugo.Object {
	return ugo.Int(time.Duration(d).Milliseconds())
}

func durationSecondsFunc(d int64) ugo.Object {
	return ugo.Float(time.Duration(d).Seconds())
}

func durationMinutesFunc(d int64) ugo.Object {
	return ugo.Float(time.Duration(d).Minutes())
}

func durationHoursFunc(d int64) ugo.Object {
	return ugo.Float(time.Duration(d).Hours())
}

func sleepFunc(c ugo.Call) (ugo.Object, error) {
	if err := c.CheckLen(1); err != nil {
		return ugo.Undefined, err
	}
	arg0 := c.Get(0)

	var dur time.Duration
	if v, ok := ugo.ToGoInt64(arg0); !ok {
		return newArgTypeErr("1st", "int", arg0.TypeName())
	} else {
		dur = time.Duration(v)
	}

	for {
		if dur <= 10*time.Millisecond {
			time.Sleep(dur)
			break
		}
		dur -= 10 * time.Millisecond
		time.Sleep(10 * time.Millisecond)
		if c.VM().Aborted() {
			return ugo.Undefined, ugo.ErrVMAborted
		}
	}
	return ugo.Undefined, nil
}

func parseDurationFunc(s string) (ugo.Object, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}
	return ugo.Int(d), nil
}

func durationRoundFunc(d, m int64) ugo.Object {
	return ugo.Int(time.Duration(d).Round(time.Duration(m)))
}

func durationTruncateFunc(d, m int64) ugo.Object {
	return ugo.Int(time.Duration(d).Truncate(time.Duration(m)))
}

func fixedZoneFunc(name string, sec int) ugo.Object {
	return &Location{Value: time.FixedZone(name, sec)}
}

func loadLocationFunc(name string) (ugo.Object, error) {
	l, err := time.LoadLocation(name)
	if err != nil {
		return ugo.Undefined, err
	}
	return &Location{Value: l}, nil
}

func isLocationFunc(o ugo.Object) ugo.Object {
	_, ok := o.(*Location)
	return ugo.Bool(ok)
}

func zerotimeFunc() ugo.Object { return zeroTime }

func sinceFunc(t *Time) ugo.Object { return ugo.Int(time.Since(t.Value)) }

func untilFunc(t *Time) ugo.Object { return ugo.Int(time.Until(t.Value)) }

func dateFunc(args ...ugo.Object) (ugo.Object, error) {
	return dateFuncEx(ugo.NewCall(nil, args))
}

func dateFuncEx(c ugo.Call) (ugo.Object, error) {
	size := c.Len()
	if size < 3 || size > 8 {
		return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
			"want=3..8 got=" + strconv.Itoa(size))
	}
	ymdHmsn := [7]int{}
	loc := &Location{Value: time.Local}
	var ok bool
	for i := 0; i < size; i++ {
		arg := c.Get(i)
		if i < 7 {
			ymdHmsn[i], ok = ugo.ToGoInt(arg)
			if !ok {
				return newArgTypeErr(strconv.Itoa(i+1), "int", arg.TypeName())
			}
			continue
		}
		loc, ok = arg.(*Location)
		if !ok {
			return newArgTypeErr(strconv.Itoa(i+1), "location", arg.TypeName())
		}
	}

	return &Time{
		Value: time.Date(ymdHmsn[0], time.Month(ymdHmsn[1]), ymdHmsn[2],
			ymdHmsn[3], ymdHmsn[4], ymdHmsn[5], ymdHmsn[6], loc.Value),
	}, nil
}

func nowFunc() ugo.Object { return &Time{Value: time.Now()} }

func parseFunc(args ...ugo.Object) (ugo.Object, error) {
	return parseFuncEx(ugo.NewCall(nil, args))
}

func parseFuncEx(c ugo.Call) (ugo.Object, error) {
	size := c.Len()
	if size != 2 && size != 3 {
		return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
			"want=2..3 got=" + strconv.Itoa(size))
	}
	layout, ok := ugo.ToGoString(c.Get(0))
	if !ok {
		return newArgTypeErr("1st", "string", c.Get(0).TypeName())
	}
	value, ok := ugo.ToGoString(c.Get(1))
	if !ok {
		return newArgTypeErr("2nd", "string", c.Get(1).TypeName())
	}
	if size == 2 {
		tm, err := time.Parse(layout, value)
		if err != nil {
			return ugo.Undefined, err
		}
		return &Time{Value: tm}, nil
	}
	loc, ok := ToLocation(c.Get(2))
	if !ok {
		return newArgTypeErr("3rd", "location", c.Get(2).TypeName())
	}
	tm, err := time.ParseInLocation(layout, value, loc.Value)
	if err != nil {
		return ugo.Undefined, err
	}
	return &Time{Value: tm}, nil
}

func unixFunc(args ...ugo.Object) (ugo.Object, error) {
	return unixFuncEx(ugo.NewCall(nil, args))
}

func unixFuncEx(c ugo.Call) (ugo.Object, error) {
	size := c.Len()
	if size != 1 && size != 2 {
		return ugo.Undefined, ugo.ErrWrongNumArguments.NewError(
			"want=1..2 got=" + strconv.Itoa(size))
	}

	sec, ok := ugo.ToGoInt64(c.Get(0))
	if !ok {
		return newArgTypeErr("1st", "int", c.Get(0).TypeName())
	}

	var nsec int64
	if size > 1 {
		nsec, ok = ugo.ToGoInt64(c.Get(1))
		if !ok {
			return newArgTypeErr("2nd", "int", c.Get(1).TypeName())
		}
	}
	return &Time{Value: time.Unix(sec, nsec)}, nil
}

func isTimeFunc(o ugo.Object) ugo.Object {
	_, ok := o.(*Time)
	return ugo.Bool(ok)
}
