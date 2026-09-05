package rwfs

import (
	"errors"
	"os"
	"path"
	"strings"
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
	Parent      *MemDirectory
	modTime     time.Time
	permissions DirPermission
}

// NewMemDirectory creates a new memory directory
func NewMemDirectory(name string, permissions DirPermission) *MemDirectory {
	d := &MemDirectory{
		Name:        name,
		Entries:     make(map[string]*MemFile),
		Dirs:        make(map[string]*MemDirectory),
		modTime:     time.Now(),
		permissions: permissions,
	}
	d.Parent = d
	return d
}

// resolveDir resolves a path to a target directory starting from CWD or RootDir.
func (fs *MemFileSystem) resolveDir(dirPath string) (*MemDirectory, error) {
	dirPath = path.Clean(dirPath)
	if dirPath == "." || dirPath == "" {
		return fs.CWD, nil
	}
	if dirPath == "/" {
		return fs.RootDir, nil
	}

	var cur *MemDirectory
	if strings.HasPrefix(dirPath, "/") {
		cur = fs.RootDir
		dirPath = strings.TrimPrefix(dirPath, "/")
	} else {
		cur = fs.CWD
	}

	parts := strings.Split(dirPath, "/")
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			if cur.Parent != nil {
				cur = cur.Parent
			}
			continue
		}

		if !cur.permissions.Execute {
			return nil, errors.New("execute permission denied")
		}

		child, exists := cur.Dirs[part]
		if !exists {
			return nil, errors.New("directory does not exist: " + part)
		}
		cur = child
	}

	return cur, nil
}

// resolvePath splits a path into its parent directory and target base name.
func (fs *MemFileSystem) resolvePath(targetPath string) (*MemDirectory, string, error) {
	cleanPath := path.Clean(targetPath)
	dirPart, basePart := path.Split(cleanPath)

	var parentDir *MemDirectory
	var err error

	if dirPart == "" || dirPart == "." {
		parentDir = fs.CWD
	} else {
		parentDir, err = fs.resolveDir(dirPart)
		if err != nil {
			return nil, "", err
		}
	}

	return parentDir, basePart, nil
}

// CreateDir creates a new directory within the file system
func (fs *MemFileSystem) CreateDir(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	parentDir, dirName, err := fs.resolvePath(name)
	if err != nil {
		return err
	}
	if dirName == "" || dirName == "." || dirName == ".." {
		return errors.New("invalid directory name")
	}

	// Check if parent directory has write permissions
	if !parentDir.permissions.Write {
		return errors.New("write permission denied")
	}

	if _, exists := parentDir.Dirs[dirName]; exists {
		return errors.New("directory already exists")
	}

	newDir := NewMemDirectory(dirName, DirPermission{Read: true, Write: true, Execute: true})
	newDir.Parent = parentDir
	parentDir.Dirs[dirName] = newDir
	parentDir.modTime = time.Now()

	return nil
}

// RemoveDir removes a directory from the file system
func (fs *MemFileSystem) RemoveDir(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	parentDir, dirName, err := fs.resolvePath(name)
	if err != nil {
		return err
	}

	if !parentDir.permissions.Write {
		return errors.New("write permission denied")
	}

	targetDir, exists := parentDir.Dirs[dirName]
	if !exists {
		return errors.New("directory does not exist")
	}

	// Prevent removing CWD or any ancestor of CWD
	for cur := fs.CWD; cur != nil; {
		if cur == targetDir {
			return errors.New("cannot remove directory: current working directory or ancestor")
		}
		if cur.Parent == cur || cur.Parent == nil {
			break
		}
		cur = cur.Parent
	}

	delete(parentDir.Dirs, dirName)
	parentDir.modTime = time.Now()

	return nil
}

// ChangeDir changes the current working directory
func (fs *MemFileSystem) ChangeDir(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir, err := fs.resolveDir(name)
	if err != nil {
		return err
	}

	// Check if the target directory has execute permissions
	if !dir.permissions.Execute {
		return errors.New("execute permission denied")
	}

	fs.CWD = dir
	return nil
}

// CreateFile creates a new file in the target directory
func (fs *MemFileSystem) CreateFile(name, owner string, permissions FilePermission) (File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	parentDir, fileName, err := fs.resolvePath(name)
	if err != nil {
		return nil, err
	}
	if fileName == "" || fileName == "." || fileName == ".." {
		return nil, errors.New("invalid file name")
	}

	// Check if the directory has write permissions
	if !parentDir.permissions.Write {
		return nil, errors.New("write permission denied")
	}

	if _, exists := parentDir.Entries[fileName]; exists {
		return nil, os.ErrExist
	}

	file := NewMemFile(fileName, owner, permissions)
	file.Config = fs.Config
	file.Cache = fs.Cache
	parentDir.Entries[fileName] = file
	parentDir.modTime = time.Now()
	if fs.Cache != nil {
		fs.Cache.Put(fileName, file, true)
	}
	return file, nil
}

// OpenFile opens a file in the target directory
func (fs *MemFileSystem) OpenFile(name string) (File, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	parentDir, fileName, err := fs.resolvePath(name)
	if err != nil {
		return nil, err
	}

	var file *MemFile
	if fs.Cache != nil {
		if cachedFile, exists := fs.Cache.Get(fileName); exists {
			file = cachedFile
		}
	}

	if file == nil {
		var exists bool
		file, exists = parentDir.Entries[fileName]
		if !exists {
			return nil, os.ErrNotExist
		}
	}

	// Check if the file has read permissions
	if !file.permissions.Read {
		return nil, errors.New("read permission denied")
	}

	file.mu.Lock()
	file.closed = false
	file.position = 0
	file.mu.Unlock()

	return file, nil
}

// RemoveFile removes a file from the target directory
func (fs *MemFileSystem) RemoveFile(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	parentDir, fileName, err := fs.resolvePath(name)
	if err != nil {
		return err
	}

	if !parentDir.permissions.Write {
		return errors.New("write permission denied")
	}

	file, exists := parentDir.Entries[fileName]
	if !exists {
		return os.ErrNotExist
	}

	if file.inode != nil {
		file.inode.mu.Lock()
		file.inode.refCount--
		shouldFree := file.inode.refCount <= 0
		file.inode.mu.Unlock()
		if shouldFree {
			file.inode.Free()
		}
	}
	delete(parentDir.Entries, fileName)
	parentDir.modTime = time.Now()

	// Remove from cache
	if fs.Cache != nil {
		fs.Cache.Remove(fileName)
	}
	return nil
}

// Rename renames or moves a file or directory
func (fs *MemFileSystem) Rename(oldName, newName string) error {
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

	if !oldParent.permissions.Write || !newParent.permissions.Write {
		return errors.New("write permission denied")
	}

	// Check if old is a file
	if file, exists := oldParent.Entries[oldBase]; exists {
		if _, destExists := newParent.Entries[newBase]; destExists {
			return os.ErrExist
		}
		delete(oldParent.Entries, oldBase)
		file.Name = newBase
		newParent.Entries[newBase] = file
		now := time.Now()
		oldParent.modTime = now
		newParent.modTime = now
		if fs.Cache != nil {
			fs.Cache.Remove(oldBase)
			fs.Cache.Put(newBase, file, true)
		}
		return nil
	}

	// Check if old is a directory
	if dir, exists := oldParent.Dirs[oldBase]; exists {
		if _, destExists := newParent.Dirs[newBase]; destExists {
			return errors.New("directory already exists")
		}
		delete(oldParent.Dirs, oldBase)
		dir.Name = newBase
		dir.Parent = newParent
		newParent.Dirs[newBase] = dir
		now := time.Now()
		oldParent.modTime = now
		newParent.modTime = now
		return nil
	}

	return os.ErrNotExist
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
			ModTime: file.ModTime(),
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
