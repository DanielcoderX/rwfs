package rwfs

import (
	"io"
	"os"
	"testing"
)

func TestMemFileReadWrite(t *testing.T) {
	file := NewMemFile("test.txt", "user", FilePermission{Read: true, Write: true})

	// Initial size
	if file.Size() != 0 {
		t.Fatalf("expected initial size 0, got %d", file.Size())
	}

	// Write data
	content1 := []byte("Hello, ")
	n, err := file.Write(content1)
	if err != nil || n != len(content1) {
		t.Fatalf("unexpected write error or count: n=%d, err=%v", n, err)
	}

	// Write more data (append)
	content2 := []byte("World!")
	n, err = file.Write(content2)
	if err != nil || n != len(content2) {
		t.Fatalf("unexpected write 2 error or count: n=%d, err=%v", n, err)
	}

	expectedLen := int64(len("Hello, World!"))
	if file.Size() != expectedLen {
		t.Fatalf("expected size %d, got %d", expectedLen, file.Size())
	}

	// Seek to beginning and read back
	pos, err := file.Seek(0, io.SeekStart)
	if err != nil || pos != 0 {
		t.Fatalf("unexpected seek error: pos=%d, err=%v", pos, err)
	}

	buf := make([]byte, 64)
	n, err = file.Read(buf)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(buf[:n]) != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got %q", string(buf[:n]))
	}

	// Read past EOF should return io.EOF
	n, err = file.Read(buf)
	if err != io.EOF || n != 0 {
		t.Fatalf("expected io.EOF on end of file, got n=%d, err=%v", n, err)
	}
}

func TestMemFileSeek(t *testing.T) {
	file := NewMemFile("seek_test.txt", "user", FilePermission{Read: true, Write: true})
	_, _ = file.Write([]byte("0123456789"))

	// Seek start
	pos, err := file.Seek(4, io.SeekStart)
	if err != nil || pos != 4 {
		t.Fatalf("expected pos 4, got %d, err %v", pos, err)
	}

	buf := make([]byte, 2)
	n, err := file.Read(buf)
	if err != nil || n != 2 || string(buf) != "45" {
		t.Fatalf("expected '45', got %q", string(buf[:n]))
	}

	// Seek current
	pos, err = file.Seek(1, io.SeekCurrent)
	if err != nil || pos != 7 {
		t.Fatalf("expected pos 7, got %d, err %v", pos, err)
	}

	n, err = file.Read(buf)
	if err != nil || n != 2 || string(buf) != "78" {
		t.Fatalf("expected '78', got %q", string(buf[:n]))
	}

	// Seek end
	pos, err = file.Seek(-2, io.SeekEnd)
	if err != nil || pos != 8 {
		t.Fatalf("expected pos 8, got %d, err %v", pos, err)
	}
	n, err = file.Read(buf)
	if err != nil || n != 2 || string(buf) != "89" {
		t.Fatalf("expected '89', got %q", string(buf[:n]))
	}

	// Invalid whence
	_, err = file.Seek(0, 999)
	if err == nil {
		t.Fatal("expected error for invalid whence")
	}

	// Negative position
	_, err = file.Seek(-100, io.SeekStart)
	if err == nil {
		t.Fatal("expected error for negative position")
	}
}

func TestMemFileOverwriteAtOffset(t *testing.T) {
	file := NewMemFile("overwrite.txt", "user", FilePermission{Read: true, Write: true})
	_, _ = file.Write([]byte("AAAAABBBBB"))

	// Seek to 5 and overwrite "BBBBB" with "CCCCC"
	_, _ = file.Seek(5, io.SeekStart)
	_, _ = file.Write([]byte("CCCCC"))

	_, _ = file.Seek(0, io.SeekStart)
	data := make([]byte, 10)
	_, _ = file.Read(data)

	if string(data) != "AAAAACCCCC" {
		t.Fatalf("expected AAAAACCCCC, got %s", string(data))
	}
}

func TestMemFileClose(t *testing.T) {
	file := NewMemFile("close.txt", "user", FilePermission{Read: true, Write: true})
	err := file.Close()
	if err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	// Close again should return os.ErrClosed
	if err := file.Close(); err != os.ErrClosed {
		t.Fatalf("expected os.ErrClosed, got %v", err)
	}

	// Read on closed file
	buf := make([]byte, 10)
	if _, err := file.Read(buf); err != os.ErrClosed {
		t.Fatalf("expected os.ErrClosed on Read, got %v", err)
	}

	// Write on closed file
	if _, err := file.Write(buf); err != os.ErrClosed {
		t.Fatalf("expected os.ErrClosed on Write, got %v", err)
	}

	// Seek on closed file
	if _, err := file.Seek(0, io.SeekStart); err != os.ErrClosed {
		t.Fatalf("expected os.ErrClosed on Seek, got %v", err)
	}
}

func TestMemFileStat(t *testing.T) {
	file := NewMemFile("stat.txt", "alice", FilePermission{Read: true, Write: true})
	_, _ = file.Write([]byte("12345"))

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("unexpected stat error: %v", err)
	}
	if info.Name() != "stat.txt" {
		t.Fatalf("expected name stat.txt, got %s", info.Name())
	}
	if info.Size() != 5 {
		t.Fatalf("expected size 5, got %d", info.Size())
	}
	if info.IsDir() {
		t.Fatal("expected IsDir to be false")
	}
}
