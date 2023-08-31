// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ozanh/ugo/token"
)

// Arg is a struct to destructure arguments from Call object.
type Arg struct {
	Value       Object
	AcceptTypes []string
}

// NamedArgVar is a struct to destructure named arguments from Call object.
type NamedArgVar struct {
	Name        string
	Value       Object
	ValueF      func() Object
	AcceptTypes []string
}

// NewNamedArgVar creates a new NamedArgVar struct with the given arguments.
func NewNamedArgVar(name string, value Object, types ...string) *NamedArgVar {
	return &NamedArgVar{Name: name, Value: value, AcceptTypes: types}
}

// NewNamedArgF creates a new NamedArgVar struct with the given arguments and value creator func.
func NewNamedArgVarF(name string, value func() Object, types ...string) *NamedArgVar {
	return &NamedArgVar{Name: name, ValueF: value, AcceptTypes: types}
}

type KeyValue [2]Object

var (
	_ Object       = KeyValue{}
	_ DeepCopier   = KeyValue{}
	_ Copier       = KeyValue{}
	_ LengthGetter = KeyValue{}
)

// TypeName implements Object interface.
func (KeyValue) TypeName() string {
	return "keyValue"
}

// String implements Object interface.
func (o KeyValue) String() string {
	var sb strings.Builder
	switch t := o[0].(type) {
	case String:
		if isLetterOrDigitRunes([]rune(t)) {
			sb.WriteString(string(t))
		} else {
			sb.WriteString(strconv.Quote(string(t)))
		}
	default:
		sb.WriteString(o[0].String())
	}
	if o[1] != True {
		sb.WriteString("=")
		switch t := o[1].(type) {
		case String:
			sb.WriteString(strconv.Quote(string(t)))
		default:
			sb.WriteString(t.String())
		}
	}
	return sb.String()
}

// DeepCopy implements DeepCopier interface.
func (o KeyValue) DeepCopy() Object {
	var cp KeyValue
	for i, v := range o[:] {
		if vv, ok := v.(DeepCopier); ok {
			cp[i] = vv.DeepCopy()
		} else {
			cp[i] = v
		}
	}
	return cp
}

// Copy implements Copier interface.
func (o KeyValue) Copy() Object {
	return KeyValue{o[0], o[1]}
}

// IndexSet implements Object interface.
func (o KeyValue) IndexSet(_, _ Object) error {
	return ErrNotIndexAssignable
}

// Equal implements Object interface.
func (o KeyValue) Equal(right Object) bool {
	v, ok := right.(KeyValue)
	if !ok {
		return false
	}

	return o[0].Equal(v[0]) && o[1].Equal(v[1])
}

// IsFalsy implements Object interface.
func (o KeyValue) IsFalsy() bool { return o[0] == Undefined && o[1] == Undefined }

// CanCall implements Object interface.
func (KeyValue) CanCall() bool { return false }

// Call implements Object interface.
func (KeyValue) Call(*NamedArgs, ...Object) (Object, error) {
	return nil, ErrNotCallable
}

// BinaryOp implements Object interface.
func (o KeyValue) BinaryOp(tok token.Token, right Object) (Object, error) {
	switch tok {
	case token.Less, token.LessEq:
		if right == Undefined {
			return False, nil
		}
		if kv, ok := right.(KeyValue); ok {
			if o.IsLess(kv) {
				return True, nil
			}
			if tok == token.LessEq {
				return Bool(o.Equal(kv)), nil
			}
			return False, nil
		}
	case token.Greater, token.GreaterEq:
		if right == Undefined {
			return True, nil
		}

		if tok == token.GreaterEq {
			if o.Equal(right) {
				return True, nil
			}
		}

		if kv, ok := right.(KeyValue); ok {
			return Bool(!o.IsLess(kv)), nil
		}
	}
	return nil, NewOperandTypeError(
		tok.String(),
		o.TypeName(),
		right.TypeName())
}

func (o KeyValue) IsLess(other KeyValue) bool {
	if o.Key().String() < other.Key().String() {
		return true
	}
	v, _ := o.Value().BinaryOp(token.Less, other.Value())
	return v == nil || !v.IsFalsy()
}

// CanIterate implements Object interface.
func (KeyValue) CanIterate() bool { return true }

// Iterate implements Iterable interface.
func (o KeyValue) Iterate() Iterator {
	return &ArrayIterator{V: o[:]}
}

// Len implements LengthGetter interface.
func (o KeyValue) Len() int {
	return len(o)
}

func (o KeyValue) Key() Object {
	return o[0]
}

func (o KeyValue) Value() Object {
	return o[1]
}

func (o KeyValue) IndexGet(index Object) (value Object, err error) {
	value = Undefined
	switch t := index.(type) {
	case String:
		switch t {
		case "k":
			return o[0], nil
		case "v":
			return o[1], nil
		case "array":
			return Array(o[:]), nil
		}
	case Int:
		switch t {
		case 0, 1:
			return o[t], nil
		}
	case Uint:
		switch t {
		case 0, 1:
			return o[t], nil
		}
	default:
		err = NewArgumentTypeError(
			"1st",
			"string|int|uint",
			index.TypeName(),
		)
		return
	}
	err = ErrInvalidIndex
	return
}

type KeyValueArray []KeyValue

var (
	_ Object       = KeyValueArray{}
	_ DeepCopier   = KeyValueArray{}
	_ Copier       = KeyValueArray{}
	_ LengthGetter = KeyValueArray{}
	_ Sorter       = KeyValueArray{}
	_ KeysGetter   = KeyValueArray{}
	_ ItemsGetter  = KeyValueArray{}
)

// TypeName implements Object interface.
func (KeyValueArray) TypeName() string {
	return "keyValueArray"
}

func (o KeyValueArray) Array() (ret Array) {
	ret = make(Array, len(o))
	for i, v := range o {
		ret[i] = v
	}
	return
}

func (o KeyValueArray) Map() (ret Map) {
	ret = make(Map, len(o))
	for _, v := range o {
		ret[v.Key().String()] = v.Value()
	}
	return
}

// String implements Object interface.
func (o KeyValueArray) String() string {
	var sb strings.Builder
	sb.WriteString("(;")
	last := len(o) - 1

	for i, v := range o {
		sb.WriteString(v.String())
		if i != last {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(")")
	return sb.String()
}

// DeepCopy implements DeepCopier interface.
func (o KeyValueArray) DeepCopy() Object {
	cp := make(KeyValueArray, len(o))
	for i, v := range o {
		cp[i] = v.DeepCopy().(KeyValue)
	}
	return cp
}

// Copy implements Copier interface.
func (o KeyValueArray) Copy() Object {
	cp := make(KeyValueArray, len(o))
	copy(cp, o)
	return cp
}

// IndexSet implements Object interface.
func (o KeyValueArray) IndexSet(_, _ Object) error {
	return ErrNotIndexAssignable
}

// IndexGet implements Object interface.
func (o KeyValueArray) IndexGet(index Object) (Object, error) {
	switch v := index.(type) {
	case Int:
		idx := int(v)
		if idx >= 0 && idx < len(o) {
			return o[v], nil
		}
		return nil, ErrIndexOutOfBounds
	case Uint:
		idx := int(v)
		if idx >= 0 && idx < len(o) {
			return o[v], nil
		}
		return nil, ErrIndexOutOfBounds
	case String:
		switch v {
		case "arrays":
			ret := make(Array, len(o))
			for i, v := range o {
				ret[i] = Array(v[:])
			}
			return ret, nil
		case "map":
			return o.Map(), nil
		default:
			return nil, ErrInvalidIndex.NewError(string(v))
		}
	}
	return nil, NewIndexTypeError("int|uint", index.TypeName())
}

// Equal implements Object interface.
func (o KeyValueArray) Equal(right Object) bool {
	v, ok := right.(KeyValueArray)
	if !ok {
		return false
	}

	if len(o) != len(v) {
		return false
	}

	for i := range o {
		if !o[i].Equal(v[i]) {
			return false
		}
	}
	return true
}

// IsFalsy implements Object interface.
func (o KeyValueArray) IsFalsy() bool { return len(o) == 0 }

// CanCall implements Object interface.
func (KeyValueArray) CanCall() bool { return false }

// Call implements Object interface.
func (KeyValueArray) Call(*NamedArgs, ...Object) (Object, error) {
	return nil, ErrNotCallable
}

func (o KeyValueArray) AppendArray(arr ...Array) (KeyValueArray, error) {
	var (
		i  = len(o)
		nl = i
		o2 KeyValueArray
	)

	for _, arr := range arr {
		nl += len(arr)
	}

	o2 = make(KeyValueArray, nl)
	copy(o2, o)

	for _, arr := range arr {
		for _, v := range arr {
			switch na := v.(type) {
			case KeyValue:
				o2[i] = na
				i++
			case Array:
				if len(na) == 2 {
					o2[i] = KeyValue{na[0], na[1]}
					i++
				} else {
					return nil, NewIndexValueTypeError("keyValue|[2]array",
						fmt.Sprintf("[%d]%s", len(na), v.TypeName()))
				}
			default:
				return nil, NewIndexTypeError("keyValue", v.TypeName())
			}
		}
	}
	return o2, nil
}

func (o KeyValueArray) AppendMap(m Map) KeyValueArray {
	var (
		i   = len(o)
		arr = make(KeyValueArray, i+len(m))
	)

	copy(arr, o)

	for k, v := range m {
		arr[i] = KeyValue{String(k), v}
		i++
	}

	return arr
}

func (o KeyValueArray) Append(arg ...KeyValue) KeyValueArray {
	if len(o) == 0 {
		return arg
	}
	var (
		i   = len(o)
		arr = make(KeyValueArray, i+len(arg))
	)

	copy(arr, o)
	copy(arr[i:], arg)
	return arr
}

func (o KeyValueArray) AppendObject(obj Object) (KeyValueArray, error) {
	switch v := obj.(type) {
	case KeyValue:
		return append(o, v), nil
	case Map:
		return o.AppendMap(v), nil
	case KeyValueArray:
		return o.Append(v...), nil
	case *NamedArgs:
		return o.Append(v.UnreadPairs()...), nil
	case Array:
		if o, err := o.AppendArray(v); err != nil {
			return nil, err
		} else {
			return o, nil
		}
	default:
		return nil, NewIndexTypeError("array|map|keyValue|keyValueArray", v.TypeName())
	}
}

// BinaryOp implements Object interface.
func (o KeyValueArray) BinaryOp(tok token.Token, right Object) (Object, error) {
	switch tok {
	case token.Add:
		return o.AppendObject(right)
	case token.Less, token.LessEq:
		if right == Undefined {
			return False, nil
		}
	case token.Greater, token.GreaterEq:
		if right == Undefined {
			return True, nil
		}
	}
	return nil, NewOperandTypeError(
		tok.String(),
		o.TypeName(),
		right.TypeName())
}

func (o KeyValueArray) Sort() (Object, error) {
	sort.Slice(o, func(i, j int) bool {
		return o[i].IsLess(o[j])
	})
	return o, nil
}

func (o KeyValueArray) SortReverse() (Object, error) {
	sort.Slice(o, func(i, j int) bool {
		return !o[i].IsLess(o[j])
	})
	return o, nil
}

func (o KeyValueArray) Get(keys ...Object) Object {
	if len(keys) == 0 {
		return Array{}
	}

	var e KeyValue
	if len(keys) > 1 {
		var arr Array
	keys:
		for _, key := range keys {
			for l := len(o); l > 0; l-- {
				e = o[l-1]
				if e[0].Equal(key) {
					arr = append(arr, e[1])
					continue keys
				}
			}
			arr = append(arr, Undefined)
		}
		return arr
	}
	for l := len(o); l > 0; l-- {
		e = o[l-1]
		if e[0].Equal(keys[0]) {
			return e[1]
		}
	}
	return Undefined
}

func (o KeyValueArray) Delete(keys ...Object) Object {
	if len(keys) == 0 {
		return o
	}

	var ret KeyValueArray
l:
	for _, kv := range o {
		for _, k := range keys {
			if kv[0].Equal(k) {
				continue l
			}
		}
		ret = append(ret, kv)
	}

	return ret
}

func (o KeyValueArray) CallName(name string, c Call) (_ Object, err error) {
	switch name {
	case "flag":
		if err = c.CheckLen(1); err != nil {
			return
		}
		keyArg := c.Get(0)
		var e KeyValue
		for l := len(o); l > 0; l-- {
			e = o[l-1]
			if e[0].Equal(keyArg) && !e.Value().IsFalsy() {
				return True, nil
			}
		}
		return False, nil
	case "get":
		return o.Get(c.Args()...), nil
	case "delete":
		return o.Delete(c.Args()...), nil
	case "values":
		if c.Len() == 0 {
			return o.Values(), nil
		}

		var (
			ret    Array
			keyArg Object
		)

		for _, keyArg = range c.args {
			for _, e := range o {
				if e[0].Equal(keyArg) {
					ret = append(ret, e.Value())
				}
			}
		}
		return ret, nil
	case "sort":
		switch len(c.args) {
		case 0:
		case 1:
			switch t := c.args[0].(type) {
			case Bool:
				if t {
					o2 := make(KeyValueArray, len(o))
					copy(o2, o)
					o = o2
				}
			default:
				return nil, NewArgumentTypeError(
					"1st",
					"bool",
					t.TypeName(),
				)
			}
		default:
			return nil, ErrWrongNumArguments.NewError("want<=1 got=" + strconv.Itoa(len(c.args)))
		}
		return o.Sort()
	case "sortReverse":
		switch len(c.args) {
		case 0:
		case 1:
			switch t := c.args[0].(type) {
			case Bool:
				if t {
					o2 := make(KeyValueArray, len(o))
					copy(o2, o)
					o = o2
				}
			default:
				return nil, NewArgumentTypeError(
					"1st",
					"bool",
					t.TypeName(),
				)
			}
		default:
			return nil, ErrWrongNumArguments.NewError("want<=1 got=" + strconv.Itoa(len(c.args)))
		}
		return o.SortReverse()
	default:
		return nil, ErrInvalidIndex.NewError(name)
	}
}

// CanIterate implements Object interface.
func (KeyValueArray) CanIterate() bool { return true }

// Iterate implements Iterable interface.
func (o KeyValueArray) Iterate() Iterator {
	return &KeyValueArrayIterator{V: o}
}

// Len implements LengthGetter interface.
func (o KeyValueArray) Len() int {
	return len(o)
}

func (o KeyValueArray) Items() KeyValueArray {
	return o
}

func (o KeyValueArray) Keys() (arr Array) {
	arr = make(Array, len(o))
	for i, v := range o {
		arr[i] = v[0]
	}
	return
}

func (o KeyValueArray) Values() (arr Array) {
	arr = make(Array, len(o))
	for i, v := range o {
		arr[i] = v[1]
	}
	return
}

// KeyValueArrayIterator represents an iterator for the array.
type KeyValueArrayIterator struct {
	V KeyValueArray
	i int
}

var _ Iterator = (*KeyValueArrayIterator)(nil)

// Next implements Iterator interface.
func (it *KeyValueArrayIterator) Next() bool {
	it.i++
	return it.i-1 < len(it.V)
}

// Key implements Iterator interface.
func (it *KeyValueArrayIterator) Key() Object {
	return Int(it.i - 1)
}

// Value implements Iterator interface.
func (it *KeyValueArrayIterator) Value() Object {
	i := it.i - 1
	if i > -1 && i < len(it.V) {
		return it.V[i]
	}
	return Undefined
}

type KeyValueArrays []KeyValueArray

// TypeName implements Object interface.
func (KeyValueArrays) TypeName() string {
	return "keyValueArrays"
}

func (o KeyValueArrays) Array() (ret Array) {
	ret = make(Array, len(o))
	for i, v := range o {
		ret[i] = v
	}
	return
}

// String implements Object interface.
func (o KeyValueArrays) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	last := len(o) - 1

	for i, v := range o {
		sb.WriteString(v.String())
		if i != last {
			sb.WriteString(", ")
		}
	}

	sb.WriteString("]")
	return sb.String()
}

// DeepCopy implements DeepCopier interface.
func (o KeyValueArrays) DeepCopy() Object {
	cp := make(KeyValueArrays, len(o))
	for i, v := range o {
		cp[i] = v.DeepCopy().(KeyValueArray)
	}
	return cp
}

// Copy implements Copier interface.
func (o KeyValueArrays) Copy() Object {
	cp := make(KeyValueArrays, len(o))
	copy(cp, o)
	return cp
}

// IndexSet implements Object interface.
func (o KeyValueArrays) IndexSet(_, _ Object) error {
	return ErrNotIndexAssignable
}

// IndexGet implements Object interface.
func (o KeyValueArrays) IndexGet(index Object) (Object, error) {
	switch v := index.(type) {
	case Int:
		idx := int(v)
		if idx >= 0 && idx < len(o) {
			return o[v], nil
		}
		return nil, ErrIndexOutOfBounds
	case Uint:
		idx := int(v)
		if idx >= 0 && idx < len(o) {
			return o[v], nil
		}
		return nil, ErrIndexOutOfBounds
	}
	return nil, NewIndexTypeError("int|uint", index.TypeName())
}

// Equal implements Object interface.
func (o KeyValueArrays) Equal(right Object) bool {
	v, ok := right.(KeyValueArrays)
	if !ok {
		return false
	}

	if len(o) != len(v) {
		return false
	}

	for i := range o {
		if !o[i].Equal(v[i]) {
			return false
		}
	}
	return true
}

// IsFalsy implements Object interface.
func (o KeyValueArrays) IsFalsy() bool { return len(o) == 0 }

// CanCall implements Object interface.
func (KeyValueArrays) CanCall() bool { return false }

// Call implements Object interface.
func (KeyValueArrays) Call(*NamedArgs, ...Object) (Object, error) {
	return nil, ErrNotCallable
}

// BinaryOp implements Object interface.
func (o KeyValueArrays) BinaryOp(tok token.Token, right Object) (Object, error) {
	switch tok {
	case token.Less, token.LessEq:
		if right == Undefined {
			return False, nil
		}
	case token.Greater, token.GreaterEq:
		if right == Undefined {
			return True, nil
		}
	}
	return nil, NewOperandTypeError(
		tok.String(),
		o.TypeName(),
		right.TypeName())
}

// CanIterate implements Object interface.
func (KeyValueArrays) CanIterate() bool { return true }

// Iterate implements Iterable interface.
func (o KeyValueArrays) Iterate() Iterator {
	return &NamedArgArraysIterator{V: o}
}

// Len implements LengthGetter interface.
func (o KeyValueArrays) Len() int {
	return len(o)
}
func (o KeyValueArrays) CallName(name string, c Call) (Object, error) {
	switch name {
	case "merge":
		l := len(o)
		switch l {
		case 0, 1:
			return o, nil
		default:
			var ret KeyValueArray
			for _, arr := range o {
				ret.Append(arr...)
			}
			return ret, nil
		}
	default:
		return nil, ErrInvalidIndex.NewError(name)
	}
}

// NamedArgArraysIterator represents an iterator for the array.
type NamedArgArraysIterator struct {
	V KeyValueArrays
	i int
}

var _ Iterator = (*NamedArgArraysIterator)(nil)

// Next implements Iterator interface.
func (it *NamedArgArraysIterator) Next() bool {
	it.i++
	return it.i-1 < len(it.V)
}

// Key implements Iterator interface.
func (it *NamedArgArraysIterator) Key() Object {
	return Int(it.i - 1)
}

// Value implements Iterator interface.
func (it *NamedArgArraysIterator) Value() Object {
	i := it.i - 1
	if i > -1 && i < len(it.V) {
		return it.V[i]
	}
	return Undefined
}

type NamedArgs struct {
	sources KeyValueArrays
	m       Map
	ready   Map
}

func NewNamedArgs(pairs ...KeyValueArray) *NamedArgs {
	return &NamedArgs{sources: pairs}
}

func (o *NamedArgs) Contains(key string) bool {
	if _, ok := o.ready[key]; ok {
		return false
	}
	o.check()
	_, ok := o.m[key]
	return ok
}

func (o *NamedArgs) Add(obj Object) error {
	arr, err := KeyValueArray{}.AppendObject(obj)
	if err != nil {
		return err
	}
	o.sources = append(o.sources, arr)
	return nil
}

func (o *NamedArgs) CallName(name string, c Call) (Object, error) {
	switch name {
	case "get":
		arg := &Arg{AcceptTypes: []string{"string"}}
		if err := c.DestructureArgs(arg); err != nil {
			return nil, err
		}
		return o.GetValue(string(arg.Value.(String))), nil
	default:
		return Undefined, ErrInvalidIndex.NewError(name)
	}
}

func (o *NamedArgs) TypeName() string {
	return "namedArgs"
}

func (o *NamedArgs) Join() KeyValueArray {
	switch len(o.sources) {
	case 0:
		return KeyValueArray{}
	case 1:
		return o.sources[0]
	default:
		ret := make(KeyValueArray, 0)
		for _, t := range o.sources {
			ret = append(ret, t...)
		}
		return ret
	}
}

func (o *NamedArgs) String() string {
	if len(o.ready) == 0 {
		return o.Join().String()
	}
	return o.UnreadPairs().String()
}

func (o *NamedArgs) BinaryOp(tok token.Token, right Object) (Object, error) {
	if right == Undefined {
		switch tok {
		case token.Add:
			if err := o.Add(right); err != nil {
				return nil, err
			}
			return o, nil
		case token.Less, token.LessEq:
			return False, nil
		case token.Greater, token.GreaterEq:
			return True, nil
		}
	}

	return nil, NewOperandTypeError(
		tok.String(),
		o.TypeName(),
		right.TypeName())
}

func (o *NamedArgs) IsFalsy() bool {
	return len(o.sources) == 0
}

func (o *NamedArgs) Equal(right Object) bool {
	v, ok := right.(*NamedArgs)
	if !ok {
		return false
	}
	if len(o.sources) != len(v.sources) {
		return false
	}
	for i, p := range o.sources {
		if !p.Equal(v.sources[i]) {
			return false
		}
	}
	return true
}

func (o *NamedArgs) Call(_ *NamedArgs, args ...Object) (Object, error) {
	return nil, ErrNotCallable
}

func (o *NamedArgs) CallEx(c Call) (Object, error) {
	arg := &Arg{AcceptTypes: []string{"string"}}
	if err := c.DestructureArgs(arg); err != nil {
		return nil, err
	}
	return o.GetValue(string(arg.Value.(String))), nil
}

func (o *NamedArgs) CanCall() bool {
	return true
}

func (o *NamedArgs) Iterate() Iterator {
	return o.Join().Iterate()
}

func (o *NamedArgs) CanIterate() bool {
	return true
}

func (o *NamedArgs) UnReady() *NamedArgs {
	return &NamedArgs{
		sources: KeyValueArrays{
			o.UnreadPairs(),
		},
	}
}

func (o *NamedArgs) Ready() (arr KeyValueArray) {
	if len(o.ready) == 0 {
		return
	}

	o.Walk(func(na KeyValue) error {
		if _, ok := o.ready[na.Key().String()]; ok {
			arr = append(arr, na)
		}
		return nil
	})
	return
}

func (o *NamedArgs) IndexGet(index Object) (value Object, err error) {
	switch t := index.(type) {
	case String:
		switch t {
		case "src":
			return o.sources, nil
		case "map":
			return o.Map(), nil
		case "unread":
			return o.UnReady(), nil
		case "ready":
			return o.Ready(), nil
		case "array":
			return o.Join(), nil
		case "readyNames":
			return o.ready.Keys(), nil
		default:
			return Undefined, ErrInvalidIndex.NewError(string(t))
		}
	default:
		err = NewArgumentTypeError(
			"1st",
			"string",
			index.TypeName(),
		)
		return
	}
}

func (o *NamedArgs) IndexSet(_, _ Object) error {
	return ErrNotIndexAssignable
}

func (o *NamedArgs) check() {
	if o.m == nil {
		o.m = Map{}
		o.ready = Map{}

		for i := len(o.sources) - 1; i >= 0; i-- {
			for _, v := range o.sources[i] {
				o.m[v.Key().String()] = v[1]
			}
		}
	}
}

// GetValue Must return value from key
func (o *NamedArgs) GetValue(key string) (val Object) {
	if val = o.GetValueOrNil(key); val == nil {
		val = Undefined
	}
	return
}

// GetPassedValue Get passed value
func (o *NamedArgs) GetPassedValue(key string) (val Object) {
	o.Walk(func(na KeyValue) error {
		if na.Key().String() == key {
			val = na[1]
			return io.EOF
		}
		return nil
	})
	return
}

// GetValue Must return value from key
func (o *NamedArgs) GetValueOrNil(key string) (val Object) {
	o.check()

	if val = o.m[key]; val != nil {
		delete(o.m, key)
		o.ready[key] = nil
		return
	}
	return nil
}

// Get destructure.
// Return errors:
// - ArgumentTypeError if type check of arg is fail.
// - UnexpectedNamedArg if have unexpected arg.
func (o *NamedArgs) Get(dst ...*NamedArgVar) (err error) {
	o.check()
	args := o.m.Copy().(Map)
	for k := range o.ready {
		delete(args, k)
	}

read:
	for i, d := range dst {
		if v, ok := args[d.Name]; ok && v != Undefined {
			if len(d.AcceptTypes) == 0 {
				d.Value = v
				delete(args, d.Name)
				continue
			}

			for _, t := range d.AcceptTypes {
				if v.TypeName() == t {
					d.Value = v
					delete(args, d.Name)
					continue read
				}
			}
			return NewArgumentTypeError(
				strconv.Itoa(i)+"st",
				strings.Join(d.AcceptTypes, "|"),
				v.TypeName(),
			)
		}

		if d.ValueF != nil {
			d.Value = d.ValueF()
		}
	}

	for key := range args {
		return ErrUnexpectedNamedArg.NewError(strconv.Quote(key))
	}
	return nil
}

// GetVar destructure and return others.
// Returns ArgumentTypeError if type check of arg is fail.
func (o *NamedArgs) GetVar(dst ...*NamedArgVar) (args Map, err error) {
	o.check()
	args = o.m
dst:
	for i, d := range dst {
		if v, ok := args[d.Name]; ok && v != Undefined {
			if len(d.AcceptTypes) == 0 {
				d.Value = v
				delete(args, d.Name)
				continue
			}

			for _, t := range d.AcceptTypes {
				if v.TypeName() == t {
					d.Value = v
					delete(args, d.Name)
					continue dst
				}
			}

			return nil, NewArgumentTypeError(
				strconv.Itoa(i)+"st",
				strings.Join(d.AcceptTypes, "|"),
				v.TypeName(),
			)
		}

		if d.ValueF != nil {
			d.Value = d.ValueF()
		}
	}

	return
}

// Empty return if is empty
func (o *NamedArgs) Empty() bool {
	return o.IsFalsy()
}

// Map return unread keys as Map
func (o *NamedArgs) Map() (ret Map) {
	o.check()
	return o.m.Copy().(Map)
}

func (o *NamedArgs) AllMap() (ret Map) {
	o.check()
	return o.m
}

func (o *NamedArgs) UnreadPairs() (ret KeyValueArray) {
	if len(o.ready) == 0 {
		o.Walk(func(na KeyValue) error {
			ret = append(ret, na)
			return nil
		})
		return
	}
	o.Walk(func(na KeyValue) error {
		if _, ok := o.ready[na.Key().String()]; !ok {
			ret = append(ret, na)
		}
		return nil
	})
	return
}

// Walk pass over all pairs and call `cb` function.
// if `cb` function returns any error, stop iterator and return then.
func (o *NamedArgs) Walk(cb func(na KeyValue) error) (err error) {
	o.check()
	for _, arr := range o.sources {
		for _, item := range arr {
			if err = cb(item); err != nil {
				return
			}
		}
	}
	return
}

func (o *NamedArgs) CheckNames(accept ...string) error {
	return o.Walk(func(na KeyValue) error {
		for _, name := range accept {
			if name == na.Key().String() {
				return nil
			}
		}
		return ErrUnexpectedNamedArg.NewError(strconv.Quote(na.Key().String()))
	})
}

func (o *NamedArgs) CheckNamesFromSet(set map[string]interface{}) error {
	if set == nil {
		return nil
	}
	return o.Walk(func(na KeyValue) error {
		if _, ok := set[na.Key().String()]; !ok {
			return ErrUnexpectedNamedArg.NewError(strconv.Quote(na.Key().String()))
		}
		return nil
	})
}

func isLetterOrDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' ||
		ch >= utf8.RuneSelf && (unicode.IsLetter(ch) || unicode.IsDigit(ch))
}

func isLetterOrDigitRunes(chs []rune) bool {
	for _, r := range chs {
		if !isLetterOrDigit(r) {
			return false
		}
	}
	return true
}
