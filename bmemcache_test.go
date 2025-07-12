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

func TestKeys(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	// Initially empty
	if len(cache.Keys()) != 0 {
		t.Errorf("expected 0 keys in empty cache, got: %d", len(cache.Keys()))
	}

	// Add keys
	cache.Set("val1", "a", "b", "c")
	cache.Set("val2", "x", "y")
	cache.Set("val3", "foo")
	cache.Set("val4")

	keys := cache.Keys()
	if len(keys) != 4 {
		t.Errorf("expected 4 keys, got: %d", len(keys))
	}

	// Ensure all can be retrieved
	if v, err := cache.Get("a", "b", "c"); err != nil || v != "val1" {
		t.Errorf("unexpected get result for a|b|c: %v, %v", v, err)
	}
	if v, err := cache.Get("x", "y"); err != nil || v != "val2" {
		t.Errorf("unexpected get result for x|y: %v, %v", v, err)
	}
	if v, err := cache.Get("foo"); err != nil || v != "val3" {
		t.Errorf("unexpected get result for foo: %v, %v", v, err)
	}
	if v, err := cache.Get(); err != nil || v != "val4" {
		t.Errorf("unexpected get result for empty: %v, %v", v, err)
	}
}

func TestGet(t *testing.T) {
	cache := New[string](WithCacheKeySeparator("|"))
	defer cache.Close()

	// Add keys
	cache.Set("zero")
	cache.Set("one", "a", "b", "c")
	cache.Set("two", "a", "b", "d")
	cache.Set("three", "x", "y", "z")
	cache.Set("short", "a", "b")
	cache.Set("hi", "a", "c")

	// Match: prefix a
	prefixMatches := cache.KeysFromPrefix("a")
	if len(prefixMatches) != 4 {
		t.Errorf("expected 4 prefix matches for a, got: %d", len(prefixMatches))
	}
	// Match: prefix a|b
	prefixMatches = cache.KeysFromPrefix("a", "b")
	if len(prefixMatches) != 3 {
		t.Errorf("expected 3 prefix matches for a|b, got: %d", len(prefixMatches))
	}
	// Test if Get works with returned keys
	for _, keyParts := range prefixMatches {
		if _, err := cache.Get(keyParts...); err != nil {
			t.Errorf("expected key %v to exist, got error: %v", keyParts, err)
		}
	}

	check := map[string]bool{"one": false, "two": false, "short": false, "hi": false}
	data, err := cache.GetsFromPrefix("a")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	for _, v := range data {
		if _, ok := check[v]; !ok {
			t.Errorf("expected value %v not to be exist", v)
		}
		check[v] = true
	}
	for _, v := range check {
		if !v {
			t.Errorf("expected value %v to exist", v)
		}
	}

	_, err = cache.GetsFromPrefix("a", "b", "c", "d")
	if err == nil {
		t.Errorf("expected error, got: %v", err)
	}

	check = map[string]bool{"three": false}
	data, err = cache.GetsFromPrefix("x", "y", "z")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	for _, v := range data {
		if _, ok := check[v]; !ok {
			t.Errorf("expected value %v not to be exist", v)
		}
		check[v] = true
	}
	for _, v := range check {
		if !v {
			t.Errorf("expected value %v to exist", v)
		}
	}

	check = map[string]bool{"zero": false, "one": false, "two": false, "three": false, "short": false, "hi": false}
	data = cache.Gets()
	for _, v := range data {
		if _, ok := check[v]; !ok {
			t.Errorf("expected value %v not to be exist", v)
		}
		check[v] = true
	}
	for _, v := range check {
		if !v {
			t.Errorf("expected value %v to exist", v)
		}
	}

	check = map[string]bool{"zero": false, "one": false, "two": false, "three": false, "short": false, "hi": false}
	data, err = cache.GetsFromPrefix()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	for _, v := range data {
		if _, ok := check[v]; !ok {
			t.Errorf("expected value %v not to be exist", v)
		}
		check[v] = true
	}
	for _, v := range check {
		if !v {
			t.Errorf("expected value %v to exist", v)
		}
	}

	// Match: exact full key
	full := cache.KeysFromPrefix("x", "y", "z")
	if len(full) != 1 {
		t.Errorf("expected 1 exact match for x|y|z, got: %d", len(full))
	}
	if _, err := cache.Get(full[0]...); err != nil {
		t.Errorf("unexpected get failure for exact match: %v", err)
	}

	// empty: key more than
	empty := cache.KeysFromPrefix("x", "y", "z", "d")
	if len(empty) != 0 {
		t.Errorf("expected no matches for empty prefix, got: %d", len(empty))
	}

	// Empty prefix
	prefixMatches = cache.KeysFromPrefix()
	if len(prefixMatches) != 1 {
		t.Errorf("expected 1 prefix matches for empty keys, got: %d", len(prefixMatches))
	}

	// test get value
	r, err := cache.Get(prefixMatches[0]...)
	if err != nil {
		t.Errorf("unexpected get failure for exact match: %v", err)
	}
	if r != "zero" {
		t.Errorf("expected 'zero', got: %s", r)
	}

	cache2 := New[string](WithCacheKeySeparator("|"))
	defer cache2.Close()

	// Empty prefix + empty key
	empty = cache2.KeysFromPrefix()
	if len(empty) != 0 {
		t.Errorf("expected no matches for empty prefix, got: %d", len(empty))
	}

	data = cache2.Gets()
	if len(data) != 0 {
		t.Errorf("expected length 0, got: %v", err)
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

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		keys      []string
		expected  string
	}{
		{
			name:      "multiple keys",
			separator: ":",
			keys:      []string{"user", "123", "settings"},
			expected:  "user:123:settings",
		},
		{
			name:      "single key",
			separator: "|",
			keys:      []string{"token"},
			expected:  "token",
		},
		{
			name:      "empty keys",
			separator: "-",
			keys:      []string{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCacheKey(tt.separator, tt.keys...)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestWithAutoCleanUpApply(t *testing.T) {
	tests := []struct {
		name           string
		duration       time.Duration
		expectedEnable bool
		expectedValue  time.Duration
	}{
		{
			name:           "non-zero duration",
			duration:       10 * time.Second,
			expectedEnable: true,
			expectedValue:  10 * time.Second,
		},
		{
			name:           "zero duration",
			duration:       0,
			expectedEnable: true,
			expectedValue:  time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &option{}
			w := &withAutoCleanUp{duration: tt.duration}
			w.Apply(opt)

			if !opt.AutoCleanup {
				t.Errorf("expected AutoCleanup to be true")
			}
			if opt.AutoCleanupInterval != tt.expectedValue {
				t.Errorf("expected AutoCleanupInterval to be %v, got %v", tt.expectedValue, opt.AutoCleanupInterval)
			}
		})
	}
}
