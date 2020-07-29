package file

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/satori/go.uuid"

	"github.com/cobinhood/mochi/common/aes"
	apicontext "github.com/cobinhood/mochi/common/api/context"
	apierrors "github.com/cobinhood/mochi/common/api/errors"
	"github.com/cobinhood/mochi/common/api/middleware"
	"github.com/cobinhood/mochi/common/kyc"
	models "github.com/cobinhood/mochi/models/exchange"
)

// UploadKYCFileHandler returns the kyc ephemeral file upload handler.
func UploadKYCFileHandler(expireSec int, key aes.Key) middleware.AppHandlerFunc {
	return UploadEphemeralFileHandler(
		NewUploadKYCFileDelegate(expireSec, key))
}

// NewUploadKYCFileDelegate creates an uploadKYCFileDelegate with expire and AES
// encryption key.
func NewUploadKYCFileDelegate(
	expireSec int, key aes.Key) EphemeralFileUploadDelegate {
	return uploadKYCFileDelegate{
		ephemeralFileUploadDelegateBase: ephemeralFileUploadDelegateBase{
			expireSec: expireSec,
			key:       key,
		},
	}
}

// uploadKYCFileDelegate defines the struct that implements the
// EpehemeralFileUploadDelegate.
type uploadKYCFileDelegate struct {
	ephemeralFileUploadDelegateBase
}

// FileName returns the field name of multipart/form-data POST request.
func (d uploadKYCFileDelegate) FileName() string {
	return "kyc_info"
}

// GetEncryptionAESKey returns the AES key for encryption if ShouldEncrypt()
// returns true.
func (d uploadKYCFileDelegate) GetEncryptionAESKey(appCtx *apicontext.AppContext) (
	aes.Key, error) {
	if d.key != nil {
		return d.key, nil
	}
	if appCtx.DB == nil {
		return nil, errors.New("nil db")
	}
	userID, err := appCtx.GetUserID()
	if err != nil {
		return nil, err
	}

	kycData, err := kyc.FirstOrCreateKYCDataOfUser(appCtx, appCtx.DB, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create kyc data. err: %v", err)
	}

	key, err := kycData.GetAESKey(appCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AES key of kyc data. err: %v", err)
	}
	return key, nil
}

// DownloadKYCInfo handles [GET] v1/kyc/pictures/:picture_id.
func DownloadKYCInfo(appCtx *apicontext.AppContext) {
	logger := appCtx.Logger()

	if !appCtx.ValidateAuthenticated() {
		return
	}

	pictureID := appCtx.Param("picture_id")
	pictureUUID, err := uuid.FromString(pictureID)
	if err != nil {
		logger.Error("invalid picture id. err: %v", err)
		appCtx.SetError(apierrors.ResourceNotFound)
		return
	}

	var kycData models.KYCData
	result := appCtx.DB.Model(&models.KYCData{}).
		Where("user_id = ?", *appCtx.UserID).
		First(&kycData)
	if result.RecordNotFound() {
		logger.Error("no kyc datat found.")
		appCtx.SetError(apierrors.ResourceNotFound)
		return
	}

	if result.Error != nil {
		logger.Error("failed to find kyc data. err: %v", result.Error)
		appCtx.Abort()
		return
	}

	for ID, file := range map[*uuid.UUID]models.GCSFile{
		kycData.ProofOfIdentityFrontFileID:          kycData.ProofOfIdentityFrontFile,
		kycData.ProofOfIdentityBackFileID:           kycData.ProofOfIdentityBackFile,
		kycData.SelfieWithPhotoIdentificationFileID: kycData.SelfieWithPhotoIdentificationFile,
		kycData.ProofOfResidenceFileID:              kycData.ProofOfResidenceFile,
		kycData.KYCFormFileID:                       kycData.KYCFormFile,
	} {
		if ID != nil && uuid.Equal(*ID, pictureUUID) {
			err = appCtx.DB.Model(&models.GCSFile{}).
				Where("id = ?", *ID).
				First(&file).Error
			if err != nil {
				logger.Error("failed to find gcs file. err: %v", err)
				appCtx.Abort()
				return
			}
			b, err := file.ReadGCSObject(appCtx)
			if err != nil {
				logger.Error("failed to read gcs object. err: %v", err)
				appCtx.Abort()
				return
			}
			fileContentType := http.DetectContentType(b)
			appCtx.SetResp(fileContentType, b)
			return
		}
	}

	logger.Error("no kyc data file id match.")
	appCtx.SetError(apierrors.ResourceNotFound)
}
