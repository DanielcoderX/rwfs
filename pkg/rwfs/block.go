package rwfs

import "sync"

// BlockSize is the fixed size of each memory block in bytes (4KB, standard memory page).
const BlockSize = 4096

// Block represents a single fixed-size page of file data.
type Block [BlockSize]byte

var blockPool = sync.Pool{
	New: func() any {
		return new(Block)
	},
}

// AllocBlock retrieves a Block from the pool and zeroes it out.
func AllocBlock() *Block {
	b := blockPool.Get().(*Block)
	// Clear the block to avoid dirty data leakage
	*b = Block{}
	return b
}

// FreeBlock clears the Block and returns it to the allocation pool.
func FreeBlock(b *Block) {
	if b == nil {
		return
	}
	*b = Block{}
	blockPool.Put(b)
}
