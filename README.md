# RWFS - In-Memory Read-Write File System

RWFS is a Go-based in-memory read-write file system that supports various features like directory support, file permissions, memory caching, compression, security, and more. It is designed to provide a simple and efficient way to simulate file system operations in memory.

## Features

- **Directory Support**: Create, delete, and navigate directories.
- **File and Directory Search**: Search files and directories by name or patterns.
- **Permissions and Access Control**: Manage read, write, and execute permissions for files and directories.
- **Caching**: Memory caching of files for improved read performance.
- **Compression**: Support for file compression to save space.
- **File Metadata**: Manage extended metadata for files.
- **Concurrency Support**: Concurrent access to files and directories with support for read-write locks.
- **Error Handling**: Custom error handling for file system operations.
- **Security Features**: Implement security measures like encryption and access control lists.
- **Local File System Integration**: Integration with the local file system for seamless data storage.

## Installation

To install RWFS, use the following command:

```sh
go get github.com/DanielcoderX/rwfs
```

## Usage

### Creating and Initializing the File System

```go
package main

import (
    "fmt"
    "github.com/DanielcoderX/rwfs"
)

func main() {
    // Initialize the in-memory file system
    fs := rwfs.NewMemFileSystem()

    // Create a new directory
    err := fs.CreateDir("example_dir")
    if err != nil {
        fmt.Println("Error creating directory:", err)
        return
    }

    // Change to the new directory
    err = fs.ChangeDir("example_dir")
    if err != nil {
        fmt.Println("Error changing directory:", err)
        return
    }

    // Create a new file in the current directory
    _, err = fs.CreateFile("example_file.txt", "owner1", rwfs.FilePermission{Read: true, Write: true})
    if err != nil {
        fmt.Println("Error creating file:", err)
        return
    }

    // List the contents of the current directory
    contents := fs.CWD.GetDirectoryContents()
    fmt.Println("Current directory contents:", contents)
}
```

### Core Functions

#### CreateDir

Creates a new directory in the current working directory.

```go
func (fs *MemFileSystem) CreateDir(name string) error
```

#### RemoveDir

Removes a directory from the current working directory.

```go
func (fs *MemFileSystem) RemoveDir(name string) error
```

#### ChangeDir

Changes the current working directory.

```go
func (fs *MemFileSystem) ChangeDir(name string) error
```

#### CreateFile

Creates a new file in the current working directory.

```go
func (fs *MemFileSystem) CreateFile(name, owner string, permissions FilePermission) (File, error)
```

#### OpenFile

Opens a file in the current working directory.

```go
func (fs *MemFileSystem) OpenFile(name string) (File, error)
```

#### RemoveFile

Removes a file from the current working directory.

```go
func (fs *MemFileSystem) RemoveFile(name string) error
```

#### ListFiles

Lists all files in the current working directory.

```go
func (fs *MemFileSystem) ListFiles() ([]string, error)
```

#### ListDirContents

Lists the contents of the current working directory (both files and directories).

```go
func (fs *MemFileSystem) ListDirContents() ([]DirEntry, error)
```

### Additional Utilities

#### GetDirectoryContents

Returns the contents of a directory in a structured format.

```go
// DirectoryContents represents the contents of a directory
type DirectoryContents struct {
    DirectoryName  string
    Files          []string
    Subdirectories []string
}

// GetDirectoryContents returns the contents of the directory in a structured format
func (dir *MemDirectory) GetDirectoryContents() DirectoryContents
```

### Example

Here is an example to demonstrate basic operations like creating directories, changing directories, and listing directory contents:

```go
package main

import (
    "fmt"
    "github.com/DanielcoderX/rwfs"
)

func main() {
    // Define the file system configuration
    config := rwfs.FileSystemConfig{
        Filepath:      "data.rwfs",
        Compression:   true,
        CompressLevel: 9,
        Encryption:    false,
    }
    // Initialize a new local file system
    fs, err := rwfs.NewLocalFileSystem(config)
    if err != nil {
        fmt.Println("Error initializing file system:", err)
        return
    }

// Create and navigate directories
    err := fs.CreateDir("dir1")
    if err != nil {
        fmt.Println("Error creating directory:", err)
        return
    }
    err = fs.ChangeDir("dir1")
    if err != nil {
        fmt.Println("Error changing directory:", err)
        return
    }
    err = fs.CreateDir("subdir1")
    if err != nil {
        fmt.Println("Error creating subdirectory:", err)
        return
    }

    // Create a file
    _, err = fs.CreateFile("file1.txt", "owner1", rwfs.FilePermission{Read: true, Write: true})
    if err != nil {
        fmt.Println("Error creating file:", err)
        return
    }

    // Get contents of current directory
    contents := fs.CWD.GetDirectoryContents()
    fmt.Println("Current directory contents:", contents)

    // Remove a directory
    err = fs.ChangeDir("/")
    if err != nil {
        fmt.Println("Error changing directory:", err)
        return
    }
    err = fs.RemoveDir("dir1")
    if err != nil {
        fmt.Println("Error removing directory:", err)
        return
    }
}
```

## License

This project is licensed under the GPL License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Contact

Just open an issue :)
