package kyc

import (
	"fmt"

	"github.com/jiarung/gorm"
	"github.com/satori/go.uuid"

	apicontext "github.com/jiarung/mochi/common/api/context"
	models "github.com/jiarung/mochi/models/exchange"
)

// FirstOrCreateKYCDataOfUser gets the first or creates KYC data for user.
func FirstOrCreateKYCDataOfUser(appCtx *apicontext.AppContext, tx *gorm.DB,
	userID uuid.UUID) (*models.KYCData, error) {
	kycDataCount, err := CountKYCDataOfUser(tx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count kyc data. err: %v", err)
	}
	var kycData models.KYCData
	if kycDataCount == 0 {
		kycDataPtr, err := models.NewKYCDataWithAESKeyEnsured(appCtx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to create kyc data with aes key ensured."+
				" err: %v", err)
		}
		kycData = *kycDataPtr
		err = tx.Create(&kycData).Error
		if err != nil {
			return nil, fmt.Errorf("failed to create kyc data. err: %v", err)
		}
	} else if kycDataCount == 1 {
		err = tx.Set("gorm:query_option", "FOR UPDATE").
			Model(&models.KYCData{}).
			Where("user_id = ?", userID).
			First(&kycData).Error
		if err != nil {
			return nil, fmt.Errorf("failed to find kyc data. err: %v", err)
		}
	} else {
		return nil, fmt.Errorf("strange kyc data count. [%d]", kycDataCount)
	}

	return &kycData, nil
}

// CountKYCDataOfUser returns the number of KYC data rows in DB.
func CountKYCDataOfUser(tx *gorm.DB, userID uuid.UUID) (int, error) {
	var kycDataCount int
	err := tx.Model(&models.KYCData{}).
		Where("user_id = ?", userID).
		Count(&kycDataCount).Error
	return kycDataCount, err
}
