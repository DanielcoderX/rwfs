package rwfs

// FileSystemConfig holds the configuration options for initializing a LocalFileSystem
type FileSystemConfig struct {
	Filepath      string
	Compression   bool
	CompressLevel int
	Encryption    bool
	EncryptionKey string
}
