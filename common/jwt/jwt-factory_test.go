package jwtfactory

import (
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
)

type JWTFactoryTestSuite struct {
	suite.Suite
}

func (s *JWTFactoryTestSuite) SetupSuite() {
}

func (s *JWTFactoryTestSuite) TearDownSuite() {
}

func (s *JWTFactoryTestSuite) SetupTest() {
}

func (s *JWTFactoryTestSuite) TearDownTest() {
}

func (s *JWTFactoryTestSuite) TestGetSecrets() {
	var jwtTypeServiceMap = map[jwtType][]cobxtypes.ServiceName{
		registrationJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		emailVerificationJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		addWithdrawalWalletEmailVerificationJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx},
		withdrawalFundsEmailVerificationJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APICoblet},
		fiatWithdrawalFundsEmailVerificationJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx},
		epayWithdrawalFundsEmailVerificationJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx},
		accessTokenJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.WS, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		deviceAuthorizationJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		deviceVerificationJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		resetPasswordJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		deleteAccountJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		changeEmailJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		twoFAEnableJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		twoFAEnableConfirmJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		twoFADisableConfirmJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		twoFARequiredJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.APIAdmin, cobxtypes.APICoblet},
		oauth2AuthorizationCodeJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx},
		oauth2AccessTokenJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx},
		oauth2RefreshTokenJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx},
		requestToDisableTwoFAJWT: []cobxtypes.ServiceName{
			cobxtypes.Test, cobxtypes.APICobx, cobxtypes.WS, cobxtypes.APIAdmin, cobxtypes.APICoblet},
	}
	for jwt, services := range jwtTypeServiceMap {
		for _, service := range services {
			s.Require().NotEqual(0, len(getJWTSecret(jwt, service)))
		}
	}
}

func (s *JWTFactoryTestSuite) TestGenAndValidate() {
	rObj := RegistrationObj{uuid.NewV4()}
	// Ok.
	{
		token, err := Build(rObj).Gen(cobxtypes.Test, 1800)
		s.Require().NoError(err)

		c, exp, err := Build(rObj).Validate(token, cobxtypes.Test)
		s.Require().NoError(err)
		s.Require().False(exp)
		s.Require().Equal(rObj.RegistrationID.String(), c["registration_id"])
	}

	// Expired.
	{
		token, err := Build(rObj).Gen(cobxtypes.Test, -56)
		s.Require().NoError(err)

		_, exp, err := Build(rObj).Validate(token, cobxtypes.Test)
		s.Require().True(exp)
		s.Require().Error(err)
	}

	// Error.
	{
		token, err := Build(rObj).Gen(cobxtypes.Test, -56)
		s.Require().NoError(err)

		_, _, err = Build(rObj).Validate(token, cobxtypes.APICobx)
		s.Require().Error(err)
	}
}

func (s *JWTFactoryTestSuite) TestGenAndValidateAPIToken() {
	store := NewAPIKeySecret()
	s.Require().NotNil(store)
	aObj := APITokenObj{uuid.NewV4(), uuid.NewV4(), nil}
	m := BuildWithSecret(aObj, "aaaabbbb")
	s.Require().NotNil(m)
	token, err := m.GenWithCOBSecret(cobxtypes.APICobx, store, 1800)
	s.Require().NoError(err)

	err = ValidateCOBSecret(token, store)
	s.Require().NoError(err)
}

func TestJWTFactory(t *testing.T) {
	suite.Run(t, new(JWTFactoryTestSuite))
}
