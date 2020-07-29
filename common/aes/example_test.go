package aes_test

import (
	"encoding/base64"
	"fmt"

	"github.com/jiarung/mochi/common/aes"
	"github.com/jiarung/mochi/models/exchange/exchangetest"
)

func Example() {
	key, err := aes.GenerateAESKey(aes.AES256KeySize)
	if err != nil {
		// handle err
		return
	}

	fileBytes, err := base64.StdEncoding.DecodeString(exchangetest.TestingUserDataProfilePic)
	if err != nil {
		// handle err
		return
	}

	encrypted, err := aes.CBCEncrypt(key, fileBytes)
	if err != nil {
		// handle err
		return
	}

	decrypted, err := aes.CBCDecrypt(key, encrypted)
	if err != nil {
		// handle err
		return
	}

	fmt.Printf("done. %v", decrypted)
}
