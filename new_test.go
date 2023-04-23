// Put relatively new features' tests in this test file.

package ugo_test

import (
	"testing"

	. "github.com/ozanh/ugo"
)

func TestVMDestructuring(t *testing.T) {
	expectErrHas(t, `x, y = undefined; return x`,
		newOpts().CompilerError(), `Compile Error: unresolved reference "x"`)
	expectErrHas(t, `var (x, y); x, y := undefined; return x`,
		newOpts().CompilerError(), `Compile Error: no new variable on left side`)
	expectErrHas(t, `x, y = 1, 2`, newOpts().CompilerError(),
		`Compile Error: multiple expressions on the right side not supported`)

	expectRun(t, `x, y := undefined; return x`, nil, Undefined)
	expectRun(t, `x, y := undefined; return y`, nil, Undefined)
	expectRun(t, `x, y := 1; return x`, nil, Int(1))
	expectRun(t, `x, y := 1; return y`, nil, Undefined)
	expectRun(t, `x, y := []; return x`, nil, Undefined)
	expectRun(t, `x, y := []; return y`, nil, Undefined)
	expectRun(t, `x, y := [1]; return x`, nil, Int(1))
	expectRun(t, `x, y := [1]; return y`, nil, Undefined)
	expectRun(t, `x, y := [1, 2]; return x`, nil, Int(1))
	expectRun(t, `x, y := [1, 2]; return y`, nil, Int(2))
	expectRun(t, `x, y := [1, 2, 3]; return x`, nil, Int(1))
	expectRun(t, `x, y := [1, 2, 3]; return y`, nil, Int(2))
	expectRun(t, `var x; x, y := [1]; return x`, nil, Int(1))
	expectRun(t, `var x; x, y := [1]; return y`, nil, Undefined)

	expectRun(t, `x, y, z := undefined; return x`, nil, Undefined)
	expectRun(t, `x, y, z := undefined; return y`, nil, Undefined)
	expectRun(t, `x, y, z := undefined; return z`, nil, Undefined)
	expectRun(t, `x, y, z := 1; return x`, nil, Int(1))
	expectRun(t, `x, y, z := 1; return y`, nil, Undefined)
	expectRun(t, `x, y, z := 1; return z`, nil, Undefined)
	expectRun(t, `x, y, z := []; return x`, nil, Undefined)
	expectRun(t, `x, y, z := []; return y`, nil, Undefined)
	expectRun(t, `x, y, z := []; return z`, nil, Undefined)
	expectRun(t, `x, y, z := [1]; return x`, nil, Int(1))
	expectRun(t, `x, y, z := [1]; return y`, nil, Undefined)
	expectRun(t, `x, y, z := [1]; return z`, nil, Undefined)
	expectRun(t, `x, y, z := [1, 2]; return x`, nil, Int(1))
	expectRun(t, `x, y, z := [1, 2]; return y`, nil, Int(2))
	expectRun(t, `x, y, z := [1, 2]; return z`, nil, Undefined)
	expectRun(t, `x, y, z := [1, 2, 3]; return x`, nil, Int(1))
	expectRun(t, `x, y, z := [1, 2, 3]; return y`, nil, Int(2))
	expectRun(t, `x, y, z := [1, 2, 3]; return z`, nil, Int(3))
	expectRun(t, `x, y, z := [1, 2, 3, 4]; return z`, nil, Int(3))

	// test index assignments
	expectRun(t, `
	var (x = {}, y, z)
	x.a, y, z = [1, 2, 3, 4]; return x`, nil, Map{"a": Int(1)})
	expectRun(t, `
	var (x = {}, y, z)
	x.a, y, z = [1, 2, 3, 4]; return y`, nil, Int(2))
	expectRun(t, `
	var (x = {}, y, z)
	x.a, y, z = [1, 2, 3, 4]; return z`, nil, Int(3))

	expectRun(t, `
	var (x = {}, y, z)
	y, x.a, z = [1, 2, 3, 4]; return x`, nil, Map{"a": Int(2)})
	expectRun(t, `
	var (x = {}, y, z)
	y, x.a, z = [1, 2, 3, 4]; return y`, nil, Int(1))
	expectRun(t, `
	var (x = {}, y, z)
	y, x.a, z = [1, 2, 3, 4]; return z`, nil, Int(3))

	expectRun(t, `
	var (x = [0], y, z)
	x[0], y, z = [1, 2, 3, 4]; return x`, nil, Array{Int(1)})
	expectRun(t, `
	var (x = [0], y, z)
	x[0], y, z = [1, 2, 3, 4]; return y`, nil, Int(2))
	expectRun(t, `
	var (x = [0], y, z)
	x[0], y, z = [1, 2, 3, 4]; return z`, nil, Int(3))

	expectRun(t, `
	var (x = [0], y, z)
	y, x[0], z = [1, 2, 3, 4]; return x`, nil, Array{Int(2)})
	expectRun(t, `
	var (x = [0], y, z)
	y, x[0], z = [1, 2, 3, 4]; return y`, nil, Int(1))
	expectRun(t, `
	var (x = [0], y, z)
	y, x[0], z = [1, 2, 3, 4]; return z`, nil, Int(3))

	// test function calls
	expectRun(t, `
	fn := func() { 
		return [1, error("abc")]
	}
	x, y := fn()
	return [x, string(y)]`, nil, Array{Int(1), String("error: abc")})

	expectRun(t, `
	fn := func() { 
		return [1]
	}
	x, y := fn()
	return [x, y]`, nil, Array{Int(1), Undefined})
	expectRun(t, `
	fn := func() { 
		return
	}
	x, y := fn()
	return [x, y]`, nil, Array{Undefined, Undefined})
	expectRun(t, `
	fn := func() { 
		return [1, 2, 3]
	}
	x, y := fn()
	t := {a: x}
	return [x, y, t]`, nil, Array{Int(1), Int(2), Map{"a": Int(1)}})
	expectRun(t, `
	fn := func() { 
		return {}
	}
	x, y := fn()
	return [x, y]`, nil, Array{Map{}, Undefined})
	expectRun(t, `
	fn := func(v) { 
		return [1, v, 3]
	}
	var x = 10
	x, y := fn(x)
	t := {a: x}
	return [x, y, t]`, nil, Array{Int(1), Int(10), Map{"a": Int(1)}})

	// test any expression
	expectRun(t, `x, y :=  {}; return [x, y]`, nil, Array{Map{}, Undefined})
	expectRun(t, `
	var x = 2
	if x > 0 {
		fn := func(v) { 
			return [3*v, 4*v]
		}
		var y
		x, y = fn(x)
		if y != 8 {
			throw sprintf("y value expected: %s, got: %s", 8, y)
		}
	}
	return x
	`, nil, Int(6))
	expectRun(t, `
	var x = 2
	if x > 0 {
		fn := func(v) { 
			return [3*v, 4*v]
		}
		// new x symbol is created within if block
		// x in upper block is not affected
		x, y := fn(x)
		if y != 8 {
			throw sprintf("y value expected: %s, got: %s", 8, y)
		}
	}
	return x
	`, nil, Int(2))

	expectRun(t, `
	var x = 2
	if x > 0 {
		fn := func(v) {
			try {
				ret := v/2
			} catch err {
				return [0, err]
			} finally {
				if err == undefined {
					return ret
				}
			}
		}
		a, err := fn("str")
		if !isError(err) {
			throw err
		}
		if a != 0 {
			throw "a is not 0"
		}
		a, err = fn(6)
		if err != undefined {
			throw sprintf("unexpected error: %s", err)
		}
		if a != 3 {
			throw "a is not 3"
		}
		x = a
	}
	// return map to check stack pointer is correct
	return {x: x}
	`, nil, Map{"x": Int(3)})
	expectRun(t, `
	for x,y := [1, 2]; true; x++ {
		if x == 10 {
			return [x, y]
		}
	}
	`, nil, Array{Int(10), Int(2)})
	expectRun(t, `
	if x,y := [1, 2]; true {
		return [x, y]
	}
	`, nil, Array{Int(1), Int(2)})
	expectRun(t, `
	var x = 0
	for true {
		x, y := [x]
		x++
		break
	}
	return x`, nil, Int(0))
	expectRun(t, `
	x, y := func(n) {
		return repeat([n], n)
	}(3)
	return [x, y]`, nil, Array{Int(3), Int(3)})
	// closures
	expectRun(t, `
	var x = 10
	a, b := func(n) {
		x = n
	}(3)
	return [x, a, b]`, nil, Array{Int(3), Undefined, Undefined})
	expectRun(t, `
	var x = 10
	a, b := func(...args) {
		x, y := args
		return [x, y]
	}(1, 2)
	return [x, a, b]`, nil, Array{Int(10), Int(1), Int(2)})
	expectRun(t, `
	var x = 10
	a, b := func(...args) {
		var y
		x, y = args
		return [x, y]
	}(1, 2)
	return [x, a, b]`, nil, Array{Int(1), Int(1), Int(2)})

	// return implicit array if return statement's expressions are comma
	// separated which is a part of destructuring implementation to mimic multi
	// return values.
	parseErr := `Parse Error: expected operand, found 'EOF'`
	expectErrHas(t, `return 1,`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `return 1, 2,`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `var a; return a,`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `var (a, b); return a, b,`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `return 1,`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `return 1, 2,`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `var a; return a,`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `var (a, b); return a, b,`,
		newOpts().CompilerError(), parseErr)

	parseErr = `Parse Error: expected operand, found '}'`
	expectErrHas(t, `func(){ return 1, }`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `func(){ return 1, 2,}`,
		newOpts().CompilerError(), parseErr)

	expectErrHas(t, `func(){ var a; return a,}`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `func(){ var (a, b); return a, b,}`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `func(){ return 1,}`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `func(){ return 1, 2,}`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `func(){ var a; return a,}`,
		newOpts().CompilerError(), parseErr)
	expectErrHas(t, `func(){ var (a, b); return a, b,}`,
		newOpts().CompilerError(), parseErr)

	expectRun(t, `return 1, 2`, nil, Array{Int(1), Int(2)})
	expectRun(t, `a := 1; return a, a`, nil, Array{Int(1), Int(1)})
	expectRun(t, `a := 1; return a, 2`, nil, Array{Int(1), Int(2)})
	expectRun(t, `a := 1; return 2, a`, nil, Array{Int(2), Int(1)})
	expectRun(t, `a := 1; return 2, a, [3]`, nil,
		Array{Int(2), Int(1), Array{Int(3)}})
	expectRun(t, `a := 1; return [2, a], [3]`, nil,
		Array{Array{Int(2), Int(1)}, Array{Int(3)}})
	expectRun(t, `return {}, []`, nil, Array{Map{}, Array{}})
	expectRun(t, `return func(){ return 1}(), []`, nil, Array{Int(1), Array{}})
	expectRun(t, `return func(){ return 1}(), [2]`, nil,
		Array{Int(1), Array{Int(2)}})
	expectRun(t, `
	f := func() {
		return 1, 2
	}
	a, b := f()
	return a, b`, nil, Array{Int(1), Int(2)})
	expectRun(t, `
	a, b := func() {
		return 1, error("x")
	}()
	return a, "" + b`, nil, Array{Int(1), String("error: x")})
	expectRun(t, `
	a, b := func(a, b) {
		return a + 1, b + 1
	}(1, 2)
	return a, b, a*2, 3/b`, nil, Array{Int(2), Int(3), Int(4), Int(1)})
	expectRun(t, `
	return func(a, b) {
		return a + 1, b + 1
	}(1, 2), 4`, nil, Array{Array{Int(2), Int(3)}, Int(4)})

	expectRun(t, `
	param ...args

	mapEach := func(seq, fn) {
	
		if !isArray(seq) {
			return error("want array, got " + typeName(seq))
		}
	
		var out = []
	
		if sz := len(seq); sz > 0 {
			out = repeat([0], sz)
		} else {
			return out
		}
	
		try {
			for i, v in seq {
				out[i] = fn(v)
			}
		} catch err {
			println(err)
		} finally {
			return out, err
		}
	}
	
	global multiplier
	
	v, err := mapEach(args, func(x) { return x*multiplier })
	if err != undefined {
		return err
	}
	return v
	`, newOpts().
		Globals(Map{"multiplier": Int(2)}).
		Args(Int(1), Int(2), Int(3), Int(4)),
		Array{Int(2), Int(4), Int(6), Int(8)})

	expectRun(t, `
	global goFunc
	// ...
	v, err := goFunc(2)
	if err != undefined {
		return string(err)
	}
	`, newOpts().
		Globals(Map{"goFunc": &Function{
			Value: func(args ...Object) (Object, error) {
				// ...
				return Array{
					Undefined,
					ErrIndexOutOfBounds.NewError("message"),
				}, nil
			},
		}}),
		String("IndexOutOfBoundsError: message"))
}

func TestVMConst(t *testing.T) {
	expectErrHas(t, `const x = 1; x = 2`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `const x = 1; x := 2`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `const (x = 1, x = 2)`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `const x`, newOpts().CompilerError(),
		`Parse Error: missing initializer in const declaration`)
	expectErrHas(t, `const (x, y = 2)`, newOpts().CompilerError(),
		`Parse Error: missing initializer in const declaration`)

	// After iota support, `const (x=1,y)` does not throw error, like Go. It
	// uses last expression as initializer.
	expectRun(t, `const (x = 1, y)`, nil, Undefined)

	expectErrHas(t, `const (x, y)`, newOpts().CompilerError(),
		`Parse Error: missing initializer in const declaration`)
	expectErrHas(t, `
	const x = 1
	func() {
		x = 2
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `
	const x = 1
	if x > 0 {
		x = 2
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `
	const x = 1
	if x > 0 {
		return func() {
			x = 2
		}
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `
	const x = 1
	if x = 2; x > 0 {
		return
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `
	const x = 1
	for x = 1; x < 10; x++ {
		return
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `
	const x = 1
	func() {
		var y
		x, y = [1, 2]
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `
	x := 1
	func() {
		const y = 2
		x, y = [1, 2]
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "y"`)
	expectErrHas(t, `const x = 1;global x`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `const x = 1;param x`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `global x; const x = 1`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `param x; const x = 1`, newOpts().CompilerError(),
		`Compile Error: "x" redeclared in this block`)
	expectErrHas(t, `
	const x = 1
	if [2] { // not optimized
		x = 2
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `
	const x = 1
	if [2] { // not optimized
		func() {
			x = 2
		}
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)
	expectErrHas(t, `
	const x = 1
	if x {
		func() {
			func() {
				for {
					x = 2
				}
			}
		}
	}`, newOpts().CompilerError(),
		`Compile Error: assignment to constant variable "x"`)

	// FIXME: Compiler does not compile if or else blocks if condition is
	// a *BoolLit (which may be reduced by optimizer). So compiler does not
	// check whether a constant is reassigned in block to throw an error.
	// A few examples for this issue.
	expectRun(t, `
	const x = 1
	if true {
		
	} else {
		// block is not compiled
		x = 2
	}
	return x
	`, nil, Int(1))
	expectRun(t, `
	const x = 1
	if false {
		// block is not compiled
		x = 2
	}
	return x
	`, nil, Int(1))

	expectRun(t, `const x = 1; return x`, nil, Int(1))
	expectRun(t, `const x = "1"; return x`, nil, String("1"))
	expectRun(t, `const x = []; return x`, nil, Array{})
	expectRun(t, `const x = []; return x`, nil, Array{})
	expectRun(t, `const x = undefined; return x`, nil, Undefined)
	expectRun(t, `const (x = 1, y = "2"); return x, y`, nil,
		Array{Int(1), String("2")})
	expectRun(t, `
	const (
		x = 1
		y = "2"
	)
	return x, y`, nil, Array{Int(1), String("2")})
	expectRun(t, `
	const (
		x = 1
		y = x + 1
	)
	return x, y`, nil, Array{Int(1), Int(2)})
	expectRun(t, `
	const x = 1
	return func() {
		const x = x + 1
		return x
	}()`, nil, Int(2))
	expectRun(t, `
	const x = 1
	return func() {
		x := x + 1
		return x
	}()`, nil, Int(2))
	expectRun(t, `
	const x = 1
	return func() {
		return func() {
			return x + 1
		}()
	}()`, nil, Int(2))
	expectRun(t, `
	const x = 1
	for x := 10; x < 100; x++{
		return x
	}`, nil, Int(10))
	expectRun(t, `
	const (i = 1, v = 2)
	for i,v in [10] {
		v = -1
		return i
	}`, nil, Int(0))
	expectRun(t, `
	const x = 1
	return func() {
		const y = 2
		const x = y
		return x
	}() + x
	`, nil, Int(3))
	expectRun(t, `
	const x = 1
	return func() {
		const y = 2
		var x = y
		return x
	}() + x
	`, nil, Int(3))
	expectRun(t, `
	const x = 1
	func() {
		x, y := [2, 3]
	}()
	return x
	`, nil, Int(1))
	expectRun(t, `
	const x = 1
	for i := 0; i < 1; i++ {
		x, y := [2, 3]
		break
	}
	return x
	`, nil, Int(1))
	expectRun(t, `
	const x = 1
	if [1] {
		x, y := [2, 3]
	}
	return x
	`, nil, Int(1))

	expectRun(t, `
	return func() {
		const x = 1
		func() {
			x, y := [2, 3]
		}()
		return x
	}()
	`, nil, Int(1))
	expectRun(t, `
	return func() {
		const x = 1
		for i := 0; i < 1; i++ {
			x, y := [2, 3]
			break
		}
		return x
	}()
	`, nil, Int(1))
	expectRun(t, `
	return func(){
		const x = 1
		if [1] {
			x, y := [2, 3]
		}
		return x
	}()
	`, nil, Int(1))
	expectRun(t, `
	return func(){
		const x = 1
		if [1] {
			var y
			x, y := [2, 3]
		}
		return x
	}()
	`, nil, Int(1))
}

func TestConstIota(t *testing.T) {
	expectRun(t, `const x = iota; return x`, nil, Int(0))
	expectRun(t, `const x = iota; const y = iota; return x, y`, nil, Array{Int(0), Int(0)})
	expectRun(t, `const(x = iota, y = iota); return x, y`, nil, Array{Int(0), Int(1)})
	expectRun(t, `const(x = iota, y); return x, y`, nil, Array{Int(0), Int(1)})

	expectRun(t, `const(x = 1+iota, y); return x, y`, nil, Array{Int(1), Int(2)})
	expectRun(t, `const(x = 1+iota, y=iota); return x, y`, nil, Array{Int(1), Int(1)})
	expectRun(t, `const(x = 1+iota, y, z); return x, y, z`, nil, Array{Int(1), Int(2), Int(3)})
	expectRun(t, `const(x = iota+1, y, z); return x, y, z`, nil, Array{Int(1), Int(2), Int(3)})

	expectRun(t, `const(_ = iota+1, y, z); return y, z`, nil, Array{Int(2), Int(3)})

	expectRun(t, `
	const (
		x = [iota]
	)
	return x`, nil, Array{Int(0)})

	expectRun(t, `
	const (
		x = []
	)
	return x`, nil, Array{})

	expectRun(t, `
	const (
		x = [iota, iota]
	)
	return x`, nil, Array{Int(0), Int(0)})

	expectRun(t, `
	const (
		x = [iota, iota]
		y
	)
	return x, y`, nil, Array{Array{Int(0), Int(0)}, Array{Int(1), Int(1)}})

	expectRun(t, `
	const (
		x = [iota, iota]
		y
		z
	)
	return x, y, z`, nil,
		Array{Array{Int(0), Int(0)}, Array{Int(1), Int(1)}, Array{Int(2), Int(2)}})

	expectRun(t, `
	const (
		x = [iota, iota]
		y
	)
	x[0] = 2
	return x, y`, nil, Array{Array{Int(2), Int(0)}, Array{Int(1), Int(1)}})

	expectRun(t, `
	const (
		x = {}
	)
	return x`, nil, Map{})

	expectRun(t, `
	const (
		x = {iota: 1}
	)
	return x`, nil, Map{"iota": Int(1)})

	expectRun(t, `
	const (
		x = {k: iota}
	)
	return x`, nil, Map{"k": Int(0)})

	expectRun(t, `
	const (
		x = {k: iota}
		y
	)
	return x, y`, nil, Array{Map{"k": Int(0)}, Map{"k": Int(1)}})

	expectRun(t, `
	const (
		x = {k: iota}
		y
	)
	x["k"] = 2
	return x, y`, nil, Array{Map{"k": Int(2)}, Map{"k": Int(1)}})

	expectRun(t, `
	const (
		x = {k: iota}
		y
		z
	)
	return x, y, z`, nil,
		Array{Map{"k": Int(0)}, Map{"k": Int(1)}, Map{"k": Int(2)}})

	expectRun(t, `
	const (
		_ = 1 << iota
		x
		y
	)
	return x, y`, nil, Array{Int(2), Int(4)})

	expectRun(t, `
	const (
		x = 1 << iota
		_
		y
	)
	return x, y`, nil, Array{Int(1), Int(4)})

	expectRun(t, `
	const (
		x = 1 << iota
		a
		y = a
		z
	)
	return x, y, z`, nil, Array{Int(1), Int(2), Int(2)})

	expectRun(t, `
	const (
		x = 1 << iota
		_
		_
		z
	)
	return x, z`, nil, Array{Int(1), Int(8)})

	expectRun(t, `
	iota := 1
	const (
		x = 1 << iota
	)
	return x, iota`, nil, Array{Int(2), Int(1)})

	expectRun(t, `
	iota := 1
	const (
		x = 1 << iota
		y
	)
	return x, y`, nil, Array{Int(2), Int(2)})

	expectErrHas(t, `const iota = 1`,
		newOpts().CompilerError(), "Compile Error: assignment to iota")

	expectErrHas(t, `const iota = iota + 1`,
		newOpts().CompilerError(), "Compile Error: assignment to iota")

	expectErrHas(t, `
	const (
		x = 1 << iota
		iota
		y
	)
	return x, iota, y`,
		newOpts().CompilerError(), "Compile Error: assignment to iota")

	expectErrHas(t, `const x = iota; return iota`,
		newOpts().CompilerError(), `Compile Error: unresolved reference "iota"`)

	expectRun(t, `
	const (
		x = iota
		y
	)
	iota := 3
	return x, y, iota`, nil, Array{Int(0), Int(1), Int(3)})

	expectRun(t, `
	const (
		x = iota
		y
	)
	iota := 3
	const (
		a = 10+iota
		b
	)
	return x, y, iota, a, b`, nil, Array{Int(0), Int(1), Int(3), Int(13), Int(13)})

	expectRun(t, `
	const (
		x = iota
		y
	)
	const (
		a = 10+iota
		b
	)
	return x, y, a, b`, nil, Array{Int(0), Int(1), Int(10), Int(11)})

	expectRun(t, `
	const (
		x = func() { return 1 }()
		y
		z
	)
	return x, y, z`, nil, Array{Int(1), Int(1), Int(1)})

	expectRun(t, `
	const (
		x = func(x) { return x }(iota)
		y
		z
	)
	return x, y, z`, nil, Array{Int(0), Int(1), Int(2)})

	expectRun(t, `
	a:=0
	const (
		x = func() { a++; return a }()
		y
		z
	)
	return x, y, z`, nil, Array{Int(1), Int(2), Int(3)})

	expectRun(t, `
	const (
		x = 1+iota
		y = func() { return 1+x }()
		z
	)
	return x, y, z`, nil, Array{Int(1), Int(2), Int(2)})

	expectRun(t, `
	const (
		x = func() { return 1 }
		y
		z
	)
	return x(), y(), z()`, nil, Array{Int(1), Int(1), Int(1)})

	expectRun(t, `
	const (
		x = func() { return 1 }
		y
		z
	)
	return x == y && y == z`, nil, True)

	expectRun(t, `
	var a
	const (
		x = func() { return a }
		y
		z
	)
	return x != y && y != z`, nil, True)

	expectRun(t, `
	return func() {
		const (
			x = 1 << iota
			_
			y
		)
		return x, y
	}()`, nil, Array{Int(1), Int(4)})

	expectRun(t, `
	iota := 2
	return func() {
		const (
			x = 1 << iota
			_
			y
		)
		return x, y
	}()`, nil, Array{Int(4), Int(4)})

	expectRun(t, `
	const (
		x = 1 + iota + func() { 
			const (
				_ = iota
				r
			)
			return r
		}()
		y
		_
	)
	return x,y`, nil, Array{Int(2), Int(3)})

	expectRun(t, `
	const (x = iota%2?"odd":"even", y, z)
	return x,y,z`, nil, Array{String("even"), String("odd"), String("even")})
}

func TestVM_Invoke(t *testing.T) {
	applyPool := &Function{
		Name: "applyPool",
		ValueEx: func(c Call) (Object, error) {
			args := make([]Object, 0, c.Len()-1)
			for i := 1; i < c.Len(); i++ {
				args = append(args, c.Get(i))
			}
			inv := NewInvoker(c.VM(), c.Get(0))
			inv.Acquire()
			defer inv.Release()
			return inv.Invoke(args...)
		},
	}
	applyNoPool := &Function{
		Name: "applyNoPool",
		ValueEx: func(c Call) (Object, error) {
			args := make([]Object, 0, c.Len()-1)
			for i := 1; i < c.Len(); i++ {
				args = append(args, c.Get(i))
			}
			inv := NewInvoker(c.VM(), c.Get(0))
			return inv.Invoke(args...)
		},
	}
	for _, apply := range []*Function{applyPool, applyNoPool} {
		t.Run(apply.Name, func(t *testing.T) {
			t.Run("apply", func(t *testing.T) {
				scr := `
global apply
sum := func(...args) {
	println("called f", args)
	s := 0
	for v in args {
		println("v", v)
		s += v
	}
	return s
}
return apply(sum, 1, 2, 3)
`
				expectRun(t, scr,
					newOpts().Globals(Map{"apply": apply}),
					Int(6),
				)
			})

			t.Run("apply indirect", func(t *testing.T) {
				scr := `
global apply
sum := func(...args) {
	println("sum args", args)
	s := 0
	for v in args {
		println("v", v)
		s += v
	}
	return s
}
f := func(fn, ...args) {
	return fn(...args)
}
return apply(f, sum, 1, 2, 3)
`
				expectRun(t, scr,
					newOpts().Globals(Map{"apply": apply}),
					Int(6),
				)
			})

			t.Run("apply indirect 2", func(t *testing.T) {
				scr := `
global apply
sum := func(...args) {
	println("sum args", args)
	s := 0
	for v in args {
		println("v", v)
		s += v
	}
	return s
}
f := func(fn, ...args) {
	return apply(fn, ...args)
}
return apply(f, sum, 1, 2, 3)
`
				expectRun(t, scr,
					newOpts().Globals(Map{"apply": apply}),
					Int(6),
				)
			})

			t.Run("apply go func", func(t *testing.T) {
				sum := &Function{
					ValueEx: func(c Call) (Object, error) {
						s := Int(0)
						for i := 0; i < c.Len(); i++ {
							s += c.Get(i).(Int)
						}
						return s, nil
					},
				}
				scr := `
global (apply, sum)
return apply(sum, 1, 2, 3)
`
				expectRun(t, scr,
					newOpts().Globals(Map{"apply": apply, "sum": sum}),
					Int(6),
				)
			})

			t.Run("module state", func(t *testing.T) {
				scr := `
module := import("module")
module.counter = 1

global apply

inc := func(a) {
	module := import("module")
	module.counter += a
}
apply(inc, 3)
return module.counter
`
				t.Run("builtin", func(t *testing.T) {
					expectRun(t, scr,
						newOpts().
							Globals(Map{"apply": apply}).
							Module("module", Map{}),
						Int(4),
					)
				})
				t.Run("source", func(t *testing.T) {
					expectRun(t, scr,
						newOpts().
							Globals(Map{"apply": apply}).
							Module("module", `return {}`),
						Int(4),
					)
				})
			})

			t.Run("closure", func(t *testing.T) {
				scr := `
global apply

counter := 1
f1 := func(a) {
	counter += a
}

f2 := func(a) {
	counter += a
}
apply(f1, 3)
apply(f2, 5)
return counter
`
				expectRun(t, scr,
					newOpts().Globals(Map{"apply": apply}),
					Int(9),
				)
			})

			t.Run("global", func(t *testing.T) {
				scr := `
global apply
global counter

f1 := func(a) {
	counter += a
}

f2 := func(a) {
	counter += a
}
apply(f1, 3)
apply(f2, 5)
return counter
`
				expected := Int(9)
				globals := Map{"apply": apply, "counter": Int(1)}
				expectRun(t, scr,
					newOpts().Globals(globals).Skip2Pass(),
					expected,
				)
				if expected != globals["counter"] {
					t.Fatalf("expected %s, got %v", expected, globals["counter"])
				}
			})
		})
	}
}

type nameCaller struct {
	Map
	counts map[string]int
}

func (n *nameCaller) CallName(name string, c Call) (Object, error) {
	fn := n.Map[name]
	args := make([]Object, 0, c.Len())
	for i := 0; i < c.Len(); i++ {
		args = append(args, c.Get(i))
	}
	ret, err := NewInvoker(c.VM(), fn).Invoke(args...)
	n.counts[name]++
	return ret, err
}

var _ NameCallerObject = &nameCaller{}

func TestVMCallName(t *testing.T) {
	newobject := func(extended bool) *nameCaller {
		var f *Function
		if extended {
			f = &Function{
				ValueEx: func(c Call) (Object, error) {
					return c.Get(0).(Int) + 1, nil
				},
			}
		} else {
			f = &Function{
				Value: func(args ...Object) (Object, error) {
					return args[0].(Int) + 1, nil
				},
			}
		}

		return &nameCaller{Map: Map{"add1": f}, counts: map[string]int{}}
	}
	scr := `
global object

object.sub1 = func(a) {
	return a - 1
}

return [object.add1(10), object.sub1(10)]
`

	for _, extended := range []bool{false, true} {
		t.Run("extended "+Bool(extended).String(), func(t *testing.T) {
			t.Run("basic", func(t *testing.T) {
				expectRun(t, scr,
					newOpts().Globals(Map{"object": newobject(extended)}),
					Array{Int(11), Int(9)},
				)
			})

			t.Run("counts single pass", func(t *testing.T) {
				object := newobject(extended)
				expectRun(t, scr,
					newOpts().Globals(Map{"object": object}).Skip2Pass(),
					Array{Int(11), Int(9)},
				)
				if object.counts["add1"] != 1 {
					t.Fatalf("expected 1, got %d", object.counts["add1"])
				}
				if object.counts["sub1"] != 1 {
					t.Fatalf("expected 1, got %d", object.counts["sub1"])
				}
			})

			t.Run("counts all pass", func(t *testing.T) {
				object := newobject(extended)
				expectRun(t, scr,
					newOpts().Globals(Map{"object": object}),
					Array{Int(11), Int(9)},
				)
				if object.counts["add1"] <= 0 {
					t.Fatalf("expected >0, got %d", object.counts["add1"])
				}
				if object.counts["sub1"] <= 0 {
					t.Fatalf("expected >0, got %d", object.counts["sub1"])
				}
			})
		})
	}
}
