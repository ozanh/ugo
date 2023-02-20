// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

// Opcode represents a single byte operation code.
type Opcode = byte

// List of opcodes
const (
	OpNoOp Opcode = iota
	OpConstant
	OpCall
	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpBinaryOp
	OpUnary
	OpEqual
	OpNotEqual
	OpJump
	OpJumpFalsy
	OpAndJump
	OpOrJump
	OpMap
	OpArray
	OpSliceIndex
	OpGetIndex
	OpSetIndex
	OpNull
	OpPop
	OpGetFree
	OpSetFree
	OpGetLocalPtr
	OpGetFreePtr
	OpClosure
	OpIterInit
	OpIterNext
	OpIterKey
	OpIterValue
	OpLoadModule
	OpStoreModule
	OpSetupTry
	OpSetupCatch
	OpSetupFinally
	OpThrow
	OpFinalizer
	OpReturn
	OpDefineLocal
	OpTrue
	OpFalse
	OpCallName
)

// OpcodeNames are string representation of opcodes.
var OpcodeNames = [...]string{
	OpNoOp:         "NOOP",
	OpConstant:     "CONSTANT",
	OpCall:         "CALL",
	OpGetGlobal:    "GETGLOBAL",
	OpSetGlobal:    "SETGLOBAL",
	OpGetLocal:     "GETLOCAL",
	OpSetLocal:     "SETLOCAL",
	OpGetBuiltin:   "GETBUILTIN",
	OpBinaryOp:     "BINARYOP",
	OpUnary:        "UNARY",
	OpEqual:        "EQUAL",
	OpNotEqual:     "NOTEQUAL",
	OpJump:         "JUMP",
	OpJumpFalsy:    "JUMPFALSY",
	OpAndJump:      "ANDJUMP",
	OpOrJump:       "ORJUMP",
	OpMap:          "MAP",
	OpArray:        "ARRAY",
	OpSliceIndex:   "SLICEINDEX",
	OpGetIndex:     "GETINDEX",
	OpSetIndex:     "SETINDEX",
	OpNull:         "NULL",
	OpPop:          "POP",
	OpGetFree:      "GETFREE",
	OpSetFree:      "SETFREE",
	OpGetLocalPtr:  "GETLOCALPTR",
	OpGetFreePtr:   "GETFREEPTR",
	OpClosure:      "CLOSURE",
	OpIterInit:     "ITERINIT",
	OpIterNext:     "ITERNEXT",
	OpIterKey:      "ITERKEY",
	OpIterValue:    "ITERVALUE",
	OpLoadModule:   "LOADMODULE",
	OpStoreModule:  "STOREMODULE",
	OpReturn:       "RETURN",
	OpSetupTry:     "SETUPTRY",
	OpSetupCatch:   "SETUPCATCH",
	OpSetupFinally: "SETUPFINALLY",
	OpThrow:        "THROW",
	OpFinalizer:    "FINALIZER",
	OpDefineLocal:  "DEFINELOCAL",
	OpTrue:         "TRUE",
	OpFalse:        "FALSE",
	OpCallName:     "CALLNAME",
}

// OpcodeOperands is the number of operands.
var OpcodeOperands = [...][]int{
	OpNoOp:         {},
	OpConstant:     {2},    // constant index
	OpCall:         {1, 1}, // number of arguments, flags
	OpGetGlobal:    {2},    // constant index
	OpSetGlobal:    {2},    // constant index
	OpGetLocal:     {1},    // local variable index
	OpSetLocal:     {1},    // local variable index
	OpGetBuiltin:   {1},    // builtin index
	OpBinaryOp:     {1},    // operator
	OpUnary:        {1},    // operator
	OpEqual:        {},
	OpNotEqual:     {},
	OpJump:         {2}, // position
	OpJumpFalsy:    {2}, // position
	OpAndJump:      {2}, // position
	OpOrJump:       {2}, // position
	OpMap:          {2}, // number of keys and values
	OpArray:        {2}, // number of items
	OpSliceIndex:   {},
	OpGetIndex:     {1}, // number of selectors
	OpSetIndex:     {},
	OpNull:         {},
	OpPop:          {},
	OpGetFree:      {1},    // index
	OpSetFree:      {1},    // index
	OpGetLocalPtr:  {1},    // index
	OpGetFreePtr:   {1},    // index
	OpClosure:      {2, 1}, // constant index, item count
	OpIterInit:     {},
	OpIterNext:     {},
	OpIterKey:      {},
	OpIterValue:    {},
	OpLoadModule:   {2, 2}, // constant index, module index
	OpStoreModule:  {2},    // module index
	OpReturn:       {1},    // number of items (0 or 1)
	OpSetupTry:     {2, 2},
	OpSetupCatch:   {},
	OpSetupFinally: {},
	OpThrow:        {1}, // 0:re-throw (system), 1:throw <expression>
	OpFinalizer:    {1}, // up to error handler index
	OpDefineLocal:  {1},
	OpTrue:         {},
	OpFalse:        {},
	OpCallName:     {1, 1}, // number of arguments, flags
}

// ReadOperands reads operands from the bytecode. Given operands slice is used to
// fill operands and is returned to allocate less.
func ReadOperands(numOperands []int, ins []byte, operands []int) ([]int, int) {
	operands = operands[:0]
	var offset int
	for _, width := range numOperands {
		switch width {
		case 1:
			operands = append(operands, int(ins[offset]))
		case 2:
			operands = append(operands, int(ins[offset+1])|int(ins[offset])<<8)
		}
		offset += width
	}
	return operands, offset
}
