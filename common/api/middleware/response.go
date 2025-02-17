package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apicontext "github.com/jiarung/mochi/common/api/context"
	apiutils "github.com/jiarung/mochi/common/api/utils"
	"github.com/jiarung/mochi/common/logging"
)

// ResponseHandler is the middleware for returning error code.
func ResponseHandler(ctx *gin.Context) {
	ctx.Next()

	appCtx, appCtxErr := apicontext.GetAppContext(ctx)
	if appCtxErr != nil {
		logging.NewLoggerTag("api:middleware:ErrorHandler()").Error(
			"Get appCtx error: %v", appCtxErr)
		return
	}

	if apiutils.IsRedirectSet(ctx) {
		ctx.Redirect(http.StatusFound, appCtx.Redirect())
		return
	}

	if appCtx.IsAborted() && !appCtx.IsIgnoreAbort() {
		status, failure := appCtx.Error()
		appCtx.Logger().Error("error_code returned: %s", failure)
		ctx.JSON(status, apiutils.FailureWithTag(failure, appCtx.RequestTag()))
		return
	}

	if !apiutils.IsRespSet(ctx) {
		return
	}

	if apiutils.IsRawResp(ctx) {
		mime, resp := appCtx.Resp()
		ctx.Data(http.StatusOK, mime, resp)
		return
	}

	ctx.JSON(http.StatusOK, appCtx.JSON())
}
