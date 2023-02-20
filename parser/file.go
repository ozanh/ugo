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

import (
	"strings"
)

// File represents a file unit.
type File struct {
	InputFile *SourceFile
	Stmts     []Stmt
	Comments  []*CommentGroup
}

// Pos returns the position of first character belonging to the node.
func (n *File) Pos() Pos {
	return Pos(n.InputFile.Base)
}

// End returns the position of first character immediately after the node.
func (n *File) End() Pos {
	return Pos(n.InputFile.Base + n.InputFile.Size)
}

func (n *File) String() string {
	var stmts []string
	for _, e := range n.Stmts {
		stmts = append(stmts, e.String())
	}
	return strings.Join(stmts, "; ")
}
