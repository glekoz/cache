package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TestData struct {
	name  string
	key   int
	value string
	ttl   time.Duration

	expectedCache Value[string]
	expectedQueue []int
	expectedTimes []time.Time
	expectedErr   error
}

type CacheSuite struct {
	suite.Suite
	cache *inMemoryCache[int, string]
	td    []TestData
}

func (s *CacheSuite) HappyTestAddFunc(
	tm time.Time,
	test TestData,
) {
	err := s.cache.Add(test.key, test.value, test.ttl)
	s.Require().Nil(err)
	s.Assert().Equal(test.expectedCache, s.cache.cache[test.key])
	s.Assert().Equal(test.expectedQueue, s.cache.queue[tm.Add(test.ttl)])
	s.Assert().Equal(test.expectedTimes, s.cache.times)
}

func (s *CacheSuite) SetupSuite() {
	c, err := New[int, string]()
	if err != nil {
		s.FailNow("SetupSuite New cache failed")
	}
	s.cache = c
}

// func (s *CacheSuite) BeforeTest(suiteName, testName string) {
// if suiteName == "CacheSuite" && (testName == "TestGet" || testName == "TestDelete") {
// 	data := []Data{
// 		{
// 			name: "Test",
// 		},
// 	}
// 	s.cache.Add(1, "first", time.Second)
// }
// }

func (s *CacheSuite) TearDownTest() {
	clear(s.cache.cache)
	clear(s.cache.queue)
	s.cache.times = s.cache.times[:0]
	s.cache.resetTicker()
	s.td = s.td[:0]
}

// I NEED MORE TESTS FOR DIFFERENT ADD SITUATIONS
// (order matters)
/*
// I should learn how to use synctest package
// func (s *CacheSuite) TestAdd() {
// 	minTTL := 5 * time.Second
// 	tests := []struct {
// 		name  string
// 		key   int
// 		value string
// 		ttl   time.Duration
// 		queue []int
// 		err   string
// 	}{
// 		{
// 			name:  "OK Test #1",
// 			key:   1,
// 			value: "first",
// 			ttl:   minTTL,
// 			queue: []int{1},
// 		},
// 		{
// 			name:  "OK Test #2",
// 			key:   10,
// 			value: "tenth",
// 			ttl:   minTTL,
// 			queue: []int{1, 10},
// 		},
// 		{
// 			name:  "Error Test",
// 			key:   2,
// 			value: "second",
// 			ttl:   -5 * time.Second,
// 			err:   "ttl must be more than cache check step",
// 		},
// 		{
// 			name:  "Duplicate OK Test #1 with different ttl",
// 			key:   1,
// 			value: "first",
// 			ttl:   10 * time.Second,
// 			queue: []int{1},
// 		},
// 		{
// 			name:  "Duplicate OK Test #1 with the same ttl",
// 			key:   1,
// 			value: "first",
// 			ttl:   10 * time.Second,
// 			queue: []int{1},
// 		},
// 	}
// 	for _, test := range tests {
// 		s.T().Run(test.name, func(t *testing.T) {
// 			err := s.cache.Add(test.key, test.value, test.ttl)
// 			// actually I can implement error interface the way when
// 			// nil value will return ""
// 			if err != nil {
// 				s.Require().Equal(test.err, err.Error())
// 				return
// 			}
// 			v, ok := s.cache.cache[test.key]
// 			s.Require().True(ok)
// 			s.Assert().Equal(test.value, v.value)
// 			s.Assert().Equal(test.ttl, v.time.Sub(time.Now().Truncate(s.cache.step)))
// 			//s.Assert().Equal(v.time, s.cache.times[0])
// 			s.Assert().Equal(time.Now().Truncate(time.Second).Add(minTTL), s.cache.times[0])
// 			s.Assert().Equal(test.queue, s.cache.queue[v.time])
// 		})

// 	}
// }
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

//
//

func (s *CacheSuite) TestAdd_SameKey_SameValue_SameTTL() {
	tm := time.Now().Truncate(s.cache.step)
	for range 5 {
		s.td = append(s.td, TestData{
			name:  "All The Same",
			key:   1,
			value: "1",
			ttl:   10 * time.Second,

			expectedCache: Value[string]{time: tm.Add(10 * time.Second), value: "1"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(10 * time.Second)},
		})
	}
	for _, test := range s.td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_DifferentValue_SameTTL() {
	tm := time.Now().Truncate(s.cache.step)
	for i := range 5 {
		s.td = append(s.td, TestData{
			name:  fmt.Sprintf("Value #%d", i),
			key:   1,
			value: fmt.Sprintf("Value #%d", i),
			ttl:   10 * time.Second,

			expectedCache: Value[string]{time: tm.Add(10 * time.Second), value: fmt.Sprintf("Value #%d", i)},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(10 * time.Second)},
		})
	}
	for _, test := range s.td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_SameValue_DifferentTTL() {
	tm := time.Now().Truncate(s.cache.step)

	for i := range 5 {
		ttl := 10 * time.Second * time.Duration(i+1)
		s.td = append(s.td, TestData{
			name:  fmt.Sprintf("TTL #%d", i),
			key:   1,
			value: "v",
			ttl:   ttl,

			expectedCache: Value[string]{time: tm.Add(ttl), value: "v"},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(ttl)},
		})
	}

	for _, test := range s.td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
		})
	}
}

func (s *CacheSuite) TestAdd_SameKey_DifferentValue_DifferentTTL() {
	tm := time.Now().Truncate(s.cache.step)

	for i := range 5 {
		ttl := 10 * time.Second * time.Duration(i+1)
		s.td = append(s.td, TestData{
			name:  fmt.Sprintf("Value and TTL #%d", i),
			key:   1,
			value: fmt.Sprintf("value #%d", i),
			ttl:   ttl,

			expectedCache: Value[string]{time: tm.Add(ttl), value: fmt.Sprintf("value #%d", i)},
			expectedQueue: []int{1},
			expectedTimes: []time.Time{tm.Add(ttl)},
		})
	}

	for _, test := range s.td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
		})
	}
}

func (s *CacheSuite) TestAdd_DifferentKey_AnyValue_SameTTL() {
	tm := time.Now().Truncate(s.cache.step)
	expque := make([]int, 0, 5)

	for i := range 5 {
		expque = append(expque, i)
		s.td = append(s.td, TestData{
			name:  fmt.Sprintf("Key and TTL #%d", i),
			key:   i,
			value: "v",
			ttl:   10 * time.Second,

			expectedCache: Value[string]{time: tm.Add(10 * time.Second), value: "v"},
			expectedQueue: expque,
			expectedTimes: []time.Time{tm.Add(10 * time.Second)},
		})
	}
	for _, test := range s.td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
		})
	}
}

func (s *CacheSuite) TestAdd_DifferentKey_AnyValue_DifferentTTL() {
	tm := time.Now().Truncate(s.cache.step)
	exptim := make([]time.Time, 0, 5)

	for i := range 5 {
		ttl := 10 * time.Second * time.Duration(i+1)
		exptim = append(exptim, tm.Add(ttl))

		s.td = append(s.td, TestData{
			name:  fmt.Sprintf("TTL #%d", i),
			key:   i,
			value: "v",
			ttl:   ttl,

			expectedCache: Value[string]{time: tm.Add(ttl), value: "v"},
			expectedQueue: []int{i},
			expectedTimes: exptim,
		})
	}
	for _, test := range s.td {
		s.T().Run(test.name, func(t *testing.T) {
			s.HappyTestAddFunc(tm, test)
		})
	}
}
func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheSuite))
}
