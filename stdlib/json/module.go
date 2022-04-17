// Copyright (c) 2022 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package json

import (
	"bytes"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/stdlib"
)

// Module represents json module.
var Module = map[string]ugo.Object{
	// ugo:doc
	// # json Module
	//
	// ## Functions
	// Marshal(v any) -> bytes
	// Returns the JSON encoding v or error.
	"Marshal": &ugo.Function{
		Name: "Marshal",
		Value: stdlib.FuncPORO(
			func(o ugo.Object) ugo.Object {
				b, err := Marshal(o)
				if err != nil {
					return &ugo.Error{Message: err.Error(), Cause: err}
				}
				return ugo.Bytes(b)
			},
		),
	},
	// ugo:doc
	// MarshalIndent(v any, prefix string, indent string) -> bytes
	// MarshalIndent is like Marshal but applies Indent to format the output.
	"MarshalIndent": &ugo.Function{
		Name: "MarshalIndent",
		Value: stdlib.FuncPOssRO(
			func(o ugo.Object, prefix, indent string) ugo.Object {
				b, err := MarshalIndent(o, prefix, indent)
				if err != nil {
					return &ugo.Error{Message: err.Error(), Cause: err}
				}
				return ugo.Bytes(b)
			},
		),
	},
	// ugo:doc
	// Indent(src bytes, prefix string, indent string) -> bytes
	// Returns indented form of the JSON-encoded src or error.
	"Indent": &ugo.Function{
		Name: "Indent",
		Value: stdlib.FuncPb2ssRO(
			func(src []byte, prefix, indent string) ugo.Object {
				var buf bytes.Buffer
				err := indentBuffer(&buf, src, prefix, indent)
				if err != nil {
					return &ugo.Error{Message: err.Error(), Cause: err}
				}
				return ugo.Bytes(buf.Bytes())
			},
		),
	},
	// ugo:doc
	// RawMessage(v bytes) -> rawMessage
	// Returns a wrapped bytes to provide raw encoded JSON value to Marshal
	// functions.
	"RawMessage": &ugo.Function{
		Name: "RawMessage",
		Value: stdlib.FuncPb2RO(func(b []byte) ugo.Object {
			return &RawMessage{Value: b}
		}),
	},
	// ugo:doc
	// Compact(data bytes, escape bool) -> bytes
	// Returns elided insignificant space characters from data or error.
	"Compact": &ugo.Function{
		Name: "Compact",
		Value: stdlib.FuncPb2bRO(func(data []byte, escape bool) ugo.Object {
			var buf bytes.Buffer
			err := compact(&buf, data, escape)
			if err != nil {
				return &ugo.Error{Message: err.Error(), Cause: err}
			}
			return ugo.Bytes(buf.Bytes())
		}),
	},
	// ugo:doc
	// Quote(v any) -> encoderOptions
	// Returns a wrapped object to provide Marshal functions to quote v.
	"Quote": &ugo.Function{
		Name: "Quote",
		Value: stdlib.FuncPORO(func(o ugo.Object) ugo.Object {
			if v, ok := o.(*EncoderOptions); ok {
				v.Quote = true
				return v
			}
			return &EncoderOptions{Value: o, Quote: true, EscapeHTML: true}
		}),
	},
	// ugo:doc
	// NoQuote(v any) -> encoderOptions
	// Returns a wrapped object to provide Marshal functions not to quote while
	// encoding.
	// This can be used not to quote all array or map items.
	"NoQuote": &ugo.Function{
		Name: "NoQuote",
		Value: stdlib.FuncPORO(func(o ugo.Object) ugo.Object {
			if v, ok := o.(*EncoderOptions); ok {
				v.Quote = false
				return v
			}
			return &EncoderOptions{Value: o, Quote: false, EscapeHTML: true}
		}),
	},
	// ugo:doc
	// NoEscape(v any) -> encoderOptions
	// Returns a wrapped object to provide Marshal functions not to escape html
	// while encoding.
	"NoEscape": &ugo.Function{
		Name: "NoEscape",
		Value: stdlib.FuncPORO(func(o ugo.Object) ugo.Object {
			if v, ok := o.(*EncoderOptions); ok {
				v.EscapeHTML = false
				return v
			}
			return &EncoderOptions{Value: o}
		}),
	},
	// ugo:doc
	// Unmarshal(p bytes) -> any
	// Unmarshal parses the JSON-encoded p and returns the result or error.
	"Unmarshal": &ugo.Function{
		Name: "Unmarshal",
		Value: stdlib.FuncPb2RO(func(b []byte) ugo.Object {
			v, err := Unmarshal(b)
			if err != nil {
				return &ugo.Error{Message: err.Error(), Cause: err}
			}
			return v
		}),
	},
	// ugo:doc
	// Valid(p bytes) -> bool
	// Reports whether p is a valid JSON encoding.
	"Valid": &ugo.Function{
		Name: "Valid",
		Value: stdlib.FuncPb2RO(func(b []byte) ugo.Object {
			return ugo.Bool(valid(b))
		}),
	},
}
