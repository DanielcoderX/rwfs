package rwfs

import (
	"errors"
	"os"
	"time"
)

var _ FileSystem = (*MemFileSystem)(nil)

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
		Config:  config,
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

// Open opens a file in the file system
func (fs *MemFileSystem) Open(name string) (File, error) {
	return fs.OpenFile(name)
}

// Create creates a new file with default owner and permissions
func (fs *MemFileSystem) Create(name string) (File, error) {
	return fs.CreateFile(name, "", FilePermission{Read: true, Write: true})
}

// Remove removes a file from the file system
func (fs *MemFileSystem) Remove(name string) error {
	return fs.RemoveFile(name)
}

// Stat returns file or directory information in the file system
func (fs *MemFileSystem) Stat(name string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if name == "/" {
		return &MemFileInfo{
			name:    fs.RootDir.Name,
			size:    0,
			modTime: fs.RootDir.modTime,
			mode:    os.ModeDir | 0755,
		}, nil
	}

	parentDir, baseName, err := fs.resolvePath(name)
	if err != nil {
		return nil, err
	}

	if file, exists := parentDir.Entries[baseName]; exists {
		if !file.permissions.Read {
			return nil, errors.New("read permission denied")
		}
		return file.Stat()
	}

	if dir, exists := parentDir.Dirs[baseName]; exists {
		if !dir.permissions.Read {
			return nil, errors.New("read permission denied")
		}
		return &MemFileInfo{
			name:    dir.Name,
			size:    0,
			modTime: dir.modTime,
			mode:    os.ModeDir | 0755,
		}, nil
	}

	return nil, os.ErrNotExist
}

// Link creates a hard link to an existing file
func (fs *MemFileSystem) Link(oldName, newName string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	oldParent, oldBase, err := fs.resolvePath(oldName)
	if err != nil {
		return err
	}
	newParent, newBase, err := fs.resolvePath(newName)
	if err != nil {
		return err
	}

	// Check if the old file exists
	oldFile, exists := oldParent.Entries[oldBase]
	if !exists {
		return os.ErrNotExist
	}

	// Check if the new file already exists
	if _, exists := newParent.Entries[newBase]; exists {
		return os.ErrExist
	}

	// Increment reference count and create new link
	oldFile.refCount++
	newParent.Entries[newBase] = oldFile
	newParent.modTime = time.Now()
	return nil
}

// Unlink removes a hard link to a file
func (fs *MemFileSystem) Unlink(name string) error {
	return fs.RemoveFile(name)
}
