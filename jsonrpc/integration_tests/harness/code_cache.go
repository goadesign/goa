package harness

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CodeCache caches generated code to avoid regenerating the same DSL multiple times
type CodeCache struct {
	mu       sync.RWMutex
	cacheDir string
	entries  map[string]string // DSL hash -> generated code directory
}

// NewCodeCache creates a new code cache
func NewCodeCache(baseDir string) (*CodeCache, error) {
	cacheDir := filepath.Join(baseDir, ".code_cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	return &CodeCache{
		cacheDir: cacheDir,
		entries:  make(map[string]string),
	}, nil
}

// hashDSL computes a hash of the DSL code for cache lookup
func (c *CodeCache) hashDSL(dslCode string) string {
	h := sha256.New()
	h.Write([]byte(dslCode))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Get retrieves cached generated code directory for the given DSL
func (c *CodeCache) Get(dslCode string) (string, bool) {
	hash := c.hashDSL(dslCode)
	
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	dir, ok := c.entries[hash]
	if !ok {
		return "", false
	}
	
	// Verify directory still exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", false
	}
	
	return dir, true
}

// Put stores the generated code directory for the given DSL
func (c *CodeCache) Put(dslCode string, generatedDir string) error {
	hash := c.hashDSL(dslCode)
	cacheEntryDir := filepath.Join(c.cacheDir, hash)
	
	// Copy generated code to cache
	if err := copyDir(generatedDir, cacheEntryDir); err != nil {
		return fmt.Errorf("failed to cache generated code: %w", err)
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.entries[hash] = cacheEntryDir
	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		
		dstPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		
		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		
		return os.WriteFile(dstPath, data, info.Mode())
	})
}