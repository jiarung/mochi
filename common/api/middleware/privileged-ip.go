package middleware

import (
	"github.com/gin-gonic/gin"

	apicontext "github.com/cobinhood/cobinhood-backend/common/api/context"
	apierrors "github.com/cobinhood/cobinhood-backend/common/api/errors"
	"github.com/cobinhood/cobinhood-backend/infra/api/utils"
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
