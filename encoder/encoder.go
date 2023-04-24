// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package encoder

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/parser"
	"github.com/ozanh/ugo/stdlib/json"
	"github.com/ozanh/ugo/stdlib/time"
)

// Bytecode signature and version are written to the header of encoded Bytecode.
// Bytecode is encoded with current BytecodeVersion and its format.
const (
	BytecodeSignature uint32 = 0x75474F
	BytecodeVersion   uint16 = 1
)

// Types implementing encoding.BinaryMarshaler encoding.BinaryUnmarshaler.
type (
	Bytecode         ugo.Bytecode
	CompiledFunction ugo.CompiledFunction
	BuiltinFunction  ugo.BuiltinFunction
	Function         ugo.Function
	UndefinedType    ugo.UndefinedType
	String           ugo.String
	Bytes            ugo.Bytes
	Array            ugo.Array
	Map              ugo.Map
	SyncMap          ugo.SyncMap
	Int              ugo.Int
	Uint             ugo.Uint
	Char             ugo.Char
	Float            ugo.Float
	Bool             ugo.Bool
	SourceFileSet    parser.SourceFileSet
	SourceFile       parser.SourceFile
)

const (
	binUndefinedV1 byte = iota
	binTrueV1
	binFalseV1
	binIntV1
	binUintV1
	binCharV1
	binFloatV1
	binStringV1
	binBytesV1
	binArrayV1
	binMapV1
	binSyncMapV1
	binCompiledFunctionV1
	binFunctionV1
	binBuiltinFunctionV1

	binUnkownType byte = 255
)

var (
	errVarintTooSmall = errors.New("read varint error: buf too small")
	errVarintOverflow = errors.New("read varint error: value larger than 64 bits (overflow)")
)

func init() {
	gob.Register(ugo.Undefined)
	gob.Register(ugo.Bool(true))
	gob.Register(ugo.Int(0))
	gob.Register(ugo.Uint(0))
	gob.Register(ugo.Char(0))
	gob.Register(ugo.Float(0))
	gob.Register(ugo.String(""))
	gob.Register(ugo.Bytes(nil))
	gob.Register(ugo.Array(nil))
	gob.Register(ugo.Map(nil))
	gob.Register((*ugo.Error)(nil))
	gob.Register((*ugo.RuntimeError)(nil))
	gob.Register((*ugo.SyncMap)(nil))
	gob.Register((*ugo.ObjectPtr)(nil))
	gob.Register((*time.Time)(nil))
	gob.Register((*json.EncoderOptions)(nil))
	gob.Register((*json.RawMessage)(nil))
}

// MarshalBinary implements encoding.BinaryMarshaler
func (bc *Bytecode) MarshalBinary() (data []byte, err error) {
	switch BytecodeVersion {
	case 1:
		var buf bytes.Buffer
		if err = bc.bytecodeV1Encoder(&buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		panic("invalid Bytecode version:" + strconv.Itoa(int(BytecodeVersion)))
	}
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
// Do not use this method if builtin modules are used, instead use Decode method.
func (bc *Bytecode) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return &ugo.Error{
			Name:    "encoder.Bytecode.UnmarshalBinary",
			Message: "invalid data",
		}
	}

	sig := binary.BigEndian.Uint32(data[0:4])
	if sig != BytecodeSignature {
		return &ugo.Error{
			Name:    "encoder.Bytecode.UnmarshalBinary",
			Message: "signature mismatch",
		}
	}

	version := binary.BigEndian.Uint16(data[4:6])
	switch version {
	case BytecodeVersion:
		buf := bytes.NewBuffer(data[6:])
		err := bc.bytecodeV1Decoder(buf)
		if err != nil {
			return err
		}
		return nil
	default:
		return &ugo.Error{
			Name:    "encoder.Bytecode.UnmarshalBinary",
			Message: "unsupported version:" + strconv.Itoa(int(version)),
		}
	}
}

func putBytecodeHeader(w io.Writer) (err error) {
	sig := make([]byte, 4)
	binary.BigEndian.PutUint32(sig, BytecodeSignature)
	if _, err = io.Copy(w, bytes.NewReader(sig)); err != nil {
		return
	}

	bcVersion := make([]byte, 2)
	binary.BigEndian.PutUint16(bcVersion, BytecodeVersion)

	if _, err = io.Copy(w, bytes.NewReader(bcVersion)); err != nil {
		return
	}
	return nil
}

func (bc *Bytecode) bytecodeV1Encoder(w io.Writer) (err error) {
	if err = putBytecodeHeader(w); err != nil {
		return
	}

	// FileSet, field #0
	if bc.FileSet != nil {
		_ = writeByteTo(w, 0)
		var data []byte
		fs := (*SourceFileSet)(bc.FileSet)
		if data, err = fs.MarshalBinary(); err != nil {
			return
		}
		var sz []byte
		if sz, err = Int(len(data)).MarshalBinary(); err != nil {
			return
		}
		_, _ = w.Write(sz)
		_, _ = w.Write(data)
	}

	// Main, field #1
	if bc.Main != nil {
		_ = writeByteTo(w, 1)
		var data []byte
		if data, err = (*CompiledFunction)(bc.Main).MarshalBinary(); err != nil {
			return
		}
		if _, err = w.Write(data); err != nil {
			return
		}
	}

	// Constants, field #2
	if bc.Constants != nil {
		_ = writeByteTo(w, 2)
		var data []byte
		if data, err = Array(bc.Constants).MarshalBinary(); err != nil {
			return
		}
		if _, err = w.Write(data); err != nil {
			return
		}
	}

	// NumModules, field #3
	if bc.NumModules > 0 {
		_ = writeByteTo(w, 3)
		var data []byte
		data, err = Int(bc.NumModules).MarshalBinary()
		if err != nil {
			return
		}
		if _, err = w.Write(data); err != nil {
			return
		}
	}
	return nil
}

func (bc *Bytecode) bytecodeV1Decoder(r *bytes.Buffer) error {
	for {
		field, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch field {
		case 0:
			obj, err := DecodeObject(r)
			if err != nil {
				return err
			}

			sz := obj.(ugo.Int)
			if sz <= 0 {
				continue
			}

			data := make([]byte, sz)
			if _, err = io.ReadFull(r, data); err != nil {
				return err
			}

			var fs SourceFileSet
			if err = fs.UnmarshalBinary(data); err != nil {
				return err
			}
			bc.FileSet = (*parser.SourceFileSet)(&fs)
		case 1:
			f, err := DecodeObject(r)
			if err != nil {
				return err
			}

			bc.Main = f.(*ugo.CompiledFunction)
		case 2:
			obj, err := DecodeObject(r)
			if err != nil {
				return err
			}

			bc.Constants = obj.(ugo.Array)
		case 3:
			num, err := DecodeObject(r)
			if err != nil {
				return err
			}

			bc.NumModules = int(num.(ugo.Int))
		default:
			return errors.New("unknown field:" + strconv.Itoa(int(field)))
		}
	}
}

// DecodeObject decodes and returns Object from a io.Reader which is encoded with MarshalBinary.
func DecodeObject(r io.Reader) (ugo.Object, error) {
	btype, err := readByteFrom(r)
	if err != nil {
		return nil, err
	}

	switch btype {
	case binUndefinedV1:
		return ugo.Undefined, nil
	case binTrueV1:
		return ugo.True, nil
	case binFalseV1:
		return ugo.False, nil
	case binIntV1,
		binUintV1,
		binFloatV1,
		binCharV1:

		size, err := readByteFrom(r)
		if err != nil {
			return nil, err
		}

		buf := make([]byte, 2+size)
		buf[0] = btype
		buf[1] = size
		if size > 0 {
			if _, err = io.ReadFull(r, buf[2:]); err != nil {
				return nil, err
			}
		}

		switch btype {
		case binIntV1:
			var v Int
			if err = v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return ugo.Int(v), nil
		case binUintV1:
			var v Uint
			if err = v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return ugo.Uint(v), nil
		case binFloatV1:
			var v Float
			if err = v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return ugo.Float(v), nil
		case binCharV1:
			var v Char
			if err = v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return ugo.Char(v), nil
		}
	case binCompiledFunctionV1,
		binArrayV1,
		binBytesV1,
		binStringV1,
		binMapV1,
		binSyncMapV1,
		binFunctionV1,
		binBuiltinFunctionV1:

		var vi varintConv
		value, readBytes, err := vi.readBytes(r)
		if err != nil {
			return nil, err
		}

		if value < 0 {
			return nil, errors.New("negative value")
		}

		n := 1 + len(readBytes)
		buf := make([]byte, n+int(value))
		buf[0] = btype
		copy(buf[1:], readBytes)

		if value > 0 {
			if _, err = io.ReadFull(r, buf[n:]); err != nil {
				return nil, err
			}
		}

		switch btype {
		case binCompiledFunctionV1:
			var v CompiledFunction
			if err := v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return (*ugo.CompiledFunction)(&v), nil
		case binArrayV1:
			var v = Array{}
			if err := v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return ugo.Array(v), nil
		case binBytesV1:
			var v = Bytes{}
			if err := v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return ugo.Bytes(v), nil
		case binStringV1:
			var v String
			if err := v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return ugo.String(v), nil
		case binMapV1:
			var v = Map{}
			if err := v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return ugo.Map(v), nil
		case binSyncMapV1:
			var v SyncMap
			if err := v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return (*ugo.SyncMap)(&v), nil
		case binFunctionV1:
			var v Function
			if err := v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return (*ugo.Function)(&v), nil
		case binBuiltinFunctionV1:
			var v BuiltinFunction
			if err := v.UnmarshalBinary(buf); err != nil {
				return nil, err
			}
			return (*ugo.BuiltinFunction)(&v), nil
		}
	case binUnkownType:
		var v ugo.Object
		if err := gob.NewDecoder(r).Decode(&v); err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, errors.New(
		"decode error: unknown encoding type:" + strconv.Itoa(int(btype)),
	)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o *UndefinedType) MarshalBinary() ([]byte, error) {
	return []byte{binUndefinedV1}, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *UndefinedType) UnmarshalBinary(data []byte) error {
	if len(data) < 1 || data[0] != binUndefinedV1 {
		return errors.New("invalid ugo.Undefined data")
	}
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o Bool) MarshalBinary() ([]byte, error) {
	if o {
		return []byte{binTrueV1}, nil
	}
	return []byte{binFalseV1}, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Bool) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("invalid ugo.Bool data")
	}

	if data[0] == binTrueV1 {
		*o = true
		return nil
	}

	if data[0] == binFalseV1 {
		*o = false
		return nil
	}
	return errors.New("invalid ugo.Bool data")
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o Int) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 2+binary.MaxVarintLen64)
	buf[0] = binIntV1

	if o == 0 {
		buf[1] = 0
		return buf[:2], nil
	}

	n := binary.PutVarint(buf[2:], int64(o))
	buf[1] = byte(n)
	return buf[:2+n], nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Int) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binIntV1 {
		return errors.New("invalid ugo.Int data")
	}

	size := int(data[1])
	if size <= 0 {
		return nil
	}

	if len(data) < 2+size {
		return errors.New("invalid ugo.Int data size")
	}

	v, n := binary.Varint(data[2:])
	if n < 1 {
		if n == 0 {
			return errors.New("ugo.Int data buffer too small")
		}
		return errors.New("ugo.Int value larger than 64 bits")
	}

	*o = Int(v)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o Uint) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 2+binary.MaxVarintLen64)
	buf[0] = binUintV1
	if o == 0 {
		buf[1] = 0
		return buf[:2], nil
	}

	n := binary.PutUvarint(buf[2:], uint64(o))
	buf[1] = byte(n)
	return buf[:2+n], nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Uint) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binUintV1 {
		return errors.New("invalid ugo.Uint data")
	}

	size := int(data[1])
	if size <= 0 {
		return nil
	}

	if len(data) < 2+size {
		return errors.New("invalid ugo.Uint data size")
	}

	v, n := binary.Uvarint(data[2:])
	if n < 1 {
		if n == 0 {
			return errors.New("ugo.Uint data buffer too small")
		}
		return errors.New("ugo.Uint value larger than 64 bits")
	}

	*o = Uint(v)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o Char) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 2+binary.MaxVarintLen32)
	buf[0] = binCharV1
	if o == 0 {
		buf[1] = 0
		return buf[:2], nil
	}

	n := binary.PutVarint(buf[2:], int64(o))
	buf[1] = byte(n)
	return buf[:2+n], nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Char) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binCharV1 {
		return errors.New("invalid ugo.Char data")
	}

	size := int(data[1])
	if size <= 0 {
		return nil
	}

	if len(data) < 2+size {
		return errors.New("invalid ugo.Char data size")
	}

	v, n := binary.Varint(data[2:])
	if n < 1 {
		if n == 0 {
			return errors.New("ugo.Char data buffer too small")
		}
		return errors.New("ugo.Char value larger than 64 bits")
	}

	if int64(rune(v)) != v {
		return errors.New("ugo.Char value larger than 32 bits")
	}

	*o = Char(v)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o Float) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 2+binary.MaxVarintLen64)
	buf[0] = binFloatV1
	if o == 0 {
		buf[1] = 0
		return buf[:2], nil
	}

	n := binary.PutUvarint(buf[2:], math.Float64bits(float64(o)))
	buf[1] = byte(n)
	return buf[:2+n], nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Float) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binFloatV1 {
		return errors.New("invalid ugo.Float data")
	}

	size := int(data[1])
	if size <= 0 {
		return nil
	}

	if len(data) < 2+size {
		return errors.New("invalid ugo.Float data size")
	}

	v, n := binary.Uvarint(data[2:])
	if n < 1 {
		if n == 0 {
			return errors.New("ugo.Float data buffer too small")
		}
		return errors.New("ugo.Float value larger than 64 bits")
	}

	*o = Float(math.Float64frombits(v))
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o String) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(binStringV1)
	size := int64(len(o))

	if size == 0 {
		buf.WriteByte(0)
		return buf.Bytes(), nil
	}

	var vi varintConv
	b := vi.toBytes(size)
	buf.Write(b)
	buf.WriteString(string(o))
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *String) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binStringV1 {
		return errors.New("invalid ugo.String data")
	}

	size, offset, err := toVarint(data[1:])
	if err != nil {
		return err
	}

	if size <= 0 {
		return nil
	}

	ub := 1 + offset + int(size)
	if len(data) < ub {
		return errors.New("invalid ugo.String data size")
	}

	*o = String(data[1+offset : ub])
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o Bytes) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(binBytesV1)
	size := int64(len(o))

	if size == 0 {
		buf.WriteByte(0)
		return buf.Bytes(), nil
	}

	var vi varintConv
	b := vi.toBytes(size)
	buf.Write(b)
	buf.Write(o)
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Bytes) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binBytesV1 {
		return errors.New("invalid ugo.Bytes data")
	}

	size, offset, err := toVarint(data[1:])
	if err != nil {
		return err
	}

	if size <= 0 {
		return nil
	}

	ub := 1 + offset + int(size)
	if len(data) < ub {
		return errors.New("invalid ugo.Bytes data size")
	}

	*o = []byte(string(data[1+offset : ub]))
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o Array) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(binArrayV1)
	if len(o) == 0 {
		buf.WriteByte(0)
		return buf.Bytes(), nil
	}

	var tmpBuf bytes.Buffer
	var vi varintConv
	b := vi.toBytes(int64(len(o)))
	tmpBuf.Write(b)

	for _, v := range o {
		if m := marshaler(v); m != nil {
			d, err := m.MarshalBinary()
			if err != nil {
				return nil, err
			}
			tmpBuf.Write(d)
		} else {
			tmpBuf.WriteByte(binUnkownType)
			if err := gob.NewEncoder(&tmpBuf).Encode(&v); err != nil {
				return nil, err
			}
		}
	}

	b = vi.toBytes(int64(tmpBuf.Len()))
	buf.Write(b)
	buf.Write(tmpBuf.Bytes())
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Array) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binArrayV1 {
		return errors.New("invalid ugo.Array data")
	}

	size, offset, err := toVarint(data[1:])
	if err != nil {
		return err
	}

	if size <= 0 {
		return nil
	}
	ub := 1 + offset + int(size)
	if len(data) < ub {
		return errors.New("invalid ugo.Array data size")
	}

	rd := bytes.NewReader(data[1+offset : ub])
	var vi varintConv
	vi.reader = rd

	length, err := vi.read()
	if err != nil {
		return err
	}

	arr := make([]ugo.Object, 0, int(length))
	for rd.Len() > 0 {
		o, err := DecodeObject(rd)
		if err != nil {
			return err
		}
		arr = append(arr, o)
	}

	*o = arr
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o Map) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(binMapV1)

	var tmpBuf bytes.Buffer
	var vi varintConv

	for k, v := range o {
		b := vi.toBytes(int64(len(k)))
		tmpBuf.Write(b)
		tmpBuf.WriteString(k)

		if m := marshaler(v); m != nil {
			d, err := m.MarshalBinary()
			if err != nil {
				return nil, err
			}
			tmpBuf.Write(d)
		} else {
			tmpBuf.WriteByte(binUnkownType)
			if err := gob.NewEncoder(&tmpBuf).Encode(&v); err != nil {
				return nil, err
			}
		}
	}

	b := vi.toBytes(int64(tmpBuf.Len()))
	buf.Write(b)
	buf.Write(tmpBuf.Bytes())
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Map) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binMapV1 {
		return errors.New("invalid ugo.Map data")
	}

	size, offset, err := toVarint(data[1:])
	if err != nil {
		return err
	}

	if size <= 0 {
		return nil
	}

	if len(data) < 1+offset+int(size) {
		return errors.New("invalid ugo.Map data size")
	}

	rd := bytes.NewReader(data[1+offset : 1+offset+int(size)])
	strBuf := bytes.NewBuffer(nil)
	var vi varintConv
	vi.reader = rd
	m := *o

	for rd.Len() > 0 {
		value, err := vi.read()
		if err != nil {
			return err
		}

		var k string
		if value > 0 {
			strBuf.Reset()
			if _, err = io.CopyN(strBuf, rd, value); err != nil {
				return err
			}
			k = strBuf.String()
		}

		o, err := DecodeObject(rd)
		if err != nil {
			return err
		}
		m[k] = o
	}
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o *SyncMap) MarshalBinary() ([]byte, error) {
	(*ugo.SyncMap)(o).RLock()
	defer (*ugo.SyncMap)(o).RUnlock()

	var buf bytes.Buffer
	if o.Value == nil {
		buf.WriteByte(binSyncMapV1)
		buf.WriteByte(0)
		return buf.Bytes(), nil
	}

	b, err := Map(o.Value).MarshalBinary()
	if err != nil {
		return nil, err
	}

	if len(b) > 0 {
		b[0] = binSyncMapV1
	}
	return b, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *SyncMap) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binSyncMapV1 {
		return errors.New("invalid ugo.SyncMap data")
	}

	if data[1] == 0 {
		return nil
	}

	data[0] = binMapV1
	m := Map{}
	if err := m.UnmarshalBinary(data); err != nil {
		data[0] = binSyncMapV1
		return err
	}

	data[0] = binSyncMapV1
	o.Value = (ugo.Map)(m)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o *CompiledFunction) MarshalBinary() ([]byte, error) {
	var tmpBuf bytes.Buffer
	var vi varintConv
	if o.NumParams > 0 {
		// NumParams field #0
		tmpBuf.WriteByte(0)
		b := vi.toBytes(int64(o.NumParams))
		tmpBuf.Write(b)
	}

	if o.NumLocals > 0 {
		// NumLocals field #1
		tmpBuf.WriteByte(1)
		b := vi.toBytes(int64(o.NumLocals))
		tmpBuf.Write(b)
	}

	if o.Instructions != nil {
		// Instructions field #2
		tmpBuf.WriteByte(2)
		data, err := Bytes(o.Instructions).MarshalBinary()
		if err != nil {
			return nil, err
		}
		tmpBuf.Write(data)
	}

	// Variadic field #3
	if o.Variadic {
		tmpBuf.WriteByte(3)
	}

	// Free field #4, ignore Free variables, doesn't make sense
	if o.SourceMap != nil {
		// SourceMap field #5
		tmpBuf.WriteByte(5)
		b := vi.toBytes(int64(len(o.SourceMap) * 2))
		tmpBuf.Write(b)
		for key, value := range o.SourceMap {
			b = vi.toBytes(int64(key))
			tmpBuf.Write(b)
			b = vi.toBytes(int64(value))
			tmpBuf.Write(b)
		}
	}

	var buf bytes.Buffer
	size := vi.toBytes(int64(tmpBuf.Len()))
	buf.WriteByte(binCompiledFunctionV1)
	buf.Write(size)
	buf.Write(tmpBuf.Bytes())
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *CompiledFunction) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binCompiledFunctionV1 {
		return errors.New("invalid ugo.CompiledFunction data")
	}

	size, offset, err := toVarint(data[1:])
	if err != nil {
		return err
	}

	if size <= 0 {
		return nil
	}

	rd := bytes.NewReader(data[1+offset : 1+offset+int(size)])
	var vi varintConv
	vi.reader = rd

	for rd.Len() > 0 {
		field, err := rd.ReadByte()
		if err != nil {
			return err
		}
		switch field {
		case 0:
			v, err := vi.read()
			if err != nil {
				return err
			}
			o.NumParams = int(v)
		case 1:
			v, err := vi.read()
			if err != nil {
				return err
			}
			o.NumLocals = int(v)
		case 2:
			obj, err := DecodeObject(rd)
			if err != nil {
				return err
			}
			o.Instructions = obj.(ugo.Bytes)
		case 3:
			o.Variadic = true
		case 4:
			return errors.New("unexpected field #4")
		case 5:
			length, err := vi.read()
			if err != nil {
				return err
			}

			sz := int(length / 2)
			// always put size to the map to decode faster
			o.SourceMap = make(map[int]int, sz)
			for i := 0; i < sz; i++ {
				key, err := vi.read()
				if err != nil {
					return err
				}
				value, err := vi.read()
				if err != nil {
					return err
				}
				o.SourceMap[int(key)] = int(value)
			}
		default:
			return errors.New("unknown field:" + strconv.Itoa(int(field)))
		}
	}
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o *BuiltinFunction) MarshalBinary() ([]byte, error) {
	// Note: use string name instead of index of builtin
	s, err := String(o.Name).MarshalBinary()
	if err != nil {
		return nil, err
	}

	var vi varintConv
	b := vi.toBytes(int64(len(s)))
	data := make([]byte, 0, 1+len(b)+len(s))
	data = append(data, binBuiltinFunctionV1)
	data = append(data, b...)
	data = append(data, s...)
	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *BuiltinFunction) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binBuiltinFunctionV1 {
		return errors.New("invalid ugo.BuiltinFunction data")
	}

	size, offset, err := toVarint(data[1:])
	if err != nil {
		return err
	}

	if size <= 0 {
		return errors.New("invalid ugo.BuiltinFunction data size")
	}

	var s String
	if err := s.UnmarshalBinary(data[1+offset:]); err != nil {
		return err
	}

	index, ok := ugo.BuiltinsMap[string(s)]
	if !ok {
		return fmt.Errorf("builtin '%s' not found", s)
	}

	obj := ugo.BuiltinObjects[index]
	f, ok := obj.(*BuiltinFunction)
	if ok {
		*o = *f
		return nil
	}
	return fmt.Errorf("builtin '%s' not a ugo.BuiltinFunction type", s)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o *Function) MarshalBinary() ([]byte, error) {
	s, err := String(o.Name).MarshalBinary()
	if err != nil {
		return nil, err
	}

	var vi varintConv
	b := vi.toBytes(int64(len(s)))
	data := make([]byte, 0, 1+len(b)+len(s))
	data = append(data, binFunctionV1)
	data = append(data, b...)
	data = append(data, s...)
	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Function) UnmarshalBinary(data []byte) error {
	if len(data) < 2 || data[0] != binFunctionV1 {
		return errors.New("invalid ugo.Function data")
	}

	size, offset, err := toVarint(data[1:])
	if err != nil {
		return err
	}

	if size <= 0 {
		return errors.New("invalid ugo.Function data size")
	}

	var s String
	if err := s.UnmarshalBinary(data[1+offset:]); err != nil {
		return err
	}
	o.Name = string(s)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (sf *SourceFile) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	d, err := String(sf.Name).MarshalBinary()
	if err != nil {
		return nil, err
	}

	buf.Write(d)
	var vi varintConv
	b := vi.toBytes(int64(sf.Base))
	buf.Write(b)

	b = vi.toBytes(int64(sf.Size))
	buf.Write(b)

	b = vi.toBytes(int64(len(sf.Lines)))
	buf.Write(b)

	for _, v := range sf.Lines {
		b = vi.toBytes(int64(v))
		buf.Write(b)
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (sf *SourceFile) UnmarshalBinary(data []byte) error {
	rd := bytes.NewReader(data)

	obj, err := DecodeObject(rd)
	if err != nil {
		return err
	}

	sf.Name = obj.String()
	var vi varintConv
	vi.reader = rd
	v, err := vi.read()
	if err != nil {
		return err
	}

	sf.Base = int(v)

	v, err = vi.read()
	if err != nil {
		return err
	}

	sf.Size = int(v)

	v, err = vi.read()
	if err != nil {
		return err
	}

	length := int(v)

	lines := make([]int, length)
	for i := 0; i < length; i++ {
		v, err = vi.read()
		if err != nil {
			return err
		}
		lines[i] = int(v)
	}

	if rd.Len() > 0 {
		return errors.New("unread bytes")
	}

	sf.Lines = lines
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (sfs *SourceFileSet) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	var vi varintConv
	b := vi.toBytes(int64(sfs.Base))
	buf.Write(b)

	b = vi.toBytes(int64(len(sfs.Files)))
	buf.Write(b)

	for _, v := range sfs.Files {
		if v == nil {
			continue
		}
		d, err := (*SourceFile)(v).MarshalBinary()
		if err != nil {
			return nil, err
		}
		b := vi.toBytes(int64(len(d)))
		buf.Write(b)
		buf.Write(d)
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (sfs *SourceFileSet) UnmarshalBinary(data []byte) error {
	rd := bytes.NewReader(data)
	var vi varintConv
	vi.reader = rd
	v, err := vi.read()
	if err != nil {
		return err
	}

	sfs.Base = int(v)

	v, err = vi.read()
	if err != nil {
		return err
	}

	length := int(v)
	files := make([]*parser.SourceFile, length)

	for i := 0; i < length; i++ {
		v, err = vi.read()
		if err != nil {
			return err
		}
		data := make([]byte, v)
		if _, err = io.ReadFull(rd, data); err != nil {
			return err
		}
		var file SourceFile
		if err = file.UnmarshalBinary(data); err != nil {
			return err
		}
		files[i] = (*parser.SourceFile)(&file)
	}

	if rd.Len() > 0 {
		return errors.New("unread bytes")
	}

	sfs.Files = files
	return nil
}

func readByteFrom(r io.Reader) (byte, error) {
	if br, ok := r.(io.ByteReader); ok {
		return br.ReadByte()
	}

	var one = []byte{0}
	n, err := r.Read(one)
	if err != nil {
		if err == io.EOF {
			if n == 1 {
				return one[0], nil
			}
		}
		return 0, err
	}

	if n == 1 {
		return one[0], nil
	}
	return 0, errors.New("byte read error")
}

func writeByteTo(w io.Writer, b byte) error {
	if bw, ok := w.(io.ByteWriter); ok {
		return bw.WriteByte(b)
	}

	n, err := w.Write([]byte{b})
	if err != nil {
		return err
	}

	if n != 1 {
		return errors.New("byte write error")
	}
	return nil
}

type varintConv struct {
	buf    [1 + binary.MaxVarintLen64]byte
	reader *bytes.Reader
}

func (vi *varintConv) toBytes(v int64) []byte {
	n := binary.PutVarint(vi.buf[1:], v)
	vi.buf[0] = byte(n)
	return vi.buf[:n+1]
}

func (vi *varintConv) read() (value int64, err error) {
	var n byte
	n, err = vi.reader.ReadByte()
	if err != nil {
		return
	}

	if int(n) > len(vi.buf) {
		return 0, errVarintOverflow
	}

	data := vi.buf[:n]
	if n == 0 {
		return
	}

	if _, err = io.ReadFull(vi.reader, data); err != nil {
		return
	}

	var offset int
	value, offset = binary.Varint(data)
	if offset < 1 {
		if offset == 0 {
			err = errVarintTooSmall
			return
		}
		err = errVarintOverflow
		return
	}
	return
}

func (vi *varintConv) readBytes(r io.Reader) (value int64, readBytes []byte, err error) {
	var n byte
	n, err = readByteFrom(r)
	if err != nil {
		return
	}

	if 1+int(n) > len(vi.buf) {
		return 0, nil, errVarintOverflow
	}

	readBytes = vi.buf[:1+n]
	readBytes[0] = n
	if n == 0 {
		return
	}

	if _, err = io.ReadFull(r, readBytes[1:]); err != nil {
		return
	}

	var offset int
	value, offset = binary.Varint(readBytes[1:])
	if offset < 1 {
		if offset == 0 {
			err = errVarintTooSmall
			return
		}
		err = errVarintOverflow
		return
	}
	return
}

// toVarint converts a byte slice to int64. If length of slice is 0, it panics.
func toVarint(data []byte) (value int64, offset int, err error) {
	size := int(data[0])
	if size == 0 {
		offset = 1
		return
	}

	if len(data) < 1+size {
		err = errVarintTooSmall
		return
	}

	value, offset = binary.Varint(data[1:])
	if offset < 1 {
		if offset == 0 {
			err = errVarintTooSmall
			return
		}
		err = errVarintOverflow
		return
	}

	offset++
	return
}

func marshaler(o ugo.Object) encoding.BinaryMarshaler {
	switch v := o.(type) {
	case ugo.Bool:
		return Bool(v)
	case ugo.Int:
		return Int(v)
	case ugo.Uint:
		return Uint(v)
	case ugo.Char:
		return Char(v)
	case ugo.Float:
		return Float(v)
	case ugo.String:
		return String(v)
	case ugo.Bytes:
		return Bytes(v)
	case ugo.Array:
		return Array(v)
	case ugo.Map:
		return Map(v)
	case *ugo.SyncMap:
		return (*SyncMap)(v)
	case *ugo.CompiledFunction:
		return (*CompiledFunction)(v)
	case *ugo.Function:
		return (*Function)(v)
	case *ugo.BuiltinFunction:
		return (*BuiltinFunction)(v)
	case *ugo.UndefinedType:
		return (*UndefinedType)(v)
	default:
		return nil
	}
}
