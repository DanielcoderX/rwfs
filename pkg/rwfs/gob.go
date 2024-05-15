package rwfs

import (
	"bytes"
	"encoding/gob"
)

// Custom Gob Encode method for MemFile
func (f *MemFile) GobEncode() ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode the simple fields
	if err := encoder.Encode(f.Name); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.mode); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.modTime); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.accessTime); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.changeTime); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.owner); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.position); err != nil {
		return nil, err
	}
	if err := encoder.Encode(f.closed); err != nil {
		return nil, err
	}

	// Encode the Data field as a byte slice
	if err := encoder.Encode(f.Data.Bytes()); err != nil {
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

	// Decode the simple fields
	if err := decoder.Decode(&f.Name); err != nil {
		return err
	}
	if err := decoder.Decode(&f.mode); err != nil {
		return err
	}
	if err := decoder.Decode(&f.modTime); err != nil {
		return err
	}
	if err := decoder.Decode(&f.accessTime); err != nil {
		return err
	}
	if err := decoder.Decode(&f.changeTime); err != nil {
		return err
	}
	if err := decoder.Decode(&f.owner); err != nil {
		return err
	}
	if err := decoder.Decode(&f.position); err != nil {
		return err
	}
	if err := decoder.Decode(&f.closed); err != nil {
		return err
	}

	// Decode the Data field as a byte slice
	var dataBytes []byte
	if err := decoder.Decode(&dataBytes); err != nil {
		return err
	}
	f.Data = bytes.NewBuffer(dataBytes)

	return nil
}
