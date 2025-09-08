package cache

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// optionally add new field - use bit (boolean) - to implement clock algorithm
type Value[V any] struct {
	time  time.Time
	value V
}

type inMemoryCache[K comparable, V any] struct {
	mu        sync.RWMutex
	cache     map[K]Value[V]    // map for storing keys and values
	queue     map[time.Time][]K // map that's used for storing keys and their expiration time
	step      time.Duration     // minimal time clock to check whether keys are expired
	queueSize int               // size of []K - depends on use case
	//isGC       atomic.Int32      // it's used to start goroutine that reallocates the memory
	//gcInterval time.Duration     // interval to check whether memory needs to be reallocated
	closeChan chan struct{}
	closed    atomic.Bool
	times     []time.Time
	ticker    *time.Ticker
}

type options struct {
	step       time.Duration
	queueSize  int
	gcInterval time.Duration
}

type Option func(options *options) error

func WithStep(step time.Duration) Option {
	return func(options *options) error {
		if step < 0 {
			return errors.New("step time must be positive")
		}
		options.step = step
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

// func WithGCInterval(interval time.Duration) Option {
// 	return func(options *options) error {
// 		if interval < 0 {
// 			return errors.New("GC interval must be positive")
// 		}
// 		options.gcInterval = interval
// 		return nil
// 	}
// }

func New[K comparable, V any](opts ...Option) (*inMemoryCache[K, V], error) {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	if options.step == 0 {
		options.step = time.Second
	}
	if options.queueSize == 0 {
		options.queueSize = 5
	}
	if options.gcInterval == 0 {
		options.gcInterval = time.Minute
	}

	c := &inMemoryCache[K, V]{
		cache:     make(map[K]Value[V]),
		queue:     make(map[time.Time][]K),
		step:      options.step,
		queueSize: options.queueSize,
		//gcInterval: options.gcInterval,
		closeChan: make(chan struct{}),
		times:     make([]time.Time, 0, 10),
		ticker:    time.NewTicker(time.Hour),
	}
	c.clean()

	return c, nil
}
