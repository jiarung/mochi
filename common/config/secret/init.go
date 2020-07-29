package secret

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"

	"github.com/cobinhood/mochi/common/utils"
	"github.com/cobinhood/mochi/gcp/kms"
)

var (
	kmsClient   *kms.Client
	secretsData map[string]string
)

// SourceType defines where the secret should read from.
type SourceType int

// SourceType enum.
// - None always returns empty string.
// - K8S reads secret from file which is mounted.
// - File parse $GITROOT/tmp/secrets.yaml and read from it.
// - Auto determines the source by env.
const (
	None SourceType = iota
	K8S
	File
	Auto
)

// Config defines the options to secret package.
type Config struct {
	// UseKms determines to init kms client or not. It's for fullnode env.
	UseKms bool
	Source SourceType
}

// FIXME(xnum): It should be removed.
func init() {
	Initialize(Config{
		UseKms: true,
		Source: Auto,
	})
}

// Initialize initializes kms client and sets secret getter.
func Initialize(cfg Config) {
	var err error
	if cfg.UseKms {
		// FIXME(xnum): set at cmd.
		if utils.FullnodeCluster != utils.Environment() {
			if err = initKmsClient(); err != nil {
				panic(err.Error())
			}
		}
	}

	switch cfg.Source {
	case None:
		getters = []Getter{noneGetter}
	case K8S:
		getters = []Getter{k8sGetter}
	case File:
		getters = []Getter{staticGetter}
		if err = initDataFromFile(); err != nil {
			panic(err)
		}
	// FIXME(xnum): not encourge to use. It depends on env.
	case Auto:
		if utils.Environment() == utils.LocalDevelopment ||
			utils.Environment() == utils.CI {
			getters = []Getter{staticGetter}
			err := initDataFromFile()
			if err != nil {
				log.Panicln(err)
			}
			return
		}
		getters = []Getter{k8sGetter}
	}
}

func initKmsClient() error {
	// FIXME(xnum): don't determines by env.
	keyType := kms.KeyDev
	if utils.IsProduction() {
		keyType = kms.KeySecret
	}
	var err error
	kmsClient, err = kms.NewDefaultClient(context.Background(), keyType)
	return err
}

func initDataFromFile() error {
	// XXX(xnum): To identify scrects from different app at localdev,
	// we read app value from env directly.
	// Since we are going to migrate to new config management method,
	// not using misc package doesn't matter anymore.
	input, err := ioutil.ReadFile(
		fmt.Sprintf("%s/kubernetes/%s/secret/secrets.yaml.in",
			os.Getenv("GITROOT"), os.Getenv("APP")))
	if err != nil {
		return fmt.Errorf("Invalid load secret: %v", err)
	}

	type k8sSecret struct {
		Data map[string]string
	}
	var s k8sSecret
	err = yaml.Unmarshal(input, &s)
	if err != nil {
		return fmt.Errorf("Invalid unmarshal: %+v", err)
	}
	secretsData = map[string]string{}
	for k, v := range s.Data {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			// Decode failed, maybe not base64 encoded?
			secretsData[k] = v
		} else {
			secretsData[k] = string(b)
		}
	}

	return nil
}
