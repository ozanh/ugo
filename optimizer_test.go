package ugo_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
	"github.com/ozanh/ugo/token"
)

func TestOptimizer(t *testing.T) {
	type values struct {
		s  string
		c  Object
		cf *CompiledFunction
	}

	f := compFunc(concatInsts(
		makeInst(OpConstant, 0),
		makeInst(OpPop),
		makeInst(OpReturn, 0),
	))

	testCases := []values{
		{s: `1 + 2`, c: Int(3), cf: f},
		{s: `1 - 2`, c: Int(-1), cf: f},
		{s: `2 * 2`, c: Int(4), cf: f},
		{s: `2 / 2`, c: Int(1), cf: f},
		{s: `1 << 2`, c: Int(4), cf: f},
		{s: `4 >> 2`, c: Int(1), cf: f},
		{s: `4 % 3`, c: Int(1), cf: f},
		{s: `1 & 2`, c: Int(0), cf: f},
		{s: `1 | 2`, c: Int(3), cf: f},
		{s: `1 ^ 2`, c: Int(3), cf: f},
		{s: `2 &^ 3`, c: Int(0), cf: f},
		{s: `1 == 2`, c: False, cf: f},
		{s: `1 != 2`, c: True, cf: f},
		{s: `1 < 2`, c: True, cf: f},
		{s: `1 <= 2`, c: True, cf: f},
		{s: `1 > 2`, c: False, cf: f},
		{s: `1 >= 2`, c: False, cf: f},
		{s: `!0`, c: True, cf: f},
		{s: `!1`, c: False, cf: f},
		{s: `-1`, c: Int(-1), cf: f},
		{s: `+1`, c: Int(1), cf: f},
		{s: `(1 + 2)`, c: Int(3), cf: f},
		{s: `1 + 2 + 3`, c: Int(6), cf: f},
		{s: `1 + (2 + 3)`, c: Int(6), cf: f},
		{s: `(1 + 2 + 3)`, c: Int(6), cf: f},
		{s: `1 + (2 + 3 + 4)`, c: Int(10), cf: f},
		{s: `(1 + 2) + (3 + 4)`, c: Int(10), cf: f},
		{s: `!(1 << 2)`, c: False, cf: f},

		{s: `1u + 2u`, c: Uint(3), cf: f},
		{s: `1u - 2u`, c: Uint(^uint64(0)), cf: f},
		{s: `2u * 2u`, c: Uint(4), cf: f},
		{s: `2u / 2u`, c: Uint(1), cf: f},
		{s: `1u << 2u`, c: Uint(4), cf: f},
		{s: `4u >> 2u`, c: Uint(1), cf: f},
		{s: `4u % 3u`, c: Uint(1), cf: f},
		{s: `1u & 2u`, c: Uint(0), cf: f},
		{s: `1u | 2u`, c: Uint(3), cf: f},
		{s: `1u ^ 2u`, c: Uint(3), cf: f},
		{s: `2u &^ 3u`, c: Uint(0), cf: f},
		{s: `1u == 2u`, c: False, cf: f},
		{s: `1u != 2u`, c: True, cf: f},
		{s: `1u < 2u`, c: True, cf: f},
		{s: `1u <= 2u`, c: True, cf: f},
		{s: `1u > 2u`, c: False, cf: f},
		{s: `1u >= 2u`, c: False, cf: f},
		{s: `!0u`, c: True, cf: f},
		{s: `!1u`, c: False, cf: f},
		{s: `-1u`, c: Uint(^uint64(0)), cf: f},
		{s: `+1u`, c: Uint(1), cf: f},

		{s: `1.0 + 2.0`, c: Float(3), cf: f},
		{s: `1.0 - 2.0`, c: Float(-1), cf: f},
		{s: `2.0 * 2.0`, c: Float(4), cf: f},
		{s: `2.0 / 2.0`, c: Float(1), cf: f},
		{s: `1.0 == 2.0`, c: False, cf: f},
		{s: `1.0 != 2.0`, c: True, cf: f},
		{s: `1.0 < 2.0`, c: True, cf: f},
		{s: `1.0 <= 2.0`, c: True, cf: f},
		{s: `1.0 > 2.0`, c: False, cf: f},
		{s: `1.0 >= 2.0`, c: False, cf: f},
		{s: `!0.0`, c: False, cf: f},
		{s: `!1.0`, c: False, cf: f},
		{s: `-1.0`, c: Float(-1), cf: f},
		{s: `+1.0`, c: Float(1), cf: f},

		{s: `1 + true`, c: Int(2), cf: f},
		{s: `true + 1`, c: Int(2), cf: f},
		{s: `1 - false`, c: Int(1), cf: f},
		{s: `false - 1`, c: Int(-1), cf: f},
		{s: `2 * false`, c: Int(0), cf: f},
		{s: `2 / (true + true)`, c: Int(1), cf: f},
		{s: `2 / (true + false)`, c: Int(2), cf: f},
		{s: `false / true`, c: Int(0), cf: f},
		{s: `1 << (true + 1)`, c: Int(4), cf: f},
		{s: `true << 2`, c: Int(4), cf: f},
		{s: `4 >> (1 + true)`, c: Int(1), cf: f},
		{s: `4 % true`, c: Int(0), cf: f},
		{s: `true & 2`, c: Int(0), cf: f},
		{s: `2 & true`, c: Int(0), cf: f},
		{s: `true | 2`, c: Int(3), cf: f},
		{s: `2 | true`, c: Int(3), cf: f},
		{s: `1 ^ (true + true)`, c: Int(3), cf: f},
		{s: `(true + true) ^ 1`, c: Int(3), cf: f},
		{s: `(2 * true) &^ 3`, c: Int(0), cf: f},
		{s: `1 == true * 2`, c: False, cf: f},
		{s: `true != 2`, c: True, cf: f},
		{s: `2 != true`, c: True, cf: f},
		{s: `true < 2`, c: True, cf: f},
		{s: `true <= 2`, c: True, cf: f},
		{s: `true > 2`, c: False, cf: f},
		{s: `true >= 2`, c: False, cf: f},
		{s: `2 < true`, c: False, cf: f},
		{s: `2 <= true`, c: False, cf: f},
		{s: `2 > true`, c: True, cf: f},
		{s: `2 >= true`, c: True, cf: f},
		{s: `!false`, c: True, cf: f},
		{s: `!true`, c: False, cf: f},
		{s: `-true`, c: Int(-1), cf: f},
		{s: `+true`, c: Int(1), cf: f},
		{s: `bool(0)`, c: False, cf: f},
		{s: `bool(1)`, c: True, cf: f},

		{s: `"a" + "b"`, c: String("ab"), cf: f},
		{s: `"a" + 1`, c: String("a1"), cf: f},
		{s: `"a" + 1u`, c: String("a1"), cf: f},
		{s: `"a" + 'c'`, c: String("ac"), cf: f},
		{s: `'c' + "a"`, c: String("ca"), cf: f},
		{s: `"a" + "b" + "c"`, c: String("abc"), cf: f},
		{s: `"a" + 'b' + "c"`, c: String("abc"), cf: f},
		{s: `"a" + 1 + "c"`, c: String("a1c"), cf: f},
		{s: `char(0)`, c: Char(0), cf: f},

		{s: `!undefined`, c: True, cf: f},
		{s: `!!undefined`, c: False, cf: f},
	}

	for _, tC := range testCases {
		t.Run(tC.s, func(t *testing.T) {
			expectEval(t, tC.s,
				bytecode(
					Array{tC.c},
					tC.cf,
				))
		})
	}

	testCases2 := make([]values, len(testCases))

	f = compFunc(concatInsts(
		makeInst(OpConstant, 0),
		makeInst(OpReturn, 1),
	))
	for i, tC := range testCases {
		testCases2[i].s = "return " + tC.s
		testCases2[i].c = tC.c
		testCases2[i].cf = f
	}
	for _, tC := range testCases2 {
		t.Run(tC.s, func(t *testing.T) {
			expectEval(t, tC.s,
				bytecode(
					Array{tC.c},
					tC.cf,
				))
		})
	}

	testCases3 := make([]values, len(testCases2))

	f = compFunc(concatInsts(
		makeInst(OpConstant, 1),
		makeInst(OpCall, 0, 0),
		makeInst(OpReturn, 1),
	))
	for i, tC := range testCases2 {
		testCases3[i].s = fmt.Sprintf(`return func(){ %s }()`, tC.s)
		testCases3[i].c = tC.c
		testCases3[i].cf = f
	}
	ff := compFunc(
		concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpReturn, 1),
		),
	)
	for _, tC := range testCases3 {
		t.Run(tC.s, func(t *testing.T) {
			expectEval(t, tC.s,
				bytecode(
					Array{tC.c, ff},
					tC.cf,
				))
		})
	}

	testCases4 := make([]values, len(testCases))

	f = compFunc(concatInsts(
		makeInst(OpConstant, 0),
		makeInst(OpSetLocal, 0),
		makeInst(OpReturn, 0),
	),
		withLocals(1),
	)

	for i, tC := range testCases {
		testCases4[i].s = fmt.Sprintf(`var x = %s`, tC.s)
		testCases4[i].c = tC.c
		testCases4[i].cf = f
	}
	for _, tC := range testCases4 {
		t.Run(tC.s, func(t *testing.T) {
			expectEval(t, tC.s,
				bytecode(
					Array{tC.c},
					tC.cf,
				))
		})
	}

	testCases5 := make([]values, len(testCases))

	for i, tC := range testCases {
		testCases5[i].s = fmt.Sprintf(`x := %s`, tC.s)
		testCases5[i].c = tC.c
		testCases5[i].cf = f
	}
	for _, tC := range testCases5 {
		t.Run(tC.s, func(t *testing.T) {
			expectEval(t, tC.s,
				bytecode(
					Array{tC.c},
					tC.cf,
				))
		})
	}
}

func TestOptimizerIf(t *testing.T) {
	expectEval(t, `if 1+2 {}`,
		bytecode(
			Array{},
			compFunc(concatInsts(
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `if 1+2 {} else { return 3}`,
		bytecode(
			Array{},
			compFunc(concatInsts(
				makeInst(OpReturn, 0),
			)),
		))
	// TODO: improve this, unnecessary jumps
	expectEval(t, `if 1-1 {} else if "a"+2 { return 3*4 }`,
		bytecode(
			Array{Int(12)},
			compFunc(concatInsts(
				makeInst(OpJump, 6),
				makeInst(OpJump, 11),
				makeInst(OpConstant, 0),
				makeInst(OpReturn, 1),
				makeInst(OpReturn, 0),
			)),
		))
}

func TestOptimizerFor(t *testing.T) {
	expectEval(t, `for 1+2 {}`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpJumpFalsy, 9),
				makeInst(OpJump, 0),
				makeInst(OpReturn, 0),
			)),
		))

	expectEval(t, `for { 1 + 2 }`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpPop),
				makeInst(OpJump, 0),
				makeInst(OpReturn, 0),
			)),
		))

	expectEval(t, `for i:=2*3; i<10+4; i+=2*2 {}`,
		bytecode(
			Array{Int(6), Int(14), Int(4)},
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpSetLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 1),
				makeInst(OpBinaryOp, int(token.Less)),
				makeInst(OpJumpFalsy, 27),
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 2),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpSetLocal, 0),
				makeInst(OpJump, 5),
				makeInst(OpNull),
				makeInst(OpSetLocal, 0),
				makeInst(OpReturn, 0),
			),
				withLocals(1),
			),
		))
}

func TestOptimizerTryThrow(t *testing.T) {
	expectEval(t, `
		try {
			1 + 2 
		} catch { 
			3.0 + 4.0 
		} finally {
			throw "a" + string(1) + "b"
		}`,
		bytecode(
			Array{Int(3), Float(7), String("a1b")},
			compFunc(concatInsts(
				makeInst(OpSetupTry, 12, 18),
				makeInst(OpConstant, 0),
				makeInst(OpPop),
				makeInst(OpJump, 18),
				makeInst(OpSetupCatch),
				makeInst(OpPop),
				makeInst(OpConstant, 1),
				makeInst(OpPop),
				makeInst(OpSetupFinally),
				makeInst(OpConstant, 2),
				makeInst(OpThrow, 1),
				makeInst(OpThrow, 0),
				makeInst(OpReturn, 0),
			)),
		))
}

func TestOptimizerMapSliceExpr(t *testing.T) {
	expectEval(t, `[][1+2]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpArray, 0),
				makeInst(OpConstant, 0),
				makeInst(OpGetIndex, 1),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `[][int(1+2)]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpArray, 0),
				makeInst(OpConstant, 0),
				makeInst(OpGetIndex, 1),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `[][1+2:]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpArray, 0),
				makeInst(OpConstant, 0),
				makeInst(OpNull),
				makeInst(OpSliceIndex),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `[][int(1u+2u):]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpArray, 0),
				makeInst(OpConstant, 0),
				makeInst(OpNull),
				makeInst(OpSliceIndex),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `[][:1+2]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpArray, 0),
				makeInst(OpNull),
				makeInst(OpConstant, 0),
				makeInst(OpSliceIndex),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `[][:int(1+2u)]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpArray, 0),
				makeInst(OpNull),
				makeInst(OpConstant, 0),
				makeInst(OpSliceIndex),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `[1+2]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpArray, 1),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `[bool(1+2)]`,
		bytecode(
			Array{True},
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpArray, 1),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `{}[1+2]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpMap, 0),
				makeInst(OpConstant, 0),
				makeInst(OpGetIndex, 1),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `{}[int(1+2)]`,
		bytecode(
			Array{Int(3)},
			compFunc(concatInsts(
				makeInst(OpMap, 0),
				makeInst(OpConstant, 0),
				makeInst(OpGetIndex, 1),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `{a: 1+2}`,
		bytecode(
			Array{String("a"), Int(3)},
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpConstant, 1),
				makeInst(OpMap, 2),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
	expectEval(t, `{a: uint(1+2)}`,
		bytecode(
			Array{String("a"), Uint(3)},
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpConstant, 1),
				makeInst(OpMap, 2),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		))
}

func TestOptimizerCondExpr(t *testing.T) {
	type values struct {
		s  string
		c  Object
		cf *CompiledFunction
	}
	f := compFunc(concatInsts(
		makeInst(OpConstant, 0),
		makeInst(OpPop),
		makeInst(OpReturn, 0),
	))
	testCases := []values{
		{s: `1 ? 2 : 3`, c: Int(2), cf: f},
		{s: `0 ? 2 : 3`, c: Int(3), cf: f},
		{s: `1 ? 2 + 5 : 3`, c: Int(7), cf: f},
		{s: `0 ? 2 : 3 + 4`, c: Int(7), cf: f},
		{s: `true ? 2 + 5 + 1 : 3`, c: Int(8), cf: f},
		{s: `false ? 2 : 3 + 4 + 1`, c: Int(8), cf: f},
		{s: `1 - 1 ? 2 + 5 : 3`, c: Int(3), cf: f},
		{s: `0 + 1 ? 2 : 3 + 4`, c: Int(2), cf: f},
		{s: `"" ? 2 : 3 + 4`, c: Int(7), cf: f},
		{s: `!"" ? 2 : 3 + 4`, c: Int(2), cf: f},

		{s: `a := 0; 1 ? a : 3`, c: Int(0),
			cf: compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpSetLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			),
				withLocals(1),
			),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.s, func(t *testing.T) {
			expectEval(t, tC.s,
				bytecode(
					Array{tC.c},
					tC.cf,
				))
		})
	}
}

func TestOptimizerError(t *testing.T) {
	expectEvalError(t, `
	try { 1 / 0 } catch err { } finally { }
	`, "Optimizer Error: ZeroDivisionError: \n\tat")

	// two errors found by optimizer is reported as multipleErr but
	// Error() method returns first error's message.
	// Errors on the same line are discarded by optimizer.
	bc, err := Compile([]byte(`
	1/0;2/0
	1/0;`), DefaultCompilerOptions)
	require.Nil(t, bc)
	require.Error(t, err)
	require.Equal(t,
		"Optimizer Error: ZeroDivisionError: \n\tat (main):2:2",
		err.Error(),
	)
	// test + flag gets all
	require.Equal(t,
		"multiple errors:\n Optimizer Error: ZeroDivisionError:"+
			" \n\tat (main):2:2\n Optimizer Error: ZeroDivisionError:"+
			" \n\tat (main):3:2",
		fmt.Sprintf("%+v", err),
	)
	// test error implements interface { Errors() []error }
	if m, ok := err.(interface {
		Errors() []error
	}); !ok {
		t.Fatalf("error does not implement interface { Errors() []error }")
	} else {
		require.Equal(t, 2, len(m.Errors()))
	}
}

func expectEval(t *testing.T, script string, expected *Bytecode) {
	t.Helper()
	opts := DefaultCompilerOptions
	require.True(t, opts.OptimizeConst)
	require.True(t, opts.OptimizeExpr)
	opts.OptimizerMaxCycle = 1<<8 - 1
	expectCompileWithOpts(t, script, opts, expected)
}

func expectEvalError(t *testing.T, script, errStr string) {
	t.Helper()
	opts := DefaultCompilerOptions
	require.True(t, opts.OptimizeConst)
	require.True(t, opts.OptimizeExpr)
	opts.OptimizerMaxCycle = 1<<8 - 1
	expectCompileErrorWithOpts(t, script, opts, errStr)
}
