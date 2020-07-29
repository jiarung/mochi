package scopeauth

import (
	"testing"

	cobxtypes "github.com/cobinhood/cobinhood-backend/apps/exchange/cobx-types"
	"github.com/cobinhood/cobinhood-backend/types"
	"github.com/stretchr/testify/suite"
)

type ScopeTreeTestSuite struct {
	suite.Suite
}

func (s *ScopeTreeTestSuite) SetupSuite() {
	Initialize(cobxtypes.Test)
	Initialize(cobxtypes.APIAdmin)
}

func (s *ScopeTreeTestSuite) TearDownSuite() {
	Finalize()
}

func (s *ScopeTreeTestSuite) TestTree() {
	s.testEndpoint(cobxtypes.APIAdmin,
		"GET",
		"v1/admin/system/messages",
		[]types.Scope{types.ScopeAdminSystemMessageAdministration})

	s.testEndpoint(cobxtypes.APIAdmin,
		"delete",
		"/v1/admin/system/messages/6c7da036-b627-11e7-9e41-88b111617934",
		[]types.Scope{types.ScopeAdminSystemMessageAdministration})

	s.testEndpoint(cobxtypes.Test,
		"get",
		"/alive",
		[]types.Scope{types.ScopeAuditCommitteeKYCAuditor, types.ScopeAuditCommitteeSuperAdmin})

	s.testEndpoint(cobxtypes.Test,
		"get",
		"/ready",
		[]types.Scope{types.ScopeAuditCommitteeKYCAuditor, types.ScopeAuditCommitteeSuperAdmin})

	s.testEndpoint(cobxtypes.Test,
		"post",
		"/v1/helo",
		[]types.Scope{types.ScopeAuditCommitteeKYCAuditor, types.ScopeAuditCommitteeSuperAdmin})

	s.testEndpoint(cobxtypes.Test,
		"get",
		"/v1/users",
		[]types.Scope{types.ScopePublic})

	s.testEndpoint(cobxtypes.Test,
		"get",
		"/v1/users/anything",
		[]types.Scope{types.ScopeAuditCommitteeKYCAuditor, types.ScopeAuditCommitteeSuperAdmin, types.ScopeExchangeAccountRead})

	s.testEndpoint(cobxtypes.Test,
		"put",
		"/v1/users/arbitrary_thing",
		[]types.Scope{types.ScopeAuditCommitteeKYCAuditor, types.ScopeAuditCommitteeSuperAdmin, types.ScopeExchangeAccountRead})

	s.testNotExistingEndpoint(cobxtypes.Test, "post", "v1/users/arbitrary_thing")
	s.testNotExistingEndpoint(cobxtypes.Test, "put", "v1/userss/arbitrary_thing")
}

func (s *ScopeTreeTestSuite) testEndpoint(service cobxtypes.ServiceName, method, path string, expected []types.Scope) {
	scopes, err := GetScopes(service, method, path)
	s.Require().Nil(err)
	s.Require().Equal(expected, scopes)
}

func (s *ScopeTreeTestSuite) testNotExistingEndpoint(service cobxtypes.ServiceName, method, path string) {
	_, err := GetScopes(service, method, path)
	s.Require().NotNil(err)
}

func TestScopeTree(t *testing.T) {
	suite.Run(t, new(ScopeTreeTestSuite))
}
