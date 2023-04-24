// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Copyright (c) 2019 Daniel Kang.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE.tengo file.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.golang file.

package parser

// Pos represents a position in the file set.
type Pos int

// NoPos represents an invalid position.
const NoPos Pos = 0

// IsValid returns true if the position is valid.
func (p Pos) IsValid() bool {
	return p != NoPos
}
