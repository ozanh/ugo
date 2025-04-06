// Copyright (c) 2020-2025 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/ozanh/ugo/parser"
	"github.com/ozanh/ugo/token"
)

// OptimizerError represents an optimizer error.
type OptimizerError struct {
	FilePos parser.SourceFilePos
	Node    parser.Node
	Err     error
}

func (e *OptimizerError) Error() string {
	return fmt.Sprintf("Optimizer Error: %s\n\tat %s", e.Err.Error(), e.FilePos)
}

func (e *OptimizerError) Unwrap() error {
	return e.Err
}

type optimizerScope struct {
	parent          *optimizerScope
	shadowed        []string
	handleConstLits bool
}

func (s *optimizerScope) define(ident string) {
	if _, ok := BuiltinsMap[ident]; ok {
		s.shadowed = append(s.shadowed, ident)
	}
}

// SimpleOptimizer optimizes given parsed file by evaluating constants and
// expressions. It is not safe to call methods concurrently.
type SimpleOptimizer struct {
	file          *parser.SourceFile
	scope         *optimizerScope
	compSymTab    *SymbolTable
	eval          optimizerEval
	errors        multipleErr
	trace         io.Writer
	count         int
	total         int
	limit         int
	indent        int
	evalBits      uint64
	exprLevel     uint8
	traceCompiler bool
	traceParser   bool
}

// NewOptimizer creates an Optimizer object.
func NewOptimizer(
	file *parser.SourceFile,
	symTab *SymbolTable,
	opts CompilerOptions,
) *SimpleOptimizer {

	var trace io.Writer
	if opts.TraceOptimizer {
		trace = opts.Trace
	}

	return &SimpleOptimizer{
		file:          file,
		compSymTab:    symTab,
		limit:         opts.OptimizerLimit,
		trace:         trace,
		traceCompiler: opts.TraceCompiler,
		traceParser:   opts.TraceParser,
	}
}

func (so *SimpleOptimizer) reset(symTab *SymbolTable, limit int) {
	*so = SimpleOptimizer{
		file:          so.file,
		compSymTab:    symTab,
		eval:          so.eval,
		limit:         limit,
		trace:         so.trace,
		traceCompiler: so.traceCompiler,
		traceParser:   so.traceParser,
	}
}

func canOptimizeExpr(expr parser.Expr) bool {
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

func canOptimizeInsts(constants []Object, insts []byte) bool {
	if len(insts) == 0 {
		return false
	}

	// using array here instead of map or slice is faster to look up opcode
	allowedOps := [255]bool{
		OpConstant: true, OpNull: true, OpBinaryOp: true, OpUnary: true,
		OpNoOp: true, OpAndJump: true, OpOrJump: true, OpArray: true,
		OpReturn: true, OpEqual: true, OpNotEqual: true, OpPop: true,
		OpGetBuiltin: true, OpCall: true, OpSetLocal: true, OpDefineLocal: true,
		OpTrue: true, OpFalse: true,
	}

	allowedBuiltins := [255]bool{
		BuiltinContains: true, BuiltinBool: true, BuiltinInt: true,
		BuiltinUint: true, BuiltinChar: true, BuiltinFloat: true,
		BuiltinString: true, BuiltinChars: true, BuiltinLen: true,
		BuiltinTypeName: true, BuiltinBytes: true, BuiltinError: true,
		BuiltinSprintf: true,
		BuiltinIsError: true, BuiltinIsInt: true, BuiltinIsUint: true,
		BuiltinIsFloat: true, BuiltinIsChar: true, BuiltinIsBool: true,
		BuiltinIsString: true, BuiltinIsBytes: true, BuiltinIsMap: true,
		BuiltinIsArray: true, BuiltinIsUndefined: true, BuiltinIsIterable: true,
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
		},
	)
	return canOptimize
}

func (so *SimpleOptimizer) evalExpr(expr parser.Expr) (parser.Expr, bool) {
	if len(so.errors) > 0 {
		// do not evaluate erroneous line again
		prevPos := so.errors[len(so.errors)-1].(*OptimizerError).FilePos
		if so.file.Set().Position(expr.Pos()).Line == prevPos.Line {
			return nil, false
		}
	}
	istrace := so.trace != nil
	if istrace {
		so.printTraceMsgf("eval: %s", expr)
	}

	if !so.canEval() || !canOptimizeExpr(expr) {
		if istrace {
			so.printTraceMsgf("cannot optimize expression")
		}
		return nil, false
	}

	if istrace {
		so.printTraceMsgf("slow eval: %s", expr)
	}

	x, ok := so.slowEvalExpr(expr)
	if !ok {
		so.setNoEval()
		if istrace {
			so.printTraceMsgf("cannot optimize code")
		}
	} else {
		if istrace {
			so.printTraceMsgf("optimized code")
		}
		so.count++
	}
	return x, ok
}

func (so *SimpleOptimizer) slowEvalExpr(expr parser.Expr) (parser.Expr, bool) {
	obj, ok := so.eval.eval(so, expr)
	if !ok {
		return nil, false
	}
	switch v := obj.(type) {
	case String:
		expr = &parser.StringLit{
			Value:    string(v),
			Literal:  strconv.Quote(string(v)),
			ValuePos: expr.Pos(),
		}
	case *UndefinedType:
		expr = &parser.UndefinedLit{TokenPos: expr.Pos()}
	case Bool:
		expr = &parser.BoolLit{
			Value:    bool(v),
			Literal:  strconv.FormatBool(bool(v)),
			ValuePos: expr.Pos(),
		}
	case Int:
		expr = &parser.IntLit{
			Value:    int64(v),
			Literal:  strconv.FormatInt(int64(v), 10),
			ValuePos: expr.Pos(),
		}
	case Uint:
		expr = &parser.UintLit{
			Value:    uint64(v),
			Literal:  strconv.FormatUint(uint64(v), 10),
			ValuePos: expr.Pos(),
		}
	case Float:
		expr = &parser.FloatLit{
			Value:    float64(v),
			Literal:  strconv.FormatFloat(float64(v), 'f', -1, 64),
			ValuePos: expr.Pos(),
		}
	case Char:
		expr = &parser.CharLit{
			Value:    rune(v),
			Literal:  strconv.QuoteRune(rune(v)),
			ValuePos: expr.Pos(),
		}
	default:
		return nil, false
	}
	return expr, true
}

func (so *SimpleOptimizer) canEval() bool {
	// if left bits are set, we should not eval, pointless
	return so.evalBits>>so.exprLevel == 0
}

func (so *SimpleOptimizer) setNoEval() {
	// set level bit to 1, we got an eval error
	so.evalBits |= 1 << (so.exprLevel - 1)
}

func (so *SimpleOptimizer) enterExprLevel() {
	// clear bits on the left
	shift := 64 - so.exprLevel
	so.evalBits = so.evalBits << shift >> shift
	so.exprLevel++
	// if opt.trace != nil {
	// 	opt.printTraceMsgf(fmt.Sprintf("level:%d %064b", opt.exprLevel, opt.evalBits))
	// }
}

func (so *SimpleOptimizer) leaveExprLevel() {
	// if opt.trace != nil {
	// 	opt.printTraceMsgf(fmt.Sprintf("level:%d %064b", opt.exprLevel, opt.evalBits))
	// }
	so.exprLevel--
}

// Optimize optimizes ast tree by simple constant folding and evaluating simple expressions.
func (so *SimpleOptimizer) Optimize(node parser.Node) error {
	const methodName = "Optimize"
	const singlePass = false

	err := so.optimize(
		methodName,
		singlePass,
		func() {
			so.transform(node)
		},
	)
	return err
}

func (so *SimpleOptimizer) optimizeExpr(expr parser.Expr) (parser.Expr, bool, error) {
	const methodName = "optimizeExpr"
	const singlePass = true

	var out parser.Expr
	var ok bool

	err := so.optimize(
		methodName,
		singlePass,
		func() {
			so.scope.handleConstLits = true

			out, ok = so.transform(expr)
		},
	)
	return out, ok, err
}

func (so *SimpleOptimizer) optimize(
	methodName string,
	singlePass bool,
	runFn func(),
) error {

	if so.limit <= 0 {
		return nil
	}

	so.errors = nil

	if so.trace != nil {
		so.printTraceMsgf("Enter " + methodName)
	}

	pass := 1
	for so.limit > 0 {
		so.count = 0
		so.exprLevel = 0
		if so.trace != nil {
			so.printTraceMsgf("%d. pass", pass)
			pass++
		}
		so.enterScope()
		runFn()
		so.leaveScope()

		if so.count == 0 {
			break
		}

		if len(so.errors) > 2 { // bailout
			break
		}
		so.total += so.count
		so.limit -= so.count

		if singlePass {
			break
		}
	}

	if so.trace != nil {
		if so.total > 0 {
			so.printTraceMsgf("Total: %d", so.total)
		} else {
			so.printTraceMsgf("No Optimization")
		}
		so.printTraceMsgf("File: %s", so.file.Name)
		so.printTraceMsgf("Exit " + methodName)
		so.printTraceMsgf("----------------------")
	}

	return so.errors.errorOrNil()
}

func (so *SimpleOptimizer) binaryopInts(
	op token.Token,
	left *parser.IntLit,
	right *parser.IntLit,
) (parser.Expr, bool) {

	var val int64
	switch op {
	case token.Add:
		val = left.Value + right.Value
	case token.Sub:
		val = left.Value - right.Value
	case token.Mul:
		val = left.Value * right.Value
	case token.Quo:
		if right.Value == 0 {
			return nil, false
		}
		val = left.Value / right.Value
	case token.Rem:
		val = left.Value % right.Value
	case token.And:
		val = left.Value & right.Value
	case token.Or:
		val = left.Value | right.Value
	case token.Shl:
		val = left.Value << right.Value
	case token.Shr:
		val = left.Value >> right.Value
	case token.AndNot:
		val = left.Value &^ right.Value
	default:
		return nil, false
	}
	l := strconv.FormatInt(val, 10)
	return &parser.IntLit{Value: val, Literal: l, ValuePos: left.ValuePos}, true
}

func (so *SimpleOptimizer) binaryopFloats(
	op token.Token,
	left *parser.FloatLit,
	right *parser.FloatLit,
) (parser.Expr, bool) {

	var val float64
	switch op {
	case token.Add:
		val = left.Value + right.Value
	case token.Sub:
		val = left.Value - right.Value
	case token.Mul:
		val = left.Value * right.Value
	case token.Quo:
		if right.Value == 0 {
			return nil, false
		}
		val = left.Value / right.Value
	default:
		return nil, false
	}

	return &parser.FloatLit{
		Value:    val,
		Literal:  strconv.FormatFloat(val, 'f', -1, 64),
		ValuePos: left.ValuePos,
	}, true
}

func (so *SimpleOptimizer) binaryop(
	op token.Token,
	left parser.Expr,
	right parser.Expr,
) (parser.Expr, bool) {

	switch left := left.(type) {
	case *parser.IntLit:
		if right, ok := right.(*parser.IntLit); ok {
			return so.binaryopInts(op, left, right)
		}
	case *parser.FloatLit:
		if right, ok := right.(*parser.FloatLit); ok {
			return so.binaryopFloats(op, left, right)
		}
	case *parser.StringLit:
		right, ok := right.(*parser.StringLit)
		if ok && op == token.Add {
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

func (so *SimpleOptimizer) unaryop(
	op token.Token,
	expr parser.Expr,
) (parser.Expr, bool) {

	switch expr := expr.(type) {
	case *parser.IntLit:
		switch op {
		case token.Not:
			v := expr.Value == 0
			return &parser.BoolLit{
				Value:    v,
				Literal:  strconv.FormatBool(v),
				ValuePos: expr.ValuePos,
			}, true
		case token.Sub:
			v := -expr.Value
			l := strconv.FormatInt(v, 10)
			return &parser.IntLit{
				Value:    v,
				Literal:  l,
				ValuePos: expr.ValuePos,
			}, true
		case token.Xor:
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
			v := expr.Value == 0
			return &parser.BoolLit{
				Value:    v,
				Literal:  strconv.FormatBool(v),
				ValuePos: expr.ValuePos,
			}, true
		case token.Sub:
			v := -expr.Value
			l := strconv.FormatUint(v, 10)
			return &parser.UintLit{
				Value:    v,
				Literal:  l,
				ValuePos: expr.ValuePos,
			}, true
		case token.Xor:
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

func (so *SimpleOptimizer) transform(node parser.Node) (parser.Expr, bool) {
	if so.trace != nil {
		if node != nil {
			defer untraceoptim(traceoptim(so, fmt.Sprintf("%s (%s)",
				node.String(), reflect.TypeOf(node).Elem().Name())))
		} else {
			defer untraceoptim(traceoptim(so, "<nil>"))
		}
	}

	if !parser.IsStatement(node) {
		so.enterExprLevel()
		defer so.leaveExprLevel()
	}

	var (
		expr parser.Expr
		ok   bool
	)

	switch node := node.(type) {
	case *parser.File:
		for _, stmt := range node.Stmts {
			_, _ = so.transform(stmt)
		}
	case *parser.ExprStmt:
		if node.Expr != nil {
			if expr, ok = so.transform(node.Expr); ok {
				node.Expr = expr
			}
			if expr, ok = so.evalExpr(node.Expr); ok {
				node.Expr = expr
			}
		}
	case *parser.ParenExpr:
		if node.Expr != nil {
			return so.transform(node.Expr)
		}
	case *parser.BinaryExpr:
		if expr, ok = so.transform(node.LHS); ok {
			node.LHS = expr
		}
		if expr, ok = so.transform(node.RHS); ok {
			node.RHS = expr
		}
		if expr, ok = so.binaryop(node.Token, node.LHS, node.RHS); ok {
			so.count++
			return expr, ok
		}
		return so.evalExpr(node)
	case *parser.UnaryExpr:
		if expr, ok = so.transform(node.Expr); ok {
			node.Expr = expr
		}
		if expr, ok = so.unaryop(node.Token, node.Expr); ok {
			so.count++
			return expr, ok
		}
		return so.evalExpr(node)
	case *parser.IfStmt:
		if node.Init != nil {
			_, _ = so.transform(node.Init)
		}
		if expr, ok = so.transform(node.Cond); ok {
			node.Cond = expr
		}
		if expr, ok = so.evalExpr(node.Cond); ok {
			node.Cond = expr
		}
		if falsy, ok := isLiteralFalsy(node.Cond); ok {
			// convert expression to BoolLit so that Compiler skips if block
			node.Cond = &parser.BoolLit{
				Value:    !falsy,
				Literal:  strconv.FormatBool(!falsy),
				ValuePos: node.Cond.Pos(),
			}
		}
		if node.Body != nil {
			_, _ = so.transform(node.Body)
		}
		if node.Else != nil {
			_, _ = so.transform(node.Else)
		}
	case *parser.TryStmt:
		if node.Body != nil {
			_, _ = so.transform(node.Body)
		}
		if node.Catch != nil {
			_, _ = so.transform(node.Catch)
		}
		if node.Finally != nil {
			_, _ = so.transform(node.Finally)
		}
	case *parser.CatchStmt:
		if node.Body != nil {
			_, _ = so.transform(node.Body)
		}
	case *parser.FinallyStmt:
		if node.Body != nil {
			_, _ = so.transform(node.Body)
		}
	case *parser.ThrowStmt:
		if node.Expr != nil {
			if expr, ok = so.transform(node.Expr); ok {
				node.Expr = expr
			}
			if expr, ok = so.evalExpr(node.Expr); ok {
				node.Expr = expr
			}
		}
	case *parser.ForStmt:
		if node.Init != nil {
			_, _ = so.transform(node.Init)
		}
		if node.Cond != nil {
			if expr, ok = so.transform(node.Cond); ok {
				node.Cond = expr
			}
		}
		if node.Post != nil {
			_, _ = so.transform(node.Post)
		}
		if node.Body != nil {
			_, _ = so.transform(node.Body)
		}
	case *parser.ForInStmt:
		if node.Body != nil {
			_, _ = so.transform(node.Body)
		}
	case *parser.BlockStmt:
		for _, stmt := range node.Stmts {
			_, _ = so.transform(stmt)
		}
	case *parser.AssignStmt:
		for _, lhs := range node.LHS {
			if ident, ok := lhs.(*parser.Ident); ok {
				so.scope.define(ident.Name)
			}
		}
		for i, rhs := range node.RHS {
			if expr, ok = so.transform(rhs); ok {
				node.RHS[i] = expr
			}
		}
		for i, rhs := range node.RHS {
			if expr, ok = so.evalExpr(rhs); ok {
				node.RHS[i] = expr
			}
		}
	case *parser.DeclStmt:
		decl := node.Decl.(*parser.GenDecl)
		switch decl.Tok {
		case token.Param, token.Global:
			for _, sp := range decl.Specs {
				spec := sp.(*parser.ParamSpec)
				so.scope.define(spec.Ident.Name)
			}
		case token.Var, token.Const:
			for _, sp := range decl.Specs {
				spec := sp.(*parser.ValueSpec)
				for i := range spec.Idents {
					so.scope.define(spec.Idents[i].Name)
					if i < len(spec.Values) && spec.Values[i] != nil {
						v := spec.Values[i]
						if expr, ok = so.transform(v); ok {
							spec.Values[i] = expr
							v = expr
						}
						if expr, ok = so.evalExpr(v); ok {
							spec.Values[i] = expr
						}
					}
				}
			}
		}
	case *parser.ArrayLit:
		for i := range node.Elements {
			if expr, ok = so.transform(node.Elements[i]); ok {
				node.Elements[i] = expr
			}
			if expr, ok = so.evalExpr(node.Elements[i]); ok {
				node.Elements[i] = expr
			}
		}
	case *parser.MapLit:
		for i := range node.Elements {
			if expr, ok = so.transform(node.Elements[i].Value); ok {
				node.Elements[i].Value = expr
			}
			if expr, ok = so.evalExpr(node.Elements[i].Value); ok {
				node.Elements[i].Value = expr
			}
		}
	case *parser.IndexExpr:
		if expr, ok = so.transform(node.Index); ok {
			node.Index = expr
		}
		if expr, ok = so.evalExpr(node.Index); ok {
			node.Index = expr
		}
	case *parser.SliceExpr:
		if node.Low != nil {
			if expr, ok = so.transform(node.Low); ok {
				node.Low = expr
			}
			if expr, ok = so.evalExpr(node.Low); ok {
				node.Low = expr
			}
		}
		if node.High != nil {
			if expr, ok = so.transform(node.High); ok {
				node.High = expr
			}
			if expr, ok = so.evalExpr(node.High); ok {
				node.High = expr
			}
		}
	case *parser.FuncLit:
		if so.scope.handleConstLits {
			return nil, false
		}

		so.enterScope()
		defer so.leaveScope()
		for _, ident := range node.Type.Params.List {
			so.scope.define(ident.Name)
		}
		if node.Body != nil {
			_, _ = so.transform(node.Body)
		}
	case *parser.ReturnStmt:
		if node.Result != nil {
			if expr, ok = so.transform(node.Result); ok {
				node.Result = expr
			}
			if expr, ok = so.evalExpr(node.Result); ok {
				node.Result = expr
			}
		}
	case *parser.CallExpr:
		if node.Func != nil {
			_, _ = so.transform(node.Func)
		}
		for i := range node.Args {
			if expr, ok = so.transform(node.Args[i]); ok {
				node.Args[i] = expr
			}
			if expr, ok = so.evalExpr(node.Args[i]); ok {
				node.Args[i] = expr
			}
		}
	case *parser.CondExpr:
		if expr, ok = so.transform(node.Cond); ok {
			node.Cond = expr
		}
		if expr, ok = so.evalExpr(node.Cond); ok {
			node.Cond = expr
		}
		if falsy, ok := isLiteralFalsy(node.Cond); ok {
			// convert expression to BoolLit so that Compiler skips expressions
			node.Cond = &parser.BoolLit{
				Value:    !falsy,
				Literal:  strconv.FormatBool(!falsy),
				ValuePos: node.Cond.Pos(),
			}
		}

		if expr, ok = so.transform(node.True); ok {
			node.True = expr
		}
		if expr, ok = so.evalExpr(node.True); ok {
			node.True = expr
		}
		if expr, ok = so.transform(node.False); ok {
			node.False = expr
		}
		if expr, ok = so.evalExpr(node.False); ok {
			node.False = expr
		}
	case *parser.Ident:
		if so.scope.handleConstLits {
			s := findSymbol(so.compSymTab, node.Name, ScopeConstLit)
			if s != nil && s.Assigned && s.Constant {
				return s.constLit.toExpr(), true
			}
		}
	}
	return nil, false
}

func (so *SimpleOptimizer) enterScope() {
	so.scope = &optimizerScope{parent: so.scope}
}

func (so *SimpleOptimizer) leaveScope() {
	so.scope = so.scope.parent
}

// Total returns total number of evaluated constants and expressions.
func (so *SimpleOptimizer) Total() int {
	return so.total
}

func (so *SimpleOptimizer) error(node parser.Node, err error) error {
	pos := so.file.Set().Position(node.Pos())
	return &OptimizerError{
		FilePos: pos,
		Node:    node,
		Err:     err,
	}
}

func (so *SimpleOptimizer) printTraceMsgf(format string, args ...interface{}) {
	const (
		dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
		n    = len(dots)
	)

	i := 2 * so.indent
	for i > n {
		_, _ = fmt.Fprint(so.trace, dots)
		i -= n
	}

	_, _ = fmt.Fprint(so.trace, dots[0:i], "<")
	_, _ = fmt.Fprintf(so.trace, format, args...)
	_, _ = fmt.Fprintln(so.trace, ">")
}

func traceoptim(so *SimpleOptimizer, msg string) *SimpleOptimizer {
	printTrace(so.indent, so.trace, msg, "{")
	so.indent++
	return so
}

func untraceoptim(so *SimpleOptimizer) {
	so.indent--
	printTrace(so.indent, so.trace, "}")
}

func isObjectConstant(obj Object) bool {
	switch obj.(type) {
	case Bool, Int, Uint, Float, Char, String, *UndefinedType:
		return true
	}
	return false
}

func isLiteralFalsy(expr parser.Expr) (bool, bool) {
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

type optimizerEval struct {
	vm         *VM
	compiler   Compiler
	global     Map
	symtab     *SymbolTable
	returnStmt parser.ReturnStmt
}

func (ev *optimizerEval) resetCompiler(so *SimpleOptimizer) {
	if ev.symtab == nil {
		ev.symtab = NewSymbolTable()
	} else {
		ev.symtab.reset()
	}

	ev.symtab.EnableParams(false)
	optimCopyBuiltinStates(ev.symtab, so.compSymTab)
	optimCopyBuiltinStatesFromScope(ev.symtab, so.scope)

	ev.compiler = Compiler{
		file:          so.file,
		constants:     ev.compiler.constants[:0],
		symbolTable:   ev.symtab,
		constsCache:   ev.compiler.constsCache,
		cfuncCache:    ev.compiler.cfuncCache,
		instructions:  ev.compiler.instructions[:0],
		sourceMap:     ev.compiler.sourceMap,
		moduleMap:     ev.compiler.moduleMap,
		moduleStore:   ev.compiler.moduleStore,
		loops:         ev.compiler.loops[:0],
		loopIndex:     -1,
		tryCatchIndex: -1,
		iotaVal:       -1,
		indent:        so.indent,
		trace:         so.trace,
	}

	if ev.compiler.opts == nil {
		ev.compiler.opts = new(CompilerOptions)
	}

	*ev.compiler.opts = CompilerOptions{
		Trace:         so.trace,
		TraceCompiler: so.traceCompiler,
		TraceParser:   so.traceParser,
		NoOptimize:    true,
	}

	if ev.compiler.constsCache == nil {
		ev.compiler.constsCache = make(map[Object]int)
	} else {
		for k := range ev.compiler.constsCache {
			delete(ev.compiler.constsCache, k)
		}
	}

	if ev.compiler.cfuncCache == nil {
		ev.compiler.cfuncCache = make(map[uint32][]int)
	} else {
		for k := range ev.compiler.cfuncCache {
			delete(ev.compiler.cfuncCache, k)
		}
	}

	if ev.compiler.sourceMap == nil {
		ev.compiler.sourceMap = make(map[int]int)
	} else {
		for k := range ev.compiler.sourceMap {
			delete(ev.compiler.sourceMap, k)
		}
	}

	if ev.compiler.moduleStore == nil {
		ev.compiler.moduleStore = new(moduleStore)
	} else {
		ev.compiler.moduleStore.reset()
	}
}

func (ev *optimizerEval) eval(so *SimpleOptimizer, expr parser.Expr) (Object, bool) {
	ev.resetCompiler(so)

	ev.returnStmt.Result = expr

	istrace := so.trace != nil
	if err := ev.compiler.Compile(&ev.returnStmt); err != nil {
		if istrace {
			so.printTraceMsgf("compile error: %s", err)
		}
		return nil, false
	}

	bytecode := ev.compiler.Bytecode()

	if !canOptimizeInsts(bytecode.Constants, bytecode.Main.Instructions) {
		if istrace {
			so.printTraceMsgf("cannot optimize instructions")
		}
		return nil, false
	}
	if ev.vm == nil {
		ev.vm = NewVM(nil).SetRecover(true)
	}
	defer ev.vm.Abort()

	if ev.global == nil {
		ev.global = Map{}
	} else {
		for k := range ev.global {
			delete(ev.global, k)
		}
	}

	var startTime time.Time
	if istrace {
		so.printTraceMsgf("run vm")
		startTime = time.Now()
	}

	obj, err := ev.vm.SetBytecode(bytecode).Run(ev.global)

	if istrace {
		so.printTraceMsgf("vm run time: %s", time.Since(startTime))
	}

	if err != nil {
		if istrace {
			so.printTraceMsgf("eval error: %s", err)
		}
		if !errors.Is(err, ErrVMAborted) {
			so.errors = append(so.errors, so.error(expr, err))
		}
		return nil, false
	}
	return obj, true
}
