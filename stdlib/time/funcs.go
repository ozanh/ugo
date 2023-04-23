package time

import (
	"time"

	"github.com/ozanh/ugo"
)

//go:generate go run ../../cmd/mkcallable -output zfuncs.go funcs.go

//ugo:callable:convert *Location ToLocation
//ugo:callable:convert *Time ToTime

// ToLocation will try to convert given ugo.Object to *Location value.
func ToLocation(o ugo.Object) (ret *Location, ok bool) {
	if v, isString := o.(ugo.String); isString {
		var err error
		o, err = loadLocationFunc(string(v))
		if err != nil {
			return
		}
	}
	ret, ok = o.(*Location)
	return
}

// ToTime will try to convert given given ugo.Object to *Time value.
func ToTime(o ugo.Object) (ret *Time, ok bool) {
	switch o := o.(type) {
	case *Time:
		ret, ok = o, true
	case ugo.Int:
		v := time.Unix(int64(o), 0)
		ret, ok = &Time{Value: v}, true
	case ugo.String:
		v, err := time.Parse(time.RFC3339Nano, string(o))
		if err != nil {
			v, err = time.Parse(time.RFC3339, string(o))
		}
		if err == nil {
			ret, ok = &Time{Value: v}, true
		}
	}
	return
}

// Since, Until
//
//ugo:callable funcPTRO(t *Time) (ret ugo.Object)

// Add, Round, Truncate
//
//ugo:callable funcPTi64RO(t *Time, d int64) (ret ugo.Object)

// Sub, After, Before
//
//ugo:callable funcPTTRO(t1 *Time, t2 *Time) (ret ugo.Object)

// AddDate
//
//ugo:callable funcPTiiiRO(t *Time, i1 int, i2 int, i3 int) (ret ugo.Object)

// Format
//
//ugo:callable funcPTsRO(t *Time, s string) (ret ugo.Object)

// AppendFormat
//
//ugo:callable funcPTb2sRO(t *Time, b []byte, s string) (ret ugo.Object)

// In
//
//ugo:callable funcPTLRO(t *Time, loc *Location) (ret ugo.Object)
