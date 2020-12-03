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
	expectErrHas(t, `const (x = 1, y)`, newOpts().CompilerError(),
		`Parse Error: missing initializer in const declaration`)
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
