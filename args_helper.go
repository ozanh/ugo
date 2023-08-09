// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"strconv"
	"strings"
)

// Arg is a struct to destructure arguments from Call object.
type Arg struct {
	Value Object
}

// NamedArg is a struct to destructure named arguments from Call object.
type NamedArg struct {
	Name        string
	Value       Object
	ValueF      func() Object
	AcceptTypes []string
}

// NewNamedArg creates a new NamedArg struct with the given arguments.
func NewNamedArg(name string, value Object, types ...string) *NamedArg {
	return &NamedArg{Name: name, Value: value, AcceptTypes: types}
}

// NewNamedArgF creates a new NamedArg struct with the given arguments and value creator func.
func NewNamedArgF(name string, value func() Object, types ...string) *NamedArg {
	return &NamedArg{Name: name, ValueF: value, AcceptTypes: types}
}

type NamedArgs struct {
	args  Map
	vargs Map
}

func NewNamedArgs(args Map, vargs ...Map) *NamedArgs {
	na := &NamedArgs{args: args}
	for _, na.vargs = range vargs {
	}
	return na
}

func (n *NamedArgs) Args() Map {
	return n.args
}

func (n *NamedArgs) Vargs() Map {
	return n.vargs
}

// GetValue Must return value from key
func (n *NamedArgs) GetValue(key string) (val Object) {
	if n.args != nil {
		if val = n.args[key]; val != nil {
			return
		}
	}
	if n.vargs != nil {
		if val = n.vargs[key]; val != nil {
			return
		}
	}
	return
}

// Get destructure.
// Return errors:
// - ArgumentTypeError if type check of arg is fail.
// - UnexpectedNamedArg if have unexpected arg.
func (n *NamedArgs) Get(dst ...*NamedArg) (err error) {
	vargs := Map{}

	if n.args != nil {
		for key, val := range n.args {
			vargs[key] = val
		}
	}
	if n.vargs != nil {
		for key, val := range n.vargs {
			vargs[key] = val
		}
	}

read:
	for i, d := range dst {
		if v, ok := vargs[d.Name]; ok && v != Undefined {
			if len(d.AcceptTypes) > 0 {
				for _, t := range d.AcceptTypes {
					if v.TypeName() == t {
						d.Value = v
						delete(vargs, d.Name)
						continue read
					}
				}
				return NewArgumentTypeError(
					strconv.Itoa(i)+"st",
					strings.Join(d.AcceptTypes, "|"),
					v.TypeName(),
				)
			} else {
				d.Value = v
				delete(vargs, d.Name)
				continue
			}
		}

		if d.ValueF != nil {
			d.Value = d.ValueF()
		}
	}

	for key := range vargs {
		return ErrUnexpectedNamedArg.NewError(strconv.Quote(key))
	}
	return nil
}

// GetVar destructure and return others.
// Returns ArgumentTypeError if type check of arg is fail.
func (n *NamedArgs) GetVar(dst ...*NamedArg) (vargs Map, err error) {
	vargs = Map{}

	if n.args != nil {
		for key, val := range n.args {
			vargs[key] = val
		}
	}
	if n.vargs != nil {
		for key, val := range n.vargs {
			vargs[key] = val
		}
	}

dst:
	for i, d := range dst {
		if v, ok := vargs[d.Name]; ok && v != Undefined {
			if len(d.AcceptTypes) > 0 {
				for _, t := range d.AcceptTypes {
					if v.TypeName() == t {
						d.Value = v
						delete(vargs, d.Name)
						continue dst
					}
				}
				return nil, NewArgumentTypeError(
					strconv.Itoa(i)+"st",
					strings.Join(d.AcceptTypes, "|"),
					v.TypeName(),
				)
			} else {
				d.Value = v
				delete(vargs, d.Name)
				continue
			}
		}

		if d.ValueF != nil {
			d.Value = d.ValueF()
		}
	}

	return
}

// Empty return if is empty
func (n *NamedArgs) Empty() bool {
	return (len(n.args) + len(n.vargs)) == 0
}

// All return all namedArgs
func (n *NamedArgs) All() (ret Map) {
	if n == nil {
		return
	}
	if n.vargs == nil {
		return n.args.Copy().(Map)
	}
	if n.args == nil {
		return n.vargs.Copy().(Map)
	}

	ret = make(Map, 0)

	for key, val := range n.vargs {
		ret[key] = val
	}

	for key, val := range n.args {
		ret[key] = val
	}

	return
}

// Walk pass over all pairs and call `cb` function.
// if `cb` function returns any error, stop iterator and return then.
func (n *NamedArgs) Walk(cb func(key string, val Object) error) (err error) {
	if n.vargs == nil {
		return nil
	}
	if n.args == nil {
		return nil
	}

	for key, val := range n.args {
		if err = cb(key, val); err != nil {
			return
		}
	}

	for key, val := range n.vargs {
		if err = cb(key, val); err != nil {
			return
		}
	}

	return
}

func (n *NamedArgs) CheckNames(accept ...string) error {
	return n.Walk(func(key string, val Object) error {
		for _, name := range accept {
			if name == key {
				return nil
			}
		}
		return ErrUnexpectedNamedArg.NewError(strconv.Quote(key))
	})
}

func (n *NamedArgs) CheckNamesFromSet(set map[string]interface{}) error {
	if set == nil {
		return nil
	}
	return n.Walk(func(key string, val Object) error {
		if _, ok := set[key]; !ok {
			return ErrUnexpectedNamedArg.NewError(strconv.Quote(key))
		}
		return nil
	})
}
