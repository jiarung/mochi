package maglev

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TestMaglevSuite struct {
	suite.Suite
}

func (s *TestMaglevSuite) TestMaglevNaively() {
	m, err := New([]string{})
	s.Require().NotNil(err)
	s.Require().Nil(m)

	m, err = New(
		[]string{
			"localhost:8080",
			"localhost:8081",
			"localhost:8082",
			"localhost:8083",
		},
	)
	s.Require().Nil(err)

	i := []byte("test-input")
	s.Require().Equal(m.Sum(i), m.Get(i))
}

func TestMaglev(t *testing.T) {
	suite.Run(t, new(TestMaglevSuite))
}
