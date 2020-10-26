// A modified version of Tengo SymbolTable.

// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

// Copyright (c) 2019 Daniel Kang.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE.tengo file.

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
	Original *Symbol
}

func (s *Symbol) String() string {
	return fmt.Sprintf("Symbol{Name:%s Index:%d Scope:%s Assigned:%v Original:%s}",
		s.Name, s.Index, s.Scope, s.Assigned, s.Original)
}

// SymbolTable represents a symbol table.
type SymbolTable struct {
	store            map[string]*Symbol
	skipIndex        map[int]struct{}
	frees            []*Symbol
	parent           *SymbolTable
	numParams        int
	maxDefinition    int
	numDefinition    int
	block            bool
	disableParams    bool
	disabledBuiltins map[string]struct{}
}

// NewSymbolTable creates new symbol table object.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store:     make(map[string]*Symbol),
		skipIndex: make(map[int]struct{}),
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
		symbol := &Symbol{
			Name:  param,
			Index: st.NextIndex(),
			Scope: ScopeLocal,
		}
		st.numDefinition++
		st.store[param] = symbol
		st.updateMaxDefs(symbol.Index + 1)
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
		if symbol.Scope == ScopeLocal {
			return symbol, true
		}
	}
	index := st.NextIndex()
	if _, ok := st.skipIndex[index]; ok {
		st.numDefinition++
		for {
			index = st.NextIndex()
			if _, ok := st.skipIndex[index]; ok {
				st.numDefinition++
			} else {
				break
			}
		}
	}
	symbol = &Symbol{
		Name:  name,
		Index: index,
		Scope: ScopeLocal,
	}
	st.numDefinition++
	st.store[name] = symbol
	st.updateMaxDefs(symbol.Index + 1)
	return symbol, false
}

func (st *SymbolTable) defineFree(original *Symbol) *Symbol {
	// no duplicates symbol exists in "frees" because it is stored in map
	// and next Resolve call returns existing symbol
	st.frees = append(st.frees, original)
	symbol := &Symbol{
		Name:     original.Name,
		Index:    len(st.frees) - 1,
		Scope:    ScopeFree,
		Original: original,
	}
	st.store[original.Name] = symbol
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
			return nil, fmt.Errorf("symbol %q cannot be global, already defined", name)
		}
		return sym, nil
	}
	s := &Symbol{
		Name:  name,
		Index: -1,
		Scope: ScopeGlobal,
	}
	st.store[name] = s
	return s, nil
}

// IsGlobal returns true if given name is registered global name.
func (st *SymbolTable) IsGlobal(name string) bool {
	sym, ok := st.Resolve(name)
	return ok && sym.Scope == ScopeGlobal
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

// SkipIndex marks the symbol to be skipped in upper blocks to prevent overwriting.
func (st *SymbolTable) SkipIndex(idx int) {
	if st.block {
		st.parent.SkipIndex(idx)
	}
	st.skipIndex[idx] = struct{}{}
}

// IsIndexSkipped returns true if symbol index is marked as skipped.
func (st *SymbolTable) IsIndexSkipped(idx int) bool {
	_, ok := st.skipIndex[idx]
	return ok
}

// DisableBuiltin disables given builtin name(s).
// Compiler returns `Compile Error: unresolved reference "builtin name"`
// if a disabled builtin is used.
func (st *SymbolTable) DisableBuiltin(names ...string) *SymbolTable {
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
