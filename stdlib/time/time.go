// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package time

import (
	"reflect"
	"time"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/registry"
	"github.com/ozanh/ugo/token"
)

func init() {
	registry.RegisterObjectConverter(reflect.TypeOf(time.Duration(0)),
		func(in interface{}) (interface{}, bool) {
			return ugo.Int(in.(time.Duration)), true
		},
	)

	registry.RegisterObjectConverter(reflect.TypeOf(time.Time{}),
		func(in interface{}) (interface{}, bool) {
			return &Time{Value: in.(time.Time)}, true
		},
	)
	registry.RegisterObjectConverter(reflect.TypeOf((*time.Time)(nil)),
		func(in interface{}) (interface{}, bool) {
			v := in.(*time.Time)
			if v == nil {
				return ugo.Undefined, true
			}
			return &Time{Value: *v}, true
		},
	)
	registry.RegisterAnyConverter(reflect.TypeOf((*Time)(nil)),
		func(in interface{}) (interface{}, bool) {
			return in.(*Time).Value, true
		},
	)

	registry.RegisterObjectConverter(reflect.TypeOf((*time.Location)(nil)),
		func(in interface{}) (interface{}, bool) {
			v := in.(*time.Location)
			if v == nil {
				return ugo.Undefined, true
			}
			return &Location{Value: v}, true
		},
	)
	registry.RegisterAnyConverter(reflect.TypeOf((*Location)(nil)),
		func(in interface{}) (interface{}, bool) {
			return in.(*Location).Value, true
		},
	)
}

// ugo:doc
// ## Types
// ### time
//
// Go Type
//
// ```go
// // Time represents time values and implements ugo.Object interface.
// type Time struct {
//   Value time.Time
// }
// ```

// Time represents time values and implements ugo.Object interface.
type Time struct {
	Value time.Time
}

var _ ugo.NameCallerObject = (*Time)(nil)

// TypeName implements ugo.Object interface.
func (*Time) TypeName() string {
	return "time"
}

// String implements ugo.Object interface.
func (o *Time) String() string {
	return o.Value.String()
}

// IsFalsy implements ugo.Object interface.
func (o *Time) IsFalsy() bool {
	return o.Value.IsZero()
}

// Equal implements ugo.Object interface.
func (o *Time) Equal(right ugo.Object) bool {
	if v, ok := right.(*Time); ok {
		return o.Value.Equal(v.Value)
	}
	return false
}

// CanCall implements ugo.Object interface.
func (*Time) CanCall() bool { return false }

// Call implements ugo.Object interface.
func (*Time) Call(args ...ugo.Object) (ugo.Object, error) {
	return nil, ugo.ErrNotCallable
}

// CanIterate implements ugo.Object interface.
func (*Time) CanIterate() bool { return false }

// Iterate implements ugo.Object interface.
func (*Time) Iterate() ugo.Iterator { return nil }

// ugo:doc
// #### Overloaded time Operators
//
// - `time + int` -> time
// - `time - int` -> time
// - `time - time` -> int
// - `time < time` -> bool
// - `time > time` -> bool
// - `time <= time` -> bool
// - `time >= time` -> bool
//
// Note that, `int` values as duration must be the right hand side operand.

// BinaryOp implements ugo.Object interface.
func (o *Time) BinaryOp(tok token.Token,
	right ugo.Object) (ugo.Object, error) {

	switch v := right.(type) {
	case ugo.Int:
		switch tok {
		case token.Add:
			return &Time{Value: o.Value.Add(time.Duration(v))}, nil
		case token.Sub:
			return &Time{Value: o.Value.Add(time.Duration(-v))}, nil
		}
	case *Time:
		switch tok {
		case token.Sub:
			return ugo.Int(o.Value.Sub(v.Value)), nil
		case token.Less:
			return ugo.Bool(o.Value.Before(v.Value)), nil
		case token.LessEq:
			return ugo.Bool(o.Value.Before(v.Value) || o.Value.Equal(v.Value)), nil
		case token.Greater:
			return ugo.Bool(o.Value.After(v.Value)), nil
		case token.GreaterEq:
			return ugo.Bool(o.Value.After(v.Value) || o.Value.Equal(v.Value)),
				nil
		}
	case *ugo.UndefinedType:
		switch tok {
		case token.Less, token.LessEq:
			return ugo.False, nil
		case token.Greater, token.GreaterEq:
			return ugo.True, nil
		}
	}
	return nil, ugo.NewOperandTypeError(
		tok.String(),
		o.TypeName(),
		right.TypeName())
}

// IndexSet implements ugo.Object interface.
func (*Time) IndexSet(_, _ ugo.Object) error { return ugo.ErrNotIndexAssignable }

// ugo:doc
// #### time Getters
//
// Deprecated: Use method call. These selectors will return a callable object in
// the future. See methods.
//
// Dynamically calculated getters for a time value are as follows:
//
// | Selector  | Return Type                                     |
// |:----------|:------------------------------------------------|
// |.Date      | {"year": int, "month": int, "day": int}         |
// |.Clock     | {"hour": int, "minute": int, "second": int}     |
// |.UTC       | time                                            |
// |.Unix      | int                                             |
// |.UnixNano  | int                                             |
// |.Year      | int                                             |
// |.Month     | int                                             |
// |.Day       | int                                             |
// |.Hour      | int                                             |
// |.Minute    | int                                             |
// |.Second    | int                                             |
// |.NanoSecond| int                                             |
// |.IsZero    | bool                                            |
// |.Local     | time                                            |
// |.Location  | location                                        |
// |.YearDay   | int                                             |
// |.Weekday   | int                                             |
// |.ISOWeek   | {"year": int, "week": int}                      |
// |.Zone      | {"name": string, "offset": int}                 |

// IndexGet implements ugo.Object interface.
func (o *Time) IndexGet(index ugo.Object) (ugo.Object, error) {
	v, ok := index.(ugo.String)
	if !ok {
		return ugo.Undefined, ugo.NewIndexTypeError("string", index.TypeName())
	}

	// For simplicity, we use method call for now. As getters are deprecated, we
	// will return callable object in the future here.

	switch v {
	case "Date", "Clock", "UTC", "Unix", "UnixNano", "Year", "Month", "Day",
		"Hour", "Minute", "Second", "Nanosecond", "IsZero", "Local", "Location",
		"YearDay", "Weekday", "ISOWeek", "Zone":
		return o.CallName(string(v), ugo.Call{})
	}
	return ugo.Undefined, nil
}

// ugo:doc
// #### time Methods
//
// | Method                               | Return Type                                 |
// |:-------------------------------------|:--------------------------------------------|
// |.Add(duration int)                    | time                                        |
// |.Sub(t2 time)                         | int                                         |
// |.AddDate(year int, month int, day int)| int                                         |
// |.After(t2 time)                       | bool                                        |
// |.Before(t2 time)                      | bool                                        |
// |.Format(layout string)                | string                                      |
// |.AppendFormat(b bytes, layout string) | bytes                                       |
// |.In(loc location)                     | time                                        |
// |.Round(duration int)                  | time                                        |
// |.Truncate(duration int)               | time                                        |
// |.Equal(t2 time)                       | bool                                        |
// |.Date()                               | {"year": int, "month": int, "day": int}     |
// |.Clock()                              | {"hour": int, "minute": int, "second": int} |
// |.UTC()                                | time                                        |
// |.Unix()                               | int                                         |
// |.UnixNano()                           | int                                         |
// |.Year()                               | int                                         |
// |.Month()                              | int                                         |
// |.Day()                                | int                                         |
// |.Hour()                               | int                                         |
// |.Minute()                             | int                                         |
// |.Second()                             | int                                         |
// |.NanoSecond()                         | int                                         |
// |.IsZero()                             | bool                                        |
// |.Local()                              | time                                        |
// |.Location()                           | location                                    |
// |.YearDay()                            | int                                         |
// |.Weekday()                            | int                                         |
// |.ISOWeek()                            | {"year": int, "week": int}                  |
// |.Zone()                               | {"name": string, "offset": int}             |

func (o *Time) CallName(name string, c ugo.Call) (ugo.Object, error) {
	fn, ok := methodTable[name]
	if !ok {
		return ugo.Undefined, ugo.ErrInvalidIndex.NewError(name)
	}
	return fn(o, &c)
}

var methodTable = map[string]func(*Time, *ugo.Call) (ugo.Object, error){
	"Add": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		d, ok := ugo.ToGoInt64(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "int", c.Get(0).TypeName())
		}
		return timeAdd(o, d), nil
	},
	"Sub": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		t2, ok := ToTime(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "time", c.Get(0).TypeName())
		}
		return timeSub(o, t2), nil
	},
	"AddDate": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(3); err != nil {
			return ugo.Undefined, err
		}
		year, ok := ugo.ToGoInt(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "int", c.Get(0).TypeName())
		}
		month, ok := ugo.ToGoInt(c.Get(1))
		if !ok {
			return newArgTypeErr("2nd", "int", c.Get(1).TypeName())
		}
		day, ok := ugo.ToGoInt(c.Get(2))
		if !ok {
			return newArgTypeErr("3rd", "int", c.Get(2).TypeName())
		}
		return timeAddDate(o, year, month, day), nil
	},
	"After": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		t2, ok := ToTime(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "time", c.Get(0).TypeName())
		}
		return timeAfter(o, t2), nil
	},
	"Before": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		t2, ok := ToTime(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "time", c.Get(0).TypeName())
		}
		return timeBefore(o, t2), nil
	},
	"Format": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		format, ok := ugo.ToGoString(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "string", c.Get(0).TypeName())
		}
		return timeFormat(o, format), nil
	},
	"AppendFormat": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(2); err != nil {
			return ugo.Undefined, err
		}
		b, ok := ugo.ToGoByteSlice(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "bytes", c.Get(0).TypeName())
		}
		format, ok := ugo.ToGoString(c.Get(1))
		if !ok {
			return newArgTypeErr("2nd", "string", c.Get(1).TypeName())
		}
		return timeAppendFormat(o, b, format), nil
	},
	"In": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		loc, ok := ToLocation(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "location", c.Get(0).TypeName())
		}
		return timeIn(o, loc), nil
	},
	"Round": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		d, ok := ugo.ToGoInt64(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "int", c.Get(0).TypeName())
		}
		return timeRound(o, d), nil
	},
	"Truncate": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		d, ok := ugo.ToGoInt64(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "int", c.Get(0).TypeName())
		}
		return timeTruncate(o, d), nil
	},
	"Equal": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(1); err != nil {
			return ugo.Undefined, err
		}
		t2, ok := ToTime(c.Get(0))
		if !ok {
			return newArgTypeErr("1st", "time", c.Get(0).TypeName())
		}
		return timeEqual(o, t2), nil
	},
	"Date": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		y, m, d := o.Value.Date()
		return ugo.Map{"year": ugo.Int(y), "month": ugo.Int(m),
			"day": ugo.Int(d)}, nil
	},
	"Clock": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		h, m, s := o.Value.Clock()
		return ugo.Map{"hour": ugo.Int(h), "minute": ugo.Int(m),
			"second": ugo.Int(s)}, nil
	},
	"UTC": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return &Time{Value: o.Value.UTC()}, nil
	},
	"Unix": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Unix()), nil
	},
	"UnixNano": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.UnixNano()), nil
	},
	"Year": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Year()), nil
	},
	"Month": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Month()), nil
	},
	"Day": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Day()), nil
	},
	"Hour": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Hour()), nil
	},
	"Minute": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Minute()), nil
	},
	"Second": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Second()), nil
	},
	"Nanosecond": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Nanosecond()), nil
	},
	"IsZero": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Bool(o.Value.IsZero()), nil
	},
	"Local": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return &Time{Value: o.Value.Local()}, nil
	},
	"Location": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return &Location{Value: o.Value.Location()}, nil
	},
	"YearDay": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.YearDay()), nil
	},
	"Weekday": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		return ugo.Int(o.Value.Weekday()), nil
	},
	"ISOWeek": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		y, w := o.Value.ISOWeek()
		return ugo.Map{"year": ugo.Int(y), "week": ugo.Int(w)}, nil
	},
	"Zone": func(o *Time, c *ugo.Call) (ugo.Object, error) {
		if err := c.CheckLen(0); err != nil {
			return ugo.Undefined, err
		}
		name, offset := o.Value.Zone()
		return ugo.Map{"name": ugo.String(name), "offset": ugo.Int(offset)}, nil
	},
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (o *Time) MarshalBinary() ([]byte, error) {
	return o.Value.MarshalBinary()
}

// MarshalJSON implements json.JSONMarshaler interface.
func (o *Time) MarshalJSON() ([]byte, error) {
	return o.Value.MarshalJSON()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (o *Time) UnmarshalBinary(data []byte) error {
	var t time.Time
	if err := t.UnmarshalBinary(data); err != nil {
		return err
	}
	o.Value = t
	return nil
}

// UnmarshalJSON implements json.JSONUnmarshaler interface.
func (o *Time) UnmarshalJSON(data []byte) error {
	var t time.Time
	if err := t.UnmarshalJSON(data); err != nil {
		return err
	}
	o.Value = t
	return nil
}

func timeAdd(t *Time, duration int64) ugo.Object {
	return &Time{Value: t.Value.Add(time.Duration(duration))}
}

func timeSub(t1, t2 *Time) ugo.Object {
	return ugo.Int(t1.Value.Sub(t2.Value))
}

func timeAddDate(t *Time, years, months, days int) ugo.Object {
	return &Time{Value: t.Value.AddDate(years, months, days)}
}

func timeAfter(t1, t2 *Time) ugo.Object {
	return ugo.Bool(t1.Value.After(t2.Value))
}

func timeBefore(t1, t2 *Time) ugo.Object {
	return ugo.Bool(t1.Value.Before(t2.Value))
}

func timeFormat(t *Time, layout string) ugo.Object {
	return ugo.String(t.Value.Format(layout))
}

func timeAppendFormat(t *Time, b []byte, layout string) ugo.Object {
	return ugo.Bytes(t.Value.AppendFormat(b, layout))
}

func timeIn(t *Time, loc *Location) ugo.Object {
	return &Time{Value: t.Value.In(loc.Value)}
}

func timeRound(t *Time, duration int64) ugo.Object {
	return &Time{Value: t.Value.Round(time.Duration(duration))}
}

func timeTruncate(t *Time, duration int64) ugo.Object {
	return &Time{Value: t.Value.Truncate(time.Duration(duration))}
}

func timeEqual(t1, t2 *Time) ugo.Object {
	return ugo.Bool(t1.Value.Equal(t2.Value))
}

func newArgTypeErr(pos, want, got string) (ugo.Object, error) {
	return ugo.Undefined, ugo.NewArgumentTypeError(pos, want, got)
}
