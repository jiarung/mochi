package encrypt

import (
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
)

type utilsTestSuite struct {
	suite.Suite
}

func (s *utilsTestSuite) TestEncryptAndDecryptUUID() {
	iv := []byte("aaaabbbbccccdddd")
	e := New(iv, iv)
	id := uuid.NewV4()
	encrypted := e.EncryptUUID(id)
	s.Require().False(uuid.Equal(uuid.Nil, encrypted))
	s.Require().False(uuid.Equal(id, encrypted))
	decrypted := e.DecryptUUID(encrypted)
	s.Require().True(uuid.Equal(id, decrypted))
}

func TestUtils(t *testing.T) {
	suite.Run(t, &utilsTestSuite{})
}
