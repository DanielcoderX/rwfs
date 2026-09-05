package rwfs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

// MemFile represents a file in the memory file system
type MemFile struct {
	Name        string
	Data        *bytes.Buffer
	data        []byte
	mu          RWMutex
	size        int64
	modTime     time.Time
	accessTime  time.Time
	changeTime  time.Time
	owner       string
	position    int64
	closed      bool
	permissions FilePermission
	refCount    int
	Config      FileSystemConfig
	Cache       *FileCache
}

// NewMemFile creates a new memory file
func NewMemFile(name, owner string, permissions FilePermission) *MemFile {
	now := time.Now()
	return &MemFile{
		Name:        name,
		Data:        bytes.NewBuffer(nil),
		data:        make([]byte, 0),
		modTime:     now,
		accessTime:  now,
		changeTime:  now,
		owner:       owner,
		permissions: permissions,
		refCount:    1,
	}
}

// MemFile methods

// Read data from the memory file at the current position
func (f *MemFile) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, os.ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	if f.position >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.position:])
	f.position += int64(n)
	f.accessTime = time.Now()
	return n, nil
}

// Write data to the memory file at the current position
func (f *MemFile) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, os.ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	endPos := f.position + int64(len(p))
	if endPos > int64(len(f.data)) {
		newData := make([]byte, endPos)
		copy(newData, f.data)
		f.data = newData
	}
	copy(f.data[f.position:], p)
	f.position = endPos
	f.size = int64(len(f.data))
	now := time.Now()
	f.modTime = now
	f.changeTime = now
	f.Data = bytes.NewBuffer(f.data)

	// Cache the file after write
	if f.Cache != nil {
		f.Cache.Put(f.Name, f, true)
	}
	return len(p), nil
}

// Close the memory file
func (f *MemFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return os.ErrClosed
	}
	f.closed = true
	return nil
}

func (f *MemFile) Stat() (os.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return &MemFileInfo{
		name:       f.Name,
		size:       f.size,
		modTime:    f.modTime,
		accessTime: f.accessTime,
		changeTime: f.changeTime,
		owner:      f.owner,
	}, nil
}

func (f *MemFile) Sync() error {
	return nil // No-op for in-memory files
}

func (f *MemFile) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, os.ErrClosed
	}
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.position + offset
	case io.SeekEnd:
		abs = int64(len(f.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("negative position")
	}
	f.position = abs
	return abs, nil
}

// MemFileInfo implements os.FileInfo for in-memory files
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
