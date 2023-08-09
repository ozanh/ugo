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
	"bytes"
	"strings"
)

const (
	nullRep = "<null>"
)

// Node represents a node in the AST.
type Node interface {
	// Pos returns the position of first character belonging to the node.
	Pos() Pos
	// End returns the position of first character immediately after the node.
	End() Pos
	// String returns a string representation of the node.
	String() string
}

// IdentList represents a list of identifiers.
type IdentList struct {
	LParen  Pos
	VarArgs bool
	List    []*Ident
	RParen  Pos
}

// Pos returns the position of first character belonging to the node.
func (n *IdentList) Pos() Pos {
	if n.LParen.IsValid() {
		return n.LParen
	}
	if len(n.List) > 0 {
		return n.List[0].Pos()
	}
	return NoPos
}

// End returns the position of first character immediately after the node.
func (n *IdentList) End() Pos {
	if n.RParen.IsValid() {
		return n.RParen + 1
	}
	if l := len(n.List); l > 0 {
		return n.List[l-1].End()
	}
	return NoPos
}

// NumFields returns the number of fields.
func (n *IdentList) NumFields() int {
	if n == nil {
		return 0
	}
	return len(n.List)
}

func (n *IdentList) String() string {
	var list []string
	for i, e := range n.List {
		if n.VarArgs && i == len(n.List)-1 {
			list = append(list, "..."+e.String())
		} else {
			list = append(list, e.String())
		}
	}
	return "(" + strings.Join(list, ", ") + ")"
}

// ArgsList represents a list of identifiers.
type ArgsList struct {
	Var    *Ident
	Values []*Ident
}

// Pos returns the position of first character belonging to the node.
func (n *ArgsList) Pos() Pos {
	if len(n.Values) > 0 {
		return n.Values[0].Pos()
	} else if n.Var != nil {
		return n.Var.Pos()
	}
	return NoPos
}

// End returns the position of first character immediately after the node.
func (n *ArgsList) End() Pos {
	if n.Var != nil {
		return n.Var.End()
	} else if l := len(n.Values); l > 0 {
		return n.Values[l-1].End()
	}
	return NoPos
}

// NumFields returns the number of fields.
func (n *ArgsList) NumFields() int {
	if n == nil {
		return 0
	}
	return len(n.Values)
}

func (n *ArgsList) String() string {
	var list []string
	for _, e := range n.Values {
		list = append(list, e.String())
	}
	if n.Var != nil {
		list = append(list, "..."+n.Var.String())
	}
	return strings.Join(list, ", ")
}

// NamedArgsList represents a list of identifier with value pairs.
type NamedArgsList struct {
	Var    *Ident
	Names  []*Ident
	Values []Expr
}

func (n *NamedArgsList) Add(name *Ident, value Expr) *NamedArgsList {
	n.Names = append(n.Names, name)
	n.Values = append(n.Values, value)
	return n
}

// Pos returns the position of first character belonging to the node.
func (n *NamedArgsList) Pos() Pos {
	if len(n.Names) > 0 {
		return n.Names[0].Pos()
	} else if n.Var != nil {
		return n.Var.Pos()
	}
	return NoPos
}

// End returns the position of first character immediately after the node.
func (n *NamedArgsList) End() Pos {
	if n.Var != nil {
		return n.Var.End()
	}
	if l := len(n.Names); l > 0 {
		if n.Var != nil {
			return n.Var.End()
		}
		return n.Values[l-1].End()
	}
	return NoPos
}

// NumFields returns the number of fields.
func (n *NamedArgsList) NumFields() int {
	if n == nil {
		return 0
	}
	return len(n.Names)
}

func (n *NamedArgsList) String() string {
	var list []string
	for i, e := range n.Names {
		list = append(list, e.String()+"="+n.Values[i].String())
	}
	if n.Var != nil {
		list = append(list, "..."+n.Var.String())
	}
	return strings.Join(list, ", ")
}

// FuncParams represents a function paramsw.
type FuncParams struct {
	LParen    Pos
	Args      ArgsList
	NamedArgs NamedArgsList
	RParen    Pos
}

// Pos returns the position of first character belonging to the node.
func (n *FuncParams) Pos() (pos Pos) {
	if n.LParen.IsValid() {
		return n.LParen
	}
	if pos = n.Args.Pos(); pos != NoPos {
		return pos
	}
	if pos = n.NamedArgs.Pos(); pos != NoPos {
		return pos
	}
	return NoPos
}

// End returns the position of first character immediately after the node.
func (n *FuncParams) End() (pos Pos) {
	if n.RParen.IsValid() {
		return n.RParen + 1
	}
	if pos = n.NamedArgs.End(); pos != NoPos {
		return pos
	}
	if pos = n.Args.End(); pos != NoPos {
		return pos
	}
	return NoPos
}

func (n *FuncParams) String() string {
	buf := bytes.NewBufferString("(")
	buf.WriteString(n.Args.String())
	if buf.Len() > 1 && n.NamedArgs.Pos() != NoPos {
		buf.WriteString("; ")
	}
	buf.WriteString(n.NamedArgs.String())
	buf.WriteString(")")
	return buf.String()
}

// ----------------------------------------------------------------------------
// Comments

// A Comment node represents a single //-style or /*-style comment.
type Comment struct {
	Slash Pos    // position of "/" starting the comment
	Text  string // comment text (excluding '\n' for //-style comments)
}

// Pos returns the position of the comment's slash.
func (c *Comment) Pos() Pos { return c.Slash }

// End returns the position of first character immediately after the comment.
func (c *Comment) End() Pos {
	return Pos(int(c.Slash) + len(c.Text))
}

// A CommentGroup represents a sequence of comments
// with no other tokens and no empty lines between.
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

// Pos returns the position of the first comment.
func (g *CommentGroup) Pos() Pos {
	return g.List[0].Pos()
}

// End returns the position of last comment's end position.
func (g *CommentGroup) End() Pos {
	return g.List[len(g.List)-1].End()
}

// Text returns the text of the comment.
// Comment markers (//, /*, and */), the first space of a line comment, and
// leading and trailing empty lines are removed.
// Multiple empty lines are reduced to one, and trailing space on lines is trimmed.
// Unless the result is empty, it is newline-terminated.
func (g *CommentGroup) Text() string {
	if g == nil {
		return ""
	}
	comments := make([]string, len(g.List))
	for i, c := range g.List {
		comments[i] = c.Text
	}

	lines := make([]string, 0, 10) // most comments are less than 10 lines
	for _, c := range comments {
		// Remove comment markers.
		// The parser has given us exactly the comment text.
		switch c[1] {
		case '/':
			// -style comment (no newline at the end)
			c = c[2:]
			if len(c) == 0 {
				// empty line
				break
			}
			if c[0] == ' ' {
				// strip first space - required for Example tests
				c = c[1:]
			}
		case '*':
			/*-style comment */
			c = c[2 : len(c)-2]
		}

		// Split on newlines.
		cl := strings.Split(c, "\n")

		// Walk lines, stripping trailing white space and adding to list.
		for _, l := range cl {
			lines = append(lines, stripTrailingWhitespace(l))
		}
	}

	// Remove leading blank lines; convert runs of
	// interior blank lines to a single blank line.
	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	lines = lines[0:n]

	// Add final "" entry to get trailing newline from Join.
	if n > 0 && lines[n-1] != "" {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func stripTrailingWhitespace(s string) string {
	i := len(s)
	for i > 0 && isWhitespace(s[i-1]) {
		i--
	}
	return s[0:i]
}
