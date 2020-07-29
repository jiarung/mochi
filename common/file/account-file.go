package file

import (
	"github.com/cobinhood/mochi/common/aes"
	"github.com/cobinhood/mochi/common/api/middleware"
)

// UploadAccountFileHandler returns the account ephemeral file upload handler.
func UploadAccountFileHandler(
	expireSec int, key aes.Key) middleware.AppHandlerFunc {
	return UploadEphemeralFileHandler(
		NewUploadAccountFileDelegate(expireSec, key))
}

// NewUploadAccountFileDelegate returns a account file upload delegate.
func NewUploadAccountFileDelegate(
	expireSec int, key aes.Key) EphemeralFileUploadDelegate {
	return uploadAccountFileDelegate{
		ephemeralFileUploadDelegateBase: ephemeralFileUploadDelegateBase{
			expireSec: expireSec,
			key:       key,
		},
	}
}

// uploadAccountFileDelegate defines the struct that implements the
// EpehemeralFileUploadDelegate.
type uploadAccountFileDelegate struct {
	ephemeralFileUploadDelegateBase
}

// FileName returns the field name of multipart/form-data POST request.
func (d uploadAccountFileDelegate) FileName() string {
	return "account_info"
}
