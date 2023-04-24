// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ozanh/ugo/parser"
	"github.com/ozanh/ugo/token"
)

// Bytecode holds the compiled functions and constants.
type Bytecode struct {
	FileSet    *parser.SourceFileSet
	Main       *CompiledFunction
	Constants  []Object
	NumModules int
}

// Fprint writes constants and instructions to given Writer in a human readable form.
func (bc *Bytecode) Fprint(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Bytecode")
	_, _ = fmt.Fprintf(w, "Modules:%d\n", bc.NumModules)
	bc.putConstants(w)
	bc.Main.Fprint(w)
}

func (bc *Bytecode) String() string {
	var buf bytes.Buffer
	bc.Fprint(&buf)
	return buf.String()
}

func (bc *Bytecode) putConstants(w io.Writer) {
	_, _ = fmt.Fprintf(w, "Constants:\n")
	for i := range bc.Constants {
		if cf, ok := bc.Constants[i].(*CompiledFunction); ok {
			_, _ = fmt.Fprintf(w, "%4d: CompiledFunction\n", i)

			var b bytes.Buffer
			cf.Fprint(&b)

			_, _ = fmt.Fprint(w, "\t")

			str := b.String()
			c := strings.Count(str, "\n")
			_, _ = fmt.Fprint(w, strings.Replace(str, "\n", "\n\t", c-1))
			continue
		}
		_, _ = fmt.Fprintf(w, "%4d: %#v|%s\n",
			i, bc.Constants[i], bc.Constants[i].TypeName())
	}
}

// CompiledFunction holds the constants and instructions to pass VM.
type CompiledFunction struct {
	// number of parameters
	NumParams int
	// number of local variabls including parameters NumLocals>=NumParams
	NumLocals    int
	Instructions []byte
	Variadic     bool
	Free         []*ObjectPtr
	// SourceMap holds the index of instruction and token's position.
	SourceMap map[int]int
}

var _ Object = (*CompiledFunction)(nil)

// TypeName implements Object interface
func (*CompiledFunction) TypeName() string {
	return "compiledFunction"
}

func (o *CompiledFunction) String() string {
	return "<compiledFunction>"
}

// Copy implements the Copier interface.
func (o *CompiledFunction) Copy() Object {
	var insts []byte
	if o.Instructions != nil {
		insts = make([]byte, len(o.Instructions))
		copy(insts, o.Instructions)
	}

	var free []*ObjectPtr
	if o.Free != nil {
		// DO NOT Copy() elements; these are variable pointers
		free = make([]*ObjectPtr, len(o.Free))
		copy(free, o.Free)
	}

	var sourceMap map[int]int
	if o.SourceMap != nil {
		sourceMap = make(map[int]int, len(o.SourceMap))
		for k, v := range o.SourceMap {
			sourceMap[k] = v
		}
	}

	return &CompiledFunction{
		NumParams:    o.NumParams,
		NumLocals:    o.NumLocals,
		Instructions: insts,
		Variadic:     o.Variadic,
		Free:         free,
		SourceMap:    sourceMap,
	}
}

// CanIterate implements Object interface.
func (*CompiledFunction) CanIterate() bool { return false }

// Iterate implements Object interface.
func (*CompiledFunction) Iterate() Iterator { return nil }

// IndexGet represents string values and implements Object interface.
func (*CompiledFunction) IndexGet(index Object) (Object, error) {
	return nil, ErrNotIndexable
}

// IndexSet implements Object interface.
func (*CompiledFunction) IndexSet(index, value Object) error {
	return ErrNotIndexAssignable
}

// CanCall implements Object interface.
func (*CompiledFunction) CanCall() bool { return true }

// Call implements Object interface. CompiledFunction is not directly callable.
// You should use Invoker to call it with a Virtual Machine. Because of this, it
// always returns an error.
func (*CompiledFunction) Call(...Object) (Object, error) {
	return Undefined, ErrNotCallable
}

// BinaryOp implements Object interface.
func (*CompiledFunction) BinaryOp(token.Token, Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// IsFalsy implements Object interface.
func (*CompiledFunction) IsFalsy() bool { return false }

// Equal implements Object interface.
func (o *CompiledFunction) Equal(right Object) bool {
	v, ok := right.(*CompiledFunction)
	return ok && o == v
}

// SourcePos returns the source position of the instruction at ip.
func (o *CompiledFunction) SourcePos(ip int) parser.Pos {
begin:
	if ip >= 0 {
		if p, ok := o.SourceMap[ip]; ok {
			return parser.Pos(p)
		}
		ip--
		goto begin
	}
	return parser.NoPos
}

// Fprint writes constants and instructions to given Writer in a human readable form.
func (o *CompiledFunction) Fprint(w io.Writer) {
	_, _ = fmt.Fprintf(w, "Params:%d Variadic:%t Locals:%d\n", o.NumParams, o.Variadic, o.NumLocals)
	_, _ = fmt.Fprintf(w, "Instructions:\n")

	i := 0
	var operands []int

	for i < len(o.Instructions) {

		op := o.Instructions[i]
		numOperands := OpcodeOperands[op]
		operands, offset := ReadOperands(numOperands, o.Instructions[i+1:], operands)
		_, _ = fmt.Fprintf(w, "%04d %-12s", i, OpcodeNames[op])

		if len(operands) > 0 {
			for _, r := range operands {
				_, _ = fmt.Fprint(w, "    ", strconv.Itoa(r))
			}
		}

		_, _ = fmt.Fprintln(w)
		i += offset + 1
	}

	if o.Free != nil {
		_, _ = fmt.Fprintf(w, "Free:%v\n", o.Free)
	}
	_, _ = fmt.Fprintf(w, "SourceMap:%v\n", o.SourceMap)
}

func (o *CompiledFunction) identical(other *CompiledFunction) bool {
	if o.NumParams != other.NumParams ||
		o.NumLocals != other.NumLocals ||
		o.Variadic != other.Variadic ||
		len(o.Instructions) != len(other.Instructions) ||
		len(o.Free) != len(other.Free) ||
		string(o.Instructions) != string(other.Instructions) {
		return false
	}
	for i := range o.Free {
		if o.Free[i].Equal(other.Free[i]) {
			return false
		}
	}
	return true
}

func (o *CompiledFunction) equalSourceMap(other *CompiledFunction) bool {
	if len(o.SourceMap) != len(other.SourceMap) {
		return false
	}
	for k, v := range o.SourceMap {
		vv, ok := other.SourceMap[k]
		if !ok || vv != v {
			return false
		}
	}
	return true
}

func (o *CompiledFunction) hash32() uint32 {
	hash := hashData32(2166136261, []byte{byte(o.NumParams)})
	hash = hashData32(hash, []byte{byte(o.NumLocals)})
	if o.Variadic {
		hash = hashData32(hash, []byte{1})
	} else {
		hash = hashData32(hash, []byte{0})
	}
	hash = hashData32(hash, o.Instructions)
	return hash
}

func hashData32(hash uint32, data []byte) uint32 {
	for _, c := range data {
		hash *= 16777619 // prime32
		hash ^= uint32(c)
	}
	return hash
}
