package file

import (
	"github.com/satori/go.uuid"

	apicontext "github.com/jiarung/mochi/common/api/context"
	apierrors "github.com/jiarung/mochi/common/api/errors"
	models "github.com/jiarung/mochi/models/exchange"
)

// DownloadCampaignManualReviewFile downloads file.
func DownloadCampaignManualReviewFile(
	appCtx *apicontext.AppContext,
	event string, submitID uuid.UUID, pictureID uuid.UUID) ([]byte, string) {
	logger := appCtx.Logger()
	db := appCtx.DB

	submit := &models.ManualReviewSubmit{}
	err := db.Model(&models.ManualReviewSubmit{}).
		Where("event = ?", event).
		Where("id = ?", submitID).
		Find(submit).Error
	if err != nil {
		logger.Error("cannot find submit. err: %s", err)
		return nil, apierrors.ResourceNotFound
	}

	pictureIDStr := pictureID.String()
	for _, v := range submit.Data {
		if v.IsFile && v.Text == pictureIDStr {
			file := &models.GCSFile{}
			err = db.Model(&models.GCSFile{}).
				Where("id = ?", v.Text).
				Find(file).Error
			if err != nil {
				logger.Error("get gcs file failed. err: %s", err)
				return nil, apierrors.UnexpectedError
			}
			bytes, err := file.ReadGCSObject(appCtx)
			if err != nil {
				logger.Error("read gcs file failed. err: %s", err)
				return nil, apierrors.UnexpectedError
			}
			if len(bytes) == 0 {
				logger.Error("find empty file.")
				return nil, apierrors.ResourceNotFound
			}

			return bytes, ""
		}
	}

	logger.Error("cannot find file.")
	return nil, apierrors.ResourceNotFound
}
