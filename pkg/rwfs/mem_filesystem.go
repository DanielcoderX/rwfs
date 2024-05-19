package rwfs

import (
	// "fmt"
	"errors"
	"os"
	"time"
)

// MemFileSystem represents an in-memory file system
type MemFileSystem struct {
	mu      RWMutex
	Files   map[string]*MemFile
	RootDir *MemDirectory
	CWD     *MemDirectory
	Config  FileSystemConfig
	Cache   *FileCache
}

// NewMemFileSystem creates a new in-memory file system
func NewMemFileSystem(config FileSystemConfig) *MemFileSystem {
	rootDir := NewMemDirectory("/", DirPermission{Read: true, Write: true, Execute: true})
	cache := NewFileCache()
	return &MemFileSystem{
		Files:   make(map[string]*MemFile),
		RootDir: rootDir,
		CWD:     rootDir,
		Cache:   cache,
	}
}

// MaintainCache starts a goroutine to periodically clear expired cache entries
func (fs *MemFileSystem) MaintainCache() {
	ticker := time.NewTicker(time.Minute * 5)
	for range ticker.C {
		fs.Cache.Flush()
	}
}

// Open opens a file in the current directory
func (fs *MemFileSystem) Open(name string) (File, error) {
	return fs.OpenFile(name)
}

// Create creates a new file in the current directory
func (fs *MemFileSystem) Create(name, owner string, permissions FilePermission) (File, error) {
	return fs.CreateFile(name, owner, permissions)
}

// Remove removes a file in the current directory
func (fs *MemFileSystem) Remove(name string) error {
	return fs.RemoveFile(name)
}

// Stat returns file information in the current directory
func (fs *MemFileSystem) Stat(name string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	file, exists := fs.CWD.Entries[name]
	if !exists {
		return nil, os.ErrNotExist
	}

	// Check if the file has read permissions
	if !file.permissions.Read {
		return nil, errors.New("read permission denied")
	}

	return file.Stat()
}

// Link creates a hard link to an existing file
func (fs *MemFileSystem) Link(oldName, newName string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if the old file exists
	oldFile, exists := fs.CWD.Entries[oldName]
	if !exists {
		return os.ErrNotExist
	}

	// Check if the new file already exists
	if _, exists := fs.CWD.Entries[newName]; exists {
		return os.ErrExist
	}

	// Increment reference count and create new link
	oldFile.refCount++
	fs.CWD.Entries[newName] = oldFile
	fs.CWD.modTime = time.Now()
	return nil
}

// Unlink removes a hard link to a file
func (fs *MemFileSystem) Unlink(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if the file exists
	file, exists := fs.CWD.Entries[name]
	if !exists {
		return os.ErrNotExist
	}

	// Decrement reference count and remove link
	file.refCount--
	if file.refCount == 0 {
		delete(fs.CWD.Entries, name)
	}
	fs.CWD.modTime = time.Now()
	return nil
}
