package rwfs

import (
	"os"
	"testing"
)

func TestDirectoryNavigation(t *testing.T) {
	fs := NewMemFileSystem(FileSystemConfig{})

	// Create directories: /dir1, /dir1/subdir1
	if err := fs.CreateDir("dir1"); err != nil {
		t.Fatalf("failed to create dir1: %v", err)
	}
	if err := fs.ChangeDir("dir1"); err != nil {
		t.Fatalf("failed to change to dir1: %v", err)
	}
	if fs.CWD.Name != "dir1" {
		t.Fatalf("expected CWD name 'dir1', got %s", fs.CWD.Name)
	}

	if err := fs.CreateDir("subdir1"); err != nil {
		t.Fatalf("failed to create subdir1: %v", err)
	}
	if err := fs.ChangeDir("subdir1"); err != nil {
		t.Fatalf("failed to change to subdir1: %v", err)
	}
	if fs.CWD.Name != "subdir1" {
		t.Fatalf("expected CWD name 'subdir1', got %s", fs.CWD.Name)
	}

	// Test .. navigation
	if err := fs.ChangeDir(".."); err != nil {
		t.Fatalf("failed to change to ..: %v", err)
	}
	if fs.CWD.Name != "dir1" {
		t.Fatalf("expected CWD name 'dir1' after .., got %s", fs.CWD.Name)
	}

	// Test .. to root
	if err := fs.ChangeDir(".."); err != nil {
		t.Fatalf("failed to change to .. (root): %v", err)
	}
	if fs.CWD.Name != "/" {
		t.Fatalf("expected CWD name '/' after second .., got %s", fs.CWD.Name)
	}

	// Test .. at root stays at root
	if err := fs.ChangeDir(".."); err != nil {
		t.Fatalf("failed to change to .. from root: %v", err)
	}
	if fs.CWD.Name != "/" {
		t.Fatalf("expected CWD name '/' at root .., got %s", fs.CWD.Name)
	}

	// Test nested path navigation
	if err := fs.ChangeDir("dir1/subdir1"); err != nil {
		t.Fatalf("failed nested ChangeDir 'dir1/subdir1': %v", err)
	}
	if fs.CWD.Name != "subdir1" {
		t.Fatalf("expected CWD 'subdir1', got %s", fs.CWD.Name)
	}

	// Test absolute path navigation
	if err := fs.ChangeDir("/"); err != nil {
		t.Fatalf("failed absolute ChangeDir '/': %v", err)
	}
	if fs.CWD.Name != "/" {
		t.Fatalf("expected CWD '/', got %s", fs.CWD.Name)
	}
}

func TestDirectoryRemoveCWDProtection(t *testing.T) {
	fs := NewMemFileSystem(FileSystemConfig{})
	_ = fs.CreateDir("protect_dir")
	_ = fs.ChangeDir("protect_dir")

	// Attempting to remove current directory should fail
	err := fs.RemoveDir("/protect_dir")
	if err == nil {
		t.Fatal("expected error when removing current working directory")
	}

	// Navigate out and removal should succeed
	_ = fs.ChangeDir("/")
	err = fs.RemoveDir("protect_dir")
	if err != nil {
		t.Fatalf("expected successful removal of protect_dir: %v", err)
	}
}

func TestDirectoryContentsAndList(t *testing.T) {
	fs := NewMemFileSystem(FileSystemConfig{})
	_ = fs.CreateDir("d1")
	_ = fs.CreateDir("d2")
	_, _ = fs.CreateFile("f1.txt", "user", FilePermission{Read: true, Write: true})
	_, _ = fs.CreateFile("f2.txt", "user", FilePermission{Read: true, Write: true})

	files, err := fs.ListFiles()
	if err != nil {
		t.Fatalf("unexpected error from ListFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	entries, err := fs.ListDirContents()
	if err != nil {
		t.Fatalf("unexpected error from ListDirContents: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries (2 dirs, 2 files), got %d", len(entries))
	}

	contents := fs.CWD.GetDirectoryContents()
	if len(contents.Files) != 2 || len(contents.Subdirectories) != 2 {
		t.Fatalf("unexpected GetDirectoryContents result: %+v", contents)
	}
}

func TestDirectoryRename(t *testing.T) {
	fs := NewMemFileSystem(FileSystemConfig{})
	_ = fs.CreateDir("old_dir")
	_, _ = fs.CreateFile("old_file.txt", "user", FilePermission{Read: true, Write: true})

	// Rename file
	err := fs.Rename("old_file.txt", "new_file.txt")
	if err != nil {
		t.Fatalf("failed to rename file: %v", err)
	}
	if _, exists := fs.CWD.Entries["old_file.txt"]; exists {
		t.Fatal("old_file.txt should no longer exist")
	}
	if _, exists := fs.CWD.Entries["new_file.txt"]; !exists {
		t.Fatal("new_file.txt should exist")
	}

	// Rename directory
	err = fs.Rename("old_dir", "new_dir")
	if err != nil {
		t.Fatalf("failed to rename dir: %v", err)
	}
	if _, exists := fs.CWD.Dirs["old_dir"]; exists {
		t.Fatal("old_dir should no longer exist")
	}
	if _, exists := fs.CWD.Dirs["new_dir"]; !exists {
		t.Fatal("new_dir should exist")
	}

	// Rename non-existent should return os.ErrNotExist
	err = fs.Rename("nonexistent", "whatever")
	if err != os.ErrNotExist {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}
