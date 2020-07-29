package scopeauth

import (
	"testing"

	cobxtypes "github.com/jiarung/mochi/apps/exchange/cobx-types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ScopeTestSuite struct {
	suite.Suite
}

func (at *ScopeTestSuite) TestGetScopes() {
	Initialize(cobxtypes.Test)
	defer Finalize()
	testMap, ok := scopeMap["test"]
	require.True(at.T(), ok)
	for path, endpointMap := range testMap {
		for method, v := range endpointMap {
			scopes, err := GetScopes("test", method, path)
			at.T().Logf("[%s] %s: %s", method, path, scopes)
			require.Nil(at.T(), err)
			require.Equal(at.T(), v, scopes)
		}
	}
}

func TestScope(t *testing.T) {
	suite.Run(t, new(ScopeTestSuite))
}
