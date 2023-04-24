// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"errors"
	"fmt"
	"sort"
)

// SymbolScope represents a symbol scope.
type SymbolScope string

// List of symbol scopes
const (
	ScopeGlobal  SymbolScope = "GLOBAL"
	ScopeLocal   SymbolScope = "LOCAL"
	ScopeBuiltin SymbolScope = "BUILTIN"
	ScopeFree    SymbolScope = "FREE"
)

// Symbol represents a symbol in the symbol table.
type Symbol struct {
	Name     string
	Index    int
	Scope    SymbolScope
	Assigned bool
	Constant bool
	Original *Symbol
}

func (s *Symbol) String() string {
	return fmt.Sprintf("Symbol{Name:%s Index:%d Scope:%s Assigned:%v "+
		"Original:%s Constant:%t}",
		s.Name, s.Index, s.Scope, s.Assigned, s.Original, s.Constant)
}

// SymbolTable represents a symbol table.
type SymbolTable struct {
	parent           *SymbolTable
	maxDefinition    int
	numDefinition    int
	numParams        int
	store            map[string]*Symbol
	disabledBuiltins map[string]struct{}
	frees            []*Symbol
	block            bool
	disableParams    bool
	shadowedBuiltins []string
}

// NewSymbolTable creates new symbol table object.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]*Symbol),
	}
}

// Fork creates a new symbol table for a new scope.
func (st *SymbolTable) Fork(block bool) *SymbolTable {
	fork := NewSymbolTable()
	fork.parent = st
	fork.block = block
	fork.disableParams = st.disableParams
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
			Index: st.NextIndex(),
			Scope: ScopeLocal,
		}
		st.numDefinition++
		st.store[param] = symbol
		st.updateMaxDefs(symbol.Index + 1)
		st.shadowBuiltin(param)
	}
	return nil
}

func (st *SymbolTable) find(name string, scopes ...SymbolScope) (*Symbol, bool) {
	if symbol, ok := st.store[name]; ok {
		if len(scopes) == 0 {
			return symbol, ok
		}

		for _, s := range scopes {
			if s == symbol.Scope {
				return symbol, true
			}
		}
	}
	return nil, false
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
			symbol.Scope != ScopeBuiltin {
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

	index := st.NextIndex()

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

// NextIndex returns the next symbol index.
func (st *SymbolTable) NextIndex() int {
	if st.block {
		return st.parent.NextIndex() + st.numDefinition
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

// Symbols returns registered symbols for the scope.
func (st *SymbolTable) Symbols() []*Symbol {
	out := make([]*Symbol, 0, len(st.store))
	for _, s := range st.store {
		out = append(out, s)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Index < out[j].Index
	})

	return out
}

// DisableBuiltin disables given builtin name(s).
// Compiler returns `Compile Error: unresolved reference "builtin name"`
// if a disabled builtin is used.
func (st *SymbolTable) DisableBuiltin(names ...string) *SymbolTable {
	if len(names) == 0 {
		return st
	}

	if st.parent != nil {
		return st.parent.DisableBuiltin(names...)
	}

	if st.disabledBuiltins == nil {
		st.disabledBuiltins = make(map[string]struct{})
	}

	for _, n := range names {
		st.disabledBuiltins[n] = struct{}{}
	}
	return st
}

// DisabledBuiltins returns disabled builtin names.
func (st *SymbolTable) DisabledBuiltins() []string {
	if st.parent != nil {
		return st.parent.DisabledBuiltins()
	}

	if st.disabledBuiltins == nil {
		return nil
	}

	names := make([]string, 0, len(st.disabledBuiltins))
	for n := range st.disabledBuiltins {
		names = append(names, n)
	}
	return names
}

// isBuiltinDisabled returns true if builtin name marked as disabled.
func (st *SymbolTable) isBuiltinDisabled(name string) bool {
	if st.parent != nil {
		return st.parent.isBuiltinDisabled(name)
	}

	_, ok := st.disabledBuiltins[name]
	return ok
}

// ShadowedBuiltins returns all shadowed builtin names including parent symbol
// tables'. Returing slice may contain duplicate names.
func (st *SymbolTable) ShadowedBuiltins() []string {
	var out []string
	if len(st.shadowedBuiltins) > 0 {
		out = append(out, st.shadowedBuiltins...)
	}

	if st.parent != nil {
		out = append(out, st.parent.ShadowedBuiltins()...)
	}
	return out
}

func (st *SymbolTable) shadowBuiltin(name string) {
	if _, ok := BuiltinsMap[name]; ok {
		st.shadowedBuiltins = append(st.shadowedBuiltins, name)
	}
}
