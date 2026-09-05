# RWFS - In-Memory Read-Write File System

RWFS is a high-performance Go in-memory read-write file system engine modeled after modern kernel virtual file systems (VFS). It features 4KB block-paged storage, sparse file support, per-inode concurrency, memory caching with TTL, Gzip compression, AES-GCM encryption, and local disk persistence.

## Features

- **4KB Block-Paged Storage**: File data is split into 4096-byte memory pages backed by a `sync.Pool` slab allocator, achieving zero-allocation reads.
- **Sparse File Support**: Unwritten file gaps consume zero physical memory. Random writes at gigabyte offsets allocate only touched pages.
- **Inode & File Handle Separation**: Independent file descriptors (`MemFile`) reference shared metadata (`Inode`), allowing concurrent handles with distinct read/write cursors.
- **Hierarchical Path Navigation**: Full support for absolute (`/`), relative, current (`.`), parent (`..`), and nested paths (`dir/subdir/file`).
- **File & Directory Search**: Concurrent regex-based file and directory searching.
- **Permissions & Access Control**: Granular read, write, and execute permissions for files and directories.
- **Memory Caching with TTL**: Configurable time-to-live caching layer with automatic expiration and pruning.
- **Local File System Persistence**: Complete directory tree serialization with optional Gzip compression and AES-GCM encryption.
- **Zero-Copy Streaming**: Implements `io.ReaderFrom`, `io.WriterTo`, and `BytesAt` for direct slice access.

## Installation

```sh
go get github.com/DanielcoderX/rwfs
```

## Usage

### Quick Start

```go
package main

import (
	"fmt"
	"io"
	"log"

	"github.com/DanielcoderX/rwfs/pkg/rwfs"
)

func main() {
	// Initialize the in-memory file system
	fs := rwfs.NewMemFileSystem(rwfs.FileSystemConfig{})

	// Create and navigate directories
	if err := fs.CreateDir("projects/go"); err != nil {
		log.Fatal(err)
	}
	if err := fs.ChangeDir("projects/go"); err != nil {
		log.Fatal(err)
	}

	// Create a file and write data
	file, err := fs.Create("hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := file.Write([]byte("Hello, RWFS!")); err != nil {
		log.Fatal(err)
	}
	file.Close()

	// Read data back
	opened, err := fs.Open("hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	data, _ := io.ReadAll(opened)
	opened.Close()
	fmt.Printf("Read: %s\n", data)

	// Navigate with parent path '..'
	fs.ChangeDir("..")
	fmt.Printf("CWD: %s\n", fs.CWD.Name)
}
```

## Core API

### FileSystem Interface
```go
type FileSystem interface {
	Open(name string) (File, error)
	Create(name string) (File, error)
	Remove(name string) error
	Stat(name string) (os.FileInfo, error)
	ListFiles() ([]string, error)
	Rename(oldName, newName string) error
}
```

### Directory Operations
- `CreateDir(path string) error`: Creates directory at specified path (supports nested paths).
- `RemoveDir(path string) error`: Removes directory (with protection against deleting current working directory or ancestors).
- `ChangeDir(path string) error`: Changes current working directory (supports `/`, `..`, `.`, and nested paths).
- `ListFiles() ([]string, error)`: Lists all file names in current directory.
- `ListDirContents() ([]DirEntry, error)`: Lists all entries (files and directories) with metadata.
- `Rename(oldName, newName string) error`: Renames or moves files and directories.

### File Operations (`File` Interface)
- `Read(p []byte) (int, error)`: Reads data starting from current cursor position.
- `Write(p []byte) (int, error)`: Writes data at current cursor position without wiping prior content.
- `Seek(offset int64, whence int) (int64, error)`: Repositions cursor (`SeekStart`, `SeekCurrent`, `SeekEnd`).
- `Close() error`: Closes the file handle.
- `Stat() (os.FileInfo, error)`: Returns metadata (`Name`, `Size`, `ModTime`, `Mode`, etc.).
- `ReadFrom(r io.Reader) (int64, error)`: Streams directly from reader into 4KB block storage.
- `WriteTo(w io.Writer) (int64, error)`: Streams directly from 4KB block storage to writer.
- `BytesAt(offset int64, length int) ([]byte, error)`: Zero-copy direct slice access into page memory.

## Examples

Runnable example applications are located in the `example/` directory:

- **Local Storage & Encryption** (`example/local_fs`):
  Demonstrates initializing disk persistence with AES-GCM encryption and Gzip compression.
  ```sh
  go run ./example/local_fs
  ```

- **Memory Caching & TTL** (`example/cache`):
  Demonstrates memory cache insertion, retrieval, and TTL expiration.
  ```sh
  go run ./example/cache
  ```

## Benchmarks

Run benchmarks locally:
```sh
go test -benchmem -bench=. ./pkg/rwfs
```

Benchmark results (Apple M3 Pro):
- **Sequential Read**: `139.45 MB/s` with **`0 B/op`** (0 GC allocations).
- **Concurrent Reads**: **8.5 Million ops/sec** with **`0 B/op`** (0 GC allocations).
- **Sparse 1 GiB Write**: `~10 ms` allocating only touched 4KB pages.

## License

This project is licensed under the GPL License. See the [LICENSE](LICENSE) file for details.
