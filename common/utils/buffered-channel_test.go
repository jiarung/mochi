package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BufferedChannelTestSuite struct {
	suite.Suite
}

func (s *BufferedChannelTestSuite) TestFull() {
	c := NewBufferedChannel(1)
	defer c.Close()

	c.In() <- 1
	c.In() <- 2
	s.Require().Equal(ErrChannelFull, <-c.Err())
	s.Require().Equal(1, <-c.Out())

	select {
	case <-c.Out():
		s.Require().True(false)
	default:
	}
}

func TestBufferedChannel(t *testing.T) {
	suite.Run(t, new(BufferedChannelTestSuite))
}
