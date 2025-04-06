package encoder_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ozanh/ugo"
	v1 "github.com/ozanh/ugo/encoder/v1"

	. "github.com/ozanh/ugo/encoder"
)

func Test_bytecode_v1(t *testing.T) {
	const headerV1Size = 6

	t.Run("no jumps", func(t *testing.T) {
		bcv1 := &Bytecode{
			Main: compFunc(
				concatInsts(
					makeInst(v1.OpNoOp),
					makeInst(v1.OpConstant, 0),
				),
				withParams(1),
				withLocals(1),
				withSourceMap(map[int]int{
					0: 0,
					1: 1,
				}),
				withVariadic(),
			),
			NumModules: 1,
		}

		buf := bytes.NewBuffer(nil)
		err := EncodeBytecodeV1(bcv1, buf)
		require.NoError(t, err)

		// Skip the header
		buf = bytes.NewBuffer(buf.Bytes()[headerV1Size:])

		var decoded Bytecode
		err = DecodeBytecodeV1(&decoded, buf)
		require.NoError(t, err)

		require.Equal(t, bcv1.Constants, decoded.Constants)
		require.Equal(t, bcv1.NumModules, decoded.NumModules)
		require.Equal(t, bcv1.FileSet, decoded.FileSet)

		require.Equal(t, bcv1.Main.NumParams, decoded.Main.NumParams)
		require.Equal(t, bcv1.Main.NumLocals, decoded.Main.NumLocals)
		require.Equal(t, bcv1.Main.Variadic, decoded.Main.Variadic)
		require.Equal(t, bcv1.Main.SourceMap, decoded.Main.SourceMap)
		require.Equal(t, bcv1.Main.Instructions, decoded.Main.Instructions)
	})

	t.Run("with jumps", func(t *testing.T) {

		bcv1 := &Bytecode{
			Main: compFunc(
				concatInsts(
					makeInst(v1.OpNoOp),
					makeInst(v1.OpConstant, 0),

					binary.BigEndian.AppendUint16(
						[]byte{v1.OpJump},
						1,
					),
					binary.BigEndian.AppendUint16(
						[]byte{v1.OpJumpFalsy},
						2,
					),
					binary.BigEndian.AppendUint16(
						[]byte{v1.OpAndJump},
						3,
					),
					binary.BigEndian.AppendUint16(
						[]byte{v1.OpOrJump},
						4,
					),
					binary.BigEndian.AppendUint16(
						binary.BigEndian.AppendUint16(
							[]byte{v1.OpSetupTry},
							5,
						),
						6,
					),
				),
				withParams(1),
				withLocals(1),
				withSourceMap(map[int]int{
					0: 0,
					1: 1,
					7: 2,
				}),
				withVariadic(),
			),
			NumModules: 1,
		}

		compFunc := (*bcv1.Main)
		bcv1.Constants = append(bcv1.Constants, &compFunc)

		buf := bytes.NewBuffer(nil)
		err := EncodeBytecodeV1(bcv1, buf)
		require.NoError(t, err)

		// Skip the header
		buf = bytes.NewBuffer(buf.Bytes()[headerV1Size:])

		var decoded Bytecode
		err = DecodeBytecodeV1(&decoded, buf)
		require.NoError(t, err)

		//(*ugo.Bytecode)(&decoded).Fprint(os.Stdout)

		require.Equal(t, len(bcv1.Constants), len(decoded.Constants))
		require.Equal(t, bcv1.NumModules, decoded.NumModules)
		require.Equal(t, bcv1.FileSet, decoded.FileSet)

		for _, cf := range []*ugo.CompiledFunction{
			decoded.Main,
			decoded.Constants[0].(*ugo.CompiledFunction),
		} {

			require.Equal(t, bcv1.Main.NumParams, cf.NumParams)
			require.Equal(t, bcv1.Main.NumLocals, cf.NumLocals)
			require.Equal(t, bcv1.Main.Variadic, cf.Variadic)

			require.NotEqual(t, bcv1.Main.SourceMap, cf.SourceMap)
			require.NotEqual(t, bcv1.Main.Instructions, cf.Instructions)

			ugo.IterateInstructions(
				cf.Instructions,
				func(pos int, op ugo.Opcode, operands []int, offset int) bool {
					switch op {
					case v1.OpJump:
						require.Equal(t, 4, offset)
						require.Equal(t, 1, len(operands))
						require.Equal(t, 1, operands[0])
					case v1.OpJumpFalsy:
						require.Equal(t, 4, offset)
						require.Equal(t, 1, len(operands))
						require.Equal(t, 2, operands[0])
					case v1.OpAndJump:
						require.Equal(t, 4, offset)
						require.Equal(t, 1, len(operands))
						require.Equal(t, 3, operands[0])
					case v1.OpOrJump:
						require.Equal(t, 4, offset)
						require.Equal(t, 1, len(operands))
						require.Equal(t, 4, operands[0])
					case v1.OpSetupTry:
						require.Equal(t, 8, offset)
						require.Equal(t, 2, len(operands))
						require.Equal(t, 5, operands[0])
						require.Equal(t, 6, operands[1])
					}
					return true
				})
		}
	})
}
