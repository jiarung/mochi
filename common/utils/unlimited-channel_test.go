package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type UnlimitedChannelTestSuite struct {
	suite.Suite
}

func (s *UnlimitedChannelTestSuite) TestIO() {
	c := NewUnlimitedChannel()
	defer c.Close()
	c.In() <- 1
	c.In() <- 2
	s.Require().Equal(1, <-c.Out())
	s.Require().Equal(2, <-c.Out())
}

func (s *UnlimitedChannelTestSuite) TestClose() {
	c := NewUnlimitedChannel()

	// Test channel Out() and Done() should block except In().
	select {
	case <-c.Out():
		s.Require().True(false)
	case <-c.Done():
		s.Require().True(false)
	case c.In() <- 1:
	}

	// Flush.
	<-c.Out()

	c.Close()

	// Test channel In() and Out() should block except Done().
	select {
	case c.In() <- 1:
		s.Require().True(false)
	case <-c.Out():
		s.Require().True(false)
	case <-c.Done():
	}
}

func TestUnlimitedChannel(t *testing.T) {
	suite.Run(t, new(UnlimitedChannelTestSuite))
}
