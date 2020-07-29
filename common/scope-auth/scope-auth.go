package scopeauth

import (
	"errors"
	"sync"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/common/logging"
	"github.com/cobinhood/mochi/types"
)

// Exported errors.
var (
	ErrServiceNotFound = errors.New("service not found")
	ErrNotInitialized  = errors.New("scope tree not initialized")
)

var mu sync.Mutex
var trees map[cobxtypes.ServiceName]*ScopeTree

// logger return a logger with scope-auth tag.
func logger() logging.Logger {
	return logging.NewLoggerTag("scope-auth")
}

// Initialize initializes the scope tree of service.
func Initialize(service cobxtypes.ServiceName) {
	mu.Lock()
	defer mu.Unlock()

	if !service.IsValid() {
		panic(errors.New("service is invalid"))
	}

	if trees == nil {
		trees = map[cobxtypes.ServiceName]*ScopeTree{}
	}

	if _, initialized := trees[service]; initialized {
		return
	}

	logger := logger()
	logger.Debug("building scope tree of [%s] service", service)
	serviceMap, exists := scopeMap[string(service)]
	if !exists {
		panic(ErrServiceNotFound)
	}
	t := NewTree(string(service))
	for endpoint, endpointMap := range serviceMap {
		for method, scopes := range endpointMap {
			err := t.InsertWithPath(endpoint, method, scopes)
			if err != nil {
				panic(err)
			}
		}
	}
	logger.Debug(t.String())
	trees[service] = t
}

// Finalize finalizes the scope tree.
func Finalize() {
	mu.Lock()
	defer mu.Unlock()

	trees = nil
}

// GetScopes returns a slice of scopes of the endpoint
func GetScopes(service cobxtypes.ServiceName, method string, path string) ([]types.Scope, error) {
	return GetTree(service).GetScopesWithPath(path, method)
}

// GetTree returns the full tree by service.
func GetTree(service cobxtypes.ServiceName) (tree *ScopeTree) {
	if trees == nil {
		panic(ErrNotInitialized)
	}

	tree, exists := trees[service]
	if !exists {
		panic(ErrServiceNotFound)
	}
	return
}
