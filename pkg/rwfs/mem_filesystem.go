package rwfs

import (
	// "fmt"
	"os"
)

// MemFileSystem represents an in-memory file system
type MemFileSystem struct {
	mu    RWMutex
	Files map[string]*MemFile
}

// NewMemFileSystem creates a new in-memory file system
func NewMemFileSystem() *MemFileSystem {
	return &MemFileSystem{
		Files: make(map[string]*MemFile),
	}
}

func (fs *MemFileSystem) Open(name string) (File, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	file, exists := fs.Files[name]
	if !exists {
		return nil, os.ErrNotExist
	}
	file.mu.Lock()
	defer file.mu.Unlock()
	file.closed = false
	return file, nil
}

func (fs *MemFileSystem) Create(name, owner string, mode os.FileMode) (File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if _, exists := fs.Files[name]; exists {
		return nil, os.ErrExist
	}
	if mode == 0 {
		mode = ReadWrite // Default mode if not provided
	}
	file := NewMemFile(name, owner, mode)
	fs.Files[name] = file
	return file, nil
}

func (fs *MemFileSystem) Remove(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if _, exists := fs.Files[name]; !exists {
		return os.ErrNotExist
	}
	delete(fs.Files, name)
	return nil
}
func (fs *MemFileSystem) Stat(name string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	file, exists := fs.Files[name]
	if !exists {
		return nil, os.ErrNotExist
	}
	return file.Stat()
}
func (fs *MemFileSystem) ListFiles() ([]string, error) {
    fs.mu.RLock()
    defer fs.mu.RUnlock()
    var fileList []string
    for fileName := range fs.Files {
        fileList = append(fileList, fileName)
    }
    return fileList, nil
}
