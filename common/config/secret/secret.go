package secret

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/jiarung/mochi/common/logging"
	"github.com/jiarung/mochi/common/utils"
)

var (
	logger = logging.NewLoggerTag("secret")

	secrets sync.Map
	getters []Getter
)

// RandomEncryptedSecretString generates a random k8s encrypt secret.
func RandomEncryptedSecretString(length int) string {
	var letterRunes = []rune(
		"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return Encrypt("GenRandomEncryptedSecretString", string(b))
}

// RandomEncryptedSecretBytes generates a random byte array with specified
// length, base64 encodes the bytes array and encrypts with kms.
func RandomEncryptedSecretBytes(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	str := base64.StdEncoding.EncodeToString(b)
	return Encrypt("GenRandomEncryptedSecretBytes", str)
}

// Get returns the secret compiled into the binary.
func Get(key string) string {
	// try to load from cache.
	if v, ok := secrets.Load(key); ok {
		return v.(string)
	}

	// load from getters.
	var (
		secret string
		err    error
	)
	for i, getter := range getters {
		secret, err = safeGet(key, getter)
		if err != nil {
			if utils.IsProduction() {
				logger.Warn("failed to get secret [%s] with getter [%d]. err: %v",
					key, i, err)
			}
			continue
		} else {
			break
		}
	}
	if err != nil {
		logger.Error("failed to get [%s] secret. err: %v", key, err)
		panic(err)
	}

	// save back to cache
	secrets.Store(key, secret)
	return secret
}

func safeGet(key string, getter Getter) (secret string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v\n%s", recovered, string(debug.Stack()))
		}
	}()
	secret = getter(key)
	return
}
