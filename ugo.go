// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

//go:generate go run ./cmd/mkcallable -output zfuncs.go ugo.go

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/ozanh/ugo/registry"
)

const (
	// AttrModuleName is a special attribute injected into modules to identify
	// the modules by name.
	AttrModuleName = "__module_name__"
)

// CallableFunc is a function signature for a callable function.
type CallableFunc = func(args ...Object) (ret Object, err error)

// CallableExFunc is a function signature for a callable function that accepts
// a Call struct.
type CallableExFunc = func(Call) (ret Object, err error)

// ToObject will try to convert an interface{} v to an Object.
func ToObject(v interface{}) (ret Object, err error) {
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
	case map[string]Object:
		if v != nil {
			ret = Map(v)
		} else {
			ret = Map{}
		}
	case map[string]interface{}:
		m := make(Map, len(v))
		for vk, vv := range v {
			vo, err := ToObject(vv)
			if err != nil {
				return nil, err
			}
			m[vk] = vo
		}
		ret = m
	case []Object:
		if v != nil {
			ret = Array(v)
		} else {
			ret = Array{}
		}
	case []interface{}:
		arr := make(Array, len(v))
		for i, vv := range v {
			obj, err := ToObject(vv)
			if err != nil {
				return nil, err
			}
			arr[i] = obj
		}
		ret = arr
	case Object:
		ret = v
	case CallableFunc:
		if v != nil {
			ret = &Function{Value: v}
		} else {
			ret = Undefined
		}
	case error:
		ret = &Error{Message: v.Error(), Cause: v}
	default:
		if out, ok := registry.ToObject(v); ok {
			ret, ok = out.(Object)
			if ok {
				return
			}
		}
		err = fmt.Errorf("cannot convert to object: %T", v)
	}
	return
}

// ToObjectAlt is analogous to ToObject but it will always convert signed integers to
// Int and unsigned integers to Uint. It is an alternative to ToObject.
// Note that, this function is subject to change in the future.
func ToObjectAlt(v interface{}) (ret Object, err error) {
	switch v := v.(type) {
	case nil:
		ret = Undefined
	case string:
		ret = String(v)
	case bool:
		if v {
			ret = True
		} else {
			ret = False
		}
	case int:
		ret = Int(v)
	case int64:
		ret = Int(v)
	case uint64:
		ret = Uint(v)
	case float64:
		ret = Float(v)
	case float32:
		ret = Float(v)
	case int32:
		ret = Int(v)
	case int16:
		ret = Int(v)
	case int8:
		ret = Int(v)
	case uint:
		ret = Uint(v)
	case uint32:
		ret = Uint(v)
	case uint16:
		ret = Uint(v)
	case uint8:
		ret = Uint(v)
	case uintptr:
		ret = Uint(v)
	case []byte:
		if v != nil {
			ret = Bytes(v)
		} else {
			ret = Bytes{}
		}
	case map[string]interface{}:
		m := make(Map, len(v))
		for vk, vv := range v {
			vo, err := ToObjectAlt(vv)
			if err != nil {
				return nil, err
			}
			m[vk] = vo
		}
		ret = m
	case map[string]Object:
		if v != nil {
			ret = Map(v)
		} else {
			ret = Map{}
		}
	case []interface{}:
		arr := make(Array, len(v))
		for i, vv := range v {
			obj, err := ToObjectAlt(vv)
			if err != nil {
				return nil, err
			}
			arr[i] = obj
		}
		ret = arr
	case []Object:
		if v != nil {
			ret = Array(v)
		} else {
			ret = Array{}
		}
	case Object:
		ret = v
	case CallableFunc:
		if v != nil {
			ret = &Function{Value: v}
		} else {
			ret = Undefined
		}
	case error:
		ret = &Error{Message: v.Error(), Cause: v}
	default:
		if out, ok := registry.ToObject(v); ok {
			ret, ok = out.(Object)
			if ok {
				return
			}
		}
		err = fmt.Errorf("cannot convert to object: %T", v)
	}
	return
}

// ToInterface tries to convert an Object o to an interface{} value.
func ToInterface(o Object) (ret interface{}) {
	switch o := o.(type) {
	case Int:
		ret = int64(o)
	case String:
		ret = string(o)
	case Bytes:
		ret = []byte(o)
	case Array:
		arr := make([]interface{}, len(o))
		for i, val := range o {
			arr[i] = ToInterface(val)
		}
		ret = arr
	case Map:
		m := make(map[string]interface{}, len(o))
		for key, v := range o {
			m[key] = ToInterface(v)
		}
		ret = m
	case Uint:
		ret = uint64(o)
	case Char:
		ret = rune(o)
	case Float:
		ret = float64(o)
	case Bool:
		ret = bool(o)
	case *SyncMap:
		if o == nil {
			return map[string]interface{}{}
		}
		o.RLock()
		defer o.RUnlock()
		m := make(map[string]interface{}, len(o.Value))
		for key, v := range o.Value {
			m[key] = ToInterface(v)
		}
		ret = m
	case *UndefinedType:
		ret = nil
	default:
		if out, ok := registry.ToInterface(o); ok {
			ret = out
		} else {
			ret = o
		}
	}
	return
}

// ToString will try to convert an Object to uGO string value.
func ToString(o Object) (v String, ok bool) {
	if v, ok = o.(String); ok {
		return
	}
	vv, ok := ToGoString(o)
	if ok {
		v = String(vv)
	}
	return
}

// ToBytes will try to convert an Object to uGO bytes value.
func ToBytes(o Object) (v Bytes, ok bool) {
	if v, ok = o.(Bytes); ok {
		return
	}
	vv, ok := ToGoByteSlice(o)
	if ok {
		v = Bytes(vv)
	}
	return
}

// ToInt will try to convert an Object to uGO int value.
func ToInt(o Object) (v Int, ok bool) {
	if v, ok = o.(Int); ok {
		return
	}
	vv, ok := ToGoInt64(o)
	if ok {
		v = Int(vv)
	}
	return
}

// ToUint will try to convert an Object to uGO uint value.
func ToUint(o Object) (v Uint, ok bool) {
	if v, ok = o.(Uint); ok {
		return
	}
	vv, ok := ToGoUint64(o)
	if ok {
		v = Uint(vv)
	}
	return
}

// ToFloat will try to convert an Object to uGO float value.
func ToFloat(o Object) (v Float, ok bool) {
	if v, ok = o.(Float); ok {
		return
	}
	vv, ok := ToGoFloat64(o)
	if ok {
		v = Float(vv)
	}
	return
}

// ToChar will try to convert an Object to uGO char value.
func ToChar(o Object) (v Char, ok bool) {
	if v, ok = o.(Char); ok {
		return
	}
	vv, ok := ToGoRune(o)
	if ok {
		v = Char(vv)
	}
	return
}

// ToBool will try to convert an Object to uGO bool value.
func ToBool(o Object) (v Bool, ok bool) {
	if v, ok = o.(Bool); ok {
		return
	}
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

// ToSyncMap will try to convert an Object to uGO syncMap value.
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
	case Bytes:
		v, ok = o, true
	case String:
		v, ok = make([]byte, len(o)), true
		copy(v, o)
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
		ok = true
		v, _ = utf8.DecodeRuneInString(string(o))
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

// builtin int
//
//ugo:callable func(v int64) (ret Object)

// builtin uint
//
//ugo:callable func(v uint64) (ret Object)

// builtin float
//
//ugo:callable func(v float64) (ret Object)
