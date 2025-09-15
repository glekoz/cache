package cache

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ----------------------------------------------------------------
// 							TEST SECTION
// ----------------------------------------------------------------

type SetupSuite struct {
	suite.Suite
	cache *inMemoryCache[int, string]
}

func TestSetupSuite(t *testing.T) {
	suite.Run(t, new(SetupSuite))
}

func (s *SetupSuite) SetupSuite() {
	c, err := New[int, string]()
	if err != nil {
		s.FailNow("SetupSuite New cache failed")
	}
	s.cache = c
}

func (s *SetupSuite) TearDownTest() {
	c, err := New[int, string]()
	if err != nil {
		s.FailNow("SetupSuite New cache failed")
	}
	s.cache = c
}

// ----------------------------------------------------------------

func (s *SetupSuite) TestDefaultNew() {
	c, err := New[int, string]()
	if err != nil {
		s.FailNow("New failed")
	}
	s.cache = c
	s.Assert().Equal(5, s.cache.cacheSize)
	s.Assert().Equal(5, cap(s.cache.times))
	s.Assert().Equal(5, s.cache.queueKeySize)
}

func (s *SetupSuite) TestWithCacheSize() {
	c, err := New[int, string](WithCacheSize(10))
	if err != nil {
		s.FailNow("New failed")
	}
	s.cache = c
	s.Assert().Equal(10, s.cache.cacheSize)
	s.Assert().Equal(5, cap(s.cache.times))
	s.Assert().Equal(5, s.cache.queueKeySize)
}

func (s *SetupSuite) TestWithTimeSize() {
	c, err := New[int, string](WithTimeSize(10))
	if err != nil {
		s.FailNow("New failed")
	}
	s.cache = c
	s.Assert().Equal(5, s.cache.cacheSize)
	s.Assert().Equal(10, cap(s.cache.times))
	s.Assert().Equal(5, s.cache.queueKeySize)
}

func (s *SetupSuite) TestWithQueueKeySize() {
	c, err := New[int, string](WithQueueKeySize(10))
	if err != nil {
		s.FailNow("New failed")
	}
	s.cache = c
	s.Assert().Equal(5, s.cache.cacheSize)
	s.Assert().Equal(5, cap(s.cache.times))
	s.Assert().Equal(10, s.cache.queueKeySize)
}

func (s *SetupSuite) TestWithAllOptions() {
	c, err := New[int, string](
		WithCacheSize(10),
		WithQueueSize(11),
		WithTimeSize(12),
		WithQueueKeySize(13),
	)
	if err != nil {
		s.FailNow("New failed")
	}
	s.cache = c
	s.Assert().Equal(10, s.cache.cacheSize)
	s.Assert().Equal(12, cap(s.cache.times))
	s.Assert().Equal(13, s.cache.queueKeySize)
}

func (s *SetupSuite) TestWithNegativeOptions() {
	_, err := New[int, string](WithCacheSize(-1))
	s.Assert().NotNil(err)
	_, err = New[int, string](WithQueueSize(-1))
	s.Assert().NotNil(err)
	_, err = New[int, string](WithTimeSize(-1))
	s.Assert().NotNil(err)
	_, err = New[int, string](WithQueueKeySize(-1))
	s.Assert().NotNil(err)
}
