// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package time

import (
	"time"

	"github.com/ozanh/ugo"
)

// ugo:doc
// ## Types
// ### location
//
// Go Type
//
// ```go
// // Location represents location values and implements ugo.Object interface.
// type Location struct {
//    ugo.ObjectImpl
//    Value *time.Location
// }
// ```

// Location represents location values and implements ugo.Object interface.
type Location struct {
	ugo.ObjectImpl
	Value *time.Location
}

// TypeName implements ugo.Object interface.
func (*Location) TypeName() string {
	return "location"
}

// String implements ugo.Object interface.
func (o *Location) String() string {
	return o.Value.String()
}

// IsFalsy implements ugo.Object interface.
func (o *Location) IsFalsy() bool {
	return o.Value == nil
}

// Equal implements ugo.Object interface.
func (o *Location) Equal(right ugo.Object) bool {
	if v, ok := right.(*Location); ok {
		return v == o || v.String() == o.String()
	}
	if v, ok := right.(ugo.String); ok {
		return o.String() == v.String()
	}
	return false
}
