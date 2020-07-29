package middleware

import (
	"github.com/gin-gonic/gin"

	apicontext "github.com/jiarung/mochi/common/api/context"
	apierrors "github.com/jiarung/mochi/common/api/errors"
	"github.com/jiarung/mochi/infra/api/utils"
)

// PrivilegedIPRequired check if the IP of http request is privileged ip.
func PrivilegedIPRequired(ctx *gin.Context) {
	appCtx, err := apicontext.GetAppContext(ctx)
	if err != nil {
		panic(err)
	}
	logger := appCtx.Logger()

	if !appCtx.IsPrivilegedIP() {
		userID := ""
		if appCtx.UserID != nil {
			userID = appCtx.UserID.String()
		}
		logger.Notice(
			"!appCtx.IsPrivilegedIP(). appCtx.RequestIP(%s), "+
				"PrivilegedIPRange(%s),userID(%s)",
			appCtx.RequestIP,
			utils.PrivilegedIPRange(),
			userID)
		appCtx.SetError(apierrors.NotPrivilegedIP)
		return
	}
}
