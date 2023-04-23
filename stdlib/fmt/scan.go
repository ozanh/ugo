package fmt

import (
	"reflect"
	"strconv"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/registry"
)

func init() {
	registry.RegisterAnyConverter(reflect.TypeOf((*scanArg)(nil)),
		func(in interface{}) (interface{}, bool) {
			sa := in.(*scanArg)
			if sa.argValue != nil {
				return sa.Arg(), true
			}
			return ugo.Undefined, false
		},
	)
}

// ScanArg is an interface that wraps methods required to scan argument with
// scan functions.
type ScanArg interface {
	// Set sets status of scanning. It is set false before scanning and true
	// after scanning if argument is scanned.
	Set(bool)
	// Arg must return either a pointer to a basic Go type or implementations of
	// fmt.Scanner interface.
	Arg() interface{}
	// Value must return scanned, non-nil uGO Object.
	Value() ugo.Object
}

// argValue is an interface implemented by the basic scannable types and used by
// scanArg type.
type argValue interface {
	Arg() interface{}
	Value() ugo.Object
}

// scanArg implements ugo.Object and ScanArg interfaces to provide arguments to
// scan functions.
// "Value" selector in uGO scripts gives the scanned value if scan was successful.
type scanArg struct {
	ugo.ObjectImpl
	argValue
	ok bool
}

var _ ScanArg = (*scanArg)(nil)

func (*scanArg) TypeName() string { return "scanArg" }

func (o *scanArg) String() string { return "<scanArg>" }

func (o *scanArg) IsFalsy() bool { return !o.ok }

func (o *scanArg) IndexGet(index ugo.Object) (ugo.Object, error) {
	if o.ok && index.String() == "Value" {
		return o.Value(), nil
	}
	return ugo.Undefined, nil
}

func (o *scanArg) Set(scanned bool) { o.ok = scanned }

func newScanArgFunc(c ugo.Call) (ugo.Object, error) {
	typ := "string"
	if c.Len() > 0 {
		v := c.Get(0)
		if b, ok := v.(*ugo.BuiltinFunction); ok {
			typ = b.Name
		} else {
			typ = v.String()
		}
	}
	var scan scanArg
	switch typ {
	case "string":
		scan.argValue = &stringType{}
	case "int":
		scan.argValue = &intType{}
	case "uint":
		scan.argValue = &uintType{}
	case "float":
		scan.argValue = &floatType{}
	case "bool":
		scan.argValue = &boolType{}
	case "char":
		scan.argValue = &charType{}
	case "bytes":
		scan.argValue = &bytesType{}
	default:
		return nil, ugo.ErrType.NewError(strconv.Quote(typ), "not implemented")
	}
	return &scan, nil
}

type stringType struct {
	v string
}

func (st *stringType) Arg() interface{} {
	return &st.v
}

func (st *stringType) Value() ugo.Object {
	return ugo.String(st.v)
}

type bytesType struct {
	v []byte
}

func (bt *bytesType) Arg() interface{} {
	return &bt.v
}

func (bt *bytesType) Value() ugo.Object {
	return ugo.Bytes(bt.v)
}

type intType struct {
	v int64
}

func (it *intType) Arg() interface{} {
	return &it.v
}

func (it *intType) Value() ugo.Object {
	return ugo.Int(it.v)
}

type uintType struct {
	v uint64
}

func (ut *uintType) Arg() interface{} {
	return &ut.v
}

func (ut *uintType) Value() ugo.Object {
	return ugo.Uint(ut.v)
}

type floatType struct {
	v float64
}

func (ft *floatType) Arg() interface{} {
	return &ft.v
}

func (ft *floatType) Value() ugo.Object {
	return ugo.Float(ft.v)
}

type charType struct {
	v rune
}

func (ct *charType) Arg() interface{} {
	return &ct.v
}

func (ct *charType) Value() ugo.Object {
	return ugo.Char(ct.v)
}

type boolType struct {
	v bool
}

func (bt *boolType) Arg() interface{} {
	return &bt.v
}

func (bt *boolType) Value() ugo.Object {
	return ugo.Bool(bt.v)
}
