package file

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/jiarung/gorm"

	"github.com/jiarung/mochi/cache"
	"github.com/jiarung/mochi/common/aes"
	apicontext "github.com/jiarung/mochi/common/api/context"
	apierrors "github.com/jiarung/mochi/common/api/errors"
	"github.com/jiarung/mochi/common/api/middleware"
	models "github.com/jiarung/mochi/models/exchange"
	"github.com/jiarung/mochi/types"
)

const (
	// JPEGType represents the "image/jpeg" content type
	JPEGType = "image/jpeg"
	// PNGType represents the "image/png" content type
	PNGType = "image/png"
	// PDFType represents the "application/pdf" content type
	PDFType = "application/pdf"
	// PlainTextUTF8Type represents the "text/plain" content type
	PlainTextUTF8Type = "text/plain; charset=utf-8"
)

const (
	// KB stands for 1024 (bytes)
	KB = 1024
	// MB stands for 1024 KB
	MB = 1024 * KB
)

type uploadResponse struct {
	CacheKey string `json:"cache_key"`
}

// EphemeralFileUploadDelegate defines the interface that configures the
// ephemeral file upload handler.
type EphemeralFileUploadDelegate interface {
	// SizeLimit returns the size limit of file.
	SizeLimit() int64
	// FileName returns the field name of multipart/form-data POST request.
	FileName() string
	// ValidContentTypes returns the allowed file types.
	ValidContentTypes() []string
	// ExpireSec returns the expiration second of the cached file.
	ExpireSec() int
	// ShouldEncrypt returns if the file should be encrypted before cached.
	ShouldEncrypt() bool
	// GetEncryptionAESKey returns the AES key for encryption if ShouldEncrypt()
	// returns true.
	GetEncryptionAESKey(appCtx *apicontext.AppContext) (aes.Key, error)
	// ShouldSetAccessKey returns if the cached file should be set with an
	// access key.
	ShouldSetAccessKey() bool
	// GetAccessKey returns the cached file access key if ShouldSetAccessKey
	// return true.
	GetAccessKey(appCtx *apicontext.AppContext) (*string, error)
}

// ephemeralFileUploadDelegateBase defines the basic struct that implements the
// EpehemeralFileUploadDelegate.
type ephemeralFileUploadDelegateBase struct {
	expireSec int
	key       aes.Key
}

// SizeLimit returns the size limit of file.
func (e ephemeralFileUploadDelegateBase) SizeLimit() int64 {
	return 5 * MB
}

// FileName returns the field name of multipart/form-data POST request.
func (e ephemeralFileUploadDelegateBase) FileName() string {
	return "temp"
}

// ValidContentTypes returns the allowed file types.
func (e ephemeralFileUploadDelegateBase) ValidContentTypes() []string {
	return []string{JPEGType, PNGType, PDFType}
}

// ExpireSec returns the expiration second of the cached file.
func (e ephemeralFileUploadDelegateBase) ExpireSec() int {
	return e.expireSec
}

// ShouldEncrypt returns if the file should be encrypted before cached.
func (e ephemeralFileUploadDelegateBase) ShouldEncrypt() bool {
	return true
}

// GetEncryptionAESKey returns the AES key for encryption if ShouldEncrypt()
// returns true.
func (e ephemeralFileUploadDelegateBase) GetEncryptionAESKey(
	appCtx *apicontext.AppContext) (aes.Key, error) {
	return e.key, nil
}

// ShouldSetAccessKey returns if the cached file should be set with an
// access key.
func (e ephemeralFileUploadDelegateBase) ShouldSetAccessKey() bool {
	return true
}

// GetAccessKey returns the cached file access key if ShouldSetAccessKey
// return true.
func (e ephemeralFileUploadDelegateBase) GetAccessKey(
	appCtx *apicontext.AppContext) (
	*string, error) {
	if !appCtx.IsAuthenticated() {
		return nil, errors.New("unauthorized app ctx")
	}
	accessKey := appCtx.UserID.String()
	return &accessKey, nil
}

// UploadEphemeralFileHandler return a `AppHandlerFunc` that handles the
// multipart/form-data POST request with specified `sizeLimit`, `filename`,
// `validContentTypes`. The file uploaded will be set in to redis with `expireSec`
// expiration time.
// The file will only be encrypted if getAESKey is not nil.
// The access key will be set if getAccessKey is not nil.
func UploadEphemeralFileHandler(
	delegate EphemeralFileUploadDelegate) middleware.AppHandlerFunc {
	return func(appCtx *apicontext.AppContext) {
		logger := appCtx.Logger()

		// Source
		fileHeader, err := appCtx.FormFile(delegate.FileName())
		if err != nil {
			logger.Error("get form err: %s", err)
			appCtx.SetError(apierrors.InvalidFilename)
			return
		}

		if fileHeader.Size > delegate.SizeLimit() {
			logger.Info("file too large. %d bytes", fileHeader.Size)
			appCtx.SetError(apierrors.FileTooLarge)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			logger.Error("open file err: %s", err)
			appCtx.Abort()
			return
		}
		defer file.Close()

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(file)
		if err != nil {
			logger.Error("read file err: %s", err)
			appCtx.Abort()
			return
		}

		fileBytes := buf.Bytes()
		fileContentType := http.DetectContentType(fileBytes)
		logger.Info("file uploaded. size: %d bytes. filename: %s. content-type: %s.",
			fileHeader.Size, fileHeader.Filename, fileContentType)

		if !isValidContentType(fileContentType, delegate.ValidContentTypes()) {
			logger.Info("invalid file content type(%s) other than %s",
				fileContentType, delegate.ValidContentTypes())
			appCtx.SetError(apierrors.InvalidContentType)
			return
		}

		if delegate.ShouldEncrypt() {
			aesKey, err := delegate.GetEncryptionAESKey(appCtx)
			if err != nil {
				logger.Error("failed to get AES key. err(%s)", err)
				appCtx.Abort()
				return
			}
			encrypted, err := aes.CBCEncrypt(aesKey, fileBytes)
			if err != nil {
				logger.Error("failed to encrypt file. err(%s)", err)
				appCtx.Abort()
				return
			}
			fileBytes = encrypted
		}

		var accessKey *string
		if delegate.ShouldSetAccessKey() {
			accessKey, err = delegate.GetAccessKey(appCtx)
			if err != nil {
				logger.Error("failed to get ephemeral file access key. err(%s)",
					err)
				appCtx.Abort()
				return
			}
		}

		cacheKey, err := cache.SetEphemeralFile(accessKey, fileBytes, delegate.ExpireSec())
		if err != nil {
			logger.Error("failed to set ephemeral file. err(%s)",
				err)
			appCtx.Abort()
			return
		}
		appCtx.SetJSON(uploadResponse{
			CacheKey: cacheKey,
		})
	}
}

// EphemeralFileDownloadDelegate defines the interface that configures the
// ephemeral file download handler.
type EphemeralFileDownloadDelegate interface {
	// PathParameter returns the path parameter, `../../:<pathParamKey>`, as
	// redis key for the cache file.
	GetPathParameter() string
	// ShouldDecrypt returns if the file should be decrypted before downloaded.
	ShouldDecrypt() bool
	// GetDecryptionAESKey returns the AES key for decryption if ShouldDecrypt()
	// returns true.
	GetDecryptionAESKey(appCtx *apicontext.AppContext) (aes.Key, error)
	// GetAccessKey returns the cached file access key if ShouldSetAccessKey
	// return true.
	GetAccessKey(appCtx *apicontext.AppContext) (*string, error)
}

// DownloadEphemeralFileHandler downloads ephemeral file from redis with path
// parameter `../../:<pathParamKey>` as redis key
// DownloadEphemeralFileHandler return a `AppHandlerFunc` that downloads the
// ephemeral file from redis with path parameter `../../:<pathParamKey>` as the
// file key .
func DownloadEphemeralFileHandler(
	delegate EphemeralFileDownloadDelegate) middleware.AppHandlerFunc {
	return func(appCtx *apicontext.AppContext) {
		logger := appCtx.Logger()

		cacheKey := appCtx.Param(delegate.GetPathParameter())

		accessKey, err := delegate.GetAccessKey(appCtx)
		if err != nil {
			logger.Error("failed to get ephemeral file access key. err(%s)",
				err)
			appCtx.Abort()
			return
		}

		b, err := cache.GetEphemeralFile(accessKey, cacheKey)
		if err != nil {
			logger.Error("failed to get from redis. err(%s)", err)
			appCtx.Abort()
			return
		}

		if !delegate.ShouldDecrypt() {
			fileContentType := http.DetectContentType(b)
			appCtx.SetResp(fileContentType, b)
			return
		}

		aesKey, err := delegate.GetDecryptionAESKey(appCtx)
		if err != nil {
			logger.Error("failed to get AES key. err(%s)", err)
			appCtx.Abort()
			return
		}

		decrypted, err := aes.CBCDecrypt(aesKey, b)
		if err != nil {
			logger.Error("failed to decrypt file. err(%s)", err)
			appCtx.Abort()
			return
		}

		fileContentType := http.DetectContentType(decrypted)
		appCtx.SetResp(fileContentType, decrypted)
	}
}

func isValidContentType(contentType string, validContentTypes []string) bool {
	validContentTypeMap := map[string]struct{}{}
	for _, validContentType := range validContentTypes {
		validContentTypeMap[validContentType] = struct{}{}
	}
	_, validType := validContentTypeMap[contentType]
	return validType
}

// CreateGCSFileWithCachedFile creates a GCSFile using data in redis.
func CreateGCSFileWithCachedFile(appCtx *apicontext.AppContext,
	cacheKey string, ephemeralFileUploadDelegate EphemeralFileUploadDelegate,
	gcsFileType types.GCSFileType, tx *gorm.DB) (*models.GCSFile, error) {
	var accessKey *string
	if ephemeralFileUploadDelegate.ShouldSetAccessKey() {
		var err error
		accessKey, err = ephemeralFileUploadDelegate.GetAccessKey(appCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to get access key. err: %v", err)
		}
	}
	file, err := cache.GetEphemeralFile(accessKey, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get picture from redis. err(%s)", err)
	}
	aesKey, err := ephemeralFileUploadDelegate.GetEncryptionAESKey(appCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption aes key. err(%v)", err)
	}
	decrypted, err := aes.CBCDecrypt(aesKey, file)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file. err(%v)", err)
	}
	gcsFile, err :=
		models.NewGCSFileWithAESKeyEnsured(appCtx, gcsFileType)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create gcs file with aes key ensured. err: %v", err)
	}
	if err = tx.Create(gcsFile).Error; err != nil {
		return nil, fmt.Errorf("failed to create gcs file row. err: %v", err)
	}
	if err = gcsFile.WriteGCSObject(appCtx, decrypted); err != nil {
		return nil, fmt.Errorf("failed to write gcs object. err: %v", err)
	}
	return gcsFile, nil
}
