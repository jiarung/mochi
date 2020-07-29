package encrypt

import (
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/satori/go.uuid"

	"github.com/cobinhood/cobinhood-backend/common/aes"
	"github.com/cobinhood/cobinhood-backend/common/config/secret"
)

var (
	// BigqueryEncrypter is encrypter using default bigquery secret settings.
	BigqueryEncrypter *Encrypter
)

// Encrypter defines settings of symmetric encryption.
type Encrypter struct {
	key []byte
	iv  []byte
}

// New receives key and iv for aes encryption and returns an Encrypter.
func New(key, iv []byte) *Encrypter {
	return &Encrypter{key: key, iv: iv}
}

// Initialize inits bigquery encrypter.
func Initialize() {
	key, err := base64.StdEncoding.DecodeString(
		secret.Get("BIGQUERY_SECRET_KEY"))
	if err != nil {
		panic(err.Error())
	}
	iv, err := base64.StdEncoding.DecodeString(
		secret.Get("BIGQUERY_SECRET_IV"))
	if err != nil {
		panic(err.Error())
	}

	BigqueryEncrypter = New(key, iv)
}

// EncryptUUID encrypts an UUID.
func (e *Encrypter) EncryptUUID(id uuid.UUID) uuid.UUID {
	encrypted, err := aes.EncryptWithoutPadding(
		e.key,
		e.iv,
		id.Bytes())
	if err != nil {
		return uuid.Nil
	}
	return uuid.FromBytesOrNil(encrypted)
}

// DecryptUUID decrypts an UUID.
func (e *Encrypter) DecryptUUID(id uuid.UUID) uuid.UUID {
	decrypted, err := aes.DecryptWithoutPadding(
		e.key,
		e.iv,
		id.Bytes())
	if err != nil {
		return uuid.Nil
	}
	return uuid.FromBytesOrNil(decrypted)
}

// EncryptString encrypts a string.
func (e *Encrypter) EncryptString(s string) string {
	encrypted, err := aes.CBCEncrypt(e.key, []byte(s))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(encrypted)
}

// DecryptString decrypts a string.
func (e *Encrypter) DecryptString(s string) string {
	encrypted, err := hex.DecodeString(s)
	if err != nil {
		return ""
	}
	decrypted, err := aes.CBCDecrypt(e.key, encrypted)
	if err != nil {
		return ""
	}
	return string(decrypted)
}

// EncryptEmail encrypts email.
func (e *Encrypter) EncryptEmail(email string) string {
	s := strings.Split(email, "@")
	if len(s) != 2 {
		return ""
	}
	return e.EncryptString(s[0]) + "@" + e.EncryptString(s[1])
}

// DecryptEmail decrypts email.
func (e *Encrypter) DecryptEmail(email string) string {
	s := strings.Split(email, "@")
	if len(s) != 2 {
		return ""
	}
	return e.DecryptString(s[0]) + "@" + e.DecryptString(s[1])
}
