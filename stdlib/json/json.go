// Copyright (c) 2022-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package json

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/registry"
)

func init() {
	registry.RegisterObjectConverter(reflect.TypeOf(json.RawMessage(nil)),
		func(in interface{}) (interface{}, bool) {
			rm := in.(json.RawMessage)
			if rm == nil {
				return &RawMessage{Value: ugo.Bytes{}}, true
			}
			return &RawMessage{Value: rm}, true
		},
	)

	registry.RegisterAnyConverter(reflect.TypeOf((*RawMessage)(nil)),
		func(in interface{}) (interface{}, bool) {
			rm := in.(*RawMessage)
			return json.RawMessage(rm.Value), true
		},
	)
}

// ugo:doc
// ## Types
// ### encoderOptions
//
// Go Type
//
// ```go
// // EncoderOptions represents the encoding options (quote, html escape) to
// // Marshal any Object.
// type EncoderOptions struct {
// 	ugo.ObjectImpl
// 	Value      ugo.Object
// 	Quote      bool
// 	EscapeHTML bool
// }
// ```

// EncoderOptions represents the encoding options (quote, html escape) to
// Marshal any Object.
type EncoderOptions struct {
	ugo.ObjectImpl
	Value      ugo.Object
	Quote      bool
	EscapeHTML bool
}

// TypeName implements ugo.Object interface.
func (eo *EncoderOptions) TypeName() string {
	return "encoderOptions"
}

// String implements ugo.Object interface.
func (eo *EncoderOptions) String() string {
	return fmt.Sprintf("encoderOptions{Quote:%t EscapeHTML:%t Value:%s}",
		eo.Quote, eo.EscapeHTML, eo.Value)
}

// ugo:doc
// #### encoderOptions Getters
//
//
// | Selector  | Return Type |
// |:----------|:------------|
// |.Value     | any         |
// |.Quote     | bool        |
// |.EscapeHTML| bool        |

// IndexGet implements ugo.Object interface.
func (eo *EncoderOptions) IndexGet(index ugo.Object) (ret ugo.Object, err error) {
	switch index.String() {
	case "Value":
		ret = eo.Value
	case "Quote":
		ret = ugo.Bool(eo.Quote)
	case "EscapeHTML":
		ret = ugo.Bool(eo.EscapeHTML)
	default:
		ret = ugo.Undefined
	}
	return
}

// ugo:doc
// #### encoderOptions Setters
//
//
// | Selector  | Value Type  |
// |:----------|:------------|
// |.Value     | any         |
// |.Quote     | bool        |
// |.EscapeHTML| bool        |

// IndexSet implements ugo.Object interface.
func (eo *EncoderOptions) IndexSet(index, value ugo.Object) error {
	switch index.String() {
	case "Value":
		eo.Value = value
	case "Quote":
		eo.Quote = !value.IsFalsy()
	case "EscapeHTML":
		eo.EscapeHTML = !value.IsFalsy()
	default:
		return ugo.ErrInvalidIndex
	}
	return nil
}

// ugo:doc
// ## Types
// ### rawMessage
//
// Go Type
//
// ```go
// // RawMessage represents raw encoded json message to directly use value of
// // MarshalJSON without encoding.
// type RawMessage struct {
// 	ugo.ObjectImpl
// 	Value []byte
// }
// ```

// RawMessage represents raw encoded json message to directly use value of
// MarshalJSON without encoding.
type RawMessage struct {
	ugo.ObjectImpl
	Value []byte
}

var _ Marshaler = (*RawMessage)(nil)

// TypeName implements ugo.Object interface.
func (rm *RawMessage) TypeName() string {
	return "rawMessage"
}

// String implements ugo.Object interface.
func (rm *RawMessage) String() string {
	return string(rm.Value)
}

// MarshalJSON implements Marshaler interface and returns rm as the JSON
// encoding of rm.Value.
func (rm *RawMessage) MarshalJSON() ([]byte, error) {
	if rm == nil || rm.Value == nil {
		return []byte("null"), nil
	}
	return rm.Value, nil
}

// ugo:doc
// #### rawMessage Getters
//
//
// | Selector  | Return Type |
// |:----------|:------------|
// |.Value     | bytes       |

// IndexGet implements ugo.Object interface.
func (rm *RawMessage) IndexGet(index ugo.Object) (ret ugo.Object, err error) {
	switch index.String() {
	case "Value":
		ret = ugo.Bytes(rm.Value)
	default:
		ret = ugo.Undefined
	}
	return
}

// ugo:doc
// #### rawMessage Setters
//
//
// | Selector  | Value Type  |
// |:----------|:------------|
// |.Value     | bytes       |

// IndexSet implements ugo.Object interface.
func (rm *RawMessage) IndexSet(index, value ugo.Object) error {
	switch index.String() {
	case "Value":
		if v, ok := ugo.ToBytes(value); ok {
			rm.Value = v
		} else {
			return ugo.ErrType
		}
	default:
		return ugo.ErrInvalidIndex
	}
	return nil
}
