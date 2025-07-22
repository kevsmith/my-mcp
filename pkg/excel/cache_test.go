package excel

import (
	"os"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestGetCacheConfig(t *testing.T) {
	// Test default config
	config := GetCacheConfig()
	if config.MaxSize != 10 {
		t.Errorf("Expected default MaxSize 10, got %d", config.MaxSize)
	}
	if config.DefaultTTL != 5*time.Minute {
		t.Errorf("Expected default TTL 5 minutes, got %v", config.DefaultTTL)
	}

	// Test environment variable overrides
	oldMaxSize := os.Getenv("EXCEL_CACHE_MAX_SIZE")
	oldTTL := os.Getenv("EXCEL_CACHE_TTL_MINUTES")
	defer func() {
		os.Setenv("EXCEL_CACHE_MAX_SIZE", oldMaxSize)
		os.Setenv("EXCEL_CACHE_TTL_MINUTES", oldTTL)
	}()

	os.Setenv("EXCEL_CACHE_MAX_SIZE", "20")
	os.Setenv("EXCEL_CACHE_TTL_MINUTES", "10")

	config = GetCacheConfig()
	if config.MaxSize != 20 {
		t.Errorf("Expected MaxSize 20 from env var, got %d", config.MaxSize)
	}
	if config.DefaultTTL != 10*time.Minute {
		t.Errorf("Expected TTL 10 minutes from env var, got %v", config.DefaultTTL)
	}
}

func TestNewFileCache(t *testing.T) {
	config := CacheConfig{MaxSize: 5, DefaultTTL: time.Minute}
	cache := NewFileCache(config)

	if cache == nil {
		t.Fatal("NewFileCache returned nil")
	}
	if cache.maxSize != 5 {
		t.Errorf("Expected maxSize 5, got %d", cache.maxSize)
	}
	if cache.defaultTTL != time.Minute {
		t.Errorf("Expected TTL 1 minute, got %v", cache.defaultTTL)
	}
}

func TestFileCacheBasicOperations(t *testing.T) {
	config := CacheConfig{MaxSize: 3, DefaultTTL: time.Hour}
	cache := NewFileCache(config)
	defer cache.Clear()

	// Test cache miss
	if _, found := cache.Get("nonexistent"); found {
		t.Error("Expected cache miss for nonexistent file")
	}

	// Create test file and put in cache
	filePath := createTestExcelFile(t)
	defer os.Remove(filePath)

	file, err := OpenTestFile(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	cache.Put(filePath, file)

	// Test cache hit
	if cachedFile, found := cache.Get(filePath); !found {
		t.Error("Expected cache hit for existing file")
	} else if cachedFile != file {
		t.Error("Cached file doesn't match original")
	}

	// Test size
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}
}

func TestFileCacheTTLExpiration(t *testing.T) {
	config := CacheConfig{MaxSize: 5, DefaultTTL: 50 * time.Millisecond}
	cache := NewFileCache(config)
	defer cache.Clear()

	filePath := createTestExcelFile(t)
	defer os.Remove(filePath)

	file, err := OpenTestFile(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	cache.Put(filePath, file)

	// Should be accessible immediately
	if _, found := cache.Get(filePath); !found {
		t.Error("File should be cached and accessible")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	if _, found := cache.Get(filePath); found {
		t.Error("File should have expired from cache")
	}
}

func TestFileCacheLRUEviction(t *testing.T) {
	config := CacheConfig{MaxSize: 2, DefaultTTL: time.Hour}
	cache := NewFileCache(config)
	defer cache.Clear()

	// Create test files
	file1Path := createTestExcelFile(t)
	file2Path := createTestExcelFile(t)
	file3Path := createTestExcelFile(t)
	defer func() {
		os.Remove(file1Path)
		os.Remove(file2Path)
		os.Remove(file3Path)
	}()

	file1, _ := OpenTestFile(file1Path)
	file2, _ := OpenTestFile(file2Path)
	file3, _ := OpenTestFile(file3Path)

	// Add first two files
	cache.Put(file1Path, file1)
	cache.Put(file2Path, file2)

	if cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Size())
	}

	// Add third file, should evict first
	cache.Put(file3Path, file3)

	if cache.Size() != 2 {
		t.Errorf("Expected cache size still 2, got %d", cache.Size())
	}

	// File1 should be evicted
	if _, found := cache.Get(file1Path); found {
		t.Error("File1 should have been evicted")
	}

	// File2 and File3 should still be there
	if _, found := cache.Get(file2Path); !found {
		t.Error("File2 should still be cached")
	}
	if _, found := cache.Get(file3Path); !found {
		t.Error("File3 should still be cached")
	}
}

func TestFileCacheCleanExpired(t *testing.T) {
	config := CacheConfig{MaxSize: 5, DefaultTTL: 50 * time.Millisecond}
	cache := NewFileCache(config)
	defer cache.Clear()

	filePath := createTestExcelFile(t)
	defer os.Remove(filePath)

	file, err := OpenTestFile(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	cache.Put(filePath, file)

	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Manually clean expired entries
	cache.CleanExpired()

	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after cleaning, got %d", cache.Size())
	}
}

func TestManagerFlushCache(t *testing.T) {
	// Create manager with small cache for testing
	config := CacheConfig{MaxSize: 3, DefaultTTL: time.Hour}
	manager := NewManagerWithConfig(config)
	defer manager.Close()

	// Create test files
	file1Path := createTestExcelFile(t)
	file2Path := createTestExcelFile(t)
	defer func() {
		os.Remove(file1Path)
		os.Remove(file2Path)
	}()

	// Open files to populate cache
	_, err1 := manager.OpenFile(file1Path)
	_, err2 := manager.OpenFile(file2Path)
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to open test files: %v, %v", err1, err2)
	}

	// Set current sheets
	err1 = manager.SetCurrentSheet(file1Path, "Sheet1")
	err2 = manager.SetCurrentSheet(file2Path, "Sheet1")
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to set current sheets: %v, %v", err1, err2)
	}

	// Verify cache has files
	if manager.cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", manager.cache.Size())
	}

	// Verify current sheets are set
	if len(manager.currentSheet) != 2 {
		t.Errorf("Expected 2 current sheet entries, got %d", len(manager.currentSheet))
	}

	// Flush cache
	filesCleared, err := manager.FlushCache()
	if err != nil {
		t.Fatalf("FlushCache failed: %v", err)
	}

	// Verify results
	if filesCleared != 2 {
		t.Errorf("Expected 2 files cleared, got %d", filesCleared)
	}

	if manager.cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after flush, got %d", manager.cache.Size())
	}

	if len(manager.currentSheet) != 0 {
		t.Errorf("Expected 0 current sheet entries after flush, got %d", len(manager.currentSheet))
	}
}

// Helper function to open a test Excel file
func OpenTestFile(filePath string) (*excelize.File, error) {
	return excelize.OpenFile(filePath)
}
