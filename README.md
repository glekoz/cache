# cache

### A lightweight, in-memory cache for Go with TTL support — simple, fast, and well-tested.

---

## Table of contents

* [Quick highlights](#quick-highlights)
* [Installation](#installation)
* [Getting started](#getting-started)
* [API overview](#api-overview)

  * [Constructor](#constructor)
  * [Add](#add)
  * [Get](#get)
  * [Delete](#delete)
  * [Options](#options)
* [Implementation details](#implementation-details)
* [Examples](#examples)
* [Tests & Coverage](#tests--coverage)
* [Contributing](#contributing)
* [License](#license)

---

## Quick highlights

* Generic cache using Go generics: `cmp.Ordered` for keys, `any` for values .
* Per-key TTL with truncation to a configurable step (default: `1s`).
* Thread-safe public methods (`Add`, `Get`, `Delete`) via `sync.RWMutex`.
* Internal structures optimize expiry cleanup: `queue map[time.Time][]K` and ordered `times []time.Time`.
* High test coverage in the repository: **96.8%**.

---

## Installation

```bash
go get github.com/glekoz/cache
```

---

## Getting started

Create a cache instance, add items with a TTL, and read them back. TTLs are rounded down to the cache `step` (default `1s`), and TTL must be at least `step`.

See the [Examples](#examples) section for a full snippet.

---

## API overview

### Constructor

```go
func New[K cmp.Ordered, V any](opts ...Option) (*inMemoryCache[K, V], error)
```

Creates a new cache instance. Returns a pointer to `inMemoryCache[K, V]` and an error if any option validation fails.

### Add

```go
func (c *inMemoryCache[K, V]) Add(key K, val V, ttl time.Duration) error
```

Adds a key with a TTL. Important rules:

* `ttl` must be **>=** cache `step` (default `1s`), otherwise an error is returned.
* `ttl` is truncated using `ttl.Truncate(step)`, so values are rounded down to the nearest step.
* If the cache reaches `cacheSize`, `clean()` is invoked to remove expired items and adjust internals.

### Get

```go
func (c *inMemoryCache[K, V]) Get(key K) (V, bool)
```

Returns `(value, true)` if the key is present and not expired. If the key is expired or absent, returns the zero value for `V` and `false`. Expired keys are removed on read.

### Delete

```go
func (c *inMemoryCache[K, V]) Delete(keys ...K)
```

Deletes one or more keys from the cache.

### Options

Options configure initial internal capacities and sizing hints.

* `WithCacheSize(size int)` — initial map size for the cache.
* `WithQueueSize(size int)` — initial map size for the expiry queue.
* `WithTimeSize(size int)` — initial capacity for the `times` slice.
* `WithQueueKeySize(size int)` — initial capacity for key slices inside `queue`.

Example:

```go
c, _ := cache.New[string, string](cache.WithCacheSize(10), cache.WithQueueKeySize(8))
```

---

## Implementation details

* Values are stored as `map[K]value[V]`, where `value` contains the expiration `time.Time` and the actual `value V`.
* Expirations are tracked in `queue map[time.Time][]K`. `times []time.Time` is a sorted slice of expiration times used to iterate over expirations efficiently.
* Several helper methods (like `addTime`, `deleteKeyFromQueue`, etc.) are marked `// not concurrent-safe` and are always called while holding internal locks.
* When `len(cache) == cacheSize`, `clean()` scans `times` for expired entries, removes expired keys, and resizes `cacheSize` to `2 * len(cache)`.

---

## Examples

```go
package main

import (
	"fmt"
	"time"

	"github.com/glekoz/cache"
)

func main() {
	c, err := cache.New[string, string](cache.WithCacheSize(10))
	if err != nil {
		panic(err)
	}

	// Add a key with 5s TTL (step = 1s, so TTL will be truncated to a multiple of 1s)
	if err := c.Add("user:1", "Ivan", 5*time.Second); err != nil {
		fmt.Println("Add failed:", err)
	}

	// Read it back
	if v, ok := c.Get("user:1"); ok {
		fmt.Println("found:", v)
	} else {
		fmt.Println("missing or expired")
	}

	// Delete the key
	c.Delete("user:1")
}
```

---

## Tests & Coverage

Run the test suite locally:

```bash
go test ./...
```

The repository currently reports **96.8%** statement coverage. [![Coverage](https://img.shields.io/badge/coverage-96.8%25-brightgreen.svg)](https://github.com/glekoz/cache)

---
