// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"errors"
	"fmt"
)

// SymbolScope represents a symbol scope.
type SymbolScope int

// List of symbol scopes
const (
	ScopeGlobal SymbolScope = iota + 1
	ScopeLocal
	ScopeBuiltin
	ScopeFree
	ScopeConstLit
)

func (s SymbolScope) String() string {
	switch s {
	case ScopeGlobal:
		return "GLOBAL"
	case ScopeLocal:
		return "LOCAL"
	case ScopeBuiltin:
		return "BUILTIN"
	case ScopeFree:
		return "FREE"
	case ScopeConstLit:
		return "CONSTLIT"
	default:
		panic("unknown symbol scope")
	}
}

// Symbol represents a symbol in the symbol table.
type Symbol struct {
	Name     string
	Index    int
	Scope    SymbolScope
	Assigned bool
	Constant bool
	Original *Symbol
	constLit constLiteral
}

func (s *Symbol) String() string {
	return fmt.Sprintf("Symbol{Name:%s Index:%d Scope:%s Assigned:%v Constant:%t Original:%s}",
		s.Name, s.Index, s.Scope, s.Assigned, s.Constant, s.Original)
}

// SymbolTable represents a symbol table.
type SymbolTable struct {
	parent            *SymbolTable
	maxDefinition     int
	numDefinition     int
	numParams         int
	store             map[string]*Symbol
	disabledBuiltins  map[string]struct{}
	frees             []*Symbol
	shadowedBuiltins  []string
	block             bool
	disableParams     bool
	hasConstLit       bool
	hasParentConstLit bool
}

// NewSymbolTable creates new symbol table object.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]*Symbol),
	}
}

func (st *SymbolTable) reset() {
	*st = SymbolTable{
		store:            st.store,
		disabledBuiltins: st.disabledBuiltins,
		frees:            st.frees[:0],
		shadowedBuiltins: st.shadowedBuiltins[:0],
	}

	for k := range st.store {
		delete(st.store, k)
	}

	for k := range st.disabledBuiltins {
		delete(st.disabledBuiltins, k)
	}

}

// Fork creates a new symbol table for a new scope.
func (st *SymbolTable) Fork(block bool) *SymbolTable {
	fork := NewSymbolTable()
	fork.parent = st
	fork.block = block
	fork.disableParams = st.disableParams
	fork.hasParentConstLit = st.hasConstLit || st.hasParentConstLit
	return fork
}

// Parent returns the outer scope of the current symbol table.
func (st *SymbolTable) Parent(skipBlock bool) *SymbolTable {
	if skipBlock && st.block {
		return st.parent.Parent(skipBlock)
	}
	return st.parent
}

// EnableParams enables or disables definition of parameters.
func (st *SymbolTable) EnableParams(v bool) *SymbolTable {
	st.disableParams = !v
	return st
}

// InBlock returns true if symbol table belongs to a block.
func (st *SymbolTable) InBlock() bool {
	return st.block
}

// SetParams sets parameters defined in the scope. This can be called only once.
func (st *SymbolTable) SetParams(params ...string) error {
	if len(params) == 0 {
		return nil
	}

	if st.numParams > 0 {
		return errors.New("parameters already defined")
	}

	if st.disableParams {
		return errors.New("parameters disabled")
	}

	st.numParams = len(params)
	for _, param := range params {
		if _, ok := st.store[param]; ok {
			return fmt.Errorf("%q redeclared in this block", param)
		}
		symbol := &Symbol{
			Name:  param,
			Index: st.nextIndex(),
			Scope: ScopeLocal,
		}
		st.numDefinition++
		st.store[param] = symbol
		st.updateMaxDefs(symbol.Index + 1)
		st.shadowBuiltin(param)
	}
	return nil
}

// Resolve resolves a symbol with a given name.
func (st *SymbolTable) Resolve(name string) (symbol *Symbol, ok bool) {
	symbol, ok = st.store[name]
	if !ok && st.parent != nil {
		symbol, ok = st.parent.Resolve(name)
		if !ok {
			return
		}

		if !st.block &&
			symbol.Scope != ScopeGlobal &&
			symbol.Scope != ScopeBuiltin &&
			symbol.Scope != ScopeConstLit {
			return st.defineFree(symbol), true
		}
	}

	if !ok && st.parent == nil && !st.isBuiltinDisabled(name) {
		if idx, exists := BuiltinsMap[name]; exists {
			symbol = &Symbol{
				Name:  name,
				Index: int(idx),
				Scope: ScopeBuiltin,
			}
			st.store[name] = symbol
			return symbol, true
		}
	}
	return
}

// DefineLocal adds a new symbol with ScopeLocal in the current scope.
func (st *SymbolTable) DefineLocal(name string) (*Symbol, bool) {
	symbol, ok := st.store[name]
	if ok {
		return symbol, true
	}

	index := st.nextIndex()

	symbol = &Symbol{
		Name:  name,
		Index: index,
		Scope: ScopeLocal,
	}

	st.numDefinition++
	st.store[name] = symbol

	st.updateMaxDefs(symbol.Index + 1)
	st.shadowBuiltin(name)

	return symbol, false
}

func (st *SymbolTable) defineConstLit(name string) (*Symbol, bool) {
	symbol, ok := st.store[name]
	if ok {
		return symbol, true
	}
	st.hasConstLit = true
	symbol = &Symbol{
		Name:     name,
		Index:    -1,
		Scope:    ScopeConstLit,
		Constant: true,
	}

	st.store[name] = symbol
	st.shadowBuiltin(name)
	return symbol, false
}

func (st *SymbolTable) defineFree(original *Symbol) *Symbol {
	// no duplicate symbol exists in "frees" because it is stored in map
	// and next Resolve call returns existing symbol
	st.frees = append(st.frees, original)
	symbol := &Symbol{
		Name:     original.Name,
		Index:    len(st.frees) - 1,
		Scope:    ScopeFree,
		Constant: original.Constant,
		Original: original,
	}

	st.store[original.Name] = symbol
	st.shadowBuiltin(original.Name)
	return symbol
}

func (st *SymbolTable) updateMaxDefs(numDefs int) {
	if numDefs > st.maxDefinition {
		st.maxDefinition = numDefs
	}

	if st.block {
		st.parent.updateMaxDefs(numDefs)
	}
}

// nextIndex returns the next symbol index.
func (st *SymbolTable) nextIndex() int {
	if st.block {
		return st.parent.nextIndex() + st.numDefinition
	}
	return st.numDefinition
}

// DefineGlobal adds a new symbol with ScopeGlobal in the current scope.
func (st *SymbolTable) DefineGlobal(name string) (*Symbol, error) {
	if st.parent != nil {
		return nil, errors.New("global declaration can be at top scope")
	}

	sym, ok := st.store[name]
	if ok {
		if sym.Scope != ScopeGlobal {
			return nil, fmt.Errorf("%q redeclared in this block", name)
		}
		return sym, nil
	}

	s := &Symbol{
		Name:  name,
		Index: -1,
		Scope: ScopeGlobal,
	}

	st.store[name] = s
	st.shadowBuiltin(name)
	return s, nil
}

// MaxSymbols returns the total number of symbols defined in the scope.
func (st *SymbolTable) MaxSymbols() int {
	return st.maxDefinition
}

// NumParams returns number of parameters for the scope.
func (st *SymbolTable) NumParams() int {
	return st.numParams
}

// FreeSymbols returns registered free symbols for the scope.
func (st *SymbolTable) FreeSymbols() []*Symbol {
	return st.frees
}

func (st *SymbolTable) Range(visitParent bool, fn func(*Symbol) bool) {
	names := make(map[string]struct{})
	ptr := st
	for ptr != nil {
		for name, sym := range ptr.store {
			if _, ok := names[name]; !ok {
				names[name] = struct{}{}
				if !fn(sym) {
					return
				}
			}
		}
		if !visitParent {
			return
		}
		ptr = ptr.parent
	}
}

func (st *SymbolTable) Find(visitParent bool, fn func(*Symbol) bool) *Symbol {
	ptr := st
	for ptr != nil {
		for _, sym := range ptr.store {
			if fn(sym) {
				return sym
			}
		}
		if !visitParent {
			return nil
		}
		ptr = ptr.parent
	}
	return nil
}

// DisableBuiltin disables given builtin name(s).
// Compiler returns `Compile Error: unresolved reference "builtin name"`
// if a disabled builtin is used.
func (st *SymbolTable) DisableBuiltin(names ...string) {
	if len(names) == 0 {
		return
	}

	root := st.root()
	root.initDisabledBuiltinsMap()

	for _, n := range names {
		root.disabledBuiltins[n] = struct{}{}
	}
}

// DisabledBuiltins returns disabled builtin names.
func (st *SymbolTable) DisabledBuiltins() []string {
	root := st.root()
	if root.disabledBuiltins == nil {
		return nil
	}

	names := make([]string, 0, len(root.disabledBuiltins))
	for n := range root.disabledBuiltins {
		names = append(names, n)
	}
	return names
}

func (st *SymbolTable) initDisabledBuiltinsMap() {
	root := st.root()
	if root.disabledBuiltins == nil {
		root.disabledBuiltins = make(map[string]struct{})
	}
}

func (st *SymbolTable) root() *SymbolTable {
	if st.parent == nil {
		return st
	}
	return st.parent.root()
}

func (st *SymbolTable) disabledBuiltinsMap() map[string]struct{} {
	if st == nil {
		return nil
	}
	root := st.root()
	return root.disabledBuiltins
}

func (st *SymbolTable) hasAnyShadowedBuiltins() bool {
	if st == nil {
		return false
	}
	if len(st.shadowedBuiltins) > 0 {
		return true
	}
	if st.parent != nil {
		return st.parent.hasAnyShadowedBuiltins()
	}
	return false
}

// isBuiltinDisabled returns true if builtin name marked as disabled.
func (st *SymbolTable) isBuiltinDisabled(name string) bool {
	root := st.root()
	_, ok := root.disabledBuiltins[name]
	return ok
}

func (st *SymbolTable) shadowBuiltin(name string) {
	if _, ok := BuiltinsMap[name]; ok {
		st.shadowedBuiltins = append(st.shadowedBuiltins, name)
	}
}

// Helper functions for symbol table to use in Compiler and optimizer.

func findSymbolSelf(st *SymbolTable, name string) *Symbol {
	const visitParent = false
	return st.Find(visitParent, func(sym *Symbol) bool {
		return sym.Name == name
	})
}

func findSymbol(st *SymbolTable, name string, scope SymbolScope) *Symbol {
	if st == nil {
		return nil
	}

	if scope == ScopeConstLit {
		if !st.hasConstLit {
			if st.hasParentConstLit {
				return findSymbol(st.parent, name, scope)
			}
			return nil
		}
	}

	const visitParent = true
	return st.Find(visitParent, func(sym *Symbol) bool {
		return sym.Name == name && sym.Scope == scope
	})
}

func hasAnyConstLit(st *SymbolTable) bool {
	return st.hasConstLit || st.hasParentConstLit
}

func copyMapStringSet(m map[string]struct{}) map[string]struct{} {
	if m == nil {
		return nil
	}
	n := make(map[string]struct{}, len(m))
	for k, v := range m {
		n[k] = v
	}
	return n
}

func optimCopyBuiltinStates(dest, src *SymbolTable) {
	otherMap := src.disabledBuiltinsMap()
	hasShadowed := src.hasAnyShadowedBuiltins()
	if len(otherMap) == 0 && !hasShadowed {
		return
	}

	root := dest.root()
	root.initDisabledBuiltinsMap()

	for name := range otherMap {
		root.disabledBuiltins[name] = struct{}{}
	}

	ptr := src
	for ptr != nil {
		for _, name := range ptr.shadowedBuiltins {
			root.disabledBuiltins[name] = struct{}{}
		}
		ptr = ptr.parent
	}
}

func optimCopyBuiltinStatesFromScope(dest *SymbolTable, src *optimizerScope) {
	if src == nil {
		return
	}

	root := dest.root()
	ptr := src
	for {
		root.DisableBuiltin(ptr.shadowed...)
		ptr = ptr.parent
		if ptr == nil {
			break
		}
	}
}
