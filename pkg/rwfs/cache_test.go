package rwfs

import (
	"testing"
	"time"
)

func TestFileCachePutGetRemove(t *testing.T) {
	cache := NewFileCache()
	cache.SetTTL(1 * time.Second)

	file := NewMemFile("file1.txt", "user", FilePermission{Read: true, Write: true})
	cache.Put("file1.txt", file, true)

	retrieved, exists := cache.Get("file1.txt")
	if !exists || retrieved == nil {
		t.Fatal("expected file1.txt to exist in cache")
	}

	cache.Remove("file1.txt")
	_, exists = cache.Get("file1.txt")
	if exists {
		t.Fatal("expected file1.txt to have been removed from cache")
	}
}

func TestFileCacheTTLExpiration(t *testing.T) {
	cache := NewFileCache()
	cache.SetTTL(50 * time.Millisecond)

	file := NewMemFile("expiring.txt", "user", FilePermission{Read: true, Write: true})
	cache.Put("expiring.txt", file, false)

	// Immediate get succeeds
	_, exists := cache.Get("expiring.txt")
	if !exists {
		t.Fatal("expected expiring.txt to exist immediately")
	}

	// Sleep past TTL
	time.Sleep(70 * time.Millisecond)

	// Get should now expire and return false
	_, exists = cache.Get("expiring.txt")
	if exists {
		t.Fatal("expected expiring.txt to have expired")
	}
}

func TestFileCacheFlush(t *testing.T) {
	cache := NewFileCache()
	cache.SetTTL(50 * time.Millisecond)

	f1 := NewMemFile("f1.txt", "user", FilePermission{Read: true, Write: true})
	f2 := NewMemFile("f2.txt", "user", FilePermission{Read: true, Write: true})

	cache.Put("f1.txt", f1, true)
	time.Sleep(60 * time.Millisecond)
	cache.Put("f2.txt", f2, true)

	cache.Flush()

	// f1 was expired, should be evicted by Flush
	_, exists := cache.Get("f1.txt")
	if exists {
		t.Fatal("expected f1.txt to be evicted by Flush")
	}

	// f2 was fresh, should still exist
	entry, exists := cache.Get("f2.txt")
	if !exists || entry == nil {
		t.Fatal("expected f2.txt to exist after Flush")
	}
}
