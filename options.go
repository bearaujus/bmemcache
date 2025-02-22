package bmemcache

import "time"

// Option defines a function that configures cache options.
type Option interface {
	// Apply sets the option on the provided configuration.
	Apply(o *option)
}

// option holds configuration settings for the cache.
type option struct {
	// AutoCleanup enables the background cleanup of expired cache entries.
	AutoCleanup bool
	// AutoCleanupInterval defines the interval between automatic cleanup operations.
	AutoCleanupInterval time.Duration
	// CacheKeySeparator is the string used to separate keys when generating the cache key.
	CacheKeySeparator string
}

// WithAutoCleanUp enables auto-cleanup and sets the cleanup interval.
//
// Parameters:
//   - interval: The time interval between automatic cleanup operations.
//
// Returns:
//   - An Option to be passed to the New() function.
func WithAutoCleanUp(interval time.Duration) Option {
	return &withAutoCleanUp{duration: interval}
}

type withAutoCleanUp struct {
	duration time.Duration
}

// Apply sets the auto-cleanup options.
func (w *withAutoCleanUp) Apply(o *option) {
	o.AutoCleanup = true
	o.AutoCleanupInterval = w.duration
	if w.duration == 0 {
		o.AutoCleanupInterval = time.Minute
	}
}

// WithCacheKeySeparator sets a custom separator for cache keys.
//
// Parameters:
//   - separator: The string to use as a separator between keys.
//
// Returns:
//   - An Option to be passed to the New() function.
func WithCacheKeySeparator(separator string) Option {
	return &withCacheKeySeparator{separator: separator}
}

type withCacheKeySeparator struct {
	separator string
}

// Apply sets the cache key separator option.
func (w *withCacheKeySeparator) Apply(o *option) {
	o.CacheKeySeparator = w.separator
}
