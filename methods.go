package cache

import (
	"cmp"
	"errors"
	"slices"
	"time"
)

func (c *inMemoryCache[K, V]) Add(key K, val V, ttl time.Duration) error {
	if ttl < c.step {
		return errors.New("ttl must be more than cache step")
	}
	ttl = ttl.Truncate(c.step)
	t := time.Now().Truncate(c.step)
	t = t.Add(ttl)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.deleteKeyFromQueue(key)
	c.cache[key] = value[V]{time: t, value: val}

	if _, ok := c.queue[t]; !ok {
		c.queue[t] = make([]K, 0, c.queueKeySize)
	}

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
			if i == 0 {
				index = i
			} else {
				index = i - 1
			}
			break
		}
	}
	for _, t := range expiredTimes {
		temp := make([]K, len(c.queue[t]))
		copy(temp, c.queue[t]) // slices.Delete in deleteKeyFromQueue leads to changing c.queue[t], so I need to create a full copy to save keys to delete
		for _, key := range temp {
			c.deleteKeyFromQueue(key)
			delete(c.cache, key)
		}
	}
	c.cacheSize = 2 * len(c.cache)
	temp := make([]time.Time, len(c.times[index:]), len(c.times[index:])*2)
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
	c.queue[v.time] = slices.Delete(c.queue[v.time], index, index+1)
	if len(c.queue[v.time]) == 0 {
		delete(c.queue, v.time)
		c.deleteTimeFromTimes(v.time)
	}
}

// not concurrent-safe
func (c *inMemoryCache[K, V]) deleteTimeFromTimes(t time.Time) {
	index, exists := findIndex(c.times, t, time.Time.Compare)
	if !exists {
		return
	}
	c.times = slices.Delete(c.times, index, index+1)
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
}

// not concurrent-safe
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
