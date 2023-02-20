// Copyright (c) 2020-2023 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package ugo

import (
	"sync"
	"unicode/utf8"
)

// Iterator wraps the methods required to iterate Objects in VM.
type Iterator interface {
	// Next returns true if there are more elements to iterate.
	Next() bool

	// Key returns the key or index value of the current element.
	Key() Object

	// Value returns the value of the current element.
	Value() Object
}

// iteratorObject is used in VM to make an iterable Object.
type iteratorObject struct {
	ObjectImpl
	Iterator
}

var _ Object = (*iteratorObject)(nil)

// ArrayIterator represents an iterator for the array.
type ArrayIterator struct {
	V Array
	i int
}

var _ Iterator = (*ArrayIterator)(nil)

// Next implements Iterator interface.
func (it *ArrayIterator) Next() bool {
	it.i++
	return it.i-1 < len(it.V)
}

// Key implements Iterator interface.
func (it *ArrayIterator) Key() Object {
	return Int(it.i - 1)
}

// Value implements Iterator interface.
func (it *ArrayIterator) Value() Object {
	i := it.i - 1
	if i > -1 && i < len(it.V) {
		return it.V[i]
	}
	return Undefined
}

// BytesIterator represents an iterator for the bytes.
type BytesIterator struct {
	V Bytes
	i int
}

var _ Iterator = (*BytesIterator)(nil)

// Next implements Iterator interface.
func (it *BytesIterator) Next() bool {
	it.i++
	return it.i-1 < len(it.V)
}

// Key implements Iterator interface.
func (it *BytesIterator) Key() Object {
	return Int(it.i - 1)
}

// Value implements Iterator interface.
func (it *BytesIterator) Value() Object {
	i := it.i - 1
	if i > -1 && i < len(it.V) {
		return Int(it.V[i])
	}
	return Undefined
}

// MapIterator represents an iterator for the map.
type MapIterator struct {
	V    Map
	keys []string
	i    int
}

var _ Iterator = (*MapIterator)(nil)

// Next implements Iterator interface.
func (it *MapIterator) Next() bool {
	it.i++
	return it.i-1 < len(it.keys)
}

// Key implements Iterator interface.
func (it *MapIterator) Key() Object {
	return String(it.keys[it.i-1])
}

// Value implements Iterator interface.
func (it *MapIterator) Value() Object {
	v, ok := it.V[it.keys[it.i-1]]
	if !ok {
		return Undefined
	}
	return v
}

// SyncIterator represents an iterator for the SyncMap.
type SyncIterator struct {
	mu sync.Mutex
	Iterator
}

// Next implements Iterator interface.
func (it *SyncIterator) Next() bool {
	it.mu.Lock()
	defer it.mu.Unlock()
	return it.Iterator.Next()
}

// Key implements Iterator interface.
func (it *SyncIterator) Key() Object {
	it.mu.Lock()
	defer it.mu.Unlock()
	return it.Iterator.Key()
}

// Value implements Iterator interface.
func (it *SyncIterator) Value() Object {
	it.mu.Lock()
	defer it.mu.Unlock()
	return it.Iterator.Value()
}

// StringIterator represents an iterator for the string.
type StringIterator struct {
	V String
	i int
	k int
	r rune
}

var _ Iterator = (*StringIterator)(nil)

// Next implements Iterator interface.
func (it *StringIterator) Next() bool {
	if it.i > len(it.V)-1 {
		return false
	}

	r, s := utf8.DecodeRuneInString(string(it.V)[it.i:])
	if r == utf8.RuneError || s == 0 {
		return false
	}

	it.k = it.i
	it.r = r
	it.i += s
	return true
}

// Key implements Iterator interface.
func (it *StringIterator) Key() Object {
	return Int(it.k)
}

// Value implements Iterator interface.
func (it *StringIterator) Value() Object {
	return Char(it.r)
}
