package bmemcache

import (
	"sync"
	"time"
)

// BMemCache defines the behavior for in-memory cache.
type BMemCache[T any] interface {
	// Set stores the given data in the cache.
	//
	// Parameters:
	//   - data: The data to cache.
	//   - keys: A variadic list of strings used to generate the cache key.
	Set(data T, keys ...string)

	// Get retrieves the cached data associated with the provided keys.
	//
	// Parameters:
	//   - keys: A variadic list of strings used to generate the cache key.
	//
	// Returns:
	//   - The cached data of type T.
	//   - An error if the key is not found or if the cached entry has expired.
	Get(keys ...string) (T, error)

	// Delete removes an item from the cache based on the provided keys.
	//
	// Parameters:
	//   - keys: A variadic list of strings used to generate the cache key.
	//
	// Returns:
	//   - An error if the key does not exist.
	Delete(keys ...string) error

	// Keys returns a list of all unique cache keys currently stored.
	//
	// Returns:
	//   - A slice of strings, where each string represents a cache key.
	Keys() [][]string

	// KeysFromPrefix returns all cache keys that match the given prefix pattern.
	//
	// Parameters:
	//   - keys: A variadic list of strings used to construct the prefix to match against stored cache keys.
	//           The prefix typically represents the beginning part of a key hierarchy.
	//
	// Returns:
	//   - A slice of strings representing cache keys that start with the specified prefix.
	KeysFromPrefix(keys ...string) [][]string

	// SetWithExp stores the data in the cache with an expiration time.
	//
	// Parameters:
	//   - data: The data to cache.
	//   - duration: The duration after which the cached data expires.
	//               If zero, the data will not expire.
	//   - keys: A variadic list of strings used to generate the cache key.
	SetWithExp(data T, duration time.Duration, keys ...string)

	// IsExist checks if an item exists in the cache for the given keys.
	//
	// Parameters:
	//   - keys: A variadic list of strings used to generate the cache key.
	//
	// Returns:
	//   - true if the item exists, false otherwise.
	IsExist(keys ...string) bool

	// IsExpired checks whether the cached item associated with the given keys is expired.
	//
	// Parameters:
	//   - keys: A variadic list of strings used to generate the cache key.
	//
	// Returns:
	//   - A boolean indicating if the item is expired.
	//   - An error if the key is not found.
	IsExpired(keys ...string) (bool, error)

	// TTL returns the remaining time before the cached item expires.
	//
	// Parameters:
	//   - keys: A variadic list of strings used to generate the cache key.
	//
	// Returns:
	//   - A time.Duration representing the remaining time until expiration.
	//   - An error if the key is not found or if the item has already expired.
	TTL(keys ...string) (time.Duration, error)

	// Clear removes all items from the cache.
	Clear()

	// Close stops the autoCleanup goroutine, releasing any associated resources.
	//
	// This method should be called when the cache is no longer needed.
	Close()
}

// New initializes a new BMemCache instance with optional configuration options.
// It sets up the underlying storage and, if enabled, starts the background auto-cleanup.
//
// Parameters:
//   - options: Variadic list of Option values to configure the cache.
//
// Returns:
//   - A BMemCache instance configured as specified.
//
// Example usage:
//
//	// Import the package and any necessary options.
//	import (
//	    "fmt"
//	    "time"
//	    "github.com/bearaujus/bmemcache"
//	)
//
//	func main() {
//	    // Create a new cache instance with auto-cleanup enabled every 30 seconds
//	    // and a custom key separator ("-").
//	    cache := bmemcache.New[string](
//	        bmemcache.WithAutoCleanUp(30*time.Second),
//	        bmemcache.WithCacheKeySeparator("-"),
//	    )
//
//	    // Set a value in the cache with no expiration.
//	    cache.Set("Hello, World!", "greeting")
//
//	    // Retrieve the value from the cache.
//	    value, err := cache.Get("greeting")
//	    if err != nil {
//	        fmt.Println("Error:", err)
//	    } else {
//	        fmt.Println("Cached Value:", value)
//	    }
//
//	    // Check if a key exists.
//	    if cache.IsExist("greeting") {
//	        fmt.Println("The key 'greeting' exists in the cache.")
//	    }
//
//	    // Get the time-to-live (TTL) for the key (if applicable).
//	    ttl, err := cache.TTL("greeting")
//	    if err == nil {
//	        fmt.Println("TTL for 'greeting':", ttl)
//	    }
//
//	    // Clear the cache when needed.
//	    cache.Clear()
//
//	    // When finished with the cache, call Close() to stop the auto-cleanup goroutine.
//	    cache.Close()
//	}
func New[T any](options ...Option) BMemCache[T] {
	o := &option{}
	for _, v := range options {
		v.Apply(o)
	}
	cache := &bmemCache[T]{
		items:             make(map[string]*cacheEntry[T]),
		cacheKeySeparator: o.CacheKeySeparator,
	}
	if o.AutoCleanup {
		cache.doneChan = make(chan struct{})
		go cache.autoCleanup(o.AutoCleanupInterval)
	}
	return cache
}

type bmemCache[T any] struct {
	items             map[string]*cacheEntry[T]
	cacheKeySeparator string
	mu                sync.RWMutex
	doneOnce          sync.Once
	doneChan          chan struct{}
}

func (c *bmemCache[T]) Set(data T, keys ...string) {
	c.SetWithExp(data, 0, keys...)
}

func (c *bmemCache[T]) SetWithExp(data T, duration time.Duration, keys ...string) {
	key := generateCacheKey(c.cacheKeySeparator, keys...)
	var exp time.Time
	if duration > 0 {
		exp = time.Now().Add(duration)
	}
	c.mu.Lock()
	c.items[key] = &cacheEntry[T]{Data: data, Exp: exp}
	c.mu.Unlock()
}

func (c *bmemCache[T]) Get(keys ...string) (T, error) {
	key := generateCacheKey(c.cacheKeySeparator, keys...)
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return generateEmptyData[T](), ErrNotFound
	}
	if entry.isExpired() {
		c.mu.Lock()
		entry.flush()
		c.mu.Unlock()
		return generateEmptyData[T](), ErrExpired
	}
	return entry.Data, nil
}

func (c *bmemCache[T]) Delete(keys ...string) error {
	key := generateCacheKey(c.cacheKeySeparator, keys...)
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.items[key]; !ok {
		return ErrNotFound
	}
	delete(c.items, key)
	return nil
}

func (c *bmemCache[T]) Keys() [][]string {
	keys := make([][]string, len(c.items))
	c.mu.RLock()
	var i int
	for k := range c.items {
		keys[i] = deGenerateCacheKey(c.cacheKeySeparator, k)
		i++
	}
	c.mu.RUnlock()
	return keys
}

func (c *bmemCache[T]) KeysFromPrefix(keys ...string) [][]string {
	if len(keys) == 0 {
		if _, err := c.Get(generateCacheKey(c.cacheKeySeparator, keys...)); err != nil {
			return [][]string{}
		}
		return [][]string{{}}
	}
	var ret [][]string
	for _, existingKeyFrags := range c.Keys() {
		if len(existingKeyFrags) < len(keys) {
			continue
		}
		match := true
		for i := range keys {
			if keys[i] != existingKeyFrags[i] {
				match = false
				break
			}
		}
		if match {
			ret = append(ret, existingKeyFrags)
		}
	}
	return ret
}

func (c *bmemCache[T]) IsExist(keys ...string) bool {
	key := generateCacheKey(c.cacheKeySeparator, keys...)
	c.mu.RLock()
	_, ok := c.items[key]
	c.mu.RUnlock()
	return ok
}

func (c *bmemCache[T]) IsExpired(keys ...string) (bool, error) {
	key := generateCacheKey(c.cacheKeySeparator, keys...)
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return false, ErrNotFound
	}
	return entry.isExpired(), nil
}

func (c *bmemCache[T]) TTL(keys ...string) (time.Duration, error) {
	key := generateCacheKey(c.cacheKeySeparator, keys...)
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return 0, ErrNotFound
	}
	if entry.Exp.IsZero() {
		return -1, nil // No expiration
	}
	remaining := time.Until(entry.Exp)
	if remaining <= 0 {
		return 0, ErrExpired
	}
	return remaining, nil
}

func (c *bmemCache[T]) Clear() {
	c.mu.Lock()
	c.items = make(map[string]*cacheEntry[T])
	c.mu.Unlock()
}

func (c *bmemCache[T]) Close() {
	c.doneOnce.Do(func() {
		if c.doneChan != nil {
			c.doneChan <- struct{}{}
			close(c.doneChan)
		}
	})
}

func (c *bmemCache[T]) autoCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for key, entry := range c.items {
				if entry.isExpired() {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		case <-c.doneChan:
			return
		}
	}
}
