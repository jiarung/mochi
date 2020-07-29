package middleware

import (
	"github.com/gin-gonic/gin"

	apicontext "github.com/cobinhood/mochi/common/api/context"
	apierrors "github.com/cobinhood/mochi/common/api/errors"
)

// OnlyIPRequired check if the IP of http request is correct.
func OnlyIPRequired(ip string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appCtx, err := apicontext.GetAppContext(ctx)
		if err != nil {
			panic(err)
		}
		logger := appCtx.Logger()

		if appCtx.RequestIP != ip {
			logger.Notice(
				"appCtx.RequestIP(%s) OnlyIPRequired(%s)",
				appCtx.RequestIP,
				ip)
			appCtx.SetError(apierrors.NotPrivilegedIP)
			return
		}
	}
}
