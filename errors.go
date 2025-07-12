package bmemcache

import "errors"

var (
	// ErrEmpty is returned when cache is empty.
	ErrEmpty = errors.New("empty")

	// ErrNotFound is returned when a cache entry is not found.
	ErrNotFound = errors.New("not found")

	// ErrExpired is returned when a cache entry has expired.
	ErrExpired = errors.New("expired")
)
