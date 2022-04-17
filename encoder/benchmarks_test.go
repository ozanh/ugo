package encoder_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/ozanh/ugo"
	. "github.com/ozanh/ugo/encoder"
)

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
	bc, err := ugo.Compile([]byte(script), ugo.CompilerOptions{})
	if err != nil {
		b.Fatal(err)
	}
	// bc.FileSet = nil
	// bc.Main.SourceMap = nil
	d, err := (*Bytecode)(bc).MarshalBinary()
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
	bc, err := ugo.Compile([]byte(script), ugo.CompilerOptions{})
	if err != nil {
		b.Fatal(err)
	}
	// bc.FileSet = nil
	// bc.Main.SourceMap = nil
	d, err := (*Bytecode)(bc).MarshalBinary()
	if err != nil {
		b.Fatal(err)
	}
	rd := bytes.NewReader(d)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rd.Reset(d)
		_, err := DecodeBytecodeFrom(rd, nil)
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
	bc, err := ugo.Compile([]byte(script), ugo.CompilerOptions{})
	if err != nil {
		b.Fatal(err)
	}

	b.Run("compileUnopt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := ugo.Compile([]byte(script), ugo.CompilerOptions{})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("compileOpt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := ugo.Compile([]byte(script), ugo.CompilerOptions{
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
			d, err := (*Bytecode)(bc).MarshalBinary()
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
			err := gob.NewEncoder(&buf).Encode((*Bytecode)(bc))
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
		d, err := (*Bytecode)(bc).MarshalBinary()
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
		err := gob.NewEncoder(&buf).Encode((*Bytecode)(bc))
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
