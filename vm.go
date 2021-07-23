// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/ozanh/ugo/parser"
	"github.com/ozanh/ugo/token"
)

const (
	stackSize = 2048
	frameSize = 1024
)

// VM executes the instructions in Bytecode.
type VM struct {
	abort        int64
	sp           int
	ip           int
	curInsts     []byte
	constants    []Object
	stack        [stackSize]Object
	frames       [frameSize]frame
	curFrame     *frame
	frameIndex   int
	bytecode     *Bytecode
	modulesCache []Object
	mu           sync.Mutex
	err          error
	noPanic      bool
}

// NewVM creates a VM object.
func NewVM(bc *Bytecode) *VM {
	var constants []Object
	if bc != nil {
		constants = bc.Constants
	}
	vm := &VM{
		bytecode:  bc,
		constants: constants,
	}
	return vm
}

// SetRecover recovers panic when Run panics and returns panic as an error.
// If error handler is present `try-catch-finally`, VM continues to run from catch/finally.
func (vm *VM) SetRecover(v bool) *VM {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.noPanic = v
	return vm
}

// SetBytecode enables to set a new Bytecode.
func (vm *VM) SetBytecode(bc *Bytecode) *VM {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.bytecode = bc
	vm.constants = bc.Constants
	vm.modulesCache = nil
	return vm
}

// Clear clears stack by setting nil to stack indexes and removes modules cache.
func (vm *VM) Clear() *VM {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	for i := range vm.stack {
		vm.stack[i] = nil
	}
	vm.modulesCache = nil
	return vm
}

// GetLocals returns variables from stack up to the NumLocals of given Bytecode.
// This must be called after Run() before Clear().
func (vm *VM) GetLocals(locals []Object) []Object {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if locals != nil {
		locals = locals[:0]
	} else {
		locals = make([]Object, 0, vm.bytecode.Main.NumLocals)
	}

	for i := range vm.stack[:vm.bytecode.Main.NumLocals] {
		locals = append(locals, vm.stack[i])
	}

	return locals
}

// RunCompiledFunction runs given CompiledFunction as if it is Main function.
// Bytecode must be set before calling this method, because Fileset and Constants are copied.
func (vm *VM) RunCompiledFunction(
	f *CompiledFunction,
	globals Object,
	args ...Object,
) (Object, error) {
	vm.mu.Lock()
	if vm.bytecode == nil {
		vm.mu.Unlock()
		return nil, errors.New("invalid Bytecode")
	}

	vm.bytecode = &Bytecode{
		FileSet:    vm.bytecode.FileSet,
		Constants:  vm.constants,
		Main:       f,
		NumModules: vm.bytecode.NumModules,
	}

	for i := range vm.stack {
		vm.stack[i] = nil
	}

	vm.mu.Unlock()
	return vm.Run(globals, args...)
}

// Abort aborts the VM execution.
func (vm *VM) Abort() {
	atomic.StoreInt64(&vm.abort, 1)
}

// Run runs VM and executes the instructions until the OpReturn Opcode or Abort call.
func (vm *VM) Run(globals Object, args ...Object) (Object, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.bytecode == nil || vm.bytecode.Main == nil {
		return nil, errors.New("invalid Bytecode")
	}

	vm.err = nil
	atomic.StoreInt64(&vm.abort, 0)
	vm.initLocals(args)
	vm.initCurFrame()
	vm.frameIndex = 1
	vm.ip = -1
	vm.sp = vm.curFrame.fn.NumLocals

	// Resize modules cache or create it if not exists.
	// Note that REPL can set module cache before running, don't recreate it add missing indexes.
	if len(vm.modulesCache) < vm.bytecode.NumModules {
		diff := vm.bytecode.NumModules - len(vm.modulesCache)
		for i := 0; i < diff; i++ {
			vm.modulesCache = append(vm.modulesCache, nil)
		}
	}

	func() {
		defer func() {
			if vm.noPanic {
				if r := recover(); r != nil {
					vm.handlePanic(r, globals)
				}
			}
			vm.clearCurFrame()
		}()
		if globals == nil {
			globals = Map{}
		}
		vm.run(globals)
	}()

	if vm.err != nil {
		return nil, vm.err
	}

	if vm.sp < stackSize {
		if vv, ok := vm.stack[vm.sp-1].(*ObjectPtr); ok {
			return *vv.Value, nil
		}
		return vm.stack[vm.sp-1], nil
	}
	return nil, ErrStackOverflow
}

func (vm *VM) run(globals Object) {
VMLoop:
	for atomic.LoadInt64(&vm.abort) == 0 {
		vm.ip++
		switch vm.curInsts[vm.ip] {
		case OpConstant:
			cidx := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			obj := vm.constants[cidx]
			vm.stack[vm.sp] = obj
			vm.sp++
			vm.ip += 2
		case OpGetLocal:
			localIdx := int(vm.curInsts[vm.ip+1])
			value := vm.stack[vm.curFrame.basePointer+localIdx]
			if v, ok := value.(*ObjectPtr); ok {
				value = *v.Value
			}
			vm.stack[vm.sp] = value
			vm.sp++
			vm.ip++
		case OpSetLocal:
			localIndex := int(vm.curInsts[vm.ip+1])
			value := vm.stack[vm.sp-1]
			index := vm.curFrame.basePointer + localIndex
			if v, ok := vm.stack[index].(*ObjectPtr); ok {
				*v.Value = value
			} else {
				vm.stack[index] = value
			}
			vm.sp--
			vm.stack[vm.sp] = nil
			vm.ip++
		case OpBinaryOp:
			tok := token.Token(vm.curInsts[vm.ip+1])
			left, right := vm.stack[vm.sp-2], vm.stack[vm.sp-1]

			var value Object
			var err error
			switch left := left.(type) {
			case Int:
				value, err = left.BinaryOp(tok, right)
			case String:
				value, err = left.BinaryOp(tok, right)
			case Float:
				value, err = left.BinaryOp(tok, right)
			case Uint:
				value, err = left.BinaryOp(tok, right)
			case Char:
				value, err = left.BinaryOp(tok, right)
			case Bool:
				value, err = left.BinaryOp(tok, right)
			default:
				value, err = left.BinaryOp(tok, right)
			}
			if err == nil {
				vm.stack[vm.sp-2] = value
				vm.sp--
				vm.stack[vm.sp] = nil
				vm.ip++
				continue
			}
			if err == ErrInvalidOperator {
				err = ErrInvalidOperator.NewError(tok.String())
			}
			if err = vm.throwGenErr(err); err != nil {
				vm.err = err
				return
			}
		case OpAndJump:
			if vm.stack[vm.sp-1].IsFalsy() {
				pos := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
				vm.ip = pos - 1
				continue
			}
			vm.stack[vm.sp-1] = nil
			vm.sp--
			vm.ip += 2
		case OpOrJump:
			if vm.stack[vm.sp-1].IsFalsy() {
				vm.stack[vm.sp-1] = nil
				vm.sp--
				vm.ip += 2
				continue
			}
			pos := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			vm.ip = pos - 1
		case OpEqual:
			left, right := vm.stack[vm.sp-2], vm.stack[vm.sp-1]

			switch left := left.(type) {
			case Int:
				vm.stack[vm.sp-2] = Bool(left.Equal(right))
			case String:
				vm.stack[vm.sp-2] = Bool(left.Equal(right))
			case Float:
				vm.stack[vm.sp-2] = Bool(left.Equal(right))
			case Bool:
				vm.stack[vm.sp-2] = Bool(left.Equal(right))
			case Uint:
				vm.stack[vm.sp-2] = Bool(left.Equal(right))
			case Char:
				vm.stack[vm.sp-2] = Bool(left.Equal(right))
			default:
				vm.stack[vm.sp-2] = Bool(left.Equal(right))
			}
			vm.sp--
			vm.stack[vm.sp] = nil
		case OpNotEqual:
			left, right := vm.stack[vm.sp-2], vm.stack[vm.sp-1]

			switch left := left.(type) {
			case Int:
				vm.stack[vm.sp-2] = Bool(!left.Equal(right))
			case String:
				vm.stack[vm.sp-2] = Bool(!left.Equal(right))
			case Float:
				vm.stack[vm.sp-2] = Bool(!left.Equal(right))
			case Bool:
				vm.stack[vm.sp-2] = Bool(!left.Equal(right))
			case Uint:
				vm.stack[vm.sp-2] = Bool(!left.Equal(right))
			case Char:
				vm.stack[vm.sp-2] = Bool(!left.Equal(right))
			default:
				vm.stack[vm.sp-2] = Bool(!left.Equal(right))
			}
			vm.sp--
			vm.stack[vm.sp] = nil
		case OpTrue:
			vm.stack[vm.sp] = True
			vm.sp++
		case OpFalse:
			vm.stack[vm.sp] = False
			vm.sp++
		case OpCall:
			err := vm.execOpCall()
			if err == nil {
				continue
			}
			if err = vm.throwGenErr(err); err != nil {
				vm.err = err
				return
			}
		case OpReturn:
			numRet := vm.curInsts[vm.ip+1]
			bp := vm.curFrame.basePointer
			if bp == 0 {
				bp = vm.curFrame.fn.NumLocals + 1
			}
			if numRet == 1 {
				vm.stack[bp-1] = vm.stack[vm.sp-1]
			} else {
				vm.stack[bp-1] = Undefined
			}

			for i := vm.sp - 1; i >= bp; i-- {
				vm.stack[i] = nil
			}

			vm.sp = bp
			if vm.frameIndex == 1 {
				return
			}
			vm.clearCurFrame()
			parent := &(vm.frames[vm.frameIndex-2])
			vm.frameIndex--
			vm.ip = parent.ip
			vm.curFrame = parent
			vm.curInsts = vm.curFrame.fn.Instructions
		case OpGetBuiltin:
			builtinIndex := BuiltinType(int(vm.curInsts[vm.ip+1]))
			var builtin Object
			if builtinIndex == BuiltinGlobals {
				// let panic if somehow builtins updated
				bf := BuiltinObjects[builtinIndex].(*BuiltinFunction)
				builtin = &BuiltinFunction{
					Name: bf.Name,
					Value: func(_ ...Object) (Object, error) {
						return globals, nil
					},
				}
			} else {
				builtin = BuiltinObjects[builtinIndex]
			}
			vm.stack[vm.sp] = builtin
			vm.sp++
			vm.ip++
		case OpClosure:
			constIdx := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			fn := vm.constants[constIdx].(*CompiledFunction)
			numFree := int(vm.curInsts[vm.ip+3])
			free := make([]*ObjectPtr, numFree)
			for i := 0; i < numFree; i++ {
				switch freeVar := (vm.stack[vm.sp-numFree+i]).(type) {
				case *ObjectPtr:
					free[i] = freeVar
				default:
					temp := vm.stack[vm.sp-numFree+i]
					free[i] = &ObjectPtr{
						Value: &temp,
					}
				}
				vm.stack[vm.sp-numFree+i] = nil
			}
			vm.sp -= numFree
			newFn := &CompiledFunction{
				Instructions: fn.Instructions,
				NumParams:    fn.NumParams,
				NumLocals:    fn.NumLocals,
				Variadic:     fn.Variadic,
				SourceMap:    fn.SourceMap,
				Free:         free,
			}
			vm.stack[vm.sp] = newFn
			vm.sp++
			vm.ip += 3
		case OpJump:
			vm.ip = (int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8) - 1
		case OpJumpFalsy:
			vm.sp--
			obj := vm.stack[vm.sp]
			vm.stack[vm.sp] = nil

			var falsy bool
			switch obj := obj.(type) {
			case Bool:
				falsy = obj.IsFalsy()
			case Int:
				falsy = obj.IsFalsy()
			case Uint:
				falsy = obj.IsFalsy()
			case Float:
				falsy = obj.IsFalsy()
			case String:
				falsy = obj.IsFalsy()
			default:
				falsy = obj.IsFalsy()
			}
			if falsy {
				vm.ip = (int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8) - 1
				continue
			}
			vm.ip += 2
		case OpGetGlobal:
			cidx := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			index := vm.constants[cidx]
			var ret Object
			var err error
			ret, err = globals.IndexGet(index)

			if err != nil {
				if err := vm.throwGenErr(err); err != nil {
					vm.err = err
					return
				}
				continue
			}

			if ret == nil {
				vm.stack[vm.sp] = Undefined
			} else {
				vm.stack[vm.sp] = ret
			}

			vm.ip += 2
			vm.sp++
		case OpSetGlobal:
			cidx := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			index := vm.constants[cidx]
			value := vm.stack[vm.sp-1]

			if v, ok := value.(*ObjectPtr); ok {
				value = *v.Value
			}

			if err := globals.IndexSet(index, value); err != nil {
				if err := vm.throwGenErr(err); err != nil {
					vm.err = err
					return
				}
				continue
			}

			vm.ip += 2
			vm.sp--
			vm.stack[vm.sp] = nil
		case OpArray:
			numItems := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			arr := make(Array, numItems)
			copy(arr, vm.stack[vm.sp-numItems:vm.sp])
			vm.sp -= numItems
			vm.stack[vm.sp] = arr

			for i := vm.sp + 1; i < vm.sp+numItems+1; i++ {
				vm.stack[i] = nil
			}

			vm.sp++
			vm.ip += 2
		case OpMap:
			numItems := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			kv := make(Map)

			for i := vm.sp - numItems; i < vm.sp; i += 2 {
				key := vm.stack[i]
				value := vm.stack[i+1]
				kv[key.String()] = value
				vm.stack[i] = nil
				vm.stack[i+1] = nil
			}

			vm.sp -= numItems
			vm.stack[vm.sp] = kv
			vm.sp++
			vm.ip += 2
		case OpGetIndex:
			numSel := int(vm.curInsts[vm.ip+1])
			tp := vm.sp - 1 - numSel
			target := vm.stack[tp]
			var value Object = Undefined

			for ; numSel > 0; numSel-- {
				ptr := vm.sp - numSel
				index := vm.stack[ptr]
				vm.stack[ptr] = nil
				v, err := target.IndexGet(index)
				if err != nil {
					switch err {
					case ErrNotIndexable:
						err = ErrNotIndexable.NewError(target.TypeName())
					case ErrIndexOutOfBounds:
						err = ErrIndexOutOfBounds.NewError(index.String())
					}
					if err = vm.throwGenErr(err); err != nil {
						vm.err = err
						return
					}
					continue VMLoop
				}
				target = v
				value = v
			}

			vm.stack[tp] = value
			vm.sp = tp + 1
			vm.ip++
		case OpSetIndex:
			value := vm.stack[vm.sp-3]
			target := vm.stack[vm.sp-2]
			index := vm.stack[vm.sp-1]
			err := target.IndexSet(index, value)

			if err != nil {
				switch err {
				case ErrNotIndexAssignable:
					err = ErrNotIndexAssignable.NewError(target.TypeName())
				case ErrIndexOutOfBounds:
					err = ErrIndexOutOfBounds.NewError(index.String())
				}
				if err = vm.throwGenErr(err); err != nil {
					vm.err = err
					return
				}
				continue
			}

			vm.stack[vm.sp-3] = nil
			vm.stack[vm.sp-2] = nil
			vm.stack[vm.sp-1] = nil
			vm.sp -= 3
		case OpSliceIndex:
			err := vm.execOpSliceIndex()
			if err == nil {
				continue
			}
			if err = vm.throwGenErr(err); err != nil {
				vm.err = err
				return
			}
		case OpGetFree:
			freeIndex := int(vm.curInsts[vm.ip+1])
			vm.stack[vm.sp] = *vm.curFrame.freeVars[freeIndex].Value
			vm.sp++
			vm.ip++
		case OpSetFree:
			freeIndex := int(vm.curInsts[vm.ip+1])
			*vm.curFrame.freeVars[freeIndex].Value = vm.stack[vm.sp-1]
			vm.sp--
			vm.stack[vm.sp] = nil
			vm.ip++
		case OpGetLocalPtr:
			localIndex := int(vm.curInsts[vm.ip+1])
			var freeVar *ObjectPtr
			value := vm.stack[vm.curFrame.basePointer+localIndex]

			if obj, ok := value.(*ObjectPtr); ok {
				freeVar = obj
			} else {
				freeVar = &ObjectPtr{Value: &value}
				vm.stack[vm.curFrame.basePointer+localIndex] = freeVar
			}

			vm.stack[vm.sp] = freeVar
			vm.sp++
			vm.ip++
		case OpGetFreePtr:
			freeIndex := int(vm.curInsts[vm.ip+1])
			value := vm.curFrame.freeVars[freeIndex]
			vm.stack[vm.sp] = value
			vm.sp++
			vm.ip++
		case OpDefineLocal:
			localIndex := int(vm.curInsts[vm.ip+1])
			vm.stack[vm.curFrame.basePointer+localIndex] = vm.stack[vm.sp-1]
			vm.sp--
			vm.stack[vm.sp] = nil
			vm.ip++
		case OpNull:
			vm.stack[vm.sp] = Undefined
			vm.sp++
		case OpPop:
			vm.sp--
			vm.stack[vm.sp] = nil
		case OpIterInit:
			dst := vm.stack[vm.sp-1]

			if dst.CanIterate() {
				it := dst.Iterate()
				vm.stack[vm.sp-1] = &iteratorObject{Iterator: it}
				continue
			}

			var err error = ErrNotIterable.NewError(dst.TypeName())
			if err = vm.throwGenErr(err); err != nil {
				vm.err = err
				return
			}
		case OpIterNext:
			iterator := vm.stack[vm.sp-1]
			hasMore := iterator.(Iterator).Next()
			vm.stack[vm.sp-1] = Bool(hasMore)
		case OpIterKey:
			iterator := vm.stack[vm.sp-1]
			val := iterator.(Iterator).Key()
			vm.stack[vm.sp-1] = val
		case OpIterValue:
			iterator := vm.stack[vm.sp-1]
			val := iterator.(Iterator).Value()
			vm.stack[vm.sp-1] = val
		case OpLoadModule:
			cidx := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			midx := int(vm.curInsts[vm.ip+4]) | int(vm.curInsts[vm.ip+3])<<8
			value := vm.modulesCache[midx]

			if value == nil {
				// module cache is empty, load the object from constants
				vm.stack[vm.sp] = vm.constants[cidx]
				vm.sp++
				// load module by putting true for subsequent OpJumpFalsy
				// if module is a compiled-function it will be called and result will be stored in module cache
				// if module is not a compiled-function, copy of object will be stored in module cache
				vm.stack[vm.sp] = True
				vm.sp++
			} else {
				vm.stack[vm.sp] = value
				vm.sp++
				// no need to load the module, put false for subsequent OpJumpFalsy
				vm.stack[vm.sp] = False
				vm.sp++
			}

			vm.ip += 4
		case OpStoreModule:
			midx := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
			value := vm.stack[vm.sp-1]

			if v, ok := value.(Copier); ok {
				// store deep copy of the module if supported
				value = v.Copy()
				vm.stack[vm.sp-1] = value
			}

			vm.modulesCache[midx] = value
			vm.ip += 2
		case OpSetupTry:
			vm.execOpSetupTry()
		case OpSetupCatch:
			vm.execOpSetupCatch()
		case OpSetupFinally:
			vm.execOpSetupFinally()
		case OpThrow:
			if err := vm.execOpThrow(); err != nil {
				vm.err = err
				return
			}
		case OpFinalizer:
			upto := int(vm.curInsts[vm.ip+1])

			pos := vm.curFrame.errHandlers.findFinally(upto)
			if pos <= 0 {
				vm.ip++
				continue
			}
			// go to finally if set
			handler := vm.curFrame.errHandlers.last()
			// save current ip to come back to same position
			handler.returnTo = vm.ip
			// save current sp to come back to same position
			handler.sp = vm.sp
			// remove current error if any
			vm.curFrame.errHandlers.err = nil
			// set ip to finally's position
			vm.ip = pos - 1
		case OpUnary:
			err := vm.execOpUnary()
			if err == nil {
				continue
			}
			if err = vm.throwGenErr(err); err != nil {
				vm.err = err
				return
			}
		case OpNoOp:
		default:
			vm.err = fmt.Errorf("unknown opcode %d", vm.curInsts[vm.ip])
			return
		}
	}
	vm.err = ErrVMAborted
}

func (vm *VM) initLocals(args []Object) {
	// init all params as undefined
	numParams := vm.bytecode.Main.NumParams
	locals := vm.stack[:vm.bytecode.Main.NumLocals]

	// TODO (ozan): check why setting numParams fails some tests!
	for i := 0; i < vm.bytecode.Main.NumLocals; i++ {
		locals[i] = Undefined
	}

	// if main function is variadic, default value for last argument is empty array
	if vm.bytecode.Main.Variadic {
		vm.stack[numParams-1] = Array{}
	}

	// copy args to stack
	if numParams > 0 {
		if len(args) < numParams {
			copy(locals, args)
			return
		}

		if len(args) >= numParams {
			for i := range args[:numParams-1] {
				locals[i] = args[i]
			}

			if vm.bytecode.Main.Variadic {
				locals[numParams-1] = append(Array{}, args[numParams-1:]...)
			} else {
				locals[numParams-1] = args[numParams-1]
			}
		}
	}
}

func (vm *VM) initCurFrame() {
	// initialize frame and pointers
	vm.curInsts = vm.bytecode.Main.Instructions
	vm.curFrame = &(vm.frames[0])
	vm.curFrame.fn = vm.bytecode.Main

	if vm.curFrame.fn.Free != nil {
		// Assign free variables if exists in compiled function.
		// This is required to run compiled functions returned from VM using RunCompiledFunction().
		vm.curFrame.freeVars = vm.curFrame.fn.Free
	}

	vm.curFrame.errHandlers = nil
	vm.curFrame.basePointer = 0
}

func (vm *VM) clearCurFrame() {
	vm.curFrame.freeVars = nil
	vm.curFrame.fn = nil
	vm.curFrame.errHandlers = nil
}

func (vm *VM) handlePanic(r interface{}, globals Object) {
	if vm.sp < stackSize &&
		vm.frameIndex <= frameSize &&
		vm.err == nil {

		if err := vm.throwGenErr(fmt.Errorf("%v", r)); err != nil {
			vm.err = err
			gostack := debug.Stack()
			if vm.err != nil {
				vm.err = fmt.Errorf("panic: %v %w\nGo Stack:\n%s",
					r, vm.err, gostack)
			}
			return
		}

		vm.run(globals)
		return
	}

	gostack := debug.Stack()

	if vm.err != nil {
		vm.err = fmt.Errorf("panic: %v error: %w\nGo Stack:\n%s",
			r, vm.err, gostack)
		return
	}
	vm.err = fmt.Errorf("panic: %v\nGo Stack:\n%s", r, gostack)
}

func (vm *VM) execOpSetupTry() {
	catch := int(vm.curInsts[vm.ip+2]) | int(vm.curInsts[vm.ip+1])<<8
	finally := int(vm.curInsts[vm.ip+4]) | int(vm.curInsts[vm.ip+3])<<8

	ptrs := errHandler{
		sp:      vm.sp,
		catch:   catch,
		finally: finally,
	}

	if vm.curFrame.errHandlers == nil {
		vm.curFrame.errHandlers = &errHandlers{
			handlers: []errHandler{ptrs},
		}
	} else {
		vm.curFrame.errHandlers.handlers = append(
			vm.curFrame.errHandlers.handlers, ptrs)
	}

	vm.ip += 4
}

func (vm *VM) execOpSetupCatch() {
	var value Object = Undefined
	errHandlers := vm.curFrame.errHandlers

	if errHandlers.hasHandler() {
		// set 0 to last catch position
		hdl := errHandlers.last()
		hdl.catch = 0

		if errHandlers.err != nil {
			value = errHandlers.err
			errHandlers.err = nil
		}
	}

	vm.stack[vm.sp] = value
	vm.sp++
	//Either OpSetLocal or OpPop is generated by compiler to handle error
}

func (vm *VM) execOpSetupFinally() {
	errHandlers := vm.curFrame.errHandlers

	if errHandlers.hasHandler() {
		hdl := errHandlers.last()
		hdl.catch = 0
		hdl.finally = 0
	}
}

func (vm *VM) execOpThrow() error {
	op := vm.curInsts[vm.ip+1]
	vm.ip++

	switch op {
	case 0: // system
		errHandlers := vm.curFrame.errHandlers
		if errHandlers.hasError() {
			errHandlers.pop()
			// do not put position info to error for re-throw after finally.
			if err := vm.throw(errHandlers.err, true); err != nil {
				return err
			}
		} else if pos := errHandlers.hasReturnTo(); pos > 0 {
			// go to OpReturn if it is set
			handler := errHandlers.last()
			errHandlers.pop()
			handler.returnTo = 0

			if vm.sp >= handler.sp {
				for i := vm.sp; i >= handler.sp; i-- {
					vm.stack[i] = nil
				}
			}
			vm.sp = handler.sp
			vm.ip = pos - 1
		}
	case 1: // user
		obj := vm.stack[vm.sp-1]
		vm.stack[vm.sp-1] = nil
		vm.sp--

		if err := vm.throw(vm.newErrorFromObject(obj), false); err != nil {
			return err
		}
	default:
		return fmt.Errorf("wrong operand for OpThrow:%d", op)
	}

	return nil
}

func (vm *VM) throwGenErr(err error) error {
	if e, ok := err.(*RuntimeError); ok {
		if e.fileSet == nil {
			e.fileSet = vm.bytecode.FileSet
		}
		return vm.throw(e, true)
	} else if e, ok := err.(*Error); ok {
		return vm.throw(vm.newError(e), false)
	}

	return vm.throw(vm.newErrorFromError(err), false)
}

func (vm *VM) throw(err *RuntimeError, noTrace bool) error {
	if !noTrace {
		err.addTrace(vm.getSourcePos())
	}

	// firstly check our frame has error handler
	if vm.curFrame.errHandlers.hasHandler() {
		return vm.handleThrownError(vm.curFrame, err)
	}

	// find previous frames having error handler
	var frame *frame
	index := vm.frameIndex - 2

	for index >= 0 {
		f := &(vm.frames[index])
		err.addTrace(getFrameSourcePos(f))
		if f.errHandlers.hasHandler() {
			frame = f
			break
		}
		f.freeVars = nil
		f.fn = nil
		index--
	}

	if frame == nil || index < 0 {
		// not handled, exit
		return err
	}

	vm.frameIndex = index + 1

	if e := vm.handleThrownError(frame, err); e != nil {
		return e
	}

	vm.curFrame = frame
	vm.curFrame.fn = frame.fn
	vm.curInsts = frame.fn.Instructions

	return nil
}

func (vm *VM) handleThrownError(frame *frame, err *RuntimeError) error {
	frame.errHandlers.err = err
	handler := frame.errHandlers.last()

	// if we have catch>0 goto catch else follow finally (one of them must be set)
	if handler.catch > 0 {
		vm.ip = handler.catch - 1
	} else if handler.finally > 0 {
		vm.ip = handler.finally - 1
	} else {
		frame.errHandlers.pop()
		return vm.throw(err, false)
	}

	if vm.sp >= handler.sp {
		for i := vm.sp; i >= handler.sp; i-- {
			vm.stack[i] = nil
		}
	}

	vm.sp = handler.sp
	return nil
}

func (vm *VM) execOpCall() error {
	numArgs := int(vm.curInsts[vm.ip+1])
	expand := int(vm.curInsts[vm.ip+2]) // 0 or 1
	callee := vm.stack[vm.sp-numArgs-1]

	if compFunc, ok := callee.(*CompiledFunction); ok {
		basePointer := vm.sp - numArgs
		numLocals := compFunc.NumLocals
		numParams := compFunc.NumParams
		if expand == 0 {
			if !compFunc.Variadic {
				if numArgs != numParams {
					return ErrWrongNumArguments.NewError(
						wantEqXGotY(numParams, numArgs),
					)
				}
			} else {
				if numArgs < numParams-1 {
					// f := func(a, ...b) {}
					// f()
					return ErrWrongNumArguments.NewError(
						wantGEqXGotY(numParams-1, numArgs),
					)
				}
				if numArgs == numParams-1 {
					// f := func(a, ...b) {} // a==1 b==[]
					// f(1)
					vm.stack[basePointer+numArgs] = Array{}
				} else {
					// f := func(a, ...b) {} // a == 1  b == [] // a == 1  b == [2, 3]
					// f(1, 2) // f(1, 2, 3)
					arr := vm.stack[basePointer+numParams-1 : basePointer+numArgs]
					vm.stack[basePointer+numParams-1] = append(Array{}, arr...)
				}
			}
		} else {
			var arrSize int
			if arr, ok := vm.stack[basePointer+numArgs-1].(Array); ok {
				arrSize = len(arr)
			} else {
				return NewArgumentTypeError("last", "array",
					vm.stack[basePointer+numArgs-1].TypeName())
			}
			if compFunc.Variadic {
				if numArgs < numParams {
					// f := func(a, ...b) {}
					// f(...[1]) // f(...[1, 2])
					if arrSize+numArgs < numParams {
						// f := func(a, ...b) {}
						// f(...[])
						return ErrWrongNumArguments.NewError(
							wantGEqXGotY(numParams-1, arrSize+numArgs-1),
						)
					}
					tempBuf := make(Array, 0, arrSize+numArgs)
					tempBuf = append(tempBuf,
						vm.stack[basePointer:basePointer+numArgs-1]...)
					tempBuf = append(tempBuf,
						vm.stack[basePointer+numArgs-1].(Array)...)
					copy(vm.stack[basePointer:], tempBuf[:numParams-1])
					arr := tempBuf[numParams-1:]
					vm.stack[basePointer+numParams-1] = append(Array{}, arr...)
				} else if numArgs > numParams {
					// f := func(a, ...b) {} // a == 1  b == [2, 3]
					// f(1, 2, ...[3])
					arr := append(Array{},
						vm.stack[basePointer+numParams-1:basePointer+numArgs-1]...)
					arr = append(arr, vm.stack[basePointer+numArgs-1].(Array)...)
					vm.stack[basePointer+numParams-1] = arr
				}
			} else {
				if arrSize+numArgs-1 != numParams {
					// f := func(a, b) {}
					// f(1, ...[2, 3, 4])
					return ErrWrongNumArguments.NewError(
						wantEqXGotY(numParams, arrSize+numArgs-1),
					)
				}
				// f := func(a, b) {}
				// f(...[1, 2])
				arr := vm.stack[basePointer+numArgs-1].(Array)
				copy(vm.stack[basePointer+numArgs-1:], arr)
			}
		}

		for i := numParams; i < numLocals; i++ {
			vm.stack[basePointer+i] = Undefined
		}

		// test if it's tail-call
		if compFunc == vm.curFrame.fn { // recursion
			nextOp := vm.curInsts[vm.ip+2+1]

			if nextOp == OpReturn ||
				(nextOp == OpPop && OpReturn == vm.curInsts[vm.ip+2+2]) {
				curBp := vm.curFrame.basePointer
				copy(vm.stack[curBp:curBp+numLocals], vm.stack[basePointer:])
				newSp := vm.sp - numArgs - 1
				for i := vm.sp; i >= newSp; i-- {
					vm.stack[i] = nil
				}
				vm.sp = newSp
				vm.ip = -1                    // reset ip to beginning of the frame
				vm.curFrame.errHandlers = nil // reset error handlers if any set
				return nil
			}
		}

		frame := &(vm.frames[vm.frameIndex])
		vm.frameIndex++

		if vm.frameIndex > frameSize-1 {
			return ErrStackOverflow
		}

		frame.fn = compFunc
		frame.freeVars = compFunc.Free
		frame.errHandlers = nil
		frame.basePointer = basePointer

		vm.curFrame.ip = vm.ip + 2
		vm.curInsts = compFunc.Instructions
		vm.curFrame = frame
		vm.sp = basePointer + numLocals
		vm.ip = -1
		return nil
	}

	if !callee.CanCall() {
		return ErrNotCallable.NewError(callee.TypeName())
	}

	args := make([]Object, numArgs-expand, numArgs)
	k := copy(args, vm.stack[vm.sp-numArgs:vm.sp-expand])

	if expand > 0 {
		switch obj := vm.stack[vm.sp-1].(type) {
		case Array:
			args = append(args[:k], obj...)
		default:
			return NewArgumentTypeError("last", "array",
				vm.stack[vm.sp-1].TypeName())
		}
	}

	for i := 0; i < numArgs+1; i++ {
		vm.sp--
		vm.stack[vm.sp] = nil
	}

	result, err := callee.Call(args...)
	if err != nil {
		return err
	}

	vm.stack[vm.sp] = result
	vm.sp++
	vm.ip += 2

	return nil
}

func (vm *VM) execOpUnary() error {
	tok := token.Token(vm.curInsts[vm.ip+1])
	right := vm.stack[vm.sp-1]
	var value Object

	switch tok {
	case token.Not:
		vm.stack[vm.sp-1] = Bool(right.IsFalsy())
		vm.ip++
		return nil
	case token.Sub:
		switch o := right.(type) {
		case Int:
			value = -o
		case Float:
			value = -o
		case Char:
			value = Int(-o)
		case Uint:
			value = -o
		case Bool:
			if o {
				value = Int(-1)
			} else {
				value = Int(0)
			}
		default:
			goto invalidType
		}
	case token.Xor:
		switch o := right.(type) {
		case Int:
			value = ^o
		case Uint:
			value = ^o
		case Char:
			value = ^Int(o)
		case Bool:
			if o {
				value = ^Int(1)
			} else {
				value = ^Int(0)
			}
		default:
			goto invalidType
		}
	case token.Add:
		switch o := right.(type) {
		case Int, Uint, Float, Char:
			value = right
		case Bool:
			if o {
				value = Int(1)
			} else {
				value = Int(0)
			}
		default:
			goto invalidType
		}
	default:
		return ErrInvalidOperator.NewError(
			fmt.Sprintf("invalid for '%s': '%s'",
				tok.String(), right.TypeName()))
	}

	vm.stack[vm.sp-1] = value
	vm.ip++
	return nil

invalidType:
	return ErrType.NewError(
		fmt.Sprintf("invalid type for unary '%s': '%s'",
			tok.String(), right.TypeName()))
}

func (vm *VM) execOpSliceIndex() error {
	obj := vm.stack[vm.sp-3]
	left := vm.stack[vm.sp-2]
	right := vm.stack[vm.sp-1]
	vm.stack[vm.sp-3] = nil
	vm.stack[vm.sp-2] = nil
	vm.stack[vm.sp-1] = nil
	vm.sp -= 3
	var objLen int

	switch obj := obj.(type) {
	case Array:
		objLen = len(obj)
	case String:
		objLen = len(obj)
	case Bytes:
		objLen = len(obj)
	default:
		return ErrType.NewError(obj.TypeName(), "cannot be sliced")
	}
	var lowIdx int
	switch v := left.(type) {
	case undefined:
		lowIdx = 0
	case Int:
		lowIdx = int(v)
	case Uint:
		lowIdx = int(v)
	case Char:
		lowIdx = int(v)
	default:
		return ErrType.NewError("invalid first index type", left.TypeName())
	}
	var highIdx int
	switch v := right.(type) {
	case undefined:
		highIdx = objLen
	case Int:
		highIdx = int(v)
	case Uint:
		highIdx = int(v)
	case Char:
		highIdx = int(v)
	default:
		return ErrType.NewError("invalid second index type", right.TypeName())
	}

	if lowIdx > highIdx {
		return ErrInvalidIndex.NewError(fmt.Sprintf("[%d:%d]", lowIdx, highIdx))
	}

	if lowIdx < 0 || highIdx < 0 ||
		lowIdx > objLen-1 || highIdx > objLen {
		return ErrIndexOutOfBounds.NewError(
			fmt.Sprintf("[%d:%d]", lowIdx, highIdx))
	}

	switch obj := obj.(type) {
	case Array:
		vm.stack[vm.sp] = obj[lowIdx:highIdx]
	case String:
		vm.stack[vm.sp] = obj[lowIdx:highIdx]
	case Bytes:
		vm.stack[vm.sp] = obj[lowIdx:highIdx]
	}

	vm.sp++
	return nil
}

func (vm *VM) newError(err *Error) *RuntimeError {
	var fileset *parser.SourceFileSet
	if vm.bytecode != nil {
		fileset = vm.bytecode.FileSet
	}
	return &RuntimeError{Err: err, fileSet: fileset}
}

func (vm *VM) newErrorFromObject(object Object) *RuntimeError {
	switch v := object.(type) {
	case *RuntimeError:
		return v
	case *Error:
		return vm.newError(v)
	default:
		return vm.newError(&Error{Message: v.String()})
	}
}
func (vm *VM) newErrorFromError(err error) *RuntimeError {
	if v, ok := err.(Object); ok {
		return vm.newErrorFromObject(v)
	}
	return vm.newError(&Error{Message: err.Error(), Cause: err})
}

func (vm *VM) getSourcePos() parser.Pos {
	if vm.curFrame == nil || vm.curFrame.fn == nil {
		return parser.NoPos
	}
	return vm.curFrame.fn.SourcePos(vm.ip)
}

type errHandler struct {
	sp       int
	catch    int
	finally  int
	returnTo int
}

type errHandlers struct {
	handlers []errHandler
	err      *RuntimeError
}

func (t *errHandlers) hasError() bool {
	return t != nil && t.err != nil
}

func (t *errHandlers) pop() bool {
	if t == nil || len(t.handlers) == 0 {
		return false
	}

	t.handlers = t.handlers[:len(t.handlers)-1]
	return true
}

func (t *errHandlers) last() *errHandler {
	if t == nil || len(t.handlers) == 0 {
		return nil
	}
	return &t.handlers[len(t.handlers)-1]
}

func (t *errHandlers) hasHandler() bool {
	return t != nil && len(t.handlers) > 0
}

func (t *errHandlers) findFinally(upto int) int {
	if t == nil {
		return 0
	}

start:
	index := len(t.handlers) - 1
	if index < upto || index < 0 {
		return 0
	}

	p := t.handlers[index].finally
	if p == 0 {
		t.pop()
		goto start
	}
	return p
}

func (t *errHandlers) hasReturnTo() int {
	if t.hasHandler() {
		return t.handlers[len(t.handlers)-1].returnTo
	}
	return 0
}

type frame struct {
	fn          *CompiledFunction
	freeVars    []*ObjectPtr
	ip          int
	basePointer int
	errHandlers *errHandlers
}

func getFrameSourcePos(frame *frame) parser.Pos {
	if frame == nil || frame.fn == nil {
		return parser.NoPos
	}
	return frame.fn.SourcePos(frame.ip + 1)
}
