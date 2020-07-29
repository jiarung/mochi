package aes

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AESTestSuite struct {
	suite.Suite
}

func (suite *AESTestSuite) TestEncryptAndDecrypt() {
	key, err := GenerateAESKey(AES256KeySize)
	require.Nil(suite.T(), err)

	fileBytes := []byte("demo file")
	require.Nil(suite.T(), err)

	encrypted, err := CBCEncrypt(key, fileBytes)
	require.Nil(suite.T(), err)

	decrypted, err := CBCDecrypt(key, encrypted)
	require.Nil(suite.T(), err)

	require.Equal(suite.T(), fileBytes, decrypted)
}

func TestAES(t *testing.T) {
	suite.Run(t, &AESTestSuite{})
}
