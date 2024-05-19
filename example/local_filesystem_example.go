package main

import (
	"fmt"
	"log"

	"github.com/DanielcoderX/rwfs/pkg/rwfs"
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
		log.Fatalf("Failed to initialize local file system: %v", err)
	}

	// Create a new directory
	err = fs.CreateDir("example_dir")
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Change directory
	err = fs.ChangeDir("example_dir")
	if err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}

	// Create a new file
	file, err := fs.CreateFile("example_file.txt", "owner1", rwfs.FilePermission{})
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	// Write data to the file
	_, err = file.Write([]byte("Hello, RWFS with local file system!"))
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	// Close the file
	err = file.Close()
	if err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}

	// Open the file
	file, err = fs.OpenFile("example_file.txt")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}

	// Read data from the file
	data := make([]byte, 1024)
	n, err := file.Read(data)
	if err != nil {
		log.Fatalf("Failed to read from file: %v", err)
	}

	// Print the read data
	fmt.Printf("Read data: %s\n", data[:n])

	// Print file information
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to get file information: %v", err)
	}
	fmt.Printf("File information: %+v\n", fileInfo)

	// Close the file
	err = file.Close()
	if err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}

	// Remove the file
	err = fs.RemoveFile("example_file.txt")
	if err != nil {
		log.Fatalf("Failed to remove file: %v", err)
	}
	// Ensure to ChangeDir before deleting the directory
	fs.ChangeDir("/")
	// Remove the directory
	err = fs.RemoveDir("example_dir")
	if err != nil {
		log.Fatalf("Failed to remove directory: %v", err)
	}

	fmt.Println(fs.CWD)
	// Save the file system state
	err = fs.SaveToFile(config.Filepath)
	if err != nil {
		log.Fatalf("Failed to save file system: %v", err)
	}
}
