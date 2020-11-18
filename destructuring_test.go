package ugo_test

import (
	"testing"

	. "github.com/ozanh/ugo"
)

func TestDestructuring(t *testing.T) {
	expectErrHas(t, `x, y = undefined; return x`,
		newOpts().CompilerError(), `Compile Error: unresolved reference "x"`)
	expectErrHas(t, `var (x, y); x, y := undefined; return x`,
		newOpts().CompilerError(), `Compile Error: no new variable on left side`)

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
}
