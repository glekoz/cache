package cache

import (
	"errors"
	"time"
)

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

func WithGCInterval(interval time.Duration) Option {
	return func(options *options) error {
		if interval < 0 {
			return errors.New("GC interval must be positive")
		}
		options.gcInterval = interval
		return nil
	}
}

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
		cache:      make(map[K]V),
		queue:      make(map[time.Time][]K),
		step:       options.step,
		queueSize:  options.queueSize,
		gcInterval: options.gcInterval,
	}
	c.clean()

	return c, nil
}
