// Copyright (c) 2020 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"errors"
	"sync"
)

// Importable interface represents importable module instance.
type Importable interface {
	// Import should return either an Object or module source code ([]byte).
	Import(moduleName string) (interface{}, error)
}

// ModuleMap represents a set of named modules. Use NewModuleMap to create a
// new module map.
type ModuleMap struct {
	mu sync.Mutex
	m  map[string]Importable
}

// NewModuleMap creates a new module map.
func NewModuleMap() *ModuleMap {
	return &ModuleMap{m: make(map[string]Importable)}
}

// Add adds an importable module.
func (m *ModuleMap) Add(name string, module Importable) *ModuleMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[name] = module
	return m
}

// AddBuiltinModule adds a builtin module.
func (m *ModuleMap) AddBuiltinModule(
	name string,
	attrs map[string]Object,
) *ModuleMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[name] = &BuiltinModule{Attrs: attrs}
	return m
}

// AddSourceModule adds a source module.
func (m *ModuleMap) AddSourceModule(name string, src []byte) *ModuleMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[name] = &SourceModule{Src: src}
	return m
}

// Remove removes a named module.
func (m *ModuleMap) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, name)
}

// Get returns an import module identified by name.
// It returns nil if the name is not found.
func (m *ModuleMap) Get(name string) Importable {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	return m.m[name]
}

// Range calls given function for each module.
func (m *ModuleMap) Range(fn func(name string, mod Importable) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, mod := range m.m {
		if !fn(name, mod) {
			break
		}
	}
}

// Copy creates a copy of the module map.
func (m *ModuleMap) Copy() *ModuleMap {
	m.mu.Lock()
	defer m.mu.Unlock()

	c := &ModuleMap{m: make(map[string]Importable)}

	for name, mod := range m.m {
		c.m[name] = mod
	}
	return c
}

// Len returns the number of modules.
func (m *ModuleMap) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.m)
}

// Merge merges modules from other ModuleMap.
func (m *ModuleMap) Merge(other *ModuleMap) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, mod := range other.m {
		m.m[name] = mod
	}
}

// SourceModule is an importable module that's written in Tengo.
type SourceModule struct {
	Src []byte
}

// Import returns a module source code.
func (m *SourceModule) Import(_ string) (interface{}, error) {
	return m.Src, nil
}

// BuiltinModule is an importable module that's written in Go.
type BuiltinModule struct {
	Attrs map[string]Object
}

// Import returns an immutable map for the module.
func (m *BuiltinModule) Import(moduleName string) (interface{}, error) {
	if m.Attrs == nil {
		return nil, errors.New("module attributes not set")
	}

	cp := Map(m.Attrs).Copy()
	cp.(Map)["__module_name__"] = String(moduleName)
	return cp, nil
}
