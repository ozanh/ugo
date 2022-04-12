// Copyright (c) 2022 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package json

import (
	"fmt"

	"github.com/ozanh/ugo"
)

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
	return "encoder-options"
}

// String implements ugo.Object interface.
func (eo *EncoderOptions) String() string {
	return fmt.Sprintf("encoder-options{Quote:%t EscapeHTML:%t Value:%s}",
		eo.Quote, eo.EscapeHTML, eo.Value)
}

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

// RawMessage represents raw encoded json message to directly use value of
// MarshalJSON without encoding.
type RawMessage struct {
	ugo.ObjectImpl
	Value []byte
}

var _ Marshaler = (*RawMessage)(nil)

// TypeName implements ugo.Object interface.
func (rm *RawMessage) TypeName() string {
	return "raw-message"
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
