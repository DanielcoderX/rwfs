package rwfs

import (
	"bytes"
	"io"
	"testing"
)

func TestInodeSparseFile(t *testing.T) {
	file := NewMemFile("sparse.bin", "user", FilePermission{Read: true, Write: true})

	// Seek to 1 GiB offset
	gigabyte := int64(1024 * 1024 * 1024)
	pos, err := file.Seek(gigabyte, io.SeekStart)
	if err != nil || pos != gigabyte {
		t.Fatalf("seek failed: pos=%d, err=%v", pos, err)
	}

	// Write 4 bytes
	magic := []byte("DATA")
	n, err := file.Write(magic)
	if err != nil || n != 4 {
		t.Fatalf("write failed: n=%d, err=%v", n, err)
	}

	// Check file size
	expectedSize := gigabyte + 4
	if file.Size() != expectedSize {
		t.Fatalf("expected size %d, got %d", expectedSize, file.Size())
	}

	// Verify only 1 block was actually allocated in the Inode
	in := file.Inode()
	in.mu.RLock()
	blockCount := len(in.blocks)
	in.mu.RUnlock()
	if blockCount != 1 {
		t.Fatalf("expected exactly 1 allocated block for sparse file, got %d", blockCount)
	}

	// Read from beginning: should return zeroes (sparse hole)
	_, _ = file.Seek(0, io.SeekStart)
	holeBuf := make([]byte, 1024)
	n, err = file.Read(holeBuf)
	if err != nil || n != len(holeBuf) {
		t.Fatalf("reading sparse hole failed: n=%d, err=%v", n, err)
	}
	for i, b := range holeBuf {
		if b != 0 {
			t.Fatalf("expected zero byte at offset %d, got %x", i, b)
		}
	}

	// Seek to the written magic bytes and read back
	_, _ = file.Seek(gigabyte, io.SeekStart)
	readMagic := make([]byte, 4)
	n, err = file.Read(readMagic)
	if err != nil || n != 4 || string(readMagic) != "DATA" {
		t.Fatalf("expected 'DATA', got %q (n=%d, err=%v)", string(readMagic), n, err)
	}
}

func TestInodeTruncate(t *testing.T) {
	file := NewMemFile("trunc.bin", "user", FilePermission{Read: true, Write: true})

	// Write 10KB (occupies 3 blocks of 4KB)
	data := make([]byte, 10*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	_, _ = file.Write(data)

	in := file.Inode()
	in.mu.RLock()
	if len(in.blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(in.blocks))
	}
	in.mu.RUnlock()

	// Truncate down to 2KB (should leave 1 block, freeing 2)
	err := file.Truncate(2048)
	if err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
	if file.Size() != 2048 {
		t.Fatalf("expected size 2048, got %d", file.Size())
	}

	in.mu.RLock()
	if len(in.blocks) != 1 {
		t.Fatalf("expected 1 block after truncate, got %d", len(in.blocks))
	}
	in.mu.RUnlock()
}

func TestMemFileReaderFromWriterTo(t *testing.T) {
	file := NewMemFile("stream.bin", "user", FilePermission{Read: true, Write: true})

	// Stream 64KB from a reader directly into file
	sourceData := bytes.Repeat([]byte("0123456789ABCDEF"), 4096) // 64KB
	srcReader := bytes.NewReader(sourceData)

	written, err := file.ReadFrom(srcReader)
	if err != nil || written != int64(len(sourceData)) {
		t.Fatalf("ReadFrom failed: written=%d, err=%v", written, err)
	}
	if file.Size() != int64(len(sourceData)) {
		t.Fatalf("expected size %d, got %d", len(sourceData), file.Size())
	}

	// Stream back out using WriteTo
	_, _ = file.Seek(0, io.SeekStart)
	var dst bytes.Buffer
	readOut, err := file.WriteTo(&dst)
	if err != nil || readOut != int64(len(sourceData)) {
		t.Fatalf("WriteTo failed: readOut=%d, err=%v", readOut, err)
	}
	if !bytes.Equal(dst.Bytes(), sourceData) {
		t.Fatal("streamed data does not match original source data")
	}
}

func TestMemFileBytesAtZeroCopy(t *testing.T) {
	file := NewMemFile("zerocopy.bin", "user", FilePermission{Read: true, Write: true})
	payload := []byte("Fast Zero-Copy Slice Access")
	_, _ = file.Write(payload)

	slice, err := file.BytesAt(5, 9)
	if err != nil {
		t.Fatalf("BytesAt failed: %v", err)
	}
	if string(slice) != "Zero-Copy" {
		t.Fatalf("expected 'Zero-Copy', got %q", string(slice))
	}
}
