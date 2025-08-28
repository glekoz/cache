package cache

import (
	"errors"
	"maps"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const elemsPerTable = 896 // it's used for memory reallocating

type inMemoryCache[K comparable, V any] struct {
	mu         sync.RWMutex
	cache      map[K]V           // map for storing keys and values
	queue      map[time.Time][]K // map that's used for storing keys and their expiration time
	step       time.Duration     // minimal time clock to check whether keys are expired
	queueSize  int               // size of []K - depends on use case
	isGC       atomic.Int32      // it's used to start goroutine that reallocates the memory
	gcInterval time.Duration     // interval to check whether memory needs to be reallocated

}

func (c *inMemoryCache[K, V]) Add(key K, value V, ttl time.Duration) error {
	if ttl < c.step {
		return errors.New("ttl must be more than cache check step")
	}
	t := time.Now().Truncate(c.step)
	ttl = ttl.Truncate(c.step)
	t = t.Add(ttl)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = value

	keys, ok := c.queue[t]
	if !ok {
		c.queue[t] = make([]K, 0, c.queueSize)
	}
	c.queue[t] = append(keys, key)
	return nil
}

func (c *inMemoryCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	value, ok := c.cache[key]
	c.mu.RUnlock()
	return value, ok
}

func (c *inMemoryCache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.cache, key)
	c.mu.Unlock()
}

func (c *inMemoryCache[K, V]) StartGC() {
	if !c.isGC.CompareAndSwap(0, 1) {
		return
	}
	go func() {
		t := time.NewTicker(c.gcInterval)
		cond := atomic.Int32{}
		(&cond).Store(1)
		for {
			<-t.C
			switch c.isGC == cond {
			case true:
				c.realloc()
			case false:
				return
			}
		}
	}()
}

func (c *inMemoryCache[K, V]) StopGC() {
	c.isGC.CompareAndSwap(1, 0)
}

func (c *inMemoryCache[K, V]) realloc() {
	mapValue := reflect.ValueOf(c.cache)
	mapPointer := unsafe.Pointer(mapValue.Pointer())
	type Map struct {
		used              uint64
		seed              uintptr
		dirPtr            unsafe.Pointer
		dirLen            int
		globalDepth       uint8
		globalShift       uint8
		writing           uint8
		tombstonePossible bool
		clearSeq          uint64
	}

	mapPtr := (*Map)(mapPointer)
	tableCount := mapPtr.dirLen
	elemCount := mapPtr.used // or len(c.cache)
	if int(math.Log2(float64(tableCount*elemsPerTable/int(elemCount)))) > 1 {
		mapCopy := make(map[K]V, elemCount)
		maps.Copy(mapCopy, c.cache)
		c.cache = mapCopy
	}
}
