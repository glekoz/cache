package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/glekoz/cache/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		step       time.Duration
		queueSize  int
		gcInterval time.Duration

		expected inMemoryCache[string, int]
		err      error
	}{
		{
			name: "Default",
			expected: inMemoryCache[string, int]{
				step:       time.Second,
				queueSize:  5,
				gcInterval: time.Minute,
			},
		},
		{
			name: "WithStep_Correct",
			step: time.Minute,
			expected: inMemoryCache[string, int]{
				step:       time.Minute,
				queueSize:  5,
				gcInterval: time.Minute,
			},
		},
		{
			name:      "WithQueueSize_Correct",
			queueSize: 10,
			expected: inMemoryCache[string, int]{
				step:       time.Second,
				queueSize:  10,
				gcInterval: time.Minute,
			},
		},
		{
			name:       "WithGCInterval_Correct",
			gcInterval: time.Hour,
			expected: inMemoryCache[string, int]{
				step:       time.Second,
				queueSize:  5,
				gcInterval: time.Hour,
			},
		},
		{
			name: "WithStep_Wrong",
			step: -1 * time.Minute,
			err:  errors.New("step must be positive"),
		},
		{
			name:      "WithQueueSize_Wrong",
			queueSize: -10,
			err:       errors.New("queue size must be positive"),
		},
		{
			name:       "WithGCInterval_Wrong",
			gcInterval: -1 * time.Hour,
			err:        errors.New("GC interval must be positive"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := New[string, int](WithStep(tc.step), WithQueueSize(tc.queueSize), WithGCInterval(tc.gcInterval))
			if err != nil {
				//t.Fatal()
				assert.Equal(t, err.Error(), tc.err.Error())
				return
			}
			assert.Equal(t, c.step, tc.expected.step)
			assert.Equal(t, c.queueSize, tc.expected.queueSize)
			assert.Equal(t, c.gcInterval, tc.expected.gcInterval)
		})
	}
}
