package utils

import (
	"os"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"

	"github.com/cobinhood/cobinhood-backend/common/config/misc"
)

// System environment enum.
const (
	Production = iota
	Staging
	Development
	Stress
	LocalDevelopment
	CI
	FullnodeCluster
)

// System environment string tag.
const (
	EnvProductionTag       = "prod"
	EnvStagingTag          = "staging"
	EnvDevelopmentTag      = "dev"
	EnvStressTag           = "stress"
	EnvLocalDevelopmentTag = "localdev"
	EnvFullnodeTag         = "fullnode"
)

// Environment returns current system environment.
func Environment() int {
	environ := misc.ServerEnvironment()
	switch environ {
	case EnvProductionTag:
		return Production
	case EnvStagingTag:
		return Staging
	case EnvDevelopmentTag:
		return Development
	case EnvStressTag:
		return Stress
	case EnvLocalDevelopmentTag:
		return LocalDevelopment
	case EnvFullnodeTag:
		return FullnodeCluster
	default:
		if _, ok := os.LookupEnv("CI"); ok {
			return CI
		}
		panic("Unknown environment.")
	}
}

// IsProduction returns true if we are in production mode.
func IsProduction() bool {
	environ := misc.ServerEnvironment()
	return environ == "prod"
}

// IsCI returns true if we are in CI mode.
func IsCI() bool {
	if _, ok := os.LookupEnv("CI"); ok {
		return true
	}
	return false
}

// IsStress retruns true if we are in stress mode/
func IsStress() bool {
	if _, ok := os.LookupEnv("stress"); ok {
		return true
	}
	return false
}

// CloneDecimal returns a pointer of copy of source.
func CloneDecimal(source decimal.Decimal) *decimal.Decimal {
	return &source
}

// CloneTime returns a pointer of copy of source.
func CloneTime(source time.Time) *time.Time {
	return &source
}

// CloneUUID returns a pointer of copy of source.
func CloneUUID(source uuid.UUID) *uuid.UUID {
	return &source
}

// CloneString returns a pointer of copy of source.
func CloneString(source string) *string {
	return &source
}

// DecimalEqual return true if d0 == d1.
func DecimalEqual(d0 *decimal.Decimal, d1 *decimal.Decimal) bool {
	if d0 == nil {
		if d1 != nil {
			return false
		}
		return true
	}
	if d1 == nil {
		return false
	}
	if (*d0).Equal(*d1) {
		return true
	}
	return false
}

// GetMaskedEmail masks the given email for privacy.
func GetMaskedEmail(email string, count ...int) string {
	template := "*******@*****.***"
	masked := []byte(template)
	name := []byte(strings.Split(email, "@")[0])
	nameLen := len(name)
	masked[0] = name[0]
	if nameLen >= 3 {
		masked[6] = name[nameLen-1]
	}
	if nameLen > 3 {
		masked[5] = name[nameLen-2]
	}
	return string(masked)
}

// JoinNonEmptyString picks up non-empty string and combines them.
func JoinNonEmptyString(strList []string, sep string) string {
	nonEmptyStrList := make([]string, 0, len(strList))
	for _, str := range strList {
		if len(str) == 0 {
			continue
		}
		nonEmptyStrList = append(nonEmptyStrList, str)
	}
	return strings.Join(nonEmptyStrList, sep)
}
