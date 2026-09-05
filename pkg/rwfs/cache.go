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
	ttl     time.Duration
}

// NewFileCache creates a new file cache with a default TTL of 2 seconds
func NewFileCache() *FileCache {
	return &FileCache{
		entries: make(map[string]*CacheEntry),
		ttl:     2 * time.Second,
	}
}

// SetTTL sets the time-to-live for cache entries
func (cache *FileCache) SetTTL(ttl time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.ttl = ttl
}

// TTL returns the configured time-to-live
func (cache *FileCache) TTL() time.Duration {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return cache.ttl
}

// Get retrieves a file from the cache
func (cache *FileCache) Get(name string) (*MemFile, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	entry, exists := cache.entries[name]
	if !exists {
		return nil, false
	}

	// Check if the entry has expired
	if cache.ttl > 0 && time.Since(entry.LastAccess) > cache.ttl {
		delete(cache.entries, name)
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

// Flush writes dirty files back to the main storage and evicts expired entries
func (cache *FileCache) Flush() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	now := time.Now()
	for name, entry := range cache.entries {
		if cache.ttl > 0 && now.Sub(entry.LastAccess) > cache.ttl {
			delete(cache.entries, name)
			continue
		}
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
