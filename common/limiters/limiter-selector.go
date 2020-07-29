package limiters

import (
	"github.com/satori/go.uuid"

	"github.com/jiarung/mochi/common/logging"
	"github.com/jiarung/gorm"
)

// SimpleLimiterSelector is just a wrapper for Limiter
type SimpleLimiterSelector struct {
	limiter Limiter
}

// NewSimpleLimiterSelector create a new simpleLimiterSelector
func NewSimpleLimiterSelector(limit int64, seconds int) *SimpleLimiterSelector {
	return &SimpleLimiterSelector{
		limiter: NewLimiter(limit, seconds),
	}
}

// SelectLimiter ignore all parameters, return the same limiter every time
func (s *SimpleLimiterSelector) SelectLimiter(db *gorm.DB,
	userID *uuid.UUID) Limiter {
	return s.limiter
}

// VIPLimiterSelector select limiter by vip rule
type VIPLimiterSelector struct {
	defaultLimiter    Limiter
	getVIPLimiterFunc func(*gorm.DB, uuid.UUID) (Limiter, error)
}

// NewVIPLimiterSelector create a vipLimiterSelector
func NewVIPLimiterSelector(defaultLimit int64, defaultSeconds int,
	getVIPLimiterFunc func(*gorm.DB,
		uuid.UUID) (Limiter, error)) *VIPLimiterSelector {
	return &VIPLimiterSelector{
		defaultLimiter:    NewLimiter(defaultLimit, defaultSeconds),
		getVIPLimiterFunc: getVIPLimiterFunc,
	}
}

// SelectLimiter select a limiter by user's rate limit related tags
func (s *VIPLimiterSelector) SelectLimiter(db *gorm.DB,
	userID *uuid.UUID) Limiter {
	limiter, err := s.getVIPLimiterFunc(db, *userID)
	if err != nil {
		logger := logging.NewLoggerTag("limiter-selector")
		logger.Error("get vip limiter fail, fallback to default limiter: %v",
			err)
		return s.defaultLimiter
	}
	return limiter
}
