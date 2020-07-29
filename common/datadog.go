package common

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-go/statsd"

	"github.com/cobinhood/mochi/common/utils"
)

// TracerEnabled returns whether or not Tracer is enabled.
func TracerEnabled() bool {
	return utils.Environment() != utils.LocalDevelopment
}

// GetStatsdClient returns a statsd client of dd-agent in k8s.
func GetStatsdClient() (*statsd.Client, error) {
	if TracerEnabled() {
		return statsd.New(
			fmt.Sprint(os.Getenv(
				"KUBERNETES_KUBELET_HOST"), ":", "8125"),
		)
	}
	return nil, nil
}
