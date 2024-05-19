package main

import (
	"fmt"
	"time"

	"github.com/DanielcoderX/rwfs/pkg/rwfs"
)

func main() {
	// Create a new file cache
	cache := rwfs.NewFileCache()

	// Create MemFile instances
	memFile1 := rwfs.NewMemFile("key1", "", rwfs.FilePermission{})
	memFile2 := rwfs.NewMemFile("key2", "", rwfs.FilePermission{})

	// Put MemFile instances into the cache
	cache.Put("key1", memFile1, false) // Not dirty
	cache.Put("key2", memFile2, false) // Not dirty

	// Retrieve data from the cache
	data1, exists1 := cache.Get("key1")
	data2, exists2 := cache.Get("key2")

	if exists1 {
		fmt.Printf("Data for key1: %s\n", data1)
	} else {
		fmt.Println("Data for key1 not found in cache")
	}

	if exists2 {
		fmt.Printf("Data for key2: %s\n", data2.Data.Bytes())
	} else {
		fmt.Println("Data for key2 not found in cache")
	}

	// Wait for some time to demonstrate cache expiration
	time.Sleep(time.Second * 3)

	// Check if the cached data expired
	data1, exists1 = cache.Get("key1")
	if exists1 {
		fmt.Printf("Data for key1: %s\n", data1.Data.Bytes())
	} else {
		fmt.Println("Data for key1 expired from cache")
	}

	// Remove data from the cache
	cache.Remove("key2")

	// Try to retrieve data that has been removed
	data2, exists2 = cache.Get("key2")
	if exists2 {
		fmt.Printf("Data for key2: %s\n", data2.Data.Bytes())
	} else {
		fmt.Println("Data for key2 not found in cache (removed)")
	}
}
