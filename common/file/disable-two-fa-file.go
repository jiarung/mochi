package file

import (
	"github.com/cobinhood/cobinhood-backend/common/aes"
	apicontext "github.com/cobinhood/cobinhood-backend/common/api/context"
	"github.com/cobinhood/cobinhood-backend/common/api/middleware"
)

// UploadDisableTwoFAFileHandler returns the disable 2FA ephemeral file upload
// handler.
func UploadDisableTwoFAFileHandler(
	expireSec int, key aes.Key) middleware.AppHandlerFunc {
	return UploadEphemeralFileHandler(
		NewUploadDisableTwoFAFileDelegate(expireSec, key))
}

// NewUploadDisableTwoFAFileDelegate returns a new disable two fa file upload
// delegate.
func NewUploadDisableTwoFAFileDelegate(
	expireSec int, key aes.Key) EphemeralFileUploadDelegate {
	return uploadDisableTwoFAFileDelegate{
		ephemeralFileUploadDelegateBase: ephemeralFileUploadDelegateBase{
			expireSec: expireSec,
			key:       key,
		},
	}
}

// uploadDisableTwoFAFileDelegate defines the struct that implements the
// EpehemeralFileUploadDelegate.
type uploadDisableTwoFAFileDelegate struct {
	ephemeralFileUploadDelegateBase
}

// FileName returns the field name of multipart/form-data POST request.
func (d uploadDisableTwoFAFileDelegate) FileName() string {
	return "disable_two_fa"
}

// ShouldEncrypt returns if the file should be encrypted before cached.
func (d uploadDisableTwoFAFileDelegate) ShouldEncrypt() bool {
	return true
}

// ShouldSetAccessKey returns if the cached file should be set with an
// access key.
func (d uploadDisableTwoFAFileDelegate) ShouldSetAccessKey() bool {
	return false
}

// GetAccessKey returns the cached file access key if ShouldSetAccessKey
// return true.
func (d uploadDisableTwoFAFileDelegate) GetAccessKey(appCtx *apicontext.AppContext) (
	*string, error) {
	return nil, nil
}
