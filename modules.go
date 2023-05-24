// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"errors"
)

// Importable interface represents importable module instance.
type Importable interface {
	// Import should return either an Object or module source code ([]byte).
	Import(moduleName string) (interface{}, error)
}

// ExtImporter wraps methods for a module which will be impored dynamically like
// a file.
type ExtImporter interface {
	Importable
	// Get returns Extimporter instance which will import a module.
	Get(moduleName string) ExtImporter
	// Name returns the full name of the module e.g. absoule path of a file.
	// Import names are generally relative, this overwrites module name and used
	// as unique key for compiler module cache.
	Name() string
	// Fork returns an ExtImporter instance which will be used to import the
	// modules. Fork will get the result of Name() if it is not empty, otherwise
	// module name will be same with the Get call.
	Fork(moduleName string) ExtImporter
}

// ModuleMap represents a set of named modules. Use NewModuleMap to create a
// new module map.
type ModuleMap struct {
	m  map[string]Importable
	im ExtImporter
}

// NewModuleMap creates a new module map.
func NewModuleMap() *ModuleMap {
	return &ModuleMap{m: make(map[string]Importable)}
}

// SetExtImporter sets an ExtImporter to ModuleMap, which will be used to
// import modules dynamically.
func (m *ModuleMap) SetExtImporter(im ExtImporter) *ModuleMap {
	m.im = im
	return m
}

// Fork creates a new ModuleMap instance if ModuleMap has an ExtImporter to
// make ExtImporter preseve state.
func (m *ModuleMap) Fork(moduleName string) *ModuleMap {
	if m == nil {
		return nil
	}
	if m.im != nil {
		fork := m.im.Fork(moduleName)
		return &ModuleMap{m: m.m, im: fork}
	}
	return m
}

// Add adds an importable module.
func (m *ModuleMap) Add(name string, module Importable) *ModuleMap {
	m.m[name] = module
	return m
}

// AddBuiltinModule adds a builtin module.
func (m *ModuleMap) AddBuiltinModule(
	name string,
	attrs map[string]Object,
) *ModuleMap {
	m.m[name] = &BuiltinModule{Attrs: attrs}
	return m
}

// AddSourceModule adds a source module.
func (m *ModuleMap) AddSourceModule(name string, src []byte) *ModuleMap {
	m.m[name] = &SourceModule{Src: src}
	return m
}

// Remove removes a named module.
func (m *ModuleMap) Remove(name string) {
	delete(m.m, name)
}

// Get returns an import module identified by name.
// It returns nil if the name is not found.
func (m *ModuleMap) Get(name string) Importable {
	if m == nil {
		return nil
	}

	v, ok := m.m[name]
	if ok || m.im == nil {
		return v
	}
	return m.im.Get(name)
}

// Copy creates a copy of the module map.
func (m *ModuleMap) Copy() *ModuleMap {
	c := &ModuleMap{m: make(map[string]Importable), im: m.im}

	for name, mod := range m.m {
		c.m[name] = mod
	}
	return c
}

// SourceModule is an importable module that's written in uGO.
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
	cp.(Map)[AttrModuleName] = String(moduleName)
	return cp, nil
}
