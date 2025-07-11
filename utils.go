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

// deGenerateCacheKey splits a full cache key back into its individual components
// based on the provided separator.
//
// Parameters:
//   - separator: The string used to split the full key.
//   - fullKey: The full cache key string to be decomposed.
//
// Returns:
//   - A slice of strings representing the original key components. Returns an
func deGenerateCacheKey(separator, fullKey string) []string {
	if len(fullKey) == 0 {
		return []string{}
	}
	return strings.Split(fullKey, separator)
}
