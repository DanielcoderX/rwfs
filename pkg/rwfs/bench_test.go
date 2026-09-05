package rwfs

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
)

func BenchmarkSequentialWrite(b *testing.B) {
	chunk := make([]byte, 4096)
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	b.SetBytes(4096)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		file := NewMemFile("bench_seq_write.bin", "user", FilePermission{Read: true, Write: true})
		for j := 0; j < 100; j++ {
			_, _ = file.Write(chunk)
		}
	}
}

func BenchmarkSequentialRead(b *testing.B) {
	file := NewMemFile("bench_seq_read.bin", "user", FilePermission{Read: true, Write: true})
	payload := bytes.Repeat([]byte("0123456789ABCDEF"), 64*1024) // 1 MB
	_, _ = file.Write(payload)

	buf := make([]byte, 4096)
	b.SetBytes(4096)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = file.Seek(0, io.SeekStart)
		for {
			_, err := file.Read(buf)
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkRandomWrite(b *testing.B) {
	file := NewMemFile("bench_rand_write.bin", "user", FilePermission{Read: true, Write: true})
	chunk := make([]byte, 512)

	b.SetBytes(512)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		offset := int64(rand.Intn(1024)) * 4096
		_, _ = file.Seek(offset, io.SeekStart)
		_, _ = file.Write(chunk)
	}
}

func BenchmarkSparseWrite(b *testing.B) {
	chunk := []byte("SPARSE_DATA")
	gigabyte := int64(1024 * 1024 * 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file := NewMemFile("bench_sparse.bin", "user", FilePermission{Read: true, Write: true})
		_, _ = file.Seek(gigabyte, io.SeekStart)
		_, _ = file.Write(chunk)
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	file := NewMemFile("bench_concurrent.bin", "user", FilePermission{Read: true, Write: true})
	payload := bytes.Repeat([]byte("CONCURRENT_READ_PAYLOAD_"), 4096) // 96 KB
	_, _ = file.Write(payload)

	in := file.Inode()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		buf := make([]byte, 1024)
		for pb.Next() {
			_, _ = in.ReadAt(buf, 0)
		}
	})
}
