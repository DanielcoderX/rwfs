package rwfs

import "errors"

// Define custom error types here if needed
var (
	ErrFileNotFound     = errors.New("file not found")
	ErrFileAlreadyExist = errors.New("file already exists")
	ErrPermissionDenied = errors.New("permission denied")
)
