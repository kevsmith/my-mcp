package excel

import (
	"container/list"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
)

// CacheEntry represents a cached Excel file with TTL
type CacheEntry struct {
	file     *excelize.File
	expireAt time.Time
	listNode *list.Element
}

// FileCache is an LRU cache with TTL for Excel files
type FileCache struct {
	mutex      sync.RWMutex
	cache      map[string]*CacheEntry
	lruList    *list.List
	maxSize    int
	defaultTTL time.Duration
}

// CacheConfig holds cache configuration parameters
type CacheConfig struct {
	MaxSize    int
	DefaultTTL time.Duration
}

// GetCacheConfig returns cache configuration from environment variables or defaults
func GetCacheConfig() CacheConfig {
	config := CacheConfig{
		MaxSize:    10,              // Default max 10 files
		DefaultTTL: 5 * time.Minute, // Default 5 minute TTL
	}

	if maxSizeStr := os.Getenv("EXCEL_CACHE_MAX_SIZE"); maxSizeStr != "" {
		if maxSize, err := strconv.Atoi(maxSizeStr); err == nil && maxSize > 0 {
			config.MaxSize = maxSize
		}
	}

	if ttlStr := os.Getenv("EXCEL_CACHE_TTL_MINUTES"); ttlStr != "" {
		if ttlMinutes, err := strconv.Atoi(ttlStr); err == nil && ttlMinutes > 0 {
			config.DefaultTTL = time.Duration(ttlMinutes) * time.Minute
		}
	}

	return config
}

// NewFileCache creates a new LRU cache with TTL for Excel files
func NewFileCache(config CacheConfig) *FileCache {
	return &FileCache{
		cache:      make(map[string]*CacheEntry),
		lruList:    list.New(),
		maxSize:    config.MaxSize,
		defaultTTL: config.DefaultTTL,
	}
}

// Get retrieves a file from the cache if it exists and hasn't expired
func (fc *FileCache) Get(filePath string) (*excelize.File, bool) {
	// Fast path with read lock for cache hit without expiration
	fc.mutex.RLock()
	entry, exists := fc.cache[filePath]
	if !exists {
		fc.mutex.RUnlock()
		return nil, false
	}

	// Check expiration with read lock first
	now := time.Now()
	if now.After(entry.expireAt) {
		fc.mutex.RUnlock()
		// Need write lock to remove expired entry
		fc.mutex.Lock()
		// Double-check after acquiring write lock (could have changed)
		if entry, exists := fc.cache[filePath]; exists && now.After(entry.expireAt) {
			fc.removeEntry(filePath, entry)
		}
		fc.mutex.Unlock()
		return nil, false
	}

	// Cache hit - need to update LRU, so upgrade to write lock
	file := entry.file
	fc.mutex.RUnlock()

	fc.mutex.Lock()
	// Double-check entry still exists and isn't expired
	if entry, exists := fc.cache[filePath]; exists && !time.Now().After(entry.expireAt) {
		fc.lruList.MoveToFront(entry.listNode)
	}
	fc.mutex.Unlock()

	return file, true
}

// Put stores a file in the cache
func (fc *FileCache) Put(filePath string, file *excelize.File) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	// If already exists, update it
	if entry, exists := fc.cache[filePath]; exists {
		entry.file = file
		entry.expireAt = time.Now().Add(fc.defaultTTL)
		fc.lruList.MoveToFront(entry.listNode)
		return
	}

	// Create new entry
	entry := &CacheEntry{
		file:     file,
		expireAt: time.Now().Add(fc.defaultTTL),
	}

	// Add to front of LRU list
	entry.listNode = fc.lruList.PushFront(filePath)
	fc.cache[filePath] = entry

	// Evict oldest entries if cache is full
	for fc.lruList.Len() > fc.maxSize {
		fc.evictOldest()
	}
}

// Clear removes all entries from the cache
func (fc *FileCache) Clear() {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	for filePath, entry := range fc.cache {
		if entry.file != nil {
			entry.file.Close()
		}
		delete(fc.cache, filePath)
	}
	fc.lruList.Init()
}

// CleanExpired removes all expired entries from the cache
func (fc *FileCache) CleanExpired() {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	now := time.Now()
	var toRemove []string

	for filePath, entry := range fc.cache {
		if now.After(entry.expireAt) {
			toRemove = append(toRemove, filePath)
		}
	}

	for _, filePath := range toRemove {
		if entry := fc.cache[filePath]; entry != nil {
			fc.removeEntry(filePath, entry)
		}
	}
}

// Size returns the current number of cached files
func (fc *FileCache) Size() int {
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()
	return len(fc.cache)
}

// removeEntry removes an entry from both cache map and LRU list
func (fc *FileCache) removeEntry(filePath string, entry *CacheEntry) {
	if entry.file != nil {
		entry.file.Close()
	}
	delete(fc.cache, filePath)
	fc.lruList.Remove(entry.listNode)
}

// evictOldest removes the least recently used entry
func (fc *FileCache) evictOldest() {
	if fc.lruList.Len() == 0 {
		return
	}

	oldest := fc.lruList.Back()
	if oldest != nil {
		filePath := oldest.Value.(string)
		if entry := fc.cache[filePath]; entry != nil {
			fc.removeEntry(filePath, entry)
		}
	}
}

// StartCleanupTicker starts a background goroutine to periodically clean expired entries
func (fc *FileCache) StartCleanupTicker(interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			fc.CleanExpired()
		}
	}()
	return ticker
}
