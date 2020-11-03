package time_test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ozanh/ugo"
	. "github.com/ozanh/ugo/stdlib/time"
)

func TestModuleTypes(t *testing.T) {
	l := &Location{Location: time.UTC}
	require.Equal(t, "location", l.TypeName())
	require.False(t, l.IsFalsy())
	require.Equal(t, "UTC", l.String())
	require.True(t, (&Location{}).Equal(&Location{}))
	require.True(t, (&Location{}).Equal(ugo.String("UTC")))
	require.False(t, (&Location{}).Equal(ugo.Int(0)))
	require.False(t, l.CanCall())
	require.False(t, l.CanIterate())
	require.Nil(t, l.Iterate())
	require.Equal(t, ugo.ErrNotIndexAssignable, l.IndexSet(nil, nil))
	_, err := l.IndexGet(nil)
	require.Equal(t, ugo.ErrNotIndexable, err)

	tm := Time{}
	require.Equal(t, "time", tm.TypeName())
	require.True(t, tm.IsFalsy())
	require.NotEmpty(t, tm.String())
	require.True(t, tm.Equal(Time{}))
	require.False(t, tm.Equal(ugo.Int(0)))
	require.False(t, tm.CanCall())
	require.False(t, tm.CanIterate())
	require.Nil(t, tm.Iterate())
	require.Equal(t, ugo.ErrNotIndexAssignable, tm.IndexSet(nil, nil))
	r, err := tm.IndexGet(ugo.String(""))
	require.NoError(t, err)
	require.Equal(t, ugo.Undefined, r)

	now := time.Now()
	tm2 := Time(now)
	require.False(t, tm2.IsFalsy())
	require.Equal(t, now.String(), tm2.String())

	var b bytes.Buffer
	err = gob.NewEncoder(&b).Encode(tm2)
	require.NoError(t, err)
	var tm3 Time
	err = gob.NewDecoder(&b).Decode(&tm3)
	require.NoError(t, err)
	require.Equal(t, time.Time(tm2).Format(time.RFC3339Nano),
		time.Time(tm3).Format(time.RFC3339Nano))
}

func TestModuleMonthWeekday(t *testing.T) {
	f := Module["MonthString"].(*ugo.Function)
	_, err := f.Call()
	require.Error(t, err)
	_, err = f.Call(ugo.String(""))
	require.Error(t, err)

	for i := 1; i <= 12; i++ {
		require.Contains(t, Module, time.Month(i).String())
		require.Equal(t, ugo.Int(i), Module[time.Month(i).String()])

		r, err := f.Call(ugo.Int(i))
		require.NoError(t, err)
		require.EqualValues(t, time.Month(i).String(), r)
	}

	f = Module["WeekdayString"].(*ugo.Function)
	_, err = f.Call()
	require.Error(t, err)
	_, err = f.Call(ugo.String(""))
	require.Error(t, err)
	for i := 0; i <= 6; i++ {
		require.Contains(t, Module, time.Weekday(i).String())
		require.Equal(t, ugo.Int(i), Module[time.Weekday(i).String()])

		r, err := f.Call(ugo.Int(i))
		require.NoError(t, err)
		require.EqualValues(t, time.Weekday(i).String(), r)
	}
}

func TestModuleFormats(t *testing.T) {
	require.Equal(t, Module["ANSIC"], ugo.String(time.ANSIC))
	require.Equal(t, Module["UnixDate"], ugo.String(time.UnixDate))
	require.Equal(t, Module["RubyDate"], ugo.String(time.RubyDate))
	require.Equal(t, Module["RFC822"], ugo.String(time.RFC822))
	require.Equal(t, Module["RFC822Z"], ugo.String(time.RFC822Z))
	require.Equal(t, Module["RFC850"], ugo.String(time.RFC850))
	require.Equal(t, Module["RFC1123"], ugo.String(time.RFC1123))
	require.Equal(t, Module["RFC1123Z"], ugo.String(time.RFC1123Z))
	require.Equal(t, Module["RFC3339"], ugo.String(time.RFC3339))
	require.Equal(t, Module["RFC3339Nano"], ugo.String(time.RFC3339Nano))
	require.Equal(t, Module["Kitchen"], ugo.String(time.Kitchen))
	require.Equal(t, Module["Stamp"], ugo.String(time.Stamp))
	require.Equal(t, Module["StampMilli"], ugo.String(time.StampMilli))
	require.Equal(t, Module["StampMicro"], ugo.String(time.StampMicro))
	require.Equal(t, Module["StampNano"], ugo.String(time.StampNano))
}

func TestModuleDuration(t *testing.T) {
	require.Equal(t, Module["Nanosecond"], ugo.Int(time.Nanosecond))
	require.Equal(t, Module["Microsecond"], ugo.Int(time.Microsecond))
	require.Equal(t, Module["Millisecond"], ugo.Int(time.Millisecond))
	require.Equal(t, Module["Second"], ugo.Int(time.Second))
	require.Equal(t, Module["Minute"], ugo.Int(time.Minute))
	require.Equal(t, Module["Hour"], ugo.Int(time.Hour))

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
	durToString := Module["DurationString"].(*ugo.Function)
	_, err := durToString.Call()
	require.Error(t, err)

	durParse := Module["ParseDuration"].(*ugo.Function)
	_, err = durParse.Call()
	require.Error(t, err)
	_, err = durParse.Call(ugo.String(""))
	require.Error(t, err)
	_, err = durParse.Call(ugo.Int(0))
	require.Error(t, err)

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
				f := Module["Duration"+fn].(*ugo.Function)
				ret, err := f.Call(ugo.Int(tC.dur))
				require.NoError(t, err)
				expect := goFnMap[fn](tC.dur)
				require.EqualValues(t, expect, ret)

				// test illegal type
				_, err = f.Call(ugo.Uint(tC.dur))
				require.Error(t, err)
				// test no arg
				_, err = f.Call()
				require.Error(t, err)

				// test to string
				s, err := durToString.Call(ugo.Int(tC.dur))
				require.NoError(t, err)
				require.EqualValues(t, tC.dur.String(), s)

				// test to string errors
				_, err = durToString.Call(ugo.Uint(tC.dur))
				require.Error(t, err)

				// test parse
				d, err := durParse.Call(s)
				require.NoError(t, err)
				ed, err := time.ParseDuration(tC.dur.String())
				require.NoError(t, err)
				require.EqualValues(t, ed, d)
			})
		}
	}

	durRound := Module["DurationRound"].(*ugo.Function)
	r, err := durRound.Call(ugo.Int(time.Second+time.Millisecond),
		ugo.Int(time.Second))
	require.NoError(t, err)
	require.EqualValues(t, time.Second, r)
	_, err = durRound.Call(ugo.Int(0))
	require.Error(t, err)
	_, err = durRound.Call(ugo.String(""), ugo.Int(0))
	require.Error(t, err)
	_, err = durRound.Call(ugo.Int(0), ugo.String(""))
	require.Error(t, err)

	durTruncate := Module["DurationTruncate"].(*ugo.Function)
	r, err = durTruncate.Call(ugo.Int(time.Second+5*time.Millisecond),
		ugo.Int(2*time.Millisecond))
	require.NoError(t, err)
	require.EqualValues(t, time.Second+4*time.Millisecond, r)
	_, err = durTruncate.Call(ugo.Int(0))
	require.Error(t, err)
	_, err = durTruncate.Call(ugo.String(""), ugo.Int(0))
	require.Error(t, err)
	_, err = durTruncate.Call(ugo.Int(0), ugo.String(""))
	require.Error(t, err)
}

func TestModuleLocation(t *testing.T) {
	fixedZone := Module["FixedZone"].(*ugo.Function)
	r, err := fixedZone.Call(ugo.String("Ankara"), ugo.Int(3*60*60))
	require.NoError(t, err)
	require.Equal(t, "Ankara", r.String())

	_, err = fixedZone.Call(ugo.String("Ankara"))
	require.Error(t, err)
	_, err = fixedZone.Call(ugo.String("Ankara"), ugo.Uint(0))
	require.Error(t, err)
	_, err = fixedZone.Call(ugo.Int(0), ugo.Int(0))
	require.Error(t, err)
	_, err = fixedZone.Call()
	require.Error(t, err)

	loadLocation := Module["LoadLocation"].(*ugo.Function)
	r, err = loadLocation.Call(ugo.String("Europe/Istanbul"))
	require.NoError(t, err)
	require.Equal(t, "Europe/Istanbul", r.String())
	r, err = loadLocation.Call(ugo.String(""))
	require.NoError(t, err)
	require.Equal(t, "UTC", r.String())
	_, err = loadLocation.Call()
	require.Error(t, err)
	_, err = loadLocation.Call(ugo.Int(0))
	require.Error(t, err)
	_, err = loadLocation.Call(ugo.String("invalid"))
	require.Error(t, err)

	isLocation := Module["IsLocation"].(*ugo.Function)
	r, err = isLocation.Call(&Location{Location: time.Local})
	require.NoError(t, err)
	require.EqualValues(t, true, r)
	r, err = isLocation.Call(ugo.Int(0))
	require.NoError(t, err)
	require.EqualValues(t, false, r)
	_, err = isLocation.Call(ugo.Int(0), ugo.Int(0))
	require.Error(t, err)
	_, err = isLocation.Call()
	require.Error(t, err)
}

func TestModuleTime(t *testing.T) {
	now := time.Now()

	require.EqualValues(t, now, Time(now))
	require.Equal(t, now.String(), Time(now).String())

	zTime := Module["Time"].(*ugo.Function)
	r, err := zTime.Call()
	require.NoError(t, err)
	require.True(t, time.Time(r.(Time)).IsZero())
	_, err = zTime.Call(ugo.String(""))
	require.Error(t, err)

	since := Module["Since"].(*ugo.Function)
	r, err = since.Call(Time(now))
	require.NoError(t, err)
	require.GreaterOrEqual(t, int64(r.(ugo.Int)), int64(0))
	_, err = since.Call()
	require.Error(t, err)
	_, err = since.Call(ugo.String(""))
	require.Error(t, err)

	until := Module["Until"].(*ugo.Function)
	r, err = until.Call(Time(now))
	require.NoError(t, err)
	require.LessOrEqual(t, int64(r.(ugo.Int)), int64(0))
	_, err = until.Call()
	require.Error(t, err)
	_, err = until.Call(ugo.String(""))
	require.Error(t, err)

	date := Module["Date"].(*ugo.Function)
	r, err = date.Call(ugo.Int(2020), ugo.Int(11), ugo.Int(8),
		ugo.Int(1), ugo.Int(2), ugo.Int(3), ugo.Int(4),
		&Location{Location: time.Local})
	require.NoError(t, err)
	require.EqualValues(t, time.Date(2020, 11, 8, 1, 2, 3, 4, time.Local), r)
	r, err = date.Call(ugo.Int(2020), ugo.Int(11), ugo.Int(8))
	require.NoError(t, err)
	require.EqualValues(t, time.Date(2020, 11, 8, 0, 0, 0, 0, time.Local), r)

	nowf := Module["Now"].(*ugo.Function)
	r, err = nowf.Call()
	require.NoError(t, err)
	require.False(t, time.Time(r.(Time)).IsZero())
	_, err = nowf.Call(ugo.Int(0))
	require.Error(t, err)

	RFC3339Nano := Module["RFC3339Nano"]
	parse := Module["Parse"].(*ugo.Function)
	r, err = parse.Call(RFC3339Nano, ugo.String(now.Format(time.RFC3339Nano)))
	require.NoError(t, err)
	require.Equal(t, now.Format(time.RFC3339Nano),
		time.Time(r.(Time)).Format(time.RFC3339Nano))

	r, err = parse.Call(RFC3339Nano, ugo.String(now.Format(time.RFC3339Nano)),
		&Location{Location: time.Local})
	require.NoError(t, err)
	require.Equal(t, now.Format(time.RFC3339Nano),
		time.Time(r.(Time)).Format(time.RFC3339Nano))

	_, err = parse.Call()
	require.Error(t, err)

	unix := Module["Unix"].(*ugo.Function)
	r, err = unix.Call(ugo.Int(now.Unix()))
	require.NoError(t, err)
	require.Equal(t, time.Unix(now.Unix(), 0), time.Time(r.(Time)))
	r, err = unix.Call(ugo.Int(now.Unix()), ugo.Int(1))
	require.NoError(t, err)
	require.Equal(t, time.Unix(now.Unix(), 1), time.Time(r.(Time)))
	_, err = unix.Call()
	require.Error(t, err)

	add := Module["Add"].(*ugo.Function)
	r, err = add.Call(Time(now), ugo.Int(time.Second))
	require.NoError(t, err)
	require.Equal(t, now.Add(time.Second), time.Time(r.(Time)))
	_, err = add.Call(Time(now))
	require.Error(t, err)
	_, err = add.Call(Time(now), Time(now))
	require.Error(t, err)
	_, err = add.Call()
	require.Error(t, err)

	sub := Module["Sub"].(*ugo.Function)
	r, err = sub.Call(Time(now), Time(now.Add(-time.Hour)))
	require.NoError(t, err)
	require.EqualValues(t, time.Hour, r.(ugo.Int))
	_, err = sub.Call(Time(now))
	require.Error(t, err)
	_, err = sub.Call(Time(now), ugo.Int(0))
	require.Error(t, err)
	_, err = sub.Call()
	require.Error(t, err)

	addDate := Module["AddDate"].(*ugo.Function)
	r, err = addDate.Call(Time(now),
		ugo.Int(1), ugo.Int(2), ugo.Int(3))
	require.NoError(t, err)
	require.EqualValues(t, now.AddDate(1, 2, 3), r.(Time))
	_, err = addDate.Call(Time(now))
	require.Error(t, err)
	_, err = addDate.Call(Time(now), ugo.Int(0))
	require.Error(t, err)
	_, err = addDate.Call()
	require.Error(t, err)

	after := Module["After"].(*ugo.Function)
	r, err = after.Call(Time(now), Time(now.Add(time.Hour)))
	require.NoError(t, err)
	require.EqualValues(t, false, r)
	r, err = after.Call(Time(now), Time(now.Add(-time.Hour)))
	require.NoError(t, err)
	require.EqualValues(t, true, r)
	_, err = after.Call(Time(now), ugo.Int(0))
	require.Error(t, err)
	_, err = after.Call(Time(now))
	require.Error(t, err)
	_, err = after.Call()
	require.Error(t, err)

	before := Module["Before"].(*ugo.Function)
	r, err = before.Call(Time(now), Time(now.Add(time.Hour)))
	require.NoError(t, err)
	require.EqualValues(t, true, r)
	r, err = before.Call(Time(now), Time(now.Add(-time.Hour)))
	require.NoError(t, err)
	require.EqualValues(t, false, r)
	_, err = before.Call(Time(now), ugo.Int(0))
	require.Error(t, err)
	_, err = before.Call(Time(now))
	require.Error(t, err)
	_, err = before.Call()
	require.Error(t, err)

	appendFormat := Module["AppendFormat"].(*ugo.Function)
	b := make(ugo.Bytes, 100)
	r, err = appendFormat.Call(Time(now), b, RFC3339Nano)
	require.NoError(t, err)
	require.EqualValues(t,
		now.AppendFormat(make([]byte, 100), time.RFC3339Nano), r)
	_, err = appendFormat.Call(Time(now), b)
	require.Error(t, err)
	_, err = appendFormat.Call(Time(now))
	require.Error(t, err)
	_, err = appendFormat.Call()
	require.Error(t, err)

	format := Module["Format"].(*ugo.Function)
	r, err = format.Call(Time(now), RFC3339Nano)
	require.NoError(t, err)
	require.EqualValues(t, now.Format(time.RFC3339Nano), r)
	_, err = format.Call(Time(now))
	require.Error(t, err)
	_, err = format.Call()
	require.Error(t, err)

	timeIn := Module["In"].(*ugo.Function)
	r, err = timeIn.Call(Time(now), &Location{Location: time.Local})
	require.NoError(t, err)
	require.False(t, time.Time(r.(Time)).IsZero())
	_, err = timeIn.Call(Time(now))
	require.Error(t, err)
	_, err = timeIn.Call()
	require.Error(t, err)

	round := Module["Round"].(*ugo.Function)
	r, err = round.Call(Time(now), ugo.Int(time.Second))
	require.NoError(t, err)
	require.EqualValues(t, now.Round(time.Second), r)
	_, err = round.Call(Time(now))
	require.Error(t, err)
	_, err = round.Call()
	require.Error(t, err)

	truncate := Module["Truncate"].(*ugo.Function)
	r, err = truncate.Call(Time(now), ugo.Int(time.Hour))
	require.NoError(t, err)
	require.EqualValues(t, now.Truncate(time.Hour), r)
	_, err = truncate.Call(Time(now))
	require.Error(t, err)
	_, err = truncate.Call()
	require.Error(t, err)

	isTime := Module["IsTime"].(*ugo.Function)
	r, err = isTime.Call(Time(now))
	require.NoError(t, err)
	require.EqualValues(t, true, r)
	r, err = isTime.Call(ugo.Int(0))
	require.NoError(t, err)
	require.EqualValues(t, false, r)
	_, err = isTime.Call(ugo.Int(0), ugo.Int(0))
	require.Error(t, err)
	_, err = isTime.Call()
	require.Error(t, err)

	y, m, d := now.Date()
	testTimeSelector(t, Time(now), "Date",
		ugo.Map{"year": ugo.Int(y), "month": ugo.Int(m), "day": ugo.Int(d)})
	h, min, s := now.Clock()
	testTimeSelector(t, Time(now), "Clock",
		ugo.Map{"hour": ugo.Int(h), "minute": ugo.Int(min), "second": ugo.Int(s)})
	testTimeSelector(t, Time(now), "UTC", Time(now.UTC()))
	testTimeSelector(t, Time(now), "Unix", ugo.Int(now.Unix()))
	testTimeSelector(t, Time(now), "UnixNano", ugo.Int(now.UnixNano()))
	testTimeSelector(t, Time(now), "Year", ugo.Int(now.Year()))
	testTimeSelector(t, Time(now), "Month", ugo.Int(now.Month()))
	testTimeSelector(t, Time(now), "Day", ugo.Int(now.Day()))
	testTimeSelector(t, Time(now), "Hour", ugo.Int(now.Hour()))
	testTimeSelector(t, Time(now), "Minute", ugo.Int(now.Minute()))
	testTimeSelector(t, Time(now), "Second", ugo.Int(now.Second()))
	testTimeSelector(t, Time(now), "Nanosecond", ugo.Int(now.Nanosecond()))
	testTimeSelector(t, Time(now), "IsZero", ugo.Bool(false))
	testTimeSelector(t, Time(now), "Local", Time(now.Local()))
	testTimeSelector(t, Time(now), "Location",
		&Location{Location: now.Location()})
	testTimeSelector(t, Time(now), "YearDay", ugo.Int(now.YearDay()))
	testTimeSelector(t, Time(now), "Weekday", ugo.Int(now.Weekday()))
	y, w := now.ISOWeek()
	testTimeSelector(t, Time(now), "ISOWeek",
		ugo.Map{"year": ugo.Int(y), "week": ugo.Int(w)})
	name, offset := now.Zone()
	testTimeSelector(t, Time(now), "Zone",
		ugo.Map{"name": ugo.String(name), "offset": ugo.Int(offset)})
	testTimeSelector(t, Time(now), "XYZ", ugo.Undefined)
}

func testTimeSelector(t *testing.T, tm ugo.Object,
	selector string, expected ugo.Object) {
	t.Helper()
	v, err := tm.IndexGet(ugo.String(selector))
	require.NoError(t, err)
	require.Equal(t, expected, v)
}

func TestScript(t *testing.T) {
	expectRun(t, `import("time")`, nil, ugo.Undefined)
	expectRun(t, `mod := import("time"); return mod.__module_name__`,
		nil, ugo.String("time"))

	tm := time.Now()
	expectRun(t, `
	param p1; time := import("time"); return time.Format(p1, time.RFC3339Nano)`,
		newOpts().Args(Time(tm)), ugo.String(tm.Format(time.RFC3339Nano)))
	expectRun(t, `param p1; return p1.UnixNano`,
		newOpts().Args(Time(tm)), ugo.Int(tm.UnixNano()))

	expectRun(t, `
	param p1
	time := import("time")
	try {
		time.Sleep(time.Millisecond)
	} finally {
		dur := time.Since(p1)
		return dur > 0 ? true: false 
	}
	`, newOpts().Args(Time(tm)), ugo.True)

	expectRun(t, `return import("time").IsTime(0)`, nil, ugo.False)
	expectRun(t, `param p1; time := import("time"); return time.IsTime(p1)`,
		newOpts().Args(Time(tm)), ugo.True)
	expectRun(t, `time := import("time"); return time.IsTime(time.Now())`,
		nil, ugo.True)
	expectRun(t, `
	time := import("time")
	return time.IsLocation(time.FixedZone("abc", 3*60*60))`, nil, ugo.True)
	expectRun(t, `param p1; return p1==p1`, newOpts().Args(Time(tm)), ugo.True)
	expectRun(t, `param p1; time := import("time"); return time.Now()==p1`,
		newOpts().Args(Time(tm)), ugo.False)
	expectRun(t, `param p1; time := import("time"); return time.Now()>=p1`,
		newOpts().Args(Time(tm)), ugo.True)
	expectRun(t, `param p1; time := import("time"); return time.Now()<p1`,
		newOpts().Args(Time(tm)), ugo.False)
	expectRun(t, `param p1; time := import("time"); return time.Now()>p1`,
		newOpts().Args(Time(tm)), ugo.True)
	expectRun(t, `time := import("time"); return (time.Now()+time.Second)>=time.Now()`,
		nil, ugo.True)
	expectRun(t, `time := import("time"); return (time.Now()+time.Second)<=time.Now()`,
		nil, ugo.False)
	expectRun(t, `time := import("time"); return (time.Now()-10*time.Second)<=time.Now()`,
		nil, ugo.True)
	expectRun(t, `time := import("time"); return time.Now() == undefined`,
		nil, ugo.False)
	expectRun(t, `time := import("time"); return time.Now() > undefined`,
		nil, ugo.True)
	expectRun(t, `time := import("time"); return time.Now() >= undefined`,
		nil, ugo.True)
	expectRun(t, `time := import("time"); return time.Now() < undefined`,
		nil, ugo.False)
	expectRun(t, `time := import("time"); return time.Now() <= undefined`,
		nil, ugo.False)
	expectRun(t, `
	time := import("time")
	t1 := time.Now()
	t2 := t1 + time.Second
	return t2 - t1
	`, nil, ugo.Int(time.Second))
}

type Opts struct {
	global ugo.Object
	args   []ugo.Object
}

func newOpts() *Opts {
	return &Opts{}
}

func (o *Opts) Args(args ...ugo.Object) *Opts {
	o.args = args
	return o
}

func (o *Opts) Globals(g ugo.Object) *Opts {
	o.global = g
	return o
}

func expectRun(t *testing.T, script string, opts *Opts, expected ugo.Object) {
	t.Helper()
	if opts == nil {
		opts = newOpts()
	}
	mm := ugo.NewModuleMap()
	mm.AddBuiltinModule("time", Module)
	c := ugo.DefaultCompilerOptions
	c.ModuleMap = mm
	bc, err := ugo.Compile([]byte(script), c)
	require.NoError(t, err)
	ret, err := ugo.NewVM(bc).Run(opts.global, opts.args...)
	require.NoError(t, err)
	require.Equal(t, expected, ret)
}
