// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
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

// Encode writes encoded data of Bytecode to writer.
func (bc *Bytecode) Encode(w io.Writer) error {
	data, err := bc.MarshalBinary()
	if err != nil {
		return err
	}
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("short write")
	}
	return nil
}

// Decode decodes Bytecode data from the reader.
func (bc *Bytecode) Decode(r io.Reader, modules *ModuleMap, tmpBuf []byte) error {
	dst := bytes.NewBuffer(tmpBuf)
	if _, err := io.Copy(dst, r); err != nil {
		return err
	}
	return bc.Unmarshal(dst.Bytes(), modules)
}

// Unmarshal unmarshals data and assigns receiver to the new Bytecode.
func (bc *Bytecode) Unmarshal(data []byte, modules *ModuleMap) error {
	err := bc.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	if modules == nil {
		modules = NewModuleMap()
	}
	return bc.fixObjects(modules)
}

func (bc *Bytecode) fixObjects(modules *ModuleMap) error {
	const moduleNameKey = "__module_name__"
	for i := range bc.Constants {
		switch obj := bc.Constants[i].(type) {
		case Map:
			if v, ok := obj[moduleNameKey]; ok {
				name, ok := v.(String)
				if !ok {
					continue
				}
				bmod := modules.GetBuiltinModule(string(name))
				if bmod == nil {
					return fmt.Errorf("module '%s' not found", name)
				}
				// copy items from given module to decoded object if key exists in obj
				for item := range obj {
					if item == moduleNameKey {
						// module name may not present in given map, skip it.
						continue
					}
					o := bmod.Attrs[item]
					// if item not exists in module, nil will not pass type check
					want := reflect.TypeOf(obj[item])
					got := reflect.TypeOf(o)
					if want != got {
						// this must not happen
						return fmt.Errorf("module '%s' item '%s' type mismatch:"+
							"want '%v', got '%v'", name, item, want, got)
					}
					obj[item] = o
				}
			}
			continue
		case *Function:
			return fmt.Errorf("Function type not decodable:'%s'", obj.Name)
		}
	}
	return nil
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
		} else {
			_, _ = fmt.Fprintf(w, "%4d: %#v|%s\n", i,
				bc.Constants[i], bc.Constants[i].TypeName())
		}
	}
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
	return "compiled-function"
}

func (o *CompiledFunction) String() string {
	return "<compiled-function>"
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

// CanIterate implements Object interface
func (o *CompiledFunction) CanIterate() bool { return false }

// Iterate implements Object interface
func (*CompiledFunction) Iterate() Iterator { return nil }

// IndexGet represents string values and implements Object interface.
func (*CompiledFunction) IndexGet(index Object) (Object, error) {
	return nil, ErrNotIndexable
}

// IndexSet implements Object interface.
func (*CompiledFunction) IndexSet(index, value Object) error {
	return ErrNotIndexAssignable
}

// CanCall implements Object interface
func (o *CompiledFunction) CanCall() bool { return true }

// Call implements Object interface
func (o *CompiledFunction) Call(...Object) (Object, error) {
	return Undefined, nil
}

// BinaryOp implements Object interface
func (o *CompiledFunction) BinaryOp(token.Token, Object) (Object, error) {
	return nil, ErrInvalidOperator
}

// IsFalsy implements Object interface
func (o *CompiledFunction) IsFalsy() bool { return false }

// Equal implements Object interface
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
	_, _ = fmt.Fprintf(w, "Params:%d Variadic:%t Locals:%d\n",
		o.NumParams, o.Variadic, o.NumLocals)
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
