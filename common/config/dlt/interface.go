package dlt

import (
	"github.com/jiarung/mochi/cache"
)

type endpoint struct {
	endpoint string
	key      string
}

// GetEndpointFunc defines the type of function ptr of Endpoint.
type GetEndpointFunc func(...bool) string

// Endpoint returns the endpoint. If assign original = true, it would ignore
// what redis overwirtes and force returning the original endpoint value.
func (e *endpoint) Endpoint(original ...bool) string {
	if cache.GetRedis() != nil {
		if len(original) > 0 && original[0] {
			return e.endpoint
		}
		if cacheEndpoint, err := cache.GetRedis().GetString(e.key); err == nil {
			return cacheEndpoint
		}
	}
	return e.endpoint
}

// Key returns the redis key for overwrite.
func (e *endpoint) Key() string {
	return e.key
}
