package cache

import (
	"cmp"
	"slices"
	"time"
)

func (c *inMemoryCache[K, V]) Add(key K, val V, ttl time.Duration) error {
	if ttl < c.step {
		ttl = c.step
	} else {
		ttl = ttl.Truncate(c.step)
	}
	t := time.Now().Truncate(c.step)
	t = t.Add(ttl)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = value[V]{time: t, value: val}
	if _, ok := c.queue[t]; !ok {
		c.queue[t] = make([]K, 0, c.queueKeySize)
	}
	//c.queue[t] = append(c.queue[t], key) // addKey
	c.deleteKeyFromQueue(key)
	c.addKey(key, t)
	c.addTime(t)

	if len(c.cache) == c.cacheSize {
		c.clean()
	}

	return nil
}

func (c *inMemoryCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	v, ok := c.cache[key]
	c.mu.RUnlock()
	if v.time.IsZero() {
		return v.value, false
	} else if time.Now().After(v.time) {
		var zero V
		c.Delete(key)
		return zero, false
	}
	return v.value, ok
}

func (c *inMemoryCache[K, V]) Delete(keys ...K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, key := range keys {
		c.deleteKeyFromQueue(key)
		delete(c.cache, key)
	}
}

// not concurrent-safe
func (c *inMemoryCache[K, V]) clean() {
	var index int
	expiredTimes := make([]time.Time, 0, 4)
	for i, t := range c.times {
		if time.Now().After(t) {
			expiredTimes = append(expiredTimes, t)
		} else {
			index = i
			break
		}
	}
	for _, t := range expiredTimes {
		for _, key := range c.queue[t] {
			c.deleteKeyFromQueue(key)
			delete(c.cache, key)
		}
	}
	c.cacheSize = 2 * len(c.cache)
	temp := make([]time.Time, len(c.times[index:]), 2*len(c.times[index:]))
	copy(temp, c.times[index:])
	c.times = temp
}

// not concurrent-safe
func (c *inMemoryCache[K, V]) deleteKeyFromQueue(key K) {
	v, ok := c.cache[key]
	if !ok {
		return
	}
	index, exists := findIndex(c.queue[v.time], key, cmp.Compare)
	if !exists {
		return
	}
	c.queue[v.time][index] = c.queue[v.time][len(c.queue[v.time])-1]
	c.queue[v.time] = c.queue[v.time][:len(c.queue[v.time])-1]
	if len(c.queue[v.time]) == 0 {
		delete(c.queue, v.time)
		c.deleteTimeFromTimes(v.time)
	}
}

// not concurrent-safe
func (c *inMemoryCache[K, V]) deleteTimeFromTimes(t time.Time) {
	//index, exists := c.findIndex(t)
	index, exists := findIndex(c.times, t, time.Time.Compare)
	if !exists {
		return
	}
	c.times = slices.Delete(c.times, index, index+1)
	// if index == 0 {
	// 	c.resetTicker()
	// }
}

// not concurrent-safe
func (c *inMemoryCache[K, V]) addTime(t time.Time) {
	index, exists := findIndex(c.times, t, time.Time.Compare)
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
	// if index == 0 {
	// 	c.resetTicker()
	// }
}

func (c *inMemoryCache[K, V]) addKey(key K, t time.Time) {
	index, exists := findIndex(c.queue[t], key, cmp.Compare)
	if exists {
		return
	}
	if len(c.queue[t]) == cap(c.queue[t]) {
		c.queue[t] = append(c.queue[t], key)
	} else {
		c.queue[t] = c.queue[t][:len(c.queue[t])+1]
	}
	copy(c.queue[t][index+1:], c.queue[t][index:])
	c.queue[t][index] = key
}

// not concurrent-safe
// func (c *inMemoryCache[K, V]) resetTicker() {
// 	if len(c.times) == 0 {
// 		c.ticker.Reset(time.Hour)
// 		return
// 	}
// 	c.ticker.Reset(c.times[0].Sub(time.Now().Truncate(c.step)))
// }

// returns index before which you should place new value
// 3 -> [0, 5, 9]: the function returns 1
// func (c *inMemoryCache[K, V]) findIndex(t time.Time) (index int, exists bool) {
// 	l := len(c.times)
// 	i, j := 0, l
// 	for i < j {
// 		mid := (i + j) >> 1
// 		switch c.times[mid].Compare(t) {
// 		case -1:
// 			i = mid + 1
// 		case 0:
// 			return mid, true
// 		case 1:
// 			j = mid
// 		}
// 	}
// 	return i, false
// }

func findIndex[S []T, T comparable](slice S, item T, compFunc func(x, y T) int) (index int, exists bool) {
	l := len(slice)
	i, j := 0, l
	for i < j {
		mid := (i + j) >> 1
		switch compFunc(slice[mid], item) {
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

// func isCmpOrdered(item any) bool {
// 	if _, ok := item.(int); ok{
// 		return true
// 	} else if _, ok := item.(string); ok {
// 		return true
// 	}
// 	// } else if _, ok := item.(float64); ok{ // because float can be asserted to int type
// 	// 	return true
// 	// }
// 	return false
// }

// I'm not sure whether I need following code
/*
const elemsPerTable = 896 // it's used for memory reallocating

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

// func (c *inMemoryCache[K, V]) StopGC() {
// 	c.isGC.CompareAndSwap(1, 0)
// }

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
*/
