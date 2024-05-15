package rwfs

import "os"

// FileSystem defines the interface for a read-write file system
type FileSystem interface {
    Open(name string) (File, error)
    Create(name string) (File, error)
    Remove(name string) error
    Stat(name string) (os.FileInfo, error)
	ListFiles() ([]File, error)
    Rename(oldName, newName string) error
}
