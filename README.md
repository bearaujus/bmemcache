# BMemCache - Generic in-memory caching library in Go

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/bearaujus/bmemcache)](https://goreportcard.com/report/github.com/bearaujus/bmemcache)

**bmemcache** is a generic, thread-safe caching library for Go. It provides a flexible interface for setting, retrieving, and managing cached data with support for auto-cleanup of expired items and customizable cache key generation.

## Installation

To install BDataMatrix, run:

```sh
go get github.com/bearaujus/bmemcache
```

## Import

```go
import "github.com/bearaujus/bmemcache"
```

## Features

- Generic cache with type safety
- Automatic expiration and cleanup
- Configurable key generation
- Simple functions for setting, retrieving, and deleting cache entries
- Thread-safe package

## Usage

Below are example of bmemcache basic usage:

```go
package main

import (
	"fmt"
	"time"

	"github.com/bearaujus/bmemcache"
)

func main() {
	// Create a new cache instance with a custom key separator and auto-cleanup enabled every 1 minute.
	cache := bmemcache.New[string](
		bmemcache.WithAutoCleanUp(1*time.Minute),
		bmemcache.WithCacheKeySeparator("|"),
	)
	defer cache.Close() // Always close the cache when done.

	// Set a value without expiration.
	cache.Set("Hello, World!", "greeting")

	// Retrieve the value.
	value, err := cache.Get("greeting")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Cached value:", value)
	}

	// Set a value with expiration.
	cache.SetWithExp("Temporary Data", 5*time.Second, "temp")

	// Check if the key exists.
	if cache.IsExist("temp") {
		fmt.Println("Temp key exists.")
	}

	// Wait for expiration.
	time.Sleep(6 * time.Second)
	if _, err := cache.Get("temp"); err != nil {
		fmt.Println("Temp key has expired:", err)
	}

	// Clear the cache.
	cache.Clear()
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/bearaujus/bmemcache/blob/master/LICENSE) file for details.
