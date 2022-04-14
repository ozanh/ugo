// Copyright (c) 2020-2022 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

//go:generate go run ./cmd/mkcallable -output zfuncs.go ugo.go

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

// CallableFunc is a function signature for a callable function.
type CallableFunc = func(args ...Object) (ret Object, err error)

// FromInterface will try to convert an interface{} v to a uGO Object.
func FromInterface(v interface{}) (ret Object, err error) {
	switch v := v.(type) {
	case nil:
		ret = Undefined
	case string:
		ret = String(v)
	case int64:
		ret = Int(v)
	case int:
		ret = Int(v)
	case uint:
		ret = Uint(v)
	case uint64:
		ret = Uint(v)
	case uintptr:
		ret = Uint(v)
	case bool:
		if v {
			ret = True
		} else {
			ret = False
		}
	case rune:
		ret = Char(v)
	case byte:
		ret = Char(v)
	case float64:
		ret = Float(v)
	case float32:
		ret = Float(v)
	case []byte:
		if v != nil {
			ret = Bytes(v)
		} else {
			ret = Bytes{}
		}
	case error:
		if v != nil {
			ret = &Error{Message: v.Error()}
		} else {
			ret = &Error{Message: "<nil>"}
		}
	case map[string]Object:
		if v != nil {
			ret = Map(v)
		} else {
			ret = Map{}
		}
	case map[string]interface{}:
		m := make(Map, len(v))
		for vk, vv := range v {
			vo, err := FromInterface(vv)
			if err != nil {
				return nil, err
			}
			m[vk] = vo
		}
		ret = Map(m)
	case []Object:
		if v != nil {
			ret = Array(v)
		} else {
			ret = Array{}
		}
	case []interface{}:
		arr := make(Array, len(v))
		for i, e := range v {
			vo, err := FromInterface(e)
			if err != nil {
				return nil, err
			}
			arr[i] = vo
		}
		ret = Array(arr)
	case Object:
		if v != nil {
			ret = v
		} else {
			ret = Undefined
		}
	case CallableFunc:
		if v != nil {
			ret = &Function{Value: v}
		} else {
			ret = Undefined
		}
	default:
		err = fmt.Errorf("cannot convert to object: %T", v)
	}
	return
}

// ToInterface tries to convert an object o to an interface{} value.
func ToInterface(o Object) (ret interface{}) {
	switch o := o.(type) {
	case Int:
		ret = int64(o)
	case String:
		ret = string(o)
	case Bytes:
		ret = []byte(o)
	case Array:
		ret = make([]interface{}, len(o))
		for i, val := range o {
			ret.([]interface{})[i] = ToInterface(val)
		}
	case Map:
		ret = make(map[string]interface{})
		for key, v := range o {
			ret.(map[string]interface{})[key] = ToInterface(v)
		}
	case *UndefinedType:
		ret = nil
	case Uint:
		ret = uint64(o)
	case Char:
		ret = rune(o)
	case Float:
		ret = float64(o)
	case *SyncMap:
		o.RLock()
		defer o.RUnlock()
		ret = make(map[string]interface{})
		for key, v := range o.Map {
			ret.(map[string]interface{})[key] = ToInterface(v)
		}
	default:
		return o
	}
	return
}

// ToString will try to convert an Object to uGO string value.
func ToString(o Object) (v String, ok bool) {
	vv, ok := ToGoString(o)
	v = String(vv)
	return
}

// ToBytes will try to convert an Object to uGO bytes value.
func ToBytes(o Object) (v Bytes, ok bool) {
	vv, ok := ToGoByteSlice(o)
	v = Bytes(vv)
	return
}

// ToInt will try to convert an Object to uGO int value.
func ToInt(o Object) (v Int, ok bool) {
	vv, ok := ToGoInt64(o)
	v = Int(vv)
	return
}

// ToUint will try to convert an Object to uGO uint value.
func ToUint(o Object) (v Uint, ok bool) {
	vv, ok := ToGoUint64(o)
	v = Uint(vv)
	return
}

// ToFloat will try to convert an Object to uGO float value.
func ToFloat(o Object) (v Float, ok bool) {
	vv, ok := ToGoFloat64(o)
	v = Float(vv)
	return
}

// ToChar will try to convert an Object to uGO char value.
func ToChar(o Object) (v Char, ok bool) {
	vv, ok := ToGoRune(o)
	v = Char(vv)
	return
}

// ToBool will try to convert an Object to uGO bool value.
func ToBool(o Object) (v Bool, ok bool) {
	vv, ok := ToGoBool(o)
	v = Bool(vv)
	return
}

// ToArray will try to convert an Object to uGO array value.
func ToArray(o Object) (v Array, ok bool) {
	v, ok = o.(Array)
	return
}

// ToMap will try to convert an Object to uGO map value.
func ToMap(o Object) (v Map, ok bool) {
	v, ok = o.(Map)
	return
}

// ToSyncMap will try to convert an Object to uGO sync-map value.
func ToSyncMap(o Object) (v *SyncMap, ok bool) {
	v, ok = o.(*SyncMap)
	return
}

// ToGoString will try to convert an Object to Go string value.
func ToGoString(o Object) (v string, ok bool) {
	if o == Undefined {
		return
	}
	v, ok = o.String(), true
	return
}

// ToGoByteSlice will try to convert an Object to Go byte slice.
func ToGoByteSlice(o Object) (v []byte, ok bool) {
	switch o := o.(type) {
	case String:
		v, ok = []byte(o), true
	case Bytes:
		v, ok = o, true
	}
	return
}

// ToGoInt will try to convert a numeric, bool or string Object to Go int value.
func ToGoInt(o Object) (v int, ok bool) {
	switch o := o.(type) {
	case Int:
		v, ok = int(o), true
	case Uint:
		v, ok = int(o), true
	case Float:
		v, ok = int(o), true
	case Char:
		v, ok = int(o), true
	case Bool:
		ok = true
		if o {
			v = 1
		}
	case String:
		if vv, err := strconv.ParseInt(string(o), 0, 0); err == nil {
			v = int(vv)
			ok = true
		}
	}
	return
}

// ToGoInt64 will try to convert a numeric, bool or string Object to Go int64
// value.
func ToGoInt64(o Object) (v int64, ok bool) {
	switch o := o.(type) {
	case Int:
		v, ok = int64(o), true
	case Uint:
		v, ok = int64(o), true
	case Float:
		v, ok = int64(o), true
	case Char:
		v, ok = int64(o), true
	case Bool:
		ok = true
		if o {
			v = 1
		}
	case String:
		if vv, err := strconv.ParseInt(string(o), 0, 64); err == nil {
			v = vv
			ok = true
		}
	}
	return
}

// ToGoUint64 will try to convert a numeric, bool or string Object to Go uint64
// value.
func ToGoUint64(o Object) (v uint64, ok bool) {
	switch o := o.(type) {
	case Int:
		v, ok = uint64(o), true
	case Uint:
		v, ok = uint64(o), true
	case Float:
		v, ok = uint64(o), true
	case Char:
		v, ok = uint64(o), true
	case Bool:
		ok = true
		if o {
			v = 1
		}
	case String:
		if vv, err := strconv.ParseUint(string(o), 0, 64); err == nil {
			v = vv
			ok = true
		}
	}
	return
}

// ToGoFloat64 will try to convert a numeric, bool or string Object to Go
// float64 value.
func ToGoFloat64(o Object) (v float64, ok bool) {
	switch o := o.(type) {
	case Int:
		v, ok = float64(o), true
	case Uint:
		v, ok = float64(o), true
	case Float:
		v, ok = float64(o), true
	case Char:
		v, ok = float64(o), true
	case Bool:
		ok = true
		if o {
			v = 1
		}
	case String:
		if vv, err := strconv.ParseFloat(string(o), 64); err == nil {
			v = vv
			ok = true
		}
	}
	return
}

// ToGoRune will try to convert a int like Object to Go rune value.
func ToGoRune(o Object) (v rune, ok bool) {
	switch o := o.(type) {
	case Int:
		v, ok = rune(o), true
	case Uint:
		v, ok = rune(o), true
	case Char:
		v, ok = rune(o), true
	case Float:
		v, ok = rune(o), true
	case String:
		v, _ = utf8.DecodeRuneInString(string(o))
		if v != utf8.RuneError {
			ok = true
		}
	case Bool:
		ok = true
		if o {
			v = 1
		}
	}
	return
}

// ToGoBool will try to convert an Object to Go bool value.
func ToGoBool(o Object) (v bool, ok bool) {
	v, ok = !o.IsFalsy(), true
	return
}

// functions to generate with mkcallable

// builtin delete
//
//ugo:callable func(o Object, k string) (err error)

// builtin copy, len, error, typeName, bool, string, isInt, isUint
// isFloat, isChar, isBool, isString, isBytes, isMap, isSyncMap, isArray
// isUndefined, isFunction, isCallable, isIterable
//
//ugo:callable func(o Object) (ret Object)

// builtin repeat
//
//ugo:callable func(o Object, n int) (ret Object, err error)

// builtin :makeArray
//
//ugo:callable func(n int, o Object) (ret Object, err error)

// builtin contains
//
//ugo:callable func(o Object, v Object) (ret Object, err error)

// builtin sort, sortReverse, int, uint, float, char, chars
//
//ugo:callable func(o Object) (ret Object, err error)

// builtin error
//
//ugo:callable func(s string) (ret Object)

// builtin int
//
//ugo:callable func(v int64) (ret Object)

// builtin uint
//
//ugo:callable func(v uint64) (ret Object)

// builtin char
//
//ugo:callable func(v rune) (ret Object)

// builtin float
//
//ugo:callable func(v float64) (ret Object)
