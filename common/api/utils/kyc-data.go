package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cobinhood/gorm"
	"github.com/satori/go.uuid"

	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/cache/keys"
	"github.com/cobinhood/mochi/models/exchange"
)

// Define cache expiration time.
const (
	UserNationalityCountryCacheExpireTime = 604800
	UserResidenceCountryCacheExpireTime   = 604800
	UserAgeCacheExpireTime                = 604800
)

// GetKYCDataNationalityCountry returns the nationality country of the user from
// KYC data.
func GetKYCDataNationalityCountry(ctx context.Context,
	db *gorm.DB, redis *cache.Redis, userID uuid.UUID, now time.Time) (
	string, string, error) {

	key := keys.GetUserNationalityCacheKey(userID)
	get := func() (string, string, error) {
		result, err := redis.GetString(key)
		if err != nil {
			return result, fmt.Sprintf("redis GET error: %v", err), err
		}
		return result, "", nil
	}

	if result, log, err := get(); err == nil {
		return result, "", nil
	} else if cache.ParseCacheErrorCode(err) != cache.ErrNilKey {
		return "", log, err
	}

	if log, err := CacheKYCData(ctx, db, redis, userID, now); err != nil {
		return "", log, err
	}
	return get()
}

// GetKYCDataResidenceCountry returns the residence country of the user from
// KYC data.
func GetKYCDataResidenceCountry(ctx context.Context,
	db *gorm.DB, redis *cache.Redis, userID uuid.UUID, now time.Time) (
	string, string, error) {

	key := keys.GetUserResidenceCacheKey(userID)
	get := func() (string, string, error) {
		result, err := redis.GetString(key)
		if err != nil {
			return result, fmt.Sprintf("redis GET error: %v", err), err
		}
		return result, "", nil
	}

	if result, log, err := get(); err == nil {
		return result, "", nil
	} else if cache.ParseCacheErrorCode(err) != cache.ErrNilKey {
		return "", log, err
	}

	if log, err := CacheKYCData(ctx, db, redis, userID, now); err != nil {
		return "", log, err
	}
	return get()
}

// GetKYCDataAge returns the age of the user from KYC data.
func GetKYCDataAge(ctx context.Context,
	db *gorm.DB, redis *cache.Redis, userID uuid.UUID, now time.Time) (
	uint8, bool, string, error) {

	key := keys.GetUserAgeKey(userID)
	get := func() (uint8, bool, string, error) {
		rawResult, err := redis.GetString(key)
		if err != nil {
			return 0, false, fmt.Sprintf("redis GET error: %v", err), err
		}

		// If the user doesn't have KYC data, the result is an empty string.
		if rawResult == "" {
			return 0, false, "", nil
		}

		// If the user has KYC data, the result is an uint8 converted to string.
		result, err := strconv.ParseUint(rawResult, 10, 8)
		if err != nil {
			return 0, false, fmt.Sprintf(
				"%s is not an uint8: %v", strconv.Quote(rawResult), err), err
		}
		return uint8(result), true, "", nil
	}

	if result, valid, log, err := get(); err == nil {
		return result, valid, "", nil
	} else if cache.ParseCacheErrorCode(err) != cache.ErrNilKey {
		return 0, false, log, err
	}

	if log, err := CacheKYCData(ctx, db, redis, userID, now); err != nil {
		return 0, false, log, err
	}
	return get()
}

// CacheKYCData retrieves KYC data from the database, decrypts them, and stores
// frequently used fields in redis.
func CacheKYCData(ctx context.Context, db *gorm.DB, redis *cache.Redis,
	userID uuid.UUID, now time.Time) (string, error) {

	set := func(info *exchange.BasicInformation) (string, error) {
		var err error
		var nationalityCountryKey string
		var nationalityCountryValue string
		var residenceCountryKey string
		var residenceCountryValue string
		var ageKey string
		var ageValue string

		nationalityCountryKey = keys.GetUserNationalityCacheKey(userID)
		residenceCountryKey = keys.GetUserResidenceCacheKey(userID)
		ageKey = keys.GetUserAgeKey(userID)

		if info != nil {
			nationalityCountryValue = info.NationalityCountry
			residenceCountryValue = info.ResidenceCountry

			nowYear, nowMonth, nowDay := now.UTC().Date()
			birthYear, birthMonth, birthDay := time.Time(info.Birthday).Date()
			age := nowYear - birthYear
			if nowMonth < birthMonth {
				age--
			} else if nowMonth == birthMonth {
				if nowDay < birthDay {
					age--
				}
			}
			if age < 0 || age >= 256 /* uint8 */ {
				birthdayJSON, _ := info.Birthday.MarshalJSON()
				return fmt.Sprintf(
						"invalid age %d (birthday %s)", age, birthdayJSON),
					strconv.ErrRange
			}
			ageValue = strconv.FormatUint(uint64(age), 10)
		}

		err = redis.Set(nationalityCountryKey, nationalityCountryValue,
			UserNationalityCountryCacheExpireTime)
		if err != nil {
			return fmt.Sprintf(
				"failed to add nationality country to cache. err(%v)", err), err
		}
		err = redis.Set(residenceCountryKey, residenceCountryValue,
			UserResidenceCountryCacheExpireTime)
		if err != nil {
			return fmt.Sprintf(
				"failed to add residence country to cache. err(%v)", err), err
		}
		err = redis.Set(ageKey, ageValue, UserAgeCacheExpireTime)
		if err != nil {
			return fmt.Sprintf(
				"failed to add age to cache. err(%v)", err), err
		}
		return "", nil
	}

	// Query from GCP.
	var kycData exchange.KYCData
	result := db.First(&kycData, "user_id = ?", userID)
	if result.RecordNotFound() {
		return set(nil)
	} else if result.Error != nil {
		return fmt.Sprintf("failed to get kyc data. err(%v)", result.Error),
			result.Error
	}

	var basicInformation exchange.BasicInformation
	err := kycData.ReadToDataBlock(ctx, &basicInformation)
	if err != nil {
		return fmt.Sprintf("failed to read basic info data block. err(%v)", err),
			err
	}
	err = basicInformation.ValidateFormat()
	if err != nil {
		return fmt.Sprintf("basic information validate failure. err(%v)", err),
			err
	}

	// Add to cache.
	return set(&basicInformation)
}
