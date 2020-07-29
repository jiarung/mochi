package utils

import "sync"

// SyncMap is a extension for sync.Map when you need Empty method.
type SyncMap struct {
	sync.Map
}

// Empty checks if the map is empty.
func (m *SyncMap) Empty() bool {
	empty := true
	m.Map.Range(func(k, v interface{}) bool {
		empty = false
		return false
	})
	return empty
}
