package rwfs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

// MemFile represents an open file handle backed by a block-paged Inode.
type MemFile struct {
	Name        string
	inode       *Inode
	mu          RWMutex
	position    int64
	closed      bool
	permissions FilePermission
	refCount    int
	Data        *bytes.Buffer
	Config      FileSystemConfig
	Cache       *FileCache
}

// NewMemFile creates a new memory file handle with an underlying Inode.
func NewMemFile(name, owner string, permissions FilePermission) *MemFile {
	inode := NewInode(owner, permissions)
	return &MemFile{
		Name:        name,
		inode:       inode,
		Data:        bytes.NewBuffer(nil),
		permissions: permissions,
		refCount:    1,
	}
}

// Inode returns the underlying Inode.
func (f *MemFile) Inode() *Inode {
	return f.inode
}

// SetInode binds this file handle to an existing Inode (e.g., for hard links).
func (f *MemFile) SetInode(in *Inode) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.inode = in
	f.permissions = in.permissions
}

// Size returns the file size in bytes.
func (f *MemFile) Size() int64 {
	if f.inode != nil {
		return f.inode.Size()
	}
	return 0
}

// ModTime returns the modification time of the underlying inode.
func (f *MemFile) ModTime() time.Time {
	if f.inode != nil {
		f.inode.mu.RLock()
		defer f.inode.mu.RUnlock()
		return f.inode.modTime
	}
	return time.Now()
}

// Read reads up to len(p) bytes from the file starting at the handle's cursor position.
func (f *MemFile) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, os.ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}

	if f.position >= f.inode.Size() {
		return 0, io.EOF
	}

	n, err := f.inode.ReadAt(p, f.position)
	f.position += int64(n)
	if n > 0 && err == io.EOF {
		return n, nil
	}
	return n, err
}

// Write writes len(p) bytes to the file at the handle's cursor position.
func (f *MemFile) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, os.ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}

	n, err := f.inode.WriteAt(p, f.position)
	if err != nil {
		return n, err
	}
	f.position += int64(n)
	f.Data = bytes.NewBuffer(f.inode.DirectBytes())

	if f.Cache != nil {
		f.Cache.Put(f.Name, f, true)
	}
	return n, nil
}

// ReadFrom implements io.ReaderFrom for streaming data directly into the block storage.
func (f *MemFile) ReadFrom(r io.Reader) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, os.ErrClosed
	}

	buf := AllocBlock()
	defer FreeBlock(buf)

	var total int64
	for {
		nr, err := r.Read(buf[:])
		if nr > 0 {
			nw, werr := f.inode.WriteAt(buf[:nr], f.position)
			f.position += int64(nw)
			total += int64(nw)
			if werr != nil {
				f.Data = bytes.NewBuffer(f.inode.DirectBytes())
				return total, werr
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			f.Data = bytes.NewBuffer(f.inode.DirectBytes())
			return total, err
		}
	}
	f.Data = bytes.NewBuffer(f.inode.DirectBytes())
	return total, nil
}

// WriteTo implements io.WriterTo for streaming directly from block storage to an io.Writer.
func (f *MemFile) WriteTo(w io.Writer) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, os.ErrClosed
	}

	buf := AllocBlock()
	defer FreeBlock(buf)

	var total int64
	for {
		nr, rerr := f.inode.ReadAt(buf[:], f.position)
		if nr > 0 {
			nw, werr := w.Write(buf[:nr])
			f.position += int64(nw)
			total += int64(nw)
			if werr != nil {
				return total, werr
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				break
			}
			return total, rerr
		}
	}
	return total, nil
}

// BytesAt returns a slice of file data, using zero-copy block slicing when within a single block.
func (f *MemFile) BytesAt(offset int64, length int) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return nil, os.ErrClosed
	}
	return f.inode.BytesAt(offset, length)
}

// Truncate resizes the file to size, freeing surplus blocks to the pool.
func (f *MemFile) Truncate(size int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return os.ErrClosed
	}
	err := f.inode.Truncate(size)
	if err == nil {
		f.Data = bytes.NewBuffer(f.inode.DirectBytes())
	}
	return err
}

// Close closes the file handle.
func (f *MemFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return os.ErrClosed
	}
	f.closed = true
	return nil
}

// Stat returns file information from the underlying inode.
func (f *MemFile) Stat() (os.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	f.inode.mu.RLock()
	defer f.inode.mu.RUnlock()

	return &MemFileInfo{
		name:       f.Name,
		size:       f.inode.size,
		modTime:    f.inode.modTime,
		accessTime: f.inode.accessTime,
		changeTime: f.inode.changeTime,
		owner:      f.inode.owner,
	}, nil
}

// Sync is a no-op for memory files.
func (f *MemFile) Sync() error {
	return nil
}

// Seek sets the handle's cursor position.
func (f *MemFile) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, os.ErrClosed
	}

	size := f.inode.Size()
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.position + offset
	case io.SeekEnd:
		abs = size + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("negative position")
	}
	f.position = abs
	return abs, nil
}

// MemFileInfo implements os.FileInfo for in-memory files.
type MemFileInfo struct {
	name       string
	size       int64
	modTime    time.Time
	accessTime time.Time
	changeTime time.Time
	mode       os.FileMode
	owner      string
}

func (fi *MemFileInfo) Name() string          { return fi.name }
func (fi *MemFileInfo) Size() int64           { return fi.size }
func (fi *MemFileInfo) Mode() os.FileMode     { return fi.mode }
func (fi *MemFileInfo) ModTime() time.Time    { return fi.modTime }
func (fi *MemFileInfo) AccessTime() time.Time { return fi.accessTime }
func (fi *MemFileInfo) ChangeTime() time.Time { return fi.changeTime }
func (fi *MemFileInfo) Owner() string         { return fi.owner }
func (fi *MemFileInfo) IsDir() bool           { return fi.mode.IsDir() }
func (fi *MemFileInfo) Sys() interface{}      { return nil }
