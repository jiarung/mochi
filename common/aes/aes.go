package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	// AES128KeySize stands for 16 bytes.
	AES128KeySize = 16
	// AES256KeySize stands for 32 bytes.
	AES256KeySize = 32
)

// EncodedKey is key but hex encoded.
type EncodedKey string

// Key stands for a AES key.
type Key []byte

// ToKey converts to Key.
func (e *EncodedKey) ToKey() Key {
	k, err := hex.DecodeString(string(*e))
	if err != nil {
		panic(err)
	}

	return k
}

// IsValid check whether e is valid.
func (e *EncodedKey) IsValid() bool {
	_, err := hex.DecodeString(string(*e))
	return err == nil
}

// streamKey defines a key that xor with aes cipher key.
type streamKey struct {
	key Key
}

var (
	keySecret streamKey
)

func realKey(key Key) Key {
	// FIXME(xnum): explictly init at cmd.
	once.Do(InitFromSecret)
	ks := keySecret.key
	realKey := make([]byte, len(key))
	for i, b := range key {
		realKey[i] = b ^ ks[i%len(ks)]
	}
	return realKey
}

// CBCEncrypt encrypts b with key in CBC mode, padding the bytes to fit the aes
// block size if needed.
func CBCEncrypt(key Key, b []byte) ([]byte, error) {
	realKey := realKey(key)

	block, err := aes.NewCipher(realKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create new aes cipher with key [%s]. err(%s)",
			key, err)
	}

	// Pad the bytes
	paddedBytes, err := pad(b)
	if err != nil {
		return nil, fmt.Errorf("failed to pad data [%v]. err(%s)", b, err)
	}

	// Generate random initialization vector and put in the beginning of cypher.
	cipherBytes := make([]byte, aes.BlockSize+len(paddedBytes))
	iv := cipherBytes[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to create new cipher with key [%s]. err(%s)",
			key, err)
	}

	// Encrypt data in CBC mode.
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherBytes[aes.BlockSize:], paddedBytes)

	return cipherBytes, nil
}

// CBCDecrypt decrypts cipherBytes with key in CBC mode and unpads the decrypted
// bytes.
func CBCDecrypt(key Key, cipherBytes []byte) ([]byte, error) {
	realKey := realKey(key)

	block, err := aes.NewCipher(realKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create new aes cipher with key [%s]. err(%s)",
			key, err)
	}

	if len(cipherBytes) < aes.BlockSize {
		return nil, errors.New("cipher bytes block size is too short")
	}

	// Use the beginning of the cipher as the initialization vector.
	iv := cipherBytes[:aes.BlockSize]
	cipherBytes = cipherBytes[aes.BlockSize:]

	// Decrypt data in CBC mode.
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherBytes))
	mode.CryptBlocks(decrypted, cipherBytes)

	unpadded, err := unpad(decrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad data. err(%s)", err)
	}

	return unpadded, nil
}

// GenerateAESKey returns a random key, calling `randomString` under the hood.
// The keySize should be either AES128KeySize or AES256KeySize.
func GenerateAESKey(keySize int) (Key, error) {
	if keySize != AES128KeySize && keySize != AES256KeySize {
		return nil, fmt.Errorf("invalid key size(%d). The key size should be"+
			"either AES128KeySize or AES256KeySize", keySize)
	}

	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes. err(%s)", err)
	}
	return key, nil
}

// pad returns the byte array passed as a parameter padded with bytes such that
// the new byte array will be an exact multiple of the aes block size (16).
// The value of each octet of the padding is the size of the padding. If the
// array passed as a parameter is already an exact multiple of the block size,
// the original array will be padded with a full block.
func pad(unpadded []byte) ([]byte, error) {
	unpaddedLen := len(unpadded)
	padLen := aes.BlockSize - (unpaddedLen % aes.BlockSize)
	padBytes := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(unpadded, padBytes...), nil
}

// unpad removes the padding of a given byte array, according to the same rules
// as described in the pad function.
func unpad(padded []byte) ([]byte, error) {
	paddedLen := len(padded)
	if paddedLen == 0 {
		return nil, errors.New("invalid padding size")
	}

	pad := padded[paddedLen-1]
	padLen := int(pad)
	if padLen > paddedLen || padLen > aes.BlockSize {
		return nil, errors.New("invalid padding size")
	}

	for _, v := range padded[paddedLen-padLen : paddedLen-1] {
		if v != pad {
			return nil, errors.New("invalid padding")
		}
	}

	return padded[:paddedLen-padLen], nil
}

// EncryptWithoutPadding encrypts data without using pad function.
func EncryptWithoutPadding(key Key, iv, data []byte) ([]byte, error) {
	if len(data)%aes.BlockSize != 0 || len(data) == 0 {
		return nil, errors.New("invalid data size")
	}
	realKey := realKey(key)

	block, err := aes.NewCipher(realKey)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create new aes cipher with key [%s]. err(%s)", key, err)
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("iv and blocksize mismatch")
	}

	cipherBytes := make([]byte, len(data))

	// Encrypt data in CBC mode.
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherBytes, data)

	return cipherBytes, nil
}

// DecryptWithoutPadding descrypts data without using unpad function.
func DecryptWithoutPadding(key Key, iv, data []byte) ([]byte, error) {
	if len(data)%aes.BlockSize != 0 || len(data) == 0 {
		return nil, fmt.Errorf("invalid data size")
	}
	realKey := realKey(key)

	block, err := aes.NewCipher(realKey)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create new aes cipher with key [%s]. err(%s)", key, err)
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("iv and blocksize mismatch")
	}

	// Decrypt data in CBC mode.
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	return decrypted, nil
}
