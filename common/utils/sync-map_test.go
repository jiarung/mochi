package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SyncMapTestSuite struct {
	suite.Suite
}

func (s *SyncMapTestSuite) TestEmpty() {
	m := SyncMap{}

	s.Require().True(m.Empty())
	m.Store(1, nil)
	s.Require().False(m.Empty())
	m.Store(2, nil)
	s.Require().False(m.Empty())
	m.Delete(2)
	s.Require().False(m.Empty())
	m.Store(1, nil)
	s.Require().False(m.Empty())
	m.Delete(1)
	s.Require().True(m.Empty())
}

func TestSyncMap(t *testing.T) {
	suite.Run(t, new(SyncMapTestSuite))
}
