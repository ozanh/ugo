// Copyright (c) 2020-2025 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package encoder

import (
	"bytes"
	"fmt"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/encoder/opv1"
)

func decodeBytecodeV1(bc *Bytecode, r *bytes.Buffer) error {
	err := decodeBytecodeV2(bc, r)
	if err != nil {
		return err
	}

	err = convBytecodeV1ToV2(bc)
	if err != nil {
		return fmt.Errorf("bytecodeV1Decoder: %w", err)
	}
	return nil
}

func convBytecodeV1ToV2(bc *Bytecode) error {
	opWidth := make([]int, len(opv1.OpcodeOperands))
	for op, operands := range opv1.OpcodeOperands {
		var total int
		for _, operand := range operands {
			total += operand
		}
		opWidth[op] = total
	}

	err := convCompFuncV1ToV2(bc.Main, opWidth)
	if err != nil {
		return fmt.Errorf("unable to convert main function to v2: %w", err)
	}

	for i := range bc.Constants {
		if cf, ok := bc.Constants[i].(*ugo.CompiledFunction); ok {
			if err = convCompFuncV1ToV2(cf, opWidth); err != nil {
				return fmt.Errorf("unable to convert function #%d to v2: %w", i, err)
			}
		}
	}
	return nil
}

func convCompFuncV1ToV2(cf *ugo.CompiledFunction, opWidth []int) error {
	if cf == nil {
		return nil
	}

	var hasJump bool
	for i := 0; !hasJump && i < len(cf.Instructions); {
		op := cf.Instructions[i]

		switch op {
		case
			opv1.OpJump, opv1.OpJumpFalsy, opv1.OpAndJump, opv1.OpOrJump, opv1.OpSetupTry:
			hasJump = true
			continue
		}

		w := opWidth[op]
		i += 1 + w
	}

	if !hasJump {
		return nil
	}

	var newInsts []byte
	newSrcMap := make(map[int]int, len(cf.SourceMap))
	operands := make([]int, 0, 4)
	instBuf := make([]byte, 0, 8)

	for i := 0; i < len(cf.Instructions); {
		op := cf.Instructions[i]
		newInsts = append(newInsts, op)

		if pos, ok := cf.SourceMap[i]; ok {
			newSrcMap[len(newInsts)-1] = pos
		}

		w := opWidth[op]

		switch op {
		case opv1.OpJump, opv1.OpJumpFalsy, opv1.OpAndJump, opv1.OpOrJump, opv1.OpSetupTry:
			operands, _ = ugo.ReadOperands(
				opv1.OpcodeOperands[op],
				cf.Instructions[i+1:],
				operands[:0],
			)

			var err error
			instBuf, err = ugo.MakeInstruction(instBuf[:0], op, operands...)
			if err != nil {
				return fmt.Errorf("unable to make instruction: %w", err)
			}
			// Skip op byte, already added.
			newInsts = append(newInsts, instBuf[1:]...)
		default:
			if w > 0 {
				newInsts = append(newInsts, cf.Instructions[i+1:i+1+w]...)
			}
		}

		i += 1 + w
	}

	cf.Instructions = newInsts
	cf.SourceMap = newSrcMap

	return nil
}
