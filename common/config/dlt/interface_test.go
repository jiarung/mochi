package dlt

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cobinhood/cobinhood-backend/cache"
	"github.com/cobinhood/cobinhood-backend/infra/app"
)

type interfaceTestSuite struct {
	suite.Suite
	testcase *endpoint
}

func (s *interfaceTestSuite) SetupSuite() {
	var config struct {
		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	cache.Initialize(config.Cache)
	s.testcase = blockchainBitcoinEndpoint
}

func (s *interfaceTestSuite) TearDownSuite() {
	cache.GetRedis().Delete(s.testcase.Key())
	cache.Finalize()
}

func (s *interfaceTestSuite) TestOverwriteEndpoint() {
	// When haven't overwrite.
	s.Require().NotNil(blockchainBitcoinEndpoint)
	s.Require().Equal(s.testcase.endpoint, s.testcase.Endpoint(true))
	s.Require().Equal(s.testcase.endpoint, s.testcase.Endpoint(false))

	// Set the new endpoint.
	newEndpoint := "http://go.rough"
	s.Require().Nil(cache.GetRedis().Set(s.testcase.Key(), newEndpoint))

	s.Require().Equal(newEndpoint, s.testcase.Endpoint())
	s.Require().Equal(newEndpoint, s.testcase.Endpoint(false))
	s.Require().Equal(s.testcase.endpoint, s.testcase.Endpoint(true))
}

func TestInterface(t *testing.T) {
	suite.Run(t, &interfaceTestSuite{})
}
