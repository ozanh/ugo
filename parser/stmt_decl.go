// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.golang file.

// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package parser

import (
	"fmt"
	"strings"

	"github.com/ozanh/ugo/token"
)

// ----------------------------------------------------------------------------
// Declarations

type (
	// Spec node represents a single (non-parenthesized) variable declaration.
	// The Spec type stands for any of *ParamSpec or *ValueSpec.
	Spec interface {
		Node
		specNode()
	}

	// A ValueSpec node represents a variable declaration
	ValueSpec struct {
		Idents []*Ident    // TODO: slice is reserved for tuple assignment
		Values []Expr      // initial values; or nil
		Data   interface{} // iota
	}

	// A ParamSpec node represents a parameter declaration
	ParamSpec struct {
		Ident    *Ident
		Variadic bool
	}
)

// Pos returns the position of first character belonging to the spec.
func (s *ParamSpec) Pos() Pos { return s.Ident.Pos() }

// Pos returns the position of first character belonging to the spec.
func (s *ValueSpec) Pos() Pos { return s.Idents[0].Pos() }

// End returns the position of first character immediately after the spec.
func (s *ParamSpec) End() Pos {
	return s.Ident.End()
}

// End returns the position of first character immediately after the spec.
func (s *ValueSpec) End() Pos {
	if n := len(s.Values); n > 0 && s.Values[n-1] != nil {
		return s.Values[n-1].End()
	}
	return s.Idents[len(s.Idents)-1].End()
}

func (s *ParamSpec) String() string {
	str := s.Ident.String()
	if s.Variadic {
		str = token.Ellipsis.String() + str
	}
	return str
}
func (s *ValueSpec) String() string {
	vals := make([]string, 0, len(s.Idents))
	for i := range s.Idents {
		if s.Values[i] != nil {
			vals = append(vals, fmt.Sprintf("%s = %v", s.Idents[i], s.Values[i]))
		} else {
			vals = append(vals, s.Idents[i].String())
		}
	}
	return strings.Join(vals, ", ")
}

// specNode() ensures that only spec nodes can be assigned to a Spec.
func (*ParamSpec) specNode() {}

// specNode() ensures that only spec nodes can be assigned to a Spec.
func (*ValueSpec) specNode() {}

// Decl wraps methods for all declaration nodes.
type Decl interface {
	Node
	declNode()
}

// A DeclStmt node represents a declaration in a statement list.
type DeclStmt struct {
	Decl // *GenDecl with VAR token
}

func (*DeclStmt) stmtNode() {}

// A BadDecl node is a placeholder for declarations containing
// syntax errors for which no correct declaration nodes can be
// created.
type BadDecl struct {
	From, To Pos // position range of bad declaration
}

// A GenDecl node (generic declaration node) represents a variable declaration.
// A valid Lparen position (Lparen.Line > 0) indicates a parenthesized declaration.
//
// Relationship between Tok value and Specs element type:
//
//	token.Var     *ValueSpec
type GenDecl struct {
	TokPos Pos         // position of Tok
	Tok    token.Token // Var
	Lparen Pos         // position of '(', if any
	Specs  []Spec
	Rparen Pos // position of ')', if any
}

// Pos returns the position of first character belonging to the node.
func (d *BadDecl) Pos() Pos { return d.From }

// Pos returns the position of first character belonging to the node.
func (d *GenDecl) Pos() Pos { return d.TokPos }

// End returns the position of first character immediately after the node.
func (d *BadDecl) End() Pos { return d.To }

// End returns the position of first character immediately after the node.
func (d *GenDecl) End() Pos {
	if d.Rparen.IsValid() {
		return d.Rparen + 1
	}
	return d.Specs[0].End()
}

func (*BadDecl) declNode() {}
func (*GenDecl) declNode() {}

func (*BadDecl) String() string { return "<bad declaration>" }
func (d *GenDecl) String() string {
	var sb strings.Builder
	sb.WriteString(d.Tok.String())
	if d.Lparen > 0 {
		sb.WriteString(" (")
	} else {
		sb.WriteString(" ")
	}
	last := len(d.Specs) - 1
	for i := range d.Specs {
		sb.WriteString(d.Specs[i].String())
		if i != last {
			sb.WriteString(", ")
		}
	}
	if d.Rparen > 0 {
		sb.WriteString(")")
	}
	return sb.String()
}
