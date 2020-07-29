package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"

	apierrors "github.com/cobinhood/cobinhood-backend/common/api/errors"
	"github.com/cobinhood/cobinhood-backend/models/exchange"
	ctime "github.com/cobinhood/cobinhood-backend/time"
	"github.com/cobinhood/cobinhood-backend/types"
)

// ErrorKey defines shared key to set error in `gin.Context`, use unusual string
// starts with underline which is not GO recommended convention to prevent
// duplicated.
const ErrorKey = "_error_code_"

// ErrorKeyArgs use as format args for ErrorKey
const ErrorKeyArgs = "_error_code_args_"

// RespKey defines shared key to set resp in `gin.Context`.
const RespKey = "_resp_"

// RespMimeKey defines shared key to set resp mime.
const RespMimeKey = "_resp_mime_"

// RespRawKey defines shared key to specify raw output.
const RespRawKey = "_resp_raw_"

// RedirectKey defines shared key to specify redirect url.
const RedirectKey = "_rediect_"

// IgnoreAbortKey is a flag that skip abort check on early return response, e.g.
// CacheMiddleware writes cached response and hijacks processing flow with
// `Abort()` method, and shouldn't be processed by this handler.
const IgnoreAbortKey = "_ignore_abort_"

// SuccessObj defines struct of success object.
type SuccessObj struct {
	Success bool        `json:"success" binding:"required"`
	Result  interface{} `json:"result" binding:"required"`
}

// Success returns an HTTP success response object.
func Success(response interface{}) *SuccessObj {
	return &SuccessObj{
		Success: true,
		Result:  response,
	}
}

// FailureObj defines struct of failure object.
type FailureObj struct {
	Success bool        `json:"success" binding:"required"`
	Error   interface{} `json:"error" binding:"required"`
}

// String returns internal error string.
func (f *FailureObj) String() string {
	return fmt.Sprintf("%s", f.Error)
}

// Failure returns an HTTP failure response object.
func Failure(response interface{}) *FailureObj {
	return &FailureObj{
		Success: false,
		Error:   response,
	}
}

// FailureWithTagObj defines struct of failure with tag object.
type FailureWithTagObj struct {
	*FailureObj

	// Expose tag to client for addressing error request
	Tag string `json:"request_id" binding:"required"`
}

// FailureWithTag returns an HTTP failure response object.
func FailureWithTag(f *FailureObj, tag string) *FailureWithTagObj {
	return &FailureWithTagObj{
		FailureObj: f,
		Tag:        tag,
	}
}

// ErrorCodeObj defines struct of error code object.
type ErrorCodeObj struct {
	ErrorCode string `json:"error_code"`
}

// String returns error string.
func (e *ErrorCodeObj) String() string {
	return e.ErrorCode
}

// ErrorCode returns an error code wrapper object.
func ErrorCode(code string) *ErrorCodeObj {
	return &ErrorCodeObj{code}
}

// SetError sets error.
func SetError(ctx *gin.Context, code string, args ...string) {
	ctx.Set(ErrorKey, code)
	ctx.Set(ErrorKeyArgs, args)
}

// SetJSON sets json response.
func SetJSON(ctx *gin.Context, resp interface{}) {
	ctx.Set(RespKey, resp)
	ctx.Set(RespRawKey, false)
}

// SetResp sets response.
func SetResp(ctx *gin.Context, mime string, resp []byte) {
	ctx.Set(RespKey, resp)
	ctx.Set(RespMimeKey, mime)
	ctx.Set(RespRawKey, true)
}

// IsRawResp returns if is raw resp.
func IsRawResp(ctx *gin.Context) bool {
	return ctx.GetBool(RespRawKey)
}

// IsRespSet returns if resp is set.
func IsRespSet(ctx *gin.Context) bool {
	_, ok := ctx.Get(RespKey)
	return ok
}

// SetRedirect sets redirect.
func SetRedirect(ctx *gin.Context, url string) {
	ctx.Set(RedirectKey, url)
}

// IsRedirectSet returns if redirect is set.
func IsRedirectSet(ctx *gin.Context) bool {
	_, ok := ctx.Get(RedirectKey)
	return ok
}

// SetIgnoreAbort sets ignore abort.
func SetIgnoreAbort(ctx *gin.Context) {
	ctx.Set(IgnoreAbortKey, true)
}

// IsIgnoreAbort returns if is ignore abort.
func IsIgnoreAbort(ctx *gin.Context) bool {
	return ctx.GetBool(IgnoreAbortKey)
}

// UnexpectedFailure creates an unexpected error failure.
func UnexpectedFailure() *FailureObj {
	return Failure(ErrorCode(apierrors.UnexpectedError))
}

// MessageResponse is the general Response format.
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents the response of error.
type ErrorResponse struct {
	IsSuccessful bool `json:"success"`
	Error        struct {
		ErrorCode string `json:"error_code"`
	} `json:"error"`
}

// ReviewMotionResponse is the response for voting.
type ReviewMotionResponse struct {
	VoterID    uuid.UUID                      `json:"voter_id"`
	MotionID   uuid.UUID                      `json:"motion_id"`
	Action     types.AuditCommitteeVoteType   `json:"action"`
	MotionType types.AuditCommitteeMotionType `json:"motion_type"`
	Resolution types.AuditCommitteeResolution `json:"resoultion"`
}

// ShortUser is the response for user model.
type ShortUser struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"user_email"`
	FirstName          *string   `json:"first_name"`
	LastName           *string   `json:"last_name"`
	IsFreezed          bool      `json:"is_freezed"`
	IsDisabled         bool      `json:"is_disabled"`
	IsDeleted          bool      `json:"is_deleted"`
	RiskIcons          *string   `json:"risk_icons"`
	NationalityCountry string    `json:"nationality_country"`
	ResidenceCountry   string    `json:"residence_country"`
	CreatedAt          int64     `json:"created_at"`
	UpdatedAt          int64     `json:"updated_at"`
}

// ShortUserFromModel returns short user for some given user.
func ShortUserFromModel(user *exchange.User) *ShortUser {
	return &ShortUser{
		ID:         user.ID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		IsFreezed:  user.IsFreezed,
		IsDisabled: user.IsDisabled,
		IsDeleted:  user.IsUserDeleted(),
		CreatedAt:  ctime.GetMTS(user.CreatedAt),
		UpdatedAt:  ctime.GetMTS(user.UpdatedAt),
	}
}
