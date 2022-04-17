package encoder_test

// import (
// 	. "github.com/ozanh/ugo"
// )

// func TestBytecode_Encode(t *testing.T) {
// 	testBytecodeSerialization(t,
// 		bytecode(
// 			nil,
// 			compFunc(nil),
// 		),
// 		nil,
// 	)
// 	testBytecodeSerialization(t,
// 		bytecode(
// 			[]Object{
// 				Undefined,
// 				Int(-1), Int(0), Int(1),
// 				Uint(0), ^Uint(0),
// 				Char('x'),
// 				Bool(true), Bool(false),
// 				Float(0), Float(1.2),
// 				String(""), String("abc"),
// 				ErrIndexOutOfBounds,
// 			},
// 			compFunc(
// 				nil,
// 			),
// 		),
// 		nil,
// 	)
// }

// func testBytecodeSerialization(t *testing.T, b *Bytecode, modules *ModuleMap) {
// 	var buf bytes.Buffer
// 	err := b.Encode(&buf)
// 	require.NoError(t, err)

// 	r := &Bytecode{}
// 	err = r.Decode(bytes.NewReader(buf.Bytes()), modules)
// 	require.NoError(t, err)

// 	require.Equal(t, b.FileSet, r.FileSet)
// 	require.Equal(t, b.Main, r.Main)
// 	require.Equalf(t, b.Constants, r.Constants,
// 		"expected:%s\nactual:%s", sdump(b.Constants), sdump(r.Constants))
// 	require.Equal(t, b.NumModules, r.NumModules)
// }
