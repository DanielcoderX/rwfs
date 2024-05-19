package rwfs

import (
	"regexp"
	"sync"
)

// SearchResult represents a search result containing the name and whether it's a file or directory
type SearchResult struct {
	Name  string
	IsDir bool
}

// Search searches for files and directories based on the provided pattern
func (fs *MemFileSystem) Search(pattern string) ([]SearchResult, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var results []SearchResult
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for name, file := range fs.CWD.Entries {
		wg.Add(1)
		go func(name string, file *MemFile) {
			defer wg.Done()
			if re.MatchString(name) {
				results = append(results, SearchResult{Name: name, IsDir: false})
			}
		}(name, file)
	}
	for name, dir := range fs.CWD.Dirs {
		wg.Add(1)
		go func(name string, dir *MemDirectory) {
			defer wg.Done()
			if re.MatchString(name) {
				results = append(results, SearchResult{Name: name, IsDir: true})
			}
		}(name, dir)
	}
	wg.Wait()

	return results, nil
}
