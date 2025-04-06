package encoder

import "io"

// Exported for testing purposes.

var DecodeBytecodeV1 = decodeBytecodeV1

var WriteBytecodeHeader = writeBytecodeHeader

func EncodeBytecodeV1(bc *Bytecode, w io.Writer) error {
	err := writeBytecodeHeader(w, BytecodeVersion1)
	if err != nil {
		return err
	}
	return encodeBytecodeCommon(bc, w)
}
