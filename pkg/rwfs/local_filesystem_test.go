package rwfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileSystemPersistencePlain(t *testing.T) {
	tempDir := t.TempDir()
	savePath := filepath.Join(tempDir, "fs_plain.rwfs")

	config := FileSystemConfig{
		Filepath:      savePath,
		Compression:   false,
		Encryption:    false,
	}

	// Create and populate filesystem
	fs, err := NewLocalFileSystem(config)
	if err != nil {
		t.Fatalf("failed to create local fs: %v", err)
	}

	err = fs.CreateDir("mydir")
	if err != nil {
		t.Fatalf("failed to create mydir: %v", err)
	}

	err = fs.ChangeDir("mydir")
	if err != nil {
		t.Fatalf("failed to change to mydir: %v", err)
	}

	f, err := fs.CreateFile("hello.txt", "alice", FilePermission{Read: true, Write: true})
	if err != nil {
		t.Fatalf("failed to create hello.txt: %v", err)
	}
	_, err = f.Write([]byte("Persisted Content!"))
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	_ = f.Close()

	// Save to file
	err = fs.SaveToFile(savePath)
	if err != nil {
		t.Fatalf("failed to save to file: %v", err)
	}

	// Verify file exists on host disk
	if _, err := os.Stat(savePath); err != nil {
		t.Fatalf("save file does not exist on disk: %v", err)
	}

	// Load into a new LocalFileSystem instance
	fs2, err := NewLocalFileSystem(config)
	if err != nil {
		t.Fatalf("failed to load local fs: %v", err)
	}

	// Verify directory tree and file content in loaded instance
	err = fs2.ChangeDir("mydir")
	if err != nil {
		t.Fatalf("failed to change to mydir in loaded fs: %v", err)
	}

	loadedFile, err := fs2.OpenFile("hello.txt")
	if err != nil {
		t.Fatalf("failed to open hello.txt in loaded fs: %v", err)
	}
	data := make([]byte, 64)
	n, err := loadedFile.Read(data)
	if err != nil {
		t.Fatalf("failed to read from loaded file: %v", err)
	}
	if string(data[:n]) != "Persisted Content!" {
		t.Fatalf("expected 'Persisted Content!', got %q", string(data[:n]))
	}

	// Test navigation with .. in loaded filesystem
	err = fs2.ChangeDir("..")
	if err != nil || fs2.CWD.Name != "/" {
		t.Fatalf("failed to navigate .. to root in loaded fs: %v", err)
	}
}

func TestLocalFileSystemCompressedAndEncrypted(t *testing.T) {
	tempDir := t.TempDir()
	savePath := filepath.Join(tempDir, "fs_secure.rwfs")

	config := FileSystemConfig{
		Filepath:      savePath,
		Compression:   true,
		CompressLevel: 6,
		Encryption:    true,
		EncryptionKey: "supersecretpassphrase123!",
	}

	fs, err := NewLocalFileSystem(config)
	if err != nil {
		t.Fatalf("failed to create secure fs: %v", err)
	}

	_ = fs.CreateDir("secure_dir")
	_ = fs.ChangeDir("secure_dir")

	f, err := fs.Create("confidential.txt")
	if err != nil {
		t.Fatalf("failed to create confidential.txt: %v", err)
	}
	_, _ = f.Write([]byte("Secret data protected by AES-GCM and gzip"))
	_ = f.Close()

	err = fs.SaveToFile(savePath)
	if err != nil {
		t.Fatalf("failed to save secure fs: %v", err)
	}

	// Load with correct key
	fsLoaded, err := NewLocalFileSystem(config)
	if err != nil {
		t.Fatalf("failed to reload secure fs: %v", err)
	}

	opened, err := fsLoaded.Open("secure_dir/confidential.txt")
	if err != nil {
		t.Fatalf("failed to open file via path: %v", err)
	}
	buf := make([]byte, 100)
	n, _ := opened.Read(buf)
	if string(buf[:n]) != "Secret data protected by AES-GCM and gzip" {
		t.Fatalf("unexpected content: %s", string(buf[:n]))
	}

	// Attempt to load with wrong key should fail
	wrongConfig := config
	wrongConfig.EncryptionKey = "wrongpassword"
	_, err = NewLocalFileSystem(wrongConfig)
	if err == nil {
		t.Fatal("expected error loading with incorrect encryption key")
	}
}
