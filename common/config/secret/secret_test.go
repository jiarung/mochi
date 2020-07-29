package secret

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

var mockGetter = func(key string) string {
	if key == "panic" {
		panic("GG")
	}
	return key
}

// SecretSuite tests filter
type SecretSuite struct {
	suite.Suite
}

func (s *SecretSuite) TestStaticGet() {
	getters = []Getter{staticGetter}

	{
		v := Get("SIG_KEY")
		fmt.Println(v)
	}

}

func (s *SecretSuite) TestGet() {
	getters = []Getter{mockGetter, staticGetter}

	// get from the first getter
	val := Get("xxx")
	s.Require().Equal("xxx", val)

	// all panic
	s.Require().Panics(func() {
		Get("panic")
	})
}

func (s *SecretSuite) TestSaveGet() {
	val, err := safeGet("xxx", mockGetter)
	s.Require().Equal("xxx", val)
	s.Require().NoError(err)

	val, err = safeGet("panic", mockGetter)
	s.Require().NotNil(err)
}

func TestSecret(t *testing.T) {
	suite.Run(t, new(SecretSuite))
}
