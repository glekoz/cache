package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CacheSuite struct {
	suite.Suite
	cache *inMemoryCache[int, string]
}

func (s *CacheSuite) SetupSuite() {
	c, err := New[int, string]()
	if err != nil {
		s.FailNow("SetupSuite New cache failed")
	}
	s.cache = c
}

// I should learn how to use synctest package
func (s *CacheSuite) TestAdd() {
	var key = 1
	var value = "first"
	var ttl = 5 * time.Second
	s.Assert().Nil(s.cache.Add(key, value, ttl))
	v, ok := s.cache.cache[key]
	if !ok {
		s.FailNow("no value added")
	}
	s.Assert().Equal(value, v.value)
	s.Assert().Equal(ttl, v.time.Sub(time.Now().Truncate(s.cache.step)))
	s.Assert().Equal(v.time, s.cache.times[0])
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheSuite))
}
