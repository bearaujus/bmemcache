package bmemcache

import "strings"

// generateEmptyData returns the zero value for a given type T.
//
// Returns:
//   - The zero value of type T.
func generateEmptyData[T any]() T {
	var emptyValue T
	return emptyValue
}

// generateCacheKey creates a cache key by joining the provided keys using the given separator.
//
// Parameters:
//   - separator: The string to use as a separator between keys.
//   - keys: Variadic list of key parts.
//
// Returns:
//   - A string that represents the combined cache key.
func generateCacheKey(separator string, keys ...string) string {
	if len(keys) == 0 {
		keys = []string{""}
	}
	return strings.Join(keys, separator)
}
