// Code generated by cobctl. This go file is generated automatically, DO NOT EDIT.

package thirdparty

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cobinhood/cobinhood-backend/cache/cacher"
)

// --------------------------------------
// NECAPTCHA_ID
// --------------------------------------
var (
	necaptchaIdStr = os.Getenv("NECAPTCHA_ID")

	necaptchaIdIntCacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseInt(necaptchaIdStr, 0, 0)
		if err != nil {
			panic(fmt.Errorf("failed to parse int for env %s", necaptchaIdStr))
		}
		val := int(v)
		return &val
	})
	necaptchaIdInt64Cacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseInt(necaptchaIdStr, 0, 64)
		if err != nil {
			panic(fmt.Errorf("failed to parse int64 for env %s", necaptchaIdStr))
		}
		val := int64(v)
		return &val
	})
	necaptchaIdUintCacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseUint(necaptchaIdStr, 0, 32)
		if err != nil {
			panic(fmt.Errorf("failed to parse int for env %s", necaptchaIdStr))
		}
		val := uint(v)
		return &val
	})
	necaptchaIdBoolCacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseBool(necaptchaIdStr)
		if err != nil {
			panic(fmt.Errorf("failed to parse bool for env %s", necaptchaIdStr))
		}
		return &v
	})
	necaptchaIdMsCacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseInt(necaptchaIdStr, 0, 64)
		if err != nil {
			panic(fmt.Errorf("failed to parse time for env %s", necaptchaIdStr))
		}

		val := time.Millisecond * time.Duration(v)
		return &val
	})
)

// NecaptchaId returns the cached NECAPTCHA_ID variable.
func NecaptchaId() string {
	return necaptchaIdStr
}

// NecaptchaIdInt returns the cached int of NECAPTCHA_ID variable.
func NecaptchaIdInt() int {
	return *((necaptchaIdIntCacher.Get()).(*int))
}

// NecaptchaIdInt64 returns the cached int64 of NECAPTCHA_ID variable.
func NecaptchaIdInt64() int64 {
	return *((necaptchaIdInt64Cacher.Get()).(*int64))
}

// NecaptchaIdUint returns the cached uint of NECAPTCHA_ID variable.
func NecaptchaIdUint() uint {
	return *((necaptchaIdUintCacher.Get()).(*uint))
}

// NecaptchaIdBool returns the cached bool of NECAPTCHA_ID variable.
func NecaptchaIdBool() bool {
	return *((necaptchaIdBoolCacher.Get()).(*bool))
}

// NecaptchaIdMs returns the cached millisecond of NECAPTCHA_ID variable.
func NecaptchaIdMs() time.Duration {
	return *((necaptchaIdMsCacher.Get()).(*time.Duration))
}

// SetNecaptchaId sets the cached value.
func SetNecaptchaId(v string) {
	necaptchaIdStr = v
	necaptchaIdIntCacher.Clear()
	necaptchaIdInt64Cacher.Clear()
	necaptchaIdUintCacher.Clear()
	necaptchaIdBoolCacher.Clear()
	necaptchaIdMsCacher.Clear()
}

// --------------------------------------
// EPAY_ACCOUNT
// --------------------------------------
var (
	epayAccountStr = os.Getenv("EPAY_ACCOUNT")

	epayAccountIntCacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseInt(epayAccountStr, 0, 0)
		if err != nil {
			panic(fmt.Errorf("failed to parse int for env %s", epayAccountStr))
		}
		val := int(v)
		return &val
	})
	epayAccountInt64Cacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseInt(epayAccountStr, 0, 64)
		if err != nil {
			panic(fmt.Errorf("failed to parse int64 for env %s", epayAccountStr))
		}
		val := int64(v)
		return &val
	})
	epayAccountUintCacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseUint(epayAccountStr, 0, 32)
		if err != nil {
			panic(fmt.Errorf("failed to parse int for env %s", epayAccountStr))
		}
		val := uint(v)
		return &val
	})
	epayAccountBoolCacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseBool(epayAccountStr)
		if err != nil {
			panic(fmt.Errorf("failed to parse bool for env %s", epayAccountStr))
		}
		return &v
	})
	epayAccountMsCacher = cacher.NewConst(func() interface{} {
		v, err := strconv.ParseInt(epayAccountStr, 0, 64)
		if err != nil {
			panic(fmt.Errorf("failed to parse time for env %s", epayAccountStr))
		}

		val := time.Millisecond * time.Duration(v)
		return &val
	})
)

// EpayAccount returns the cached EPAY_ACCOUNT variable.
func EpayAccount() string {
	return epayAccountStr
}

// EpayAccountInt returns the cached int of EPAY_ACCOUNT variable.
func EpayAccountInt() int {
	return *((epayAccountIntCacher.Get()).(*int))
}

// EpayAccountInt64 returns the cached int64 of EPAY_ACCOUNT variable.
func EpayAccountInt64() int64 {
	return *((epayAccountInt64Cacher.Get()).(*int64))
}

// EpayAccountUint returns the cached uint of EPAY_ACCOUNT variable.
func EpayAccountUint() uint {
	return *((epayAccountUintCacher.Get()).(*uint))
}

// EpayAccountBool returns the cached bool of EPAY_ACCOUNT variable.
func EpayAccountBool() bool {
	return *((epayAccountBoolCacher.Get()).(*bool))
}

// EpayAccountMs returns the cached millisecond of EPAY_ACCOUNT variable.
func EpayAccountMs() time.Duration {
	return *((epayAccountMsCacher.Get()).(*time.Duration))
}

// SetEpayAccount sets the cached value.
func SetEpayAccount(v string) {
	epayAccountStr = v
	epayAccountIntCacher.Clear()
	epayAccountInt64Cacher.Clear()
	epayAccountUintCacher.Clear()
	epayAccountBoolCacher.Clear()
	epayAccountMsCacher.Clear()
}
