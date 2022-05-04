package ugo_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ozanh/ugo/tests"
	"github.com/ozanh/ugo/token"

	. "github.com/ozanh/ugo"
)

func makeInst(op Opcode, args ...int) []byte {
	b, err := MakeInstruction(make([]byte, 8), op, args...)
	if err != nil {
		panic(err)
	}
	return b
}

type bytecodeOption func(*Bytecode)

func withModules(numOfModules int) bytecodeOption {
	return func(bc *Bytecode) {
		bc.NumModules = numOfModules
	}
}

func bytecode(
	consts []Object,
	cf *CompiledFunction,
	opts ...bytecodeOption,
) *Bytecode {

	bc := &Bytecode{
		Constants: consts,
		Main:      cf,
	}
	for _, f := range opts {
		f(bc)
	}
	return bc
}

type funcOpt func(*CompiledFunction)

func withParams(numParams int) funcOpt {
	return func(cf *CompiledFunction) {
		cf.NumParams = numParams
	}
}

func withVariadic() funcOpt {
	return func(cf *CompiledFunction) {
		cf.Variadic = true
	}
}

func withLocals(numLocals int) funcOpt {
	return func(cf *CompiledFunction) {
		cf.NumLocals = numLocals
	}
}

func compFunc(insts []byte, opts ...funcOpt) *CompiledFunction {
	cf := &CompiledFunction{
		Instructions: insts,
	}
	for _, f := range opts {
		f(cf)
	}
	return cf
}

func concatInsts(insts ...[]byte) []byte {
	var out []byte
	for i := range insts {
		out = append(out, insts[i]...)
	}
	return out
}

func TestCompiler_Compile(t *testing.T) {
	// all local variables are initialized as undefined
	expectCompile(t, `var a`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))
	expectCompile(t, `var (a, b, c)`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpNull),
			makeInst(OpDefineLocal, 1),
			makeInst(OpNull),
			makeInst(OpDefineLocal, 2),
			makeInst(OpReturn, 0),
		),
			withLocals(3),
		),
	))
	expectCompile(t, `var a = undefined`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))
	expectCompile(t, `a := undefined`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))

	// multiple declaration requires parentheses
	expectCompileError(t, `param a, b`, `Parse Error: expected ';', found ','`)
	expectCompileError(t, `global a, b`, `Parse Error: expected ';', found ','`)
	expectCompileError(t, `var a, b`, `Parse Error: expected ';', found ','`)
	// param declaration can only be at the top scope
	expectCompileError(t, `func() { param a }`, `Compile Error: param not allowed in this scope`)

	// force to set undefined
	expectCompile(t, `a := (undefined)`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))
	expectCompile(t, `var (a, b=1, c=2)`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 1),
			makeInst(OpConstant, 1),
			makeInst(OpDefineLocal, 2),
			makeInst(OpReturn, 0),
		),
			withLocals(3),
		),
	))
	// parameters are initialized as undefined
	expectCompile(t, `param a`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpReturn, 0),
		),
			withParams(1),
			withLocals(1),
		),
	))
	expectCompile(t, `param (a, b, ...c)`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpReturn, 0),
		),
			withParams(3),
			withLocals(3),
			withVariadic(),
		),
	))
	expectCompile(t, `global a`, bytecode(
		Array{String("a")},
		compFunc(concatInsts(
			makeInst(OpReturn, 0),
		)),
	))
	expectCompile(t, `global (a, b); var c`, bytecode(
		Array{String("a"), String("b")},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))
	expectCompile(t, `param (arg1, ...varg); global (a, b); var c = arg1; c = b`, bytecode(
		Array{String("a"), String("b")},
		compFunc(concatInsts(
			makeInst(OpGetLocal, 0),
			makeInst(OpDefineLocal, 2),
			makeInst(OpGetGlobal, 1),
			makeInst(OpSetLocal, 2),
			makeInst(OpReturn, 0),
		),
			withParams(2),
			withLocals(3),
			withVariadic(),
		),
	))

	expectCompile(t, `1 + 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Add)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1; 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpPop),
			makeInst(OpConstant, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1 - 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Sub)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1 * 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Mul)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `2 / 1`, bytecode(
		Array{Int(2), Int(1)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Quo)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `true`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpTrue),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `false`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpFalse),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1 > 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Greater)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1 < 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Less)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1 >= 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.GreaterEq)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1 <= 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.LessEq)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1 == 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpEqual),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `1 != 2`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpNotEqual),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `true == false`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpTrue),
			makeInst(OpFalse),
			makeInst(OpEqual),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `true != false`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpTrue),
			makeInst(OpFalse),
			makeInst(OpNotEqual),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `-1`, bytecode(
		Array{Int(1)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpUnary, int(token.Sub)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `!true`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpTrue),
			makeInst(OpUnary, int(token.Not)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))
	// `if true` => skips else
	expectCompile(t, `if true { 10 }; 3333`, bytecode(
		Array{Int(10), Int(3333)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpPop),
			makeInst(OpConstant, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	// `if (true)` => normal if
	expectCompile(t, `if (true) { 10 }; 3333`, bytecode(
		Array{Int(10), Int(3333)},
		compFunc(concatInsts(
			makeInst(OpTrue),         // 0000
			makeInst(OpJumpFalsy, 8), // 0001
			makeInst(OpConstant, 0),  // 0004
			makeInst(OpPop),          // 0007
			makeInst(OpConstant, 1),  // 0008
			makeInst(OpPop),          // 0011
			makeInst(OpReturn, 0),    // 0012
		)),
	))

	// `if true` => skips else
	expectCompile(t, `if true { 10 } else { 20 }; 3333;`, bytecode(
		Array{Int(10), Int(3333)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpPop),
			makeInst(OpConstant, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	// `if true` => skips else
	expectCompile(t, `if true { 10 } else {}; 3333;`, bytecode(
		Array{Int(10), Int(3333)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpPop),
			makeInst(OpConstant, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	// `if true` => no jumps
	expectCompile(t, `if true { 10 }; 3333;`, bytecode(
		Array{Int(10), Int(3333)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpPop),
			makeInst(OpConstant, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	// `if false` => skip if block but OpJump is put
	// TODO: improve this, unnecessary jump
	expectCompile(t, `if false { 10 }; 3333;`, bytecode(
		Array{Int(3333)},
		compFunc(concatInsts(
			makeInst(OpJump, 3),     // 0000
			makeInst(OpConstant, 0), // 0003
			makeInst(OpPop),         // 0006
			makeInst(OpReturn, 0),   // 0007
		)),
	))

	// `if false` => goes to else block
	// TODO: improve this, unnecessary jump
	expectCompile(t, `if false { 10 } else { 20 }; 3333;`, bytecode(
		Array{Int(20), Int(3333)},
		compFunc(concatInsts(
			makeInst(OpJump, 6),     // 0000
			makeInst(OpJump, 10),    // 0003
			makeInst(OpConstant, 0), // 0006
			makeInst(OpPop),         // 0009
			makeInst(OpConstant, 1), // 0010
			makeInst(OpPop),         // 0013
			makeInst(OpReturn, 0),   // 0014
		)),
	))

	// `if (true)` => normal if
	expectCompile(t, `if (true) { 10 } else { 20 }; 3333;`, bytecode(
		Array{Int(10), Int(20), Int(3333)},
		compFunc(concatInsts(
			makeInst(OpTrue),          // 0000
			makeInst(OpJumpFalsy, 11), // 0001
			makeInst(OpConstant, 0),   // 0004
			makeInst(OpPop),           // 0007
			makeInst(OpJump, 15),      // 0008
			makeInst(OpConstant, 1),   // 0011
			makeInst(OpPop),           // 0014
			makeInst(OpConstant, 2),   // 0015
			makeInst(OpPop),           // 0018
			makeInst(OpReturn, 0),     // 0019
		)),
	))

	expectCompile(t, `"string"`, bytecode(
		Array{String("string")},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `"str" + "ing"`, bytecode(
		Array{String("str"), String("ing")},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Add)),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `a := 1; b := 2; a += b`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpConstant, 1),
			makeInst(OpDefineLocal, 1),
			makeInst(OpGetLocal, 0),
			makeInst(OpGetLocal, 1),
			makeInst(OpBinaryOp, int(token.Add)),
			makeInst(OpSetLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(2),
		)))

	expectCompile(t, `var (a = 1, b = 2); a += b`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpConstant, 1),
			makeInst(OpDefineLocal, 1),
			makeInst(OpGetLocal, 0),
			makeInst(OpGetLocal, 1),
			makeInst(OpBinaryOp, int(token.Add)),
			makeInst(OpSetLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(2),
		)))

	expectCompile(t, `var (a, b = 1); a = b + 1`, bytecode(
		Array{Int(1)},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 1),
			makeInst(OpGetLocal, 1),
			makeInst(OpConstant, 0),
			makeInst(OpBinaryOp, int(token.Add)),
			makeInst(OpSetLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(2),
		)))

	expectCompile(t, `var (a, b)`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpNull),
			makeInst(OpDefineLocal, 0),
			makeInst(OpNull),
			makeInst(OpDefineLocal, 1),
			makeInst(OpReturn, 0),
		),
			withLocals(2),
		)))

	expectCompile(t, `a := 1; b := 2; a /= b`, bytecode(
		Array{Int(1), Int(2)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpConstant, 1),
			makeInst(OpDefineLocal, 1),
			makeInst(OpGetLocal, 0),
			makeInst(OpGetLocal, 1),
			makeInst(OpBinaryOp, int(token.Quo)),
			makeInst(OpSetLocal, 0),
			makeInst(OpReturn, 0),
		),
			withLocals(2),
		)))

	expectCompile(t, `[]`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpArray, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `[1, 2, 3]`, bytecode(
		Array{Int(1), Int(2), Int(3)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpArray, 3),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `[1 + 2, 3 - 4, 5 * 6]`, bytecode(
		Array{Int(1), Int(2), Int(3), Int(4), Int(5), Int(6)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Add)),
			makeInst(OpConstant, 2),
			makeInst(OpConstant, 3),
			makeInst(OpBinaryOp, int(token.Sub)),
			makeInst(OpConstant, 4),
			makeInst(OpConstant, 5),
			makeInst(OpBinaryOp, int(token.Mul)),
			makeInst(OpArray, 3),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `{}`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpMap, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `{a: 2, b: 4, c: 6}`, bytecode(
		Array{String("a"), Int(2), String("b"), Int(4), String("c"), Int(6)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpConstant, 3),
			makeInst(OpConstant, 4),
			makeInst(OpConstant, 5),
			makeInst(OpMap, 6),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `{a: 2 + 3, b: 5 * 6}`, bytecode(
		Array{String("a"), Int(2), Int(3), String("b"), Int(5), Int(6)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpBinaryOp, int(token.Add)),
			makeInst(OpConstant, 3),
			makeInst(OpConstant, 4),
			makeInst(OpConstant, 5),
			makeInst(OpBinaryOp, int(token.Mul)),
			makeInst(OpMap, 4),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `[1, 2, 3][1 + 1]`, bytecode(
		Array{Int(1), Int(2), Int(3)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpArray, 3),
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 0),
			makeInst(OpBinaryOp, int(token.Add)),
			makeInst(OpGetIndex, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `{a: 2}[2 - 1]`, bytecode(
		Array{String("a"), Int(2), Int(1)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpMap, 2),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpBinaryOp, int(token.Sub)),
			makeInst(OpGetIndex, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `[1, 2, 3][:]`, bytecode(
		Array{Int(1), Int(2), Int(3)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpArray, 3),
			makeInst(OpNull),
			makeInst(OpNull),
			makeInst(OpSliceIndex),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `[1, 2, 3][0 : 2]`, bytecode(
		Array{Int(1), Int(2), Int(3), Int(0)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpArray, 3),
			makeInst(OpConstant, 3),
			makeInst(OpConstant, 1),
			makeInst(OpSliceIndex),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `[1, 2, 3][ : 2]`, bytecode(
		Array{Int(1), Int(2), Int(3)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpArray, 3),
			makeInst(OpNull),
			makeInst(OpConstant, 1),
			makeInst(OpSliceIndex),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `[1, 2, 3][0 : ]`, bytecode(
		Array{Int(1), Int(2), Int(3), Int(0)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpArray, 3),
			makeInst(OpConstant, 3),
			makeInst(OpNull),
			makeInst(OpSliceIndex),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `f1 := func(a) { return a }; f1(...[1, 2]);`, bytecode(
		Array{
			compFunc(concatInsts(
				makeInst(OpGetLocal, 0),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withLocals(1),
			),
			Int(1),
			Int(2),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpArray, 2),
			makeInst(OpCall, 1, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0)),
			withLocals(1),
		),
	))

	expectCompile(t, `func() { return 5 + 10 }`, bytecode(
		Array{
			Int(5),
			Int(10),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpConstant, 1),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpReturn, 1),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 2),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { 5 + 10 }`, bytecode(
		Array{
			Int(5),
			Int(10),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpConstant, 1),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 2),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { 1; 2 }`, bytecode(
		Array{
			Int(1),
			Int(2),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpPop),
				makeInst(OpConstant, 1),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 2),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { 1; return 2 }`, bytecode(
		Array{
			Int(1),
			Int(2),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpPop),
				makeInst(OpConstant, 1),
				makeInst(OpReturn, 1),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 2),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { if(true) { return 1 } else { return 2 } }`, bytecode(
		Array{
			Int(1),
			Int(2),
			compFunc(concatInsts(
				makeInst(OpTrue),          // 0000
				makeInst(OpJumpFalsy, 12), // 0001
				makeInst(OpConstant, 0),   // 0004
				makeInst(OpReturn, 1),     // 0007
				makeInst(OpJump, 17),      // 0009
				makeInst(OpConstant, 1),   // 0012
				makeInst(OpReturn, 1),     // 0015
				makeInst(OpReturn, 0),     // 0017
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 2),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { 1; if(true) { 2 } else { 3 }; 4 }`, bytecode(
		Array{
			Int(1),
			Int(2),
			Int(3),
			Int(4),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),   // 0000
				makeInst(OpPop),           // 0003
				makeInst(OpTrue),          // 0004
				makeInst(OpJumpFalsy, 15), // 0005
				makeInst(OpConstant, 1),   // 0008
				makeInst(OpPop),           // 0011
				makeInst(OpJump, 19),      // 0012
				makeInst(OpConstant, 2),   // 0015
				makeInst(OpPop),           // 0018
				makeInst(OpConstant, 3),   // 0019
				makeInst(OpPop),           // 0022
				makeInst(OpReturn, 0),     // 0023
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 4),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { }`, bytecode(
		Array{
			compFunc(concatInsts(
				makeInst(OpReturn, 0),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { 24 }()`, bytecode(
		Array{
			Int(24),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 1),
			makeInst(OpCall, 0, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { return 24 }()`, bytecode(
		Array{
			Int(24),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpReturn, 1),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 1),
			makeInst(OpCall, 0, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `f := func() { 24 }; f();`, bytecode(
		Array{
			Int(24),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 1),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpCall, 0, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))

	expectCompile(t, `f := func() { return 24 }; f();`, bytecode(
		Array{
			Int(24),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpReturn, 1),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 1),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpCall, 0, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))

	expectCompile(t, `n := 55; func() { n };`, bytecode(
		Array{
			Int(55),
			compFunc(concatInsts(
				makeInst(OpGetFree, 0),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocalPtr, 0),
			makeInst(OpClosure, 1, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))

	expectCompile(t, `func() { n := 55; return n }`, bytecode(
		Array{
			Int(55),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { a := 55; b := 77; return a + b }`, bytecode(
		Array{
			Int(55),
			Int(77),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpDefineLocal, 0),
				makeInst(OpConstant, 1),
				makeInst(OpDefineLocal, 1),
				makeInst(OpGetLocal, 0),
				makeInst(OpGetLocal, 1),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpReturn, 1),
			),
				withLocals(2),
			),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 2),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `f := func(a) { return a }; f(24);`, bytecode(
		Array{
			compFunc(concatInsts(
				makeInst(OpGetLocal, 0),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withLocals(1),
			),
			Int(24),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpConstant, 1),
			makeInst(OpCall, 1, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))

	expectCompile(t, `f := func(...a) { return a }; f(1, 2, 3);`, bytecode(
		Array{
			compFunc(concatInsts(
				makeInst(OpGetLocal, 0),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withVariadic(),
				withLocals(1),
			),
			Int(1),
			Int(2),
			Int(3),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpConstant, 3),
			makeInst(OpCall, 3, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))

	expectCompile(t, `f := func(a, b, c) { a; b; return c; }; f(24, 25, 26);`, bytecode(
		Array{
			compFunc(concatInsts(
				makeInst(OpGetLocal, 0),
				makeInst(OpPop),
				makeInst(OpGetLocal, 1),
				makeInst(OpPop),
				makeInst(OpGetLocal, 2),
				makeInst(OpReturn, 1),
			),
				withParams(3),
				withLocals(3),
			),
			Int(24),
			Int(25),
			Int(26),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpConstant, 1),
			makeInst(OpConstant, 2),
			makeInst(OpConstant, 3),
			makeInst(OpCall, 3, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))

	expectCompile(t, `func() { n := 55; n = 23; return n }`, bytecode(
		Array{
			Int(55),
			Int(23),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpDefineLocal, 0),
				makeInst(OpConstant, 1),
				makeInst(OpSetLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 2),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `len([]);`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpGetBuiltin, int(BuiltinLen)),
			makeInst(OpArray, 0),
			makeInst(OpCall, 1, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func() { return len([]) }`, bytecode(
		Array{
			compFunc(concatInsts(
				makeInst(OpGetBuiltin, int(BuiltinLen)),
				makeInst(OpArray, 0),
				makeInst(OpCall, 1, 0),
				makeInst(OpReturn, 1),
			)),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func(a) { func(b) { return a + b } }`, bytecode(
		Array{
			compFunc(concatInsts(
				makeInst(OpGetFree, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withLocals(1),
			),

			compFunc(concatInsts(
				makeInst(OpGetLocalPtr, 0),
				makeInst(OpClosure, 0, 1),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			),
				withParams(1),
				withLocals(1),
			),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `func(a) {
		return func(b) {
			return func(c) {
				return a + b + c
			}
		}
	}`, bytecode(
		Array{
			compFunc(concatInsts(
				makeInst(OpGetFree, 0),
				makeInst(OpGetFree, 1),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpGetLocal, 0),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withLocals(1),
			),

			compFunc(concatInsts(
				makeInst(OpGetFreePtr, 0),
				makeInst(OpGetLocalPtr, 0),
				makeInst(OpClosure, 0, 2),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withLocals(1),
			),

			compFunc(concatInsts(
				makeInst(OpGetLocalPtr, 0),
				makeInst(OpClosure, 1, 1),
				makeInst(OpReturn, 1),
			),
				withParams(1),
				withLocals(1),
			),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 2),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))

	expectCompile(t, `
	g := 55;
	func() {
		a := 66;

		return func() {
			b := 77;

			return func() {
				c := 88;

				return g + a + b + c;
			}
		}
	}`, bytecode(
		Array{
			Int(55),
			Int(66),
			Int(77),
			Int(88),
			compFunc(concatInsts(
				makeInst(OpConstant, 3),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetFree, 0),
				makeInst(OpGetFree, 1),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpGetFree, 2),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpGetLocal, 0),
				makeInst(OpBinaryOp, int(token.Add)),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),

			compFunc(concatInsts(
				makeInst(OpConstant, 2),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetFreePtr, 0),
				makeInst(OpGetFreePtr, 1),
				makeInst(OpGetLocalPtr, 0),
				makeInst(OpClosure, 4, 3),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),

			compFunc(concatInsts(
				makeInst(OpConstant, 1),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetFreePtr, 0),
				makeInst(OpGetLocalPtr, 0),
				makeInst(OpClosure, 5, 2),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocalPtr, 0),
			makeInst(OpClosure, 6, 1),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		),
			withLocals(1),
		),
	))

	// Block variables not used as free variable is set to undefined after loop.
	// If block variable is not used as free variable it is reused.
	expectCompile(t, `for i:=0; i<10; i++ {}; j := 1`, bytecode(
		Array{Int(0), Int(10), Int(1)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),               // 0000
			makeInst(OpDefineLocal, 0),            // 0003
			makeInst(OpGetLocal, 0),               // 0005
			makeInst(OpConstant, 1),               // 0007
			makeInst(OpBinaryOp, int(token.Less)), // 0010
			makeInst(OpJumpFalsy, 27),             // 0012
			makeInst(OpGetLocal, 0),               // 0015
			makeInst(OpConstant, 2),               // 0017
			makeInst(OpBinaryOp, int(token.Add)),  // 0020
			makeInst(OpSetLocal, 0),               // 0022
			makeInst(OpJump, 5),                   // 0024
			makeInst(OpConstant, 2),               // 0027
			makeInst(OpDefineLocal, 0),            // 0030
			makeInst(OpReturn, 0),                 // 0032
		),
			withLocals(1),
		),
	))

	expectCompile(t, `m := {}; for k, v in m { }`, bytecode(
		Array{},
		compFunc(concatInsts(
			makeInst(OpMap, 0),         // 0000
			makeInst(OpDefineLocal, 0), // 0003
			makeInst(OpGetLocal, 0),    // 0005
			makeInst(OpIterInit),       // 0007
			makeInst(OpDefineLocal, 1), // 0008 :it
			makeInst(OpGetLocal, 1),    // 0010 :it
			makeInst(OpIterNext),       // 0012
			makeInst(OpJumpFalsy, 29),  // 0013
			makeInst(OpGetLocal, 1),    // 0016
			makeInst(OpIterKey),        // 0018
			makeInst(OpDefineLocal, 2), // 0019 k
			makeInst(OpGetLocal, 1),    // 0021 :it
			makeInst(OpIterValue),      // 0023
			makeInst(OpDefineLocal, 3), // 0024 v
			makeInst(OpJump, 10),       // 0026
			makeInst(OpReturn, 0),      // 0029
		),
			withLocals(4), // m, :it, k, v
		),
	))

	expectCompile(t, `a := 0; a == 0 && a != 1 || a < 1`, bytecode(
		Array{Int(0), Int(1)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),               // 0000
			makeInst(OpDefineLocal, 0),            // 0003
			makeInst(OpGetLocal, 0),               // 0005
			makeInst(OpConstant, 0),               // 0007
			makeInst(OpEqual),                     // 0010
			makeInst(OpAndJump, 20),               // 0011
			makeInst(OpGetLocal, 0),               // 0014
			makeInst(OpConstant, 1),               // 0016
			makeInst(OpNotEqual),                  // 0019
			makeInst(OpOrJump, 30),                // 0020
			makeInst(OpGetLocal, 0),               // 0023
			makeInst(OpConstant, 1),               // 0025
			makeInst(OpBinaryOp, int(token.Less)), // 0028
			makeInst(OpPop),                       // 0030
			makeInst(OpReturn, 0),                 // 0031
		),
			withLocals(1),
		),
	))

	expectCompile(t, `try { a:=0 } catch err { } finally { err; a; }; x:=1`, bytecode(
		Array{Int(0), Int(1)},
		compFunc(concatInsts(
			makeInst(OpSetupTry, 16, 19), // 0000 // catch and finally positions
			makeInst(OpConstant, 0),      // 0005
			makeInst(OpDefineLocal, 0),   // 0008 a
			makeInst(OpNull),             // 0010
			makeInst(OpDefineLocal, 1),   // 0011 err
			makeInst(OpJump, 19),         // 0013 // jump to finally if no error
			makeInst(OpSetupCatch),       // 0016
			makeInst(OpSetLocal, 1),      // 0017
			makeInst(OpSetupFinally),     // 0019
			makeInst(OpGetLocal, 1),      // 0020
			makeInst(OpPop),              // 0022
			makeInst(OpGetLocal, 0),      // 0023
			makeInst(OpPop),              // 0025
			makeInst(OpThrow, 0),         // 0026
			makeInst(OpConstant, 1),      // 0028
			makeInst(OpDefineLocal, 0),   // 0031 x
			makeInst(OpReturn, 0),        // 0033
		),
			withLocals(2),
		),
	))

	expectCompile(t, `try { a:=0 } catch err { }`, bytecode(
		Array{Int(0)},
		compFunc(concatInsts(
			makeInst(OpSetupTry, 16, 19), // 0000
			makeInst(OpConstant, 0),      // 0005
			makeInst(OpDefineLocal, 0),   // 0008 a
			makeInst(OpNull),             // 0010
			makeInst(OpDefineLocal, 1),   // 0011 err
			makeInst(OpJump, 19),         // 0010
			makeInst(OpSetupCatch),       // 0013
			makeInst(OpSetLocal, 1),      // 0014
			makeInst(OpSetupFinally),     // 0016 always OpSetupFinally
			makeInst(OpThrow, 0),         // 0023
			makeInst(OpReturn, 0),        // 0025
		),
			withLocals(2),
		),
	))

	expectCompile(t, `try { a:=0; throw "an error" } catch { }`, bytecode(
		Array{Int(0), String("an error")},
		compFunc(concatInsts(
			makeInst(OpSetupTry, 18, 20), // 0000
			makeInst(OpConstant, 0),      // 0005
			makeInst(OpDefineLocal, 0),   // 0008 a
			makeInst(OpConstant, 1),      // 0010
			makeInst(OpThrow, 1),         // 0013
			makeInst(OpJump, 20),         // 0015
			makeInst(OpSetupCatch),       // 0018
			makeInst(OpPop),              // 0019
			makeInst(OpSetupFinally),     // 0020
			makeInst(OpThrow, 0),         // 0021
			makeInst(OpReturn, 0),        // 0023
		),
			withLocals(1),
		),
	))
	expectCompileError(t, `try {}`, `Parse Error: expected 'finally', found newline`)
	expectCompileError(t, `catch {}`, `Parse Error: expected statement, found 'catch'`)
	expectCompileError(t, `finally {}`, `Parse Error: expected statement, found 'finally'`)
	// catch and finally must in the same line with right brace.
	expectCompileError(t, `try {}
	catch {}`, `Parse Error: expected 'finally', found newline`)
	expectCompileError(t, `try {
	} catch {}
	finally {}`, `Parse Error: expected statement, found 'finally'`)

	// 4 instructions are generated for every source module import.
	// If module's returned value is already stored, ignore storing.
	moduleMap := NewModuleMap()
	moduleMap.AddSourceModule("mod", []byte(``))
	expectCompileWithOpts(t, `import("mod")`,
		CompilerOptions{
			ModuleMap: moduleMap,
		},
		bytecode(
			Array{
				compFunc(concatInsts(
					makeInst(OpReturn, 0),
				)),
			},
			compFunc(concatInsts(
				makeInst(OpLoadModule, 0, 0), // 0000 constant, module indexes
				makeInst(OpJumpFalsy, 14),    // 0005 if loaded no call is required
				makeInst(OpCall, 0, 0),       // 0008 obtain return value from module
				makeInst(OpStoreModule, 0),   // 0011 store returned value to module cache
				makeInst(OpPop),              // 0014
				makeInst(OpReturn, 0),        // 0015
			)),
			withModules(1),
		),
	)

	// 3 instructions are generated for non-source module import.
	// If module's value is already stored, ignore storing.
	moduleMap = NewModuleMap()
	moduleMap.AddBuiltinModule("mod", Map{})
	expectCompileWithOpts(t, `import("mod")`,
		CompilerOptions{
			ModuleMap: moduleMap,
		},
		bytecode(
			Array{
				Map{AttrModuleName: String("mod")},
			},
			compFunc(concatInsts(
				makeInst(OpLoadModule, 0, 0), // 0000 constant, module indexes
				makeInst(OpJumpFalsy, 11),    // 0005 if loaded no call is required
				makeInst(OpStoreModule, 0),   // 0008 store value to module cache
				makeInst(OpPop),              // 0011
				makeInst(OpReturn, 0),        // 0012
			)),
			withModules(1),
		),
	)

	// unknown module name
	expectCompileError(t, `import("user1")`, "Compile Error: module 'user1' not found")
	expectCompileError(t, `import("")`, "Compile Error: empty module name")
	// too many errors
	expectCompileError(t, `
	r["x"] = {
		@a:1,
		@b:1,
		@c:1,
		@d:1,
		@e:1,
		@f:1,
		@g:1,
		@h:1,
		@i:1,
		@j:1,
		@k:1
	}
	`, "Parse Error: illegal character U+0040 '@'\n\tat (main):3:3 (and 10 more errors)")
	expectCompileError(t, `
	(func() {
		fn := fn()
	})()	
	`, `Compile Error: unresolved reference "fn"`)

	expectCompile(t, `x, y := []`,
		bytecode(
			// 2: number of LHS idents
			// 0: array index to assign to x
			// 1: array index to assign to y
			Array{Int(2), Int(0), Int(1)},
			compFunc(concatInsts(
				makeInst(OpGetBuiltin,
					int(BuiltinMakeArray)), // load builtin to call
				makeInst(OpConstant, 0),    // load lhs length
				makeInst(OpArray, 0),       // rhs empty array
				makeInst(OpCall, 2, 0),     // call builtin :makeArray(2, [])
				makeInst(OpDefineLocal, 0), // set builtin call result to :array
				makeInst(OpGetLocal, 0),    // load :array
				makeInst(OpConstant, 1),    // load 0 (array index)
				makeInst(OpGetIndex, 1),    // :array[0]
				makeInst(OpDefineLocal, 1), // x = :array[0]
				makeInst(OpGetLocal, 0),    // load :array
				makeInst(OpConstant, 2),    // load 1 (array index)
				makeInst(OpGetIndex, 1),    // :array[1]
				makeInst(OpDefineLocal, 2), // y = :array[1]
				makeInst(OpNull),           // load undefined
				makeInst(OpSetLocal, 0),    // cleanup -> :array = undefined
				makeInst(OpReturn, 0),
			),
				// x,y and :array hidden variable
				withLocals(3),
			),
		),
	)

	expectCompile(t, `func() { return 1, 2 }`,
		bytecode(
			Array{
				Int(1),
				Int(2),
				compFunc(concatInsts(
					makeInst(OpConstant, 0),
					makeInst(OpConstant, 1),
					makeInst(OpArray, 2),
					makeInst(OpReturn, 1),
				),
				),
			},
			compFunc(concatInsts(
				makeInst(OpConstant, 2),
				makeInst(OpPop),
				makeInst(OpReturn, 0),
			),
			),
		),
	)

	expectCompile(t, `var a; return a["b"]`,
		bytecode(
			Array{String("b")},
			compFunc(concatInsts(
				makeInst(OpNull),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 0),
				makeInst(OpGetIndex, 1),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),
		),
	)

	expectCompile(t, `var a; return a["b"]["c"][2]`,
		bytecode(
			Array{String("b"), String("c"), Int(2)},
			compFunc(concatInsts(
				makeInst(OpNull),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpConstant, 0),
				makeInst(OpConstant, 1),
				makeInst(OpConstant, 2),
				makeInst(OpGetIndex, 3),
				makeInst(OpReturn, 1),
			),
				withLocals(1),
			),
		),
	)
}

func TestCompilerScopes(t *testing.T) {
	expectCompile(t, `
	if a := 1; a {
		a = 2
		b := a
	} else {
		a = 3
		b := a
	}`, bytecode(
		Array{Int(1), Int(2), Int(3)},
		compFunc(concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpDefineLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpJumpFalsy, 22),
			makeInst(OpConstant, 1),
			makeInst(OpSetLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpDefineLocal, 1),
			makeInst(OpJump, 31),
			makeInst(OpConstant, 2),
			makeInst(OpSetLocal, 0),
			makeInst(OpGetLocal, 0),
			makeInst(OpDefineLocal, 1),
			makeInst(OpReturn, 0),
		),
			withLocals(2),
		)),
	)

	expectCompile(t, `
	func() {
		if a := 1; a {
			a = 2
			b := a
		} else {
			a = 3
			b := a
		}
	}`, bytecode(
		Array{
			Int(1),
			Int(2),
			Int(3),
			compFunc(concatInsts(
				makeInst(OpConstant, 0),
				makeInst(OpDefineLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpJumpFalsy, 22),
				makeInst(OpConstant, 1),
				makeInst(OpSetLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpDefineLocal, 1),
				makeInst(OpJump, 31),
				makeInst(OpConstant, 2),
				makeInst(OpSetLocal, 0),
				makeInst(OpGetLocal, 0),
				makeInst(OpDefineLocal, 1),
				makeInst(OpReturn, 0),
			),
				withLocals(2),
			),
		},
		compFunc(concatInsts(
			makeInst(OpConstant, 3),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)),
	))
}

func expectCompileError(t *testing.T, script string, errStr string) {
	t.Helper()
	expectCompileErrorWithOpts(t, script, CompilerOptions{}, errStr)
}

func expectCompileErrorWithOpts(t *testing.T,
	script string, opts CompilerOptions, errStr string) {

	t.Helper()
	_, err := Compile([]byte(script), opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), errStr)
}

func expectCompile(t *testing.T, script string, expected *Bytecode) {
	t.Helper()
	expectCompileWithOpts(t, script, CompilerOptions{}, expected)
}

// SourceMap comparison is ignored if it is nil.
func expectCompileWithOpts(t *testing.T,
	script string, opts CompilerOptions, expected *Bytecode) {

	t.Helper()
	got, err := Compile([]byte(script), opts)
	require.NoError(t, err)
	testBytecodesEqual(t, expected, got, expected.Main.SourceMap != nil)
}

func testBytecodesEqual(t *testing.T,
	expected, got *Bytecode, checkSourceMap bool) {

	t.Helper()
	if expected.NumModules != got.NumModules {
		t.Fatalf("NumModules not equal expected %d, got %d\n",
			expected.NumModules, got.NumModules)
	}
	if len(expected.Constants) != len(got.Constants) {
		var buf bytes.Buffer
		got.Fprint(&buf)
		t.Fatalf("Constants not equal\nDump:\n%s\n"+
			"Expected Constants:\n%s\nGot Constants:\n%s\n",
			buf.String(), tests.Sdump(expected.Constants), tests.Sdump(got.Constants))
	}
	if !assertCompiledFunctionsEqual(t,
		expected.Main, got.Main, checkSourceMap) {
		t.Fatal("Main functions not equal")
	}
	for i, gotObj := range got.Constants {
		expectObj := expected.Constants[i]

		switch g := gotObj.(type) {
		case *CompiledFunction:
			ex, ok := expectObj.(*CompiledFunction)
			if !ok {
				t.Fatalf("%T expected at index %d but got %T",
					expectObj, i, gotObj)
			}
			if !assertCompiledFunctionsEqual(t, ex, g, checkSourceMap) {
				t.Fatalf("CompiledFunctions not equal at %d\nExpected:\n"+
					"%s\nGot:\n%s\n", i, ex, g)
			}
			continue
		}
		if !reflect.DeepEqual(expectObj, gotObj) {
			t.Fatalf("Constants not equal at %d\nExpected:\n%s\nGot:\n%s\n",
				i, expectObj, gotObj)
		}
	}
}

func assertCompiledFunctionsEqual(t *testing.T,
	expected, got *CompiledFunction, checkSourceMap bool) bool {
	t.Helper()
	if expected.NumParams != got.NumParams {
		t.Errorf("NumParams not equal expected %d, got %d\n",
			expected.NumParams, got.NumParams)
		return false
	}
	if expected.Variadic != got.Variadic {
		t.Errorf("Variadic not equal expected %t, got %t\n",
			expected.Variadic, got.Variadic)
		return false
	}
	if expected.NumLocals != got.NumLocals {
		t.Errorf("NumLocals not equal expected %d, got %d\n",
			expected.NumLocals, got.NumLocals)
		return false
	}
	if string(expected.Instructions) != string(got.Instructions) {
		var buf bytes.Buffer
		buf.WriteString("Expected:\n")
		expected.Fprint(&buf)
		buf.WriteString("\nGot:\n")
		got.Fprint(&buf)
		t.Fatalf("Instructions not equal\n%s", buf.String())
	}
	if len(expected.Free) != len(got.Free) {
		t.Errorf("Free not equal expected %d, got %d\n",
			len(expected.Free), len(got.Free))
		return false
	}
	if checkSourceMap &&
		!reflect.DeepEqual(got.SourceMap, expected.SourceMap) {
		t.Errorf("sourceMaps not equal\n"+
			"Expected:\n%s\nGot:\n%s\n"+
			"Bytecode dump:\n%s\n",
			tests.Sdump(expected.SourceMap), tests.Sdump(got.SourceMap), got)
		return false
	}
	return true
}
