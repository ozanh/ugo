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

	defF := compFunc(concatInsts(makeInst(OpConstant, 0)))

	falseF := compFunc(concatInsts(makeInst(OpFalse)))

	trueF := compFunc(concatInsts(makeInst(OpTrue)))

	testCases := []values{
		{s: `1 + 2`, c: Int(3), cf: defF},
		{s: `1 - 2`, c: Int(-1), cf: defF},
		{s: `2 * 2`, c: Int(4), cf: defF},
		{s: `2 / 2`, c: Int(1), cf: defF},
		{s: `1 << 2`, c: Int(4), cf: defF},
		{s: `4 >> 2`, c: Int(1), cf: defF},
		{s: `4 % 3`, c: Int(1), cf: defF},
		{s: `1 & 2`, c: Int(0), cf: defF},
		{s: `1 | 2`, c: Int(3), cf: defF},
		{s: `1 ^ 2`, c: Int(3), cf: defF},
		{s: `2 &^ 3`, c: Int(0), cf: defF},
		{s: `1 == 2`, cf: falseF},
		{s: `1 != 2`, cf: trueF},
		{s: `1 < 2`, cf: trueF},
		{s: `1 <= 2`, cf: trueF},
		{s: `1 > 2`, cf: falseF},
		{s: `1 >= 2`, cf: falseF},
		{s: `!0`, cf: trueF},
		{s: `!1`, cf: falseF},
		{s: `-1`, c: Int(-1), cf: defF},
		{s: `+1`, c: Int(1), cf: defF},
		{s: `(1 + 2)`, c: Int(3), cf: defF},
		{s: `1 + 2 + 3`, c: Int(6), cf: defF},
		{s: `1 + (2 + 3)`, c: Int(6), cf: defF},
		{s: `(1 + 2 + 3)`, c: Int(6), cf: defF},
		{s: `1 + (2 + 3 + 4)`, c: Int(10), cf: defF},
		{s: `(1 + 2) + (3 + 4)`, c: Int(10), cf: defF},
		{s: `!(1 << 2)`, cf: falseF},

		{s: `1u + 2u`, c: Uint(3), cf: defF},
		{s: `1u - 2u`, c: Uint(^uint64(0)), cf: defF},
		{s: `2u * 2u`, c: Uint(4), cf: defF},
		{s: `2u / 2u`, c: Uint(1), cf: defF},
		{s: `1u << 2u`, c: Uint(4), cf: defF},
		{s: `4u >> 2u`, c: Uint(1), cf: defF},
		{s: `4u % 3u`, c: Uint(1), cf: defF},
		{s: `1u & 2u`, c: Uint(0), cf: defF},
		{s: `1u | 2u`, c: Uint(3), cf: defF},
		{s: `1u ^ 2u`, c: Uint(3), cf: defF},
		{s: `2u &^ 3u`, c: Uint(0), cf: defF},
		{s: `1u == 2u`, cf: falseF},
		{s: `1u != 2u`, cf: trueF},
		{s: `1u < 2u`, cf: trueF},
		{s: `1u <= 2u`, cf: trueF},
		{s: `1u > 2u`, cf: falseF},
		{s: `1u >= 2u`, cf: falseF},
		{s: `!0u`, cf: trueF},
		{s: `!1u`, cf: falseF},
		{s: `-1u`, c: Uint(^uint64(0)), cf: defF},
		{s: `+1u`, c: Uint(1), cf: defF},

		{s: `1.0 + 2.0`, c: Float(3), cf: defF},
		{s: `1.0 - 2.0`, c: Float(-1), cf: defF},
		{s: `2.0 * 2.0`, c: Float(4), cf: defF},
		{s: `2.0 / 2.0`, c: Float(1), cf: defF},
		{s: `1.0 == 2.0`, cf: falseF},
		{s: `1.0 != 2.0`, cf: trueF},
		{s: `1.0 < 2.0`, cf: trueF},
		{s: `1.0 <= 2.0`, cf: trueF},
		{s: `1.0 > 2.0`, cf: falseF},
		{s: `1.0 >= 2.0`, cf: falseF},
		{s: `!0.0`, cf: falseF},
		{s: `!1.0`, cf: falseF},
		{s: `-1.0`, c: Float(-1), cf: defF},
		{s: `+1.0`, c: Float(1), cf: defF},

		{s: `1 + true`, c: Int(2), cf: defF},
		{s: `true + 1`, c: Int(2), cf: defF},
		{s: `1 - false`, c: Int(1), cf: defF},
		{s: `false - 1`, c: Int(-1), cf: defF},
		{s: `2 * false`, c: Int(0), cf: defF},
		{s: `2 / (true + true)`, c: Int(1), cf: defF},
		{s: `2 / (true + false)`, c: Int(2), cf: defF},
		{s: `false / true`, c: Int(0), cf: defF},
		{s: `1 << (true + 1)`, c: Int(4), cf: defF},
		{s: `true << 2`, c: Int(4), cf: defF},
		{s: `4 >> (1 + true)`, c: Int(1), cf: defF},
		{s: `4 % true`, c: Int(0), cf: defF},
		{s: `true & 2`, c: Int(0), cf: defF},
		{s: `2 & true`, c: Int(0), cf: defF},
		{s: `true | 2`, c: Int(3), cf: defF},
		{s: `2 | true`, c: Int(3), cf: defF},
		{s: `1 ^ (true + true)`, c: Int(3), cf: defF},
		{s: `(true + true) ^ 1`, c: Int(3), cf: defF},
		{s: `(2 * true) &^ 3`, c: Int(0), cf: defF},
		{s: `1 == true * 2`, cf: falseF},
		{s: `true != 2`, cf: trueF},
		{s: `2 != true`, cf: trueF},
		{s: `true < 2`, cf: trueF},
		{s: `true <= 2`, cf: trueF},
		{s: `true > 2`, cf: falseF},
		{s: `true >= 2`, cf: falseF},
		{s: `2 < true`, cf: falseF},
		{s: `2 <= true`, cf: falseF},
		{s: `2 > true`, cf: trueF},
		{s: `2 >= true`, cf: trueF},
		{s: `!false`, cf: trueF},
		{s: `!true`, cf: falseF},
		{s: `-true`, c: Int(-1), cf: defF},
		{s: `+true`, c: Int(1), cf: defF},
		{s: `bool(0)`, cf: falseF},
		{s: `bool(1)`, cf: trueF},

		{s: `"a" + "b"`, c: String("ab"), cf: defF},
		{s: `"a" + 1`, c: String("a1"), cf: defF},
		{s: `"a" + 1u`, c: String("a1"), cf: defF},
		{s: `"a" + 'c'`, c: String("ac"), cf: defF},
		{s: `'c' + "a"`, c: String("ca"), cf: defF},
		{s: `"a" + "b" + "c"`, c: String("abc"), cf: defF},
		{s: `"a" + 'b' + "c"`, c: String("abc"), cf: defF},
		{s: `"a" + 1 + "c"`, c: String("a1c"), cf: defF},
		{s: `char(0)`, c: Char(0), cf: defF},

		{s: `!undefined`, cf: trueF},
		{s: `!!undefined`, cf: falseF},
	}

	for _, tC := range testCases {
		t.Run(tC.s, func(t *testing.T) {
			var consts Array
			if tC.c != nil {
				consts = Array{tC.c}
			}

			f := tC.cf.Copy().(*CompiledFunction)
			f.Instructions = append(f.Instructions, makeInst(OpPop)...)
			f.Instructions = append(f.Instructions, makeInst(OpReturn, 0)...)

			expectEval(t, tC.s, bytecode(consts, f))
		})
	}

	testCases2 := make([]values, len(testCases))

	for i, tC := range testCases {
		testCases2[i].s = "return " + tC.s
		testCases2[i].c = tC.c
		testCases2[i].cf = tC.cf
	}

	for _, tC := range testCases2 {
		t.Run(tC.s, func(t *testing.T) {
			var consts Array
			if tC.c != nil {
				consts = Array{tC.c}
			}

			f := tC.cf.Copy().(*CompiledFunction)
			f.Instructions = append(f.Instructions, makeInst(OpReturn, 1)...)

			expectEval(t, tC.s, bytecode(consts, f))
		})
	}

	testCases3 := make([]values, len(testCases2))

	callF0 := compFunc(concatInsts(
		makeInst(OpConstant, 0),
		makeInst(OpCall, 0, 0),
		makeInst(OpReturn, 1),
	))

	callF1 := compFunc(concatInsts(
		makeInst(OpConstant, 1),
		makeInst(OpCall, 0, 0),
		makeInst(OpReturn, 1),
	))

	for i, tC := range testCases2 {
		testCases3[i].s = fmt.Sprintf(`return func(){ %s }()`, tC.s)
		testCases3[i].c = tC.c
		testCases3[i].cf = tC.cf
	}

	for _, tC := range testCases3 {
		t.Run(tC.s, func(t *testing.T) {
			var consts Array
			var cf *CompiledFunction
			if tC.c != nil {
				consts = append(consts, tC.c)
				cf = callF1
			} else {
				cf = callF0
			}
			sub := tC.cf.Copy().(*CompiledFunction)
			sub.Instructions = append(sub.Instructions, makeInst(OpReturn, 1)...)
			consts = append(consts, sub)

			expectEval(t, tC.s, bytecode(consts, cf))
		})
	}

	testCases4 := make([]values, len(testCases))

	for i, tC := range testCases {
		testCases4[i].s = fmt.Sprintf(`var x = %s`, tC.s)
		testCases4[i].c = tC.c
		testCases4[i].cf = tC.cf
	}

	for _, tC := range testCases4 {
		t.Run(tC.s, func(t *testing.T) {
			var consts Array
			if tC.c != nil {
				consts = append(consts, tC.c)
			}

			f := tC.cf.Copy().(*CompiledFunction)
			f.Instructions = append(f.Instructions, makeInst(OpDefineLocal, 0)...)
			f.Instructions = append(f.Instructions, makeInst(OpReturn, 0)...)
			f.NumLocals = 1
			expectEval(t, tC.s, bytecode(consts, f))
		})
	}

	testCases5 := make([]values, len(testCases))

	for i, tC := range testCases {
		testCases5[i].s = fmt.Sprintf(`x := %s`, tC.s)
		testCases5[i].c = tC.c
		testCases5[i].cf = tC.cf
	}

	for _, tC := range testCases5 {
		t.Run(tC.s, func(t *testing.T) {
			var consts Array
			if tC.c != nil {
				consts = append(consts, tC.c)
			}

			f := tC.cf.Copy().(*CompiledFunction)
			f.Instructions = append(f.Instructions, makeInst(OpDefineLocal, 0)...)
			f.Instructions = append(f.Instructions, makeInst(OpReturn, 0)...)
			f.NumLocals = 1
			expectEval(t, tC.s, bytecode(consts, f))
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
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 1),
				makeInst(OpBinaryOp, int(token.Less)),
				makeInst(OpJumpFalsy, 27),
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 2),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpSetLocal, 0),
				makeInst(OpJump, 5),
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
			Array{},
			compFunc(concatInsts(
				makeInst(OpTrue),
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
				makeInst(OpDefineLocal, 0),
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

func TestOptimizerShadowing(t *testing.T) {
	// int is shadowed by a param declaration, should not evalute int("1") to 1
	expectEval(t, `param int; return int("1")`,
		bytecode(
			Array{String("1")},
			compFunc(concatInsts(
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 0),
				makeInst(OpCall, 1, 0),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withLocals(1),
			),
		))
	// int is shadowed by a var declaration, should not evalute int("1") to 1
	expectEval(t, `var int; return int("1")`,
		bytecode(
			Array{String("1")},
			compFunc(concatInsts(
				makeInst(OpNull),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 0),
				makeInst(OpCall, 1, 0),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),
		))
	// int is shadowed by a var declaration in upper scope,
	// should not evalute int("1") to 1 within function
	expectEval(t, `var int; return func() {return int("1")}`,
		bytecode(
			Array{
				String("1"),
				compFunc(concatInsts(
					makeInst(OpGetFree, 0),
					makeInst(OpConstant, 0),
					makeInst(OpCall, 1, 0),
					makeInst(OpReturn, 1),
				)),
			},
			compFunc(concatInsts(
				makeInst(OpNull),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetLocalPtr, 0),
				makeInst(OpClosure, 1, 1),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),
		))

	opts := DefaultCompilerOptions
	opts.OptimizeConst = true
	opts.OptimizeExpr = true

	st := NewSymbolTable()
	require.NoError(t, st.SetParams("int"))
	opts.SymbolTable = st
	expectCompileWithOpts(t, `return int("1")`, opts,
		bytecode(
			Array{String("1")},
			compFunc(concatInsts(
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 0),
				makeInst(OpCall, 1, 0),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withLocals(1),
			),
		),
	)

	st = NewSymbolTable()
	_, err := st.DefineGlobal("int")
	require.NoError(t, err)
	opts.SymbolTable = st
	expectCompileWithOpts(t, `return int("1")`, opts,
		bytecode(
			Array{String("int"), String("1")},
			compFunc(concatInsts(
				makeInst(OpGetGlobal, 0),
				makeInst(OpConstant, 1),
				makeInst(OpCall, 1, 0),
				makeInst(OpReturn, 1),
			),
			),
		),
	)

	st = NewSymbolTable()
	_, err = st.DefineGlobal("int")
	require.NoError(t, err)
	opts.SymbolTable = st
	expectCompileWithOpts(t, `return func() {return  int("1")}()`, opts,
		bytecode(
			Array{
				String("int"),
				String("1"),
				compFunc(concatInsts(
					makeInst(OpGetGlobal, 0),
					makeInst(OpConstant, 1),
					makeInst(OpCall, 1, 0),
					makeInst(OpReturn, 1),
				)),
			},
			compFunc(concatInsts(
				makeInst(OpConstant, 2),
				makeInst(OpCall, 0, 0),
				makeInst(OpReturn, 1),
			),
			),
		),
	)

	opts.Constants = nil
	opts.SymbolTable = nil
	expectCompileWithOpts(t, `func(int) {return  int("1")}; return int("1")`,
		opts,
		bytecode(
			Array{
				String("1"),
				compFunc(concatInsts(
					makeInst(OpGetLocal, 0),
					makeInst(OpConstant, 0),
					makeInst(OpCall, 1, 0),
					makeInst(OpReturn, 1),
				),
					withParams(1),
					withLocals(1),
				),
				Int(1),
			},
			compFunc(concatInsts(
				makeInst(OpConstant, 1),
				makeInst(OpPop),
				makeInst(OpConstant, 2),
				makeInst(OpReturn, 1),
			),
			),
		),
	)

	// https://github.com/ozanh/ugo/issues/2
	expectRun(t, `
	string := func(x) { return "ok" }
	return string(1)
	`, nil, String("ok"))
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
