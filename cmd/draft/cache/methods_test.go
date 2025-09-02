package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// type Data struct {
// name  string
// key   int
// value string
// ttl   time.Time
// }

type CacheSuite struct {
	suite.Suite
	cache *inMemoryCache[int, string]
	//data  map[string][]Data
}

func (s *CacheSuite) SetupSuite() {
	c, err := New[int, string]()
	if err != nil {
		s.FailNow("SetupSuite New cache failed")
	}
	s.cache = c
}

func (s *CacheSuite) BeforeTest(suiteName, testName string) {
	if suiteName == "CacheSuite" && testName == "TestGet" {
		s.T().Log("BeforeTest: TestGet")
		s.cache.Add(10, "ten", time.Second)
	}
}

func (s *CacheSuite) TearDownTest() {
	clear(s.cache.cache)
	clear(s.cache.queue)
	s.cache.times = nil
	s.cache.resetTicker()
}

// I should learn how to use synctest package
func (s *CacheSuite) TestAdd() {
	tests := []struct {
		name  string
		key   int
		value string
		ttl   time.Duration
		err   string
	}{
		{
			name:  "OK Test",
			key:   1,
			value: "first",
			ttl:   5 * time.Second,
			err:   "",
		},
		{
			name:  "Error Test",
			key:   2,
			value: "second",
			ttl:   -5 * time.Second,
			err:   "ttl must be more than cache check step",
		},
		{
			name:  "Duplicate OK test",
			key:   1,
			value: "first",
			ttl:   10 * time.Second,
			err:   "",
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			err := s.cache.Add(test.key, test.value, test.ttl)
			if err != nil {
				s.Require().Equal(test.err, err.Error())
				return
			}
			v, ok := s.cache.cache[test.key]
			s.Require().True(ok)
			s.Assert().Equal(test.value, v.value)
			s.Assert().Equal(test.ttl, v.time.Sub(time.Now().Truncate(s.cache.step)))
			s.Assert().Equal(v.time, s.cache.times[0])
			s.Assert().Equal(s.cache.queue[v.time], []int{test.key})
		})

	}
}

func (s *CacheSuite) TestGet() {
	v, ok := s.cache.Get(10)
	if !ok {
		s.FailNow("TestGet failed")
	}
	s.Assert().Equal("ten", v)
	_, ok = s.cache.Get(1)
	s.Assert().False(ok)
	time.Sleep(1 * time.Second)
	_, ok = s.cache.Get(10)
	s.Assert().False(ok)
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheSuite))
}
