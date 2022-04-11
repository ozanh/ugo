// Copyright (c) 2022 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package json

import (
	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/futils"
)

// Module represents json module.
var Module = map[string]ugo.Object{
	"Marshal": &ugo.Function{
		Name: "Marshal",
		Value: futils.FuncPORO(
			func(o ugo.Object) ugo.Object {
				b, err := Marshal(o)
				if err != nil {
					return &ugo.Error{Message: err.Error(), Cause: err}
				}
				return ugo.Bytes(b)
			},
		),
	},
	"MarshalIndent": &ugo.Function{
		Name: "Marshal",
		Value: futils.FuncPOssRO(
			func(o ugo.Object, prefix, indent string) ugo.Object {
				b, err := MarshalIndent(o, prefix, indent)
				if err != nil {
					return &ugo.Error{Message: err.Error(), Cause: err}
				}
				return ugo.Bytes(b)
			},
		),
	},
	"Quote": &ugo.Function{
		Name: "Quote",
		Value: futils.FuncPORO(func(o ugo.Object) ugo.Object {
			if v, ok := o.(*EncoderOptions); ok {
				v.Quote = true
				return v
			}
			return &EncoderOptions{Value: o, Quote: true, EscapeHTML: true}
		}),
	},
	"Unquote": &ugo.Function{
		Name: "Unquote",
		Value: futils.FuncPORO(func(o ugo.Object) ugo.Object {
			if v, ok := o.(*EncoderOptions); ok {
				v.Quote = false
				return v
			}
			return &EncoderOptions{Value: o, Quote: false, EscapeHTML: true}
		}),
	},
	"NoEscape": &ugo.Function{
		Name: "NoEscape",
		Value: futils.FuncPORO(func(o ugo.Object) ugo.Object {
			if v, ok := o.(*EncoderOptions); ok {
				v.EscapeHTML = false
				return v
			}
			return &EncoderOptions{Value: o}
		}),
	},
	"Unmarshal": &ugo.Function{
		Name: "Unmarshal",
		Value: futils.FuncPb2RO(func(b []byte) ugo.Object {
			v, err := Unmarshal(b)
			if err != nil {
				return &ugo.Error{Message: err.Error(), Cause: err}
			}
			return v
		}),
	},
}
