package cache

import (
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
			name: "WithStep",
			step: time.Minute,
			expected: inMemoryCache[string, int]{
				step:       time.Minute,
				queueSize:  5,
				gcInterval: time.Minute,
			},
		},
		{
			name:      "WithQueueSize",
			queueSize: 10,
			expected: inMemoryCache[string, int]{
				step:       time.Second,
				queueSize:  10,
				gcInterval: time.Minute,
			},
		},
		{
			name:       "WithGCInterval",
			gcInterval: time.Hour,
			expected: inMemoryCache[string, int]{
				step:       time.Second,
				queueSize:  5,
				gcInterval: time.Hour,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := New[string, int](WithStep(tc.step), WithQueueSize(tc.queueSize), WithGCInterval(tc.gcInterval))
			if err != nil {
				t.Fatal()
			}
			assert.Equal(t, c.step, tc.expected.step)
			assert.Equal(t, c.queueSize, tc.expected.queueSize)
			assert.Equal(t, c.gcInterval, tc.expected.gcInterval)
		})
	}
}
