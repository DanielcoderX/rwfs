package rwfs

import (
	"sync"
	"testing"
	"time"
)

func TestRWMutexBasicLockUnlock(t *testing.T) {
	var mu RWMutex
	mu.Lock()
	mu.Unlock()

	mu.RLock()
	mu.RUnlock()
}

func TestRWMutexTryLock(t *testing.T) {
	var mu RWMutex

	// Should succeed when uncontended
	if !mu.TryLock() {
		t.Fatal("expected TryLock to succeed")
	}

	// TryLock should fail when already write-locked
	if mu.TryLock() {
		t.Fatal("expected TryLock to fail when already locked")
	}

	// TryRLock should fail when write-locked
	if mu.TryRLock() {
		t.Fatal("expected TryRLock to fail when write-locked")
	}

	mu.Unlock()

	// Should succeed again after unlocking
	if !mu.TryLock() {
		t.Fatal("expected TryLock to succeed after unlock")
	}
	mu.Unlock()
}

func TestRWMutexTryRLock(t *testing.T) {
	var mu RWMutex

	if !mu.TryRLock() {
		t.Fatal("expected TryRLock to succeed")
	}

	// Another read lock should succeed
	if !mu.TryRLock() {
		t.Fatal("expected second TryRLock to succeed")
	}

	// Write lock should fail while read-locked
	if mu.TryLock() {
		t.Fatal("expected TryLock to fail when read-locked")
	}

	mu.RUnlock()
	mu.RUnlock()

	if !mu.TryLock() {
		t.Fatal("expected TryLock to succeed after releasing read locks")
	}
	mu.Unlock()
}

func TestRWMutexConcurrentNoDeadlock(t *testing.T) {
	var mu RWMutex
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if mu.TryLock() {
					time.Sleep(time.Microsecond * 50)
					mu.Unlock()
				} else if mu.TryRLock() {
					time.Sleep(time.Microsecond * 50)
					mu.RUnlock()
				}
			}
		}()
	}

	wg.Wait()
}
