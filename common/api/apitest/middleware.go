package apitest

import (
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"

	"github.com/cobinhood/mochi/common/api/context"
)

// SetUserIDfromRequestHeader assigns user id to app context.
func SetUserIDfromRequestHeader(c *gin.Context) {
	appCtx, err := context.GetAppContext(c)
	if err != nil {
		panic(err)
	}
	userID, _ := uuid.FromString(c.Request.Header.Get("user_id"))
	appCtx.UserID = &userID
	c.Next()
}
