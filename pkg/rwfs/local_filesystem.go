package rwfs

import (
	"bytes"
	"encoding/gob"
	"io"
	"os"
	"path"
)

// LocalFileSystem extends MemFileSystem with persistent storage capabilities
type LocalFileSystem struct {
	*MemFileSystem
	mu            RWMutex
	compression   bool
	compressLevel int
	encryption    bool
	encryptionKey string
	RootPath      string
}

// NewLocalFileSystem creates a new LocalFileSystem using the provided configuration
func NewLocalFileSystem(config FileSystemConfig) (*LocalFileSystem, error) {
	fs := &LocalFileSystem{
		MemFileSystem: NewMemFileSystem(config),
		compression:   config.Compression,
		compressLevel: config.CompressLevel,
		encryption:    config.Encryption,
		encryptionKey: config.EncryptionKey,
	}
	if config.Filepath != "" {
		if err := fs.LoadFromFile(config.Filepath); err != nil {
			return nil, err
		}
	}
	return fs, nil
}

// populateFilesFlatMap recursively traverses dir and populates fs.Files
func populateFilesFlatMap(dir *MemDirectory, currentPath string, files map[string]*MemFile) {
	for name, file := range dir.Entries {
		filePath := path.Join(currentPath, name)
		files[filePath] = file
	}
	for dirName, subDir := range dir.Dirs {
		populateFilesFlatMap(subDir, path.Join(currentPath, dirName), files)
	}
}

// SaveToFile saves the in-memory file system to a binary file with optional compression and encryption
func (fs *LocalFileSystem) SaveToFile(filepath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	snapshot := fs.RootDir.ToSerialized()
	if err := encoder.Encode(snapshot); err != nil {
		return err
	}

	data := buf.Bytes()
	if fs.compression {
		var err error
		data, err = CompressData(data, fs.compressLevel)
		if err != nil {
			return err
		}
	}

	if fs.encryption {
		var err error
		data, err = EncryptData(data, fs.encryptionKey)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// LoadFromFile loads the in-memory file system from a binary file with optional decompression and decryption
func (fs *LocalFileSystem) LoadFromFile(filepath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // It's okay if the file doesn't exist
		}
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if fs.encryption {
		data, err = DecryptData(data, fs.encryptionKey)
		if err != nil {
			return err
		}
	}

	if fs.compression {
		data, err = DecompressData(data)
		if err != nil {
			return err
		}
	}

	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	var snapshot SerializedDirectory
	if err := decoder.Decode(&snapshot); err != nil {
		return err
	}

	loadedRoot := snapshot.ToMemDirectory(nil)
	fs.RootDir = loadedRoot
	fs.CWD = loadedRoot

	// Rebuild fs.Files flat map
	fs.Files = make(map[string]*MemFile)
	populateFilesFlatMap(loadedRoot, "", fs.Files)

	return nil
}
