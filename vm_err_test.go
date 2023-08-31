package ugo_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ozanh/ugo/parser"
	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
)

func TestVMErrorHandlers(t *testing.T) {
	expectRun(t, `try {} catch err {} finally {}`, newOpts().Skip2Pass(), Undefined)
	expectRun(t, `try {} finally {}`, newOpts().Skip2Pass(), Undefined)
	expectRun(t, `try {} catch err {}`, newOpts().Skip2Pass(), Undefined)
	expectRun(t, `try {} catch {}`, newOpts().Skip2Pass(), Undefined)
	// test symbol scope
	expectRun(t, `var a = 1; try { a := 2 } catch err {} finally {}; return a`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `var a = 1; try {} catch err { a := 2 } finally {}; return a`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `var a = 1; try { a = 2 } catch err {} finally {}; return a`,
		newOpts().Skip2Pass(), Int(2))
	expectRun(t, `var a = 1; try {} catch err { a = 2; } finally {}; return a`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `var a; try {} catch err {} finally { a = 1 }; return a`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `var a = 1; try {} catch err {} finally { a := 2 }; return a`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `try { a := 1 } catch err {} finally { return a }`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `try { a := 1 } catch err { a = 2 } finally { return a }`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `var a = 1; try { a := 2 } catch err { a = 3 } finally { return a }`,
		newOpts().Skip2Pass(), Int(2))
	expectRun(t, `try {} catch err {} finally { return err }`,
		newOpts().Skip2Pass(), Undefined)
	expectRun(t, `try { a := 1 } catch err {} finally { return err }; return 0`,
		newOpts().Skip2Pass(), Undefined)
	expectErrHas(t, `try {} catch err {} finally { err := 1 }`,
		newOpts().Skip2Pass().CompilerError(), `Compile Error: "err" redeclared in this block`)
	expectRun(t, `
	try {
		a := 1; try {} catch err {} finally { err = 2 }
	} catch err {} finally { return err }; return 0`,
		newOpts().Skip2Pass(), Undefined)

	// return
	expectRun(t, `var a = 1; try { return a } finally { a = 2 }`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `var a = 1; try { throw "an error" } catch {} finally { a = 2 }; return a`,
		newOpts().Skip2Pass(), Int(2))
	expectRun(t, `var a = 1; try { throw "an error" } catch { return a } finally { a = 2 }; return 0`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `var a = 1; try { throw "an error" } catch {} finally { a = 2 }; return a`,
		newOpts().Skip2Pass(), Int(2))
	expectRun(t, `var a = 1; try { throw "an error" } catch err {} finally { return string(err) }; return a`,
		newOpts().Skip2Pass(), String((&Error{Message: "an error"}).String()))
	expectRun(t, `var a = 1; try { throw "an error" } catch err {} finally { return typeName(err) }; return a`,
		newOpts().Skip2Pass(), String("error"))
	expectRun(t, `var a = 1; try { a = 2 } finally { return a }; return 0`,
		newOpts().Skip2Pass(), Int(2))
	expectRun(t, `var a = 1; try { return a } finally { return 2 }; return 0`,
		newOpts().Skip2Pass(), Int(2))
	expectRun(t, `a := 1; b := 2; c := func(){ try { return a+1 } finally { b = 3 } }(); return [a, b, c]`,
		newOpts().Skip2Pass(), Array{Int(1), Int(3), Int(2)})
	expectRun(t, `
	var a;
	try {
		a := 1; try {} finally { return 2 }
	} finally { return a }; return 0`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `
	var a;
	try {
		a := 1; try {} finally { a++; return a }
	} finally { return a }; return 0`,
		newOpts().Skip2Pass(), Int(2))
	expectRun(t, `
	var a;
	try {
		a := 1; try {} finally { return a }
	} finally { a++; }; return 0`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `
	var a;
	try {
		a := 1; try { throw "an error" } catch { return a }
	} finally { a++; }; return 0`,
		newOpts().Skip2Pass(), Int(1))
	expectRun(t, `
	var a = 1;
	try {
		a := 2; try { throw "an error" } catch { return a }
	} finally { return a }; return 0`,
		newOpts().Skip2Pass(), Int(2))

	// errors
	expectErrIs(t, `throw InvalidOperatorError`, newOpts().Skip2Pass(), ErrInvalidOperator)
	var invOpErr *RuntimeError
	expectErrAs(t, `throw InvalidOperatorError`, newOpts().Skip2Pass(), &invOpErr, nil)
	require.NotNil(t, invOpErr.Err)
	require.Equal(t, "", invOpErr.Err.Message)
	require.Nil(t, invOpErr.Err.Cause)
	require.Equal(t, 1, len(invOpErr.Trace))
	require.Equal(t, parser.Pos(1), invOpErr.Trace[0])

	expectErrIs(t, `try { throw WrongNumArgumentsError } catch err { throw err }`,
		newOpts().Skip2Pass(), ErrWrongNumArguments)
	expectErrHas(t, `try { throw WrongNumArgumentsError.New("expected 1 got 2") } catch err { throw err }`,
		newOpts().Skip2Pass(), "WrongNumberOfArgumentsError: expected 1 got 2")
	var errZeroDiv *RuntimeError
	expectErrAs(t, `try { throw ZeroDivisionError.New("x") } catch err { throw err }`,
		newOpts().Skip2Pass(), &errZeroDiv, nil)
	require.NotNil(t, errZeroDiv.Err)
	require.Equal(t, "x", errZeroDiv.Err.Message)
	require.Equal(t, ErrZeroDivision, errZeroDiv.Err.Cause)
	require.Equal(t, 2, len(errZeroDiv.Trace))
	require.Equal(t, parser.Pos(7), errZeroDiv.Trace[0])
	require.Equal(t, parser.Pos(54), errZeroDiv.Trace[1])

	errZeroDiv = nil
	expectErrAs(t, `func(x) { return 1/x }(0)`, newOpts().Skip2Pass(), &errZeroDiv, nil)
	require.NotNil(t, errZeroDiv.Err)
	require.Equal(t, "", errZeroDiv.Err.Message)
	require.Equal(t, nil, errZeroDiv.Err.Cause)
	require.Equal(t, 2, len(errZeroDiv.Trace))
	require.Equal(t, parser.Pos(18), errZeroDiv.Trace[0])
	require.Equal(t, parser.Pos(1), errZeroDiv.Trace[1])

	errZeroDiv = nil
	expectErrAs(t, `1/0`, newOpts().Skip2Pass(), &errZeroDiv, nil)
	require.NotNil(t, invOpErr.Err)
	require.Equal(t, "", errZeroDiv.Err.Message)
	require.Equal(t, nil, errZeroDiv.Err.Cause)
	require.Equal(t, 1, len(errZeroDiv.Trace))
	require.Equal(t, parser.Pos(1), errZeroDiv.Trace[0])
}

func TestVMNoPanic(t *testing.T) {
	panicFunc := &Function{
		Name: "panicFunc",
		Value: func(args ...Object) (Object, error) {
			panic("panic:" + args[0].String())
		},
	}
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic but got nil")
			}
		}()
		// expectRun() is not used because panic somehow cannot be recovered in testing.
		c, err := Compile([]byte(`param panic; panic();`), CompilerOptions{})
		require.NoError(t, err)
		vm := NewVM(c)
		v, err := vm.Run(nil, panicFunc)
		t.Fatalf("expected panic but got err=%v\nreturn value=%v", err, v)
	}()

	expectRun(t, `param panic; out := 0; 
	try { panic("1") } catch { out |= 1 } finally { out |= 2 }; return out`,
		newOpts().NoPanic().Args(panicFunc), Int(1|2))
	expectRun(t, `param panic; out := 0;
	try { 
	try { panic("1") } finally { out |= 1 }
	} catch { out |= 2 } finally { out |= 4 }; return out`,
		newOpts().NoPanic().Args(panicFunc), Int(1|2|4))
	expectRun(t, `param panic; out := 0;
	try { 
	try { panic() } finally { out |= 1 }
	} catch { out |= 2 } finally { out |= 4 }; return out`,
		newOpts().NoPanic().Args(panicFunc), Int(1|2|4))
	expectRun(t, `param panic; out := 0;
	try { 
	try {} finally {  panic(); out |= 1 }
	} catch { out |= 2 } finally { out |= 4 }; return out`,
		newOpts().NoPanic().Args(panicFunc), Int(2|4))
	expectRun(t, `param panic; out := 0;
	try { 
	try {} finally { out |= 1 }; panic();
	} catch { out |= 2 } finally { out |= 4 }; return out`,
		newOpts().NoPanic().Args(panicFunc), Int(1|2|4))
	expectRun(t, `param panic; out := 0;
	try { 
	panic(); try {} finally { out |= 1 };
	} catch { out |= 2 } finally { out |= 4 }; return out`,
		newOpts().NoPanic().Args(panicFunc), Int(2|4))
	expectRun(t, `param panic; out := 0;
	try { 
	try {} catch { panic() } finally { out |= 1 };
	} catch { out |= 2 } finally { out |= 4 }; return out`,
		newOpts().NoPanic().Args(panicFunc), Int(1|4))
	expectRun(t, `param panic;
	try { 
	panic()
	} catch { return 1 } finally { return 2 }; return 0`,
		newOpts().NoPanic().Args(panicFunc), Int(2))
	expectRun(t, `param panic;
	try { 
	panic()
	} catch { return 1 } finally {}; return 0`,
		newOpts().NoPanic().Args(panicFunc), Int(1))
	expectRun(t, `param panic;
	try { 
	panic()
	} catch {} finally {}; return 0`,
		newOpts().NoPanic().Args(panicFunc), Int(0))

	expectRun(t, `param panic;
	try { 
	func() { panic() }()
	} catch { return 1 } finally {}; return 0`,
		newOpts().NoPanic().Args(panicFunc), Int(1))

	expectErrHas(t, `param panic; panic();`,
		newOpts().NoPanic().Args(panicFunc), `index out of range [0] with length 0`)
	expectRun(t, `param panic;
	try { 
		try { func() { panic() }() } finally { return 5 }
	} catch { return 1 } finally { return 2 }; return 0`,
		newOpts().NoPanic().Args(panicFunc), Int(2))
	expectRun(t, `param panic;
	try { 
		try { func() { panic() }() } finally { return 5 }
	} catch { return 1 } finally {}; return 0`,
		newOpts().NoPanic().Args(panicFunc), Int(5))
	expectRun(t, `param panic;
	try { 
		try { func() { panic() }() } catch { return 5 }
	} catch { return 1 } finally {}; return 0`,
		newOpts().NoPanic().Args(panicFunc), Int(5))
	expectRun(t, `param panic;
	try { 
		try { func() { panic() }() } finally {}
	} catch {}; return 0`,
		newOpts().NoPanic().Args(panicFunc), Int(0))
	expectErrHas(t, `param panic;
	try { 
		try { func() { panic() }() } finally {}
	} finally {}; return 0`,
		newOpts().NoPanic().Args(panicFunc), `index out of range [0] with length 0`)
}

func TestVMCatchAll(t *testing.T) {
	catchAll := `
	return func(callable, ...args) {
		try {
			return callable(...args)
		} catch err {
			return err
		}
	}`
	expectRun(t, `
	catchAll := import("catchAll")

	sum := func(a, b, c) {
		return a + b + c
	}

	strArray := func(arr) {
		var out = []
		for v in arr {
			out = append(out, string(v))
		}
		return out
	}

	return strArray([
		catchAll(sum, 1, 2, 3),
		catchAll(sum, 1, 2),
		catchAll(sum, 1),
		catchAll(sum),
		catchAll(sum, 1, 2, 3, 4),
	])
	`, newOpts().Module("catchAll", catchAll),
		Array{
			String("6"),
			String("WrongNumberOfArgumentsError: want=3 got=2"),
			String("WrongNumberOfArgumentsError: want=3 got=1"),
			String("WrongNumberOfArgumentsError: want=3 got=0"),
			String("WrongNumberOfArgumentsError: want=3 got=4"),
		},
	)

	catchAll2 := `
	return func(callable, onError, ...args) {
		var ret
		try {
			return callable(...args)
		} catch err {
			try {
				ret = onError(err)
			} catch err2 {
				ret = err2
			}
		} finally {
			if err != undefined {
				return ret
			}
		}
	}`
	expectRun(t, `
	catchAll2 := import("catchAll2")

	sum := func(a, b, c) {
		return a + b + c
	}
	var counter = -1
	onError := func(err) {
		if isError(err) {
			try {
				return counter
			} finally {
				counter--
			}
		}
		throw "onError must be called on error"
	}

	return [
		catchAll2(sum, onError, 10, 20, 30),
		catchAll2(sum, onError, 10, 20),
		catchAll2(sum, onError, 10),
		catchAll2(sum, onError),
		catchAll2(sum, onError, 10, 20, 30, 40),
		catchAll2(sum, onError, 11, 21, 31),
	]
	`, newOpts().Module("catchAll2", catchAll2),
		Array{
			Int(60),
			Int(-1),
			Int(-2),
			Int(-3),
			Int(-4),
			Int(63),
		},
	)
}

func TestVMAssert(t *testing.T) {
	g := Map{}
	expectRun(t, `
	global errs

	assertTrue := func(v, msg) {
		if !v {
			msg := string(msg)
			throw msg
		}
	}

	assertTrue(errs == undefined, "errs must be undefined")
	assertTrue(isCallable(sprintf), "sprintf is not a callable")
	assertTrue(isFunction(sprintf), "sprintf is not a function")
	assertTrue(isCallable(assertTrue), "assertTrue is not a callable")
	assertTrue(isFunction(assertTrue), "assertTrue is not a function")

	var (
		numFails = 0,
		numRun = 0
	)
	arr := [1, 2u, 3.0, "", error("x")]
	assertTrue(isIterable(arr), "arr is not iterable")
	for i, v in arr {
		try {
			assertTrue(bool(v), sprintf("#%d is not true", i))
		} catch err {
			numFails++
			errs = append(errs, string(err))
		} finally {
			numRun++
		}
	}
	assertTrue(numFails > 0, sprintf("numFails expected > 0 but got %d", numFails))
	assertTrue(numRun == len(arr), sprintf("numRun expected %d but got %d", len(arr), numRun))
	return [numFails, numRun]
	`, newOpts().Globals(g).Skip2Pass(),
		Array{Int(2), Int(5)},
	)
	require.Equal(t, 1, len(g))
	require.Equal(t, Array{
		String("error: #3 is not true"),
		String("error: #4 is not true"),
	}, g["errs"])
}

func TestVMLoop(t *testing.T) {
	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
		  try {
			try {
			  //continue
			} finally {
			  x++
			}
		  } catch err {
			throw err
		  } finally {
			x++
		  }
		}
		return x
	  }
	return f()
	`, nil, Int(10))

	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
		  try {
			try {
			  continue
			} finally {
			  x++
			}
		  } catch err {
			throw err
		  } finally {
			x++
		  }
		}
		return x
	  }
	return f()
	`, nil, Int(10))

	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
		  try {
			try {
			  break
			} finally {
			  x++
			}
		  } catch err {
			throw err
		  } finally {
			x++
		  }
		}
		return x
	  }
	return f()
	`, nil, Int(2))

	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
		  break
		  try {
			try {
			  
			} finally {
			  x++
			}
		  } catch err {
			throw err
		  } finally {
			x++
		  }
		}
		return x
	  }
	return f()
	`, nil, Int(0))

	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
		  continue
		  try {
			try {
			  
			} finally {
			  x++
			}
		  } catch err {
			throw err
		  } finally {
			x++
		  }
		}
		return x
	  }
	return f()
	`, nil, Int(0))

	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
		  try {
			try {
			  
			} finally {
			  x++
			}
		  } catch err {
			
		  } finally {
			break
			x++
		  }
		}
		return x
	  }
	return f()
	`, nil, Int(1))

	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
		  try {
			try {
			  break
			} finally {
			  x++
			}
		  } catch err {
			
		  } finally {
			break
			x++
		  }
		}
		return x
	  }
	return f()
	`, nil, Int(1))

	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
		  try {
			try {
			  break
			} finally {
			  x++
			}
		  } catch err {
			
		  } finally {
			continue
			x++
		  }
		}
		return x
	  }
	return f()
	`, nil, Int(5))

	expectRun(t, `
	var f = func() {
		var x = 0
		for i := 0; i < 5; i++ {
			break
			x++
		}
		return x
	  }
	return f()
	`, nil, Int(0))

	expectRun(t, `
	var f = func() {
		var x = 0
		try {
			for i := 5; i > 0; i-- {
				continue
			}
		} finally {
			x++
		}
		return x
	  }
	return f()
	`, nil, Int(1))

	expectRun(t, `
	var x = 0
	try {
		for i := 5; i > 0; i-- {
			continue
		}
	} finally {
		x++
	}
	return x
	`, nil, Int(1))
}

func TestVMErrorUnwrap(t *testing.T) {
	err1 := errors.New("err1")
	var g Object = Map{"fn": &Function{
		Value: func(args ...Object) (Object, error) {
			return nil, err1
		},
	}}
	expectErrIs(t, `global fn; fn()`, newOpts().Globals(g), err1)
	expectErrIs(t, `import("module")()`,
		newOpts().Globals(g).Module("module", `global fn; return fn`), err1)

	g.(Map)["fn"] = &Function{
		Value: func(args ...Object) (Object, error) {
			return &Error{Cause: err1}, nil
		},
	}
	expectErrIs(t, `global fn; throw fn()`, newOpts().Globals(g), err1)

	g.(Map)["fn"] = &Function{
		Value: func(args ...Object) (Object, error) {
			return ErrZeroDivision, nil
		},
	}
	expectErrIs(t, `global fn; throw fn()`,
		newOpts().Globals(g), ErrZeroDivision)

	g.(Map)["fn"] = &Function{
		Value: func(args ...Object) (Object, error) {
			return nil, ErrZeroDivision
		},
	}
	expectErrIs(t, `global fn; fn()`,
		newOpts().Globals(g), ErrZeroDivision)

	expectErrIs(t, `throw TypeError`, newOpts().Globals(g), ErrType)
	expectErrIs(t, `throw TypeError.New("foo")`, newOpts().Globals(g), ErrType)
}

func TestVMExamples(t *testing.T) {
	ex1Module := `
	var numOfErrors = 0

	sum := func(check, ...args) {
		total := 0
		try {
			i := 0
			for i, v in args {
				if err := check.Value(v); err != undefined {
					throw err
				}
				total += v
			}
		} catch err  {
			printf("sum func has error: %v at index %d\n", err, i)
			throw err // re-throw error after printing
		} finally {
			if err != undefined {
				numOfErrors++ // free variable
			}
		}
		return total
	}
	// return a map to the module importer to export objects.
	return {
		Sum: sum,
		NumOfErrors: func() { return numOfErrors },
	}
`
	ex1MainScript := `
	// This example is to show some features of uGO.

	// provide arguments as if main module body is a function.
	param (a0, a1, ...args)
	
	intSum := func(callback, ...args) {
		/* functions can accept variable number of arguments. */
		var check = {
			Value: func(v) {
				if !isInt(v) {
					return TypeError.New(sprintf("want int, got %s", typeName(v)))
				}
			},
		}
		return callback(check, ...args)
	}
	
	// use global to export an Object to script
	global DoCleanup
	
	// import a source module or a builtin module
	var module = import("module")
	
	try {
		try {
			total := intSum(module.Sum, a0, a1, ...(args || []))
		} catch err {
			if isError(err, TypeError) {
				// handle specific error type.
			}
		} finally {
			// variables defined in try or catch block are visible in finally block.
			// re-importing source module does not reset state of the module.
			module := import("module")
			return {
				Total: total,
				Error: sprintf("%+v", err),
				ModuleErrors: module.NumOfErrors(),
			}
		}
	} finally {
		DoCleanup()
	}
`

	var cleanupCall int
	expectRun(t, ex1MainScript,
		newOpts().Module("module", ex1Module).Globals(Map{
			"DoCleanup": &Function{
				Value: func(args ...Object) (Object, error) {
					// a dummy callable to export to script
					cleanupCall++
					return Undefined, nil
				},
			},
		}).Args(Int(1), Int(2), Int(3)).Skip2Pass(),
		Map{"Total": Int(6), "ModuleErrors": Int(0), "Error": String("undefined")})
	require.Equal(t, 1, cleanupCall)

	oldPrintWriter := PrintWriter
	defer func() {
		PrintWriter = oldPrintWriter
	}()
	printWriter := bytes.NewBuffer(nil)
	PrintWriter = printWriter
	cleanupCall = 0
	expectRun(t, ex1MainScript,
		newOpts().Module("module", ex1Module).Globals(Map{
			"DoCleanup": &Function{
				Value: func(args ...Object) (Object, error) {
					// a dummy callable to export to script
					cleanupCall++
					return Undefined, nil
				},
			},
		}).Args(Undefined, Undefined).Skip2Pass(),
		Map{
			"Total":        Undefined,
			"ModuleErrors": Int(1),
			"Error": String(`TypeError: want int, got undefined
	at (main):27:4
	   (main):16:3
	   module:16:4
	   module:10:6`),
		})
	require.Equal(t, 1, cleanupCall)
	require.Equal(t,
		"sum func has error: TypeError: want int, got undefined at index 0\n",
		printWriter.String())

	expectRun(t, `
	param ...args

	mapEach := func(seq, fn) {
		if !isArray(seq) {
			return error("want array, got "+typeName(seq))
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
			return err == undefined ? out : err
		}
	}

	global multiplier

	return mapEach(args, func(x) { return x*multiplier })
	`, newOpts().
		Globals(Map{"multiplier": Int(2)}).
		Args(Int(1), Int(2), Int(3), Int(4)),
		Array{Int(2), Int(4), Int(6), Int(8)},
	)

	scr := `
	param a0
	global (notAnInt, zeroDivision)

	var ErrNotAnInt = error("not an integer")

	fn := func(x) {
		if !isInt(x) {
			throw ErrNotAnInt
		}
		return 10 / x
	}
	
	try {
	   result := fn(a0)
	} catch myerr {
	
		if isError(myerr, ErrNotAnInt) {
			notAnInt = true
		} else if isError(myerr, ZeroDivisionError) {
			zeroDivision = true
		}
	
	} finally {
		if myerr != undefined {
			return -1
		}
		return result
	}
`
	var g Object = Map{}
	expectRun(t, scr, newOpts().Globals(g).Args(Undefined), Int(-1))
	require.Equal(t, 1, len(g.(Map)))
	require.Equal(t, True, g.(Map)["notAnInt"])

	g = Map{}
	expectRun(t, scr, newOpts().Globals(g).Args(Int(0)), Int(-1))
	require.Equal(t, 1, len(g.(Map)))
	require.Equal(t, True, g.(Map)["zeroDivision"])

	expectRun(t, scr, newOpts().Args(Int(2)), Int(5))

	g = &SyncMap{Value: Map{"stats": Map{"fn1": Int(0), "fn2": Int(0)}}}
	expectRun(t, `
	global stats

	fn1 := func() {
		stats.fn1++
		/* ... */
	}

	fn1()

	fn2 := import("module")
	fn2()
	`, newOpts().Module("module", `
	global stats

	return func() {
		stats.fn2++
		/* ... */
	}
	`).Globals(g).Skip2Pass(), Undefined)
	require.Equal(t, Int(1), g.(*SyncMap).Value["stats"].(Map)["fn1"])
	require.Equal(t, Int(1), g.(*SyncMap).Value["stats"].(Map)["fn2"])
}
