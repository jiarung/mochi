package context

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jiarung/gorm"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/satori/go.uuid"

	cobxtypes "github.com/jiarung/mochi/apps/exchange/cobx-types"
	"github.com/jiarung/mochi/cache"
	"github.com/jiarung/mochi/cache/instances"
	"github.com/jiarung/mochi/cache/keys"
	apierrors "github.com/jiarung/mochi/common/api/errors"
	apiutils "github.com/jiarung/mochi/common/api/utils"
	"github.com/jiarung/mochi/common/config"
	"github.com/jiarung/mochi/common/logging"
	"github.com/jiarung/mochi/infra/api/utils"
	models "github.com/jiarung/mochi/models/exchange"
	"github.com/jiarung/mochi/types"
)

var (
	errUnexpectedAppCtx = errors.New("unexpected app ctx")
	errNotAuthenticated = errors.New("not authenticated")
)

// Defines static variables.
const (
	UserEmailCacheExpireTime = 604800
)

// GetAppContext is shortcut to obtain AppContext from map
func GetAppContext(ctx *gin.Context) (*AppContext, error) {
	var err error

	value, exists := ctx.Get(config.AppContext)
	if !exists {
		err = fmt.Errorf(
			"GetAppContext: ctx.Get(config.AppContext) failed. "+
				"config.AppContext(%s)",
			config.AppContext)
		return nil, err
	}

	appCtx, ok := value.(*AppContext)
	if !ok {
		err = fmt.Errorf(
			"GetAppContext: value.(*context.AppContext) failed. value(%+v)",
			value)
		return nil, err
	}
	return appCtx, nil
}

// AppContext contains shared resources and client, and
// passed through handlers as key-value pair in `gin.Context"
type AppContext struct {
	ctx    *gin.Context
	logger logging.Logger

	DB    *gorm.DB
	Cache *cache.Redis

	isPrivilegedIP bool

	RequestRawIP net.IP
	RequestIP    string

	ServiceName cobxtypes.ServiceName

	RequiredScopes []types.Scope

	APITokenID              *uuid.UUID
	UserID                  *uuid.UUID
	UserAuthorizationScopes []types.Scope
	DeviceAuthorizationID   *uuid.UUID
	AccessTokenID           *uuid.UUID
	OAuth2TokenID           *uuid.UUID
	OAuth2ClientID          uuid.UUID

	// Client side information.
	Platform        types.DevicePlatform
	PlatformVersion apiutils.SemanticVersion
	AppVersion      apiutils.SemanticVersion
}

// NewAppCtx creates AppContext.
func NewAppCtx(
	ctx *gin.Context,
	logger logging.Logger,
	db *gorm.DB,
	cache *cache.Redis) (*AppContext, error) {
	if logger == nil {
		logging.NewLogger().Error("%s", errors.New("logger == nil"))
		return nil, errors.New("logger == nil")
	}
	if ctx == nil {
		logger.Error("gin ctx == nil")
		return nil, errors.New("gin ctx == nil")
	}
	if db == nil {
		logger.Error("db == nil")
		return nil, errors.New("db == nil")
	}
	if cache == nil {
		logger.Error("cache == nil")
		return nil, errors.New("cache == nil")
	}

	appCtx := &AppContext{
		ctx:    ctx,
		logger: logger,
		DB:     db,
		Cache:  cache,
	}

	err := appCtx.SetRequestIP()
	if err != nil {
		logger.Error("failed to set request ip. err(%s)", err)
		return nil, err
	}

	// TODO(wmin0): remove deprecated code after m42.
	appVersion := appCtx.getHeader("App-Version")
	if !strings.Contains(appVersion, ".") {
		// XXX(wmin0): fallback to minor version.
		appVersion = "3." + appVersion + ".0"
	}
	appCtx.AppVersion = apiutils.NewSemanticVersion(appVersion)

	appCtx.PlatformVersion =
		apiutils.NewSemanticVersion(appCtx.getHeader("Platform-Version"))

	appCtx.Platform = types.DevicePlatform(appCtx.getHeader("Platform"))

	ctx.Set(config.AppContext, appCtx)
	return appCtx, nil
}

func (appCtx *AppContext) getHeader(key string) string {
	if appCtx.ctx.Request == nil || appCtx.ctx.Request.Header == nil {
		return ""
	}

	return appCtx.ctx.GetHeader(key)
}

// RequestBody returns request body.
func (appCtx *AppContext) RequestBody() []byte {
	body, ok := appCtx.ctx.Get(config.RequestBody)
	if ok {
		return body.([]byte)
	}

	logger := appCtx.Logger()
	req := appCtx.Request()

	var buffer []byte
	var err error
	if req.Body != nil {
		buffer, err = ioutil.ReadAll(req.Body)
		if err != nil {
			logger.Error("Failed to read body: %s", err)
			buffer = nil
		} else {
			// Restore request body for external use.
			req.Body = ioutil.NopCloser(bytes.NewBuffer(buffer))
		}
	}

	appCtx.ctx.Set(config.RequestBody, buffer)
	return buffer
}

// SetRequestBody sets request body.
func (appCtx *AppContext) SetRequestBody(body []byte) {
	if appCtx.ctx.Request != nil {
		appCtx.ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
	appCtx.ctx.Set(config.RequestBody, body)
}

// SetParams sets request params.
func (appCtx *AppContext) SetParams(params map[string]string) {
	appCtx.ctx.Params = gin.Params{}
	for key, value := range params {
		appCtx.ctx.Params = append(appCtx.ctx.Params, gin.Param{
			Key:   key,
			Value: value,
		})
	}
}

// GetUserEmail gets user email from User data.
func (appCtx *AppContext) GetUserEmail() string {
	if !appCtx.IsAuthenticated() {
		return ""
	}

	// Check cache.
	email, err := appCtx.Cache.GetString(
		keys.GetUserEmailCacheKey(*appCtx.UserID))
	if err == nil {
		return email
	}

	// Query from User.
	user := models.User{}
	result := appCtx.DB.
		Model(&models.User{}).
		Select("email").
		Where("id = ?", *appCtx.UserID).
		First(&user)

	if result.Error != nil {
		return ""
	}

	email = user.Email

	// Add to cache.
	err = appCtx.Cache.Set(keys.GetUserEmailCacheKey(*appCtx.UserID),
		email,
		UserEmailCacheExpireTime)
	if err != nil {
		appCtx.Logger().Error("failed to add to cache. err(%v)", err)
	}

	return email
}

// GetUserNationality gets nationality from KYCData.
func (appCtx *AppContext) GetUserNationality() string {
	if !appCtx.IsAuthenticated() {
		return ""
	}

	nationality, log, err := apiutils.GetKYCDataNationalityCountry(
		appCtx, appCtx.DB, appCtx.Cache, *appCtx.UserID, time.Now())
	if err != nil {
		appCtx.Logger().Error("%s", log)
		return ""
	}
	return nationality
}

// GetUserResidence gets residence from KYCData.
func (appCtx *AppContext) GetUserResidence() string {
	if !appCtx.IsAuthenticated() {
		return ""
	}

	residence, log, err := apiutils.GetKYCDataResidenceCountry(
		appCtx, appCtx.DB, appCtx.Cache, *appCtx.UserID, time.Now())
	if err != nil {
		appCtx.Logger().Error("%s", log)
		return ""
	}
	return residence
}

// CheckEmployeeIdentity checks if the user is employee.
func (appCtx *AppContext) CheckEmployeeIdentity() bool {
	email := appCtx.GetUserEmail()
	return strings.HasSuffix(email, "jiarung.com")
}

// RequestTag returns request tag.
func (appCtx *AppContext) RequestTag() string {
	return appCtx.ctx.GetString(logging.LabelTag)
}

// Logger returns appCtx.logger and it must not be nil.
func (appCtx *AppContext) Logger() logging.Logger {
	return appCtx.logger
}

// SetError sets error.
func (appCtx *AppContext) SetError(code string, args ...string) {
	ctx := appCtx.ctx
	apiutils.SetError(ctx, code, args...)
	ctx.Abort()
}

// SetIgnoreAndAbort sets ignore abort.
func (appCtx *AppContext) SetIgnoreAndAbort() {
	ctx := appCtx.ctx
	apiutils.SetIgnoreAbort(ctx)
	ctx.Abort()
}

// IsIgnoreAbort gets if ignore abort.
func (appCtx *AppContext) IsIgnoreAbort() bool {
	return apiutils.IsIgnoreAbort(appCtx.ctx)
}

// Error returns error depends on data in context.
func (appCtx *AppContext) Error() (int, *apiutils.FailureObj) {
	ctx := appCtx.ctx
	code := ctx.GetString(apiutils.ErrorKey)

	strs := ctx.GetStringSlice(apiutils.ErrorKeyArgs)
	args := []interface{}{}
	for idx := range strs {
		args = append(args, strs[idx])
	}

	status := apierrors.HTTPStatus(code)
	errStr := fmt.Sprintf(code, args...)
	return status, apiutils.Failure(apiutils.ErrorCode(errStr))
}

// SetJSON sets json response.
func (appCtx *AppContext) SetJSON(resp interface{}) {
	apiutils.SetJSON(appCtx.ctx, resp)
}

// SetResp sets response.
func (appCtx *AppContext) SetResp(mime string, resp []byte) {
	apiutils.SetResp(appCtx.ctx, mime, resp)
}

// JSON returns json response depends on data in context.
func (appCtx *AppContext) JSON() *apiutils.SuccessObj {
	ctx := appCtx.ctx
	// Ignore not exist.
	resp, _ := ctx.Get(apiutils.RespKey)
	return apiutils.Success(resp)
}

// Resp returns response depends on data in context.
func (appCtx *AppContext) Resp() (string, []byte) {
	ctx := appCtx.ctx
	// Ignore not exist.
	resp, _ := ctx.Get(apiutils.RespKey)
	mime := ctx.GetString(apiutils.RespMimeKey)
	return mime, resp.([]byte)
}

// Abort aborts the appCtx.
func (appCtx *AppContext) Abort() {
	appCtx.ctx.Abort()
}

// IsAborted returns if appCtx is aborted.
func (appCtx *AppContext) IsAborted() bool {
	return appCtx.ctx.IsAborted()
}

// BindJSON bind body request to input object.
func (appCtx *AppContext) BindJSON(req interface{}) error {
	// Because gin.Context binding only can bind once, we use our own body to
	// decode.
	body := appCtx.RequestBody()
	err := json.Unmarshal(body, req)
	if err != nil {
		return err
	}
	if binding.Validator == nil {
		return nil
	}
	return binding.Validator.ValidateStruct(req)
}

// MustBindJSON bind body request to obj and returns false if unmarshal JSON
// failed then logs it and set error.
func (appCtx *AppContext) MustBindJSON(obj interface{}) bool {
	err := appCtx.BindJSON(obj)
	if err != nil {
		appCtx.Logger().Error("appCtx.MustBindJSON(): %v", err)
		appCtx.SetError(apierrors.ParseJSONError)
		return false
	}

	return true
}

// Query gets data from query string.
func (appCtx *AppContext) Query(key string) string {
	return appCtx.ctx.Query(key)
}

// DefaultQuery gets data from query string with default.
func (appCtx *AppContext) DefaultQuery(key string, def string) string {
	return appCtx.ctx.DefaultQuery(key, def)
}

// GetQuery gets data from query string.
func (appCtx *AppContext) GetQuery(key string) (string, bool) {
	return appCtx.ctx.GetQuery(key)
}

// Param gets data from url params.
func (appCtx *AppContext) Param(key string) string {
	return appCtx.ctx.Param(key)
}

// FormFile gets file from post form params.
func (appCtx *AppContext) FormFile(
	filename string) (*multipart.FileHeader, error) {
	return appCtx.ctx.FormFile(filename)
}

// PostForm gets data from post form params.
func (appCtx *AppContext) PostForm(key string) string {
	return appCtx.ctx.PostForm(key)
}

// GetPostForm gets data from post form params.
func (appCtx *AppContext) GetPostForm(key string) (string, bool) {
	return appCtx.ctx.GetPostForm(key)
}

// PostFormArray gets data from post form params.
func (appCtx *AppContext) PostFormArray(key string) []string {
	return appCtx.ctx.PostFormArray(key)
}

// GetPostFormArray gets data from post form params.
func (appCtx *AppContext) GetPostFormArray(key string) ([]string, bool) {
	return appCtx.ctx.GetPostFormArray(key)
}

// Request returns request object.
func (appCtx *AppContext) Request() *http.Request {
	return appCtx.ctx.Request
}

// Writer returns response writer object.
func (appCtx *AppContext) Writer() http.ResponseWriter {
	return appCtx.ctx.Writer
}

// SetRedirect sets redirect.
func (appCtx *AppContext) SetRedirect(url string) {
	apiutils.SetRedirect(appCtx.ctx, url)
}

// Redirect returns rediect url.
func (appCtx *AppContext) Redirect() string {
	return appCtx.ctx.GetString(apiutils.RedirectKey)
}

// Deadline implements context.Context interface.
func (appCtx *AppContext) Deadline() (deadline time.Time, ok bool) {
	return appCtx.Request().Context().Deadline()
}

// Done implements context.Context interface.
func (appCtx *AppContext) Done() <-chan struct{} {
	return appCtx.Request().Context().Done()
}

// Err implements context.Context interface.
func (appCtx *AppContext) Err() error {
	return appCtx.Request().Context().Err()
}

// Value implements context.Context interface.
func (appCtx *AppContext) Value(key interface{}) interface{} {
	return appCtx.ctx.Value(key)
}

// GetUserID returns their user ID (uuid.UUID) in AppContext.
func (appCtx *AppContext) GetUserID() (uuid.UUID, error) {
	if !appCtx.IsAuthenticated() {
		return uuid.Nil, errNotAuthenticated
	}
	return *appCtx.UserID, nil
}

// GetUser returns user by user ID in AppContext.
func (appCtx *AppContext) GetUser() (*models.User, error) {
	if !appCtx.IsAuthenticated() {
		return nil, errNotAuthenticated
	}

	user := &models.User{ID: *appCtx.UserID}

	err := appCtx.DB.Find(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

// IsAuthenticated indicates the user ID from App Context is authenticated or not.
func (appCtx *AppContext) IsAuthenticated() bool {
	return appCtx.UserID != nil && !uuid.Equal(*appCtx.UserID, uuid.Nil)
}

// ValidateAuthenticated checks whether the user is logon and set error if not.
func (appCtx *AppContext) ValidateAuthenticated() bool {
	if appCtx.IsAuthenticated() {
		return true
	}

	appCtx.SetError(apierrors.AuthenticationError)
	return false
}

// IsPrivilegedIP returns appCtx.isPrivilegedIP.
func (appCtx *AppContext) IsPrivilegedIP() bool {
	return appCtx.isPrivilegedIP
}

// IsMobile indicates whether an request is from Mobile or not.
func (appCtx *AppContext) IsMobile() bool {
	return (appCtx.Platform == types.DeviceAndroid ||
		appCtx.Platform == types.DeviceIOS)
}

// IsAPIToken indicates whether an API token is used.
func (appCtx *AppContext) IsAPIToken() bool {
	return appCtx.APITokenID != nil
}

// IsOAuth2Token indicates whether an OAuth2 access token is used.
func (appCtx *AppContext) IsOAuth2Token() bool {
	return appCtx.OAuth2TokenID != nil
}

// Set store object in the context
func (appCtx *AppContext) Set(key string, value interface{}) {
	appCtx.ctx.Set(key, value)
}

// SetRequestIP set appCtx.RequestIP and appCtx.isPrivilegedIP.
func (appCtx *AppContext) SetRequestIP() error {
	ctx := appCtx.ctx
	if ctx.Request == nil {
		return nil
	}
	appCtx.RequestRawIP = utils.GetIP(ctx.Request)
	appCtx.RequestIP = appCtx.RequestRawIP.String()
	isPrivilegedIP, err := utils.IsPrivilegedIP(ctx.Request)
	if err != nil {
		return err
	}
	appCtx.isPrivilegedIP = isPrivilegedIP
	return nil
}

// GetAndFilterCurrencies returns cached currencies with items blacklisted by
// the request IP address removed.
func (appCtx *AppContext) GetAndFilterCurrencies() (
	[]models.Currency, error) {

	currencies, err := instances.GetCurrencies()
	if err != nil {
		return nil, err
	}

	// privileged IP is not filtered, and
	// can bypass USD withdrawal/deposit frozen.
	if appCtx.isPrivilegedIP {
		for index := range currencies {
			if currencies[index].ID == "USD" {
				currencies[index].WithdrawalFrozen = false
				currencies[index].DepositFrozen = false
			}
		}
		return currencies, nil
	}

	results := make([]models.Currency, 0, len(currencies))
	for _, currency := range currencies {
		if !currency.BlackList.Contains(appCtx.RequestRawIP) {
			results = append(results, currency)
		}
	}
	return results, nil
}

// GetFilteredCurrency return cached currency with items blacklisted by
// the request IP address removed.
func (appCtx *AppContext) GetFilteredCurrency(
	currencyID string) (models.Currency, error) {

	currency, err := instances.GetCurrency(currencyID)
	if err != nil {
		return models.Currency{}, err
	}

	// privileged IP is not filtered and can
	// bypass USD withdrawal/deposit frozen.
	if appCtx.isPrivilegedIP {
		if currency.ID == "USD" {
			currency.WithdrawalFrozen = false
			currency.DepositFrozen = false
		}
		return currency, nil
	}

	if currency.BlackList.Contains(appCtx.RequestRawIP) {
		return models.Currency{}, errors.New("invalid currency")
	}
	return currency, nil
}

// GetAndFilterTradingPairs returns cached trading pairs with items blacklisted
// by the request IP address removed.
func (appCtx *AppContext) GetAndFilterTradingPairs() (
	[]models.TradingPair, error) {

	tradingPairs, err := instances.GetTradingPairs()
	if err != nil {
		return nil, err
	}

	if appCtx.isPrivilegedIP {
		return tradingPairs, nil
	}

	results := make([]models.TradingPair, 0, len(tradingPairs))
	for _, tradingPair := range tradingPairs {
		if !tradingPair.BlackList.Contains(appCtx.RequestRawIP) {
			results = append(results, tradingPair)
		}
	}
	return results, nil
}

// IsCurrencyBlackListed checks whether a currency is blacklisted on the request
// IP address recorded in the AppContext.
func (appCtx *AppContext) IsCurrencyBlackListed(currencyID string) (
	bool, error) {
	if appCtx.isPrivilegedIP {
		return false, nil
	}

	currency, err := instances.GetCurrency(currencyID)
	if err != nil {
		return false, err
	}

	return currency.BlackList.Contains(appCtx.RequestRawIP), nil
}

// IsTradingPairBlackListed checks whether a trading pair is blacklisted on the
// request IP address recorded in the AppContext.
func (appCtx *AppContext) IsTradingPairBlackListed(tradingPairID string) (
	bool, error) {
	if appCtx.isPrivilegedIP {
		return false, nil
	}

	tradingPair, err := instances.GetTradingPair(tradingPairID)
	if err != nil {
		return false, err
	}
	return tradingPair.BlackList.Contains(appCtx.RequestRawIP), nil
}

// SetServiceName set appCtx.ServiceName
func (appCtx *AppContext) SetServiceName(serviceName cobxtypes.ServiceName) error {
	if !serviceName.IsValid() {
		return errors.New("service name is not valid")
	}
	appCtx.ServiceName = serviceName
	return nil
}

// CreateSubRequestCtx creates a sub request related to current context but
// without any request data.
func (appCtx *AppContext) CreateSubRequestCtx() (*AppContext, error) {
	req := appCtx.Request()

	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	newReq, err := http.NewRequest(req.Method, req.RequestURI, nil)
	if err != nil {
		return nil, err
	}

	// TODO(wmin0): if need specify timeout.
	newReq.WithContext(req.Context())

	newReq.Proto = req.Proto
	newReq.ProtoMajor = req.ProtoMajor
	newReq.ProtoMinor = req.ProtoMinor
	newReq.RemoteAddr = req.RemoteAddr
	for key, value := range req.Header {
		newReq.Header[key] = append([]string(nil), value...)
	}
	for key, value := range req.Trailer {
		newReq.Trailer[key] = append([]string(nil), value...)
	}

	ctx := appCtx.ctx
	newCtx := &gin.Context{}
	*newCtx = *appCtx.ctx
	newCtx.Request = newReq

	newCtx.Params = nil
	newCtx.Keys = map[string]interface{}{}
	for key, value := range ctx.Keys {
		switch key {
		case config.RequestBody, config.AppContext:
			continue
		default:
			newCtx.Keys[key] = value
		}
	}
	newCtx.Accepted = append([]string(nil), ctx.Accepted...)

	newAppCtx := &AppContext{}
	*newAppCtx = *appCtx
	newAppCtx.ctx = newCtx

	newCtx.Set(config.AppContext, newAppCtx)
	return newAppCtx, nil
}
