package auth

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testDigestSuite struct {
	suite.Suite
}

func (s *testDigestSuite) TestNewDigestAuthParams() {
	da1 := NewDigestAuthParams("")
	s.Require().Nil(da1)

	da2 := NewDigestAuthParams(" ")
	s.Require().Nil(da2)

	da4 := NewDigestAuthParams(`Digest realm="testrealm@host.com",
                        qop="auth,auth-int",
                        nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093",
                        opaque="5ccc069c403ebaf9f0171e9517f40e41"`)
	s.Require().NotNil(da4)
	s.Require().Equal("testrealm@host.com", da4.Realm)
	s.Require().Equal("auth,auth-int", da4.Qop)
	s.Require().Equal("dcd98b7102dd2f0e8b11d0f600bfb0c093", da4.Nonce)
	s.Require().Equal("5ccc069c403ebaf9f0171e9517f40e41", da4.Opaque)
}

func TestDigest(t *testing.T) {
	suite.Run(t, &testDigestSuite{})
}
