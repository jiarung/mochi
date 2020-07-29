package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/cobinhood/mochi/common/config/misc"
	"github.com/cobinhood/mochi/common/utils"
)

// Getter defines the secret getter func.
type Getter func(string) string

// Encrypt secrets.
func Encrypt(name, data string) string {
	result, err := kmsClient.Encrypt([]byte(data))
	if err != nil {
		panic(fmt.Errorf("failed to encrypt secret '%s': %v", name, err))
	}
	return result
}

// Decrypt secrets.
func Decrypt(name, data string) string {
	result, err := kmsClient.Decrypt(data)
	if err != nil {
		panic(fmt.Errorf("failed to decrypt secret '%s': %v", name, err))
	}
	return string(result)
}

// k8sGetter returns k8s secrets.
func k8sGetter(key string) string {
	file := path.Join(misc.CobinhoodSecretPath(), key)
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	s := string(fileData)
	return decryptedSecret(key, s)
}

// staticGetter returns tmp file secrets.
func staticGetter(key string) string {
	if val, ok := secretsData[key]; ok {
		return decryptedSecret(key, val)
	}
	panic(fmt.Errorf("Invalid secret name: %s", key))
}

func noneGetter(key string) string {
	fmt.Fprintf(os.Stderr, "Get: %v", key)
	return ""
}

// decryptedSecret decrets the secret for prod and returns the raw one for the
// others.
func decryptedSecret(key, val string) string {
	if utils.IsProduction() {
		// XXX(hao): hack for prod to work
		if !strings.HasPrefix(val, "CiQ") {
			logger.Warn("not encrypted secret. %s", key)
			return val
		}
		return Decrypt(key, val)
	}
	return val
}
