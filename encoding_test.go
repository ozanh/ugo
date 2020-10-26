package ugo_test

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
	"github.com/ozanh/ugo/token"
)

func TestEncDecObjects(t *testing.T) {
	data, err := Undefined.(encoding.BinaryMarshaler).MarshalBinary()
	require.NoError(t, err)
	if obj, err := DecodeObject(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	} else {
		require.Equal(t, Undefined, obj)
	}

	boolObjects := []Bool{True, False, Bool(true), Bool(false)}
	for _, tC := range boolObjects {
		msg := fmt.Sprintf("Bool(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Bool
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	intObjects := []Int{
		Int(-1), Int(0), Int(1), Int(1<<63 - 1),
	}
	for i := 0; i < 1000; i++ {
		v := seededRand.Int63()
		if i%2 == 0 {
			intObjects = append(intObjects, Int(-v))
		} else {
			intObjects = append(intObjects, Int(v))
		}
	}
	for _, tC := range intObjects {
		msg := fmt.Sprintf("Int(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Int
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	uintObjects := []Uint{Uint(0), Uint(1), ^Uint(0)}
	for i := 0; i < 1000; i++ {
		v := seededRand.Uint64()
		uintObjects = append(uintObjects, Uint(v))
	}
	for _, tC := range uintObjects {
		msg := fmt.Sprintf("Uint(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Uint
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	charObjects := []Char{Char(0)}
	for i := 0; i < 1000; i++ {
		v := seededRand.Int31()
		charObjects = append(charObjects, Char(v))
	}
	for _, tC := range charObjects {
		msg := fmt.Sprintf("Char(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Char
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	floatObjects := []Float{Float(0), Float(-1)}
	for i := 0; i < 1000; i++ {
		v := seededRand.Float64()
		floatObjects = append(floatObjects, Float(v))
	}
	floatObjects = append(floatObjects, Float(math.NaN()))
	for _, tC := range floatObjects {
		msg := fmt.Sprintf("Float(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v Float
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		if math.IsNaN(float64(tC)) {
			require.True(t, math.IsNaN(float64(v)))
		} else {
			require.Equal(t, tC, v, msg)
		}

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		if math.IsNaN(float64(tC)) {
			require.True(t, math.IsNaN(float64(obj.(Float))))
		} else {
			require.Equal(t, tC, obj, msg)
		}
	}
	// remove NaN from Floats slice, array tests below requires NaN check otherwise fails.
	floatObjects = floatObjects[:len(floatObjects)-1]

	stringObjects := []String{String(""), String("çığöşü")}
	for i := 0; i < 1000; i++ {
		stringObjects = append(stringObjects, String(randString(i)))
	}
	for _, tC := range stringObjects {
		msg := fmt.Sprintf("String(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v String
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	bytesObjects := []Bytes{{}, Bytes("çığöşü")}
	for i := 0; i < 1000; i++ {
		bytesObjects = append(bytesObjects, Bytes(randString(i)))
	}
	for _, tC := range bytesObjects {
		msg := fmt.Sprintf("Bytes(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = Bytes{}
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	arrays := []Array{}
	temp1 := Array{}
	for i := range bytesObjects[:100] {
		temp1 = append(temp1, bytesObjects[i])
	}
	arrays = append(arrays, temp1)
	temp2 := Array{}
	for i := range stringObjects[:100] {
		temp2 = append(temp2, stringObjects[i])
	}
	arrays = append(arrays, temp2)
	temp3 := Array{}
	for i := range floatObjects[:100] {
		temp3 = append(temp3, floatObjects[i])
	}
	arrays = append(arrays, temp3)
	temp4 := Array{}
	for i := range charObjects[:100] {
		temp4 = append(temp4, charObjects[i])
	}
	arrays = append(arrays, temp4)
	temp5 := Array{}
	for i := range uintObjects[:100] {
		temp5 = append(temp5, uintObjects[i])
	}
	arrays = append(arrays, temp5)
	temp6 := Array{}
	for i := range intObjects[:100] {
		temp6 = append(temp6, intObjects[i])
	}
	arrays = append(arrays, temp6)
	temp7 := Array{}
	for i := range boolObjects {
		temp7 = append(temp7, boolObjects[i])
	}
	arrays = append(arrays, temp7)
	arrays = append(arrays, Array{Undefined})

	for _, tC := range arrays {
		msg := fmt.Sprintf("Array(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = Array{}
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	maps := []Map{}
	for _, array := range arrays {
		m := Map{}
		s := randString(10)
		r := seededRand.Intn(len(array))
		m[s] = array[r]
		maps = append(maps, m)
	}

	for _, tC := range maps {
		msg := fmt.Sprintf("Map(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = Map{}
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	syncMaps := []*SyncMap{}
	for _, m := range maps {
		syncMaps = append(syncMaps, &SyncMap{Map: m})
	}
	for _, tC := range syncMaps {
		msg := fmt.Sprintf("SyncMap(%v)", tC)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = &SyncMap{}
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	compFuncs := []*CompiledFunction{
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
			makeInst(OpConstant, 0),
			makeInst(OpConstant, 1),
			makeInst(OpBinaryOp, int(token.Add)),
		),
			withParams(1),
			withVariadic(),
			withLocals(2),
			withSourceMap(map[int]int{0: 1, 3: 1, 5: 1}),
		),
	}
	for i, tC := range compFuncs {
		msg := fmt.Sprintf("CompiledFunction #%d", i)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = &CompiledFunction{}
		err = v.UnmarshalBinary(data)
		require.NoError(t, err, msg)
		require.Equal(t, tC, v, msg)

		obj, err := DecodeObject(bytes.NewReader(data))
		require.NoError(t, err, msg)
		require.Equal(t, tC, obj, msg)
	}

	builtinFuncs := []*BuiltinFunction{}
	for _, o := range BuiltinObjects {
		if f, ok := o.(*BuiltinFunction); ok {
			builtinFuncs = append(builtinFuncs, f)
		}
	}
	for _, tC := range builtinFuncs {
		msg := fmt.Sprintf("BuiltinFunction %s", tC.Name)
		data, err := tC.MarshalBinary()
		require.NoError(t, err, msg)
		require.Greater(t, len(data), 0, msg)
		var v = &BuiltinFunction{}
		err = v.UnmarshalBinary(data)
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
	m := {a: 1, b: ["abc"], c: {x: bytes()}, builtins: [append, len]}
	`, nil, Undefined)

	testEncDecBytecode(t, `
	mod1 := import("mod1")
	mod2 := import("mod2")
	return mod1.run() + mod2.run()
	`, newOpts().Module("mod1", Map{
		"run": &Function{
			Name: "run",
			Value: func(args ...Object) (Object, error) {
				return String("mod1"), nil
			},
		},
	}).Module("mod2", `return {run: func(){ return "mod2" }}`),
		String("mod1mod2"))

	// t.Logf("Original:\n%s\nSecond:\n%s\nThird:\n%s", bc, &bc2, &bc2)

}

func testEncDecBytecode(t *testing.T, script string, opts *testopts, expected Object) {
	t.Helper()
	if opts == nil {
		opts = newOpts()
	}
	var initialModuleMap *ModuleMap
	if opts.moduleMap != nil {
		initialModuleMap = opts.moduleMap.Copy()
	}
	bc, err := Compile([]byte(script), CompilerOptions{
		ModuleMap: opts.moduleMap,
	})
	require.NoError(t, err)
	ret, err := NewVM(bc).Run(opts.globals, opts.args...)
	require.NoError(t, err)
	require.Equal(t, expected, ret)

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(bc)
	require.NoError(t, err)
	t.Logf("GobSize:%d", len(buf.Bytes()))
	bcData, err := bc.MarshalBinary()
	require.NoError(t, err)
	t.Logf("BinSize:%d", len(bcData))

	if opts.moduleMap == nil {
		var bc2 Bytecode
		err = gob.NewDecoder(&buf).Decode(&bc2)
		require.NoError(t, err)
		testDecodedBytecodeEqual(t, bc, &bc2)
		ret, err := NewVM(&bc2).Run(opts.globals, opts.args...)
		require.NoError(t, err)
		require.Equal(t, expected, ret)

		var bc3 Bytecode
		err = bc3.UnmarshalBinary(bcData)
		require.NoError(t, err)
		testDecodedBytecodeEqual(t, bc, &bc3)
		ret, err = NewVM(&bc3).Run(opts.globals, opts.args...)
		require.NoError(t, err)
		require.Equal(t, expected, ret)
	}

	var bc4 Bytecode
	err = bc4.Decode(bytes.NewReader(bcData), opts.moduleMap, nil)
	require.NoError(t, err)
	testDecodedBytecodeEqual(t, bc, &bc4)
	ret, err = NewVM(&bc4).Run(opts.globals, opts.args...)
	require.NoError(t, err)
	require.Equal(t, expected, ret)
	// ensure moduleMap is not updated during compilation and decoding
	require.Equal(t, initialModuleMap, opts.moduleMap)
}

func testDecodedBytecodeEqual(t *testing.T, actual, decoded *Bytecode) {
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

func getModuleName(obj Object) (string, bool) {
	if m, ok := obj.(Map); ok {
		if n, ok := m["__module_name__"]; ok {
			return string(n.(String)), true
		}
	}
	return "", false
}

func testBytecodeConstants(t *testing.T, expected, decoded []Object) {
	t.Helper()
	if len(decoded) != len(expected) {
		t.Fatalf("constants length not equal want %d, got %d", len(decoded), len(expected))
	}
	Len := func(v Object) Object {
		ret, err := BuiltinObjects[BuiltinLen].Call(v)
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
				f1, ok := v1.(*Function)
				if ok {
					f2 := v2.(*Function)
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

func BenchmarkBytecodeUnmarshal(b *testing.B) {
	b.ReportAllocs()
	script := `
	f := func() {
		return [undefined, true, false, "", -1, 0, 1, 2u, 3.0, 'a', bytes(0, 1, 2)]
	}
	f()
	m := {a: 1, b: ["abc"], c: {x: bytes()}, builtins: [append, len]}
	`
	var err error
	bc, err := Compile([]byte(script), CompilerOptions{})
	if err != nil {
		b.Fatal(err)
	}
	// bc.FileSet = nil
	// bc.Main.SourceMap = nil
	d, err := bc.MarshalBinary()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var bc2 Bytecode
		err := bc2.UnmarshalBinary(d)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(len(d)), "Bytes")
}

func BenchmarkBytecodeDecode(b *testing.B) {
	b.ReportAllocs()
	script := `
	f := func() {
		return [undefined, true, false, "", -1, 0, 1, 2u, 3.0, 'a', bytes(0, 1, 2)]
	}
	f()
	m := {a: 1, b: ["abc"], c: {x: bytes()}, builtins: [append, len]}
	`
	var err error
	bc, err := Compile([]byte(script), CompilerOptions{})
	if err != nil {
		b.Fatal(err)
	}
	// bc.FileSet = nil
	// bc.Main.SourceMap = nil
	d, err := bc.MarshalBinary()
	if err != nil {
		b.Fatal(err)
	}
	rd := bytes.NewReader(d)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rd.Reset(d)
		var bc2 Bytecode
		err := bc2.Decode(rd, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(len(d)), "Bytes")
}

func BenchmarkBytecodeEncDec(b *testing.B) {
	b.ReportAllocs()
	script := `
	f := func() {
		return [undefined, true, false, "", -1, 0, 1, 2u, 3.0, 'a', bytes(0, 1, 2)]
	}
	f()
	m := {a: 1, b: ["abc"], c: {x: bytes()}, builtins: [append, len]}
	`
	var err error
	bc, err := Compile([]byte(script), CompilerOptions{})
	if err != nil {
		b.Fatal(err)
	}

	b.Run("compileUnopt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Compile([]byte(script), CompilerOptions{})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("compileOpt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Compile([]byte(script), CompilerOptions{
				OptimizeConst:     true,
				OptimizeExpr:      true,
				OptimizerMaxCycle: 100,
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("marshal", func(b *testing.B) {
		var size int
		for i := 0; i < b.N; i++ {
			d, err := bc.MarshalBinary()
			if err != nil {
				b.Fatal(err)
			}
			if size == 0 {
				size = len(d)
			}
		}
		b.ReportMetric(float64(size), "Bytes")
	})
	b.Run("gobEncode", func(b *testing.B) {
		var size int
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			err := gob.NewEncoder(&buf).Encode(bc)
			if err != nil {
				b.Fatal(err)
			}
			if size == 0 {
				size = buf.Len()
			}
		}
		b.ReportMetric(float64(size), "Bytes")
	})
	b.Run("unmarshal", func(b *testing.B) {
		d, err := bc.MarshalBinary()
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var bc2 Bytecode
			err := bc2.UnmarshalBinary(d)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(len(d)), "Bytes")
	})
	b.Run("gobDecode", func(b *testing.B) {
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(bc)
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var bc2 Bytecode
			err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&bc2)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(len(buf.Bytes())), "Bytes")
	})
}

func BenchmarkIntEncDec(b *testing.B) {
	b.Run("marshal unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data, err := Int(i).MarshalBinary()
			if err != nil {
				b.Fatal(err)
			}
			var v Int
			err = v.UnmarshalBinary(data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("decode object", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data, err := Int(i).MarshalBinary()
			if err != nil {
				b.Fatal(err)
			}
			_, err = DecodeObject(bytes.NewReader(data))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

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
