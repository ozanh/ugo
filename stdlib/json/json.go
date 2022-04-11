// Copyright (c) 2022 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package json

import (
	"fmt"

	"github.com/ozanh/ugo"
)

type EncoderOptions struct {
	ugo.ObjectImpl
	Value      ugo.Object
	Quote      bool
	EscapeHTML bool
}

func (eo *EncoderOptions) TypeName() string {
	return "encoder-options"
}

func (eo *EncoderOptions) String() string {
	return fmt.Sprintf("encoder-options{Quote:%t EscapeHTML:%t Value:%s}",
		eo.Quote, eo.EscapeHTML, eo.Value)
}

func (eo *EncoderOptions) IndexGet(index ugo.Object) (value ugo.Object, err error) {
	switch index.String() {
	case "Value":
		value = eo.Value
	case "Quote":
		value = ugo.Bool(eo.Quote)
	case "EscapeHTML":
		value = ugo.Bool(eo.EscapeHTML)
	default:
		value = ugo.Undefined
	}
	return
}

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
