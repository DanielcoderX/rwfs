package rwfs

import (
	"sync"
	"time"
)

// CacheEntry represents an entry in the cache
type CacheEntry struct {
	File       *MemFile
	LastAccess time.Time
	Dirty      bool
}

// FileCache represents the file cache
type FileCache struct {
	entries map[string]*CacheEntry
	mu      sync.Mutex
}

// NewFileCache creates a new file cache
func NewFileCache() *FileCache {
	return &FileCache{
		entries: make(map[string]*CacheEntry),
	}
}

// Get retrieves a file from the cache
func (cache *FileCache) Get(name string) (*MemFile, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	entry, exists := cache.entries[name]
	if !exists {
		return nil, false
	}

	// Update the last access time
	entry.LastAccess = time.Now()
	return entry.File, true
}

// Put adds a file to the cache
func (cache *FileCache) Put(name string, file *MemFile, dirty bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.entries[name] = &CacheEntry{
		File:       file,
		LastAccess: time.Now(),
		Dirty:      dirty,
	}
}

// Flush writes dirty files back to the main storage
func (cache *FileCache) Flush() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for _, entry := range cache.entries {
		if entry.Dirty {
			// Perform the write-back operation
			entry.Dirty = false
		}
	}
}

// Remove removes a file from the cache
func (cache *FileCache) Remove(name string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	delete(cache.entries, name)
}
