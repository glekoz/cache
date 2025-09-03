package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type Data struct {
	name  string
	key   int
	value string
	ttl   time.Time
}

type CacheSuite struct {
	suite.Suite
	cache         *inMemoryCache[int, string]
	getDeleteData map[string][]Data
}

func (s *CacheSuite) SetupSuite() {
	c, err := New[int, string]()
	if err != nil {
		s.FailNow("SetupSuite New cache failed")
	}
	s.cache = c
}

func (s *CacheSuite) BeforeTest(suiteName, testName string) {
	// if suiteName == "CacheSuite" && (testName == "TestGet" || testName == "TestDelete") {
	// 	data := []Data{
	// 		{
	// 			name: "Test",
	// 		},
	// 	}
	// 	s.cache.Add(1, "first", time.Second)
	// }
}

func (s *CacheSuite) TearDownTest() {
	clear(s.cache.cache)
	clear(s.cache.queue)
	s.cache.times = nil
	s.cache.resetTicker()
}

// I NEED MORE TESTS FOR DIFFERENT ADD SITUATIONS
// (order matters)

/*
// I should learn how to use synctest package
func (s *CacheSuite) TestAdd() {
	minTTL := 5 * time.Second
	tests := []struct {
		name  string
		key   int
		value string
		ttl   time.Duration
		queue []int
		err   string
	}{
		{
			name:  "OK Test #1",
			key:   1,
			value: "first",
			ttl:   minTTL,
			queue: []int{1},
		},
		{
			name:  "OK Test #2",
			key:   10,
			value: "tenth",
			ttl:   minTTL,
			queue: []int{1, 10},
		},
		{
			name:  "Error Test",
			key:   2,
			value: "second",
			ttl:   -5 * time.Second,
			err:   "ttl must be more than cache check step",
		},
		{
			name:  "Duplicate OK Test #1 with different ttl",
			key:   1,
			value: "first",
			ttl:   10 * time.Second,
			queue: []int{1},
		},
		{
			name:  "Duplicate OK Test #1 with the same ttl",
			key:   1,
			value: "first",
			ttl:   10 * time.Second,
			queue: []int{1},
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			err := s.cache.Add(test.key, test.value, test.ttl)
			// actually I can implement error interface the way when
			// nil value will return ""
			if err != nil {
				s.Require().Equal(test.err, err.Error())
				return
			}
			v, ok := s.cache.cache[test.key]
			s.Require().True(ok)
			s.Assert().Equal(test.value, v.value)
			s.Assert().Equal(test.ttl, v.time.Sub(time.Now().Truncate(s.cache.step)))
			//s.Assert().Equal(v.time, s.cache.times[0])
			s.Assert().Equal(time.Now().Truncate(time.Second).Add(minTTL), s.cache.times[0])
			s.Assert().Equal(test.queue, s.cache.queue[v.time])
		})

	}
}
*/
// func (s *CacheSuite) TestGet() {
// 	v, ok := s.cache.Get(10)
// 	if !ok {
// 		s.FailNow("TestGet failed")
// 	}
// 	s.Assert().Equal("ten", v)
// 	_, ok = s.cache.Get(1)
// 	s.Assert().False(ok)
// 	time.Sleep(1 * time.Second)
// 	_, ok = s.cache.Get(10)
// 	s.Assert().False(ok)
// }

func (s *CacheSuite) TestAdd_SameKey_SameValue_SameTTL() {
	tm := time.Now().Truncate(s.cache.step).Add(10 * time.Second)
	test := struct {
		name  string
		key   int
		value string
		ttl   time.Duration

		expectedCache Value[string]
		expectedQueue []int
		expectedTimes []time.Time
		expectedErr   error
	}{
		name:  "Sample",
		key:   1,
		value: "first",
		ttl:   10 * time.Second,

		expectedCache: Value[string]{time: tm, value: "first"},
		expectedQueue: []int{1},
		expectedTimes: []time.Time{tm},
	}
	for range 5 {
		s.T().Run(test.name, func(t *testing.T) {
			err := s.cache.Add(test.key, test.value, test.ttl)
			s.Require().Nil(err)
			s.Assert().Equal(test.expectedCache, s.cache.cache[test.key])
			s.Assert().Equal(test.expectedQueue, s.cache.queue[tm])
			s.Assert().Equal(test.expectedTimes, s.cache.times)
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_DifferentValue_SameTTL() {
	tm := time.Now().Truncate(s.cache.step).Add(10 * time.Second)
	tests := []struct {
		name  string
		key   int
		value string
		ttl   time.Duration

		expectedCache Value[string]
		expectedQueue []int
		expectedTimes []time.Time
		expectedErr   error
	}{
		{
			name:  "First Value",
			key:   1,
			value: "first",
			ttl:   10 * time.Second,

			expectedCache: Value[string]{time: tm, value: "first"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm},
		},
		{
			name:  "Second Value",
			key:   1,
			value: "second",
			ttl:   10 * time.Second,

			expectedCache: Value[string]{time: tm, value: "second"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm},
		},
		{
			name:  "Third Value",
			key:   1,
			value: "third",
			ttl:   10 * time.Second,

			expectedCache: Value[string]{time: tm, value: "third"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm},
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			err := s.cache.Add(test.key, test.value, test.ttl)
			s.Require().Nil(err)
			s.Assert().Equal(test.expectedCache, s.cache.cache[test.key])
			s.Assert().Equal(test.expectedQueue, s.cache.queue[tm])
			s.Assert().Equal(test.expectedTimes, s.cache.times)
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_SameValue_DifferentTTL() {
	tm := time.Now().Truncate(s.cache.step)
	tests := []struct {
		name  string
		key   int
		value string
		ttl   time.Duration

		expectedCache Value[string]
		expectedQueue []int
		expectedTimes []time.Time
		expectedErr   error
	}{
		{
			name:  "First TTL",
			key:   1,
			value: "first",
			ttl:   10 * time.Second,

			expectedCache: Value[string]{time: tm.Add(10 * time.Second), value: "first"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(10 * time.Second)},
		},
		{
			name:  "Second TTL",
			key:   1,
			value: "first",
			ttl:   20 * time.Second,

			expectedCache: Value[string]{time: tm.Add(20 * time.Second), value: "first"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(20 * time.Second)},
		},
		{
			name:  "Third TTL",
			key:   1,
			value: "first",
			ttl:   30 * time.Second,

			expectedCache: Value[string]{time: tm.Add(30 * time.Second), value: "first"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(30 * time.Second)},
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			err := s.cache.Add(test.key, test.value, test.ttl)
			s.Require().Nil(err)
			s.Assert().Equal(test.expectedCache, s.cache.cache[test.key])
			s.Assert().Equal(test.expectedQueue, s.cache.queue[tm.Add(test.ttl)])
			s.Assert().Equal(test.expectedTimes, s.cache.times)
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_DifferentValue_DifferentTTL() {
	tm := time.Now().Truncate(s.cache.step)
	tests := []struct {
		name  string
		key   int
		value string
		ttl   time.Duration

		expectedCache Value[string]
		expectedQueue []int
		expectedTimes []time.Time
		expectedErr   error
	}{
		{
			name:  "First TTL",
			key:   1,
			value: "first",
			ttl:   10 * time.Second,

			expectedCache: Value[string]{time: tm.Add(10 * time.Second), value: "first"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(10 * time.Second)},
		},
		{
			name:  "Second TTL",
			key:   1,
			value: "second",
			ttl:   20 * time.Second,

			expectedCache: Value[string]{time: tm.Add(20 * time.Second), value: "second"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(20 * time.Second)},
		},
		{
			name:  "Third TTL",
			key:   1,
			value: "third",
			ttl:   30 * time.Second,

			expectedCache: Value[string]{time: tm.Add(30 * time.Second), value: "third"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(30 * time.Second)},
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			err := s.cache.Add(test.key, test.value, test.ttl)
			s.Require().Nil(err)
			s.Assert().Equal(test.expectedCache, s.cache.cache[test.key])
			s.Assert().Equal(test.expectedQueue, s.cache.queue[tm.Add(test.ttl)])
			s.Assert().Equal(test.expectedTimes, s.cache.times)
		})
	}
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheSuite))
}
