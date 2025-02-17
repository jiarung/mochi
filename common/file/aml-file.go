package file

import (
	"errors"

	"github.com/jiarung/mochi/common/aes"
	apicontext "github.com/jiarung/mochi/common/api/context"
)

// uploadAMLFileDelegate defines the struct that implements the
// EpehemeralFileUploadDelegate.
type uploadAMLFileDelegate struct {
	EphemeralFileUploadDelegate
}

// NewUploadAMLFileDelegate creates a uploadAMLFileDelegate with expire and
// given AES key.
func NewUploadAMLFileDelegate(
	expireSec int, key aes.Key) EphemeralFileUploadDelegate {
	return uploadAMLFileDelegate{
		EphemeralFileUploadDelegate: NewUploadKYCFileDelegate(expireSec, key),
	}
}

// GetEncryptionAESKey returns the AES key for encryption if ShouldEncrypt()
// returns true.
func (d uploadAMLFileDelegate) GetEncryptionAESKey(
	appCtx *apicontext.AppContext) (aes.Key, error) {
	if !appCtx.IsAuthenticated() {
		return nil, errors.New("unauthenticated context")
	}
	return d.EphemeralFileUploadDelegate.GetEncryptionAESKey(appCtx)
}
