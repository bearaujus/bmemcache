package bmemcache

import (
	"encoding/json"
)

// generateEmptyData returns the zero value for a given type T.
//
// Returns:
//   - The zero value of type T.
func generateEmptyData[T any]() T {
	var emptyValue T
	return emptyValue
}

// serializeKey converts a slice of strings into a JSON-encoded string to be used
// as a safe cache key.
//
// This approach avoids ambiguity caused by string separators when generating
// composite keys.
//
// Parameters:
//   - keys: A slice of strings representing the individual parts of the cache key.
//
// Returns:
//   - A JSON-formatted string representing the composite key.
func serializeKey(keys []string) string {
	if keys == nil {
		keys = []string{}
	}
	b, _ := json.Marshal(keys) // Always succeeds for []string
	return string(b)
}

// deserializeKey converts a JSON-encoded cache key string back into its original
// slice of string components.
//
// This function is the inverse of serializeKey and is used for extracting key parts
// from stored cache map keys.
//
// Parameters:
//   - s: A JSON-formatted string representing a composite cache key.
//
// Returns:
//   - A slice of strings representing the original key components.
func deserializeKey(s string) []string {
	var keys []string
	_ = json.Unmarshal([]byte(s), &keys)
	return keys
}
