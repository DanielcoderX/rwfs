package rwfs

import (
	"errors"
	"os"
	"time"
)

// DirEntry represents a directory entry, which could be a file or a directory
type DirEntry struct {
	Name    string
	IsDir   bool
	ModTime time.Time
}

// MemDirectory represents a directory in the memory file system
type MemDirectory struct {
	Name        string
	mu          RWMutex
	Entries     map[string]*MemFile
	Dirs        map[string]*MemDirectory
	modTime     time.Time
	permissions DirPermission
}

// NewMemDirectory creates a new memory directory
func NewMemDirectory(name string, permissions DirPermission) *MemDirectory {
	return &MemDirectory{
		Name:        name,
		Entries:     make(map[string]*MemFile),
		Dirs:        make(map[string]*MemDirectory),
		modTime:     time.Now(),
		permissions: permissions,
	}
}

// CreateDir creates a new directory within the file system
func (fs *MemFileSystem) CreateDir(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if the current directory has write permissions
	if !fs.CWD.permissions.Write {
		return errors.New("write permission denied")
	}

	if _, exists := fs.CWD.Dirs[name]; exists {
		return errors.New("directory already exists")
	}

	newDir := NewMemDirectory(name, DirPermission{Read: true, Write: true, Execute: true})
	// Add the new directory to the appropriate parent directory
	if fs.CWD == fs.RootDir {
		fs.RootDir.Dirs[name] = newDir
	} else {
		fs.CWD.Dirs[name] = newDir
	}

	fs.CWD.modTime = time.Now()

	return nil
}

// RemoveDir removes a directory from the file system
func (fs *MemFileSystem) RemoveDir(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if the directory being removed is the current working directory of any user
	for _, dir := range fs.CWD.Dirs {
		if dir == fs.CWD {
			return errors.New("cannot remove directory: current working directory")
		}
	}

	// Check if the current directory has write permissions
	if !fs.CWD.permissions.Write {
		return errors.New("write permission denied")
	}
	// Check if the directory exists
	if _, exists := fs.CWD.Dirs[name]; !exists {
		return errors.New("directory does not exist. Try PWD and ChangeDir")
	}
	// Remove the directory
	delete(fs.CWD.Dirs, name)
	fs.CWD.modTime = time.Now()

	return nil
}

// ChangeDir changes the current working directory
func (fs *MemFileSystem) ChangeDir(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()


	var dir *MemDirectory
	var exists bool

	if name == "/" {
		dir = fs.RootDir
	} else {
		dir, exists = fs.CWD.Dirs[name]
		if !exists {
			return errors.New("directory does not exist")
		}
	}

	// Check if the target directory has execute permissions
	if !dir.permissions.Execute {
		return errors.New("execute permission denied")
	}

	fs.CWD = dir
	return nil
}

// CreateFile creates a new file in the current directory
func (fs *MemFileSystem) CreateFile(name, owner string, permissions FilePermission) (File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if the current directory has write permissions
	if !fs.CWD.permissions.Write {
		return nil, errors.New("write permission denied")
	}

	if _, exists := fs.CWD.Entries[name]; exists {
		return nil, os.ErrExist
	}

	file := NewMemFile(name, owner, permissions)
	fs.CWD.Entries[name] = file
	fs.CWD.modTime = time.Now()
	fs.Cache.Put(name, file, true)
	return file, nil
}

// OpenFile opens a file in the current directory
func (fs *MemFileSystem) OpenFile(name string) (File, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	if file, exists := fs.Cache.Get(name); exists {
		file.closed = false
		return file, nil
	}
	file, exists := fs.CWD.Entries[name]
	if !exists {
		return nil, os.ErrNotExist
	}
	// Check if the file has read permissions
	if !file.permissions.Read {
		return nil, errors.New("read permission denied")
	}

	file.closed = false
	return file, nil
}

// RemoveFile removes a file from the current directory
func (fs *MemFileSystem) RemoveFile(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if the current directory has write permissions
	if !fs.CWD.permissions.Write {
		return errors.New("write permission denied")
	}
	file, exists := fs.CWD.Entries[name]
	if !exists {
		return os.ErrNotExist
	}
	if file.refCount > 1 {
		file.refCount--
	} else {
		delete(fs.CWD.Entries, name)

	}
	fs.CWD.modTime = time.Now()

	// Remove from cache
	fs.Cache.mu.Lock()
	defer fs.Cache.mu.Unlock()
	delete(fs.Cache.entries, name)
	return nil
}

func (fs *MemFileSystem) ListFiles() ([]string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var fileList []string
	for fileName := range fs.CWD.Entries {
		fileList = append(fileList, fileName)
	}
	return fileList, nil
}

// ListDirContents lists the contents of the current working directory
func (fs *MemFileSystem) ListDirContents() ([]DirEntry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var entries []DirEntry
	for name, file := range fs.CWD.Entries {
		entries = append(entries, DirEntry{
			Name:    name,
			IsDir:   false,
			ModTime: file.modTime,
		})
	}
	for name, dir := range fs.CWD.Dirs {
		entries = append(entries, DirEntry{
			Name:    name,
			IsDir:   true,
			ModTime: dir.modTime,
		})
	}
	return entries, nil
}

// DirectoryContents represents the contents of a directory
type DirectoryContents struct {
    DirectoryName  string
    Files          []string
    Subdirectories []string
}

// GetDirectoryContents returns the contents of the directory in a structured format
func (dir *MemDirectory) GetDirectoryContents() DirectoryContents {
    files := make([]string, 0, len(dir.Entries))
    for fileName := range dir.Entries {
        files = append(files, fileName)
    }

    subDirs := make([]string, 0, len(dir.Dirs))
    for subDirName := range dir.Dirs {
        subDirs = append(subDirs, subDirName)
    }

    return DirectoryContents{
        DirectoryName:  dir.Name,
        Files:          files,
        Subdirectories: subDirs,
    }
}
