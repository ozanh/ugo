// Copyright (c) 2020-2025 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package v1

type Opcode = byte

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

var OpcodeOperands = [...][]int{
	OpNoOp:         {},
	OpConstant:     {2},
	OpCall:         {1, 1},
	OpGetGlobal:    {2},
	OpSetGlobal:    {2},
	OpGetLocal:     {1},
	OpSetLocal:     {1},
	OpGetBuiltin:   {1},
	OpBinaryOp:     {1},
	OpUnary:        {1},
	OpEqual:        {},
	OpNotEqual:     {},
	OpJump:         {2},
	OpJumpFalsy:    {2},
	OpAndJump:      {2},
	OpOrJump:       {2},
	OpMap:          {2},
	OpArray:        {2},
	OpSliceIndex:   {},
	OpGetIndex:     {1},
	OpSetIndex:     {},
	OpNull:         {},
	OpPop:          {},
	OpGetFree:      {1},
	OpSetFree:      {1},
	OpGetLocalPtr:  {1},
	OpGetFreePtr:   {1},
	OpClosure:      {2, 1},
	OpIterInit:     {},
	OpIterNext:     {},
	OpIterKey:      {},
	OpIterValue:    {},
	OpLoadModule:   {2, 2},
	OpStoreModule:  {2},
	OpReturn:       {1},
	OpSetupTry:     {2, 2},
	OpSetupCatch:   {},
	OpSetupFinally: {},
	OpThrow:        {1},
	OpFinalizer:    {1},
	OpDefineLocal:  {1},
	OpTrue:         {},
	OpFalse:        {},
	OpCallName:     {1, 1},
}
