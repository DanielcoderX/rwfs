package rwfs

import (
	"os"
)

// File interface to be implemented by MemFile for file operations.
type File interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
	Stat() (os.FileInfo, error)
	Seek(offset int64, whence int) (int64, error)
}
