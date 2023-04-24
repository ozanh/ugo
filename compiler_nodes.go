// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"github.com/ozanh/ugo/parser"
	"github.com/ozanh/ugo/token"
)

func (c *Compiler) compileIfStmt(node *parser.IfStmt) error {
	// open new symbol table for the statement
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		c.symbolTable = c.symbolTable.Parent(false)
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
	return nil
}

func (c *Compiler) compileTryStmt(node *parser.TryStmt) error {
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
	// fork new symbol table for the statement
	c.symbolTable = c.symbolTable.Fork(true)
	c.tryCatchIndex++
	defer func() {
		c.symbolTable = c.symbolTable.Parent(false)
		c.emit(node, OpThrow, 0) // implicit re-throw
	}()

	optry := c.emit(node, OpSetupTry, 0, 0)
	var catchPos, finallyPos int
	if node.Body != nil && len(node.Body.Stmts) > 0 {
		// in order not to fork symbol table in Body, compile stmts here instead of in *BlockStmt
		for _, stmt := range node.Body.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}
	}

	var opjump int
	if node.Catch != nil {
		// if there is no thrown error before catch statement, set catch ident to undefined
		// otherwise jumping to finally and accessing ident in finally access previous set same index variable.
		if node.Catch.Ident != nil {
			c.emit(node.Catch, OpNull)
			symbol, exists := c.symbolTable.DefineLocal(node.Catch.Ident.Name)
			if exists {
				c.emit(node, OpSetLocal, symbol.Index)
			} else {
				c.emit(node, OpDefineLocal, symbol.Index)
			}
		}

		opjump = c.emit(node, OpJump, 0)
		catchPos = len(c.instructions)
		if err := c.Compile(node.Catch); err != nil {
			return err
		}
	}

	c.tryCatchIndex--
	// always emit OpSetupFinally to cleanup
	if node.Finally != nil {
		finallyPos = c.emit(node.Finally, OpSetupFinally)
		if err := c.Compile(node.Finally); err != nil {
			return err
		}
	} else {
		finallyPos = c.emit(node, OpSetupFinally)
	}

	c.changeOperand(optry, catchPos, finallyPos)
	if node.Catch != nil {
		// no need jumping if catch is not defined
		c.changeOperand(opjump, finallyPos)
	}
	return nil
}

func (c *Compiler) compileCatchStmt(node *parser.CatchStmt) error {
	c.emit(node, OpSetupCatch)
	if node.Ident != nil {
		symbol, exists := c.symbolTable.DefineLocal(node.Ident.Name)
		if exists {
			c.emit(node, OpSetLocal, symbol.Index)
		} else {
			c.emit(node, OpDefineLocal, symbol.Index)
		}
	} else {
		c.emit(node, OpPop)
	}

	if node.Body == nil {
		return nil
	}

	// in order not to fork symbol table in Body, compile stmts here instead of in *BlockStmt
	for _, stmt := range node.Body.Stmts {
		if err := c.Compile(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileFinallyStmt(node *parser.FinallyStmt) error {
	if node.Body == nil {
		return nil
	}

	// in order not to fork symbol table in Body, compile stmts here instead of in *BlockStmt
	for _, stmt := range node.Body.Stmts {
		if err := c.Compile(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileThrowStmt(node *parser.ThrowStmt) error {
	if node.Expr != nil {
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
	}
	c.emit(node, OpThrow, 1)
	return nil
}

func (c *Compiler) compileDeclStmt(node *parser.DeclStmt) error {
	decl := node.Decl.(*parser.GenDecl)
	if len(decl.Specs) == 0 {
		return c.errorf(node, "empty declaration not allowed")
	}

	switch decl.Tok {
	case token.Param:
		return c.compileDeclParam(decl)
	case token.Global:
		return c.compileDeclGlobal(decl)
	case token.Var, token.Const:
		return c.compileDeclValue(decl)
	}
	return nil
}

func (c *Compiler) compileDeclParam(node *parser.GenDecl) error {
	if c.symbolTable.parent != nil {
		return c.errorf(node, "param not allowed in this scope")
	}

	names := make([]string, 0, len(node.Specs))
	for _, sp := range node.Specs {
		spec := sp.(*parser.ParamSpec)
		names = append(names, spec.Ident.Name)
		if spec.Variadic {
			if c.variadic {
				return c.errorf(node,
					"multiple variadic param declaration")
			}
			c.variadic = true
		}
	}

	if err := c.symbolTable.SetParams(names...); err != nil {
		return c.error(node, err)
	}
	return nil
}

func (c *Compiler) compileDeclGlobal(node *parser.GenDecl) error {
	if c.symbolTable.parent != nil {
		return c.errorf(node, "global not allowed in this scope")
	}

	for _, sp := range node.Specs {
		spec := sp.(*parser.ParamSpec)
		symbol, err := c.symbolTable.DefineGlobal(spec.Ident.Name)
		if err != nil {
			return c.error(node, err)
		}

		idx := c.addConstant(String(spec.Ident.Name))
		symbol.Index = idx
	}
	return nil
}

func (c *Compiler) compileDeclValue(node *parser.GenDecl) error {
	var (
		isConst  bool
		lastExpr parser.Expr
	)
	if node.Tok == token.Const {
		isConst = true
		defer func() { c.iotaVal = -1 }()
	}

	for _, sp := range node.Specs {
		spec := sp.(*parser.ValueSpec)
		if isConst {
			if v, ok := spec.Data.(int); ok {
				c.iotaVal = v
			} else {
				return c.errorf(node, "invalid iota value")
			}
		}
		for i, ident := range spec.Idents {
			leftExpr := []parser.Expr{ident}
			var v parser.Expr
			if i < len(spec.Values) {
				v = spec.Values[i]
			}

			if v == nil {
				if isConst && lastExpr != nil {
					v = lastExpr
				} else {
					v = &parser.UndefinedLit{TokenPos: ident.Pos()}
				}
			} else {
				lastExpr = v
			}

			rightExpr := []parser.Expr{v}
			err := c.compileAssignStmt(node, leftExpr, rightExpr, node.Tok, token.Define)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Compiler) checkAssignment(
	node parser.Node,
	lhs []parser.Expr,
	rhs []parser.Expr,
	keyword token.Token,
	op token.Token,
) (bool, error) {
	_, numRHS := len(lhs), len(rhs)
	if numRHS > 1 {
		return false, c.errorf(node,
			"multiple expressions on the right side not supported")
	}

	var selector bool
Loop:
	for _, expr := range lhs {
		switch expr.(type) {
		case *parser.SelectorExpr, *parser.IndexExpr:
			selector = true
			break Loop
		}
	}

	if selector {
		if op == token.Define {
			// using selector on new variable does not make sense
			return false, c.errorf(node, "operator ':=' not allowed with selector")
		}
	}

	return true, nil
}

func (c *Compiler) compileAssignStmt(
	node parser.Node,
	lhs []parser.Expr,
	rhs []parser.Expr,
	keyword token.Token,
	op token.Token,
) error {
	compile, err := c.checkAssignment(node, lhs, rhs, keyword, op)
	if err != nil || !compile {
		return err
	}

	var isArrDestruct bool
	var tempArrSymbol *Symbol
	// +=, -=, *=, /=
	if op != token.Assign && op != token.Define {
		if err := c.Compile(lhs[0]); err != nil {
			return err
		}
	} else if len(lhs) > 1 {
		isArrDestruct = true
		// ignore redefinition of :array symbol, it can be used multiple times
		// within a block.
		tempArrSymbol, _ = c.symbolTable.DefineLocal(":array")
		// ignore disabled builtins of symbol table for BuiltinMakeArray because
		// it is required to handle destructuring assignment.
		c.emit(node, OpGetBuiltin, int(BuiltinMakeArray))
		c.emit(node, OpConstant, c.addConstant(Int(len(lhs))))
	}

	// compile RHSs
	for _, expr := range rhs {
		if err := c.Compile(expr); err != nil {
			return err
		}
	}

	if isArrDestruct {
		return c.compileDestructuring(node, lhs, tempArrSymbol, keyword, op)
	}

	if op != token.Assign && op != token.Define {
		c.compileCompoundAssignment(node, op)
	}
	return c.compileDefineAssign(node, lhs[0], keyword, op, false)
}

func (c *Compiler) compileCompoundAssignment(
	node parser.Node,
	op token.Token,
) {
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
}

func (c *Compiler) compileDestructuring(
	node parser.Node,
	lhs []parser.Expr,
	tempArrSymbol *Symbol,
	keyword token.Token,
	op token.Token,
) error {
	c.emit(node, OpCall, 2, 0)
	c.emit(node, OpDefineLocal, tempArrSymbol.Index)
	numLHS := len(lhs)
	var found int

	for lhsIndex, expr := range lhs {
		if op == token.Define {
			if term, ok := expr.(*parser.Ident); ok {
				if _, ok = c.symbolTable.find(term.Name); ok {
					found++
				}
			}
			if found == numLHS {
				return c.errorf(node, "no new variable on left side")
			}
		}

		c.emit(node, OpGetLocal, tempArrSymbol.Index)
		c.emit(node, OpConstant, c.addConstant(Int(lhsIndex)))
		c.emit(node, OpGetIndex, 1)
		err := c.compileDefineAssign(node, expr, keyword, op, keyword != token.Const)
		if err != nil {
			return err
		}
	}

	if !c.symbolTable.InBlock() {
		// blocks set undefined to variables defined in it after block
		c.emit(node, OpNull)
		c.emit(node, OpSetLocal, tempArrSymbol.Index)
	}
	return nil
}

func (c *Compiler) compileDefine(
	node parser.Node,
	ident string,
	allowRedefine bool,
	keyword token.Token,
) error {
	symbol, exists := c.symbolTable.DefineLocal(ident)
	if !allowRedefine && exists && ident != "_" {
		return c.errorf(node, "%q redeclared in this block", ident)
	}

	if symbol.Constant {
		return c.errorf(node, "assignment to constant variable %q", ident)
	}
	if c.iotaVal > -1 && ident == "iota" && keyword == token.Const {
		return c.errorf(node, "assignment to iota")
	}

	c.emit(node, OpDefineLocal, symbol.Index)
	symbol.Assigned = true
	symbol.Constant = keyword == token.Const && ident != "_"
	return nil
}

func (c *Compiler) compileAssign(
	node parser.Node,
	symbol *Symbol,
	ident string,
) error {
	if symbol.Constant {
		return c.errorf(node, "assignment to constant variable %q", ident)
	}

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

func (c *Compiler) compileDefineAssign(
	node parser.Node,
	lhs parser.Expr,
	keyword token.Token,
	op token.Token,
	allowRedefine bool,
) error {
	ident, selectors := resolveAssignLHS(lhs)
	numSel := len(selectors)
	if numSel == 0 && op == token.Define {
		return c.compileDefine(node, ident, allowRedefine, keyword)
	}

	symbol, ok := c.symbolTable.Resolve(ident)
	if !ok {
		return c.errorf(node, "unresolved reference %q", ident)
	}

	if numSel == 0 {
		return c.compileAssign(node, symbol, ident)
	}

	// get indexes until last one and set the value to the last index
	switch symbol.Scope {
	case ScopeLocal:
		c.emit(node, OpGetLocal, symbol.Index)
	case ScopeFree:
		c.emit(node, OpGetFree, symbol.Index)
	case ScopeGlobal:
		c.emit(node, OpGetGlobal, symbol.Index)
	default:
		return c.errorf(node, "unexpected scope %q for symbol %q",
			symbol.Scope, ident)
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
	case *parser.IndexExpr:
		name, selectors = resolveAssignLHS(term.Expr)
		selectors = append(selectors, term.Index)
	case *parser.Ident:
		name = term.Name
	}
	return
}

func (c *Compiler) compileBranchStmt(node *parser.BranchStmt) error {
	switch node.Token {
	case token.Break:
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
		curLoop.breaks = append(curLoop.breaks, pos)
	case token.Continue:
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
		curLoop.continues = append(curLoop.continues, pos)
	default:
		return c.errorf(node, "invalid branch statement: %s", node.Token.String())
	}
	return nil
}

func (c *Compiler) compileBlockStmt(node *parser.BlockStmt) error {
	if len(node.Stmts) == 0 {
		return nil
	}

	c.symbolTable = c.symbolTable.Fork(true)
	for _, stmt := range node.Stmts {
		if err := c.Compile(stmt); err != nil {
			return err
		}
	}

	c.symbolTable = c.symbolTable.Parent(false)
	return nil
}

func (c *Compiler) compileReturnStmt(node *parser.ReturnStmt) error {
	if node.Result == nil {
		if c.tryCatchIndex > -1 {
			c.emit(node, OpFinalizer, 0)
		}
		c.emit(node, OpReturn, 0)
		return nil
	}

	if err := c.Compile(node.Result); err != nil {
		return err
	}

	if c.tryCatchIndex > -1 {
		c.emit(node, OpFinalizer, 0)
	}

	c.emit(node, OpReturn, 1)
	return nil
}

func (c *Compiler) compileForStmt(stmt *parser.ForStmt) error {
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		c.symbolTable = c.symbolTable.Parent(false)
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
	for _, pos := range loop.breaks {
		c.changeOperand(pos, postStmtPos)
	}

	for _, pos := range loop.continues {
		c.changeOperand(pos, postBodyPos)
	}
	return nil
}

func (c *Compiler) compileForInStmt(stmt *parser.ForInStmt) error {
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		c.symbolTable = c.symbolTable.Parent(false)
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
	c.emit(stmt, OpDefineLocal, itSymbol.Index)

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
		c.emit(stmt, OpDefineLocal, keySymbol.Index)
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
		c.emit(stmt, OpDefineLocal, valueSymbol.Index)
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
	for _, pos := range loop.breaks {
		c.changeOperand(pos, postStmtPos)
	}

	for _, pos := range loop.continues {
		c.changeOperand(pos, postBodyPos)
	}
	return nil
}

func (c *Compiler) compileFuncLit(node *parser.FuncLit) error {
	params := make([]string, len(node.Type.Params.List))
	for i, ident := range node.Type.Params.List {
		params[i] = ident.Name
	}

	symbolTable := c.symbolTable.Fork(false)
	if err := symbolTable.SetParams(params...); err != nil {
		return c.error(node, err)
	}

	fork := c.fork(c.file, c.modulePath, c.moduleMap, symbolTable)
	fork.variadic = node.Type.Params.VarArgs
	if err := fork.Compile(node.Body); err != nil {
		return err
	}

	freeSymbols := fork.symbolTable.FreeSymbols()
	for _, s := range freeSymbols {
		switch s.Scope {
		case ScopeLocal:
			c.emit(node, OpGetLocalPtr, s.Index)
		case ScopeFree:
			c.emit(node, OpGetFreePtr, s.Index)
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
	return nil
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

func (c *Compiler) compileBinaryExpr(node *parser.BinaryExpr) error {
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
	default:
		if !node.Token.IsBinaryOperator() {
			return c.errorf(node, "invalid binary operator: %s",
				node.Token.String())
		}
		c.emit(node, OpBinaryOp, int(node.Token))
	}
	return nil
}

func (c *Compiler) compileUnaryExpr(node *parser.UnaryExpr) error {
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
	return nil
}

func (c *Compiler) compileSelectorExpr(node *parser.SelectorExpr) error {
	expr, selectors := resolveSelectorExprs(node)
	if err := c.Compile(expr); err != nil {
		return err
	}
	for _, selector := range selectors {
		if err := c.Compile(selector); err != nil {
			return err
		}
	}
	c.emit(node, OpGetIndex, len(selectors))
	return nil
}

func resolveSelectorExprs(node parser.Expr) (expr parser.Expr, selectors []parser.Expr) {
	expr = node
	if v, ok := node.(*parser.SelectorExpr); ok {
		expr, selectors = resolveIndexExprs(v.Expr)
		selectors = append(selectors, v.Sel)
	}
	return
}

func (c *Compiler) compileIndexExpr(node *parser.IndexExpr) error {
	expr, indexes := resolveIndexExprs(node)
	if err := c.Compile(expr); err != nil {
		return err
	}
	for _, index := range indexes {
		if err := c.Compile(index); err != nil {
			return err
		}
	}
	c.emit(node, OpGetIndex, len(indexes))
	return nil
}

func resolveIndexExprs(node parser.Expr) (expr parser.Expr, indexes []parser.Expr) {
	expr = node
	if v, ok := node.(*parser.IndexExpr); ok {
		expr, indexes = resolveIndexExprs(v.Expr)
		indexes = append(indexes, v.Index)
	}
	return
}

func (c *Compiler) compileSliceExpr(node *parser.SliceExpr) error {
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
	return nil
}

func (c *Compiler) compileCallExpr(node *parser.CallExpr) error {
	var op = OpCall
	var selExpr *parser.SelectorExpr
	var isSelector bool
	if node.Func != nil {
		selExpr, isSelector = node.Func.(*parser.SelectorExpr)
	}

	if isSelector {
		if err := c.Compile(selExpr.Expr); err != nil {
			return err
		}
		op = OpCallName
	} else {
		if err := c.Compile(node.Func); err != nil {
			return err
		}
	}

	for _, arg := range node.Args {
		if err := c.Compile(arg); err != nil {
			return err
		}
	}

	if isSelector {
		if err := c.Compile(selExpr.Sel); err != nil {
			return err
		}
	}

	var expand int
	if node.Ellipsis.IsValid() {
		expand = 1
	}

	c.emit(node, op, len(node.Args), expand)
	return nil
}

func (c *Compiler) compileImportExpr(node *parser.ImportExpr) error {
	moduleName := node.ModuleName
	if moduleName == "" {
		return c.errorf(node, "empty module name")
	}

	importer := c.moduleMap.Get(moduleName)
	if importer == nil {
		return c.errorf(node, "module '%s' not found", moduleName)
	}

	extImp, isExt := importer.(ExtImporter)
	if isExt {
		if name := extImp.Name(); name != "" {
			moduleName = name
		}
	}

	module, exists := c.getModule(moduleName)
	if !exists {
		mod, err := importer.Import(moduleName)
		if err != nil {
			return c.error(node, err)
		}
		switch v := mod.(type) {
		case []byte:
			var moduleMap *ModuleMap
			if isExt {
				moduleMap = c.moduleMap.Fork(moduleName)
			} else {
				moduleMap = c.baseModuleMap()
			}
			cidx, err := c.compileModule(node, moduleName, moduleMap, v)
			if err != nil {
				return err
			}
			module = c.addModule(moduleName, 1, cidx)
		case Object:
			module = c.addModule(moduleName, 2, c.addConstant(v))
		default:
			return c.errorf(node, "invalid import value type: %T", v)
		}
	}

	switch module.typ {
	case 1:
		var numParams int
		mod := c.constants[module.constantIndex]
		if cf, ok := mod.(*CompiledFunction); ok {
			numParams = cf.NumParams
			if cf.Variadic {
				numParams--
			}
		}
		// load module
		// if module is already stored, load from VM.modulesCache otherwise call compiled function
		// and store copy of result to VM.modulesCache.
		c.emit(node, OpLoadModule, module.constantIndex, module.moduleIndex)
		jumpPos := c.emit(node, OpJumpFalsy, 0)
		// modules should not accept parameters, to suppress the wrong number of arguments error
		// set all params to undefined
		for i := 0; i < numParams; i++ {
			c.emit(node, OpNull)
		}
		c.emit(node, OpCall, numParams, 0)
		c.emit(node, OpStoreModule, module.moduleIndex)
		c.changeOperand(jumpPos, len(c.instructions))
	case 2:
		// load module
		// if module is already stored, load from VM.modulesCache otherwise copy object
		// and store it to VM.modulesCache.
		c.emit(node, OpLoadModule, module.constantIndex, module.moduleIndex)
		jumpPos := c.emit(node, OpJumpFalsy, 0)
		c.emit(node, OpStoreModule, module.moduleIndex)
		c.changeOperand(jumpPos, len(c.instructions))
	default:
		return c.errorf(node, "invalid module type: %v", module.typ)
	}
	return nil
}

func (c *Compiler) compileCondExpr(node *parser.CondExpr) error {
	if v, ok := node.Cond.(*parser.BoolLit); ok {
		if v.Value {
			return c.Compile(node.True)
		}
		return c.Compile(node.False)
	}

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
	return nil
}

func (c *Compiler) compileIdent(node *parser.Ident) error {
	symbol, ok := c.symbolTable.Resolve(node.Name)
	if !ok {
		if c.iotaVal < 0 || node.Name != "iota" {
			return c.errorf(node, "unresolved reference %q", node.Name)
		}
		c.emit(node, OpConstant, c.addConstant(Int(c.iotaVal)))
		return nil
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
	return nil
}

func (c *Compiler) compileArrayLit(node *parser.ArrayLit) error {
	for _, elem := range node.Elements {
		if err := c.Compile(elem); err != nil {
			return err
		}
	}

	c.emit(node, OpArray, len(node.Elements))
	return nil
}

func (c *Compiler) compileMapLit(node *parser.MapLit) error {
	for _, elt := range node.Elements {
		// key
		c.emit(node, OpConstant, c.addConstant(String(elt.Key)))
		// value
		if err := c.Compile(elt.Value); err != nil {
			return err
		}
	}

	c.emit(node, OpMap, len(node.Elements)*2)
	return nil
}
