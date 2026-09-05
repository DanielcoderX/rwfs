package rwfs

import (
	"bytes"
	"encoding/gob"
	"time"
)

func init() {
	gob.Register(&SerializedDirectory{})
	gob.Register(&MemFile{})
	gob.Register(DirPermission{})
	gob.Register(FilePermission{})
}

// SerializedDirectory represents a serializable snapshot of a directory tree
type SerializedDirectory struct {
	Name        string
	Entries     map[string]*MemFile
	Dirs        map[string]*SerializedDirectory
	ModTime     time.Time
	Permissions DirPermission
}

// ToSerialized converts a MemDirectory hierarchy to a SerializedDirectory
func (dir *MemDirectory) ToSerialized() *SerializedDirectory {
	if dir == nil {
		return nil
	}
	s := &SerializedDirectory{
		Name:        dir.Name,
		Entries:     make(map[string]*MemFile, len(dir.Entries)),
		Dirs:        make(map[string]*SerializedDirectory, len(dir.Dirs)),
		ModTime:     dir.modTime,
		Permissions: dir.permissions,
	}
	for k, v := range dir.Entries {
		s.Entries[k] = v
	}
	for k, v := range dir.Dirs {
		s.Dirs[k] = v.ToSerialized()
	}
	return s
}

// ToMemDirectory reconstructs a MemDirectory hierarchy with Parent references
func (s *SerializedDirectory) ToMemDirectory(parent *MemDirectory) *MemDirectory {
	if s == nil {
		return nil
	}
	dir := &MemDirectory{
		Name:        s.Name,
		Entries:     make(map[string]*MemFile, len(s.Entries)),
		Dirs:        make(map[string]*MemDirectory, len(s.Dirs)),
		Parent:      parent,
		modTime:     s.ModTime,
		permissions: s.Permissions,
	}
	if parent == nil {
		dir.Parent = dir
	}
	for k, v := range s.Entries {
		dir.Entries[k] = v
	}
	for k, childS := range s.Dirs {
		dir.Dirs[k] = childS.ToMemDirectory(dir)
	}
	return dir
}

// Custom Gob Encode method for MemFile
func (f *MemFile) GobEncode() ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	var modTime, accessTime, changeTime time.Time
	var owner string
	var permissions FilePermission
	var content []byte

	if f.inode != nil {
		f.inode.mu.RLock()
		modTime = f.inode.modTime
		accessTime = f.inode.accessTime
		changeTime = f.inode.changeTime
		owner = f.inode.owner
		permissions = f.inode.permissions
		content = f.inode.DirectBytes()
		f.inode.mu.RUnlock()
	} else if f.Data != nil {
		content = f.Data.Bytes()
	}

	// Encode the simple fields
	if err := encoder.Encode(f.Name); err != nil {
		return nil, err
	}
	if err := encoder.Encode(modTime); err != nil {
		return nil, err
	}
	if err := encoder.Encode(accessTime); err != nil {
		return nil, err
	}
	if err := encoder.Encode(changeTime); err != nil {
		return nil, err
	}
	if err := encoder.Encode(owner); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.position); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.closed); err != nil {
		return nil, err
	}
	if err := encoder.Encode(permissions); err != nil {
		return nil, err
	}

	// Encode the Data field as a byte slice
	if err := encoder.Encode(content); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Custom Gob Decode method for MemFile
func (f *MemFile) GobDecode(data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	var name, owner string
	var modTime, accessTime, changeTime time.Time
	var position int64
	var closed bool
	var permissions FilePermission

	// Decode the simple fields
	if err := decoder.Decode(&name); err != nil {
		return err
	}
	if err := decoder.Decode(&modTime); err != nil {
		return err
	}
	if err := decoder.Decode(&accessTime); err != nil {
		return err
	}
	if err := decoder.Decode(&changeTime); err != nil {
		return err
	}
	if err := decoder.Decode(&owner); err != nil {
		return err
	}
	if err := decoder.Decode(&position); err != nil {
		return err
	}
	if err := decoder.Decode(&closed); err != nil {
		return err
	}
	if err := decoder.Decode(&permissions); err != nil {
		return err
	}

	// Decode the Data field as a byte slice
	var dataBytes []byte
	if err := decoder.Decode(&dataBytes); err != nil {
		return err
	}

	f.Name = name
	f.position = position
	f.closed = closed
	f.permissions = permissions
	f.refCount = 1

	f.inode = NewInode(owner, permissions)
	f.inode.modTime = modTime
	f.inode.accessTime = accessTime
	f.inode.changeTime = changeTime
	if len(dataBytes) > 0 {
		_, _ = f.inode.WriteAt(dataBytes, 0)
	}
	f.Data = bytes.NewBuffer(dataBytes)

	return nil
}
