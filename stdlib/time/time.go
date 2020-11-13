// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package time

import (
	"time"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/token"
)

// ugo:doc
// ## Types
// ### time
//
// Go Type
//
// ```go
// // Time represents time values and implements ugo.Object interface.
// type Time time.Time
// ```

// Time represents time values and implements ugo.Object interface.
type Time time.Time

// TypeName implements ugo.Object interface.
func (Time) TypeName() string {
	return "time"
}

// String implements ugo.Object interface.
func (o Time) String() string {
	return time.Time(o).String()
}

// IsFalsy implements ugo.Object interface.
func (o Time) IsFalsy() bool {
	return bool(time.Time(o).IsZero())
}

// Equal implements ugo.Object interface.
func (o Time) Equal(right ugo.Object) bool {
	if v, ok := right.(Time); ok {
		return time.Time(o).Equal(time.Time(v))
	}
	return false
}

// CanCall implements ugo.Object interface.
func (Time) CanCall() bool { return false }

// Call implements ugo.Object interface.
func (Time) Call(args ...ugo.Object) (ugo.Object, error) {
	return nil, ugo.ErrNotCallable
}

// CanIterate implements ugo.Object interface.
func (Time) CanIterate() bool { return false }

// Iterate implements ugo.Object interface.
func (Time) Iterate() ugo.Iterator { return nil }

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
func (o Time) BinaryOp(tok token.Token,
	right ugo.Object) (ugo.Object, error) {

	switch v := right.(type) {
	case ugo.Int:
		switch tok {
		case token.Add:
			return Time(time.Time(o).Add(time.Duration(v))), nil
		case token.Sub:
			return Time(time.Time(o).Add(time.Duration(-v))), nil
		}
	case Time:
		switch tok {
		case token.Sub:
			return ugo.Int(time.Time(o).Sub(time.Time(v))), nil
		case token.Less:
			return ugo.Bool(time.Time(o).Before(time.Time(v))), nil
		case token.LessEq:
			return ugo.Bool(time.Time(o).Before(time.Time(v)) || o.Equal(v)),
				nil
		case token.Greater:
			return ugo.Bool(time.Time(o).After(time.Time(v))), nil
		case token.GreaterEq:
			return ugo.Bool(time.Time(o).After(time.Time(v)) || o.Equal(v)),
				nil
		}
	}
	if right == ugo.Undefined {
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
func (Time) IndexSet(_, _ ugo.Object) error { return ugo.ErrNotIndexAssignable }

// ugo:doc
// #### time Getters
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
func (o Time) IndexGet(index ugo.Object) (ugo.Object, error) {
	v, ok := index.(ugo.String)
	if !ok {
		return nil, ugo.NewIndexTypeError("string", index.TypeName())
	}
	switch v {
	case "Date":
		y, m, d := time.Time(o).Date()
		return ugo.Map{"year": ugo.Int(y), "month": ugo.Int(m),
			"day": ugo.Int(d)}, nil
	case "Clock":
		h, m, s := time.Time(o).Clock()
		return ugo.Map{"hour": ugo.Int(h), "minute": ugo.Int(m),
			"second": ugo.Int(s)}, nil
	case "UTC":
		return Time(time.Time(o).UTC()), nil
	case "Unix":
		return ugo.Int(time.Time(o).Unix()), nil
	case "UnixNano":
		return ugo.Int(time.Time(o).UnixNano()), nil
	case "Year":
		return ugo.Int(time.Time(o).Year()), nil
	case "Month":
		return ugo.Int(time.Time(o).Month()), nil
	case "Day":
		return ugo.Int(time.Time(o).Day()), nil
	case "Hour":
		return ugo.Int(time.Time(o).Hour()), nil
	case "Minute":
		return ugo.Int(time.Time(o).Minute()), nil
	case "Second":
		return ugo.Int(time.Time(o).Second()), nil
	case "Nanosecond":
		return ugo.Int(time.Time(o).Nanosecond()), nil
	case "IsZero":
		return ugo.Bool(time.Time(o).IsZero()), nil
	case "Local":
		return Time(time.Time(o).Local()), nil
	case "Location":
		return &Location{Location: time.Time(o).Location()}, nil
	case "YearDay":
		return ugo.Int(time.Time(o).YearDay()), nil
	case "Weekday":
		return ugo.Int(time.Time(o).Weekday()), nil
	case "ISOWeek":
		y, w := time.Time(o).ISOWeek()
		return ugo.Map{"year": ugo.Int(y), "week": ugo.Int(w)}, nil
	case "Zone":
		name, offset := time.Time(o).Zone()
		return ugo.Map{"name": ugo.String(name), "offset": ugo.Int(offset)}, nil
	}
	return ugo.Undefined, nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (o Time) MarshalBinary() ([]byte, error) {
	return time.Time(o).MarshalBinary()
}

// MarshalJSON implements json.JSONMarshaler interface.
func (o Time) MarshalJSON() ([]byte, error) {
	return time.Time(o).MarshalJSON()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (o *Time) UnmarshalBinary(data []byte) error {
	return (*time.Time)(o).UnmarshalBinary(data)
}

// UnmarshalJSON implements json.JSONUnmarshaler interface.
func (o *Time) UnmarshalJSON(data []byte) error {
	return (*time.Time)(o).UnmarshalJSON(data)
}
