// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ozanh/ugo/parser"
	"github.com/ozanh/ugo/token"
)

// OptimizerError represents an optimizer error.
type OptimizerError struct {
	FileSet *parser.SourceFileSet
	Node    parser.Node
	Err     error
}

func (e *OptimizerError) Error() string {
	filePos := e.FileSet.Position(e.Node.Pos())
	return fmt.Sprintf("Optimizer Error: %s\n\tat %s", e.Err.Error(), filePos)
}

func (e *OptimizerError) Unwrap() error {
	return e.Err
}

// SimpleOptimizer optimizes given parsed file by evaluating constants and expressions.
type SimpleOptimizer struct {
	ctx              context.Context
	file             *parser.File
	vm               *VM
	count            int
	total            int
	maxCycle         int
	indent           int
	optimConsts      bool
	optimExpr        bool
	disabledBuiltins []string
	duration         time.Duration
	errors           multipleErr
	trace            io.Writer
}

// NewOptimizer creates an Optimizer object.
func NewOptimizer(
	ctx context.Context,
	file *parser.File,
	optimConst bool,
	optimExpr bool,
	maxCycle int,
	trace io.Writer,
) *SimpleOptimizer {

	return &SimpleOptimizer{
		ctx:         ctx,
		file:        file,
		vm:          NewVM(nil),
		trace:       trace,
		maxCycle:    maxCycle,
		optimConsts: optimConst,
		optimExpr:   optimExpr,
	}
}

// DisableBuiltins passes disabled builtins to symbol table.
func (opt *SimpleOptimizer) DisableBuiltins(names []string) *SimpleOptimizer {
	opt.disabledBuiltins = names
	return opt
}

// File returns parsed file which is modified after calling Optimize().
func (opt *SimpleOptimizer) File() *parser.File {
	return opt.file
}

func (*SimpleOptimizer) canOptimizeExpr(expr parser.Expr) bool {
	if parser.IsStatement(expr) {
		return false
	}
	switch expr.(type) {
	case *parser.BoolLit,
		*parser.IntLit,
		*parser.UintLit,
		*parser.FloatLit,
		*parser.CharLit,
		*parser.StringLit,
		*parser.UndefinedLit:
		return false
	}
	return true
}

func (*SimpleOptimizer) canOptimizeInsts(constants []Object, insts []byte) bool {
	if len(insts) == 0 {
		return false
	}
	// using array here instead of map or slice is faster to look up opcode
	allowedOps := [...]bool{
		OpConstant: true, OpNull: true, OpBinaryOp: true, OpUnary: true,
		OpNoOp: true, OpAndJump: true, OpOrJump: true, OpArray: true,
		OpReturn: true, OpEqual: true, OpNotEqual: true, OpPop: true,
		OpGetBuiltin: true, OpCall: true, OpSetLocal: true,
		^byte(0): false,
	}
	allowedBuiltins := [...]bool{
		BuiltinContains: true, BuiltinBool: true, BuiltinInt: true,
		BuiltinUint: true, BuiltinChar: true, BuiltinFloat: true,
		BuiltinString: true, BuiltinChars: true, BuiltinLen: true,
		BuiltinTypeName: true, BuiltinBytes: true, BuiltinError: true,
		BuiltinSprintf: true,
		BuiltinIsError: true, BuiltinIsInt: true, BuiltinIsUint: true,
		BuiltinIsFloat: true, BuiltinIsChar: true, BuiltinIsBool: true,
		BuiltinIsString: true, BuiltinIsBytes: true, BuiltinIsMap: true,
		BuiltinIsArray: true, BuiltinIsUndefined: true, BuiltinIsIterable: true,
		^byte(0): false,
	}
	canOptimize := true
	IterateInstructions(insts,
		func(_ int, opcode Opcode, operands []int, _ int) bool {
			if !allowedOps[opcode] {
				canOptimize = false
				return false
			}
			if opcode == OpConstant &&
				!isObjectConstant(constants[operands[0]]) {
				canOptimize = false
				return false
			}
			if opcode == OpGetBuiltin &&
				!allowedBuiltins[operands[0]] {
				canOptimize = false
				return false
			}
			return true
		})
	return canOptimize
}

func (opt *SimpleOptimizer) evalExpr(expr parser.Expr) (parser.Expr, bool) {
	if !opt.optimExpr {
		return nil, false
	}
	if opt.trace != nil {
		opt.printTraceMsg(fmt.Sprintf("Eval:%s", expr))
	}
	if !opt.canOptimizeExpr(expr) {
		if opt.trace != nil {
			opt.printTraceMsg("cannot optimize expression")
		}
		return nil, false
	}
	return opt.slowEvalExpr(expr)
}

func (opt *SimpleOptimizer) slowEvalExpr(expr parser.Expr) (parser.Expr, bool) {
	var trace io.Writer
	if opt.trace != nil {
		trace = opt.trace
	}
	st := NewSymbolTable().
		EnableParams(false).
		DisableBuiltin(opt.disabledBuiltins...)

	compiler := NewCompiler(
		opt.file.InputFile,
		CompilerOptions{
			SymbolTable: st,
			Trace:       trace,
		},
	)

	compiler.indent = opt.indent

	f := &parser.File{
		Stmts: []parser.Stmt{
			&parser.ReturnStmt{
				Result: expr,
			},
		},
	}

	if err := compiler.Compile(f); err != nil {
		return nil, false
	}
	bytecode := compiler.Bytecode()
	if !opt.canOptimizeInsts(bytecode.Constants, bytecode.Main.Instructions) {
		if opt.trace != nil {
			opt.printTraceMsg("cannot optimize instructions")
		}
		return nil, false
	}
	obj, err := opt.vm.SetBytecode(bytecode).Clear().Run(nil)
	if err != nil {
		if opt.trace != nil {
			opt.printTraceMsg(fmt.Sprintf("eval error: %s", err))
		}
		if !errors.Is(err, ErrVMAborted) {
			opt.errors = append(opt.errors, opt.error(expr, err))
		}
		return nil, false
	}
	switch v := obj.(type) {
	case String:
		opt.count++
		l := strconv.Quote(string(v))
		return &parser.StringLit{
			Value:    string(v),
			Literal:  l,
			ValuePos: expr.Pos(),
		}, true
	case undefined:
		opt.count++
		return &parser.UndefinedLit{TokenPos: expr.Pos()}, true
	case Bool:
		opt.count++
		l := strconv.FormatBool(bool(v))
		return &parser.BoolLit{
			Value:    bool(v),
			Literal:  l,
			ValuePos: expr.Pos(),
		}, true
	case Int:
		opt.count++
		l := strconv.FormatInt(int64(v), 10)
		return &parser.IntLit{
			Value:    int64(v),
			Literal:  l,
			ValuePos: expr.Pos(),
		}, true
	case Uint:
		opt.count++
		l := strconv.FormatUint(uint64(v), 10)
		return &parser.UintLit{
			Value:    uint64(v),
			Literal:  l,
			ValuePos: expr.Pos(),
		}, true
	case Float:
		opt.count++
		l := strconv.FormatFloat(float64(v), 'f', -1, 64)
		return &parser.FloatLit{
			Value:    float64(v),
			Literal:  l,
			ValuePos: expr.Pos(),
		}, true
	case Char:
		opt.count++
		l := strconv.QuoteRune(rune(v))
		return &parser.CharLit{
			Value:    rune(v),
			Literal:  l,
			ValuePos: expr.Pos(),
		}, true
	}
	return nil, false
}

// Optimize optimizes ast tree by simple constant folding and evaluating simple expressions.
func (opt *SimpleOptimizer) Optimize() error {
	opt.errors = nil
	opt.duration = 0
	if opt.ctx != nil {
		defer close(opt.abortVM())
	}

	if opt.trace != nil {
		opt.printTraceMsg("Enter Optimizer")
		defer func() {
			opt.printTraceMsg(fmt.Sprintf("File: %s", opt.file))
			opt.printTraceMsg(fmt.Sprintf("Duration: %s", opt.duration))
			opt.printTraceMsg("Exit Optimizer")
			opt.printTraceMsg("----------------------")
		}()
	}
	start := time.Now()
	i := 1
	for i <= opt.maxCycle {
		opt.count = 0
		if opt.trace != nil {
			opt.printTraceMsg(fmt.Sprintf("%d. pass", i))
		}
		opt.optimize(opt.file)
		if opt.count == 0 {
			break
		}
		if len(opt.errors) == 3 {
			break
		}
		opt.total += opt.count
		i++
	}
	opt.duration = time.Since(start)
	if opt.trace != nil {
		if opt.total > 0 {
			opt.printTraceMsg(fmt.Sprintf("Total: %d", opt.total))
		} else {
			opt.printTraceMsg("No Optimization")
		}
	}
	if opt.errors == nil {
		return nil
	}
	return opt.errors
}

func (opt *SimpleOptimizer) abortVM() chan struct{} {
	done := make(chan struct{})
	go func() {
		select {
		case <-opt.ctx.Done():
			goto abort
		case <-done:
			goto abort
		}
	abort:
		if opt.vm != nil {
			opt.vm.Abort()
		}
	}()
	return done
}

func (opt *SimpleOptimizer) binaryopInts(op token.Token,
	left, right *parser.IntLit) (parser.Expr, bool) {

	var val int64
	switch op {
	case token.Add:
		val = left.Value + right.Value
		goto result
	case token.Sub:
		val = left.Value - right.Value
		goto result
	case token.Mul:
		val = left.Value * right.Value
		goto result
	case token.Quo:
		if right.Value == 0 {
			return nil, false
		}
		val = left.Value / right.Value
		goto result
	case token.Rem:
		val = left.Value % right.Value
		goto result
	case token.And:
		val = left.Value & right.Value
		goto result
	case token.Or:
		val = left.Value | right.Value
		goto result
	case token.Shl:
		val = left.Value << right.Value
		goto result
	case token.Shr:
		val = left.Value >> right.Value
		goto result
	case token.AndNot:
		val = left.Value &^ right.Value
		goto result
	}
	return nil, false
result:
	opt.count++
	l := strconv.FormatInt(val, 10)
	return &parser.IntLit{Value: val, Literal: l, ValuePos: left.ValuePos}, true
}

func (opt *SimpleOptimizer) binaryopFloats(op token.Token,
	left, right *parser.FloatLit) (parser.Expr, bool) {

	var val float64
	switch op {
	case token.Add:
		val = left.Value + right.Value
		goto result
	case token.Sub:
		val = left.Value - right.Value
		goto result
	case token.Mul:
		val = left.Value * right.Value
		goto result
	case token.Quo:
		if right.Value == 0 {
			return nil, false
		}
		val = left.Value / right.Value
		goto result
	}
	return nil, false
result:
	opt.count++
	l := strconv.FormatFloat(val, 'f', -1, 64)
	return &parser.FloatLit{
		Value:    val,
		Literal:  l,
		ValuePos: left.ValuePos,
	}, true
}

func (opt *SimpleOptimizer) binaryop(op token.Token,
	left, right parser.Expr) (parser.Expr, bool) {

	if !opt.optimConsts {
		return nil, false
	}
	switch left := left.(type) {
	case *parser.IntLit:
		if right, ok := right.(*parser.IntLit); ok {
			return opt.binaryopInts(op, left, right)
		}
	case *parser.FloatLit:
		if right, ok := right.(*parser.FloatLit); ok {
			return opt.binaryopFloats(op, left, right)
		}
	case *parser.StringLit:
		right, ok := right.(*parser.StringLit)
		if ok && op == token.Add {
			opt.count++
			v := left.Value + right.Value
			return &parser.StringLit{
				Value:    v,
				Literal:  strconv.Quote(v),
				ValuePos: left.ValuePos,
			}, true
		}
	}
	return nil, false
}

func (opt *SimpleOptimizer) unaryop(op token.Token,
	expr parser.Expr) (parser.Expr, bool) {

	if !opt.optimConsts {
		return nil, false
	}
	switch expr := expr.(type) {
	case *parser.IntLit:
		switch op {
		case token.Not:
			opt.count++
			v := expr.Value == 0
			return &parser.BoolLit{
				Value:    v,
				Literal:  strconv.FormatBool(v),
				ValuePos: expr.ValuePos,
			}, true
		case token.Sub:
			opt.count++
			v := -expr.Value
			l := strconv.FormatInt(v, 10)
			return &parser.IntLit{
				Value:    v,
				Literal:  l,
				ValuePos: expr.ValuePos,
			}, true
		case token.Xor:
			opt.count++
			v := ^expr.Value
			l := strconv.FormatInt(v, 10)
			return &parser.IntLit{
				Value:    v,
				Literal:  l,
				ValuePos: expr.ValuePos,
			}, true
		}
	case *parser.UintLit:
		switch op {
		case token.Not:
			opt.count++
			v := expr.Value == 0
			return &parser.BoolLit{
				Value:    v,
				Literal:  strconv.FormatBool(v),
				ValuePos: expr.ValuePos,
			}, true
		case token.Sub:
			opt.count++
			v := -expr.Value
			l := strconv.FormatUint(v, 10)
			return &parser.UintLit{
				Value:    v,
				Literal:  l,
				ValuePos: expr.ValuePos,
			}, true
		case token.Xor:
			opt.count++
			v := ^expr.Value
			l := strconv.FormatUint(v, 10)
			return &parser.UintLit{
				Value:    v,
				Literal:  l,
				ValuePos: expr.ValuePos,
			}, true
		}
	case *parser.FloatLit:
		switch op {
		case token.Sub:
			opt.count++
			v := -expr.Value
			l := strconv.FormatFloat(v, 'f', -1, 64)
			return &parser.FloatLit{
				Value:    v,
				Literal:  l,
				ValuePos: expr.ValuePos,
			}, true
		}
	}
	return nil, false
}

func (opt *SimpleOptimizer) optimize(node parser.Node) (parser.Expr, bool) {
	if opt.trace != nil {
		if node != nil {
			defer untraceoptim(traceoptim(opt, fmt.Sprintf("%s (%s)",
				node.String(), reflect.TypeOf(node).Elem().Name())))
		} else {
			defer untraceoptim(traceoptim(opt, "<nil>"))
		}
	}
	var (
		expr parser.Expr
		ok   bool
	)
	switch node := node.(type) {
	case *parser.File:
		for _, stmt := range node.Stmts {
			_, _ = opt.optimize(stmt)
		}
	case *parser.ExprStmt:
		if expr, ok = opt.optimize(node.Expr); ok {
			node.Expr = expr
		}
		if expr, ok = opt.evalExpr(node.Expr); ok {
			node.Expr = expr
		}
	case *parser.ParenExpr:
		return opt.optimize(node.Expr)
	case *parser.BinaryExpr:
		if expr, ok = opt.optimize(node.LHS); ok {
			node.LHS = expr
		}
		if expr, ok = opt.optimize(node.RHS); ok {
			node.RHS = expr
		}
		if expr, ok = opt.binaryop(node.Token, node.LHS, node.RHS); ok {
			return expr, ok
		}
		return opt.evalExpr(node)
	case *parser.UnaryExpr:
		if expr, ok = opt.optimize(node.Expr); ok {
			node.Expr = expr
		}
		if expr, ok = opt.unaryop(node.Token, node.Expr); ok {
			return expr, ok
		}
		return opt.evalExpr(node)
	case *parser.IfStmt:
		if node.Init != nil {
			_, _ = opt.optimize(node.Init)
		}
		if expr, ok = opt.optimize(node.Cond); ok {
			node.Cond = expr
		}
		if expr, ok = opt.evalExpr(node.Cond); ok {
			node.Cond = expr
		}
		falsy, ok := isLitFalsy(node.Cond)
		if ok {
			// convert expression to BoolLit so that Compiler skips if block
			node.Cond = &parser.BoolLit{
				Value:    !falsy,
				Literal:  strconv.FormatBool(!falsy),
				ValuePos: node.Cond.Pos(),
			}
		}
		if node.Body != nil {
			_, _ = opt.optimize(node.Body)
		}
		if node.Else != nil {
			_, _ = opt.optimize(node.Else)
		}
	case *parser.TryStmt:
		if node.Body != nil {
			_, _ = opt.optimize(node.Body)
		}
		if node.Catch != nil {
			_, _ = opt.optimize(node.Catch)
		}
		if node.Finally != nil {
			_, _ = opt.optimize(node.Finally)
		}
	case *parser.CatchStmt:
		if node.Body != nil {
			_, _ = opt.optimize(node.Body)
		}
	case *parser.FinallyStmt:
		if node.Body != nil {
			_, _ = opt.optimize(node.Body)
		}
	case *parser.ThrowStmt:
		if node.Expr != nil {
			if expr, ok = opt.optimize(node.Expr); ok {
				node.Expr = expr
			}
			if expr, ok = opt.evalExpr(node.Expr); ok {
				node.Expr = expr
			}
		}
	case *parser.ForStmt:
		if node.Init != nil {
			_, _ = opt.optimize(node.Init)
		}
		if node.Cond != nil {
			if expr, ok = opt.optimize(node.Cond); ok {
				node.Cond = expr
			}
		}
		if node.Post != nil {
			_, _ = opt.optimize(node.Post)
		}
		if node.Body != nil {
			_, _ = opt.optimize(node.Body)
		}
	case *parser.ForInStmt:
		if node.Body != nil {
			_, _ = opt.optimize(node.Body)
		}
	case *parser.BlockStmt:
		for _, stmt := range node.Stmts {
			_, _ = opt.optimize(stmt)
		}
	case *parser.AssignStmt:
		for i, rhs := range node.RHS {
			if expr, ok = opt.optimize(rhs); ok {
				node.RHS[i] = expr
			}
		}
		for i, rhs := range node.RHS {
			if expr, ok = opt.evalExpr(rhs); ok {
				node.RHS[i] = expr
			}
		}
	case *parser.DeclStmt:
		decl, ok := node.Decl.(*parser.GenDecl)
		if !ok {
			panic("only GenDecl is supported in DeclStmt")
		}
		switch decl.Tok {
		case token.Var:
			for _, sp := range decl.Specs {
				spec, ok := sp.(*parser.ValueSpec)
				if !ok {
					return nil, false
				}
				for i := range spec.Idents {
					var v parser.Expr
					if i < len(spec.Values) {
						v = spec.Values[i]
						if expr, ok = opt.optimize(v); ok {
							spec.Values[i] = expr
							v = expr
						}
						if expr, ok = opt.evalExpr(v); ok {
							spec.Values[i] = expr
						}
					}
				}
			}
		}
	case *parser.ArrayLit:
		for i := range node.Elements {
			if expr, ok = opt.optimize(node.Elements[i]); ok {
				node.Elements[i] = expr
			}
			if expr, ok = opt.evalExpr(node.Elements[i]); ok {
				node.Elements[i] = expr
			}
		}
	case *parser.MapLit:
		for i := range node.Elements {
			if expr, ok = opt.optimize(node.Elements[i].Value); ok {
				node.Elements[i].Value = expr
			}
			if expr, ok = opt.evalExpr(node.Elements[i].Value); ok {
				node.Elements[i].Value = expr
			}
		}
	case *parser.IndexExpr:
		if expr, ok = opt.optimize(node.Index); ok {
			node.Index = expr
		}
		if expr, ok = opt.evalExpr(node.Index); ok {
			node.Index = expr
		}
	case *parser.SliceExpr:
		if node.Low != nil {
			if expr, ok = opt.optimize(node.Low); ok {
				node.Low = expr
			}
			if expr, ok = opt.evalExpr(node.Low); ok {
				node.Low = expr
			}
		}
		if node.High != nil {
			if expr, ok = opt.optimize(node.High); ok {
				node.High = expr
			}
			if expr, ok = opt.evalExpr(node.High); ok {
				node.High = expr
			}
		}
	case *parser.FuncLit:
		_, _ = opt.optimize(node.Body)
	case *parser.ReturnStmt:
		if expr, ok = opt.optimize(node.Result); ok {
			node.Result = expr
		}
		if expr, ok = opt.evalExpr(node.Result); ok {
			node.Result = expr
		}
	case *parser.CallExpr:
		if node.Func != nil {
			_, _ = opt.optimize(node.Func)
		}
		for i := range node.Args {
			if expr, ok = opt.optimize(node.Args[i]); ok {
				node.Args[i] = expr
			}
			if expr, ok = opt.evalExpr(node.Args[i]); ok {
				node.Args[i] = expr
			}
		}
	case *parser.CondExpr:
		if expr, ok = opt.optimize(node.Cond); ok {
			node.Cond = expr
		}
		if expr, ok = opt.evalExpr(node.Cond); ok {
			node.Cond = expr
		}
		falsy, ok := isLitFalsy(node.Cond)
		if ok {
			// convert expression to BoolLit so that Compiler skips expressions
			node.Cond = &parser.BoolLit{
				Value:    !falsy,
				Literal:  strconv.FormatBool(!falsy),
				ValuePos: node.Cond.Pos(),
			}
		}

		if expr, ok = opt.optimize(node.True); ok {
			node.True = expr
		}
		if expr, ok = opt.evalExpr(node.True); ok {
			node.True = expr
		}
		if expr, ok = opt.optimize(node.False); ok {
			node.False = expr
		}
		if expr, ok = opt.evalExpr(node.False); ok {
			node.False = expr
		}
	}
	return nil, false
}

// Total returns total number of evaluated constants and expressions.
func (opt *SimpleOptimizer) Total() int {
	return opt.total
}

// Duration returns total elapsed time of Optimize() call.
func (opt *SimpleOptimizer) Duration() time.Duration {
	return opt.duration
}

func (opt *SimpleOptimizer) error(node parser.Node, err error) error {
	return &OptimizerError{
		FileSet: opt.file.InputFile.Set(),
		Node:    node,
		Err:     err,
	}
}

func (opt *SimpleOptimizer) printTrace(a ...interface{}) {
	const (
		dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
		n    = len(dots)
	)

	i := 2 * opt.indent
	for i > n {
		_, _ = fmt.Fprint(opt.trace, dots)
		i -= n
	}
	_, _ = fmt.Fprint(opt.trace, dots[0:i])
	_, _ = fmt.Fprintln(opt.trace, a...)
}

func (opt *SimpleOptimizer) printTraceMsg(a ...interface{}) {
	const (
		dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
		n    = len(dots)
	)

	i := 2 * opt.indent
	for i > n {
		_, _ = fmt.Fprint(opt.trace, dots)
		i -= n
	}
	_, _ = fmt.Fprint(opt.trace, dots[0:i], "<")
	_, _ = fmt.Fprint(opt.trace, a...)
	_, _ = fmt.Fprintln(opt.trace, ">")
}

func traceoptim(cf *SimpleOptimizer, msg string) *SimpleOptimizer {
	cf.printTrace(msg, "{")
	cf.indent++
	return cf
}

func untraceoptim(cf *SimpleOptimizer) {
	cf.indent--
	cf.printTrace("}")
}

func isObjectConstant(obj Object) bool {
	switch obj.(type) {
	case Bool, Int, Uint, Float, Char, String, undefined:
		return true
	}
	return false
}

func isLitFalsy(expr parser.Expr) (bool, bool) {
	if expr == nil {
		return false, false
	}
	switch v := expr.(type) {
	case *parser.BoolLit:
		return !v.Value, true
	case *parser.IntLit:
		return Int(v.Value).IsFalsy(), true
	case *parser.UintLit:
		return Uint(v.Value).IsFalsy(), true
	case *parser.FloatLit:
		return Float(v.Value).IsFalsy(), true
	case *parser.StringLit:
		return String(v.Value).IsFalsy(), true
	case *parser.CharLit:
		return Char(v.Value).IsFalsy(), true
	case *parser.UndefinedLit:
		return Undefined.IsFalsy(), true
	}
	return false, false
}

type multipleErr []error

func (m multipleErr) Error() string {
	if len(m) == 0 {
		return ""
	}
	if len(m) == 1 {
		return m[0].Error()
	}
	var sb strings.Builder
	sb.WriteString(m[0].Error())
	for _, err := range m[1:] {
		sb.WriteString("\n")
		sb.WriteString(err.Error())
	}
	return sb.String()
}
