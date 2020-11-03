// A modified version of Tengo Compiler.

// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Copyright (c) 2019 Daniel Kang.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE.tengo file.

package ugo

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/ozanh/ugo/parser"
	"github.com/ozanh/ugo/token"
)

// CompilerOptions represents customizable options for Compile().
type CompilerOptions struct {
	ModuleMap         *ModuleMap
	ModulePath        string
	ModuleIndexes     *ModuleIndexes
	Constants         []Object
	SymbolTable       *SymbolTable
	Trace             io.Writer
	TraceParser       bool
	TraceCompiler     bool
	TraceOptimizer    bool
	OptimizerMaxCycle int
	OptimizeConst     bool
	OptimizeExpr      bool
	constsCache       map[Object]int
}

var (
	// DefaultCompilerOptions holds default Compiler options.
	DefaultCompilerOptions = CompilerOptions{
		OptimizerMaxCycle: 100,
		OptimizeConst:     true,
		OptimizeExpr:      true,
	}
	// TraceCompilerOptions holds Compiler options to print trace output
	// to stdout for Parser, Optimizer, Compiler.
	TraceCompilerOptions = CompilerOptions{
		Trace:             os.Stdout,
		TraceParser:       true,
		TraceCompiler:     true,
		TraceOptimizer:    true,
		OptimizerMaxCycle: 1<<8 - 1,
		OptimizeConst:     true,
		OptimizeExpr:      true,
	}
)

// loopStmts represents a loopStmts construct that the compiler uses to track the current loopStmts.
type loopStmts struct {
	Continues         []int
	Breaks            []int
	lastTryCatchIndex int
}

// CompilerError represents a compiler error.
type CompilerError struct {
	FileSet *parser.SourceFileSet
	Node    parser.Node
	Err     error
}

func (e *CompilerError) Error() string {
	filePos := e.FileSet.Position(e.Node.Pos())
	return fmt.Sprintf("Compile Error: %s\n\tat %s", e.Err.Error(), filePos)
}

func (e *CompilerError) Unwrap() error {
	return e.Err
}

// ModuleIndex represents indexes of a single module.
type ModuleIndex struct {
	ConstantIndex int
	ModuleIndex   int
}

// ModuleIndexes represents modules indexes and total count that are defined while compiling.
type ModuleIndexes struct {
	Count   int
	Indexes map[string]ModuleIndex
}

// NewModuleIndexes returns a new ModuleIndexes object.
func NewModuleIndexes() *ModuleIndexes {
	return &ModuleIndexes{
		Indexes: make(map[string]ModuleIndex),
	}
}

// Compiler compiles the AST into a bytecode.
type Compiler struct {
	parent        *Compiler
	file          *parser.SourceFile
	constants     []Object
	constsCache   map[Object]int
	symbolTable   *SymbolTable
	instructions  []byte
	sourceMap     map[int]int
	moduleMap     *ModuleMap
	moduleIndexes *ModuleIndexes
	modulePath    string
	variadic      bool
	loops         []*loopStmts
	loopIndex     int
	tryCatchIndex int
	opts          CompilerOptions
	trace         io.Writer
	indent        int
}

// NewCompiler creates a new Compiler object.
func NewCompiler(file *parser.SourceFile, opts CompilerOptions) *Compiler {
	st := opts.SymbolTable
	if st == nil {
		st = NewSymbolTable()
	}
	if opts.constsCache == nil {
		opts.constsCache = make(map[Object]int)
		for i := range opts.Constants {
			switch opts.Constants[i].(type) {
			case Int, Uint, String, Bool, Float, Char, undefined,
				*CompiledFunction:
				opts.constsCache[opts.Constants[i]] = i
			}
		}
	}
	if opts.ModuleMap == nil {
		opts.ModuleMap = NewModuleMap()
	}
	if opts.ModuleIndexes == nil {
		opts.ModuleIndexes = NewModuleIndexes()
	}
	var trace io.Writer
	if opts.TraceCompiler {
		trace = opts.Trace
	}
	return &Compiler{
		file:          file,
		constants:     opts.Constants,
		constsCache:   opts.constsCache,
		symbolTable:   st,
		sourceMap:     make(map[int]int),
		moduleMap:     opts.ModuleMap,
		moduleIndexes: opts.ModuleIndexes,
		modulePath:    opts.ModulePath,
		loopIndex:     -1,
		tryCatchIndex: -1,
		opts:          opts,
		trace:         trace,
	}
}

// Compile compiles given script to Bytecode.
func Compile(script []byte, opts CompilerOptions) (*Bytecode, error) {
	fileSet := parser.NewFileSet()
	moduleName := opts.ModulePath
	if moduleName == "" {
		moduleName = "(main)"
	}
	srcFile := fileSet.AddFile(moduleName, -1, len(script))
	var trace io.Writer
	if opts.TraceParser {
		trace = opts.Trace
	}
	p := parser.NewParser(srcFile, script, trace)
	pf, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	compiler := NewCompiler(srcFile, opts)
	if opts.OptimizeConst || opts.OptimizeExpr {
		optim, err := compiler.optimize(pf)
		if err != nil {
			return nil, err
		}
		if optim != nil {
			opts.OptimizerMaxCycle -= optim.Total()
			if opts.TraceCompiler && !opts.TraceOptimizer {
				_, _ = fmt.Fprintf(opts.Trace,
					"<Optimization Took: %s>\n", optim.Duration())
			}
		}

	}
	if err := compiler.Compile(pf); err != nil {
		return nil, err
	}
	bc := compiler.Bytecode()
	if bc.Main.NumLocals > 256 {
		return nil, ErrSymbolLimit
	}
	return bc, nil
}

// optimize runs the Optimizer and returns Optimizer object and error from Optimizer.
// Note:If optimizer cannot run for some reason, all returned values will be nil.
func (c *Compiler) optimize(file *parser.File) (*SimpleOptimizer, error) {
	if c.opts.OptimizerMaxCycle < 1 {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var trace io.Writer
	if c.opts.TraceOptimizer {
		trace = c.opts.Trace
	}
	o := NewOptimizer(
		ctx,
		file,
		c.opts.OptimizeConst,
		c.opts.OptimizeExpr,
		c.opts.OptimizerMaxCycle,
		trace,
	)
	dis := c.symbolTable.DisabledBuiltins()
	if err := o.DisableBuiltins(dis).Optimize(); err != nil {
		return o, err
	}
	c.opts.OptimizerMaxCycle -= o.Total()
	return o, nil
}

// Bytecode returns compiled Bytecode ready to run in VM.
func (c *Compiler) Bytecode() *Bytecode {
	var lastOp Opcode
	var operands = make([]int, 0, 4)
	var jumpPos = make(map[int]struct{})
	var offset int
	var i int
	for i < len(c.instructions) {
		lastOp = c.instructions[i]
		numOperands := OpcodeOperands[lastOp]
		operands, offset = ReadOperands(
			numOperands,
			c.instructions[i+1:],
			operands,
		)
		if lastOp == OpJump || lastOp == OpJumpFalsy ||
			lastOp == OpAndJump || lastOp == OpOrJump {
			jumpPos[operands[0]] = struct{}{}
		}
		delete(jumpPos, i)
		i += offset + 1
	}
	if lastOp != OpReturn || len(jumpPos) > 0 {
		c.emit(nil, OpReturn, 0)
	}
	return &Bytecode{
		FileSet:   c.file.Set(),
		Constants: c.constants,
		Main: &CompiledFunction{
			NumParams:    c.symbolTable.NumParams(),
			NumLocals:    c.symbolTable.MaxSymbols(),
			Variadic:     c.variadic,
			Instructions: c.instructions,
			SourceMap:    c.sourceMap,
		},
		NumModules: c.moduleIndexes.Count,
	}
}

// Compile compiles parser.Node and builds Bytecode.
func (c *Compiler) Compile(node parser.Node) error {
	if c.trace != nil {
		if node != nil {
			defer untracec(tracec(c, fmt.Sprintf("%s (%s)",
				node.String(), reflect.TypeOf(node).Elem().Name())))
		} else {
			defer untracec(tracec(c, "<nil>"))
		}
	}
	switch node := node.(type) {
	case *parser.File:
		for _, stmt := range node.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}
	case *parser.ExprStmt:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, OpPop)
	case *parser.IncDecStmt:
		op := token.AddAssign
		if node.Token == token.Dec {
			op = token.SubAssign
		}
		return c.compileAssign(node, []parser.Expr{node.Expr},
			[]parser.Expr{&parser.IntLit{Value: 1}}, op)
	case *parser.ParenExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
	case *parser.BinaryExpr:
		if node.Token == token.LAnd || node.Token == token.LOr {
			return c.compileLogical(node)
		}
		if err := c.Compile(node.LHS); err != nil {
			return err
		}
		if err := c.Compile(node.RHS); err != nil {
			return err
		}
		switch node.Token {
		case token.Equal:
			c.emit(node, OpEqual)
		case token.NotEqual:
			c.emit(node, OpNotEqual)
		case token.Add:
			c.emit(node, OpBinaryOp, int(token.Add))
		case token.Sub:
			c.emit(node, OpBinaryOp, int(token.Sub))
		case token.Mul:
			c.emit(node, OpBinaryOp, int(token.Mul))
		case token.Quo:
			c.emit(node, OpBinaryOp, int(token.Quo))
		case token.Rem:
			c.emit(node, OpBinaryOp, int(token.Rem))
		case token.Less:
			c.emit(node, OpBinaryOp, int(token.Less))
		case token.LessEq:
			c.emit(node, OpBinaryOp, int(token.LessEq))
		case token.Greater:
			c.emit(node, OpBinaryOp, int(token.Greater))
		case token.GreaterEq:
			c.emit(node, OpBinaryOp, int(token.GreaterEq))
		case token.And:
			c.emit(node, OpBinaryOp, int(token.And))
		case token.Or:
			c.emit(node, OpBinaryOp, int(token.Or))
		case token.Xor:
			c.emit(node, OpBinaryOp, int(token.Xor))
		case token.AndNot:
			c.emit(node, OpBinaryOp, int(token.AndNot))
		case token.Shl:
			c.emit(node, OpBinaryOp, int(token.Shl))
		case token.Shr:
			c.emit(node, OpBinaryOp, int(token.Shr))
		default:
			return c.errorf(node, "invalid binary operator: %s",
				node.Token.String())
		}
	case *parser.IntLit:
		c.emit(node, OpConstant, c.addConstant(Int(node.Value)))
	case *parser.UintLit:
		c.emit(node, OpConstant, c.addConstant(Uint(node.Value)))
	case *parser.FloatLit:
		c.emit(node, OpConstant, c.addConstant(Float(node.Value)))
	case *parser.BoolLit:
		if node.Value {
			c.emit(node, OpConstant, c.addConstant(True))
		} else {
			c.emit(node, OpConstant, c.addConstant(False))
		}
	case *parser.StringLit:
		c.emit(node, OpConstant, c.addConstant(String(node.Value)))
	case *parser.CharLit:
		c.emit(node, OpConstant, c.addConstant(Char(node.Value)))
	case *parser.UndefinedLit:
		c.emit(node, OpNull)
	case *parser.UnaryExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		switch node.Token {
		case token.Not:
			c.emit(node, OpUnary, int(token.Not))
		case token.Sub:
			c.emit(node, OpUnary, int(token.Sub))
		case token.Xor:
			c.emit(node, OpUnary, int(token.Xor))
		case token.Add:
			c.emit(node, OpUnary, int(token.Add))
		default:
			return c.errorf(node,
				"invalid unary operator: %s", node.Token.String())
		}
	case *parser.IfStmt:
		// open new symbol table for the statement
		nextIndex := c.symbolTable.NextIndex()
		c.symbolTable = c.symbolTable.Fork(true)
		defer func() {
			parent := c.symbolTable.Parent(false)
			// set undefined to variables having no reference
			maxSymbols := c.symbolTable.MaxSymbols()
			for i := nextIndex; i < maxSymbols; i++ {
				if !parent.IsIndexSkipped(i) {
					c.emit(node, OpNull)
					c.emit(node, OpSetLocal, i)
				}
			}
			c.symbolTable = parent
		}()
		if node.Init != nil {
			if err := c.Compile(node.Init); err != nil {
				return err
			}
		}
		jumpPos1 := -1
		var skipElse bool
		if v, ok := node.Cond.(*parser.BoolLit); !ok {
			if err := c.Compile(node.Cond); err != nil {
				return err
			}
			// first jump placeholder
			jumpPos1 = c.emit(node, OpJumpFalsy, 0)
			if err := c.Compile(node.Body); err != nil {
				return err
			}
		} else if v.Value {
			if err := c.Compile(node.Body); err != nil {
				return err
			}
			skipElse = true
		} else {
			jumpPos1 = c.emit(node, OpJump, 0)
		}
		if !skipElse && node.Else != nil {
			// second jump placeholder
			jumpPos2 := c.emit(node, OpJump, 0)
			if jumpPos1 > -1 {
				// update first jump offset
				curPos := len(c.instructions)
				c.changeOperand(jumpPos1, curPos)
			}
			if err := c.Compile(node.Else); err != nil {
				return err
			}
			// update second jump offset
			curPos := len(c.instructions)
			c.changeOperand(jumpPos2, curPos)
		} else {
			if jumpPos1 > -1 {
				// update first jump offset
				curPos := len(c.instructions)
				c.changeOperand(jumpPos1, curPos)
			}
		}
	case *parser.TryStmt:
		/*
			// create a single symbol table for try-catch-finally
			// any `return` statement in finally block ignores already thrown error.
			try {
				// emit: OpSetupTry (CatchPos, FinallyPos)

				// emit: OpJump (FinallyPos) // jump to finally block to skip catch block.
			} catch err {
				// emit: OpSetupCatch
				//
				// catch block is optional.
				// if err is elided  in `catch {}`, OpPop removes the error from stack.
				// catch pops the error from error handler, re-throw requires explicit
				// throw expression `throw err`.
			} finally {
				// emit: OpSetupFinally
				//
				// finally block is optional if catch block is defined but
				// instructions are always generated for finally block even if not explicitly defined
				// to cleanup symbols and re-throw error if not handled in catch block.
				//
				// emit: OpThrow 0 // this is implicit re-throw operation without putting stack trace
			}
		*/
		// open new symbol table for the statement
		nextIndex := c.symbolTable.NextIndex()
		c.symbolTable = c.symbolTable.Fork(true)
		c.tryCatchIndex++
		defer func() {
			parent := c.symbolTable.Parent(false)
			// set undefined to variables having no reference
			maxSymbols := c.symbolTable.MaxSymbols()
			for i := nextIndex; i < maxSymbols; i++ {
				if !parent.IsIndexSkipped(i) {
					c.emit(node, OpNull)
					c.emit(node, OpSetLocal, i)
				}
			}
			c.symbolTable = parent
			c.emit(node, OpThrow, 0) // implicit re-throw
		}()
		optry := c.emit(node, OpSetupTry, 0, 0)
		var catchPos, finallyPos int
		if len(node.Body.Stmts) != 0 {
			// in order not to fork symbol table in Body, compile stmts here instead of in *BlockStmt
			for _, stmt := range node.Body.Stmts {
				if err := c.Compile(stmt); err != nil {
					return err
				}
			}
		}
		var opjump int
		if node.Catch != nil {
			opjump = c.emit(node, OpJump, 0)
			catchPos = len(c.instructions)
			if err := c.Compile(node.Catch); err != nil {
				return err
			}
		}
		c.tryCatchIndex--
		// always emit OpSetupFinally to cleanup
		finallyPos = c.emit(node, OpSetupFinally)
		if node.Finally != nil {
			if err := c.Compile(node.Finally); err != nil {
				return err
			}
		}
		c.changeOperand(optry, catchPos, finallyPos)
		if node.Catch != nil {
			// no need jumping if catch is not defined
			c.changeOperand(opjump, finallyPos)
		}
	case *parser.CatchStmt:
		c.emit(node, OpSetupCatch)
		var symbol *Symbol
		if node.Ident != nil {
			symbol, _ = c.symbolTable.DefineLocal(node.Ident.Name)
			c.emit(node, OpSetLocal, symbol.Index)
		} else {
			c.emit(node, OpPop)
		}
		if len(node.Body.Stmts) == 0 {
			return nil
		}
		// in order not to fork symbol table in Body, compile stmts here instead of in *BlockStmt
		for _, stmt := range node.Body.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}
	case *parser.FinallyStmt:
		if len(node.Body.Stmts) == 0 {
			return nil
		}
		// in order not to fork symbol table in Body, compile stmts here instead of in *BlockStmt
		for _, stmt := range node.Body.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}
	case *parser.ThrowStmt:
		if node.Expr != nil {
			if err := c.Compile(node.Expr); err != nil {
				return err
			}
		}
		c.emit(node, OpThrow, 1)
	case *parser.ForStmt:
		return c.compileForStmt(node)
	case *parser.ForInStmt:
		return c.compileForInStmt(node)
	case *parser.BranchStmt:
		if node.Token == token.Break {
			curLoop := c.currentLoop()
			if curLoop == nil {
				return c.errorf(node, "break not allowed outside loop")
			}
			var pos int
			if curLoop.lastTryCatchIndex == c.tryCatchIndex {
				pos = c.emit(node, OpJump, 0)
			} else {
				c.emit(node, OpFinalizer, curLoop.lastTryCatchIndex+1)
				pos = c.emit(node, OpJump, 0)
			}
			curLoop.Breaks = append(curLoop.Breaks, pos)
		} else if node.Token == token.Continue {
			curLoop := c.currentLoop()
			if curLoop == nil {
				return c.errorf(node, "continue not allowed outside loop")
			}
			var pos int
			if curLoop.lastTryCatchIndex == c.tryCatchIndex {
				pos = c.emit(node, OpJump, 0)
			} else {
				c.emit(node, OpFinalizer, curLoop.lastTryCatchIndex+1)
				pos = c.emit(node, OpJump, 0)
			}
			curLoop.Continues = append(curLoop.Continues, pos)
		} else {
			panic(fmt.Errorf("invalid branch statement: %s",
				node.Token.String()))
		}
	case *parser.BlockStmt:
		if len(node.Stmts) == 0 {
			return nil
		}
		c.symbolTable = c.symbolTable.Fork(true)
		defer func() {
			c.symbolTable = c.symbolTable.Parent(false)
		}()
		for _, stmt := range node.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}
	case *parser.AssignStmt:
		err := c.compileAssign(node, node.LHS, node.RHS, node.Token)
		if err != nil {
			return err
		}
	case *parser.DeclStmt:
		decl, ok := node.Decl.(*parser.GenDecl)
		if !ok {
			return c.errorf(node, "only GenDecl is supported in DeclStmt")
		}
		switch decl.Tok {
		case token.Param:
			if len(decl.Specs) == 0 {
				return c.errorf(node, "empty param declaration not allowed")
			}
			if c.symbolTable.parent != nil {
				return c.errorf(node, "param not allowed in this scope")
			}
			names := make([]string, 0)
			var variadic bool
			for _, sp := range decl.Specs {
				spec, ok := sp.(*parser.ParamSpec)
				if !ok {
					return c.errorf(node,
						"ParamSpec is expected but got %T at GenDecl.Specs",
						decl.Specs[0],
					)
				}
				if spec.Ident == nil {
					return c.errorf(node, "param ident not defined")
				}
				names = append(names, spec.Ident.Name)
				if spec.Variadic {
					if variadic {
						return c.errorf(node,
							"multiple variadic param declaration")
					}
					variadic = true
				}
			}
			c.variadic = variadic
			err := c.symbolTable.SetParams(names...)
			if err != nil {
				return c.error(node, err)
			}
			return nil
		case token.Global:
			if len(decl.Specs) == 0 {
				return c.errorf(node, "empty global declaration not allowed")
			}
			if c.symbolTable.parent != nil {
				return c.errorf(node, "global not allowed in this scope")
			}
			for _, sp := range decl.Specs {
				spec, ok := sp.(*parser.ParamSpec)
				if !ok {
					return c.errorf(node,
						"ParamSpec is expected but got %T at GenDecl.Specs", sp)
				}
				if spec.Ident != nil {
					if c.symbolTable.IsGlobal(spec.Ident.Name) {
						return c.errorf(node,
							"duplicate global variable declaration or shadowed variable")
					}
					symbol, err := c.symbolTable.DefineGlobal(spec.Ident.Name)
					if err != nil {
						return c.error(node, err)
					}
					idx := c.addConstant(String(spec.Ident.Name))
					symbol.Index = idx
				} else {
					return c.errorf(node, "global ident not defined")
				}
			}
			return nil
		case token.Var:
			if len(decl.Specs) == 0 {
				return c.errorf(node, "empty var declaration not allowed")
			}
			for _, sp := range decl.Specs {
				spec, ok := sp.(*parser.ValueSpec)
				if !ok {
					return c.errorf(node,
						"ValueSpec is expected but got %T at GenDecl.Specs", sp)
				}
				for i, ident := range spec.Idents {
					leftExpr := []parser.Expr{ident}
					var v parser.Expr
					if i < len(spec.Values) {
						v = spec.Values[i]
					} else {
						v = &parser.UndefinedLit{TokenPos: ident.Pos()}
					}
					rightExpr := []parser.Expr{v}
					err := c.compileAssign(node, leftExpr, rightExpr, token.Define)
					if err != nil {
						return err
					}
				}
			}
			return nil
		}

	case *parser.Ident:
		symbol, ok := c.symbolTable.Resolve(node.Name)
		if !ok {
			return c.errorf(node, "unresolved reference %q", node.Name)
		}
		switch symbol.Scope {
		case ScopeGlobal:
			c.emit(node, OpGetGlobal, symbol.Index)
		case ScopeLocal:
			c.emit(node, OpGetLocal, symbol.Index)
		case ScopeBuiltin:
			c.emit(node, OpGetBuiltin, symbol.Index)
		case ScopeFree:
			c.emit(node, OpGetFree, symbol.Index)
		}
	case *parser.ArrayLit:
		for _, elem := range node.Elements {
			if err := c.Compile(elem); err != nil {
				return err
			}
		}
		c.emit(node, OpArray, len(node.Elements))
	case *parser.MapLit:
		for _, elt := range node.Elements {
			// key
			c.emit(node, OpConstant, c.addConstant(String(elt.Key)))
			// value
			if err := c.Compile(elt.Value); err != nil {
				return err
			}
		}
		c.emit(node, OpMap, len(node.Elements)*2)
	case *parser.SelectorExpr: // selector on RHS side
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		if err := c.Compile(node.Sel); err != nil {
			return err
		}
		c.emit(node, OpGetIndex, 1)
	case *parser.IndexExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		if err := c.Compile(node.Index); err != nil {
			return err
		}
		c.emit(node, OpGetIndex, 1)
	case *parser.SliceExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		if node.Low != nil {
			if err := c.Compile(node.Low); err != nil {
				return err
			}
		} else {
			c.emit(node, OpNull)
		}
		if node.High != nil {
			if err := c.Compile(node.High); err != nil {
				return err
			}
		} else {
			c.emit(node, OpNull)
		}
		c.emit(node, OpSliceIndex)
	case *parser.FuncLit:
		params := make([]string, len(node.Type.Params.List))
		for i, ident := range node.Type.Params.List {
			params[i] = ident.Name
		}
		symbolTable := c.symbolTable.Fork(false)
		if err := symbolTable.SetParams(params...); err != nil {
			return c.error(node, err)
		}
		fork := c.fork(c.file, c.modulePath, symbolTable)
		fork.variadic = node.Type.Params.VarArgs
		if err := fork.Compile(node.Body); err != nil {
			return err
		}
		freeSymbols := fork.symbolTable.FreeSymbols()
		for _, s := range freeSymbols {
			switch s.Scope {
			case ScopeLocal:
				c.emit(node, OpGetLocalPtr, s.Index)
				if c.symbolTable.InBlock() {
					c.symbolTable.SkipIndex(s.Index)
				}
			case ScopeFree:
				c.emit(node, OpGetFreePtr, s.Index)
				if c.symbolTable.InBlock() {
					ls := s.Original
					for ls != nil {
						if ls.Scope == ScopeLocal {
							c.symbolTable.SkipIndex(s.Index)
							break
						}
						ls = ls.Original
					}
				}
			}
		}

		bc := fork.Bytecode()
		if bc.Main.NumLocals > 256 {
			return c.error(node, ErrSymbolLimit)
		}
		c.constants = bc.Constants
		index := c.addConstant(bc.Main)
		if len(freeSymbols) > 0 {
			c.emit(node, OpClosure, index, len(freeSymbols))
		} else {
			c.emit(node, OpConstant, index)
		}
	case *parser.ReturnStmt:
		if node.Result == nil {
			if c.tryCatchIndex > -1 {
				c.emit(node, OpFinalizer, 0)
			}
			c.emit(node, OpNull)
			c.emit(node, OpReturn, 0)
		} else {
			if err := c.Compile(node.Result); err != nil {
				return err
			}
			if c.tryCatchIndex > -1 {
				c.emit(node, OpFinalizer, 0)
			}
			c.emit(node, OpReturn, 1)
		}
	case *parser.CallExpr:
		if err := c.Compile(node.Func); err != nil {
			return err
		}
		for _, arg := range node.Args {
			if err := c.Compile(arg); err != nil {
				return err
			}
		}
		var expand int
		if node.Ellipsis.IsValid() {
			expand = 1
		}
		c.emit(node, OpCall, len(node.Args), expand)
	case *parser.ImportExpr:
		if node.ModuleName == "" {
			return c.errorf(node, "empty module name")
		}
		if mod := c.moduleMap.Get(node.ModuleName); mod != nil {
			v, err := mod.Import(node.ModuleName)
			if err != nil {
				return err
			}

			switch v := v.(type) {
			case []byte:
				moduleIndexes, exists := c.getModule(node.ModuleName)
				if !exists {
					moduleIndexes, err = c.compileModule(
						node, node.ModuleName, v)
					if err != nil {
						return err
					}
				}
				var numParams int
				mod := c.constants[moduleIndexes.ConstantIndex]
				if cf, ok := mod.(*CompiledFunction); ok {
					numParams = cf.NumParams
					if cf.Variadic {
						numParams--
					}
				}
				// load module
				// if module is already stored, load from VM.modulesCache otherwise call compiled function
				// and store copy of result to VM.modulesCache.
				c.emit(node, OpLoadModule,
					moduleIndexes.ConstantIndex, moduleIndexes.ModuleIndex)
				jumpPos := c.emit(node, OpJumpFalsy, 0)
				// modules should not accept parameters, to suppress the wrong number of arguments error
				// set all params to undefined
				for i := 0; i < numParams; i++ {
					c.emit(node, OpNull)
				}
				c.emit(node, OpCall, numParams, 0)
				c.emit(node, OpStoreModule, moduleIndexes.ModuleIndex)
				c.changeOperand(jumpPos, len(c.instructions))
			case Object:
				moduleIndexes, exists := c.getModule(node.ModuleName)
				if !exists {
					moduleIndexes = c.addModule(node.ModuleName, c.addConstant(v))
				}
				// load module
				// if module is already stored, load from VM.modulesCache otherwise copy object
				// and store it to VM.modulesCache.
				c.emit(node, OpLoadModule,
					moduleIndexes.ConstantIndex, moduleIndexes.ModuleIndex)
				jumpPos := c.emit(node, OpJumpFalsy, 0)
				c.emit(node, OpStoreModule, moduleIndexes.ModuleIndex)
				c.changeOperand(jumpPos, len(c.instructions))
			default:
				panic(fmt.Errorf("invalid import value type: %T", v))
			}
		} else {
			return c.errorf(node, "module '%s' not found", node.ModuleName)
		}
	case *parser.CondExpr:
		if v, ok := node.Cond.(*parser.BoolLit); !ok {
			if err := c.Compile(node.Cond); err != nil {
				return err
			}
			// first jump placeholder
			jumpPos1 := c.emit(node, OpJumpFalsy, 0)
			if err := c.Compile(node.True); err != nil {
				return err
			}

			// second jump placeholder
			jumpPos2 := c.emit(node, OpJump, 0)

			// update first jump offset
			curPos := len(c.instructions)
			c.changeOperand(jumpPos1, curPos)
			if err := c.Compile(node.False); err != nil {
				return err
			}
			// update second jump offset
			curPos = len(c.instructions)
			c.changeOperand(jumpPos2, curPos)
		} else if v.Value {
			if err := c.Compile(node.True); err != nil {
				return err
			}
		} else {
			if err := c.Compile(node.False); err != nil {
				return err
			}
		}
	case *parser.EmptyStmt:
	case nil:
	default:
		return c.errorf(node, "%[1]T \"%[1]v\" not implemented", node)
	}

	return nil
}

func (c *Compiler) changeOperand(opPos int, operand ...int) {
	op := c.instructions[opPos]
	inst, err := MakeInstruction(op, operand...)
	if err != nil {
		panic(err)
	}
	c.replaceInstruction(opPos, inst)
}

func (c *Compiler) replaceInstruction(pos int, inst []byte) {
	copy(c.instructions[pos:], inst)
	if c.trace != nil {
		c.printTrace(fmt.Sprintf("REPLC %s",
			FormatInstructions(
				c.instructions[pos:], pos)[0]))
	}
}

func (c *Compiler) compileAssign(node parser.Node, lhs, rhs []parser.Expr,
	op token.Token) error {

	numLHS, numRHS := len(lhs), len(rhs)
	if numLHS > 1 || numRHS > 1 {
		return c.errorf(node, "tuple assignment not allowed")
	}

	// resolve and compile left-hand side
	ident, selectors := resolveAssignLHS(lhs[0])
	numSel := len(selectors)

	if op == token.Define && numSel > 0 {
		// using selector on new variable does not make sense
		return c.errorf(node, "operator ':=' not allowed with selector")
	} else if op == token.Define && numSel == 0 && len(rhs) == 1 {
		// exception for variable := undefined
		// all local variables are inited as undefined at VM, ignore if rhs[0] == undefined
		if _, ok := rhs[0].(*parser.UndefinedLit); ok {
			symbol, ok := c.symbolTable.DefineLocal(ident)
			if ok {
				return c.errorf(node, "%q ", ident)
			}
			symbol.Assigned = true
			return nil
		}
	}

	// +=, -=, *=, /=
	if op != token.Assign && op != token.Define {
		if err := c.Compile(lhs[0]); err != nil {
			return err
		}
	}

	// compile RHSs
	for _, expr := range rhs {
		if err := c.Compile(expr); err != nil {
			return err
		}
	}

	switch op {
	case token.AddAssign:
		c.emit(node, OpBinaryOp, int(token.Add))
	case token.SubAssign:
		c.emit(node, OpBinaryOp, int(token.Sub))
	case token.MulAssign:
		c.emit(node, OpBinaryOp, int(token.Mul))
	case token.QuoAssign:
		c.emit(node, OpBinaryOp, int(token.Quo))
	case token.RemAssign:
		c.emit(node, OpBinaryOp, int(token.Rem))
	case token.AndAssign:
		c.emit(node, OpBinaryOp, int(token.And))
	case token.OrAssign:
		c.emit(node, OpBinaryOp, int(token.Or))
	case token.AndNotAssign:
		c.emit(node, OpBinaryOp, int(token.AndNot))
	case token.XorAssign:
		c.emit(node, OpBinaryOp, int(token.Xor))
	case token.ShlAssign:
		c.emit(node, OpBinaryOp, int(token.Shl))
	case token.ShrAssign:
		c.emit(node, OpBinaryOp, int(token.Shr))
	}

	if op == token.Define {
		symbol, ok := c.symbolTable.DefineLocal(ident)
		if ok {
			return c.errorf(node, "%q redeclared in this block", ident)
		}
		c.emit(node, OpSetLocal, symbol.Index)
		symbol.Assigned = true
		return nil
	}
	symbol, ok := c.symbolTable.Resolve(ident)
	if !ok {
		return c.errorf(node, "unresolved reference %q", ident)
	}

	if numSel == 0 {
		switch symbol.Scope {
		case ScopeLocal:
			c.emit(node, OpSetLocal, symbol.Index)
			symbol.Assigned = true
		case ScopeFree:
			c.emit(node, OpSetFree, symbol.Index)
			symbol.Assigned = true
			s := symbol
			for s != nil {
				if s.Original != nil && s.Original.Scope == ScopeLocal {
					s.Original.Assigned = true
				}
				s = s.Original
			}
		case ScopeGlobal:
			c.emit(node, OpSetGlobal, symbol.Index)
			symbol.Assigned = true
		default:
			return c.errorf(node, "unresolved reference %q", ident)
		}
		return nil
	}
	switch symbol.Scope {
	case ScopeLocal:
		c.emit(node, OpGetLocal, symbol.Index)
	case ScopeFree:
		c.emit(node, OpGetFree, symbol.Index)
	case ScopeGlobal:
		c.emit(node, OpGetGlobal, symbol.Index)
	default:
		return c.errorf(node, "unresolved reference %q", ident)
	}
	if numSel > 1 {
		for i := 0; i < numSel-1; i++ {
			if err := c.Compile(selectors[i]); err != nil {
				return err
			}
		}
		c.emit(node, OpGetIndex, numSel-1)
	}
	if err := c.Compile(selectors[numSel-1]); err != nil {
		return err
	}
	c.emit(node, OpSetIndex)
	return nil
}

func resolveAssignLHS(expr parser.Expr) (name string, selectors []parser.Expr) {
	switch term := expr.(type) {
	case *parser.SelectorExpr:
		name, selectors = resolveAssignLHS(term.Expr)
		selectors = append(selectors, term.Sel)
		return
	case *parser.IndexExpr:
		name, selectors = resolveAssignLHS(term.Expr)
		selectors = append(selectors, term.Index)
	case *parser.Ident:
		name = term.Name
	}
	return
}

func (c *Compiler) addConstant(obj Object) (index int) {
	defer func() {
		if c.trace != nil {
			c.printTrace(fmt.Sprintf("CONST %04d %s", index, obj))
		}
	}()
	switch obj.(type) {
	case Int, Uint, String, Bool, Float, Char, undefined:
		i, ok := c.constsCache[obj]
		if ok {
			index = i
			return
		}
	case *CompiledFunction:
		for i, v := range c.constants {
			if f, ok := v.(*CompiledFunction); ok {
				if reflect.DeepEqual(f, obj) {
					index = i
					return
				}
			}
		}
	default:
		// unhashable types cannot be stored in constsCache, append them to constants slice
		// and return index
		c.constants = append(c.constants, obj)
		index = len(c.constants) - 1
		return
	}
	c.constants = append(c.constants, obj)
	index = len(c.constants) - 1
	c.constsCache[obj] = index
	return
}

func (c *Compiler) emit(node parser.Node, opcode Opcode, operands ...int) int {
	filePos := parser.NoPos
	if node != nil {
		filePos = node.Pos()
	}

	inst, err := MakeInstruction(opcode, operands...)
	if err != nil {
		panic(err)
	}
	pos := c.addInstruction(inst)
	c.sourceMap[pos] = int(filePos)

	if c.trace != nil {
		c.printTrace(fmt.Sprintf("EMIT  %s",
			FormatInstructions(
				c.instructions[pos:], pos)[0]))
	}
	return pos
}

func (c *Compiler) addInstruction(b []byte) int {
	posNewIns := len(c.instructions)
	c.instructions = append(c.instructions, b...)
	return posNewIns
}

func (c *Compiler) compileLogical(node *parser.BinaryExpr) error {
	// left side term
	if err := c.Compile(node.LHS); err != nil {
		return err
	}

	// jump position
	var jumpPos int
	if node.Token == token.LAnd {
		jumpPos = c.emit(node, OpAndJump, 0)
	} else {
		jumpPos = c.emit(node, OpOrJump, 0)
	}

	// right side term
	if err := c.Compile(node.RHS); err != nil {
		return err
	}
	c.changeOperand(jumpPos, len(c.instructions))
	return nil
}

func (c *Compiler) compileForStmt(stmt *parser.ForStmt) error {
	nextIndex := c.symbolTable.NextIndex()
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		parent := c.symbolTable.Parent(false)
		// set undefined to variables having no reference
		maxSymbols := c.symbolTable.MaxSymbols()
		for i := nextIndex; i < maxSymbols; i++ {
			if !parent.IsIndexSkipped(i) {
				c.emit(stmt, OpNull)
				c.emit(stmt, OpSetLocal, i)
			}
		}
		c.symbolTable = parent
	}()

	// init statement
	if stmt.Init != nil {
		if err := c.Compile(stmt.Init); err != nil {
			return err
		}
	}

	// pre-condition position
	preCondPos := len(c.instructions)

	// condition expression
	postCondPos := -1
	if stmt.Cond != nil {
		if err := c.Compile(stmt.Cond); err != nil {
			return err
		}
		// condition jump position
		postCondPos = c.emit(stmt, OpJumpFalsy, 0)
	}

	// enter loop
	loop := c.enterLoop()

	// body statement
	if err := c.Compile(stmt.Body); err != nil {
		c.leaveLoop()
		return err
	}

	c.leaveLoop()

	// post-body position
	postBodyPos := len(c.instructions)

	// post statement
	if stmt.Post != nil {
		if err := c.Compile(stmt.Post); err != nil {
			return err
		}
	}

	// back to condition
	c.emit(stmt, OpJump, preCondPos)

	// post-statement position
	postStmtPos := len(c.instructions)
	if postCondPos >= 0 {
		c.changeOperand(postCondPos, postStmtPos)
	}

	// update all break/continue jump positions
	for _, pos := range loop.Breaks {
		c.changeOperand(pos, postStmtPos)
	}
	for _, pos := range loop.Continues {
		c.changeOperand(pos, postBodyPos)
	}
	return nil
}

func (c *Compiler) compileForInStmt(stmt *parser.ForInStmt) error {
	nextIndex := c.symbolTable.NextIndex()
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		parent := c.symbolTable.Parent(false)
		// set undefined to variables having no reference
		maxSymbols := c.symbolTable.MaxSymbols()
		for i := nextIndex; i < maxSymbols; i++ {
			if !parent.IsIndexSkipped(i) {
				c.emit(stmt, OpNull)
				c.emit(stmt, OpSetLocal, i)
			}
		}
		c.symbolTable = parent
	}()

	// for-in statement is compiled like following:
	//
	//   for :it := iterator(iterable); :it.next();  {
	//     k, v := :it.get()  // set locals
	//
	//     ... body ...
	//   }
	//
	// ":it" is a local variable but it will not conflict with other user variables
	// because character ":" is not allowed in the variable names.

	// init
	//   :it = iterator(iterable)
	itSymbol, exists := c.symbolTable.DefineLocal(":it")
	if exists {
		return c.errorf(stmt, ":it redeclared in this block")
	}
	if err := c.Compile(stmt.Iterable); err != nil {
		return err
	}
	c.emit(stmt, OpIterInit)
	c.emit(stmt, OpSetLocal, itSymbol.Index)

	// pre-condition position
	preCondPos := len(c.instructions)

	// condition
	//  :it.Next()
	c.emit(stmt, OpGetLocal, itSymbol.Index)
	c.emit(stmt, OpIterNext)

	// condition jump position
	postCondPos := c.emit(stmt, OpJumpFalsy, 0)

	// enter loop
	loop := c.enterLoop()

	// assign key variable
	if stmt.Key.Name != "_" {
		keySymbol, exists := c.symbolTable.DefineLocal(stmt.Key.Name)
		if exists {
			return c.errorf(stmt, "%q redeclared in this block", stmt.Key.Name)
		}
		c.emit(stmt, OpGetLocal, itSymbol.Index)
		c.emit(stmt, OpIterKey)
		keySymbol.Assigned = true
		c.emit(stmt, OpSetLocal, keySymbol.Index)
	}

	// assign value variable
	if stmt.Value.Name != "_" {
		valueSymbol, exists := c.symbolTable.DefineLocal(stmt.Value.Name)
		if exists {
			return c.errorf(stmt, "%q redeclared in this block", stmt.Value.Name)
		}
		c.emit(stmt, OpGetLocal, itSymbol.Index)
		c.emit(stmt, OpIterValue)
		valueSymbol.Assigned = true
		c.emit(stmt, OpSetLocal, valueSymbol.Index)
	}

	// body statement
	if err := c.Compile(stmt.Body); err != nil {
		c.leaveLoop()
		return err
	}

	c.leaveLoop()

	// post-body position
	postBodyPos := len(c.instructions)

	// back to condition
	c.emit(stmt, OpJump, preCondPos)

	// post-statement position
	postStmtPos := len(c.instructions)
	c.changeOperand(postCondPos, postStmtPos)

	// update all break/continue jump positions
	for _, pos := range loop.Breaks {
		c.changeOperand(pos, postStmtPos)
	}
	for _, pos := range loop.Continues {
		c.changeOperand(pos, postBodyPos)
	}
	return nil
}

func (c *Compiler) checkCyclicImports(node parser.Node, modulePath string) error {
	if c.modulePath == modulePath {
		return c.errorf(node, "cyclic module import: %s", modulePath)
	} else if c.parent != nil {
		return c.parent.checkCyclicImports(node, modulePath)
	}
	return nil
}

func (c *Compiler) addModule(name string, constantIndex int) ModuleIndex {
	index := c.moduleIndexes.Count
	c.moduleIndexes.Count++
	c.moduleIndexes.Indexes[name] = ModuleIndex{
		ConstantIndex: constantIndex,
		ModuleIndex:   index,
	}
	return c.moduleIndexes.Indexes[name]
}

func (c *Compiler) getModule(name string) (ModuleIndex, bool) {
	indexes, ok := c.moduleIndexes.Indexes[name]
	return indexes, ok
}

func (c *Compiler) compileModule(node parser.Node,
	modulePath string, src []byte) (modIndex ModuleIndex, err error) {
	if err = c.checkCyclicImports(node, modulePath); err != nil {
		return
	}
	modIndex, exists := c.getModule(modulePath)
	if exists {
		return modIndex, nil
	}
	modFile := c.file.Set().AddFile(modulePath, -1, len(src))
	var trace io.Writer
	if c.opts.TraceParser {
		trace = c.trace
	}
	p := parser.NewParser(modFile, src, trace)
	var file *parser.File
	file, err = p.ParseFile()
	if err != nil {
		return
	}
	symbolTable := NewSymbolTable().
		DisableBuiltin(c.symbolTable.DisabledBuiltins()...)
	fork := c.fork(modFile, modulePath, symbolTable)
	_, err = fork.optimize(file)
	if err != nil {
		err = c.error(node, err)
		return
	}
	if err = fork.Compile(file); err != nil {
		return
	}
	bc := fork.Bytecode()
	if bc.Main.NumLocals > 256 {
		err = c.error(node, ErrSymbolLimit)
		return
	}
	c.constants = bc.Constants
	index := c.addConstant(bc.Main)
	return c.addModule(modulePath, index), nil
}

func (c *Compiler) enterLoop() *loopStmts {
	loop := &loopStmts{
		lastTryCatchIndex: c.tryCatchIndex,
	}
	c.loops = append(c.loops, loop)
	c.loopIndex++
	if c.trace != nil {
		c.printTrace("LOOPE", c.loopIndex)
	}
	return loop
}

func (c *Compiler) leaveLoop() {
	if c.trace != nil {
		c.printTrace("LOOPL", c.loopIndex)
	}
	c.loops = c.loops[:len(c.loops)-1]
	c.loopIndex--
}

func (c *Compiler) currentLoop() *loopStmts {
	if c.loopIndex >= 0 {
		return c.loops[c.loopIndex]
	}
	return nil
}

func (c *Compiler) fork(file *parser.SourceFile, modulePath string,
	symbolTable *SymbolTable) *Compiler {
	child := NewCompiler(file, CompilerOptions{
		ModuleMap:         c.moduleMap,
		ModuleIndexes:     c.moduleIndexes,
		ModulePath:        modulePath,
		Constants:         c.constants,
		SymbolTable:       symbolTable,
		Trace:             c.trace,
		TraceParser:       c.opts.TraceParser,
		TraceCompiler:     c.opts.TraceCompiler,
		TraceOptimizer:    c.opts.TraceOptimizer,
		OptimizerMaxCycle: c.opts.OptimizerMaxCycle,
		OptimizeConst:     c.opts.OptimizeConst,
		OptimizeExpr:      c.opts.OptimizeExpr,
		constsCache:       c.constsCache,
	})
	child.parent = c
	if modulePath == c.modulePath {
		child.indent = c.indent
	}
	return child
}

func (c *Compiler) error(node parser.Node, err error) error {
	return &CompilerError{
		FileSet: c.file.Set(),
		Node:    node,
		Err:     err,
	}
}

func (c *Compiler) errorf(node parser.Node,
	format string, args ...interface{}) error {

	return &CompilerError{
		FileSet: c.file.Set(),
		Node:    node,
		Err:     fmt.Errorf(format, args...),
	}
}

func (c *Compiler) printTrace(a ...interface{}) {
	const (
		dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
		n    = len(dots)
	)

	i := 2 * c.indent
	for i > n {
		_, _ = fmt.Fprint(c.trace, dots)
		i -= n
	}
	_, _ = fmt.Fprint(c.trace, dots[0:i])
	_, _ = fmt.Fprintln(c.trace, a...)
}

func tracec(c *Compiler, msg string) *Compiler {
	c.printTrace(msg, "{")
	c.indent++
	return c
}

func untracec(c *Compiler) {
	c.indent--
	c.printTrace("}")
}

// MakeInstruction returns a bytecode for an opcode and the operands.
func MakeInstruction(op Opcode, args ...int) ([]byte, error) {
	operands := OpcodeOperands[op]
	if len(operands) != len(args) {
		return nil, fmt.Errorf("MakeInstruction: %s expected %d operands, but got %d",
			OpcodeNames[op], len(operands), len(args))
	}
	switch op {
	case OpConstant, OpMap, OpArray,
		OpGetGlobal, OpSetGlobal,
		OpJump, OpJumpFalsy, OpAndJump, OpOrJump,
		OpStoreModule:
		inst := make([]byte, 3)
		inst[0] = op
		inst[1] = byte(args[0] >> 8)
		inst[2] = byte(args[0])
		return inst, nil
	case OpLoadModule, OpSetupTry:
		inst := make([]byte, 5)
		inst[0] = op
		inst[1] = byte(args[0] >> 8)
		inst[2] = byte(args[0])
		inst[3] = byte(args[1] >> 8)
		inst[4] = byte(args[1])
		return inst, nil
	case OpClosure:
		inst := make([]byte, 4)
		inst[0] = op
		inst[1] = byte(args[0] >> 8)
		inst[2] = byte(args[0])
		inst[3] = byte(args[1])
		return inst, nil
	case OpCall:
		inst := make([]byte, 3)
		inst[0] = op
		inst[1] = byte(args[0])
		inst[2] = byte(args[1])
		return inst, nil
	case OpGetBuiltin, OpReturn,
		OpBinaryOp, OpUnary,
		OpGetIndex,
		OpGetLocal, OpSetLocal,
		OpGetFree, OpSetFree,
		OpGetLocalPtr, OpGetFreePtr,
		OpThrow, OpFinalizer:
		inst := make([]byte, 2)
		inst[0] = op
		inst[1] = byte(args[0])
		return inst, nil
	case OpEqual, OpNotEqual, OpNull,
		OpPop, OpSliceIndex, OpSetIndex,
		OpIterInit, OpIterNext, OpIterKey, OpIterValue,
		OpSetupCatch, OpSetupFinally:
		return []byte{op}, nil
	default:
		panic(fmt.Errorf("MakeInstruction: unknown Opcode %d %s",
			op, OpcodeNames[op]))
	}
}

// FormatInstructions returns string representation of bytecode instructions.
func FormatInstructions(b []byte, posOffset int) []string {
	var out []string
	var operands = make([]int, 0, 4)
	var offset int
	var i int
	for i < len(b) {
		numOperands := OpcodeOperands[b[i]]
		operands, offset = ReadOperands(numOperands, b[i+1:], operands)

		switch len(numOperands) {
		case 0:
			out = append(out, fmt.Sprintf("%04d %-7s",
				posOffset+i, OpcodeNames[b[i]]))
		case 1:
			out = append(out, fmt.Sprintf("%04d %-7s %-5d",
				posOffset+i, OpcodeNames[b[i]], operands[0]))
		case 2:
			out = append(out, fmt.Sprintf("%04d %-7s %-5d %-5d",
				posOffset+i, OpcodeNames[b[i]],
				operands[0], operands[1]))
		}
		i += 1 + offset
	}
	return out
}

// IterateInstructions iterate instructions and call given function for each instruction.
// Note: Do not use operands slice in callback, it is reused for less allocation.
func IterateInstructions(insts []byte,
	fn func(pos int, opcode Opcode, operands []int, offset int) bool) {
	operands := make([]int, 0, 4)
	var offset int
	for i := 0; i < len(insts); i++ {
		numOperands := OpcodeOperands[insts[i]]
		operands, offset = ReadOperands(numOperands, insts[i+1:], operands)
		if !fn(i, insts[i], operands, offset) {
			break
		}
		i += offset
	}
}
