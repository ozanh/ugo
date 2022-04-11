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
		Value: futils.FuncPOROe(
			func(o ugo.Object) (ugo.Object, error) {
				b, err := Marshal(o)
				return ugo.Bytes(b), err
			},
		),
	},
	"MarshalIndent": &ugo.Function{
		Name: "Marshal",
		Value: futils.FuncPOssROe(
			func(o ugo.Object, prefix, indent string) (ugo.Object, error) {
				b, err := MarshalIndent(o, prefix, indent)
				return ugo.Bytes(b), err
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
		Name:  "Unmarshal",
		Value: futils.FuncPb2ROe(Unmarshal),
	},
}
