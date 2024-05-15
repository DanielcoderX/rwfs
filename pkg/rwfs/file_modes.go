package rwfs

import "os"

// Predefined file modes for convenience
const (
	ReadOnly  = os.FileMode(0444)
	WriteOnly = os.FileMode(0222)
	ReadWrite = os.FileMode(0666)
	Exec      = os.FileMode(0777)
)
