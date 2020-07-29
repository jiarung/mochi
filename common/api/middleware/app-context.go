package middleware

import (
	"errors"
	"net/http"
	"reflect"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	cobxtypes "github.com/jiarung/mochi/apps/exchange/cobx-types"
	"github.com/jiarung/mochi/cache"
	"github.com/jiarung/mochi/common"
	apicontext "github.com/jiarung/mochi/common/api/context"
	apiutils "github.com/jiarung/mochi/common/api/utils"
	"github.com/jiarung/mochi/common/logging"
	"github.com/jiarung/mochi/database"
	"github.com/jiarung/mochi/infra/api/middleware/logger"
)

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// AppHandlerFunc defines handler with app context type.
type AppHandlerFunc func(*apicontext.AppContext)

// RequireAppContext returns a wrapper for passing app context.
func RequireAppContext(
	serviceName cobxtypes.ServiceName, fn AppHandlerFunc) func(*gin.Context) {
	fnName := getFunctionName(fn)

	return func(ctx *gin.Context) {
		var appCtx *apicontext.AppContext
		var err error

		// Add ddtracer span.
		if common.TracerEnabled() {
			// Create our span and patch it to the context for downstream.
			span := tracer.StartSpan("gin.request",
				tracer.ServiceName(string(serviceName)),
				tracer.ResourceName(fnName),
			)
			defer func() {
				var spanErr error
				var code int
				var failure *apiutils.FailureObj
				defer func() {
					// Setup metadata.
					span.SetTag(ext.HTTPMethod, ctx.Request.Method)
					span.SetTag(ext.HTTPURL, ctx.Request.URL.Path)

					if code != 0 {
						span.SetTag(ext.HTTPCode, strconv.Itoa(code))
					}
					// Set any error information.
					if failure != nil {
						span.SetTag("appctx.errors", failure.String()) // set all errors
					}

					span.Finish(tracer.WithError(spanErr))
				}()

				if err != nil {
					code = http.StatusInternalServerError
					spanErr = err
					return
				}

				if apiutils.IsRedirectSet(ctx) {
					code = http.StatusFound
					return
				}

				if appCtx.IsAborted() && !appCtx.IsIgnoreAbort() {
					code, failure = appCtx.Error()
					spanErr = errors.New(failure.String()) // but use the first for standard fields
					return
				}

				if !apiutils.IsRespSet(ctx) {
					return
				}

				if apiutils.IsRawResp(ctx) {
					code = http.StatusOK
					return
				}

			}()
		}
		appCtx, err = apicontext.GetAppContext(ctx)
		if err != nil {
			logging.NewLoggerTag(ctx.GetString(logging.LabelTag)).Error(
				"%s: Can't get AppContext: %v", fnName, err)
			ctx.Abort()
			return
		}
		appCtx.Logger().SetLabel(logging.LabelApp, fnName)
		fn(appCtx)
	}
}

// AppContextMiddleware create a app context which contains
// various clients and pass through handlers with gin.Context
func AppContextMiddleware(serviceName cobxtypes.ServiceName) func(*gin.Context) {
	return func(ctx *gin.Context) {
		var err error
		logger := logger.Get(ctx)
		logger.SetLabel(logging.LabelAuthMethod, "jwt")

		// Compose & set application context
		appCtx, err := apicontext.NewAppCtx(
			ctx,
			logger,
			database.GetDB(database.Default),
			cache.GetRedis(),
		)
		if err != nil {
			logger.Error("context.NewAppCtx() failed. err(%s)", err)
			ctx.Abort()
			return
		}
		appCtx.SetServiceName(serviceName)
	}
}
