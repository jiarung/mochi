package sync

import (
	"sync/atomic"
)

// Array provides thread-safe array collection to any type with fixed length.
type Array struct {
	arr []atomic.Value
}

// NewArray returns new Array.
func NewArray(length int) *Array {
	return &Array{arr: make([]atomic.Value, length)}
}

// Store saves value. user should check k is not out of range.
func (s *Array) Store(k int, v interface{}) {
	s.arr[k].Store(v)
}

// Load returns value. nil if not store yet.
// user should check k is not out of range.
func (s *Array) Load(k int) interface{} {
	return s.arr[k].Load()
}
