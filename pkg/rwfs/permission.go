package rwfs

import (
	"os"
)

// Predefined file modes for convenience
const (
	ReadOnly  = os.FileMode(0444)
	WriteOnly = os.FileMode(0222)
	ReadWrite = os.FileMode(0666)
	Exec      = os.FileMode(0777)
)

// FilePermission represents file permissions
type FilePermission struct {
	Read    bool
	Write   bool
	Execute bool
}

// DirPermission represents directory permissions
type DirPermission struct {
	Read    bool
	Write   bool
	Execute bool
	List    bool
}

// SetFilePermissions sets permissions for a file
func (f *MemFile) SetFilePermissions(permissions FilePermission) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.permissions = permissions
}

// SetDirPermissions sets permissions for a directory
func (d *MemDirectory) SetDirPermissions(permissions DirPermission) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.permissions = permissions
}
func (f *MemFile) CheckFilePermission(permission os.FileMode) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if permission&os.FileMode(0400) != 0 && !f.permissions.Read {
		return false
	}
	if permission&os.FileMode(0200) != 0 && !f.permissions.Write {
		return false
	}
	if permission&os.FileMode(0100) != 0 && !f.permissions.Execute {
		return false
	}
	return true
}

func (d *MemDirectory) CheckDirPermission(permission os.FileMode) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if permission&os.FileMode(0400) != 0 && !d.permissions.Read {
		return false
	}
	if permission&os.FileMode(0200) != 0 && !d.permissions.Write {
		return false
	}
	if permission&os.FileMode(0100) != 0 && !d.permissions.Execute {
		return false
	}
	return true
}
