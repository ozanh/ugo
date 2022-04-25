package encoder_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
	gotime "time"

	"github.com/stretchr/testify/require"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/stdlib/fmt"
	"github.com/ozanh/ugo/stdlib/json"
	"github.com/ozanh/ugo/stdlib/strings"
	"github.com/ozanh/ugo/stdlib/time"
	"github.com/ozanh/ugo/tests"

	. "github.com/ozanh/ugo/encoder"
)

var baz ugo.Object = ugo.String("baz")
var testObjects = []ugo.Object{
	ugo.Undefined,
	ugo.Int(-1), ugo.Int(0), ugo.Int(1),
	ugo.Uint(0), ^ugo.Uint(0),
	ugo.Char('x'),
	ugo.Bool(true), ugo.Bool(false),
	ugo.Float(0), ugo.Float(1.2),
	ugo.String(""), ugo.String("abc"),
	ugo.Bytes{}, ugo.Bytes("foo"),
	ugo.ErrIndexOutOfBounds,
	&ugo.RuntimeError{Err: ugo.ErrInvalidIndex},
	ugo.Map{"key": &ugo.Function{Name: "f"}},
	&ugo.SyncMap{Value: ugo.Map{"k": ugo.String("")}},
	ugo.Array{ugo.Undefined, ugo.True, ugo.False},
	&time.Time{Value: gotime.Time{}},
	&json.EncoderOptions{Value: ugo.Int(1)},
	&json.RawMessage{Value: ugo.Bytes("bar")},
	&ugo.ObjectPtr{Value: &baz},
}

func TestBytecode_Encode(t *testing.T) {
	testBytecodeSerialization(t, &ugo.Bytecode{Main: compFunc(nil)}, nil)

	testBytecodeSerialization(t,
		&ugo.Bytecode{Constants: testObjects,
			Main: compFunc(
				[]byte("test instructions"),
				withLocals(1), withParams(1), withVariadic(),
			),
		},
		nil,
	)
}

func TestBytecode_file(t *testing.T) {
	temp := t.TempDir()

	bc := &ugo.Bytecode{Constants: testObjects,
		Main: compFunc(
			[]byte("test instructions"),
			withLocals(4), withParams(0), withVariadic(),
			withSourceMap(map[int]int{0: 1, 1: 2}),
		),
	}
	f, err := ioutil.TempFile(temp, "mod.ugoc")
	require.NoError(t, err)
	defer f.Close()

	err = EncodeBytecodeTo(bc, f)
	require.NoError(t, err)

	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)

	got, err := DecodeBytecodeFrom(f, nil)
	require.NoError(t, err)
	testBytecodesEqual(t, bc, got)
}

func TestBytecode_full(t *testing.T) {
	src := `
fmt := import("fmt")
strings := import("strings")
time := import("time")
json := import("json")
srcmod := import("srcmod")

v := int(json.Unmarshal(json.Marshal(1)))
v = int(strings.Join([v], ""))
v = srcmod.Incr(v)
v = srcmod.Decr(v)
v = int(fmt.Sprintf("%d", v))
return v*time.Second/time.Second // 1
`

	opts := ugo.DefaultCompilerOptions
	opts.ModuleMap = ugo.NewModuleMap().
		AddBuiltinModule("fmt", fmt.Module).
		AddBuiltinModule("strings", strings.Module).
		AddBuiltinModule("time", time.Module).
		AddBuiltinModule("json", json.Module).
		AddSourceModule("srcmod", []byte(`
return {
	Incr: func(x) { return x + 1 },
	Decr: func(x) { return x - 1 },
}
		`))

	mmCopy := opts.ModuleMap.Copy()

	bc, err := ugo.Compile([]byte(src), opts)
	require.NoError(t, err)

	wantRet, err := ugo.NewVM(bc).Run(nil)
	require.NoError(t, err)
	require.Equal(t, ugo.Int(1), wantRet)

	temp := t.TempDir()
	f, err := ioutil.TempFile(temp, "program.ugoc")
	require.NoError(t, err)
	defer f.Close()

	var buf bytes.Buffer

	logmicros(t, "encode time: %d microsecs", func() {
		err = EncodeBytecodeTo(bc, &buf)
	})
	require.NoError(t, err)

	t.Logf("written size: %v bytes", buf.Len())

	_, err = buf.WriteTo(f)
	require.NoError(t, err)

	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)

	var gotBc *ugo.Bytecode
	logmicros(t, "decode time: %d microsecs", func() {
		gotBc, err = DecodeBytecodeFrom(f, mmCopy)
	})
	require.NoError(t, err)
	require.NotNil(t, gotBc)

	var gotRet ugo.Object
	logmicros(t, "run time: %d microsecs", func() {
		gotRet, err = ugo.NewVM(gotBc).Run(nil)
	})
	require.NoError(t, err)

	require.Equal(t, wantRet, gotRet)
}

func testBytecodeSerialization(t *testing.T, b *ugo.Bytecode, modules *ugo.ModuleMap) {
	t.Helper()

	var buf bytes.Buffer
	err := (*Bytecode)(b).Encode(&buf)
	require.NoError(t, err)

	r := &ugo.Bytecode{}
	err = (*Bytecode)(r).Decode(bytes.NewReader(buf.Bytes()), modules)
	require.NoError(t, err)

	testBytecodesEqual(t, b, r)
}

func testBytecodesEqual(t *testing.T, want, got *ugo.Bytecode) {
	t.Helper()

	require.Equal(t, want.FileSet, got.FileSet)
	require.Equal(t, want.Main, got.Main)
	require.Equalf(t, want.Constants, got.Constants,
		"expected:%s\nactual:%s", tests.Sdump(want.Constants), tests.Sdump(want.Constants))
	testBytecodeConstants(t, want.Constants, got.Constants)
	require.Equal(t, want.NumModules, got.NumModules)
}

func logmicros(t *testing.T, format string, f func()) {
	t0 := gotime.Now()
	f()
	t.Logf(format, gotime.Since(t0).Microseconds())
}
