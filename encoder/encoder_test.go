package encoder_test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	gotime "time"

	"github.com/stretchr/testify/require"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/stdlib/json"
	"github.com/ozanh/ugo/stdlib/time"
	"github.com/ozanh/ugo/token"

	. "github.com/ozanh/ugo/encoder"
)

func TestGobEncoder(t *testing.T) {
	objects := []ugo.Object{
		ugo.Undefined,
		ugo.Bool(true),
		ugo.Int(0),
		ugo.Uint(0),
		ugo.Char(0),
		ugo.Float(0),
		ugo.String("abc"),
		ugo.Bytes{},
		ugo.Array{ugo.Bool(true), ugo.String("")},
		ugo.Map{"b": ugo.Bool(true), "s": ugo.String("")},
		&ugo.SyncMap{Value: ugo.Map{"i": ugo.Int(0), "u": ugo.Uint(0)}},
		&ugo.ObjectPtr{},
		&time.Time{Value: gotime.Now()},
		&json.EncoderOptions{Value: ugo.Float(0)},
		&json.RawMessage{},
	}
	for _, obj := range objects {
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(obj)
		require.NoError(t, err)
	}
}

func TestEncDecObjects(t *testing.T) {
	data, err := (*UndefinedType)(ugo.Undefined.(*ugo.UndefinedType)).MarshalBinary()
	require.NoError(t, err)
	if obj, err := DecodeObject(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	} else {
		require.Equal(t, ugo.Undefined, obj)
	}

	boolObjects := []ugo.Bool{ugo.True, ugo.False, ugo.Bool(true), ugo.Bool(false)}
	for _, tC := range boolObjects {
		msg := fmt.Sprintf("Bool(%v)", tC)
		data, err := Bool(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Bool
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, Bool(tC), v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	intObjects := []ugo.Int{
		ugo.Int(-1), ugo.Int(0), ugo.Int(1), ugo.Int(1<<63 - 1),
	}
	for i := 0; i < 1000; i++ {
		v := seededRand.Int63()
		if i%2 == 0 {
			intObjects = append(intObjects, ugo.Int(-v))
		} else {
			intObjects = append(intObjects, ugo.Int(v))
		}
	}
	for _, tC := range intObjects {
		msg := fmt.Sprintf("Int(%v)", tC)
		data, err := Int(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Int
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, Int(tC), v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	uintObjects := []ugo.Uint{ugo.Uint(0), ugo.Uint(1), ^ugo.Uint(0)}
	for i := 0; i < 1000; i++ {
		v := seededRand.Uint64()
		uintObjects = append(uintObjects, ugo.Uint(v))
	}
	for _, tC := range uintObjects {
		msg := fmt.Sprintf("Uint(%v)", tC)
		data, err := Uint(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Uint
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, Uint(tC), v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	charObjects := []ugo.Char{ugo.Char(0)}
	for i := 0; i < 1000; i++ {
		v := seededRand.Int31()
		charObjects = append(charObjects, ugo.Char(v))
	}
	for _, tC := range charObjects {
		msg := fmt.Sprintf("Char(%v)", tC)
		data, err := Char(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Char
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, Char(tC), v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	floatObjects := []ugo.Float{ugo.Float(0), ugo.Float(-1)}
	for i := 0; i < 1000; i++ {
		v := seededRand.Float64()
		floatObjects = append(floatObjects, ugo.Float(v))
	}
	floatObjects = append(floatObjects, ugo.Float(math.NaN()))
	for _, tC := range floatObjects {
		msg := fmt.Sprintf("Float(%v)", tC)
		data, err := Float(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Float
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		if math.IsNaN(float64(tC)) {
			require.True(t, math.IsNaN(float64(v)))
		} else {
			require.Equal(t, Float(tC), v, msg)
		}

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		if math.IsNaN(float64(tC)) {
			require.True(t, math.IsNaN(float64(obj.(ugo.Float))))
		} else {
			require.Equal(t, tC, obj, msg)
		}
	}
	// remove NaN from Floats slice, array tests below requires NaN check otherwise fails.
	floatObjects = floatObjects[:len(floatObjects)-1]

	stringObjects := []ugo.String{ugo.String(""), ugo.String("çığöşü")}
	for i := 0; i < 1000; i++ {
		stringObjects = append(stringObjects, ugo.String(randString(i)))
	}
	for _, tC := range stringObjects {
		msg := fmt.Sprintf("String(%v)", tC)
		data, err := String(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v String
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, String(tC), v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	bytesObjects := []ugo.Bytes{{}, ugo.Bytes("çığöşü")}
	for i := 0; i < 1000; i++ {
		bytesObjects = append(bytesObjects, ugo.Bytes(randString(i)))
	}
	for _, tC := range bytesObjects {
		msg := fmt.Sprintf("Bytes(%v)", tC)
		data, err := Bytes(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = Bytes{}
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, Bytes(tC), v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	arrays := []ugo.Array{}
	temp1 := ugo.Array{}
	for i := range bytesObjects[:100] {
		temp1 = append(temp1, bytesObjects[i])
	}
	arrays = append(arrays, temp1)
	temp2 := ugo.Array{}
	for i := range stringObjects[:100] {
		temp2 = append(temp2, stringObjects[i])
	}
	arrays = append(arrays, temp2)
	temp3 := ugo.Array{}
	for i := range floatObjects[:100] {
		temp3 = append(temp3, floatObjects[i])
	}
	arrays = append(arrays, temp3)
	temp4 := ugo.Array{}
	for i := range charObjects[:100] {
		temp4 = append(temp4, charObjects[i])
	}
	arrays = append(arrays, temp4)
	temp5 := ugo.Array{}
	for i := range uintObjects[:100] {
		temp5 = append(temp5, uintObjects[i])
	}
	arrays = append(arrays, temp5)
	temp6 := ugo.Array{}
	for i := range intObjects[:100] {
		temp6 = append(temp6, intObjects[i])
	}
	arrays = append(arrays, temp6)
	temp7 := ugo.Array{}
	for i := range boolObjects {
		temp7 = append(temp7, boolObjects[i])
	}
	arrays = append(arrays, temp7)
	arrays = append(arrays, ugo.Array{ugo.Undefined})

	for _, tC := range arrays {
		msg := fmt.Sprintf("Array(%v)", tC)
		data, err := Array(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = ugo.Array{}
		err = (*Array)(&v).UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	maps := []ugo.Map{}
	for _, array := range arrays {
		m := ugo.Map{}
		s := randString(10)
		r := seededRand.Intn(len(array))
		m[s] = array[r]
		maps = append(maps, m)
	}

	for _, tC := range maps {
		msg := fmt.Sprintf("Map(%v)", tC)
		data, err := Map(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = ugo.Map{}
		err = (*Map)(&v).UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	syncMaps := []*ugo.SyncMap{}
	for _, m := range maps {
		syncMaps = append(syncMaps, &ugo.SyncMap{Value: m})
	}
	for _, tC := range syncMaps {
		msg := fmt.Sprintf("SyncMap(%v)", tC)
		data, err := (*SyncMap)(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = &ugo.SyncMap{}
		err = (*SyncMap)(v).UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	compFuncs := []*ugo.CompiledFunction{
		compFunc(nil),
		compFunc(nil,
			withLocals(10),
		),
		compFunc(nil,
			withParams(2),
		),
		compFunc(nil,
			withVariadic(),
		),
		compFunc(nil,
			withSourceMap(map[int]int{0: 1, 3: 1, 5: 1}),
		),
		compFunc(concatInsts(
			makeInst(ugo.OpConstant, 0),
			makeInst(ugo.OpConstant, 1),
			makeInst(ugo.OpBinaryOp, int(token.Add)),
		),
			withParams(1),
			withVariadic(),
			withLocals(2),
			withSourceMap(map[int]int{0: 1, 3: 1, 5: 1}),
		),
	}
	for i, tC := range compFuncs {
		msg := fmt.Sprintf("CompiledFunction #%d", i)
		data, err := (*CompiledFunction)(tC).MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = &ugo.CompiledFunction{}
		err = (*CompiledFunction)(v).UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	builtinFuncs := []*BuiltinFunction{}
	for _, o := range ugo.BuiltinObjects {
		if f, ok := o.(*BuiltinFunction); ok {
			builtinFuncs = append(builtinFuncs, f)
		}
	}
	for _, tC := range builtinFuncs {
		msg := fmt.Sprintf("BuiltinFunction %s", tC.Name)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = &ugo.BuiltinFunction{}
		err = (*BuiltinFunction)(v).UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC.Name, v.Name)
		require.NotNil(t, v.Value)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC.Name, obj.(*BuiltinFunction).Name, msg)
		require.NotNil(t, obj.(*BuiltinFunction).Value, msg)
	}

}

func TestEncDecBytecode(t *testing.T) {
	testEncDecBytecode(t, `
	f := func() {
		return [undefined, true, false, "", -1, 0, 1, 2u, 3.0, 'a', bytes(0, 1, 2)]
	}
	f()
	m := {a: 1, b: ["abc"], c: {x: bytes()}, builtins: [append, len]}`, nil, ugo.Undefined)
}

func TestEncDecBytecode_modules(t *testing.T) {
	testEncDecBytecode(t, `
	mod1 := import("mod1")
	mod2 := import("mod2")
	return mod1.run() + mod2.run()
	`, newOpts().Module("mod1", ugo.Map{
		"run": &ugo.Function{
			Name: "run",
			Value: func(args ...ugo.Object) (ugo.Object, error) {
				return ugo.String("mod1"), nil
			},
		},
	}).Module("mod2", `return {run: func(){ return "mod2" }}`), ugo.String("mod1mod2"))
}

func testEncDecBytecode(t *testing.T, script string, opts *testopts, expected ugo.Object) {
	t.Helper()
	if opts == nil {
		opts = newOpts()
	}
	var initialModuleMap *ugo.ModuleMap
	if opts.moduleMap != nil {
		initialModuleMap = opts.moduleMap.Copy()
	}
	bc, err := ugo.Compile([]byte(script),
		ugo.CompilerOptions{
			ModuleMap: opts.moduleMap,
		},
	)
	require.NoError(t, err)
	ret, err := ugo.NewVM(bc).Run(opts.globals, opts.args...)
	require.NoError(t, err)
	require.Equal(t, expected, ret)

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode((*Bytecode)(bc))
	require.NoError(t, err)
	t.Logf("GobSize:%d", len(buf.Bytes()))
	bcData, err := (*Bytecode)(bc).MarshalBinary()
	require.NoError(t, err)
	t.Logf("BinSize:%d", len(bcData))

	if opts.moduleMap == nil {
		var bc2 ugo.Bytecode
		err = gob.NewDecoder(&buf).Decode((*Bytecode)(&bc2))
		require.NoError(t, err)
		testDecodedBytecodeEqual(t, bc, &bc2)
		ret, err := ugo.NewVM(&bc2).Run(opts.globals, opts.args...)
		require.NoError(t, err)
		require.Equal(t, expected, ret)

		var bc3 ugo.Bytecode
		err = (*Bytecode)(&bc3).UnmarshalBinary(bcData)
		require.NoError(t, err)
		testDecodedBytecodeEqual(t, bc, &bc3)
		ret, err = ugo.NewVM(&bc3).Run(opts.globals, opts.args...)
		require.NoError(t, err)
		require.Equal(t, expected, ret)
	}

	bc4, err := DecodeBytecodeFrom(bytes.NewReader(bcData), opts.moduleMap)
	require.NoError(t, err)
	testDecodedBytecodeEqual(t, bc, bc4)
	ret, err = ugo.NewVM(bc4).Run(opts.globals, opts.args...)
	require.NoError(t, err)
	require.Equal(t, expected, ret)
	// ensure moduleMap is not updated during compilation and decoding
	require.Equal(t, initialModuleMap, opts.moduleMap)
}

func testDecodedBytecodeEqual(t *testing.T, actual, decoded *ugo.Bytecode) {
	t.Helper()
	msg := fmt.Sprintf("actual:%s\ndecoded:%s\n", actual, decoded)

	testBytecodeConstants(t, actual.Constants, decoded.Constants)
	require.Equal(t, actual.Main, decoded.Main, msg)
	require.Equal(t, actual.NumModules, decoded.NumModules, msg)
	if actual.FileSet == nil {
		require.Nil(t, decoded.FileSet, msg)
	} else {
		require.Equal(t, actual.FileSet.Base, decoded.FileSet.Base, msg)
		require.Equal(t, len(actual.FileSet.Files), len(decoded.FileSet.Files), msg)
		for i, f := range actual.FileSet.Files {
			f2 := decoded.FileSet.Files[i]
			require.Equal(t, f.Base, f2.Base, msg)
			require.Equal(t, f.Lines, f2.Lines, msg)
			require.Equal(t, f.Name, f2.Name, msg)
			require.Equal(t, f.Size, f2.Size, msg)
		}
		require.NotNil(t, actual.FileSet.LastFile, msg)
		require.Nil(t, decoded.FileSet.LastFile, msg)
	}
}

func getModuleName(obj ugo.Object) (string, bool) {
	if m, ok := obj.(ugo.Map); ok {
		if n, ok := m[ugo.AttrModuleName]; ok {
			return string(n.(ugo.String)), true
		}
	}
	return "", false
}

func testBytecodeConstants(t *testing.T, expected, decoded []ugo.Object) {
	t.Helper()
	if len(decoded) != len(expected) {
		t.Fatalf("constants length not equal want %d, got %d", len(decoded), len(expected))
	}
	Len := func(v ugo.Object) ugo.Object {
		ret, err := ugo.BuiltinObjects[ugo.BuiltinLen].Call(v)
		if err != nil {
			t.Fatalf("%v: length error for '%v'", err, v)
		}
		return ret
	}
	for i := range decoded {
		modName, ok1 := getModuleName(expected[i])
		decModName, ok2 := getModuleName(decoded[i])
		if ok1 {
			require.True(t, ok2)
			require.Equal(t, modName, decModName)
			require.Equal(t, reflect.TypeOf(expected[i]), reflect.TypeOf(decoded[i]))
			require.Equal(t, Len(expected[i]), Len(decoded[i]))
			if !expected[i].CanIterate() {
				require.False(t, decoded[i].CanIterate())
				continue
			}
			it := expected[i].Iterate()
			decIt := decoded[i].Iterate()
			for decIt.Next() {
				require.True(t, it.Next())
				key := decIt.Key()
				v1, err := expected[i].IndexGet(key)
				require.NoError(t, err)
				v2 := decIt.Value()
				if (v1 != nil && v2 == nil) || (v1 == nil && v2 != nil) {
					t.Fatalf("decoded constant index %d not equal", i)
				}
				f1, ok := v1.(*ugo.Function)
				if ok {
					f2 := v2.(*ugo.Function)
					require.Equal(t, f1.Name, f2.Name)
					require.NotNil(t, f2.Value)
					// Note that this is not a guaranteed way to compare func pointers
					require.Equal(t, reflect.ValueOf(f1.Value).Pointer(),
						reflect.ValueOf(f2.Value).Pointer())
				} else {
					require.Equal(t, v1, v2)
				}
			}
			require.False(t, it.Next())
			continue
		}
		require.Equalf(t, expected[i], decoded[i],
			"constant index %d not equal want %v, got %v", i, expected[i], decoded[i])
		require.NotNil(t, decoded[i])
	}
}

type funcOpt func(*ugo.CompiledFunction)

func withParams(numParams int) funcOpt {
	return func(cf *ugo.CompiledFunction) {
		cf.NumParams = numParams
	}
}

func withLocals(numLocals int) funcOpt {
	return func(cf *ugo.CompiledFunction) {
		cf.NumLocals = numLocals
	}
}

func withVariadic() funcOpt {
	return func(cf *ugo.CompiledFunction) {
		cf.Variadic = true
	}
}

func withSourceMap(m map[int]int) funcOpt {
	return func(cf *ugo.CompiledFunction) {
		cf.SourceMap = m
	}
}

func compFunc(insts []byte, opts ...funcOpt) *ugo.CompiledFunction {
	cf := &ugo.CompiledFunction{
		Instructions: insts,
	}
	for _, f := range opts {
		f(cf)
	}
	return cf
}

func makeInst(op ugo.Opcode, args ...int) []byte {
	b, err := ugo.MakeInstruction(make([]byte, 8), op, args...)
	if err != nil {
		panic(err)
	}
	return b
}

func concatInsts(insts ...[]byte) []byte {
	var out []byte
	for i := range insts {
		out = append(out, insts[i]...)
	}
	return out
}

type testopts struct {
	globals       ugo.Object
	args          []ugo.Object
	moduleMap     *ugo.ModuleMap
	skip2pass     bool
	isCompilerErr bool
	noPanic       bool
}

func newOpts() *testopts {
	return &testopts{}
}

func (t *testopts) Globals(globals ugo.Object) *testopts {
	t.globals = globals
	return t
}

func (t *testopts) Args(args ...ugo.Object) *testopts {
	t.args = args
	return t
}

func (t *testopts) Skip2Pass() *testopts {
	t.skip2pass = true
	return t
}

func (t *testopts) CompilerError() *testopts {
	t.isCompilerErr = true
	return t
}

func (t *testopts) NoPanic() *testopts {
	t.noPanic = true
	return t
}

func (t *testopts) Module(name string, module interface{}) *testopts {
	if t.moduleMap == nil {
		t.moduleMap = ugo.NewModuleMap()
	}
	switch v := module.(type) {
	case []byte:
		t.moduleMap.AddSourceModule(name, v)
	case string:
		t.moduleMap.AddSourceModule(name, []byte(v))
	case map[string]ugo.Object:
		t.moduleMap.AddBuiltinModule(name, v)
	case ugo.Map:
		t.moduleMap.AddBuiltinModule(name, v)
	case ugo.Importable:
		t.moduleMap.Add(name, v)
	default:
		panic(fmt.Errorf("invalid module type: %T", module))
	}
	return t
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(gotime.Now().UnixNano()))

func randStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randString(length int) string {
	return randStringWithCharset(length, charset)
}
