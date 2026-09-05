package rwfs

import (
	"errors"
	"os"
	"testing"
)

func TestFileSystemInterfaceCompliance(t *testing.T) {
	var _ FileSystem = (*MemFileSystem)(nil)
	var _ FileSystem = (*LocalFileSystem)(nil)
}

func TestMemFileSystemFileOperations(t *testing.T) {
	fs := NewMemFileSystem(FileSystemConfig{})

	// Create file using FileSystem interface method
	f, err := fs.Create("test_file.txt")
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Write and read
	_, err = f.Write([]byte("Testing FS interface"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	_ = f.Close()

	// Open file using FileSystem interface method
	opened, err := fs.Open("test_file.txt")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	data := make([]byte, 64)
	n, err := opened.Read(data)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data[:n]) != "Testing FS interface" {
		t.Fatalf("unexpected content: %s", string(data[:n]))
	}
	_ = opened.Close()

	// Stat
	info, err := fs.Stat("test_file.txt")
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size() != int64(len("Testing FS interface")) {
		t.Fatalf("expected size %d, got %d", len("Testing FS interface"), info.Size())
	}

	// Remove
	err = fs.Remove("test_file.txt")
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// Stat should now return os.ErrNotExist
	_, err = fs.Stat("test_file.txt")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func TestMemFileSystemPermissions(t *testing.T) {
	fs := NewMemFileSystem(FileSystemConfig{})

	// Create file with write-only permission (no read)
	_, err := fs.CreateFile("no_read.txt", "user", FilePermission{Read: false, Write: true})
	if err != nil {
		t.Fatalf("failed to create no_read.txt: %v", err)
	}

	// Open should fail with permission error
	_, err = fs.OpenFile("no_read.txt")
	if err == nil {
		t.Fatal("expected error opening file without read permission")
	}

	// Stat should fail with read permission error
	_, err = fs.Stat("no_read.txt")
	if err == nil {
		t.Fatal("expected error stating file without read permission")
	}
}

func TestMemFileSystemHardLinks(t *testing.T) {
	fs := NewMemFileSystem(FileSystemConfig{})

	f, err := fs.Create("orig.txt")
	if err != nil {
		t.Fatalf("failed to create orig.txt: %v", err)
	}
	_, _ = f.Write([]byte("linked data"))
	_ = f.Close()

	// Create link
	err = fs.Link("orig.txt", "linked.txt")
	if err != nil {
		t.Fatalf("failed to create hard link: %v", err)
	}

	// Remove original
	err = fs.Remove("orig.txt")
	if err != nil {
		t.Fatalf("failed to remove orig.txt: %v", err)
	}

	// Linked file should still exist and have data
	linkedFile, err := fs.Open("linked.txt")
	if err != nil {
		t.Fatalf("failed to open linked.txt: %v", err)
	}
	buf := make([]byte, 32)
	n, _ := linkedFile.Read(buf)
	if string(buf[:n]) != "linked data" {
		t.Fatalf("expected 'linked data', got %s", string(buf[:n]))
	}
}
