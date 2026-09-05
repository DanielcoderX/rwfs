package rwfs

import (
	"fmt"
	"testing"
)

func TestSearch(t *testing.T) {
	fs := NewMemFileSystem(FileSystemConfig{})

	// Create a variety of files and directories
	for i := 0; i < 20; i++ {
		fileName := fmt.Sprintf("doc_%02d.txt", i)
		_, err := fs.CreateFile(fileName, "user", FilePermission{Read: true, Write: true})
		if err != nil {
			t.Fatalf("failed to create %s: %v", fileName, err)
		}
	}
	for i := 0; i < 10; i++ {
		dirName := fmt.Sprintf("dir_%02d", i)
		err := fs.CreateDir(dirName)
		if err != nil {
			t.Fatalf("failed to create %s: %v", dirName, err)
		}
	}

	// Search matching files
	results, err := fs.Search(`^doc_0[0-4]\.txt$`)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
	for _, res := range results {
		if res.IsDir {
			t.Fatalf("expected file result, got directory: %s", res.Name)
		}
	}

	// Search matching directories
	dirResults, err := fs.Search(`^dir_0[0-2]$`)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(dirResults) != 3 {
		t.Fatalf("expected 3 results, got %d", len(dirResults))
	}
	for _, res := range dirResults {
		if !res.IsDir {
			t.Fatalf("expected directory result, got file: %s", res.Name)
		}
	}

	// Invalid regex pattern
	_, err = fs.Search(`[invalid(`)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}
