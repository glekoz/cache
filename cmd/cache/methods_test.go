package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ----------------------------------------------------------------
// 							TEST SECTION
// ----------------------------------------------------------------

type CacheSuite struct {
	suite.Suite
	cache *inMemoryCache[int, string]
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheSuite))
}

func (s *CacheSuite) SetupSuite() {
	c, err := New[int, string]()
	if err != nil {
		s.FailNow("SetupSuite New cache failed")
	}
	s.cache = c
}

func (s *CacheSuite) TearDownTest() {
	clear(s.cache.cache)
	clear(s.cache.queue)
	s.cache.times = s.cache.times[:0]
}

// ----------------------------------------------------------------

type TestData struct {
	name  string
	key   int
	value string
	ttl   time.Duration

	expectedCache value[string]
	expectedQueue []int
	expectedTimes []time.Time
	expectedErr   string
}

func (s *CacheSuite) HappyTestAddFunc(
	tm time.Time,
	test TestData,
) {
	s.T().Helper()
	err := s.cache.Add(test.key, test.value, test.ttl)
	s.Require().Nil(err)
	s.Assert().Equal(test.expectedCache, s.cache.cache[test.key])
	s.Assert().Equal(test.expectedQueue, s.cache.queue[tm.Add(test.ttl)])
	s.Assert().Equal(test.expectedTimes, s.cache.times)
}

// ----------------------------------------------------------------

func (s *CacheSuite) TestAdd_SameKey_SameValue_SameTTL() {
	tm := time.Now().Truncate(s.cache.step)
	n := 5
	td := make([]TestData, 0, n)
	for i := range n {
		td = append(td, TestData{
			name:  fmt.Sprintf("Same #%d", i),
			key:   1,
			value: "v",
			ttl:   5 * time.Second,

			expectedCache: value[string]{time: tm.Add(5 * time.Second), value: "v"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(5 * time.Second)},
		})
	}
	for _, test := range td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
			s.Assert().Equal(1, len(s.cache.cache))
			s.Assert().Equal(1, len(s.cache.queue))
			s.Assert().Equal(1, len(s.cache.times))
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_DifferentValue_SameTTL() {
	tm := time.Now().Truncate(s.cache.step)
	n := 5
	td := make([]TestData, 0, n)
	for i := range n {
		td = append(td, TestData{
			name:  fmt.Sprintf("Same #%d", i),
			key:   1,
			value: fmt.Sprintf("Same Value #%d", i),
			ttl:   5 * time.Second,

			expectedCache: value[string]{time: tm.Add(5 * time.Second), value: fmt.Sprintf("Same Value #%d", i)},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(5 * time.Second)},
		})
	}
	for _, test := range td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
			s.Assert().Equal(1, len(s.cache.cache))
			s.Assert().Equal(1, len(s.cache.queue))
			s.Assert().Equal(1, len(s.cache.times))
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_SameValue_DifferentTTL() {
	tm := time.Now().Truncate(s.cache.step)
	n := 5
	td := make([]TestData, 0, n)

	for i := range n {
		ttl := 10 * time.Second * time.Duration(i+1)
		td = append(td, TestData{
			name:  fmt.Sprintf("TTL #%d", i),
			key:   1,
			value: "v",
			ttl:   ttl,

			expectedCache: value[string]{time: tm.Add(ttl), value: "v"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(ttl)},
		})
	}

	for _, test := range td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
			s.Assert().Equal(1, len(s.cache.cache))
			s.Assert().Equal(1, len(s.cache.queue))
			s.Assert().Equal(1, len(s.cache.times))
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_DifferentValue_DifferentTTL() {
	tm := time.Now().Truncate(s.cache.step)
	n := 5
	td := make([]TestData, 0, n)

	for i := range n {
		ttl := 10 * time.Second * time.Duration(i+1)
		td = append(td, TestData{
			name:  fmt.Sprintf("Value and TTL #%d", i),
			key:   1,
			value: fmt.Sprintf("value #%d", i),
			ttl:   ttl,

			expectedCache: value[string]{time: tm.Add(ttl), value: fmt.Sprintf("value #%d", i)},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(ttl)},
		})
	}

	for _, test := range td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
			s.Assert().Equal(1, len(s.cache.cache))
			s.Assert().Equal(1, len(s.cache.queue))
			s.Assert().Equal(1, len(s.cache.times))
		})
	}
}

func (s *CacheSuite) TestAdd_DifferentKey_AnyValue_SameTTL() {
	tm := time.Now().Truncate(s.cache.step)
	n := 5
	td := make([]TestData, 0, n)
	expque := make([]int, 0, n)

	for i := range n {
		expque = append(expque, i)
		td = append(td, TestData{
			name:  fmt.Sprintf("Key and TTL #%d", i),
			key:   i,
			value: "v",
			ttl:   10 * time.Second,

			expectedCache: value[string]{time: tm.Add(10 * time.Second), value: "v"},
			expectedQueue: expque,
			expectedTimes: []time.Time{tm.Add(10 * time.Second)},
		})
	}
	for i, test := range td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
			s.Assert().Equal(i+1, len(s.cache.cache))
			s.Assert().Equal(1, len(s.cache.queue))
			s.Assert().Equal(1, len(s.cache.times))
		})
	}
}

func (s *CacheSuite) TestAdd_DifferentKey_AnyValue_DifferentTTL() {
	tm := time.Now().Truncate(s.cache.step)
	n := 5
	td := make([]TestData, 0, n)
	exptim := make([]time.Time, 0, n)

	for i := range n {
		ttl := 10 * time.Second * time.Duration(i+1)
		exptim = append(exptim, tm.Add(ttl))

		td = append(td, TestData{
			name:  fmt.Sprintf("TTL #%d", i),
			key:   i,
			value: "v",
			ttl:   ttl,

			expectedCache: value[string]{time: tm.Add(ttl), value: "v"},
			expectedQueue: []int{i},
			expectedTimes: exptim,
		})
	}
	for i, test := range td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
			s.Assert().Equal(i+1, len(s.cache.cache))
			s.Assert().Equal(i+1, len(s.cache.queue))
			s.Assert().Equal(i+1, len(s.cache.times))
		})
	}
}

func (s *CacheSuite) TestAdd_Parallel() {
	var wg sync.WaitGroup
	var n int

	wg.Add(n)
	for j := range n {
		go func() {
			defer wg.Done()
			for i := range n {
				s.cache.Add((j*10 + i), "v", time.Second*5+time.Second*time.Duration(j*10+i))
			}
		}()
	}
	wg.Wait()
	s.Assert().Equal(n*n, len(s.cache.cache))
	s.Assert().Equal(n*n, len(s.cache.queue))
	s.Assert().Equal(n*n, len(s.cache.times))
}

func (s *CacheSuite) TestAdd_Err() {
	n := 5
	td := make([]TestData, 0, n)
	for i := range n {
		ttl := 10 * time.Second * time.Duration(-i)

		td = append(td, TestData{
			name:  fmt.Sprintf("TTL #%d", i),
			key:   i,
			value: "v",
			ttl:   ttl,

			expectedCache: value[string]{},
			expectedQueue: []int{},
			expectedTimes: []time.Time{},
			expectedErr:   "ttl must be more than cache step",
		})
	}
	for _, test := range td {
		s.T().Run(test.name, func(t *testing.T) {
			err := s.cache.Add(test.key, test.value, test.ttl)
			s.Assert().Equal(test.expectedErr, err.Error())
			s.Assert().Equal(0, len(s.cache.cache))
			s.Assert().Equal(0, len(s.cache.queue))
			s.Assert().Equal(0, len(s.cache.times))
		})
	}
}

func (s *CacheSuite) TestGet() {
	if testing.Short() {
		s.T().Skip("Skipping long-running test in short mode")
	}
	n := 10
	expected := make([]string, 10)
	for i := range n {
		expected[i] = fmt.Sprintf("Value #%d", i)
		s.cache.Add(i, expected[i], time.Second*time.Duration(i+1))
	}
	for i := range n {
		s.T().Run(fmt.Sprintf("Get Test #%d", i), func(t *testing.T) {
			value, ok := s.cache.Get(i)
			s.Require().True(ok)
			s.Assert().Equal(expected[i], value)
		})
	}
	time.Sleep(11 * time.Second)
	s.Assert().Equal(n, len(s.cache.cache))
	s.Assert().Equal(n, len(s.cache.queue))
	s.Assert().Equal(n, len(s.cache.times))
	for i := range n {
		value, ok := s.cache.Get(i)
		s.Require().False(ok)
		s.Assert().Equal("", value)
	}
	s.Assert().Equal(0, len(s.cache.cache))
	s.Assert().Equal(0, len(s.cache.queue))
	s.Assert().Equal(0, len(s.cache.times))
}

func (s *CacheSuite) TestDelete() {
	n := 10
	expected := make([]string, 10)
	for i := range n {
		expected[i] = fmt.Sprintf("Value #%d", i)
		s.cache.Add(i, expected[i], time.Second*time.Duration(i+1))
	}
	for i := range n {
		s.cache.Delete(i)
		s.Assert().Equal(n-i-1, len(s.cache.cache), "s.cache.cache")
		s.Assert().Equal(n-i-1, len(s.cache.queue), "s.cache.queue")
		s.Assert().Equal(n-i-1, len(s.cache.times), "s.cache.times")
	}
}

// ----------------------------------------------------------------
// 						BENCHMARK SECTION
// ----------------------------------------------------------------

func BenchmarkAdd(b *testing.B) {
	c, _ := New[int, string]()
	for b.Loop() {
		c.Add(b.N, "v", 5*time.Second)
	}
}

func BenchmarkGet(b *testing.B) {
	c, _ := New[int, string]()
	for i := range 10 {
		c.Add(i, "v", 5*time.Second)
	}
	for b.Loop() {
		c.Get(b.N % 10)
	}
}

func BenchmarkComplete(b *testing.B) {
	c, _ := New[int, string]()
	for b.Loop() {
		c.Add(b.N, "v", 5*time.Second)
		c.Get(b.N)
		c.Delete(b.N)
	}
}
