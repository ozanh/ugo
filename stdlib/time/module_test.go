package time_test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
	. "github.com/ozanh/ugo/stdlib/time"
)

func TestModuleTypes(t *testing.T) {
	l := &Location{Value: time.UTC}
	require.Equal(t, "location", l.TypeName())
	require.False(t, l.IsFalsy())
	require.Equal(t, "UTC", l.String())
	require.True(t, (&Location{}).Equal(&Location{}))
	require.True(t, (&Location{}).Equal(String("UTC")))
	require.False(t, (&Location{}).Equal(Int(0)))
	require.False(t, l.CanCall())
	require.False(t, l.CanIterate())
	require.Nil(t, l.Iterate())
	require.Equal(t, ErrNotIndexAssignable, l.IndexSet(nil, nil))
	_, err := l.IndexGet(nil)
	require.Equal(t, ErrNotIndexable, err)

	tm := &Time{}
	require.Equal(t, "time", tm.TypeName())
	require.True(t, tm.IsFalsy())
	require.NotEmpty(t, tm.String())
	require.True(t, tm.Equal(&Time{}))
	require.False(t, tm.Equal(Int(0)))
	require.False(t, tm.CanCall())
	require.False(t, tm.CanIterate())
	require.Nil(t, tm.Iterate())
	require.Equal(t, ErrNotIndexAssignable, tm.IndexSet(nil, nil))
	r, err := tm.IndexGet(String(""))
	require.NoError(t, err)
	require.Equal(t, Undefined, r)

	now := time.Now()
	tm2 := &Time{Value: now}
	require.False(t, tm2.IsFalsy())
	require.Equal(t, now.String(), tm2.String())

	var b bytes.Buffer
	err = gob.NewEncoder(&b).Encode(tm2)
	require.NoError(t, err)
	var tm3 Time
	err = gob.NewDecoder(&b).Decode(&tm3)
	require.NoError(t, err)
	require.Equal(t, tm2.Value.Format(time.RFC3339Nano),
		tm3.Value.Format(time.RFC3339Nano))

	// test registers
	//time
	ret, err := ToObject(now)
	require.NoError(t, err)
	require.IsType(t, &Time{}, ret)
	require.Equal(t, now.String(), ret.String())

	iface := ToInterface(ret)
	require.Equal(t, now, iface)

	ret, err = ToObject(&now)
	require.NoError(t, err)
	require.IsType(t, &Time{}, ret)
	require.Equal(t, now.String(), ret.String())

	ret, err = ToObject((*time.Time)(nil))
	require.NoError(t, err)
	require.Equal(t, Undefined, ret)

	//duration
	ret, err = ToObject(time.Second)
	require.NoError(t, err)
	require.IsType(t, Int(0), ret)
	require.Equal(t, Int(time.Second), ret)

	//location
	ret, err = ToObject(time.UTC)
	require.NoError(t, err)
	require.IsType(t, &Location{}, ret)
	require.Equal(t, time.UTC.String(), ret.String())

	iface = ToInterface(ret)
	require.Equal(t, time.UTC, iface)

	ret, err = ToObject((*time.Location)(nil))
	require.NoError(t, err)
	require.Equal(t, Undefined, ret)
}

func TestModuleMonthWeekday(t *testing.T) {
	f := Module["MonthString"].(*Function)
	_, err := f.Call()
	require.Error(t, err)
	_, err = f.Call(String(""))
	require.Error(t, err)

	for i := 1; i <= 12; i++ {
		require.Contains(t, Module, time.Month(i).String())
		require.Equal(t, Int(i), Module[time.Month(i).String()])

		r, err := f.Call(Int(i))
		require.NoError(t, err)
		require.EqualValues(t, time.Month(i).String(), r)
	}

	f = Module["WeekdayString"].(*Function)
	_, err = f.Call()
	require.Error(t, err)
	_, err = f.Call(String(""))
	require.Error(t, err)
	for i := 0; i <= 6; i++ {
		require.Contains(t, Module, time.Weekday(i).String())
		require.Equal(t, Int(i), Module[time.Weekday(i).String()])

		r, err := f.Call(Int(i))
		require.NoError(t, err)
		require.EqualValues(t, time.Weekday(i).String(), r)
	}
}

func TestModuleFormats(t *testing.T) {
	require.Equal(t, Module["ANSIC"], String(time.ANSIC))
	require.Equal(t, Module["UnixDate"], String(time.UnixDate))
	require.Equal(t, Module["RubyDate"], String(time.RubyDate))
	require.Equal(t, Module["RFC822"], String(time.RFC822))
	require.Equal(t, Module["RFC822Z"], String(time.RFC822Z))
	require.Equal(t, Module["RFC850"], String(time.RFC850))
	require.Equal(t, Module["RFC1123"], String(time.RFC1123))
	require.Equal(t, Module["RFC1123Z"], String(time.RFC1123Z))
	require.Equal(t, Module["RFC3339"], String(time.RFC3339))
	require.Equal(t, Module["RFC3339Nano"], String(time.RFC3339Nano))
	require.Equal(t, Module["Kitchen"], String(time.Kitchen))
	require.Equal(t, Module["Stamp"], String(time.Stamp))
	require.Equal(t, Module["StampMilli"], String(time.StampMilli))
	require.Equal(t, Module["StampMicro"], String(time.StampMicro))
	require.Equal(t, Module["StampNano"], String(time.StampNano))
}

func TestModuleDuration(t *testing.T) {
	require.Equal(t, Module["Nanosecond"], Int(time.Nanosecond))
	require.Equal(t, Module["Microsecond"], Int(time.Microsecond))
	require.Equal(t, Module["Millisecond"], Int(time.Millisecond))
	require.Equal(t, Module["Second"], Int(time.Second))
	require.Equal(t, Module["Minute"], Int(time.Minute))
	require.Equal(t, Module["Hour"], Int(time.Hour))

	goFnMap := map[string]func(time.Duration) interface{}{
		"Nanoseconds": func(d time.Duration) interface{} {
			return d.Nanoseconds()
		},
		"Microseconds": func(d time.Duration) interface{} {
			return d.Microseconds()
		},
		"Milliseconds": func(d time.Duration) interface{} {
			return d.Milliseconds()
		},
		"Seconds": func(d time.Duration) interface{} {
			return d.Seconds()
		},
		"Minutes": func(d time.Duration) interface{} {
			return d.Minutes()
		},
		"Hours": func(d time.Duration) interface{} {
			return d.Hours()
		},
	}
	durToString := Module["DurationString"].(*Function)
	_, err := durToString.Call()
	require.Error(t, err)

	durParse := Module["ParseDuration"].(*Function)
	_, err = durParse.Call()
	require.Error(t, err)
	_, err = durParse.Call(String(""))
	require.Error(t, err)
	_, err = durParse.Call(Int(0))
	require.NoError(t, err)

	testCases := []struct {
		dur time.Duration
	}{
		{time.Nanosecond}, {time.Microsecond}, {time.Millisecond}, {time.Second},
		{time.Minute}, {time.Hour},
		{time.Hour + time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond},
		{2*time.Hour + 3*time.Minute + 4*time.Second + 5*time.Millisecond + 6*time.Microsecond + 7*time.Nanosecond},
		{-2*time.Hour + 3*time.Minute + 4*time.Second + 5*time.Millisecond + 6*time.Microsecond + 7*time.Nanosecond},
	}

	for _, tC := range testCases {
		for fn := range goFnMap {
			t.Run(fmt.Sprintf("%s:%s", tC.dur, fn), func(t *testing.T) {
				f := Module["Duration"+fn].(*Function)
				ret, err := f.Call(Int(tC.dur))
				require.NoError(t, err)
				expect := goFnMap[fn](tC.dur)
				require.EqualValues(t, expect, ret)

				// test illegal type
				_, err = f.Call(&illegalDur{Value: tC.dur})
				require.Error(t, err)
				// test no arg
				_, err = f.Call()
				require.Error(t, err)

				// test to string
				s, err := durToString.Call(Int(tC.dur))
				require.NoError(t, err)
				require.EqualValues(t, tC.dur.String(), s)

				// test parse
				d, err := durParse.Call(s)
				require.NoError(t, err)
				ed, err := time.ParseDuration(tC.dur.String())
				require.NoError(t, err)
				require.EqualValues(t, ed, d)
			})
		}
	}

	durRound := Module["DurationRound"].(*Function)
	r, err := durRound.Call(Int(time.Second+time.Millisecond),
		Int(time.Second))
	require.NoError(t, err)
	require.EqualValues(t, time.Second, r)
	_, err = durRound.Call(Int(0))
	require.Error(t, err)
	_, err = durRound.Call(String(""), Int(0))
	require.Error(t, err)
	_, err = durRound.Call(Int(0), String(""))
	require.Error(t, err)

	durTruncate := Module["DurationTruncate"].(*Function)
	r, err = durTruncate.Call(Int(time.Second+5*time.Millisecond),
		Int(2*time.Millisecond))
	require.NoError(t, err)
	require.EqualValues(t, time.Second+4*time.Millisecond, r)
	_, err = durTruncate.Call(Int(0))
	require.Error(t, err)
	_, err = durTruncate.Call(String(""), Int(0))
	require.Error(t, err)
	_, err = durTruncate.Call(Int(0), String(""))
	require.Error(t, err)
}

func TestModuleLocation(t *testing.T) {
	fixedZone := Module["FixedZone"].(*Function)
	r, err := fixedZone.Call(String("Ankara"), Int(3*60*60))
	require.NoError(t, err)
	require.Equal(t, "Ankara", r.String())

	_, err = fixedZone.Call(String("Ankara"))
	require.Error(t, err)
	_, err = fixedZone.Call(String("Ankara"), Uint(0))
	require.NoError(t, err)
	_, err = fixedZone.Call(Int(0), Array{})
	require.Error(t, err)
	_, err = fixedZone.Call()
	require.Error(t, err)

	loadLocation := Module["LoadLocation"].(*Function)
	r, err = loadLocation.Call(String("Europe/Istanbul"))
	require.NoError(t, err)
	require.Equal(t, "Europe/Istanbul", r.String())
	r, err = loadLocation.Call(String(""))
	require.NoError(t, err)
	require.Equal(t, "UTC", r.String())
	_, err = loadLocation.Call()
	require.Error(t, err)
	_, err = loadLocation.Call(Int(0))
	require.Error(t, err)
	_, err = loadLocation.Call(String("invalid"))
	require.Error(t, err)

	isLocation := Module["IsLocation"].(*Function)
	r, err = isLocation.Call(&Location{Value: time.Local})
	require.NoError(t, err)
	require.EqualValues(t, true, r)
	r, err = isLocation.Call(Int(0))
	require.NoError(t, err)
	require.EqualValues(t, false, r)
	_, err = isLocation.Call(Int(0), Int(0))
	require.Error(t, err)
	_, err = isLocation.Call()
	require.Error(t, err)
}

func TestModuleTime(t *testing.T) {
	now := time.Now()

	require.Equal(t, now.String(), (&Time{Value: now}).String())

	_, err := (&Time{}).Call()
	require.Same(t, ErrNotCallable, err)

	zTime := Module["Time"].(*Function)
	r, err := zTime.Call()
	require.NoError(t, err)
	require.True(t, r.(*Time).Value.IsZero())
	_, err = zTime.Call(String(""))
	require.Error(t, err)

	since := Module["Since"].(*Function)
	r, err = since.Call(&Time{Value: now})
	require.NoError(t, err)
	require.GreaterOrEqual(t, int64(r.(Int)), int64(0))
	_, err = since.Call()
	require.Error(t, err)
	_, err = since.Call(String(""))
	require.Error(t, err)

	until := Module["Until"].(*Function)
	r, err = until.Call(&Time{Value: now})
	require.NoError(t, err)
	require.LessOrEqual(t, int64(r.(Int)), int64(0))
	_, err = until.Call()
	require.Error(t, err)
	_, err = until.Call(String(""))
	require.Error(t, err)

	date := Module["Date"].(*Function)
	r, err = date.Call(Int(2020), Int(11), Int(8),
		Int(1), Int(2), Int(3), Int(4),
		&Location{Value: time.Local})
	require.NoError(t, err)
	require.Equal(t,
		time.Date(2020, 11, 8, 1, 2, 3, 4, time.Local), r.(*Time).Value)
	r, err = date.Call(Int(2020), Int(11), Int(8))
	require.NoError(t, err)
	require.Equal(t,
		time.Date(2020, 11, 8, 0, 0, 0, 0, time.Local), r.(*Time).Value)

	nowf := Module["Now"].(*Function)
	r, err = nowf.Call()
	require.NoError(t, err)
	require.False(t, r.(*Time).Value.IsZero())
	_, err = nowf.Call(Int(0))
	require.Error(t, err)

	RFC3339Nano := Module["RFC3339Nano"]
	parse := Module["Parse"].(*Function)
	r, err = parse.Call(RFC3339Nano, String(now.Format(time.RFC3339Nano)))
	require.NoError(t, err)
	require.Equal(t, now.Format(time.RFC3339Nano),
		r.(*Time).Value.Format(time.RFC3339Nano))

	r, err = parse.Call(RFC3339Nano, String(now.Format(time.RFC3339Nano)),
		&Location{Value: time.Local})
	require.NoError(t, err)
	require.Equal(t, now.Format(time.RFC3339Nano),
		r.(*Time).Value.Format(time.RFC3339Nano))

	_, err = parse.Call()
	require.Error(t, err)

	unix := Module["Unix"].(*Function)
	r, err = unix.Call(Int(now.Unix()))
	require.NoError(t, err)
	require.Equal(t, time.Unix(now.Unix(), 0), r.(*Time).Value)
	r, err = unix.Call(Int(now.Unix()), Int(1))
	require.NoError(t, err)
	require.Equal(t, time.Unix(now.Unix(), 1), r.(*Time).Value)
	_, err = unix.Call()
	require.Error(t, err)

	add := Module["Add"].(*Function)
	r, err = add.Call(&Time{Value: now}, Int(time.Second))
	require.NoError(t, err)
	require.Equal(t, now.Add(time.Second), r.(*Time).Value)
	_, err = add.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = add.Call(&Time{Value: now}, &Time{Value: now})
	require.Error(t, err)
	_, err = add.Call()
	require.Error(t, err)

	sub := Module["Sub"].(*Function)
	r, err = sub.Call(&Time{Value: now}, &Time{Value: now.Add(-time.Hour)})
	require.NoError(t, err)
	require.EqualValues(t, time.Hour, r.(Int))
	_, err = sub.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = sub.Call(&Time{Value: now}, Int(0))
	require.NoError(t, err)
	_, err = sub.Call()
	require.Error(t, err)

	addDate := Module["AddDate"].(*Function)
	r, err = addDate.Call(&Time{Value: now},
		Int(1), Int(2), Int(3))
	require.NoError(t, err)
	require.EqualValues(t, now.AddDate(1, 2, 3), r.(*Time).Value)
	_, err = addDate.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = addDate.Call(&Time{Value: now}, Int(0))
	require.Error(t, err)
	_, err = addDate.Call()
	require.Error(t, err)

	after := Module["After"].(*Function)
	r, err = after.Call(&Time{Value: now}, &Time{Value: now.Add(time.Hour)})
	require.NoError(t, err)
	require.EqualValues(t, false, r)
	r, err = after.Call(&Time{Value: now}, &Time{Value: now.Add(-time.Hour)})
	require.NoError(t, err)
	require.EqualValues(t, true, r)
	_, err = after.Call(&Time{Value: now}, Int(0))
	require.NoError(t, err)
	_, err = after.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = after.Call()
	require.Error(t, err)

	before := Module["Before"].(*Function)
	r, err = before.Call(&Time{Value: now}, &Time{Value: now.Add(time.Hour)})
	require.NoError(t, err)
	require.EqualValues(t, true, r)
	r, err = before.Call(&Time{Value: now}, &Time{Value: now.Add(-time.Hour)})
	require.NoError(t, err)
	require.EqualValues(t, false, r)
	_, err = before.Call(&Time{Value: now}, Int(0))
	require.NoError(t, err)
	_, err = before.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = before.Call()
	require.Error(t, err)

	appendFormat := Module["AppendFormat"].(*Function)
	b := make(Bytes, 100)
	r, err = appendFormat.Call(&Time{Value: now}, b, RFC3339Nano)
	require.NoError(t, err)
	require.EqualValues(t,
		now.AppendFormat(make([]byte, 100), time.RFC3339Nano), r)
	_, err = appendFormat.Call(&Time{Value: now}, b)
	require.Error(t, err)
	_, err = appendFormat.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = appendFormat.Call()
	require.Error(t, err)

	format := Module["Format"].(*Function)
	r, err = format.Call(&Time{Value: now}, RFC3339Nano)
	require.NoError(t, err)
	require.EqualValues(t, now.Format(time.RFC3339Nano), r)
	_, err = format.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = format.Call()
	require.Error(t, err)

	timeIn := Module["In"].(*Function)
	r, err = timeIn.Call(&Time{Value: now}, &Location{Value: time.Local})
	require.NoError(t, err)
	require.False(t, r.(*Time).Value.IsZero())
	_, err = timeIn.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = timeIn.Call()
	require.Error(t, err)

	round := Module["Round"].(*Function)
	r, err = round.Call(&Time{Value: now}, Int(time.Second))
	require.NoError(t, err)
	require.Equal(t, now.Round(time.Second), r.(*Time).Value)
	_, err = round.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = round.Call()
	require.Error(t, err)

	truncate := Module["Truncate"].(*Function)
	r, err = truncate.Call(&Time{Value: now}, Int(time.Hour))
	require.NoError(t, err)
	require.Equal(t, now.Truncate(time.Hour), r.(*Time).Value)
	_, err = truncate.Call(&Time{Value: now})
	require.Error(t, err)
	_, err = truncate.Call()
	require.Error(t, err)

	isTime := Module["IsTime"].(*Function)
	r, err = isTime.Call(&Time{Value: now})
	require.NoError(t, err)
	require.EqualValues(t, true, r)
	r, err = isTime.Call(Int(0))
	require.NoError(t, err)
	require.EqualValues(t, false, r)
	_, err = isTime.Call(Int(0), Int(0))
	require.Error(t, err)
	_, err = isTime.Call()
	require.Error(t, err)

	y, m, d := now.Date()
	testTimeSelector(t, &Time{Value: now}, "Date",
		Map{"year": Int(y), "month": Int(m), "day": Int(d)})
	h, min, s := now.Clock()
	testTimeSelector(t, &Time{Value: now}, "Clock",
		Map{"hour": Int(h), "minute": Int(min), "second": Int(s)})
	testTimeSelector(t, &Time{Value: now}, "UTC", &Time{Value: now.UTC()})
	testTimeSelector(t, &Time{Value: now}, "Unix", Int(now.Unix()))
	testTimeSelector(t, &Time{Value: now}, "UnixNano", Int(now.UnixNano()))
	testTimeSelector(t, &Time{Value: now}, "Year", Int(now.Year()))
	testTimeSelector(t, &Time{Value: now}, "Month", Int(now.Month()))
	testTimeSelector(t, &Time{Value: now}, "Day", Int(now.Day()))
	testTimeSelector(t, &Time{Value: now}, "Hour", Int(now.Hour()))
	testTimeSelector(t, &Time{Value: now}, "Minute", Int(now.Minute()))
	testTimeSelector(t, &Time{Value: now}, "Second", Int(now.Second()))
	testTimeSelector(t, &Time{Value: now}, "Nanosecond", Int(now.Nanosecond()))
	testTimeSelector(t, &Time{Value: now}, "IsZero", Bool(false))
	testTimeSelector(t, &Time{Value: now}, "Local", &Time{Value: now.Local()})
	testTimeSelector(t, &Time{Value: now}, "Location",
		&Location{Value: now.Location()})
	testTimeSelector(t, &Time{Value: now}, "YearDay", Int(now.YearDay()))
	testTimeSelector(t, &Time{Value: now}, "Weekday", Int(now.Weekday()))
	y, w := now.ISOWeek()
	testTimeSelector(t, &Time{Value: now}, "ISOWeek",
		Map{"year": Int(y), "week": Int(w)})
	name, offset := now.Zone()
	testTimeSelector(t, &Time{Value: now}, "Zone",
		Map{"name": String(name), "offset": Int(offset)})
	testTimeSelector(t, &Time{Value: now}, "XYZ", Undefined)
}

func testTimeSelector(t *testing.T, tm Object,
	selector string, expected Object) {
	t.Helper()
	v, err := tm.IndexGet(String(selector))
	require.NoError(t, err)
	require.Equal(t, expected, v)
}

func TestScript(t *testing.T) {
	catch := func(s string) string {
		return fmt.Sprintf(`
		time := import("time")
		try {
			return %s
		} catch err {
			return string(err)
		}
		`, s)
	}
	idxTypeErr := func(expected, got string) String {
		return String(NewIndexTypeError(expected, got).String())
	}
	opTypeErr := func(tok, lhs, rhs string) String {
		return String(NewOperandTypeError(
			tok, lhs, rhs).String())
	}
	typeErr := func(pos, expected, got string) String {
		return String(NewArgumentTypeError(pos, expected, got).String())
	}
	nwrongArgs := func(want1, want2, got int) String {
		var msg string
		if want2 <= 0 {
			msg = fmt.Sprintf("want=%d got=%d", want1, got)
		} else {
			msg = fmt.Sprintf("want=%d..%d got=%d", want1, want2, got)
		}
		return String(ErrWrongNumArguments.NewError(msg).String())
	}
	expectRun(t, `import("time")`, nil, Undefined)

	expectRun(t, catch(`time.Now()[1]`),
		nil, idxTypeErr("string", "int"))
	expectRun(t, catch(`time.Now() + 'c'`),
		nil, opTypeErr("+", "time", "char"))
	expectRun(t, catch(`time.Now()()`), nil, String("NotCallableError: time"))
	expectRun(t, catch(`time.Date()`), nil, nwrongArgs(3, 8, 0))
	expectRun(t, catch(`time.Date(1)`), nil, nwrongArgs(3, 8, 1))
	expectRun(t, catch(`time.Date(1, 2)`), nil, nwrongArgs(3, 8, 2))
	expectRun(t, catch(`time.Date(1, 2, "")`),
		nil, typeErr("3", "int", "string"))
	expectRun(t, catch(`time.Date(1, 2, 3, 4, 5, 6, 7, "")`),
		nil, typeErr("8", "location", "string"))
	expectRun(t, catch(`time.Parse("", 1)`),
		nil, String("error: parsing time \"1\": extra text: \"1\""))
	expectRun(t, catch(`time.Parse("", "", 1)`),
		nil, typeErr("3rd", "location", "int"))
	expectRun(t, catch(`time.Unix("")`),
		nil, typeErr("1st", "int", "string"))
	expectRun(t, catch(`time.Unix(1, "")`),
		nil, typeErr("2nd", "int", "string"))
	expectRun(t, catch(`time.AddDate(time.Now(), "", 1, 2)`),
		nil, typeErr("2nd", "int", "string"))
	expectRun(t, catch(`time.AddDate(time.Now(), 1, "", 2)`),
		nil, typeErr("3rd", "int", "string"))
	expectRun(t, catch(`time.AddDate(time.Now(), 1, 2, "")`),
		nil, typeErr("4th", "int", "string"))
	expectRun(t, catch(`time.After(1, 2)`), nil, False)
	expectRun(t, catch(`time.Before(1, 2)`), nil, True)
	expectRun(t, catch(`time.AppendFormat(1, 2, 3)`),
		nil, typeErr("2nd", "bytes", "int"))
	expectRun(t, catch(`time.AppendFormat(time.Now(), 1, 2)`),
		nil, typeErr("2nd", "bytes", "int"))
	expectRun(t, catch(`time.AppendFormat(time.Time(), bytes(), 1)`),
		nil, Bytes{0x31})
	expectRun(t, catch(`time.In(1, 2)`),
		nil, typeErr("2nd", "location", "int"))
	expectRun(t, catch(`time.In(time.Now(), 2)`),
		nil, typeErr("2nd", "location", "int"))
	expectRun(t, catch(`time.Round(time.Now(), "")`),
		nil, typeErr("2nd", "int", "string"))
	expectRun(t, catch(`time.Truncate(time.Now(), "")`),
		nil, typeErr("2nd", "int", "string"))
	expectRun(t, catch(`time.Sleep("")`),
		nil, typeErr("1st", "int", "string"))

	expectRun(t, `mod := import("time"); return mod.__module_name__`,
		nil, String("time"))

	tm := time.Now()
	expectRun(t, `
	param p1; time := import("time"); return time.Format(p1, time.RFC3339Nano)`,
		newOpts().Args(&Time{Value: tm}), String(tm.Format(time.RFC3339Nano)))
	expectRun(t, `param p1; return p1.UnixNano`,
		newOpts().Args(&Time{Value: tm}), Int(tm.UnixNano()))

	expectRun(t, `
	param p1
	time := import("time")
	try {
		time.Sleep(time.Millisecond)
	} finally {
		dur := time.Since(p1)
		return dur > 0 ? true: false 
	}
	`, newOpts().Args(&Time{Value: tm}), True)

	expectRun(t, `return import("time").IsTime(0)`, nil, False)
	expectRun(t, `param p1; time := import("time"); return time.IsTime(p1)`,
		newOpts().Args(&Time{Value: tm}), True)
	expectRun(t, `time := import("time"); return time.IsTime(time.Now())`,
		nil, True)
	expectRun(t, `
	time := import("time")
	return time.IsLocation(time.FixedZone("abc", 3*60*60))`, nil, True)
	expectRun(t, `param p1; return p1==p1`,
		newOpts().Args(&Time{Value: tm}), True)
	expectRun(t, `param p1; time := import("time"); return time.Now()==p1`,
		newOpts().Args(&Time{Value: tm}), False)
	expectRun(t, `param p1; time := import("time"); return time.Now()>=p1`,
		newOpts().Args(&Time{Value: tm}), True)
	expectRun(t, `param p1; time := import("time"); return time.Now()<p1`,
		newOpts().Args(&Time{Value: tm}), False)
	expectRun(t, `param p1; time := import("time"); return time.Now()>p1`,
		newOpts().Args(&Time{Value: tm}), True)
	expectRun(t, `time := import("time"); return (time.Now()+time.Second)>=time.Now()`, nil, True)
	expectRun(t, `time := import("time"); return (time.Now()+time.Second)<=time.Now()`, nil, False)
	expectRun(t, `time := import("time"); return (time.Now()-10*time.Second)<=time.Now()`, nil, True)
	expectRun(t, `time := import("time"); return time.Now() == undefined`, nil, False)
	expectRun(t, `time := import("time"); return time.Now() > undefined`, nil, True)
	expectRun(t, `time := import("time"); return time.Now() >= undefined`, nil, True)
	expectRun(t, `time := import("time"); return time.Now() < undefined`, nil, False)
	expectRun(t, `time := import("time"); return time.Now() <= undefined`, nil, False)
	expectRun(t, `
	time := import("time")
	t1 := time.Now()
	t2 := t1 + time.Second
	return t2 - t1
	`, nil, Int(time.Second))

	// methods
	// .Add
	expectRun(t, `time := import("time"); return time.Time().Add(10*time.Second)`,
		nil, &Time{Value: time.Time{}.Add(10 * time.Second)})
	expectRun(t, catch(`time.Time().Add()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().Add(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().Add(undefined)`), nil, typeErr("1st", "int", "undefined"))

	// .Sub
	expectRun(t, `time := import("time");
	t1 := time.Time()
	t2 := time.Time().Add(10*time.Second)
	return t2.Sub(t1)`,
		nil, Int(10*time.Second))
	expectRun(t, catch(`time.Time().Sub()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().Sub(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().Sub(undefined)`), nil, typeErr("1st", "time", "undefined"))

	// .AddDate
	expectRun(t, `time := import("time"); return time.Time().AddDate(1, 2, 3)`,
		nil, &Time{Value: time.Time{}.AddDate(1, 2, 3)})
	expectRun(t, catch(`time.Time().AddDate()`), nil, nwrongArgs(3, -1, 0))
	expectRun(t, catch(`time.Time().AddDate(1, 2)`), nil, nwrongArgs(3, -1, 2))
	expectRun(t, catch(`time.Time().AddDate(1, 2, 3, 4)`), nil, nwrongArgs(3, -1, 4))
	expectRun(t, catch(`time.Time().AddDate(undefined, 2, 3)`), nil, typeErr("1st", "int", "undefined"))

	// .After
	expectRun(t, `time := import("time"); return time.Time().After(time.Time())`, nil, False)
	expectRun(t, `time := import("time"); return time.Time().Add(time.Second).After(time.Time())`, nil, True)
	expectRun(t, catch(`time.Time().After()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().After(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().After(undefined)`), nil, typeErr("1st", "time", "undefined"))

	// .Before
	expectRun(t, `time := import("time"); return time.Time().Before(time.Time())`, nil, False)
	expectRun(t, `time := import("time"); return time.Time().Add(-time.Second).Before(time.Time())`, nil, True)
	expectRun(t, catch(`time.Time().Before()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().Before(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().Before(undefined)`), nil, typeErr("1st", "time", "undefined"))

	// .Format
	expectRun(t, `time := import("time"); return time.Time().Format("2006-01-02")`, nil, String("0001-01-01"))
	expectRun(t, catch(`time.Time().Format()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().Format(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().Format(undefined)`), nil, typeErr("1st", "string", "undefined"))

	// .AppendFormat
	expectRun(t, `time := import("time"); return time.Time().AppendFormat("", "2006-01-02")`, nil, Bytes("0001-01-01"))
	expectRun(t, catch(`time.Time().AppendFormat()`), nil, nwrongArgs(2, -1, 0))
	expectRun(t, catch(`time.Time().AppendFormat(1)`), nil, nwrongArgs(2, -1, 1))
	expectRun(t, catch(`time.Time().AppendFormat(1, 2, 3)`), nil, nwrongArgs(2, -1, 3))
	expectRun(t, catch(`time.Time().AppendFormat(undefined, "2006-01-02")`), nil, typeErr("1st", "bytes", "undefined"))

	// .In
	expectRun(t, `param p1; time := import("time"); return p1.In(time.UTC())`,
		newOpts().Args(&Time{Value: tm}), &Time{Value: tm.In(time.UTC)})
	expectRun(t, catch(`time.Time().In()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().In(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().In(undefined)`), nil, typeErr("1st", "location", "undefined"))

	// .Round
	expectRun(t, `param p1; time := import("time"); return p1.Round(time.Second)`,
		newOpts().Args(&Time{Value: tm}), &Time{Value: tm.Round(time.Second)})
	expectRun(t, catch(`time.Time().Round()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().Round(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().Round(undefined)`), nil, typeErr("1st", "int", "undefined"))

	// .Truncate
	expectRun(t, `param p1; time := import("time"); return p1.Truncate(time.Second)`,
		newOpts().Args(&Time{Value: tm}), &Time{Value: tm.Truncate(time.Second)})
	expectRun(t, catch(`time.Time().Truncate()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().Truncate(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().Truncate(undefined)`), nil, typeErr("1st", "int", "undefined"))

	// .Equal
	expectRun(t, `time := import("time"); return time.Time().Equal(time.Time())`, nil, True)
	expectRun(t, `param (p1,p2); return p1.Equal(p2)`,
		newOpts().Args(&Time{Value: tm}, &Time{Value: tm}), True)
	expectRun(t, `param (p1,p2); return p1.Equal(p2)`,
		newOpts().Args(&Time{Value: tm}, &Time{Value: tm.Add(time.Second)}), False)
	expectRun(t, catch(`time.Time().Equal()`), nil, nwrongArgs(1, -1, 0))
	expectRun(t, catch(`time.Time().Equal(1, 2)`), nil, nwrongArgs(1, -1, 2))
	expectRun(t, catch(`time.Time().Equal(undefined)`), nil, typeErr("1st", "time", "undefined"))

	// .Date
	expectRun(t, `time := import("time"); return time.Time().Date()`,
		nil, Map{"day": Int(1), "month": Int(1), "year": Int(1)})
	expectRun(t, catch(`time.Time().Date(1)`), nil, nwrongArgs(0, -1, 1))

	// .Clock
	hour, minute, second := tm.Clock()
	expectRun(t, `param p1; return p1.Clock()`,
		newOpts().Args(&Time{Value: tm}),
		Map{"hour": Int(hour), "minute": Int(minute), "second": Int(second)})
	expectRun(t, catch(`time.Time().Clock(1)`), nil, nwrongArgs(0, -1, 1))

	// .UTC
	expectRun(t, `param p1; return p1.UTC()`,
		newOpts().Args(&Time{Value: tm}), &Time{Value: tm.UTC()})
	expectRun(t, catch(`time.Time().UTC(1)`), nil, nwrongArgs(0, -1, 1))

	// .Unix
	expectRun(t, `param p1; return p1.Unix()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Unix()))
	expectRun(t, catch(`time.Time().Unix(1)`), nil, nwrongArgs(0, -1, 1))

	// .UnixNano
	expectRun(t, `param p1; return p1.UnixNano()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.UnixNano()))
	expectRun(t, catch(`time.Time().UnixNano(1)`), nil, nwrongArgs(0, -1, 1))

	// .Year
	expectRun(t, `param p1; return p1.Year()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Year()))
	expectRun(t, catch(`time.Time().Year(1)`), nil, nwrongArgs(0, -1, 1))

	// .Month
	expectRun(t, `param p1; return p1.Month()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Month()))
	expectRun(t, catch(`time.Time().Month(1)`), nil, nwrongArgs(0, -1, 1))

	// .Day
	expectRun(t, `param p1; return p1.Day()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Day()))
	expectRun(t, catch(`time.Time().Day(1)`), nil, nwrongArgs(0, -1, 1))

	// .Hour
	expectRun(t, `param p1; return p1.Hour()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Hour()))
	expectRun(t, catch(`time.Time().Hour(1)`), nil, nwrongArgs(0, -1, 1))

	// .Minute
	expectRun(t, `param p1; return p1.Minute()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Minute()))
	expectRun(t, catch(`time.Time().Minute(1)`), nil, nwrongArgs(0, -1, 1))

	// .Second
	expectRun(t, `param p1; return p1.Second()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Second()))
	expectRun(t, catch(`time.Time().Second(1)`), nil, nwrongArgs(0, -1, 1))

	// .Nanosecond
	expectRun(t, `param p1; return p1.Nanosecond()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Nanosecond()))
	expectRun(t, catch(`time.Time().Nanosecond(1)`), nil, nwrongArgs(0, -1, 1))

	// .Weekday
	expectRun(t, `param p1; return p1.Weekday()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.Weekday()))
	expectRun(t, catch(`time.Time().Weekday(1)`), nil, nwrongArgs(0, -1, 1))

	// .ISOWeek
	year, week := tm.ISOWeek()
	expectRun(t, `param p1; return p1.ISOWeek()`,
		newOpts().Args(&Time{Value: tm}), Map{"year": Int(year), "week": Int(week)})
	expectRun(t, catch(`time.Time().ISOWeek(1)`), nil, nwrongArgs(0, -1, 1))

	// .YearDay
	expectRun(t, `param p1; return p1.YearDay()`,
		newOpts().Args(&Time{Value: tm}), Int(tm.YearDay()))
	expectRun(t, catch(`time.Time().YearDay(1)`), nil, nwrongArgs(0, -1, 1))

	// .Location
	expectRun(t, `time := import("time"); return time.Time().Location()`, nil, &Location{Value: time.Time{}.Location()})
	expectRun(t, catch(`time.Time().Location(1)`), nil, nwrongArgs(0, -1, 1))

	// .Zone
	zone, offset := tm.Zone()
	expectRun(t, `param p1; return p1.Zone()`,
		newOpts().Args(&Time{Value: tm}), Map{"name": String(zone), "offset": Int(offset)})
	expectRun(t, catch(`time.Time().Zone(1)`), nil, nwrongArgs(0, -1, 1))
}

type illegalDur struct {
	ObjectImpl
	Value time.Duration
}

func (*illegalDur) String() string   { return "illegal" }
func (*illegalDur) TypeName() string { return "illegal" }

type Opts struct {
	global Object
	args   []Object
}

func newOpts() *Opts {
	return &Opts{}
}

func (o *Opts) Args(args ...Object) *Opts {
	o.args = args
	return o
}

func (o *Opts) Globals(g Object) *Opts {
	o.global = g
	return o
}

func expectRun(t *testing.T, script string, opts *Opts, expected Object) {
	t.Helper()
	if opts == nil {
		opts = newOpts()
	}
	mm := NewModuleMap()
	mm.AddBuiltinModule("time", Module)
	c := DefaultCompilerOptions
	c.ModuleMap = mm
	bc, err := Compile([]byte(script), c)
	require.NoError(t, err)
	ret, err := NewVM(bc).Run(opts.global, opts.args...)
	require.NoError(t, err)
	require.Equal(t, expected, ret)
}
