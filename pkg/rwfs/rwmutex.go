package rwfs

import (
	"sync"
)

// RWMutex is a wrapper around sync.RWMutex to provide additional functionalities.
type RWMutex struct {
	mu sync.RWMutex
}

// Lock locks the mutex for writing.
func (m *RWMutex) Lock() {
	m.mu.Lock()
}

// Unlock unlocks the mutex for writing.
func (m *RWMutex) Unlock() {
	m.mu.Unlock()
}

// RLock locks the mutex for reading.
func (m *RWMutex) RLock() {
	m.mu.RLock()
}

// RUnlock unlocks the mutex for reading.
func (m *RWMutex) RUnlock() {
	m.mu.RUnlock()
}

// TryLock attempts to lock the mutex for writing without blocking.
func (m *RWMutex) TryLock() bool {
	return m.mu.TryLock()
}

// TryRLock attempts to lock the mutex for reading without blocking.
func (m *RWMutex) TryRLock() bool {
	return m.mu.TryRLock()
}

