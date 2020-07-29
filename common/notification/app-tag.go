package notification

import (
	"crypto/sha256"
	"encoding/base64"

	cobxtypes "github.com/cobinhood/cobinhood-backend/apps/exchange/cobx-types"
)

// AppTag defines config of app notification tag.
type AppTag struct {
	TagSecret string `config:"NotificationTagSecret"`
}

// GetNotificationTag returns tag string for registering the device in 3rd party
// server. This tag string is encrypted with a secret key.
func (a *AppTag) GetNotificationTag(
	userID string, service cobxtypes.ServiceName) string {
	// FIXME(xnum): shouldn't rely on types.
	key := a.TagSecret
	if service != cobxtypes.APICobx {
		key += string(service)
	}
	sha256sum := sha256.Sum256([]byte(userID + key))
	return base64.StdEncoding.EncodeToString(sha256sum[:])
}
