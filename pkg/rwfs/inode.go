package rwfs

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

var globalInoCounter uint64

// Inode represents the low-level metadata and block-paged storage of an individual file.
type Inode struct {
	ino         uint64
	mu          sync.RWMutex
	size        int64
	blocks      map[int64]*Block
	modTime     time.Time
	accessTime  time.Time
	changeTime  time.Time
	owner       string
	permissions FilePermission
	refCount    int
}

// NewInode creates a new Inode with initialized metadata and block map.
func NewInode(owner string, permissions FilePermission) *Inode {
	now := time.Now()
	return &Inode{
		ino:         atomic.AddUint64(&globalInoCounter, 1),
		size:        0,
		blocks:      make(map[int64]*Block),
		modTime:     now,
		accessTime:  now,
		changeTime:  now,
		owner:       owner,
		permissions: permissions,
		refCount:    1,
	}
}

// Ino returns the inode number.
func (in *Inode) Ino() uint64 {
	return in.ino
}

// Size returns the size of the file in bytes.
func (in *Inode) Size() int64 {
	in.mu.RLock()
	defer in.mu.RUnlock()
	return in.size
}

// ReadAt reads up to len(p) bytes from the file starting at byte offset off.
// Reading from sparse holes returns zero-filled bytes without allocating memory.
func (in *Inode) ReadAt(p []byte, off int64) (int, error) {
	in.mu.RLock()
	defer in.mu.RUnlock()

	if off < 0 {
		return 0, errors.New("negative offset")
	}
	if len(p) == 0 {
		return 0, nil
	}
	if off >= in.size {
		return 0, io.EOF
	}

	toRead := int64(len(p))
	if off+toRead > in.size {
		toRead = in.size - off
	}

	var bytesRead int64
	currentOff := off

	for bytesRead < toRead {
		blockIdx := currentOff / BlockSize
		blockOff := currentOff % BlockSize
		chunk := BlockSize - blockOff
		if bytesRead+chunk > toRead {
			chunk = toRead - bytesRead
		}

		dest := p[bytesRead : bytesRead+chunk]
		if b, exists := in.blocks[blockIdx]; exists && b != nil {
			copy(dest, b[blockOff:blockOff+chunk])
		} else {
			// Sparse hole: zero fill
			clear(dest)
		}

		bytesRead += chunk
		currentOff += chunk
	}

	in.accessTime = time.Now()
	if bytesRead < int64(len(p)) {
		return int(bytesRead), io.EOF
	}
	return int(bytesRead), nil
}

// WriteAt writes len(p) bytes to the file starting at byte offset off.
// Blocks are allocated from the block pool on-demand only for target pages.
func (in *Inode) WriteAt(p []byte, off int64) (int, error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	if off < 0 {
		return 0, errors.New("negative offset")
	}
	if len(p) == 0 {
		return 0, nil
	}

	var bytesWritten int64
	currentOff := off
	totalBytes := int64(len(p))

	for bytesWritten < totalBytes {
		blockIdx := currentOff / BlockSize
		blockOff := currentOff % BlockSize
		chunk := BlockSize - blockOff
		if bytesWritten+chunk > totalBytes {
			chunk = totalBytes - bytesWritten
		}

		b, exists := in.blocks[blockIdx]
		if !exists || b == nil {
			b = AllocBlock()
			in.blocks[blockIdx] = b
		}

		copy(b[blockOff:blockOff+chunk], p[bytesWritten:bytesWritten+chunk])

		bytesWritten += chunk
		currentOff += chunk
	}

	if currentOff > in.size {
		in.size = currentOff
	}

	now := time.Now()
	in.modTime = now
	in.changeTime = now

	return int(bytesWritten), nil
}

// Truncate resizes the file to the specified size, releasing surplus blocks to the pool.
func (in *Inode) Truncate(newSize int64) error {
	in.mu.Lock()
	defer in.mu.Unlock()

	if newSize < 0 {
		return errors.New("negative size")
	}

	if newSize < in.size {
		for idx, b := range in.blocks {
			blockStart := idx * BlockSize
			if blockStart >= newSize {
				FreeBlock(b)
				delete(in.blocks, idx)
			} else if blockStart+BlockSize > newSize {
				// Zero out trailing portion in partially kept block
				zeroStart := newSize - blockStart
				clear(b[zeroStart:])
			}
		}
	}

	in.size = newSize
	now := time.Now()
	in.modTime = now
	in.changeTime = now
	return nil
}

// Free releases all allocated blocks back to the pool.
func (in *Inode) Free() {
	in.mu.Lock()
	defer in.mu.Unlock()

	for idx, b := range in.blocks {
		FreeBlock(b)
		delete(in.blocks, idx)
	}
	in.size = 0
}

// BytesAt returns a slice of file data.
// If the requested range falls entirely within a single allocated block, it returns
// a zero-copy slice directly referencing block memory.
func (in *Inode) BytesAt(off int64, length int) ([]byte, error) {
	in.mu.RLock()
	defer in.mu.RUnlock()

	if off < 0 || length < 0 {
		return nil, errors.New("invalid offset or length")
	}
	if off >= in.size {
		return nil, io.EOF
	}

	actualLen := int64(length)
	if off+actualLen > in.size {
		actualLen = in.size - off
	}

	blockIdx := off / BlockSize
	blockOff := off % BlockSize

	// Check if the range fits within a single block
	if blockOff+actualLen <= BlockSize {
		if b, exists := in.blocks[blockIdx]; exists && b != nil {
			// Zero-copy slice view
			return b[blockOff : blockOff+actualLen], nil
		}
		// Sparse hole: return zero slice
		return make([]byte, actualLen), nil
	}

	// Range crosses block boundaries: allocate buffer and read
	buf := make([]byte, actualLen)
	in.mu.RUnlock()
	n, err := in.ReadAt(buf, off)
	in.mu.RLock()
	return buf[:n], err
}

// DirectBytes returns the complete file data as a contiguous byte slice.
func (in *Inode) DirectBytes() []byte {
	in.mu.RLock()
	defer in.mu.RUnlock()

	if in.size == 0 {
		return nil
	}

	data := make([]byte, in.size)
	for idx, b := range in.blocks {
		start := idx * BlockSize
		if start >= in.size {
			continue
		}
		end := start + BlockSize
		if end > in.size {
			end = in.size
		}
		copy(data[start:end], b[:end-start])
	}
	return data
}
