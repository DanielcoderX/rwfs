package main

import (
	// "io"
	"io"
	"log"

	"github.com/DanielcoderX/rwfs/pkg/rwfs"
)

func main() {
	config := rwfs.FileSystemConfig{
		Filepath:      "data.rwfs",
		Compression:   false,
		CompressLevel: 9, // Compression level between 1 (best speed) and 9 (best compression)
		Encryption:    false,
		EncryptionKey: "your-secret-key-here",
	}

	fs, err := rwfs.NewLocalFileSystem(config)
	if err != nil {
		log.Fatalf("Failed to initialize filesystem: %v", err)
	}

	// Example usage with predefined file modes
	file, err := fs.Create("example.txt", "owner1", rwfs.ReadWrite)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	_, err = file.Write([]byte("Hello, RWFS with encryption and extended metadata!"))
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}
	file.Close()

	// Save the file system state
	if err := fs.SaveToFile(config.Filepath); err != nil {
		log.Fatalf("Failed to save file system: %v", err)
	}
	// Load the file system state
	loadedFS, err := rwfs.NewLocalFileSystem(config)
	if err != nil {
		log.Fatalf("Failed to load file system: %v", err)
	}

	loadedFile, err := loadedFS.Open("example.txt")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	data := make([]byte, 1024)
	n, err := loadedFile.Read(data)
	if err != nil && err != io.EOF {
		log.Fatalf("Failed to read from file: %v", err)
	}
	log.Printf("Read data: %s", data[:n])
	log.Println(loadedFile.Stat())
	loadedFile.Close()
	log.Println(loadedFile.Stat())
}
