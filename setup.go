package cache

import (
	"cmp"
	"errors"
	"sync"
	"time"
)

type value[V any] struct {
	time  time.Time //time.Now() + ttl; min ttl = 1s
	value V
}

type inMemoryCache[K cmp.Ordered, V any] struct {
	mu           sync.RWMutex      // no comments
	cache        map[K]value[V]    // map for storing keys and values
	cacheSize    int               // initial size of cache
	queue        map[time.Time][]K // map that's used for storing expiration time and corresponding ordered slice of keys
	queueKeySize int               // initial size of []K in queue
	times        []time.Time       // ordered slice of all expiration times (it's used to retrieve specific keys)
	step         time.Duration     // ttl step of 1 ms
}

type options struct {
	cacheSize    int // initial size of cache
	queueSize    int // initial size of queue
	timeSize     int // initial size of times
	queueKeySize int // initial size of []K in queue
}

type Option func(options *options) error

func WithCacheSize(size int) Option {
	return func(options *options) error {
		if size < 0 {
			return errors.New("cache size must be positive")
		}
		options.queueSize = size
		return nil
	}
}

func WithQueueSize(size int) Option {
	return func(options *options) error {
		if size < 0 {
			return errors.New("queue size must be positive")
		}
		options.queueSize = size
		return nil
	}
}

func WithTimeSize(size int) Option {
	return func(options *options) error {
		if size < 0 {
			return errors.New("times size must be positive")
		}
		options.timeSize = size
		return nil
	}
}

func WithQueueKeySize(size int) Option {
	return func(options *options) error {
		if size < 0 {
			return errors.New("queue keys size must be positive")
		}
		options.queueKeySize = size
		return nil
	}
}

func New[K cmp.Ordered, V any](opts ...Option) (*inMemoryCache[K, V], error) {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	if options.cacheSize == 0 {
		options.cacheSize = 5
	}
	if options.queueSize == 0 {
		options.queueSize = 5
	}
	if options.timeSize == 0 {
		options.timeSize = 5
	}
	if options.queueKeySize == 0 {
		options.queueKeySize = 5
	}

	c := &inMemoryCache[K, V]{
		cache:        make(map[K]value[V], options.cacheSize),
		queue:        make(map[time.Time][]K, options.queueSize),
		times:        make([]time.Time, 0, options.timeSize),
		cacheSize:    options.cacheSize,
		queueKeySize: options.queueKeySize,
		step:         time.Second,
	}

	return c, nil
}
