package bmemcache

import (
	"testing"
	"time"
)

// TestSetAndGet verifies that a value can be stored and then retrieved.
func TestSetAndGet(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	cache.Set("hello", "greeting")
	value, err := cache.Get("greeting")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if value != "hello" {
		t.Errorf("expected 'hello', got: %s", value)
	}
}

// TestGetNonExistentKey verifies that attempting to retrieve a non-existent key returns ErrNotFound.
func TestGetNonExistentKey(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	_, err := cache.Get("nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

// TestSetWithExp checks that a value set with an expiration is available initially,
// then returns ErrExpired after the duration passes.
func TestSetWithExp(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	cache.SetWithExp("temp", 100*time.Millisecond, "key")
	value, err := cache.Get("key")
	if err != nil {
		t.Fatalf("expected value before expiration, got error: %v", err)
	}
	if value != "temp" {
		t.Errorf("expected 'temp', got: %s", value)
	}
	// Wait for expiration
	time.Sleep(150 * time.Millisecond)
	_, err = cache.Get("key")
	if err != ErrExpired {
		t.Errorf("expected ErrExpired after expiration, got: %v", err)
	}
}

// TestTTL verifies that TTL returns the correct duration for expiring values and -1 for non-expiring values.
func TestTTL(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	// For a non-expiring value, TTL should return -1.
	cache.Set("permanent", "key1")
	ttl, err := cache.TTL("key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ttl != -1 {
		t.Errorf("expected TTL to be -1 for non-expiring value, got: %v", ttl)
	}

	// For an expiring value, TTL should be > 0 initially.
	cache.SetWithExp("temp", 200*time.Millisecond, "key2")
	ttl, err = cache.TTL("key2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ttl <= 0 {
		t.Errorf("expected TTL > 0 for expiring value, got: %v", ttl)
	}

	// Wait until it expires and check that TTL returns ErrExpired.
	time.Sleep(250 * time.Millisecond)
	_, err = cache.TTL("key2")
	if err != ErrExpired {
		t.Errorf("expected ErrExpired after expiration, got: %v", err)
	}

	// Not found
	_, err = cache.TTL("invalid")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

// TestIsExist verifies the IsExist method.
func TestIsExist(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	if cache.IsExist("key") {
		t.Error("expected key to not exist")
	}
	cache.Set("value", "key")
	if !cache.IsExist("key") {
		t.Error("expected key to exist")
	}
}

// TestIsExpired verifies that IsExpired correctly identifies expired and non-expired cache entries.
func TestIsExpired(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	// Case 1: Non-existent key should return an error
	_, err := cache.IsExpired("nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound for non-existent key, got: %v", err)
	}

	// Case 2: A value without expiration should return false
	cache.Set("value", "key")
	isExpired, err := cache.IsExpired("key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isExpired {
		t.Error("expected false for non-expiring key, got true")
	}

	// Case 3: A value with expiration should return true after expiration
	cache.SetWithExp("temp", 100*time.Millisecond, "expiringKey")
	time.Sleep(150 * time.Millisecond) // Wait for expiration
	isExpired, err = cache.IsExpired("expiringKey")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isExpired {
		t.Error("expected true for expired key, got false")
	}
}

// TestDelete checks that deletion of keys works as expected.
func TestDelete(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	err := cache.Delete("key")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound for deleting non-existent key, got: %v", err)
	}
	cache.Set("value", "key")
	err = cache.Delete("key")
	if err != nil {
		t.Errorf("unexpected error on delete: %v", err)
	}
	if cache.IsExist("key") {
		t.Error("expected key to be deleted")
	}
}

// TestClear verifies that the Clear method removes all entries.
func TestClear(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	cache.Set("value", "key1")
	cache.Set("value", "key2")
	cache.Clear()
	if cache.IsExist("key1") || cache.IsExist("key2") {
		t.Error("expected cache to be empty after clear")
	}
}

// TestAutoCleanup verifies that autoCleanup removes expired entries automatically.
func TestAutoCleanup(t *testing.T) {
	// Enable auto-cleanup with a short interval.
	cache := New[string](WithAutoCleanUp(50*time.Millisecond), WithCacheKeySeparator("|"))
	defer cache.Close()

	cache.SetWithExp("temp", 30*time.Millisecond, "key")
	// Wait for auto cleanup to run.
	time.Sleep(100 * time.Millisecond)
	if cache.IsExist("key") {
		t.Error("expected expired key to be auto-cleaned up")
	}
}

// TestClose ensures that calling Close stops the cleanup goroutine and is safe to call multiple times.
func TestClose(t *testing.T) {
	cache := New[string](WithAutoCleanUp(50*time.Millisecond), WithCacheKeySeparator("|"))
	// Call Close multiple times to ensure no panic occurs.
	cache.Close()
	cache.Close()
}
