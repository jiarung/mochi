package limiters

import (
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/common/logging"
)

var logger = logging.NewLoggerTag("api-limiter")

// Result result the reached limiter count and limit.
type Result struct {
	Reached   bool
	Count     int64
	Limit     int64
	ExpiredAt int64
	Seconds   int64
}

// Limiter interface test given key and returns result and error
type Limiter interface {
	Get(string) (Result, error)
}

type concreteLimiter struct {
	RedisCli *cache.Redis
	Limit    int64
	Seconds  int
}

func (c *concreteLimiter) Get(key string) (res Result, err error) {

	rCli, release := c.RedisCli.GetConn()
	defer release()
	if _, err = rCli.Do(
		"SET", key, 0, "EX", c.Seconds, "NX"); err != nil {
		return
	}

	res.Count, err = redis.Int64(rCli.Do("INCR", key))
	if err != nil {
		logger.Error("redis INCR error %v", err)
	} else if res.Count > c.Limit {
		res.Reached = true
		logger.Warn("Reach limit res (%v) > limit (%v)", res.Count, c.Limit)
	}

	res.Limit = c.Limit
	res.Seconds = int64(c.Seconds)
	ttl, tErr := redis.Int(rCli.Do("TTL", key))
	if ttl < 0 {
		// tErr and eErr prevents variable shadowed
		if _, eErr := rCli.Do(
			"EXPIRE", key, c.Seconds); eErr != nil {
			err = eErr
			return
		}
	} else if tErr != nil {
		logger.Error("redis TTL error %v", tErr)
		err = tErr
		return
	}
	res.ExpiredAt = time.Now().Unix() + int64(ttl)
	return
}

// NewLimiter with condition: <limit> requests per <period>
func NewLimiter(limit int64, seconds int) Limiter {
	return &concreteLimiter{
		RedisCli: cache.GetRedis(),
		Limit:    limit,
		Seconds:  seconds,
	}
}

// ReachLimitation checks by given limiter and key
func ReachLimitation(cLimiter Limiter, key string) Result {
	logger.Debug("Checking key: %v", key)
	result, err := cLimiter.Get(key)
	if err != nil {
		logger.Error(
			"Fail to check limitation with limiter(%v) and key(%v). Err: %v\n",
			cLimiter, key, err)
		result.Reached = true
		return result
	}

	if result.Reached {
		logger.Warn("Reach limitation with key: %v. Result: %v\n", key, result)
	}
	logger.Debug("Limitation result: %v, %v\n", key, result)
	return result
}
