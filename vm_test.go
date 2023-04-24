package ugo_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ozanh/ugo/tests"

	. "github.com/ozanh/ugo"
)

func TestVMArray(t *testing.T) {
	expectRun(t, `return [1, 2 * 2, 3 + 3]`, nil, Array{Int(1), Int(4), Int(6)})

	// array copy-by-reference
	expectRun(t, `a1 := [1, 2, 3]; a2 := a1; a1[0] = 5; return a2`,
		nil, Array{Int(5), Int(2), Int(3)})
	expectRun(t, `var out; func () { a1 := [1, 2, 3]; a2 := a1; a1[0] = 5; out = a2 }(); return out`,
		nil, Array{Int(5), Int(2), Int(3)})

	// array index set
	expectErrIs(t, `a1 := [1, 2, 3]; a1[3] = 5`, nil, ErrIndexOutOfBounds)

	// index operator
	arr := Array{Int(1), Int(2), Int(3), Int(4), Int(5), Int(6)}
	arrStr := `[1, 2, 3, 4, 5, 6]`
	arrLen := 6
	for idx := 0; idx < arrLen; idx++ {
		expectRun(t, fmt.Sprintf("return %s[%d]", arrStr, idx),
			nil, arr[idx])
		expectRun(t, fmt.Sprintf("return %s[0 + %d]", arrStr, idx),
			nil, arr[idx])
		expectRun(t, fmt.Sprintf("return %s[1 + %d - 1]", arrStr, idx),
			nil, arr[idx])
		expectRun(t, fmt.Sprintf("idx := %d; return %s[idx]", idx, arrStr),
			nil, arr[idx])
	}
	expectErrIs(t, fmt.Sprintf("%s[%d]", arrStr, -1), nil, ErrIndexOutOfBounds)
	expectErrIs(t, fmt.Sprintf("%s[%d]", arrStr, arrLen), nil, ErrIndexOutOfBounds)

	// slice operator
	for low := 0; low < arrLen; low++ {
		expectRun(t, fmt.Sprintf("return %s[%d:%d]", arrStr, low, low),
			nil, Array{})
		for high := low; high <= arrLen; high++ {
			expectRun(t, fmt.Sprintf("return %s[%d:%d]", arrStr, low, high),
				nil, arr[low:high])
			expectRun(t, fmt.Sprintf("return %s[0 + %d : 0 + %d]",
				arrStr, low, high), nil, arr[low:high])
			expectRun(t, fmt.Sprintf("return %s[1 + %d - 1 : 1 + %d - 1]",
				arrStr, low, high), nil, arr[low:high])
			expectRun(t, fmt.Sprintf("return %s[:%d]", arrStr, high),
				nil, arr[:high])
			expectRun(t, fmt.Sprintf("return %s[%d:]", arrStr, low),
				nil, arr[low:])
		}
	}

	expectRun(t, fmt.Sprintf("return %s[:]", arrStr), nil, arr)
	expectRun(t, fmt.Sprintf("return %s[%d:%d]", arrStr, 2, 2), nil, Array{})
	expectErrIs(t, fmt.Sprintf("return %s[%d:\"\"]", arrStr, -1), nil, ErrType)
	expectErrIs(t, fmt.Sprintf("return %s[%d:]", arrStr, -1), nil, ErrIndexOutOfBounds)
	expectErrIs(t, fmt.Sprintf("return %s[:%d]", arrStr, arrLen+1), nil, ErrIndexOutOfBounds)
	expectErrIs(t, fmt.Sprintf("%s[%d:%d]", arrStr, 2, 1), nil, ErrInvalidIndex)
	expectErrIs(t, fmt.Sprintf("%s[%d:]", arrStr, arrLen+1), nil, ErrInvalidIndex)
	expectErrIs(t, fmt.Sprintf("%s[:%d]", arrStr, -1), nil, ErrInvalidIndex)
	expectErrIs(t, "return 1[0:]", nil, ErrType)
	expectErrIs(t, "return 1[0]", nil, ErrNotIndexable)
}

func TestVMDecl(t *testing.T) {
	expectRun(t, `param a; return a`, nil, Undefined)
	expectRun(t, `param (a); return a`, nil, Undefined)
	expectRun(t, `param ...a; return a`, nil, Array{})
	expectRun(t, `param (a, ...b); return b`, nil, Array{})
	expectRun(t, `param (a, b); return [a, b]`,
		nil, Array{Undefined, Undefined})
	expectRun(t, `param a; return a`,
		newOpts().Args(Int(1)), Int(1))
	expectRun(t, `param (a, b); return a + b`,
		newOpts().Args(Int(1), Int(2)), Int(3))
	expectRun(t, `param (a, ...b); return b`,
		newOpts().Args(Int(1)), Array{})
	expectRun(t, `param (a, ...b); return b+a`,
		newOpts().Args(Int(1), Int(2)), Array{Int(2), Int(1)})
	expectRun(t, `param ...a; return a`,
		newOpts().Args(Int(1), Int(2)), Array{Int(1), Int(2)})
	expectErrHas(t, `func(){ param x; }`, newOpts().CompilerError(),
		`Compile Error: param not allowed in this scope`)

	expectRun(t, `global a; return a`, nil, Undefined)
	expectRun(t, `global (a); return a`, nil, Undefined)
	expectRun(t, `global (a, b); return [a, b]`,
		nil, Array{Undefined, Undefined})
	expectRun(t, `global a; return a`,
		newOpts().Globals(Map{"a": String("ok")}), String("ok"))
	expectRun(t, `global (a, b); return a+b`,
		newOpts().Globals(Map{"a": Int(1), "b": Int(2)}), Int(3))
	expectErrHas(t, `func() { global a }`, newOpts().CompilerError(),
		`Compile Error: global not allowed in this scope`)

	expectRun(t, `var a; return a`, nil, Undefined)
	expectRun(t, `var (a); return a`, nil, Undefined)
	expectRun(t, `var (a = 1); return a`, nil, Int(1))
	expectRun(t, `var (a, b = 1); return a`, nil, Undefined)
	expectRun(t, `var (a, b = 1); return b`, nil, Int(1))
	expectRun(t, `var (a,
		b = 1); return a`, nil, Undefined)
	expectRun(t, `var (a,
		b = 1); return b`, nil, Int(1))
	expectRun(t, `var (a = 1, b = "x"); return b`, nil, String("x"))
	expectRun(t, `var (a = 1, b = "x"); return a`, nil, Int(1))
	expectRun(t, `var (a = 1, b); return a`, nil, Int(1))
	expectRun(t, `var (a = 1, b); return b`, nil, Undefined)
	expectRun(t, `var b = 1; return b`, nil, Int(1))
	expectRun(t, `var (a, b, c); return [a, b, c]`,
		nil, Array{Undefined, Undefined, Undefined})
	expectRun(t, `return func(a) { var (b = 2,c); return [a, b, c] }(1)`,
		nil, Array{Int(1), Int(2), Undefined})

	expectErrHas(t, `param x; global x`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `param x; var x`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `var x; param x`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `var x; global x`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `a := 1; if a { param x }`, newOpts().CompilerError(),
		`Compile Error: param not allowed in this scope`)
	expectErrHas(t, `a := 1; if a { global x }`, newOpts().CompilerError(),
		`Compile Error: global not allowed in this scope`)
	expectErrHas(t, `func() { param x }`, newOpts().CompilerError(),
		`Compile Error: param not allowed in this scope`)
	expectErrHas(t, `func() { global x }`, newOpts().CompilerError(),
		`Compile Error: global not allowed in this scope`)

	expectRun(t, `param x; return func(x) { return x }(1)`, nil, Int(1))
	expectRun(t, `
	param x
	return func(x) { 
		for i := 0; i < 1; i++ {
			return x
		}
	}(1)`, nil, Int(1))
	expectRun(t, `
	param x
	func() {
		if x || !x {
			x = 2
		}
	}()
	return x`, newOpts().Args(Int(0)), Int(2))
	expectRun(t, `
	param x
	func() {
		if x || !x {
			func() {
				x = 2
			}()
		}
	}()
	return x`, newOpts().Args(Int(0)), Int(2))
	expectRun(t, `
	param x
	return func(x) { 
		for i := 0; i < 1; i++ {
			return x
		}
	}(1)`, nil, Int(1))
	expectRun(t, `
	global x
	func() {
		if x || !x {
			x = 2
		}
	}()
	return x`, nil, Int(2))
	expectRun(t, `
	global x
	func() {
		if x || !x {
			func() {
				x = 2
			}()
		}
	}()
	return x`, nil, Int(2))
}

func TestVMAssignment(t *testing.T) {
	expectErrHas(t, `a.b := 1`, newOpts().CompilerError(),
		`Compile Error: operator ':=' not allowed with selector`)

	expectRun(t, `a := 1; a = 2; return a`, nil, Int(2))
	expectRun(t, `a := 1; a = a + 4; return a`, nil, Int(5))
	expectRun(t, `a := 1; f1 := func() { a = 2; return a }; return f1()`,
		nil, Int(2))
	expectRun(t, `a := 1; f1 := func() { a := 3; a = 2; return a }; return f1()`,
		nil, Int(2))

	expectRun(t, `a := 1; return a`, nil, Int(1))
	expectRun(t, `a := 1; func() { a = 2 }(); return a`, nil, Int(2))
	expectRun(t, `a := 1; func() { a := 2 }(); return a`, nil, Int(1)) // "a := 2" shadows variable 'a' in upper scope
	expectRun(t, `a := 1; return func() { b := 2; return b }()`, nil, Int(2))
	expectRun(t, `
	return func() { 
		a := 2
		func() {
			a = 3 // a is free (non-local) variable
		}()
		return a
	}()
	`, nil, Int(3))

	expectRun(t, `
	var out
	func() {
		a := 5
		out = func() {  	
			a := 4						
			return a
		}()
	}()
	return out`, nil, Int(4))

	expectErrHas(t, `a := 1; a := 2`, newOpts().CompilerError(),
		`Compile Error: "a" redeclared in this block`) // redeclared in the same scope
	expectErrHas(t, `func() { a := 1; a := 2 }()`, newOpts().CompilerError(),
		`Compile Error: "a" redeclared in this block`) // redeclared in the same scope

	expectRun(t, `a := 1; a += 2; return a`, nil, Int(3))
	expectRun(t, `a := 1; a += 4 - 2; return a`, nil, Int(3))
	expectRun(t, `a := 3; a -= 1; return a`, nil, Int(2))
	expectRun(t, `a := 3; a -= 5 - 4; return a`, nil, Int(2))
	expectRun(t, `a := 2; a *= 4; return a`, nil, Int(8))
	expectRun(t, `a := 2; a *= 1 + 3; return a`, nil, Int(8))
	expectRun(t, `a := 10; a /= 2; return a`, nil, Int(5))
	expectRun(t, `a := 10; a /= 5 - 3; return a`, nil, Int(5))

	// compound assignment operator does not define new variable
	expectErrHas(t, `a += 4`, newOpts().CompilerError(), `Compile Error: unresolved reference "a"`)
	expectErrHas(t, `a -= 4`, newOpts().CompilerError(), `Compile Error: unresolved reference "a"`)
	expectErrHas(t, `a *= 4`, newOpts().CompilerError(), `Compile Error: unresolved reference "a"`)
	expectErrHas(t, `a /= 4`, newOpts().CompilerError(), `Compile Error: unresolved reference "a"`)

	expectRun(t, `
	f1 := func() {
		f2 := func() {
			a := 1
			a += 2
			return a
		};
		return f2();
	};
	return f1();`, nil, Int(3))
	expectRun(t, `f1 := func() { f2 := func() { a := 1; a += 4 - 2; return a }; return f2(); }; return f1()`,
		nil, Int(3))
	expectRun(t, `f1 := func() { f2 := func() { a := 3; a -= 1; return a }; return f2(); }; return f1()`,
		nil, Int(2))
	expectRun(t, `f1 := func() { f2 := func() { a := 3; a -= 5 - 4; return a }; return f2(); }; return f1()`,
		nil, Int(2))
	expectRun(t, `f1 := func() { f2 := func() { a := 2; a *= 4; return a }; return f2(); }; return f1()`,
		nil, Int(8))
	expectRun(t, `f1 := func() { f2 := func() { a := 2; a *= 1 + 3; return a }; return f2(); }; return f1()`,
		nil, Int(8))
	expectRun(t, `f1 := func() { f2 := func() { a := 10; a /= 2; return a }; return f2(); }; return f1()`,
		nil, Int(5))
	expectRun(t, `f1 := func() { f2 := func() { a := 10; a /= 5 - 3; return a }; return f2(); }; return f1()`,
		nil, Int(5))
	expectRun(t, `a := 1; f1 := func() { f2 := func() { a += 2; return a }; return f2(); }; return f1()`,
		nil, Int(3))
	expectRun(t, `
	f1 := func(a) {
		return func(b) {
			c := a
			c += b * 2
			return c
		}
	}
	return f1(3)(4)
	`, nil, Int(11))

	expectRun(t, `
	return func() {
		a := 1
		func() {
			a = 2
			func() {
				a = 3
				func() {
					a := 4 // declared new
				}()
			}()
		}()
		return a
	}()
	`, nil, Int(3))

	// write on free variables
	expectRun(t, `
	f1 := func() {
		a := 5
		return func() {
			a += 3
			return a
		}()
	}
	return f1()
	`, nil, Int(8))

	expectRun(t, `
	return func() {
		f1 := func() {
			a := 5
			add1 := func() { a += 1 }
			add2 := func() { a += 2 }
			a += 3
			return func() { a += 4; add1(); add2(); a += 5; return a }
		}
		return f1()
	}()()
	`, nil, Int(20))

	expectRun(t, `
	it := func(seq, fn) {
		fn(seq[0])
		fn(seq[1])
		fn(seq[2])
	}

	foo := func(a) {
		b := 0
		it([1, 2, 3], func(x) {
			b = x + a
		})
		return b
	}
	return foo(2)
	`, nil, Int(5))

	expectRun(t, `
	it := func(seq, fn) {
		fn(seq[0])
		fn(seq[1])
		fn(seq[2])
	}

	foo := func(a) {
		b := 0
		it([1, 2, 3], func(x) {
			b += x + a
		})
		return b
	}
	return foo(2)
	`, nil, Int(12))

	expectRun(t, `
	return func() {
		a := 1
		func() {
			a = 2
		}()
		return a
	}()
	`, nil, Int(2))

	expectRun(t, `
	f := func() {
		a := 1
		return {
			b: func() { a += 3 },
			c: func() { a += 2 },
			d: func() { return a },
		}
	}
	m := f()
	m.b()
	m.c()
	return m.d()
	`, nil, Int(6))

	expectRun(t, `
	each := func(s, x) { for i:=0; i<len(s); i++ { x(s[i]) } }

	return func() {
		a := 100
		each([1, 2, 3], func(x) {
			a += x
		})
		a += 10
		return func(b) {
			return a + b
		}
	}()(20)
	`, nil, Int(136))

	// assigning different type value
	expectRun(t, `a := 1; a = "foo"; return a`, nil, String("foo"))
	expectRun(t, `return func() { a := 1; a = "foo"; return a }()`, nil, String("foo"))
	expectRun(t, `
	return func() {
		a := 5
		return func() {
			a = "foo"
			return a
		}()
	}()`, nil, String("foo")) // free

	// variables declared in if/for blocks
	expectRun(t, `for a:=0; a<5; a++ {}; a := "foo"; return a`, nil, String("foo"))
	expectRun(t, `var out; func() { for a:=0; a<5; a++ {}; a := "foo"; out = a }(); return out`,
		nil, String("foo"))
	expectRun(t, `a:=0; if a:=1; a>0 { return a }; return 0`, nil, Int(1))
	expectRun(t, `a:=1; if a:=0; a>0 { return a }; return a`, nil, Int(1))

	// selectors
	expectRun(t, `a:=[1,2,3]; a[1] = 5; return a[1]`, nil, Int(5))
	expectRun(t, `a:=[1,2,3]; a[1] += 5; return a[1]`, nil, Int(7))
	expectRun(t, `a:={b:1,c:2}; a.b = 5; return a.b`, nil, Int(5))
	expectRun(t, `a:={b:1,c:2}; a.b += 5; return a.b`, nil, Int(6))
	expectRun(t, `a:={b:1,c:2}; a.b += a.c; return a.b`, nil, Int(3))
	expectRun(t, `a:={b:1,c:2}; a.b += a.c; return a.c`, nil, Int(2))
	expectRun(t, `
	a := {
		b: [1, 2, 3],
		c: {
			d: 8,
			e: "foo",
			f: [9, 8],
		},
	}
	a.c.f[1] += 2
	return a["c"]["f"][1]
	`, nil, Int(10))

	expectRun(t, `
	a := {
		b: [1, 2, 3],
		c: {
			d: 8,
			e: "foo",
			f: [9, 8],
		},
	}
	a.c.h = "bar"
	return a.c.h
	`, nil, String("bar"))

	expectErrIs(t, `
	a := {
		b: [1, 2, 3],
		c: {
			d: 8,
			e: "foo",
			f: [9, 8],
		},
	}
	a.x.e = "bar"`, nil, ErrNotIndexAssignable)

	// order of evaluation
	// left to right but in assignment RHS first then LHS
	expectRun(t, `
	a := 1
	f := func() {
		a*=10
		return a
	}
	g := func() {
		a++
		return a
	}
	h := func() {
		a+=2
		return a
	}
	d := {}
	d[f()] = [g(), h()]
	return d
	`, nil, Map{"40": Array{Int(2), Int(4)}})
}

func TestVMBitwise(t *testing.T) {
	expectRun(t, `return 1 & 1`, nil, Int(1))
	expectRun(t, `return 1 & 0`, nil, Int(0))
	expectRun(t, `return 0 & 1`, nil, Int(0))
	expectRun(t, `return 0 & 0`, nil, Int(0))
	expectRun(t, `return 1 | 1`, nil, Int(1))
	expectRun(t, `return 1 | 0`, nil, Int(1))
	expectRun(t, `return 0 | 1`, nil, Int(1))
	expectRun(t, `return 0 | 0`, nil, Int(0))
	expectRun(t, `return 1 ^ 1`, nil, Int(0))
	expectRun(t, `return 1 ^ 0`, nil, Int(1))
	expectRun(t, `return 0 ^ 1`, nil, Int(1))
	expectRun(t, `return 0 ^ 0`, nil, Int(0))
	expectRun(t, `return 1 &^ 1`, nil, Int(0))
	expectRun(t, `return 1 &^ 0`, nil, Int(1))
	expectRun(t, `return 0 &^ 1`, nil, Int(0))
	expectRun(t, `return 0 &^ 0`, nil, Int(0))
	expectRun(t, `return 1 << 2`, nil, Int(4))
	expectRun(t, `return 16 >> 2`, nil, Int(4))

	expectRun(t, `return 1u & 1u`, nil, Uint(1))
	expectRun(t, `return 1u & 0u`, nil, Uint(0))
	expectRun(t, `return 0u & 1u`, nil, Uint(0))
	expectRun(t, `return 0u & 0u`, nil, Uint(0))
	expectRun(t, `return 1u | 1u`, nil, Uint(1))
	expectRun(t, `return 1u | 0u`, nil, Uint(1))
	expectRun(t, `return 0u | 1u`, nil, Uint(1))
	expectRun(t, `return 0u | 0u`, nil, Uint(0))
	expectRun(t, `return 1u ^ 1u`, nil, Uint(0))
	expectRun(t, `return 1u ^ 0u`, nil, Uint(1))
	expectRun(t, `return 0u ^ 1u`, nil, Uint(1))
	expectRun(t, `return 0u ^ 0u`, nil, Uint(0))
	expectRun(t, `return 1u &^ 1u`, nil, Uint(0))
	expectRun(t, `return 1u &^ 0u`, nil, Uint(1))
	expectRun(t, `return 0u &^ 1u`, nil, Uint(0))
	expectRun(t, `return 0u &^ 0u`, nil, Uint(0))
	expectRun(t, `return 1u << 2u`, nil, Uint(4))
	expectRun(t, `return 16u >> 2u`, nil, Uint(4))

	expectRun(t, `out := 1; out &= 1; return out`, nil, Int(1))
	expectRun(t, `out := 1; out |= 0; return out`, nil, Int(1))
	expectRun(t, `out := 1; out ^= 0; return out`, nil, Int(1))
	expectRun(t, `out := 1; out &^= 0; return out`, nil, Int(1))
	expectRun(t, `out := 1; out <<= 2; return out`, nil, Int(4))
	expectRun(t, `out := 16; out >>= 2; return out`, nil, Int(4))

	expectRun(t, `out := 1u; out &= 1u; return out`, nil, Uint(1))
	expectRun(t, `out := 1u; out |= 0u; return out`, nil, Uint(1))
	expectRun(t, `out := 1u; out ^= 0u; return out`, nil, Uint(1))
	expectRun(t, `out := 1u; out &^= 0u; return out`, nil, Uint(1))
	expectRun(t, `out := 1u; out <<= 2u; return out`, nil, Uint(4))
	expectRun(t, `out := 16u; out >>= 2u; return out`, nil, Uint(4))

	expectRun(t, `out := ^0; return out`, nil, Int(^0))
	expectRun(t, `out := ^1; return out`, nil, Int(^1))
	expectRun(t, `out := ^55; return out`, nil, Int(^55))
	expectRun(t, `out := ^-55; return out`, nil, Int(^-55))

	expectRun(t, `out := ^0u; return out`, nil, Uint(^uint64(0)))
	expectRun(t, `out := ^1u; return out`, nil, Uint(^uint64(1)))
	expectRun(t, `out := ^55u; return out`, nil, Uint(^uint64(55)))
}

func TestVMBoolean(t *testing.T) {
	expectRun(t, `return true`, nil, True)
	expectRun(t, `return false`, nil, False)
	expectRun(t, `return 1 < 2`, nil, True)
	expectRun(t, `return 1 > 2`, nil, False)
	expectRun(t, `return 1 < 1`, nil, False)
	expectRun(t, `return 1 > 2`, nil, False)
	expectRun(t, `return 1 == 1`, nil, True)
	expectRun(t, `return 1 != 1`, nil, False)
	expectRun(t, `return 1 == 2`, nil, False)
	expectRun(t, `return 1 != 2`, nil, True)
	expectRun(t, `return 1 <= 2`, nil, True)
	expectRun(t, `return 1 >= 2`, nil, False)
	expectRun(t, `return 1 <= 1`, nil, True)
	expectRun(t, `return 1 >= 2`, nil, False)

	expectRun(t, `return true == true`, nil, True)
	expectRun(t, `return false == false`, nil, True)
	expectRun(t, `return true == false`, nil, False)
	expectRun(t, `return true != false`, nil, True)
	expectRun(t, `return false != true`, nil, True)
	expectRun(t, `return (1 < 2) == true`, nil, True)
	expectRun(t, `return (1 < 2) == false`, nil, False)
	expectRun(t, `return (1 > 2) == true`, nil, False)
	expectRun(t, `return (1 > 2) == false`, nil, True)
	expectRun(t, `return !true`, nil, False)
	expectRun(t, `return !false`, nil, True)

	expectRun(t, `return 5 + true`, nil, Int(6))
	expectRun(t, `return 5 + false`, nil, Int(5))
	expectRun(t, `return 5 * true`, nil, Int(5))
	expectRun(t, `return 5 * false`, nil, Int(0))
	expectRun(t, `return -true`, nil, Int(-1))
	expectRun(t, `return true + false`, nil, Int(1))
	expectRun(t, `return true*false`, nil, Int(0))
	expectRun(t, `return func() { return true + false }()`, nil, Int(1))
	expectRun(t, `if (true + false) { return 10 }`, nil, Int(10))
	expectRun(t, `return 10 + (true + false)`, nil, Int(11))
	expectRun(t, `return (true + false) + 20`, nil, Int(21))
	expectRun(t, `return !(true + false)`, nil, False)
	expectRun(t, `return !(true - false)`, nil, False)
	expectErrIs(t, `return true/false`, nil, ErrZeroDivision)
	expectErrIs(t, `return 1/false`, nil, ErrZeroDivision)
}

func TestVMUndefined(t *testing.T) {
	expectRun(t, `return undefined`, nil, Undefined)
	expectRun(t, `return undefined.a`, nil, Undefined)
	expectRun(t, `return undefined[1]`, nil, Undefined)
	expectRun(t, `return undefined.a.b`, nil, Undefined)
	expectRun(t, `return undefined[1][2]`, nil, Undefined)
	expectRun(t, `return undefined ? 1 : 2`, nil, Int(2))
	expectRun(t, `return undefined == undefined`, nil, True)
	expectRun(t, `return undefined == (undefined ? 1 : undefined)`,
		nil, True)
	expectRun(t, `return copy(undefined)`, nil, Undefined)
	expectRun(t, `return len(undefined)`, nil, Int(0))

	testCases := []string{
		"true", "false", "0", "1", "1u", `""`, `"a"`, `bytes(0)`, "[]", "{}",
		"[1]", "{a:1}", `'a'`, "1.1", "0.0",
	}
	for _, tC := range testCases {
		t.Run(tC, func(t *testing.T) {
			expectRun(t, fmt.Sprintf(`return undefined == %s`, tC), nil, False)
			expectRun(t, fmt.Sprintf(`return undefined != %s`, tC), nil, True)
			expectRun(t, fmt.Sprintf(`return undefined < %s`, tC), nil, True)
			expectRun(t, fmt.Sprintf(`return undefined <= %s`, tC), nil, True)
			expectRun(t, fmt.Sprintf(`return undefined > %s`, tC), nil, False)
			expectRun(t, fmt.Sprintf(`return undefined >= %s`, tC), nil, False)

			expectRun(t, fmt.Sprintf(`return %s == undefined`, tC), nil, False)
			expectRun(t, fmt.Sprintf(`return %s != undefined`, tC), nil, True)
			expectRun(t, fmt.Sprintf(`return %s > undefined`, tC), nil, True)
			expectRun(t, fmt.Sprintf(`return %s >= undefined`, tC), nil, True)
			expectRun(t, fmt.Sprintf(`return %s < undefined`, tC), nil, False)
			expectRun(t, fmt.Sprintf(`return %s <= undefined`, tC), nil, False)
		})
	}
}

func TestVMBuiltinFunction(t *testing.T) {
	expectRun(t, `return append(undefined)`,
		nil, Array{})
	expectRun(t, `return append(undefined, 1)`,
		nil, Array{Int(1)})
	expectRun(t, `return append([], 1)`,
		nil, Array{Int(1)})
	expectRun(t, `return append([], 1, 2)`,
		nil, Array{Int(1), Int(2)})
	expectRun(t, `return append([0], 1, 2)`,
		nil, Array{Int(0), Int(1), Int(2)})
	expectRun(t, `return append(bytes())`,
		nil, Bytes{})
	expectRun(t, `return append(bytes(), 1, 2)`,
		nil, Bytes{1, 2})
	expectErrIs(t, `append()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `append({})`, nil, ErrType)

	expectRun(t, `out := {}; delete(out, "a"); return out`,
		nil, Map{})
	expectRun(t, `out := {a: 1}; delete(out, "a"); return out`,
		nil, Map{})
	expectRun(t, `out := {a: 1}; delete(out, "b"); return out`,
		nil, Map{"a": Int(1)})
	expectErrIs(t, `delete({})`, nil, ErrWrongNumArguments)
	expectErrIs(t, `delete({}, "", "")`, nil, ErrWrongNumArguments)
	expectErrIs(t, `delete([], "")`, nil, ErrType)
	expectRun(t, `delete({}, 1)`, nil, Undefined)

	g := &SyncMap{Value: Map{"out": &SyncMap{Value: Map{"a": Int(1)}}}}
	expectRun(t, `global out; delete(out, "a"); return out`,
		newOpts().Globals(g).Skip2Pass(), &SyncMap{Value: Map{}})

	expectRun(t, `return copy(undefined)`, nil, Undefined)
	expectRun(t, `return copy(1)`, nil, Int(1))
	expectRun(t, `return copy(1u)`, nil, Uint(1))
	expectRun(t, `return copy('a')`, nil, Char('a'))
	expectRun(t, `return copy(1.0)`, nil, Float(1.0))
	expectRun(t, `return copy("x")`, nil, String("x"))
	expectRun(t, `return copy(true)`, nil, True)
	expectRun(t, `return copy(false)`, nil, False)
	expectRun(t, `a := {x: 1}; b := copy(a); a.x = 2; return b`,
		nil, Map{"x": Int(1)})
	expectRun(t, `a := {x: 1}; b := copy(a); b.x = 2; return a`,
		nil, Map{"x": Int(1)})
	expectRun(t, `a := {x: 1}; b := copy(a); return a == b`,
		nil, True)
	expectRun(t, `a := [1]; b := copy(a); a[0] = 2; return b`,
		nil, Array{Int(1)})
	expectRun(t, `a := [1]; b := copy(a); b[0] = 2; return a`,
		nil, Array{Int(1)})
	expectRun(t, `a := [1]; b := copy(a); return a == b`,
		nil, True)
	expectRun(t, `a := bytes(1); b := copy(a); a[0] = 2; return b`,
		nil, Bytes{1})
	expectRun(t, `a := bytes(1); b := copy(a); b[0] = 2; return a`,
		nil, Bytes{1})
	expectRun(t, `a := bytes(1); b := copy(a); return a == b`,
		nil, True)
	expectErrIs(t, `copy()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `copy(1, 2)`, nil, ErrWrongNumArguments)

	expectRun(t, `return repeat("abc", 3)`, nil, String("abcabcabc"))
	expectRun(t, `return repeat("abc", 2)`, nil, String("abcabc"))
	expectRun(t, `return repeat("abc", 1)`, nil, String("abc"))
	expectRun(t, `return repeat("abc", 0)`, nil, String(""))
	expectRun(t, `return repeat(bytes(1, 2, 3), 3)`,
		nil, Bytes{1, 2, 3, 1, 2, 3, 1, 2, 3})
	expectRun(t, `return repeat(bytes(1, 2, 3), 2)`,
		nil, Bytes{1, 2, 3, 1, 2, 3})
	expectRun(t, `return repeat(bytes(1, 2, 3), 1)`,
		nil, Bytes{1, 2, 3})
	expectRun(t, `return repeat(bytes(1, 2, 3), 0)`,
		nil, Bytes{})
	expectRun(t, `return repeat([1, 2], 2)`,
		nil, Array{Int(1), Int(2), Int(1), Int(2)})
	expectRun(t, `return repeat([1, 2], 1)`,
		nil, Array{Int(1), Int(2)})
	expectRun(t, `return repeat([1, 2], 0)`,
		nil, Array{})
	expectRun(t, `return repeat([true], 1)`, nil, Array{True})
	expectRun(t, `return repeat([true], 2)`, nil, Array{True, True})
	expectRun(t, `return repeat("", 3)`, nil, String(""))
	expectRun(t, `return repeat(bytes(), 3)`, nil, Bytes{})
	expectRun(t, `return repeat([], 2)`, nil, Array{})
	expectErrIs(t, `return repeat("abc", -1)`, nil, ErrType)
	expectErrIs(t, `return repeat(bytes(1), -1)`, nil, ErrType)
	expectErrIs(t, `return repeat([1], -1)`, nil, ErrType)
	expectErrIs(t, `return repeat("abc", "")`, nil, ErrType)
	expectErrIs(t, `return repeat(bytes(1), [])`, nil, ErrType)
	expectErrIs(t, `return repeat([1], {})`, nil, ErrType)
	expectErrIs(t, `return repeat(undefined, 1)`, nil, ErrType)
	expectErrIs(t, `return repeat(true, 1)`, nil, ErrType)
	expectErrIs(t, `return repeat(false, 1)`, nil, ErrType)
	expectErrIs(t, `return repeat(1, 1)`, nil, ErrType)
	expectErrIs(t, `return repeat(1u, 1)`, nil, ErrType)
	expectErrIs(t, `return repeat(1.1, 1)`, nil, ErrType)
	expectErrIs(t, `return repeat('a', 1)`, nil, ErrType)
	expectErrIs(t, `return repeat({}, 1)`, nil, ErrType)

	expectRun(t, `return contains("xyz", "y")`, nil, True)
	expectRun(t, `return contains("xyz", "a")`, nil, False)
	expectRun(t, `return contains({a: 1}, "a")`, nil, True)
	expectRun(t, `return contains({a: 1}, "b")`, nil, False)
	expectRun(t, `return contains([1, 2, 3], 2)`, nil, True)
	expectRun(t, `return contains([1, 2, 3], 4)`, nil, False)
	expectRun(t, `return contains(bytes(1, 2, 3), 3)`, nil, True)
	expectRun(t, `return contains(bytes(1, 2, 3), 4)`, nil, False)
	expectRun(t, `return contains(bytes("abc"), "b")`, nil, True)
	expectRun(t, `return contains(bytes("abc"), "d")`, nil, False)
	expectRun(t, `return contains(bytes(1, 2, 3, 4), bytes(2, 3))`, nil, True)
	expectRun(t, `return contains(bytes(1, 2, 3, 4), bytes(1, 3))`, nil, False)
	expectRun(t, `return contains(undefined, "")`, nil, False)
	expectRun(t, `return contains(undefined, 1)`, nil, False)
	g = &SyncMap{Value: Map{"a": Int(1)}}
	expectRun(t, `return contains(globals(), "a")`,
		newOpts().Globals(g).Skip2Pass(), True)
	expectErrIs(t, `contains()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `contains("", "", "")`, nil, ErrWrongNumArguments)
	expectErrIs(t, `contains(1, 2)`, nil, ErrType)

	expectRun(t, `return len(undefined)`, nil, Int(0))
	expectRun(t, `return len(1)`, nil, Int(0))
	expectRun(t, `return len(1u)`, nil, Int(0))
	expectRun(t, `return len(true)`, nil, Int(0))
	expectRun(t, `return len(1.1)`, nil, Int(0))
	expectRun(t, `return len("")`, nil, Int(0))
	expectRun(t, `return len([])`, nil, Int(0))
	expectRun(t, `return len({})`, nil, Int(0))
	expectRun(t, `return len(bytes())`, nil, Int(0))
	expectRun(t, `return len("xyzw")`, nil, Int(4))
	expectRun(t, `return len("çığöşü")`, nil, Int(12))
	expectRun(t, `return len(chars("çığöşü"))`, nil, Int(6))
	expectRun(t, `return len(["a"])`, nil, Int(1))
	expectRun(t, `return len({a: 2})`, nil, Int(1))
	expectRun(t, `return len(bytes(0, 1, 2))`, nil, Int(3))
	g = &SyncMap{Value: Map{"a": Int(5)}}
	expectRun(t, `return len(globals())`,
		newOpts().Globals(g).Skip2Pass(), Int(1))
	expectErrIs(t, `len()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `len([], [])`, nil, ErrWrongNumArguments)

	expectRun(t, `return cap(undefined)`, nil, Int(0))
	expectRun(t, `return cap(1)`, nil, Int(0))
	expectRun(t, `return cap(1u)`, nil, Int(0))
	expectRun(t, `return cap(true)`, nil, Int(0))
	expectRun(t, `return cap(1.1)`, nil, Int(0))
	expectRun(t, `return cap("")`, nil, Int(0))
	expectRun(t, `return cap([])`, nil, Int(0))
	expectRun(t, `return cap({})`, nil, Int(0))
	expectRun(t, `return cap(bytes())`, nil, Int(0))
	expectRun(t, `return cap(bytes("a"))>=1`, nil, True)
	expectRun(t, `return cap(bytes("abc"))>=3`, nil, True)
	expectRun(t, `return cap(bytes("abc")[:3])>=3`, nil, True)
	expectRun(t, `return cap([1])>0`, nil, True)
	expectRun(t, `return cap([1,2,3])>=3`, nil, True)
	expectRun(t, `return cap([1,2,3][:3])>=3`, nil, True)

	expectRun(t, `return sort(undefined)`,
		nil, Undefined)
	expectRun(t, `return sort("acb")`,
		nil, String("abc"))
	expectRun(t, `return sort(bytes("acb"))`,
		nil, Bytes(String("abc")))
	expectRun(t, `return sort([3, 2, 1])`,
		nil, Array{Int(1), Int(2), Int(3)})
	expectRun(t, `return sort([3u, 2.0, 1])`,
		nil, Array{Int(1), Float(2), Uint(3)})
	expectRun(t, `a := [3, 2, 1]; sort(a); return a`,
		nil, Array{Int(1), Int(2), Int(3)})
	expectErrIs(t, `sort()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `sort([], [])`, nil, ErrWrongNumArguments)
	expectErrIs(t, `sort({})`, nil, ErrType)

	expectRun(t, `return sortReverse(undefined)`,
		nil, Undefined)
	expectRun(t, `return sortReverse("acb")`,
		nil, String("cba"))
	expectRun(t, `return sortReverse(bytes("acb"))`,
		nil, Bytes(String("cba")))
	expectRun(t, `return sortReverse([1, 2, 3])`,
		nil, Array{Int(3), Int(2), Int(1)})
	expectRun(t, `a := [1, 2, 3]; sortReverse(a); return a`,
		nil, Array{Int(3), Int(2), Int(1)})
	expectErrIs(t, `sortReverse()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `sortReverse([], [])`, nil, ErrWrongNumArguments)
	expectErrIs(t, `sortReverse({})`, nil, ErrType)

	expectRun(t, `return error("x")`, nil,
		&Error{Name: "error", Message: "x"})
	expectRun(t, `return error(1)`, nil,
		&Error{Name: "error", Message: "1"})
	expectRun(t, `return error(undefined)`, nil,
		&Error{Name: "error", Message: "undefined"})
	expectErrIs(t, `error()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `error(1,2,3)`, nil, ErrWrongNumArguments)

	expectRun(t, `return typeName(true)`, nil, String("bool"))
	expectRun(t, `return typeName(undefined)`, nil, String("undefined"))
	expectRun(t, `return typeName(1)`, nil, String("int"))
	expectRun(t, `return typeName(1u)`, nil, String("uint"))
	expectRun(t, `return typeName(1.1)`, nil, String("float"))
	expectRun(t, `return typeName('a')`, nil, String("char"))
	expectRun(t, `return typeName("")`, nil, String("string"))
	expectRun(t, `return typeName([])`, nil, String("array"))
	expectRun(t, `return typeName({})`, nil, String("map"))
	expectRun(t, `return typeName(error(""))`, nil, String("error"))
	expectRun(t, `return typeName(bytes())`, nil, String("bytes"))
	expectRun(t, `return typeName(func(){})`, nil, String("compiledFunction"))
	expectRun(t, `return typeName(append)`, nil, String("builtinFunction"))
	expectErrIs(t, `typeName()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `typeName("", "")`, nil, ErrWrongNumArguments)

	convs := []struct {
		f      string
		inputs map[string]Object
	}{
		{
			"int",
			map[string]Object{
				"1":       Int(1),
				"1u":      Int(1),
				"1.0":     Int(1),
				`'\x01'`:  Int(1),
				`'a'`:     Int(97),
				"true":    Int(1),
				"false":   Int(0),
				`"1"`:     Int(1),
				`"+123"`:  Int(123),
				`"-123"`:  Int(-123),
				`"0x10"`:  Int(16),
				`"0b101"`: Int(5),
			},
		},
		{
			"uint",
			map[string]Object{
				"1":       Uint(1),
				"1u":      Uint(1),
				"1.0":     Uint(1),
				`'\x01'`:  Uint(1),
				`'a'`:     Uint(97),
				"true":    Uint(1),
				"false":   Uint(0),
				`"1"`:     Uint(1),
				"-1":      ^Uint(0),
				`"0x10"`:  Uint(16),
				`"0b101"`: Uint(5),
			},
		},
		{
			"char",
			map[string]Object{
				"1":      Char(1),
				"1u":     Char(1),
				"1.1":    Char(1),
				`'\x01'`: Char(1),
				"true":   Char(1),
				"false":  Char(0),
				`"1"`:    Char('1'),
				`""`:     Undefined,
			},
		},
		{
			"float",
			map[string]Object{
				"1":      Float(1.0),
				"1u":     Float(1.0),
				"1.0":    Float(1.0),
				`'\x01'`: Float(1.0),
				"true":   Float(1.0),
				"false":  Float(0.0),
				`"1"`:    Float(1.0),
				`"1.1"`:  Float(1.1),
			},
		},
		{
			"string",
			map[string]Object{
				"1":                 String("1"),
				"1u":                String("1"),
				"1.0":               String("1"),
				`'\x01'`:            String("\x01"),
				"true":              String("true"),
				"false":             String("false"),
				`"1"`:               String("1"),
				`"1.1"`:             String("1.1"),
				`undefined`:         String("undefined"),
				`[]`:                String("[]"),
				`[1]`:               String("[1]"),
				`[1, 2]`:            String("[1, 2]"),
				`{}`:                String("{}"),
				`{a: 1}`:            String(`{"a": 1}`),
				`error("an error")`: String(`error: an error`),
			},
		},
		{
			"bytes",
			map[string]Object{
				"1":           Bytes{1},
				"1u":          Bytes{1},
				`'\x01'`:      Bytes{1},
				"1, 2u":       Bytes{1, 2},
				"1, '\x02'":   Bytes{1, 2},
				"1u, 2":       Bytes{1, 2},
				`'\x01', 2u`:  Bytes{1, 2},
				`'\x01', 2`:   Bytes{1, 2},
				`bytes(1, 2)`: Bytes{1, 2},
				`"abc"`:       Bytes{'a', 'b', 'c'},
			},
		},
		{
			"chars",
			map[string]Object{
				`""`:             Array{},
				`"abc"`:          Array{Char('a'), Char('b'), Char('c')},
				`bytes("abc")`:   Array{Char('a'), Char('b'), Char('c')},
				`"a\xc5"`:        Undefined, // incorrect UTF-8
				`bytes("a\xc5")`: Undefined, // incorrect UTF-8
			},
		},
	}
	for i, conv := range convs {
		for k, v := range conv.inputs {
			t.Run(fmt.Sprintf("%s#%d#%v", conv.f, i, k), func(t *testing.T) {
				expectRun(t, fmt.Sprintf(`return %s(%s)`, conv.f, k), nil, v)
			})
		}
	}

	expectErrIs(t, `int(1, 2)`, nil, ErrWrongNumArguments)
	expectErrIs(t, `uint(1, 2)`, nil, ErrWrongNumArguments)
	expectErrIs(t, `char(1, 2)`, nil, ErrWrongNumArguments)
	expectErrIs(t, `float(1, 2)`, nil, ErrWrongNumArguments)
	expectErrIs(t, `string(1, 2)`, nil, ErrWrongNumArguments)
	expectErrIs(t, `chars(1, 2)`, nil, ErrWrongNumArguments)

	expectErrIs(t, `int([])`, nil, ErrType)
	expectErrIs(t, `uint([])`, nil, ErrType)
	expectErrIs(t, `char([])`, nil, ErrType)
	expectErrIs(t, `float([])`, nil, ErrType)
	expectErrIs(t, `chars([])`, nil, ErrType)
	expectErrIs(t, `bytes(1, 2, "")`, nil, ErrType)

	type trueValues []string
	type falseValues []string

	isfuncs := []struct {
		f           string
		trueValues  trueValues
		falseValues falseValues
	}{
		{
			`isError`,
			trueValues{
				`error("test")`,
			},
			falseValues{
				"1", "1u", `""`, "1.1", "'\x01'", `bytes()`, "undefined",
				"true", "false", "[]", "{}",
			},
		},
		{
			`isInt`,
			trueValues{
				"0", "1", "-1",
			},
			falseValues{
				"1u", `""`, "1.1", "'\x01'", `bytes()`, "undefined",
				`error("x")`,
				"true", "false", "[]", "{}",
			},
		},
		{
			`isUint`,
			trueValues{
				"0u", "1u", "-1u",
			},
			falseValues{
				"1", "-1", `""`, "1.1", "'\x01'", `bytes()`, "undefined",
				`error("x")`, "true", "false", "[]", "{}",
			},
		},
		{
			`isFloat`,
			trueValues{
				"0.0", "1.0", "-1.0",
			},
			falseValues{
				"1", "-1", `""`, "1u", "'\x01'", `bytes()`, "undefined",
				`error("x")`, "true", "false", "[]", "{}",
			},
		},
		{
			`isChar`,
			trueValues{
				"'\x01'", `'a'`, `'b'`,
			},
			falseValues{
				"1", "-1", `""`, "1u", "1.1", `bytes()`, "undefined",
				`error("x")`, "true", "false", "[]", "{}",
			},
		},
		{
			`isBool`,
			trueValues{
				"true", "false",
			},
			falseValues{
				"1", "-1", `""`, "1u", "1.1", "'\x01'", `bytes()`, "undefined",
				`error("x")`, "[]", "{}",
			},
		},
		{
			`isString`,
			trueValues{
				`""`, `"abc"`,
			},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", `bytes()`, "undefined",
				`error("x")`, "true", "false", "[]", "{}",
			},
		},
		{
			`isBytes`,
			trueValues{
				`bytes()`, `bytes(1, 2)`,
			},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", `""`, "undefined",
				`error("x")`, "true", "false", "[]", "{}",
			},
		},
		{
			`isMap`,
			trueValues{
				`{}`, `{a: 1}`,
			},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", `""`, `bytes()`, "undefined",
				`error("x")`, "true", "false", "[]",
			},
		},
		{
			`isSyncMap`,
			trueValues{},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", `""`, `bytes()`, "undefined",
				`error("x")`, "true", "false", "[]", "{}",
			},
		},
		{
			`isArray`,
			trueValues{
				`[]`, `[0]`,
			},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", `""`, `bytes()`, "undefined",
				`error("x")`, "true", "false", "{}",
			},
		},
		{
			`isUndefined`,
			trueValues{
				`undefined`,
			},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", `""`, `bytes()`, `error("x")`,
				"true", "false", "{}", "[]",
			},
		},
		{
			`isFunction`,
			trueValues{
				`len`, `append`, `func(){}`,
			},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", `""`, `bytes()`, "undefined",
				`error("x")`, "true", "false", "{}", "[]",
			},
		},
		{
			`isCallable`,
			trueValues{
				`len`, `append`,
			},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", `""`, `bytes()`, "undefined",
				`error("x")`, "true", "false", "{}", "[]",
			},
		},
		{
			`isIterable`,
			trueValues{
				`[]`, `{}`, `"abc"`, `""`, `bytes()`,
			},
			falseValues{
				"1", "-1", "1u", "1.1", "'\x01'", "undefined", `error("x")`,
				"true", "false",
			},
		},
		{
			`bool`,
			trueValues{
				"1", "1u", "-1", "1.1", "'\x01'", "true", `"abc"`, `bytes(1)`,
			},
			falseValues{
				"0", "0u", "undefined", `error("x")`, "false", `[]`, `{}`, `""`, `bytes()`,
			},
		},
	}
	for i, isfunc := range isfuncs {
		for _, v := range isfunc.trueValues {
			t.Run(fmt.Sprintf("%s#%d %v true", isfunc.f, i, v), func(t *testing.T) {
				expectRun(t, fmt.Sprintf(`return %s(%s)`, isfunc.f, v), nil, True)
			})
		}
		for _, v := range isfunc.falseValues {
			t.Run(fmt.Sprintf("%s#%d %v false", isfunc.f, i, v), func(t *testing.T) {
				expectRun(t, fmt.Sprintf(`return %s(%s)`, isfunc.f, v), nil, False)
			})
		}
		if isfunc.f != "isError" {
			t.Run(fmt.Sprintf("%s#%d 2args", isfunc.f, i), func(t *testing.T) {
				expectErrIs(t, fmt.Sprintf(`%s(undefined, undefined)`, isfunc.f),
					nil, ErrWrongNumArguments)
			})
		} else {
			t.Run(fmt.Sprintf("%s#%d 3args", isfunc.f, i), func(t *testing.T) {
				expectErrIs(t, fmt.Sprintf(`%s(undefined, undefined, undefined)`, isfunc.f),
					nil, ErrWrongNumArguments)
			})
		}
	}

	expectRun(t, `global sm; return isSyncMap(sm)`,
		newOpts().Globals(Map{"sm": &SyncMap{Value: Map{}}}), True)

	expectRun(t, `return isError(WrongNumArgumentsError.New(""), WrongNumArgumentsError)`,
		nil, True)
	expectRun(t, `
	f := func(){ 
		throw NotImplementedError.New("test") 
	}
	try {
		f()
	} catch err {
		return isError(err, NotImplementedError)
	}`, nil, True)

	var stdOut bytes.Buffer
	oldWriter := PrintWriter
	PrintWriter = &stdOut
	defer func() {
		PrintWriter = oldWriter
	}()
	stdOut.Reset()
	expectRun(t, `printf("test")`, newOpts().Skip2Pass(), Undefined)
	require.Equal(t, "test", stdOut.String())

	stdOut.Reset()
	expectRun(t, `printf("test %d", 1)`, newOpts().Skip2Pass(), Undefined)
	require.Equal(t, "test 1", stdOut.String())

	stdOut.Reset()
	expectRun(t, `printf("test %d %d", 1, 2u)`, newOpts().Skip2Pass(), Undefined)
	require.Equal(t, "test 1 2", stdOut.String())

	stdOut.Reset()
	expectRun(t, `println()`, newOpts().Skip2Pass(), Undefined)
	require.Equal(t, "\n", stdOut.String())

	stdOut.Reset()
	expectRun(t, `println("test")`, newOpts().Skip2Pass(), Undefined)
	require.Equal(t, "test\n", stdOut.String())

	stdOut.Reset()
	expectRun(t, `println("test", 1)`, newOpts().Skip2Pass(), Undefined)
	require.Equal(t, "test 1\n", stdOut.String())

	stdOut.Reset()
	expectRun(t, `println("test", 1, 2u)`, newOpts().Skip2Pass(), Undefined)
	require.Equal(t, "test 1 2\n", stdOut.String())

	expectRun(t, `return sprintf("test")`,
		newOpts().Skip2Pass(), String("test"))
	expectRun(t, `return sprintf("test %d", 1)`,
		newOpts().Skip2Pass(), String("test 1"))
	expectRun(t, `return sprintf("test %d %t", 1, true)`,
		newOpts().Skip2Pass(), String("test 1 true"))

	expectErrIs(t, `printf()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `sprintf()`, nil, ErrWrongNumArguments)
}

func TestBytes(t *testing.T) {
	expectRun(t, `return bytes("Hello World!")`, nil, Bytes("Hello World!"))
	expectRun(t, `return bytes("Hello") + bytes(" ") + bytes("World!")`,
		nil, Bytes("Hello World!"))
	expectRun(t, `return bytes("Hello") + bytes(" ") + "World!"`,
		nil, Bytes("Hello World!"))
	expectRun(t, `return "Hello " + bytes("World!")`,
		nil, String("Hello World!"))

	//slice
	expectRun(t, `return bytes("")[:]`, nil, Bytes{})
	expectRun(t, `return bytes("abcde")[:]`, nil, Bytes(String("abcde")))
	expectRun(t, `return bytes("abcde")[0:]`, nil, Bytes(String("abcde")))
	expectRun(t, `return bytes("abcde")[:0]`, nil, Bytes{})
	expectRun(t, `return bytes("abcde")[:1]`, nil, Bytes(String("a")))
	expectRun(t, `return bytes("abcde")[:2]`, nil, Bytes(String("ab")))
	expectRun(t, `return bytes("abcde")[0:2]`, nil, Bytes(String("ab")))
	expectRun(t, `return bytes("abcde")[1:]`, nil, Bytes(String("bcde")))
	expectRun(t, `return bytes("abcde")[1:5]`, nil, Bytes(String("bcde")))
	expectRun(t, `
	b1 := bytes("abcde")
	b2 := b1[:2]
	return b2[:len(b1)]`, nil, Bytes(String("abcde")))
	expectRun(t, `
	b1 := bytes("abcde")
	b2 := b1[:2]
	return cap(b1) == cap(b2)`, nil, True)

	// bytes[] -> int
	expectRun(t, `return bytes("abcde")[0]`, nil, Int('a'))
	expectRun(t, `return bytes("abcde")[1]`, nil, Int('b'))
	expectRun(t, `return bytes("abcde")[4]`, nil, Int('e'))
	expectErrIs(t, `return bytes("abcde")[-1]`, nil, ErrIndexOutOfBounds)
	expectErrIs(t, `return bytes("abcde")[100]`, nil, ErrIndexOutOfBounds)
	expectErrIs(t, `b1 := bytes("abcde");	b2 := b1[:cap(b1)+1]`, nil, ErrIndexOutOfBounds)
}

func TestVMChar(t *testing.T) {
	expectRun(t, `return 'a'`, nil, Char('a'))
	expectRun(t, `return '九'`, nil, Char(20061))
	expectRun(t, `return 'Æ'`, nil, Char(198))
	expectRun(t, `return '0' + '9'`, nil, Char(105))
	expectRun(t, `return '0' + 9`, nil, Char('9'))
	expectRun(t, `return 1 + '9'`, nil, Char(1)+Char('9'))
	expectRun(t, `return '9' - 4`, nil, Char('5'))
	expectRun(t, `return '0' == '0'`, nil, True)
	expectRun(t, `return '0' != '0'`, nil, False)
	expectRun(t, `return '2' < '4'`, nil, True)
	expectRun(t, `return '2' > '4'`, nil, False)
	expectRun(t, `return '2' <= '4'`, nil, True)
	expectRun(t, `return '2' >= '4'`, nil, False)
	expectRun(t, `return '4' < '4'`, nil, False)
	expectRun(t, `return '4' > '4'`, nil, False)
	expectRun(t, `return '4' <= '4'`, nil, True)
	expectRun(t, `return '4' >= '4'`, nil, True)
	expectRun(t, `return '九' + "Hello"`, nil, String("九Hello"))
	expectRun(t, `return "Hello" + '九'`, nil, String("Hello九"))
}

func TestVMCondExpr(t *testing.T) {
	expectRun(t, `true ? 5 : 10`, nil, Undefined)
	expectRun(t, `false ? 5 : 10; var a; return a`, nil, Undefined)
	expectRun(t, `return true ? 5 : 10`, nil, Int(5))
	expectRun(t, `return false ? 5 : 10`, nil, Int(10))
	expectRun(t, `return (1 == 1) ? 2 + 3 : 12 - 2`, nil, Int(5))
	expectRun(t, `return (1 != 1) ? 2 + 3 : 12 - 2`, nil, Int(10))
	expectRun(t, `return (1 == 1) ? true ? 10 - 8 : 1 + 3 : 12 - 2`, nil, Int(2))
	expectRun(t, `return (1 == 1) ? false ? 10 - 8 : 1 + 3 : 12 - 2`, nil, Int(4))

	expectRun(t, `
	out := 0
	f1 := func() { out += 10 }
	f2 := func() { out = -out }
	true ? f1() : f2()
	return out
	`, nil, Int(10))
	expectRun(t, `
	out := 5
	f1 := func() { out += 10 }
	f2 := func() { out = -out }
	false ? f1() : f2()
	return out
	`, nil, Int(-5))
	expectRun(t, `
	f1 := func(a) { return a + 2 }
	f2 := func(a) { return a - 2 }
	f3 := func(a) { return a + 10 }
	f4 := func(a) { return -a }

	f := func(c) {
		return c == 0 ? f1(c) : f2(c) ? f3(c) : f4(c)
	}

	return [f(0), f(1), f(2)]
	`, nil, Array{Int(2), Int(11), Int(-2)})

	expectRun(t, `f := func(a) { return -a }; return f(true ? 5 : 3)`, nil, Int(-5))
	expectRun(t, `return [false?5:10, true?1:2]`, nil, Array{Int(10), Int(1)})

	expectRun(t, `
	return 1 > 2 ?
		1 + 2 + 3 :
		10 - 5`, nil, Int(5))
}

func TestVMEquality(t *testing.T) {
	testEquality(t, `1`, `1`, true)
	testEquality(t, `1`, `2`, false)

	testEquality(t, `1.0`, `1.0`, true)
	testEquality(t, `1.0`, `1.1`, false)

	testEquality(t, `true`, `true`, true)
	testEquality(t, `true`, `false`, false)

	testEquality(t, `"foo"`, `"foo"`, true)
	testEquality(t, `"foo"`, `"bar"`, false)

	testEquality(t, `'f'`, `'f'`, true)
	testEquality(t, `'f'`, `'b'`, false)

	testEquality(t, `[]`, `[]`, true)
	testEquality(t, `[1]`, `[1]`, true)
	testEquality(t, `[1]`, `[1, 2]`, false)
	testEquality(t, `["foo", "bar"]`, `["foo", "bar"]`, true)
	testEquality(t, `["foo", "bar"]`, `["bar", "foo"]`, false)

	testEquality(t, `{}`, `{}`, true)
	testEquality(t, `{a: 1, b: 2}`, `{b: 2, a: 1}`, true)
	testEquality(t, `{a: 1, b: 2}`, `{b: 2}`, false)
	testEquality(t, `{a: 1, b: {}}`, `{b: {}, a: 1}`, true)

	testEquality(t, `1`, `"foo"`, false)
	testEquality(t, `1`, `true`, true)
	testEquality(t, `[1]`, `["1"]`, false)
	testEquality(t, `[1, [2]]`, `[1, ["2"]]`, false)
	testEquality(t, `{a: 1}`, `{a: "1"}`, false)
	testEquality(t, `{a: 1, b: {c: 2}}`, `{a: 1, b: {c: "2"}}`, false)
}

func testEquality(t *testing.T, lhs, rhs string, expected bool) {
	t.Helper()
	// 1. equality is commutative
	// 2. equality and inequality must be always opposite
	expectRun(t, fmt.Sprintf("return %s == %s", lhs, rhs), nil, Bool(expected))
	expectRun(t, fmt.Sprintf("return %s == %s", rhs, lhs), nil, Bool(expected))
	expectRun(t, fmt.Sprintf("return %s != %s", lhs, rhs), nil, Bool(!expected))
	expectRun(t, fmt.Sprintf("return %s != %s", rhs, lhs), nil, Bool(!expected))
}

func TestVMBuiltinError(t *testing.T) {
	expectRun(t, `return error(1)`, nil, &Error{Name: "error", Message: "1"})
	expectRun(t, `return error(1).Name`, nil, String("error"))
	expectRun(t, `return error(1).Message`, nil, String("1"))
	expectRun(t, `return error("some error")`, nil,
		&Error{Name: "error", Message: "some error"})
	expectRun(t, `return error("some" + " error")`, nil,
		&Error{Name: "error", Message: "some error"})

	expectRun(t, `return func() { return error(5) }()`, nil,
		&Error{Name: "error", Message: "5"})
	expectRun(t, `return error(error("foo"))`, nil, &Error{Name: "error", Message: "error: foo"})

	expectRun(t, `return error("some error").Name`, nil, String("error"))
	expectRun(t, `return error("some error")["Name"]`, nil, String("error"))
	expectRun(t, `return error("some error").Message`, nil, String("some error"))
	expectRun(t, `return error("some error")["Message"]`, nil, String("some error"))

	expectRun(t, `error("error").err`, nil, Undefined)
	expectRun(t, `error("error").value_`, nil, Undefined)
	expectRun(t, `error([1,2,3])[1]`, nil, Undefined)
}

func TestVMFloat(t *testing.T) {
	expectRun(t, `return 0.0`, nil, Float(0.0))
	expectRun(t, `return -10.3`, nil, Float(-10.3))
	expectRun(t, `return 3.2 + 2.0 * -4.0`, nil, Float(-4.8))
	expectRun(t, `return 4 + 2.3`, nil, Float(6.3))
	expectRun(t, `return 2.3 + 4`, nil, Float(6.3))
	expectRun(t, `return +5.0`, nil, Float(5.0))
	expectRun(t, `return -5.0 + +5.0`, nil, Float(0.0))
}

func TestVMForIn(t *testing.T) {
	// array
	expectRun(t, `out := 0; for x in [1, 2, 3] { out += x }; return out`,
		nil, Int(6)) // value
	expectRun(t, `out := 0; for i, x in [1, 2, 3] { out += i + x }; return out`,
		nil, Int(9)) // index, value
	expectRun(t, `out := 0; func() { for i, x in [1, 2, 3] { out += i + x } }(); return out`,
		nil, Int(9)) // index, value
	expectRun(t, `out := 0; for i, _ in [1, 2, 3] { out += i }; return out`,
		nil, Int(3)) // index, _
	expectRun(t, `out := 0; func() { for i, _ in [1, 2, 3] { out += i  } }(); return out`,
		nil, Int(3)) // index, _

	// map
	expectRun(t, `out := 0; for v in {a:2,b:3,c:4} { out += v }; return out`,
		nil, Int(9)) // value
	expectRun(t, `out := ""; for k, v in {a:2,b:3,c:4} { out = k; if v==3 { break } }; return out`,
		nil, String("b")) // key, value
	expectRun(t, `out := ""; for k, _ in {a:2} { out += k }; return out`,
		nil, String("a")) // key, _
	expectRun(t, `out := 0; for _, v in {a:2,b:3,c:4} { out += v }; return out`,
		nil, Int(9)) // _, value
	expectRun(t, `out := ""; func() { for k, v in {a:2,b:3,c:4} { out = k; if v==3 { break } } }(); return out`,
		nil, String("b")) // key, value

	// syncMap
	g := Map{"syncMap": &SyncMap{Value: Map{"a": Int(2), "b": Int(3), "c": Int(4)}}}
	expectRun(t, `out := 0; for v in globals().syncMap { out += v }; return out`,
		newOpts().Globals(g).Skip2Pass(), Int(9)) // value
	expectRun(t, `out := ""; for k, v in globals().syncMap { out = k; if v==3 { break } }; return out`,
		newOpts().Globals(g).Skip2Pass(), String("b")) // key, value
	expectRun(t, `out := ""; for k, _ in globals().syncMap { out += k }; return out`,
		newOpts().Globals(Map{"syncMap": &SyncMap{Value: Map{"a": Int(2)}}}).Skip2Pass(), String("a")) // key, _
	expectRun(t, `out := 0; for _, v in globals().syncMap { out += v }; return out`,
		newOpts().Globals(g).Skip2Pass(), Int(9)) // _, value
	expectRun(t, `out := ""; func() { for k, v in globals().syncMap { out = k; if v==3 { break } } }(); return out`,
		newOpts().Globals(g).Skip2Pass(), String("b")) // key, value

	// string
	expectRun(t, `out := ""; for c in "abcde" { out += c }; return out`, nil, String("abcde"))
	expectRun(t, `out := ""; for i, c in "abcde" { if i == 2 { continue }; out += c }; return out`,
		nil, String("abde"))

	// bytes
	expectRun(t, `out := ""; for c in bytes("abcde") { out += char(c) }; return out`, nil, String("abcde"))
	expectRun(t, `out := ""; for i, c in bytes("abcde") { if i == 2 { continue }; out += char(c) }; return out`,
		nil, String("abde"))

	expectErrIs(t, `a := 1; for k,v in a {}`, nil, ErrNotIterable)
}

func TestFor(t *testing.T) {
	expectRun(t, `
	out := 0
	for {
		out++
		if out == 5 {
			break
		}
	}
	return out`, nil, Int(5))

	expectRun(t, `
	out := 0
	a := 0
	for {
		a++
		if a == 3 { continue }
		if a == 5 { break }
		out += a
	}
	return out`, nil, Int(7)) // 1 + 2 + 4

	expectRun(t, `
	out := 0
	a := 0
	for {
		a++
		if a == 3 { continue }
		out += a
		if a == 5 { break }
	}
	return out`, nil, Int(12)) // 1 + 2 + 4 + 5

	expectRun(t, `
	out := 0
	for true {
		out++
		if out == 5 {
			break
		}
	}
	return out`, nil, Int(5))

	expectRun(t, `
	a := 0
	for true {
		a++
		if a == 5 {
			break
		}
	}
	return a`, nil, Int(5))

	expectRun(t, `
	out := 0
	a := 0
	for true {
		a++
		if a == 3 { continue }
		if a == 5 { break }
		out += a
	}
	return out`, nil, Int(7)) // 1 + 2 + 4

	expectRun(t, `
	out := 0
	a := 0
	for true {
		a++
		if a == 3 { continue }
		out += a
		if a == 5 { break }
	}
	return out`, nil, Int(12)) // 1 + 2 + 4 + 5

	expectRun(t, `
	out := 0
	func() {
		for true {
			out++
			if out == 5 {
				return
			}
		}
	}()
	return out`, nil, Int(5))

	expectRun(t, `
	out := 0
	for a:=1; a<=10; a++ {
		out += a
	}
	return out`, nil, Int(55))

	expectRun(t, `
	out := 0
	for a:=1; a<=3; a++ {
		for b:=3; b<=6; b++ {
			out += b
		}
	}
	return out`, nil, Int(54))

	expectRun(t, `
	out := 0
	func() {
		for {
			out++
			if out == 5 {
				break
			}
		}
	}()
	return out`, nil, Int(5))

	expectRun(t, `
	out := 0
	func() {
		for true {
			out++
			if out == 5 {
				break
			}
		}
	}()
	return out`, nil, Int(5))

	expectRun(t, `
	return func() {
		a := 0
		for {
			a++
			if a == 5 {
				break
			}
		}
		return a
	}()`, nil, Int(5))

	expectRun(t, `
	return func() {
		a := 0
		for true {
			a++
			if a== 5 {
				break
			}
		}
		return a
	}()`, nil, Int(5))

	expectRun(t, `
	return func() {
		a := 0
		func() {
			for {
				a++
				if a == 5 {
					break
				}
			}
		}()
		return a
	}()`, nil, Int(5))

	expectRun(t, `
	return func() {
		a := 0
		func() {
			for true {
				a++
				if a == 5 {
					break
				}
			}
		}()
		return a
	}()`, nil, Int(5))

	expectRun(t, `
	return func() {
		sum := 0
		for a:=1; a<=10; a++ {
			sum += a
		}
		return sum
	}()`, nil, Int(55))

	expectRun(t, `
	return func() {
		sum := 0
		for a:=1; a<=4; a++ {
			for b:=3; b<=5; b++ {
				sum += b
			}
		}
		return sum
	}()`, nil, Int(48)) // (3+4+5) * 4

	expectRun(t, `
	a := 1
	for ; a<=10; a++ {
		if a == 5 {
			break
		}
	}
	return a`, nil, Int(5))

	expectRun(t, `
	out := 0
	for a:=1; a<=10; a++ {
		if a == 3 {
			continue
		}
		out += a
		if a == 5 {
			break
		}
	}
	return out`, nil, Int(12)) // 1 + 2 + 4 + 5

	expectRun(t, `
	out := 0
	for a:=1; a<=10; {
		if a == 3 {
			a++
			continue
		}
		out += a
		if a == 5 {
			break
		}
		a++
	}
	return out`, nil, Int(12)) // 1 + 2 + 4 + 5
}

func TestVMFunction(t *testing.T) {
	// function with no "return" statement returns undefined value.
	expectRun(t, `f1 := func() {}; return f1()`, nil, Undefined)
	expectRun(t, `f1 := func() {}; f2 := func() { return f1(); }; f1(); return f2()`,
		nil, Undefined)
	expectRun(t, `f := func(x) { x; }; return f(5);`, nil, Undefined)

	expectRun(t, `f := func(...x) { return x; }; return f(1, 2, 3);`,
		nil, Array{Int(1), Int(2), Int(3)})

	expectRun(t, `f := func(a, b, ...x) { return [a, b, x]; }; return f(8, 9, 1, 2, 3);`,
		nil, Array{Int(8), Int(9), Array{Int(1), Int(2), Int(3)}})

	expectRun(t, `f := func(v) { x := 2; return func(a, ...b){ return [a, b, v+x]}; }; return f(5)("a", "b");`,
		nil, Array{String("a"), Array{String("b")}, Int(7)})

	expectRun(t, `f := func(...x) { return x; }; return f();`, nil, Array{})

	expectRun(t, `f := func(a, b, ...x) { return [a, b, x]; }; return f(8, 9);`,
		nil, Array{Int(8), Int(9), Array{}})

	expectRun(t, `f := func(v) { x := 2; return func(a, ...b){ return [a, b, v+x]}; }; return f(5)("a");`,
		nil, Array{String("a"), Array{}, Int(7)})

	expectErrIs(t, `f := func(a, b, ...x) { return [a, b, x]; }; f();`, nil, ErrWrongNumArguments)
	expectErrHas(t, `f := func(a, b, ...x) { return [a, b, x]; }; f();`, nil, "want>=2 got=0")

	expectErrIs(t, `f := func(a, b, ...x) { return [a, b, x]; }; f(1);`, nil, ErrWrongNumArguments)
	expectErrHas(t, `f := func(a, b, ...x) { return [a, b, x]; }; f(1);`, nil, "want>=2 got=1")

	expectRun(t, `f := func(x) { return x; }; return f(5);`, nil, Int(5))
	expectRun(t, `f := func(x) { return x * 2; }; return f(5);`, nil, Int(10))
	expectRun(t, `f := func(x, y) { return x + y; }; return f(5, 5);`, nil, Int(10))
	expectRun(t, `f := func(x, y) { return x + y; }; return f(5 + 5, f(5, 5));`,
		nil, Int(20))
	expectRun(t, `return func(x) { return x; }(5)`, nil, Int(5))
	expectRun(t, `x := 10; f := func(x) { return x; }; f(5); return x;`, nil, Int(10))

	expectRun(t, `
	f2 := func(a) {
		f1 := func(a) {
			return a * 2;
		};

		return f1(a) * 3;
	}
	return f2(10)`, nil, Int(60))

	expectRun(t, `
	f1 := func(f) {
		a := [undefined]
		a[0] = func() { return f(a) }
		return a[0]()
	}
	return f1(func(a) { return 2 })
	`, nil, Int(2))

	// closures
	expectRun(t, `
	newAdder := func(x) {
		return func(y) { return x + y }
	}
	add2 := newAdder(2)
	return add2(5)`, nil, Int(7))
	expectRun(t, `
	var out
	m := {a: 1}
	for k,v in m {
		func(){
			out = k
		}()
	}
	return out`, nil, String("a"))

	expectRun(t, `
	var out
	m := {a: 1}
	for k,v in m {
		func(){
			out = v
		}()
	}; return out`, nil, Int(1))
	// function as a argument
	expectRun(t, `
	add := func(a, b) { return a + b };
	sub := func(a, b) { return a - b };
	applyFunc := func(a, b, f) { return f(a, b) };

	return applyFunc(applyFunc(2, 2, add), 3, sub);
	`, nil, Int(1))

	expectRun(t, `f1 := func() { return 5 + 10; }; return f1();`,
		nil, Int(15))
	expectRun(t, `f1 := func() { return 1 }; f2 := func() { return 2 }; return f1() + f2()`,
		nil, Int(3))
	expectRun(t, `f1 := func() { return 1 }; f2 := func() { return f1() + 2 }; f3 := func() { return f2() + 3 }; return f3()`,
		nil, Int(6))
	expectRun(t, `f1 := func() { return 99; 100 }; return f1();`,
		nil, Int(99))
	expectRun(t, `f1 := func() { return 99; return 100 }; return f1();`,
		nil, Int(99))
	expectRun(t, `f1 := func() { return 33; }; f2 := func() { return f1 }; return f2()();`,
		nil, Int(33))
	expectRun(t, `var one; one = func() { one = 1; return one }; return one()`,
		nil, Int(1))
	expectRun(t, `three := func() { one := 1; two := 2; return one + two }; return three()`,
		nil, Int(3))
	expectRun(t, `three := func() { one := 1; two := 2; return one + two }; seven := func() { three := 3; four := 4; return three + four }; return three() + seven()`,
		nil, Int(10))
	expectRun(t, `
	foo1 := func() {
		foo := 50
		return foo
	}
	foo2 := func() {
		foo := 100
		return foo
	}
	return foo1() + foo2()`, nil, Int(150))
	expectRun(t, `
	g := 50;
	minusOne := func() {
		n := 1;
		return g - n;
	};
	minusTwo := func() {
		n := 2;
		return g - n;
	};
	return minusOne() + minusTwo()`, nil, Int(97))
	expectRun(t, `
	f1 := func() {
		f2 := func() { return 1; }
		return f2
	};
	return f1()()`, nil, Int(1))

	expectRun(t, `
	f1 := func(a) { return a; };
	return f1(4)`, nil, Int(4))
	expectRun(t, `
	f1 := func(a, b) { return a + b; };
	return f1(1, 2)`, nil, Int(3))

	expectRun(t, `
	sum := func(a, b) {
		c := a + b;
		return c;
	};
	return sum(1, 2);`, nil, Int(3))

	expectRun(t, `
	sum := func(a, b) {
		c := a + b;
		return c;
	};
	return sum(1, 2) + sum(3, 4);`, nil, Int(10))

	expectRun(t, `
	sum := func(a, b) {
		c := a + b
		return c
	};
	outer := func() {
		return sum(1, 2) + sum(3, 4)
	};
	return outer();`, nil, Int(10))

	expectRun(t, `
	g := 10;

	sum := func(a, b) {
		c := a + b;
		return c + g;
	}

	outer := func() {
		return sum(1, 2) + sum(3, 4) + g;
	}

	return outer() + g
	`, nil, Int(50))

	expectErrIs(t, `func() { return 1; }(1)`, nil, ErrWrongNumArguments)
	expectErrIs(t, `func(a) { return a; }()`, nil, ErrWrongNumArguments)
	expectErrIs(t, `func(a, b) { return a + b; }(1)`, nil, ErrWrongNumArguments)

	expectRun(t, `
	f1 := func(a) {
		return func() { return a; };
	};
	f2 := f1(99);
	return f2()
	`, nil, Int(99))

	expectRun(t, `
	f1 := func(a, b) {
		return func(c) { return a + b + c };
	};
	f2 := f1(1, 2);
	return f2(8);
	`, nil, Int(11))
	expectRun(t, `
	f1 := func(a, b) {
		c := a + b;
		return func(d) { return c + d };
	};
	f2 := f1(1, 2);
	return f2(8);
	`, nil, Int(11))
	expectRun(t, `
	f1 := func(a, b) {
		c := a + b;
		return func(d) {
			e := d + c;
			return func(f) { return e + f };
		}
	};
	f2 := f1(1, 2);
	f3 := f2(3);
	return f3(8);
	`, nil, Int(14))
	expectRun(t, `
	a := 1;
	f1 := func(b) {
		return func(c) {
			return func(d) { return a + b + c + d }
		};
	};
	f2 := f1(2);
	f3 := f2(3);
	return f3(8);
	`, nil, Int(14))
	expectRun(t, `
	f1 := func(a, b) {
		one := func() { return a; };
		two := func() { return b; };
		return func() { return one() + two(); }
	};
	f2 := f1(9, 90);
	return f2();
	`, nil, Int(99))

	// function recursion
	expectRun(t, `
	var fib
	fib = func(x) {
		if x == 0 {
			return 0
		} else if x == 1 {
			return 1
		} else {
			return fib(x-1) + fib(x-2)
		}
	}
	return fib(15)`, nil, Int(610))

	// function recursion
	expectRun(t, `
	return func() {
		var sum
		sum = func(x) {
			return x == 0 ? 0 : x + sum(x-1)
		}
		return sum(5)
	}()`, nil, Int(15))

	// closure and block scopes
	expectRun(t, `
	var out
	func() {
		a := 10
		func() {
			b := 5
			if true {
				out = a + b
			}
		}()
	}(); return out`, nil, Int(15))
	expectRun(t, `
	var out
	func() {
		a := 10
		b := func() { return 5 }
		func() {
			if b() {
				out = a + b()
			}
		}()
	}(); return out`, nil, Int(15))
	expectRun(t, `
	var out
	func() {
		a := 10
		func() {
			b := func() { return 5 }
			func() {
				if true {
					out = a + b()
				}
			}()
		}()
	}(); return out`, nil, Int(15))

	expectRun(t, `return func() {}()`, nil, Undefined)
	expectRun(t, `return func(v) { if v { return true } }(1)`, nil, True)
	expectRun(t, `return func(v) { if v { return true } }(0)`, nil, Undefined)
	expectRun(t, `return func(v) { if v { } else { return true } }(1)`, nil, Undefined)
	expectRun(t, `return func(v) { if v { return } }(1)`, nil, Undefined)
	expectRun(t, `return func(v) { if v { return } }(0)`, nil, Undefined)
	expectRun(t, `return func(v) { if v { } else { return } }(1)`, nil, Undefined)
	expectRun(t, `return func(v) { for ;;v++ { if v == 3 { return true } } }(1)`, nil, True)
	expectRun(t, `return func(v) { for ;;v++ { if v == 3 { break } } }(1)`, nil, Undefined)
	expectRun(t, `
	f := func() { return 2 }
	return (func() {
		f := f()
		return f
	})()`, nil, Int(2))
}

func TestBlocksScope(t *testing.T) {
	expectRun(t, `
	var f
	if true {
		a := 1
		f = func() {
			a = 2
		}
	}
	b := 3
	f()
	return b`, nil, Int(3))

	expectRun(t, `
	var out
	func() {
		f := undefined
		if true {
			a := 10
			f = func() {
				a = 20
			}
		}
		b := 5
		f()
		out = b
	}()
	return out`, nil, Int(5))

	expectRun(t, `
	f := undefined
	if true {
		a := 1
		b := 2
		f = func() {
			a = 3
			b = 4
		}
	}
	c := 5
	d := 6
	f()
	return c + d`, nil, Int(11))

	expectRun(t, `
	fn := undefined
	if true {
		a := 1
		b := 2
		if true {
			c := 3
			d := 4
			fn = func() {
				a = 5
				b = 6
				c = 7
				d = 8
			}
		}
	}
	e := 9
	f := 10
	fn()
	return e + f`, nil, Int(19))

	expectRun(t, `
	out := 0
	func() {
		for x in [1, 2, 3] {
			out += x
		}
	}()
	return out`, nil, Int(6))

	expectRun(t, `
	out := 0
	for x in [1, 2, 3] {
		out += x
	}
	return out`, nil, Int(6))

	expectRun(t, `
	out := 1
	x := func(){
		out := out // get free variable's value with the same name
		return out
	}()
	out = 2
	return x`, nil, Int(1))

	expectRun(t, `
	out := 1
	func(){
		out := out // get free variable's value with the same name
		return func(){
			out = 3 // this refers to out in upper block, not 'out' at top
		}
	}()()
	return out`, nil, Int(1))

	// symbol must be defined before compiling right hand side otherwise not resolved.
	expectErrHas(t, `
	f := func() {
		f()
	}`, newOpts().CompilerError(), `Compile Error: unresolved reference "f"`)
}

func TestVMIf(t *testing.T) {
	expectRun(t, `var out; if (true) { out = 10 }; return out`,
		nil, Int(10))
	expectRun(t, `var out; if (false) { out = 10 }; return out`,
		nil, Undefined)
	expectRun(t, `var out; if (false) { out = 10 } else { out = 20 }; return out`,
		nil, Int(20))
	expectRun(t, `var out; if (1) { out = 10 }; return out`,
		nil, Int(10))
	expectRun(t, `var out; if (0) { out = 10 } else { out = 20 }; return out`,
		nil, Int(20))
	expectRun(t, `var out; if (1 < 2) { out = 10 }; return out`,
		nil, Int(10))
	expectRun(t, `var out; if (1 > 2) { out = 10 }; return out`,
		nil, Undefined)
	expectRun(t, `var out; if (1 < 2) { out = 10 } else { out = 20 }; return out`,
		nil, Int(10))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else { out = 20 }; return out`,
		nil, Int(20))
	expectRun(t, `var out; if (1 < 2) { out = 10 } else if (1 > 2) { out = 20 } else { out = 30 }; return out`,
		nil, Int(10))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 < 2) { out = 20 } else { out = 30 }; return out`,
		nil, Int(20))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 == 2) { out = 20 } else { out = 30 }; return out`,
		nil, Int(30))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 == 2) { out = 20 } else if (1 < 2) { out = 30 } else { out = 40 }; return out`,
		nil, Int(30))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 < 2) { out = 20; out = 21; out = 22 } else { out = 30 }; return out`,
		nil, Int(22))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 == 2) { out = 20 } else { out = 30; out = 31; out = 32}; return out`,
		nil, Int(32))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 < 2) { if (1 == 2) { out = 21 } else { out = 22 } } else { out = 30 }; return out`,
		nil, Int(22))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 < 2) { if (1 == 2) { out = 21 } else if (2 == 3) { out = 22 } else { out = 23 } } else { out = 30 }; return out`,
		nil, Int(23))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 == 2) { if (1 == 2) { out = 21 } else if (2 == 3) { out = 22 } else { out = 23 } } else { out = 30 }; return out`,
		nil, Int(30))
	expectRun(t, `var out; if (1 > 2) { out = 10 } else if (1 == 2) { out = 20 } else { if (1 == 2) { out = 31 } else if (2 == 3) { out = 32 } else { out = 33 } }; return out`,
		nil, Int(33))

	expectRun(t, `var out; if a:=0; a<1 { out = 10 }; return out`, nil, Int(10))
	expectRun(t, `var out; a:=0; if a++; a==1 { out = 10 }; return out`, nil, Int(10))
	expectRun(t, `
	var out
	func() {
		a := 1
		if a++; a > 1 {
			out = a
		}
	}()
	return out`, nil, Int(2))
	expectRun(t, `
	var out
	func() {
		a := 1
		if a++; a == 1 {
			out = 10
		} else {
			out = 20
		}
	}()
	return out`, nil, Int(20))
	expectRun(t, `
	var out
	func() {
		a := 1

		func() {
			if a++; a > 1 {
				a++
			}
		}()

		out = a
	}()
	return out`, nil, Int(3))

	// expression statement in init (should not leave objects on stack)
	expectRun(t, `a := 1; if a; a { return a }`, nil, Int(1))
	expectRun(t, `a := 1; if a + 4; a { return a }`, nil, Int(1))
}

func TestVMIncDec(t *testing.T) {
	expectRun(t, `out := 0; out++; return out`, nil, Int(1))
	expectRun(t, `out := 0; out--; return out`, nil, -Int(1))
	expectRun(t, `a := 0; a++; out := a; return out`, nil, Int(1))
	expectRun(t, `a := 0; a++; a--; out := a; return out`, nil, Int(0))

	// this seems strange but it works because 'a += b' is
	// translated into 'a = a + b' and string type takes other types for + operator.
	expectRun(t, `a := "foo"; a++; return a`, nil, String("foo1"))
	expectErrIs(t, `a := "foo"; a--`, nil, ErrType)
	expectErrHas(t, `a := "foo"; a--`, nil,
		`TypeError: unsupported operand types for '-': 'string' and 'int'`)

	expectErrHas(t, `a++`, newOpts().CompilerError(),
		`Compile Error: unresolved reference "a"`) // not declared
	expectErrHas(t, `a--`, newOpts().CompilerError(),
		`Compile Error: unresolved reference "a"`) // not declared
	expectErrHas(t, `4++`, newOpts().CompilerError(),
		`Compile Error: unresolved reference ""`)
}

func TestVMInteger(t *testing.T) {
	expectRun(t, `return 5`, nil, Int(5))
	expectRun(t, `return 10`, nil, Int(10))
	expectRun(t, `return -5`, nil, Int(-5))
	expectRun(t, `return -10`, nil, Int(-10))
	expectRun(t, `return 5 + 5 + 5 + 5 - 10`, nil, Int(10))
	expectRun(t, `return 2 * 2 * 2 * 2 * 2`, nil, Int(32))
	expectRun(t, `return -50 + 100 + -50`, nil, Int(0))
	expectRun(t, `return 5 * 2 + 10`, nil, Int(20))
	expectRun(t, `return 5 + 2 * 10`, nil, Int(25))
	expectRun(t, `return 20 + 2 * -10`, nil, Int(0))
	expectRun(t, `return 50 / 2 * 2 + 10`, nil, Int(60))
	expectRun(t, `return 2 * (5 + 10)`, nil, Int(30))
	expectRun(t, `return 3 * 3 * 3 + 10`, nil, Int(37))
	expectRun(t, `return 3 * (3 * 3) + 10`, nil, Int(37))
	expectRun(t, `return (5 + 10 * 2 + 15 /3) * 2 + -10`, nil, Int(50))
	expectRun(t, `return 5 % 3`, nil, Int(2))
	expectRun(t, `return 5 % 3 + 4`, nil, Int(6))
	expectRun(t, `return +5`, nil, Int(5))
	expectRun(t, `return +5 + -5`, nil, Int(0))

	expectRun(t, `return 9 + '0'`, nil, Char('9'))
	expectRun(t, `return '9' - 5`, nil, Char('4'))

	expectRun(t, `return 5u`, nil, Uint(5))
	expectRun(t, `return 10u`, nil, Uint(10))
	expectRun(t, `return -5u`, nil, Uint(^uint64(0)-4))
	expectRun(t, `return -10u`, nil, Uint(^uint64(0)-9))
	expectRun(t, `return 5 + 5 + 5 + 5 - 10u`, nil, Uint(10))
	expectRun(t, `return 2 * 2 * 2u * 2 * 2`, nil, Uint(32))
	expectRun(t, `return -50 + 100u + -50`, nil, Uint(0))
	expectRun(t, `return 5u * 2 + 10`, nil, Uint(20))
	expectRun(t, `return 5 + 2u * 10`, nil, Uint(25))
	expectRun(t, `return 20u + 2 * -10`, nil, Uint(0))
	expectRun(t, `return 50 / 2u * 2 + 10`, nil, Uint(60))
	expectRun(t, `return 2 * (5u + 10)`, nil, Uint(30))
	expectRun(t, `return 3 * 3 * 3u + 10`, nil, Uint(37))
	expectRun(t, `return 3u * (3 * 3) + 10`, nil, Uint(37))
	expectRun(t, `return (5 + 10u * 2 + 15 /3) * 2 + -10`, nil, Uint(50))
	expectRun(t, `return 5 % 3u`, nil, Uint(2))
	expectRun(t, `return 5u % 3 + 4`, nil, Uint(6))
	expectRun(t, `return 5 % 3 + 4u`, nil, Uint(6))
	expectRun(t, `return +5u`, nil, Uint(5))
	expectRun(t, `return +5u + -5`, nil, Uint(0))

	expectRun(t, `return 9u + '0'`, nil, Char('9'))
	expectRun(t, `return '9' - 5u`, nil, Char('4'))
}

func TestVMLogical(t *testing.T) {
	expectRun(t, `true && true`, nil, Undefined)
	expectRun(t, `false || true`, nil, Undefined)
	expectRun(t, `return true && true`, nil, True)
	expectRun(t, `return true && false`, nil, False)
	expectRun(t, `return false && true`, nil, False)
	expectRun(t, `return false && false`, nil, False)
	expectRun(t, `return !true && true`, nil, False)
	expectRun(t, `return !true && false`, nil, False)
	expectRun(t, `return !false && true`, nil, True)
	expectRun(t, `return !false && false`, nil, False)

	expectRun(t, `return true || true`, nil, True)
	expectRun(t, `return true || false`, nil, True)
	expectRun(t, `return false || true`, nil, True)
	expectRun(t, `return false || false`, nil, False)
	expectRun(t, `return !true || true`, nil, True)
	expectRun(t, `return !true || false`, nil, False)
	expectRun(t, `return !false || true`, nil, True)
	expectRun(t, `return !false || false`, nil, True)

	expectRun(t, `return 1 && 2`, nil, Int(2))
	expectRun(t, `return 1 || 2`, nil, Int(1))
	expectRun(t, `return 1 && 0`, nil, Int(0))
	expectRun(t, `return 1 || 0`, nil, Int(1))
	expectRun(t, `return 1 && (0 || 2)`, nil, Int(2))
	expectRun(t, `return 0 || (0 || 2)`, nil, Int(2))
	expectRun(t, `return 0 || (0 && 2)`, nil, Int(0))
	expectRun(t, `return 0 || (2 && 0)`, nil, Int(0))

	expectRun(t, `var out; t:=func() {out = 3; return true}; f:=func() {out = 7; return false}; t() && f(); return out`,
		nil, Int(7))
	expectRun(t, `var out; t:=func() {out = 3; return true}; f:=func() {out = 7; return false}; f() && t(); return out`,
		nil, Int(7))
	expectRun(t, `var out; t:=func() {out = 3; return true}; f:=func() {out = 7; return false}; f() || t(); return out`,
		nil, Int(3))
	expectRun(t, `var out; t:=func() {out = 3; return true}; f:=func() {out = 7; return false}; t() || f(); return out`,
		nil, Int(3))
	expectRun(t, `var out; t:=func() {out = 3; return true}; f:=func() {out = 7; return false}; !t() && f(); return out`,
		nil, Int(3))
	expectRun(t, `var out; t:=func() {out = 3; return true}; f:=func() {out = 7; return false}; !f() && t(); return out`,
		nil, Int(3))
	expectRun(t, `var out; t:=func() {out = 3; return true}; f:=func() {out = 7; return false}; !f() || t(); return out`,
		nil, Int(7))
	expectRun(t, `var out; t:=func() {out = 3; return true}; f:=func() {out = 7; return false}; !t() || f(); return out`,
		nil, Int(7))
}

func TestVMMap(t *testing.T) {
	expectRun(t, `
	return {
		one: 10 - 9,
		two: 1 + 1,
		three: 6 / 2,
	}`, nil, Map{
		"one":   Int(1),
		"two":   Int(2),
		"three": Int(3),
	})

	expectRun(t, `
	return {
		"one": 10 - 9,
		"two": 1 + 1,
		"three": 6 / 2,
	}`, nil, Map{
		"one":   Int(1),
		"two":   Int(2),
		"three": Int(3),
	})

	expectRun(t, `return {foo: 5}["foo"]`, nil, Int(5))
	expectRun(t, `return {foo: 5}["bar"]`, nil, Undefined)
	expectRun(t, `key := "foo"; return {foo: 5}[key]`, nil, Int(5))
	expectRun(t, `return {}["foo"]`, nil, Undefined)

	expectRun(t, `
	m := {
		foo: func(x) {
			return x * 2
		},
	}
	return m["foo"](2) + m["foo"](3)
	`, nil, Int(10))

	// map assignment is copy-by-reference
	expectRun(t, `m1 := {k1: 1, k2: "foo"}; m2 := m1; m1.k1 = 5; return m2.k1`,
		nil, Int(5))
	expectRun(t, `m1 := {k1: 1, k2: "foo"}; m2 := m1; m2.k1 = 3; return m1.k1`,
		nil, Int(3))
	expectRun(t, `var out; func() { m1 := {k1: 1, k2: "foo"}; m2 := m1; m1.k1 = 5; out = m2.k1 }(); return out`,
		nil, Int(5))
	expectRun(t, `var out; func() { m1 := {k1: 1, k2: "foo"}; m2 := m1; m2.k1 = 3; out = m1.k1 }(); return out`,
		nil, Int(3))
}

func TestVMSourceModules(t *testing.T) {
	// module return none
	expectRun(t, `out := import("mod1"); return out`,
		newOpts().Module("mod1", `fn := func() { return 5.0 }; a := 2`),
		Undefined)

	// module return values
	expectRun(t, `return import("mod1")`,
		newOpts().Module("mod1", `return 5`), Int(5))
	expectRun(t, `return import("mod1")`,
		newOpts().Module("mod1", `return "foo"`), String("foo"))

	// module return compound types
	expectRun(t, `out := import("mod1"); return out`,
		newOpts().Module("mod1", `return [1, 2, 3]`), Array{Int(1), Int(2), Int(3)})
	expectRun(t, `out := import("mod1"); return out`,
		newOpts().Module("mod1", `return {a: 1, b: 2}`), Map{"a": Int(1), "b": Int(2)})

	// if returned values are not imumutable, they can be updated
	expectRun(t, `m1 := import("mod1"); m1.a = 5; return m1`,
		newOpts().Module("mod1", `return {a: 1, b: 2}`), Map{"a": Int(5), "b": Int(2)})
	expectRun(t, `m1 := import("mod1"); m1[1] = 5; return m1`,
		newOpts().Module("mod1", `return [1, 2, 3]`), Array{Int(1), Int(5), Int(3)})
	// modules are evaluated once, calling in different scopes returns same object
	expectRun(t, `
	m1 := import("mod1")
	m1.a = 5
	func(){
		m11 := import("mod1")
		m11.a = 6
	}()
	return m1`, newOpts().Module("mod1", `return {a: 1, b: 2}`), Map{"a": Int(6), "b": Int(2)})

	// module returning function
	expectRun(t, `out := import("mod1")(); return out`,
		newOpts().Module("mod1", `return func() { return 5.0 }`), Float(5.0))
	// returned function that reads module variable
	expectRun(t, `out := import("mod1")(); return out`,
		newOpts().Module("mod1", `a := 1.5; return func() { return a + 5.0 }`), Float(6.5))
	// returned function that reads local variable
	expectRun(t, `out := import("mod1")(); return out`,
		newOpts().Module("mod1", `return func() { a := 1.5; return a + 5.0 }`), Float(6.5))
	// returned function that reads free variables
	expectRun(t, `out := import("mod1")(); return out`,
		newOpts().Module("mod1", `return func() { a := 1.5; return func() { return a + 5.0 }() }`), Float(6.5))

	// recursive function in module
	expectRun(t, `return import("mod1")`,
		newOpts().Module("mod1", `
	var a
	a = func(x) {
		return x == 0 ? 0 : x + a(x-1)
	}
	return a(5)`), Int(15))

	expectRun(t, `out := import("mod1"); return out`,
		newOpts().Module("mod1", `
	return func() {
		var a
		a = func(x) {
			return x == 0 ? 0 : x + a(x-1)
		}
		return a(5)
	}()
	`), Int(15))

	// (main) -> mod1 -> mod2
	expectRun(t, `return import("mod1")()`,
		newOpts().Module("mod1", `return import("mod2")`).
			Module("mod2", `return func() { return 5.0 }`),
		Float(5.0))
	// (main) -> mod1 -> mod2
	//        -> mod2
	expectRun(t, `import("mod1"); return import("mod2")()`,
		newOpts().Module("mod1", `return import("mod2")`).
			Module("mod2", `return func() { return 5.0 }`),
		Float(5.0))
	// (main) -> mod1 -> mod2 -> mod3
	//        -> mod2 -> mod3
	expectRun(t, `import("mod1"); return import("mod2")()`,
		newOpts().Module("mod1", `return import("mod2")`).
			Module("mod2", `return import("mod3")`).
			Module("mod3", `return func() { return 5.0 }`),
		Float(5.0))

	// cyclic imports
	// (main) -> mod1 -> mod2 -> mod1
	expectErrHas(t, `import("mod1")`,
		newOpts().Module("mod1", `import("mod2")`).
			Module("mod2", `import("mod1")`).CompilerError(),
		"Compile Error: cyclic module import: mod1\n\tat mod2:1:1")
	// (main) -> mod1 -> mod2 -> mod3 -> mod1
	expectErrHas(t, `import("mod1")`,
		newOpts().Module("mod1", `import("mod2")`).
			Module("mod2", `import("mod3")`).
			Module("mod3", `import("mod1")`).CompilerError(),
		"Compile Error: cyclic module import: mod1\n\tat mod3:1:1")
	// (main) -> mod1 -> mod2 -> mod3 -> mod2
	expectErrHas(t, `import("mod1")`,
		newOpts().Module("mod1", `import("mod2")`).
			Module("mod2", `import("mod3")`).
			Module("mod3", `import("mod2")`).CompilerError(),
		"Compile Error: cyclic module import: mod2\n\tat mod3:1:1")

	// unknown modules
	expectErrHas(t, `import("mod0")`,
		newOpts().Module("mod1", `a := 5`).CompilerError(), "Compile Error: module 'mod0' not found")
	expectErrHas(t, `import("mod1")`,
		newOpts().Module("mod1", `import("mod2")`).CompilerError(), "Compile Error: module 'mod2' not found")

	expectRun(t, `m1 := import("mod1"); m1.a.b = 5; return m1.a.b`,
		newOpts().Module("mod1", `return {a: {b: 3}}`), Int(5))

	// make sure module has same builtin functions
	expectRun(t, `out := import("mod1"); return out`,
		newOpts().Module("mod1", `return func() { return typeName(0) }()`), String("int"))

	// module cannot access outer scope
	expectErrHas(t, `a := 5; import("mod1")`, newOpts().Module("mod1", `return a`).CompilerError(),
		"Compile Error: unresolved reference \"a\"\n\tat mod1:1:8")

	// runtime error within modules
	expectErrIs(t, `
	a := 1;
	b := import("mod1");
	b(a)`,
		newOpts().Module("mod1", `
	return func(a) {
	   a()
	}
	`), ErrNotCallable)

	// module with no return
	expectRun(t, `out := import("mod0"); return out`,
		newOpts().Module("mod0", ``), Undefined)
	expectRun(t, `out := import("mod0"); return out`,
		newOpts().Module("mod0", `if 0 { return true }`), Undefined)
	expectRun(t, `out := import("mod0"); return out`,
		newOpts().Module("mod0", `if 1 { } else { }`), Undefined)
	expectRun(t, `out := import("mod0"); return out`,
		newOpts().Module("mod0", `for v:=0;;v++ { if v == 3 { break } }`), Undefined)

	// importing same module multiple times returns same object
	expectRun(t, `
	m1 := import("mod")
	m2 := import("mod")
	return m1 == m2
	`, newOpts().Module("mod", `return { x: 1 }`), True)
	expectRun(t, `
	m1 := import("mod")
	m2 := import("mod")
	m1.x = 2
	f := func() {
		return import("mod")
	}
	return [m1 == m2, m2 == import("mod"), m1 == f()]
	`, newOpts().Module("mod", `return { x: 1 }`), Array{True, True, True})
	expectRun(t, `
	mod2 := import("mod2")
	mod1 := import("mod1")
	return mod1.mod2 == mod2
	`, newOpts().Module("mod1", `m2 := import("mod2"); m2.x = 2; return { x: 1, mod2: m2 }`).
		Module("mod2", "m := { x: 0 }; return m"), True)

}

func TestVMUnary(t *testing.T) {
	expectRun(t, `!true`, nil, Undefined)
	expectRun(t, `true`, nil, Undefined)
	expectRun(t, `!false`, nil, Undefined)
	expectRun(t, `false`, nil, Undefined)
	expectRun(t, `return !false`, nil, True)
	expectRun(t, `return !0`, nil, True)
	expectRun(t, `return !5`, nil, False)
	expectRun(t, `return !!true`, nil, True)
	expectRun(t, `return !!false`, nil, False)
	expectRun(t, `return !!5`, nil, True)

	expectRun(t, `-1`, nil, Undefined)
	expectRun(t, `+1`, nil, Undefined)
	expectRun(t, `return -1`, nil, Int(-1))
	expectRun(t, `return +1`, nil, Int(1))
	expectRun(t, `return -0`, nil, Int(0))
	expectRun(t, `return +0`, nil, Int(0))
	expectRun(t, `return ^1`, nil, Int(^int64(1)))
	expectRun(t, `return ^0`, nil, Int(^int64(0)))

	expectRun(t, `-1u`, nil, Undefined)
	expectRun(t, `+1u`, nil, Undefined)
	expectRun(t, `return -1u`, nil, Uint(^uint64(0)))
	expectRun(t, `return +1u`, nil, Uint(1))
	expectRun(t, `return -0u`, nil, Uint(0))
	expectRun(t, `return +0u`, nil, Uint(0))
	expectRun(t, `return ^1u`, nil, Uint(^uint64(1)))
	expectRun(t, `return ^0u`, nil, Uint(^uint64(0)))

	expectRun(t, `-true`, nil, Undefined)
	expectRun(t, `+false`, nil, Undefined)
	expectRun(t, `return -true`, nil, Int(-1))
	expectRun(t, `return +true`, nil, Int(1))
	expectRun(t, `return -false`, nil, Int(0))
	expectRun(t, `return +false`, nil, Int(0))
	expectRun(t, `return ^true`, nil, Int(^int64(1)))
	expectRun(t, `return ^false`, nil, Int(^int64(0)))

	expectRun(t, `-'a'`, nil, Undefined)
	expectRun(t, `+'a'`, nil, Undefined)
	expectRun(t, `return -'a'`, nil, Int(-rune('a')))
	expectRun(t, `return +'a'`, nil, Char('a'))
	expectRun(t, `return ^'a'`, nil, Int(^rune('a')))

	expectRun(t, `-1.0`, nil, Undefined)
	expectRun(t, `+1.0`, nil, Undefined)
	expectRun(t, `return -1.0`, nil, Float(-1.0))
	expectRun(t, `return +1.0`, nil, Float(1.0))
	expectRun(t, `return -0.0`, nil, Float(0.0))
	expectRun(t, `return +0.0`, nil, Float(0.0))

	expectErrIs(t, `return ^1.0`, nil, ErrType)
	expectErrHas(t, `return ^1.0`, nil, `TypeError: invalid type for unary '^': 'float'`)
}

func TestVMScopes(t *testing.T) {
	// shadowed local variable
	expectRun(t, `
	c := 5
	if a := 3; a {
		c := 6
	} else {
		c := 7
	}
	return c
	`, nil, Int(5))

	// shadowed function local variable
	expectRun(t, `
	return func() {
		c := 5
		if a := 3; a {
			c := 6
		} else {
			c := 7
		}
		return c
	}()
	`, nil, Int(5))

	// 'b' is declared in 2 separate blocks
	expectRun(t, `
	c := 5
	if a := 3; a {
		b := 8
		c = b
	} else {
		b := 9
		c = b
	}
	return c
	`, nil, Int(8))

	// shadowing inside for statement
	expectRun(t, `
	a := 4
	b := 5
	for i:=0;i<3;i++ {
		b := 6
		for j:=0;j<2;j++ {
			b := 7
			a = i*j
		}
	}
	return a`, nil, Int(2))

	// shadowing variable declared in init statement
	expectRun(t, `
	var out
	if a := 5; a {
		a := 6
		out = a
	}
	return out`, nil, Int(6))
	expectRun(t, `
	var out
	a := 4
	if a := 5; a {
		a := 6
		out = a
	}
	return out`, nil, Int(6))
	expectRun(t, `
	var out
	a := 4
	if a := 0; a {
		a := 6
		out = a
	} else {
		a := 7
		out = a
	}
	return out`, nil, Int(7))
	expectRun(t, `
	var out
	a := 4
	if a := 0; a {
		out = a
	} else {
		out = a
	}
	return out`, nil, Int(0))

	// shadowing function level
	expectRun(t, `
	a := 5
	func() {
		a := 6
		a = 7
	}()
	return a`, nil, Int(5))
	expectRun(t, `
	a := 5
	func() {
		if a := 7; true {
			a = 8
		}
	}()
	return a`, nil, Int(5))
	expectRun(t, `
	a := 5
	func() {
		if a := 7; true {
			a = 8
		}
	}()
	var (b, c, d)
	return [a, b, c, d]`, nil, Array{Int(5), Undefined, Undefined, Undefined})
	expectRun(t, `
	var f
	a := 5
	func() {
		if a := 7; true {
			f = func() {
				a = 8
			}
		}
	}()
	f()
	return a`, nil, Int(5))
	expectRun(t, `
	if a := 7; false {
		a = 8
		return a
	} else {
		a = 9
		return a
	}`, nil, Int(9))
	expectRun(t, `
	if a := 7; false {
		a = 8
		return a
	} else if false {
		a = 9
		return a
	} else {
		a = 10
		return a	
	}`, nil, Int(10))
}

func TestVMSelector(t *testing.T) {
	expectRun(t, `a := {k1: 5, k2: "foo"}; return a.k1`, nil, Int(5))
	expectRun(t, `a := {k1: 5, k2: "foo"}; return a.k2`, nil, String("foo"))
	expectRun(t, `a := {k1: 5, k2: "foo"}; return a.k3`, nil, Undefined)

	expectRun(t, `
	a := {
		b: {
			c: 4,
			a: false,
		},
		c: "foo bar",
	}
	_ := a.b.c
	return a.b.c`, nil, Int(4))

	expectRun(t, `
	a := {
		b: {
			c: 4,
			a: false,
		},
		c: "foo bar",
	}
	_ := a.x.c
	return a.x.c`, nil, Undefined)

	expectRun(t, `
	a := {
		b: {
			c: 4,
			a: false,
		},
		c: "foo bar",
	}
	_ := a.x.y
	return a.x.y`, nil, Undefined)

	expectRun(t, `a := {b: 1, c: "foo"}; a.b = 2; return a.b`, nil, Int(2))
	expectRun(t, `a := {b: 1, c: "foo"}; a.c = 2; return a.c`, nil, Int(2))
	expectRun(t, `a := {b: {c: 1}}; a.b.c = 2; return a.b.c`, nil, Int(2))
	expectRun(t, `a := {b: 1}; a.c = 2; return a`, nil, Map{"b": Int(1), "c": Int(2)})
	expectRun(t, `a := {b: {c: 1}}; a.b.d = 2; return a`, nil,
		Map{"b": Map{"c": Int(1), "d": Int(2)}})

	expectRun(t, `return func() { a := {b: 1, c: "foo"}; a.b = 2; return a.b }()`, nil, Int(2))
	expectRun(t, `return func() { a := {b: 1, c: "foo"}; a.c = 2; return a.c }()`, nil, Int(2))
	expectRun(t, `return func() { a := {b: {c: 1}}; a.b.c = 2; return a.b.c }()`, nil, Int(2))
	expectRun(t, `return func() { a := {b: 1}; a.c = 2; return a }()`, nil,
		Map{"b": Int(1), "c": Int(2)})
	expectRun(t, `return func() { a := {b: {c: 1}}; a.b.d = 2; return a }()`, nil,
		Map{"b": Map{"c": Int(1), "d": Int(2)}})

	expectRun(t, `return func() { a := {b: 1, c: "foo"}; func() { a.b = 2 }(); return a.b }()`, nil, Int(2))
	expectRun(t, `return func() { a := {b: 1, c: "foo"}; func() { a.c = 2 }(); return a.c }()`, nil, Int(2))
	expectRun(t, `return func() { a := {b: {c: 1}}; func() { a.b.c = 2 }(); return a.b.c }()`, nil, Int(2))
	expectRun(t, `return func() { a := {b: 1}; func() { a.c = 2 }(); return a }()`, nil,
		Map{"b": Int(1), "c": Int(2)})
	expectRun(t, `return func() { a := {b: {c: 1}}; func() { a.b.d = 2 }(); return a }()`,
		nil, Map{"b": Map{"c": Int(1), "d": Int(2)}})

	expectRun(t, `
	a := {
		b: [1, 2, 3],
		c: {
			d: 8,
			e: "foo",
			f: [9, 8],
		},
	}
	return [a.b[2], a.c.d, a.c.e, a.c.f[1]]
	`, nil, Array{Int(3), Int(8), String("foo"), Int(8)})

	expectRun(t, `
	var out
	func() {
		a := [1, 2, 3]
		b := 9
		a[1] = b
		b = 7     // make sure a[1] has a COPY of value of 'b'
		out = a[1]
	}()
	return out
	`, nil, Int(9))

	expectErrIs(t, `a := {b: {c: 1}}; a.d.c = 2`, nil, ErrNotIndexAssignable)
	expectErrIs(t, `a := [1, 2, 3]; a.b = 2`, nil, ErrType)
	expectErrIs(t, `a := "foo"; a.b = 2`, nil, ErrNotIndexAssignable)
	expectErrIs(t, `func() { a := {b: {c: 1}}; a.d.c = 2 }()`, nil, ErrNotIndexAssignable)
	expectErrIs(t, `func() { a := [1, 2, 3]; a.b = 2 }()`, nil, ErrType)
	expectErrIs(t, `func() { a := "foo"; a.b = 2 }()`, nil, ErrNotIndexAssignable)
}

func TestVMStackOverflow(t *testing.T) {
	expectErrIs(t, `var f; f = func() { return f() + 1 }; f()`, nil, ErrStackOverflow)
}

func TestVMString(t *testing.T) {
	expectRun(t, `return "Hello World!"`, nil, String("Hello World!"))
	expectRun(t, `return "Hello" + " " + "World!"`, nil, String("Hello World!"))

	expectRun(t, `return "Hello" == "Hello"`, nil, True)
	expectRun(t, `return "Hello" == "World"`, nil, False)
	expectRun(t, `return "Hello" != "Hello"`, nil, False)
	expectRun(t, `return "Hello" != "World"`, nil, True)

	expectRun(t, `return "Hello" > "World"`, nil, False)
	expectRun(t, `return "World" < "Hello"`, nil, False)
	expectRun(t, `return "Hello" < "World"`, nil, True)
	expectRun(t, `return "World" > "Hello"`, nil, True)
	expectRun(t, `return "Hello" >= "World"`, nil, False)
	expectRun(t, `return "Hello" <= "World"`, nil, True)
	expectRun(t, `return "Hello" >= "Hello"`, nil, True)
	expectRun(t, `return "World" <= "World"`, nil, True)

	// index operator
	str := "abcdef"
	strStr := `"abcdef"`
	strLen := 6
	for idx := 0; idx < strLen; idx++ {
		expectRun(t, fmt.Sprintf("return %s[%d]", strStr, idx), nil, Int(str[idx]))
		expectRun(t, fmt.Sprintf("return %s[0 + %d]", strStr, idx), nil, Int(str[idx]))
		expectRun(t, fmt.Sprintf("return %s[1 + %d - 1]", strStr, idx), nil, Int(str[idx]))
		expectRun(t, fmt.Sprintf("idx := %d; return %s[idx]", idx, strStr), nil, Int(str[idx]))
	}

	expectErrIs(t, fmt.Sprintf("%s[%d]", strStr, -1), nil, ErrIndexOutOfBounds)
	expectErrIs(t, fmt.Sprintf("%s[%d]", strStr, strLen), nil, ErrIndexOutOfBounds)

	// slice operator
	for low := 0; low < strLen; low++ {
		expectRun(t, fmt.Sprintf("return %s[%d:%d]", strStr, low, low), nil, String(""))
		for high := low; high <= strLen; high++ {
			expectRun(t, fmt.Sprintf("return %s[%d:%d]", strStr, low, high),
				nil, String(str[low:high]))
			expectRun(t,
				fmt.Sprintf("return %s[0 + %d : 0 + %d]", strStr, low, high),
				nil, String(str[low:high]))
			expectRun(t,
				fmt.Sprintf("return %s[1 + %d - 1 : 1 + %d - 1]",
					strStr, low, high),
				nil, String(str[low:high]))
			expectRun(t,
				fmt.Sprintf("return %s[:%d]", strStr, high),
				nil, String(str[:high]))
			expectRun(t,
				fmt.Sprintf("return %s[%d:]", strStr, low),
				nil, String(str[low:]))
		}
	}

	expectRun(t, fmt.Sprintf("return %s[:]", strStr), nil, String(str[:]))
	expectRun(t, fmt.Sprintf("return %s[:]", strStr), nil, String(str))
	expectRun(t, fmt.Sprintf("return %s[%d:]", strStr, 0), nil, String(str))
	expectRun(t, fmt.Sprintf("return %s[:%d]", strStr, strLen), nil, String(str))
	expectRun(t, fmt.Sprintf("return %s[%d:%d]", strStr, 2, 2), nil, String(""))

	expectErrIs(t, fmt.Sprintf("%s[:%d]", strStr, -1), nil, ErrInvalidIndex)
	expectErrIs(t, fmt.Sprintf("%s[%d:]", strStr, strLen+1), nil, ErrInvalidIndex)
	expectErrIs(t, fmt.Sprintf("%s[%d:%d]", strStr, 0, -1), nil, ErrInvalidIndex)
	expectErrIs(t, fmt.Sprintf("%s[%d:%d]", strStr, 2, 1), nil, ErrInvalidIndex)

	// string concatenation with other types
	expectRun(t, `return "foo" + 1`, nil, String("foo1"))
	// Float.String() returns the smallest number of digits
	// necessary such that ParseFloat will return f exactly.
	expectErrIs(t, `return 1 + "foo"`, nil, ErrType)
	expectRun(t, `return "foo" + 1.0`, nil, String("foo1")) // <- note '1' instead of '1.0'
	expectErrIs(t, `return 1.0 + "foo"`, nil, ErrType)
	expectRun(t, `return "foo" + 1.5`, nil, String("foo1.5"))
	expectErrIs(t, `return 1.5 + "foo"`, nil, ErrType)
	expectRun(t, `return "foo" + true`, nil, String("footrue"))
	expectErrIs(t, `return true + "foo"`, nil, ErrType)
	expectRun(t, `return "foo" + 'X'`, nil, String("fooX"))
	expectRun(t, `return 'X' + "foo"`, nil, String("Xfoo"))
	expectRun(t, `return "foo" + error(5)`, nil, String("fooerror: 5"))
	expectRun(t, `return "foo" + undefined`, nil, String("fooundefined"))
	expectErrIs(t, `return undefined + "foo"`, nil, ErrType)
	// array adds rhs object to the array
	expectRun(t, `return [1, 2, 3] + "foo"`,
		nil, Array{Int(1), Int(2), Int(3), String("foo")})
	// also works with "+=" operator
	expectRun(t, `out := "foo"; out += 1.5; return out`, nil, String("foo1.5"))
	expectErrHas(t, `"foo" - "bar"`,
		nil, `TypeError: unsupported operand types for '-': 'string' and 'string'`)
}

func TestVMTailCall(t *testing.T) {
	expectRun(t, `
	var fac
	fac = func(n, a) {
		if n == 1 {
			return a
		}
		return fac(n-1, n*a)
	}
	return fac(5, 1)`, nil, Int(120))

	expectRun(t, `
	var fac
	fac = func(n, a) {
		if n == 1 {
			return a
		}
		x := {foo: fac} // indirection for test
		return x.foo(n-1, n*a)
	}
	return fac(5, 1)`, nil, Int(120))

	expectRun(t, `
	var fib
	fib = func(x, s) {
		if x == 0 {
			return 0 + s
		} else if x == 1 {
			return 1 + s
		}
		return fib(x-1, fib(x-2, s))
	}
	return fib(15, 0)`, nil, Int(610))

	expectRun(t, `
	var fib
	fib = func(n, a, b) {
		if n == 0 {
			return a
		} else if n == 1 {
			return b
		}
		return fib(n-1, b, a + b)
	}
	return fib(15, 0, 1)`, nil, Int(610))

	expectRun(t, `
	var (foo, out = 0)
	foo = func(a) {
		if a == 0 {
			return
		}
		out += a
		foo(a-1)
	}
	foo(10)
	return out`, nil, Int(55))

	expectRun(t, `
	var f1
	f1 = func() {
		var f2
		f2 = func(n, s) {
			if n == 0 { return s }
			return f2(n-1, n + s)
		}
		return f2(5, 0)
	}
	return f1()`, nil, Int(15))

	// tail-call replacing loop
	// without tail-call optimization, this code will cause stack overflow
	expectRun(t, `
	var iter
	iter = func(n, max) {
		if n == max {
			return n
		}
		return iter(n+1, max)
	}
	return iter(0, 9999)`, nil, Int(9999))

	expectRun(t, `
	var (iter, c = 0)
	iter = func(n, max) {
		if n == max {
			return
		}
		c++
		iter(n+1, max)
	}
	iter(0, 9999)
	return c`, nil, Int(9999))
}

func TestVMTailCallFreeVars(t *testing.T) {
	expectRun(t, `
	var out
	func() {
		a := 10
		f2 := 0
		f2 = func(n, s) {
			if n == 0 {
				return s + a
			}
			return f2(n-1, n+s)
		}
		out = f2(5, 0)
	}()
	return out`, nil, Int(25))
}

func TestVMCall(t *testing.T) {
	var invErr *RuntimeError
	expectErrAs(t, `f := func() {}; return f(...{})`, nil, &invErr, nil)
	require.NotNil(t, invErr)
	require.NotNil(t, invErr.Err)
	require.Equal(t, ErrType, invErr.Err.Cause)
	require.Equal(t, "invalid type for argument 'last': expected array, found map", invErr.Err.Message)

	invErr = nil
	expectErrAs(t, `f := func() {}; return f(..."")`, nil, &invErr, nil)
	require.NotNil(t, invErr)
	require.NotNil(t, invErr.Err)
	require.Equal(t, ErrType, invErr.Err.Cause)
	require.Equal(t, "invalid type for argument 'last': expected array, found string", invErr.Err.Message)

	invErr = nil
	expectErrAs(t, `f := func() {}; return f(...undefined)`, nil, &invErr, nil)
	require.NotNil(t, invErr)
	require.NotNil(t, invErr.Err)
	require.Equal(t, ErrType, invErr.Err.Cause)
	require.Equal(t, "invalid type for argument 'last': expected array, found undefined", invErr.Err.Message)

	expectRun(t, `f := func() {}; return f()`, nil, Undefined)
	expectRun(t, `f := func(a) { return a; }; return f(1)`, nil, Int(1))
	expectRun(t, `f := func(a, b) { return [a, b]; }; return f(1, 2)`, nil, Array{Int(1), Int(2)})
	expectErrIs(t, `f := func() { return; }; return f(1)`, nil, ErrWrongNumArguments)
	expectErrHas(t, `f := func() { return; }; return f(1)`, nil, `want=0 got=1`)

	expectRun(t, `f := func(...a) { return a; }; return f()`, nil, Array{})
	expectRun(t, `f := func(...a) { return a; }; return f(1)`, nil, Array{Int(1)})
	expectRun(t, `f := func(...a) { return a; }; return f(1, 2)`, nil, Array{Int(1), Int(2)})
	expectErrIs(t, `f := func(a, ...b) { return a; }; return f()`, nil, ErrWrongNumArguments)
	expectErrHas(t, `f := func(a, ...b) { return a; }; return f()`, nil, `want>=1 got=0`)
	expectErrHas(t, `f := func(a, b, ...c) { return a; }; return f(1)`, nil, `want>=2 got=1`)

	expectRun(t, `f := func(a, ...b) { return a; }; return f(1, 2)`, nil, Int(1))
	expectRun(t, `f := func(a, ...b) { return b; }; return f(1)`, nil, Array{})
	expectRun(t, `f := func(a, ...b) { return b; }; return f(1, 2)`, nil, Array{Int(2)})
	expectRun(t, `f := func(a, ...b) { return b; }; return f(1, 2, 3)`, nil, Array{Int(2), Int(3)})

	expectRun(t, `f := func(a, b, ...c) { return a; }; return f(1, 2)`, nil, Int(1))
	expectRun(t, `f := func(a, b, ...c) { return b; }; return f(1, 2)`, nil, Int(2))
	expectRun(t, `f := func(a, b, ...c) { return c; }; return f(1, 2)`, nil, Array{})
	expectRun(t, `f := func(a, b, ...c) { return c; }; return f(1, 2, 3)`, nil, Array{Int(3)})
	expectRun(t, `f := func(a, b, ...c) { return c; }; return f(1, 2, 3, 4)`, nil, Array{Int(3), Int(4)})

	expectRun(t, `f := func(a) { return a; }; return f(...[1])`, nil, Int(1))
	expectRun(t, `f := func(a, b) { return [a, b]; }; return f(...[1, 2])`, nil, Array{Int(1), Int(2)})
	expectRun(t, `f := func(a, b) { return [a, b]; }; return f(1, ...[2])`, nil, Array{Int(1), Int(2)})
	expectRun(t, `f := func() { return; }; return f(...[])`, nil, Undefined)

	expectRun(t, `f := func(a, ...b) { return a; }; return f(1, ...[2])`, nil, Int(1))
	expectRun(t, `f := func(a, ...b) { return b; }; return f(1, ...[2])`, nil, Array{Int(2)})
	expectRun(t, `f := func(a, ...b) { return b; }; return f(1, ...[2, 3])`, nil, Array{Int(2), Int(3)})
	expectRun(t, `f := func(a, ...b) { return a; }; return f(...[1, 2, 3])`, nil, Int(1))
	expectRun(t, `f := func(a, ...b) { return b; }; return f(...[1, 2, 3])`, nil, Array{Int(2), Int(3)})

	expectRun(t, `f := func(...a) { return a; }; return f(1, 2, ...[3, 4])`, nil, Array{Int(1), Int(2), Int(3), Int(4)})
	expectRun(t, `f := func(a, ...b) { return a; }; return f(1, 2, ...[3, 4])`, nil, Int(1))
	expectRun(t, `f := func(a, ...b) { return b; }; return f(1, 2, ...[3, 4])`, nil, Array{Int(2), Int(3), Int(4)})
	expectRun(t, `f := func(a, ...b) { return b; }; return f(1, 2, ...[])`, nil, Array{Int(2)})
	// if args and params match, 'c' points to the given array not undefined.
	expectRun(t, `f := func(a, b, ...c) { return c; }; return f(1, 2, ...[])`, nil, Array{})

	expectErrIs(t, `f := func(a, ...b) { return a; }; return f(...[])`, nil, ErrWrongNumArguments)
	expectErrHas(t, `f := func(a, ...b) { return a; }; return f(...[])`, nil, `want>=1 got=0`)
	expectErrHas(t, `f := func(a, b, ...c) { return a; }; return f(...[1])`, nil, `want>=2 got=1`)
	expectErrHas(t, `f := func(a, b, ...c) { return a; }; return f(1, ...[])`, nil, `want>=2 got=1`)
	expectErrHas(t, `f := func(a, b, c, ...d) { return a; }; return f(1, ...[])`, nil, `want>=3 got=1`)
	expectErrIs(t, `f := func(a, b, c, ...d) { return a; }; return f(1, ...[2])`, nil, ErrWrongNumArguments)
	expectErrHas(t, `f := func(a, b, c, ...d) { return a; }; return f(1, ...[2])`, nil, `want>=3 got=2`)

	expectErrIs(t, `f := func(a, b) { return a; }; return f(1, ...[2, 3])`, nil, ErrWrongNumArguments)
	expectErrHas(t, `f := func(a, b) { return a; }; return f(1, 2, ...[3])`, nil, `want=2 got=3`)
	expectErrHas(t, `f := func(a, b) { return a; }; return f(1, ...[2, 3])`, nil, `want=2 got=3`)
	expectErrHas(t, `f := func(a, b) { return a; }; return f(...[1, 2, 3])`, nil, `want=2 got=3`)

	expectRun(t, `f := func(a, ...b) { var x; return [x, a]; }; return f(1, 2)`, nil, Array{Undefined, Int(1)})
	expectRun(t, `f := func(a, ...b) { var x; return [x, b]; }; return f(1, 2)`, nil, Array{Undefined, Array{Int(2)}})

	expectRun(t, `global f; return f()`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Int(len(args)), nil
		}}}), Int(0))
	expectRun(t, `global f; return f(1)`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Array{Int(len(args)), args[0]}, nil
		}}}), Array{Int(1), Int(1)})
	expectRun(t, `global f; return f(1, 2)`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Array{Int(len(args)), args[0], args[1]}, nil
		}}}), Array{Int(2), Int(1), Int(2)})
	expectRun(t, `global f; return f(...[])`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Array{Int(len(args)), Array(args)}, nil
		}}}), Array{Int(0), Array{}})
	expectRun(t, `global f; return f(...[1])`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Array{Int(len(args)), Array(args)}, nil
		}}}), Array{Int(1), Array{Int(1)}})
	expectRun(t, `global f; return f(1, ...[])`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Array{Int(len(args)), Array(args)}, nil
		}}}), Array{Int(1), Array{Int(1)}})
	expectRun(t, `global f; return f(1, ...[2])`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Array{Int(len(args)), Array(args)}, nil
		}}}), Array{Int(2), Array{Int(1), Int(2)}})
	expectRun(t, `global f; return f(1, 2, ...[3])`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Array{Int(len(args)), Array(args)}, nil
		}}}), Array{Int(3), Array{Int(1), Int(2), Int(3)}})
	expectRun(t, `global f; return f(1, 2, 3)`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Array{Int(len(args)), Array(args)}, nil
		}}}), Array{Int(3), Array{Int(1), Int(2), Int(3)}})

	invErr = nil
	expectErrAs(t, `global f; var a = {}; return f(1, ...a)`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return Undefined, nil
		}}}), &invErr, nil)
	require.NotNil(t, invErr)
	require.NotNil(t, invErr.Err)
	require.Equal(t, ErrType, invErr.Err.Cause)
	require.Equal(t, "invalid type for argument 'last': expected array, found map",
		invErr.Err.Message)

	expectErrIs(t, `global f; return f()`, newOpts().Globals(
		Map{"f": &Function{Value: func(args ...Object) (Object, error) {
			return nil, ErrWrongNumArguments
		}}}), ErrWrongNumArguments)
	expectErrIs(t, `global f; return f()`, newOpts().Globals(Map{"f": Undefined}),
		ErrNotCallable)

	expectRun(t, `a := { b: func(x) { return x + 2 } }; return a.b(5)`, nil, Int(7))
	expectRun(t, `a := { b: { c: func(x) { return x + 2 } } }; return a.b.c(5)`,
		nil, Int(7))
	expectRun(t, `a := { b: { c: func(x) { return x + 2 } } }; return a["b"].c(5)`,
		nil, Int(7))
	expectErrIs(t, `
	a := 1
	b := func(a, c) {
	c(a)
	}
	c := func(a) {
	a()
	}
	b(a, c)
	`, nil, ErrNotCallable)

	expectRun(t, `return {a: string(...[0])}`, nil, Map{"a": String("0")})
	expectRun(t, `return {a: string([0])}`, nil, Map{"a": String("[0]")})
	expectRun(t, `return {a: bytes(...repeat([0], 4096))}`,
		nil, Map{"a": make(Bytes, 4096)})
}

func TestVMCallCompiledFunction(t *testing.T) {
	script := `
	var v = 0
	return {
		"add": func(x) {
			v+=x
			return v
		},
		"sub": func(x) {
			v-=x
			return v
		},
	}
	`
	c, err := Compile([]byte(script), CompilerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	vm := NewVM(c)
	f, err := vm.Run(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	//locals := vm.GetLocals(nil)
	// t.Log(f)
	require.Contains(t, f.(Map), "add")
	require.Contains(t, f.(Map), "sub")
	add := f.(Map)["add"].(*CompiledFunction)
	ret, err := vm.RunCompiledFunction(add, nil, Int(10))
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(ret)
	require.Equal(t, Int(10), ret.(Int))

	ret, err = vm.RunCompiledFunction(add, nil, Int(10))
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(ret)
	require.Equal(t, Int(20), ret.(Int))

	sub := f.(Map)["sub"].(*CompiledFunction)
	ret, err = vm.RunCompiledFunction(sub, nil, Int(1))
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(ret)
	require.Equal(t, Int(19), ret.(Int))

	ret, err = vm.RunCompiledFunction(sub, nil, Int(1))
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(ret)
	require.Equal(t, Int(18), ret.(Int))
	// for i := range locals {
	// 	fmt.Printf("%#v\n", locals[i])
	// 	fmt.Printf("%#v\n", *locals[i].(*ObjectPtr).Value)
	// }
}

func TestVMClosure(t *testing.T) {
	expectRun(t, `
	param arg0
	var (f, y=0)
	f = func(x) {
		if x<=0{
			return 0
		}
		y++
		return f(x-1)
	}
	f(arg0)
	return y`, newOpts().Args(Int(100)), Int(100))

	expectRun(t, `
	x:=func(){
		a:=10
		g:=func(){
			b:=20
			y:=func(){
				b=21
				a=11
			}()
			return b
		}
		t := g()
		return [a, t]
	}
	return x()`, nil, Array{Int(11), Int(21)})

	expectRun(t, `
	var f
	for i:=0; i<3; i++ {
		f = func(){
			return i
		}
	}
	return f()
	`, nil, Int(3))

	expectRun(t, `
	fns :=  []
	for i:=0; i<3; i++ {
		i := i
		fns = append(fns, func(){
			return i
		})
	}

	ret := []
	for f in fns {
		ret = append(ret, f())
	}
	return ret
	`, nil, Array{Int(0), Int(1), Int(2)})
}

type testopts struct {
	globals       Object
	args          []Object
	moduleMap     *ModuleMap
	skip2pass     bool
	isCompilerErr bool
	noPanic       bool
}

func newOpts() *testopts {
	return &testopts{}
}

func (t *testopts) Globals(globals Object) *testopts {
	t.globals = globals
	return t
}

func (t *testopts) Args(args ...Object) *testopts {
	t.args = args
	return t
}

func (t *testopts) Skip2Pass() *testopts {
	t.skip2pass = true
	return t
}

func (t *testopts) CompilerError() *testopts {
	t.isCompilerErr = true
	return t
}

func (t *testopts) NoPanic() *testopts {
	t.noPanic = true
	return t
}

func (t *testopts) Module(name string, module interface{}) *testopts {
	if t.moduleMap == nil {
		t.moduleMap = NewModuleMap()
	}
	switch v := module.(type) {
	case []byte:
		t.moduleMap.AddSourceModule(name, v)
	case string:
		t.moduleMap.AddSourceModule(name, []byte(v))
	case map[string]Object:
		t.moduleMap.AddBuiltinModule(name, v)
	case Map:
		t.moduleMap.AddBuiltinModule(name, v)
	case Importable:
		t.moduleMap.Add(name, v)
	default:
		panic(fmt.Errorf("invalid module type: %T", module))
	}
	return t
}

func expectErrHas(t *testing.T, script string, opts *testopts, expectMsg string) {
	t.Helper()
	if expectMsg == "" {
		panic("expected message must not be empty")
	}
	expectErrorGen(t, script, opts, func(t *testing.T, retErr error) {
		t.Helper()
		if !strings.Contains(retErr.Error(), expectMsg) {
			require.Failf(t, "expectErrHas Failed",
				"expected error: %v, got: %v", expectMsg, retErr)
		}
	})
}

func expectErrIs(t *testing.T, script string, opts *testopts, expectErr error) {
	t.Helper()
	expectErrorGen(t, script, opts, func(t *testing.T, retErr error) {
		t.Helper()
		if !errors.Is(retErr, expectErr) {
			require.Failf(t, "expectErrorIs Failed",
				"expected error: %v, got: %v", expectErr, retErr)
		}
	})
}

func expectErrAs(t *testing.T, script string, opts *testopts, asErr interface{}, eqErr interface{}) {
	t.Helper()
	expectErrorGen(t, script, opts, func(t *testing.T, retErr error) {
		t.Helper()
		if !errors.As(retErr, asErr) {
			require.Failf(t, "expectErrorAs Type Failed",
				"expected error type: %T, got: %T(%v)", asErr, retErr, retErr)
		}
		if eqErr != nil && !reflect.DeepEqual(eqErr, asErr) {
			require.Failf(t, "expectErrorAs Equality Failed",
				"errors not equal: %[1]T(%[1]v), got: %[2]T(%[2]v)", eqErr, retErr)
		}
	})
}

func expectErrorGen(
	t *testing.T,
	script string,
	opts *testopts,
	callback func(*testing.T, error),
) {
	t.Helper()
	if opts == nil {
		opts = newOpts()
	}
	type testCase struct {
		name   string
		opts   CompilerOptions
		tracer bytes.Buffer
	}
	testCases := []testCase{
		{
			name: "default",
			opts: CompilerOptions{
				ModuleMap:      opts.moduleMap,
				OptimizeConst:  true,
				TraceParser:    true,
				TraceOptimizer: true,
				TraceCompiler:  true,
			},
		},
		{
			name: "unoptimized",
			opts: CompilerOptions{
				ModuleMap:      opts.moduleMap,
				TraceParser:    true,
				TraceOptimizer: true,
				TraceCompiler:  true,
			},
		},
	}
	if opts.skip2pass {
		testCases = testCases[:1]
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			t.Helper()
			tC.opts.Trace = &tC.tracer // nolint exportloopref
			compiled, err := Compile([]byte(script), tC.opts)
			if opts.isCompilerErr {
				require.Error(t, err)
				callback(t, err)
				return
			}
			require.NoError(t, err)
			_, err = NewVM(compiled).SetRecover(opts.noPanic).Run(opts.globals, opts.args...)
			require.Error(t, err)
			callback(t, err)
		})
	}
}

func expectRun(t *testing.T, script string, opts *testopts, expect Object) {
	t.Helper()
	if opts == nil {
		opts = newOpts()
	}
	type testCase struct {
		name   string
		opts   CompilerOptions
		tracer bytes.Buffer
	}
	testCases := []testCase{
		{
			name: "default",
			opts: CompilerOptions{
				ModuleMap:      opts.moduleMap,
				OptimizeConst:  true,
				TraceParser:    true,
				TraceOptimizer: true,
				TraceCompiler:  true,
			},
		},
		{
			name: "unoptimized",
			opts: CompilerOptions{
				ModuleMap:      opts.moduleMap,
				TraceParser:    true,
				TraceOptimizer: true,
				TraceCompiler:  true,
			},
		},
	}
	if opts.skip2pass {
		testCases = testCases[:1]
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			t.Helper()
			tC.opts.Trace = &tC.tracer // nolint exportloopref
			gotBc, err := Compile([]byte(script), tC.opts)
			require.NoError(t, err)
			// create a copy of the bytecode before execution to test bytecode
			// change after execution
			expectBc := *gotBc
			expectBc.Main = gotBc.Main.Copy().(*CompiledFunction)
			expectBc.Constants = Array(gotBc.Constants).Copy().(Array)
			vm := NewVM(gotBc)
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "------- Start Trace -------\n%s"+
						"\n------- End Trace -------\n", tC.tracer.String())
					gotBc.Fprint(os.Stderr)
					panic(r)
				}
			}()
			got, err := vm.SetRecover(opts.noPanic).Run(
				opts.globals,
				opts.args...,
			)
			if !assert.NoErrorf(t, err, "Code:\n%s\n", script) {
				gotBc.Fprint(os.Stderr)
			}
			if !reflect.DeepEqual(expect, got) {
				var buf bytes.Buffer
				gotBc.Fprint(&buf)
				t.Fatalf("Objects not equal:\nExpected:\n%s\nGot:\n%s\nScript:\n%s\n%s\n",
					tests.Sdump(expect), tests.Sdump(got), script, buf.String())
			}
			testBytecodesEqual(t, &expectBc, gotBc, true)
		})
	}
}
