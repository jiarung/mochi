package file

import (
	"net/http"

	"github.com/satori/go.uuid"

	"github.com/cobinhood/mochi/common/aes"
	apicontext "github.com/cobinhood/mochi/common/api/context"
	apierrors "github.com/cobinhood/mochi/common/api/errors"
	"github.com/cobinhood/mochi/common/api/middleware"
	"github.com/cobinhood/mochi/models/exchange"
)

// SuspicionInvestigationFileHandler returns the suspicion investigation
// ephemeral file upload handler.
func SuspicionInvestigationFileHandler(
	expireSec int, key aes.Key) middleware.AppHandlerFunc {
	return UploadEphemeralFileHandler(
		NewSuspicionInvestigationFileDelegate(expireSec, key))
}

// NewSuspicionInvestigationFileDelegate returns a new suspicion investigation
// file delegate.
func NewSuspicionInvestigationFileDelegate(
	expireSec int, key aes.Key) EphemeralFileUploadDelegate {
	return suspicionInvestigationFileDelegate{
		ephemeralFileUploadDelegateBase: ephemeralFileUploadDelegateBase{
			expireSec: expireSec,
			key:       key,
		},
	}
}

// suspicionInvestigationFileDelegate defines the struct that implements the
// EpehemeralFileUploadDelegate.
type suspicionInvestigationFileDelegate struct {
	ephemeralFileUploadDelegateBase
}

// FileName returns the field name of multipart/form-data POST request.
func (d suspicionInvestigationFileDelegate) FileName() string {
	return "suspicion_investigation"
}

// DownloadSuspicionInvestigationFile downloads file by ID.
// /v1/crm/suspicion/file/:file_id [GET]
func DownloadSuspicionInvestigationFile(appCtx *apicontext.AppContext) {
	logger := appCtx.Logger()
	if !appCtx.ValidateAuthenticated() {
		return
	}

	fileID := uuid.FromStringOrNil(appCtx.Param("file_id"))
	if uuid.Equal(uuid.Nil, fileID) {
		logger.Error("invalid file id")
		appCtx.SetError(apierrors.ParameterError)
		return
	}

	var (
		file      exchange.GCSFile
		fileBytes []byte
	)
	if err := appCtx.DB.Model(&exchange.GCSFile{}).
		Where("id = ?", fileID).
		First(&file).Error; err != nil {
		logger.Error("db error: %v", err)
		appCtx.SetError(apierrors.DBError)
		return
	}
	fileBytes, err := file.ReadGCSObject(appCtx)
	if err != nil {
		logger.Error("failed to read GCS object: %v", err)
		appCtx.Abort()
		return
	}

	fileContentType := http.DetectContentType(fileBytes)
	appCtx.SetResp(fileContentType, fileBytes)
}
