package cache

import (
	"errors"
	"maps"
	"math"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

const elemsPerTable = 896 // it's used for memory reallocating

func (c *inMemoryCache[K, V]) Add(key K, value V, ttl time.Duration) error {
	// switch any(key).(type) {
	// case int:
	// default:
	//	var k K
	// 	if key == k {
	// 		return errors.New("key must be non-zero value")
	// 	}
	// }
	if ttl < c.step {
		return errors.New("ttl must be more than cache check step")
	}
	t := time.Now().Truncate(c.step)
	ttl = ttl.Truncate(c.step)
	t = t.Add(ttl)
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deleteKeyFromQueue(key) // too heavy for active usage

	c.cache[key] = Value[V]{time: t, value: value}

	if _, ok := c.queue[t]; !ok {
		c.queue[t] = make([]K, 0, c.queueSize)
	}
	c.queue[t] = append(c.queue[t], key)
	c.addTime(t)

	return nil
}

func (c *inMemoryCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	v, ok := c.cache[key]
	c.mu.RUnlock()
	if time.Now().After(v.time) {
		var zero V
		return zero, false
	}
	return v.value, ok
}

func (c *inMemoryCache[K, V]) Delete(keys ...K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, key := range keys {
		// the code is correct, but it takes too much CPU time
		// to find and delete something that will eventually be deleted
		c.deleteKeyFromQueue(key)
		delete(c.cache, key)
	}
}

func (c *inMemoryCache[K, V]) StartGC() {
	if !c.isGC.CompareAndSwap(0, 1) {
		return
	}
	go func() {
		t := time.NewTicker(c.gcInterval)
		for {
			select {
			case <-t.C:
				switch c.isGC.Load() == 1 {
				case true:
					c.realloc()
				case false:
					return
				}
			case <-c.closeChan:
				return
			}
		}
	}()
}

/*
	func (c *inMemoryCache[K, V]) StopGC() {
		c.isGC.CompareAndSwap(1, 0)
	}
*/
func (c *inMemoryCache[K, V]) Close() {
	a := new(sync.Once) // move to struct's fields
	a.Do(func() { close(c.closeChan) })

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
	elemCount := mapPtr.used // or just len(c.cache)
	c.mu.Lock()
	defer c.mu.Unlock()
	if int(math.Log2(float64(tableCount*elemsPerTable/int(elemCount)))) > 1 {
		mapCopy := make(map[K]Value[V], elemCount)
		maps.Copy(mapCopy, c.cache)
		c.cache = mapCopy
	}
	timesCopy := make([]time.Time, len(c.times), cap(c.times))
	copy(timesCopy, c.times)
	c.times = timesCopy
}

func (c *inMemoryCache[K, V]) clean() {
	go func() {
		for {
			select {
			case <-c.closeChan:
				return
			case <-c.ticker.C:
				c.mu.Lock()
				t := time.Now().Truncate(c.step)
				keys, ok := c.queue[t]
				if !ok {
					c.mu.Unlock()
					break // or continue
				}
				delete(c.queue, t)
				for _, key := range keys {
					delete(c.cache, key)
				}
				if len(c.times) != 0 {
					c.times = c.times[1:]
				}
				c.resetTicker()
				c.mu.Unlock()
			}
		}
	}()
}

// not concurrent-safe
// I should move this heavy-lifting method to my GC
func (c *inMemoryCache[K, V]) deleteKeyFromQueue(key K) {
	v, ok := c.cache[key]
	if !ok {
		return
	}
	keySlice := c.queue[v.time]
	for i, k := range keySlice {
		if k == key {
			keySlice[i] = keySlice[len(keySlice)-1]
			c.queue[v.time] = keySlice[:len(keySlice)-1]
			if len(c.queue[v.time]) == 0 {
				delete(c.queue, v.time)
				c.deleteTimeFromTimes(v.time)
			}
			//keySlice=append(keySlice[:i], keySlice[i+1:]...)
			break
		}
	}
}

// not concurrent-safe
func (c *inMemoryCache[K, V]) deleteTimeFromTimes(t time.Time) {
	index, exists := c.findIndex(t)
	if !exists {
		return
	}
	c.times[index] = c.times[len(c.times)-1]
	c.times = c.times[:len(c.times)-1]
	if index == 0 {
		c.resetTicker()
	}
}

// not concurrent-safe
func (c *inMemoryCache[K, V]) addTime(t time.Time) {
	index, exists := c.findIndex(t)
	if exists {
		return
	}
	if len(c.times) == cap(c.times) {
		c.times = append(c.times, time.Time{})
	} else {
		c.times = c.times[:len(c.times)+1]
	}
	copy(c.times[index+1:], c.times[index:])
	c.times[index] = t
	if index == 0 {
		c.resetTicker()
	}
}

// not concurrent-safe
func (c *inMemoryCache[K, V]) resetTicker() {
	if len(c.times) == 0 {
		c.ticker.Reset(time.Hour)
		return
	}
	c.ticker.Reset(c.times[0].Sub(time.Now().Truncate(c.step)))
}

// returns index before which you should place new value
// 3 -> [0, 5, 9]: the function returns 1
func (c *inMemoryCache[K, V]) findIndex(t time.Time) (index int, exists bool) {
	l := len(c.times)
	i, j := 0, l
	for i < j {
		mid := (i + j) >> 1
		switch c.times[mid].Compare(t) {
		case -1:
			i = mid + 1
		case 0:
			return mid, true
		case 1:
			j = mid
		}
	}
	return i, false
}

// Test block
/*
func findIndex(ts []time.Time, t time.Time) (index int, exists bool) {
	l := len(ts)
	i, j := 0, l
	for i < j {
		mid := (i + j) >> 1
		switch ts[mid].Compare(t) {
		case -1:
			i = mid + 1
		case 0:
			return mid, true
		case 1:
			j = mid
		}
	}
	return i, false
}

func addTime(ts []time.Time, t time.Time) []time.Time {
	index, exists := findIndex(ts, t)
	if exists {
		return ts
	}
	if len(ts) == cap(ts) {
		ts = append(ts, time.Time{})
	}
	copy(ts[index+1:], ts[index:])
	ts[index] = t
	return ts
}
*/
