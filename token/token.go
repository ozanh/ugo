// Copyright (c) 2020-2025 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package token

import "strconv"

var keywords map[string]Token

// Token represents a token.
type Token int

// List of tokens
const (
	Illegal Token = iota
	EOF
	Comment
	_literalBeg
	Ident
	Int
	Uint
	Float
	Char
	String
	_literalEnd
	_operatorBeg
	Add          // +
	Sub          // -
	Mul          // *
	Quo          // /
	Rem          // %
	And          // &
	Or           // |
	Xor          // ^
	Shl          // <<
	Shr          // >>
	AndNot       // &^
	AddAssign    // +=
	SubAssign    // -=
	MulAssign    // *=
	QuoAssign    // /=
	RemAssign    // %=
	AndAssign    // &=
	OrAssign     // |=
	XorAssign    // ^=
	ShlAssign    // <<=
	ShrAssign    // >>=
	AndNotAssign // &^=
	LAnd         // &&
	LOr          // ||
	Inc          // ++
	Dec          // --
	Equal        // ==
	Less         // <
	Greater      // >
	Assign       // =
	Not          // !
	NotEqual     // !=
	LessEq       // <=
	GreaterEq    // >=
	Define       // :=
	Ellipsis     // ...
	LParen       // (
	LBrack       // [
	LBrace       // {
	Comma        // ,
	Period       // .
	RParen       // )
	RBrack       // ]
	RBrace       // }
	Semicolon    // ;
	Colon        // :
	Question     // ?
	_operatorEnd
	_keywordBeg
	Break
	Continue
	Else
	For
	Func
	If
	Return
	True
	False
	In
	Undefined
	Import
	Param
	Global
	Var
	Const
	Try
	Catch
	Finally
	Throw
	_keywordEnd
)

var tokens = [...]string{
	Illegal:      "ILLEGAL",
	EOF:          "EOF",
	Comment:      "COMMENT",
	Ident:        "IDENT",
	Int:          "INT",
	Uint:         "UINT",
	Float:        "FLOAT",
	Char:         "CHAR",
	String:       "STRING",
	Add:          "+",
	Sub:          "-",
	Mul:          "*",
	Quo:          "/",
	Rem:          "%",
	And:          "&",
	Or:           "|",
	Xor:          "^",
	Shl:          "<<",
	Shr:          ">>",
	AndNot:       "&^",
	AddAssign:    "+=",
	SubAssign:    "-=",
	MulAssign:    "*=",
	QuoAssign:    "/=",
	RemAssign:    "%=",
	AndAssign:    "&=",
	OrAssign:     "|=",
	XorAssign:    "^=",
	ShlAssign:    "<<=",
	ShrAssign:    ">>=",
	AndNotAssign: "&^=",
	LAnd:         "&&",
	LOr:          "||",
	Inc:          "++",
	Dec:          "--",
	Equal:        "==",
	Less:         "<",
	Greater:      ">",
	Assign:       "=",
	Not:          "!",
	NotEqual:     "!=",
	LessEq:       "<=",
	GreaterEq:    ">=",
	Define:       ":=",
	Ellipsis:     "...",
	LParen:       "(",
	LBrack:       "[",
	LBrace:       "{",
	Comma:        ",",
	Period:       ".",
	RParen:       ")",
	RBrack:       "]",
	RBrace:       "}",
	Semicolon:    ";",
	Colon:        ":",
	Question:     "?",
	Break:        "break",
	Continue:     "continue",
	Else:         "else",
	For:          "for",
	Func:         "func",
	If:           "if",
	Return:       "return",
	True:         "true",
	False:        "false",
	In:           "in",
	Undefined:    "undefined",
	Import:       "import",
	Param:        "param",
	Global:       "global",
	Var:          "var",
	Const:        "const",
	Try:          "try",
	Catch:        "catch",
	Finally:      "finally",
	Throw:        "throw",
}

func (tok Token) String() string {
	s := ""

	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}

	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}

	return s
}

// LowestPrec represents lowest operator precedence.
const LowestPrec = 0

// Precedence returns the precedence for the operator token.
func (tok Token) Precedence() int {
	switch tok {
	case LOr:
		return 1
	case LAnd:
		return 2
	case Equal, NotEqual, Less, LessEq, Greater, GreaterEq:
		return 3
	case Add, Sub, Or, Xor:
		return 4
	case Mul, Quo, Rem, Shl, Shr, And, AndNot:
		return 5
	}
	return LowestPrec
}

// IsLiteral returns true if the token is a literal.
func (tok Token) IsLiteral() bool {
	return _literalBeg < tok && tok < _literalEnd
}

// IsOperator returns true if the token is an operator.
func (tok Token) IsOperator() bool {
	return _operatorBeg < tok && tok < _operatorEnd
}

// IsBinaryOperator reports whether token is a binary operator.
func (tok Token) IsBinaryOperator() bool {
	switch tok {
	case Add,
		Sub,
		Mul,
		Quo,
		Rem,
		Less,
		LessEq,
		Greater,
		GreaterEq,
		And,
		Or,
		Xor,
		AndNot,
		Shl,
		Shr,
		Equal,
		NotEqual:
		return true
	}
	return false
}

// IsKeyword returns true if the token is a keyword.
func (tok Token) IsKeyword() bool {
	return _keywordBeg < tok && tok < _keywordEnd
}

// Lookup returns corresponding keyword if ident is a keyword.
func Lookup(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return Ident
}

func init() {
	keywords = make(map[string]Token)
	for i := _keywordBeg + 1; i < _keywordEnd; i++ {
		keywords[tokens[i]] = i
	}
}
