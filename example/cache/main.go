package main

import (
	"fmt"
	"time"

	"github.com/DanielcoderX/rwfs/pkg/rwfs"
)

func main() {
	// Create a new file cache with a 2-second TTL
	cache := rwfs.NewFileCache()
	cache.SetTTL(2 * time.Second)

	// Create MemFile instances and write sample data
	memFile1 := rwfs.NewMemFile("key1", "", rwfs.FilePermission{Read: true, Write: true})
	_, _ = memFile1.Write([]byte("Sample content for key 1"))

	memFile2 := rwfs.NewMemFile("key2", "", rwfs.FilePermission{Read: true, Write: true})
	_, _ = memFile2.Write([]byte("Sample content for key 2"))

	// Put MemFile instances into the cache
	cache.Put("key1", memFile1, false) // Not dirty
	cache.Put("key2", memFile2, false) // Not dirty

	// Retrieve data from the cache
	data1, exists1 := cache.Get("key1")
	data2, exists2 := cache.Get("key2")

	if exists1 {
		fmt.Printf("Data for key1: %s\n", data1.Data.Bytes())
	} else {
		fmt.Println("Data for key1 not found in cache")
	}

	if exists2 {
		fmt.Printf("Data for key2: %s\n", data2.Data.Bytes())
	} else {
		fmt.Println("Data for key2 not found in cache")
	}

	// Wait for some time to demonstrate cache expiration
	fmt.Println("Waiting 3 seconds for cache TTL to expire...")
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
