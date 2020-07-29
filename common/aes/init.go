package aes

import (
	"sync"

	"github.com/jiarung/mochi/common/config/secret"
)

var (
	once sync.Once
)

// Initialize setups key.
// It receives any positive legnth byte array and expands to 256bit.
func Initialize(key []byte) {
	if len(key) == 0 {
		panic("invalid key size")
	}
	setupKeySecret(key)
}

// InitFromSecret gets key from secret and init from it. It may panic.
func InitFromSecret() {
	// FIXME(xnum): It shouldn't be used. This makes aes depend on secret
	// initialized.
	if len(keySecret.key) == 0 {
		key := []byte(secret.Get("AES_KEY_SECRET"))
		setupKeySecret(key)
	}
}

func setupKeySecret(key []byte) {
	// FIXME(xnum): From none getter, we get empty string.
	if len(key) == 0 {
		return
	}

	ks := key
	for len(ks) < AES256KeySize {
		ks = append(ks, key...)
	}
	keySecret.key = ks

}
