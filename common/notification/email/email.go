package email

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/cobinhood/gorm"

	"github.com/cobinhood/cobinhood-backend/cache"
	"github.com/cobinhood/cobinhood-backend/cache/keys"
	"github.com/cobinhood/cobinhood-backend/common/config/misc"
	jsonBuilder "github.com/cobinhood/cobinhood-backend/common/encode/json"
	"github.com/cobinhood/cobinhood-backend/jsonrpc"
	models "github.com/cobinhood/cobinhood-backend/models/exchange"
	"github.com/cobinhood/cobinhood-backend/types"
)

// Config defines config.
type Config struct {
	CallbackDomain  string
	EmailSenderAddr string
}

// Service provides namespace for email functions.
type Service struct {
	Config
	serviceName string
}

func genConfig() Config {
	return Config{
		EmailSenderAddr: misc.PostmanServerEndpoint(),
	}
}

// New return new service.
func New(cfg Config) *Service {
	e := &Service{Config: cfg}
	return e
}

var (
	email *Service
	once  sync.Once
)

// Default returns service as singleton.
func Default() *Service {
	once.Do(func() {
		email = New(genConfig())
		email.serviceName = "postman"
	})

	return email
}

// ServiceName returns its concrete service name.
func (e *Service) ServiceName() string {
	return e.serviceName
}

// Send calls internal send api.
func (e *Service) Send(req Request, requestTag string) error {

	rpcClient := jsonrpc.NewClient(e.EmailSenderAddr, "", "")
	rpcClient.SetVersion(2)

	requestType := reflect.Indirect(reflect.ValueOf(req)).Type().Name()
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	obj := jsonBuilder.Object(
		jsonBuilder.Attr("type", requestType),
		jsonBuilder.Attr("payload", payload),
		jsonBuilder.Attr("request_tag", requestTag),
	)
	params := jsonrpc.NewParams()
	params.UseObj(obj)
	_, err = rpcClient.Post("sendgrid_handler", params, nil)
	return err
}

// GetEmailRequest get the email's request in interface{} which for unit test.
func GetEmailRequest(requestType string, requestPayload []byte) (
	Request, error) {
	//	interface{}, error) {
	for _, request := range Requests {
		requestModelType := reflect.Indirect(reflect.ValueOf(request)).Type()
		if requestModelType.Name() == requestType {
			req := reflect.New(requestModelType).Interface()
			err := json.Unmarshal(requestPayload, req)
			if err != nil {
				return nil, err
			}
			return req.(Request), nil
		}
	}
	return nil, errors.New(`no email's request found`)
}

// GetUserPreferenceLanguage return user preference language.
func GetUserPreferenceLanguage(db *gorm.DB, email string) (
	locale string, err error) {
	if locale, err = cache.GetRedis().GetString(
		keys.GetUserLanguageKey(email)); err == nil && len(locale) > 0 {
		return
	}
	err = nil

	if db == nil {
		err = errors.New("GetUserPreferenceLanguage failed with no db instance")
		return
	}

	var user models.User
	err = db.Model(models.User{}).Find(&user, "email = ?", email).Error
	if err != nil {
		err = fmt.Errorf("email: %s is not register", email)
		return
	}

	var preferences []models.Preference
	err = db.Model(models.Preference{}).Where("user_id = ? AND key in (?)",
		user.ID, []types.Preference{
			types.PreferenceLanguage,
			types.PreferenceMobileLanguage,
		}).Find(&preferences).Error
	if err != nil {
		err = errors.New("failed to get locale preferences")
		return
	}

	if len(preferences) == 0 {
		locale = "en"
	}
	for _, preference := range preferences {
		if preference.Key == types.PreferenceLanguage {
			locale = LocaleMap[preference.Value]
			break
		}
		locale = LocaleMap[preference.Value]
	}

	cache.GetRedis().Set(
		keys.GetUserLanguageKey(email),
		locale,
	)

	return
}
