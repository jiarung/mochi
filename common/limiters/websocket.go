package limiters

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jiarung/mochi/cache"
	"github.com/jiarung/mochi/common/logging"
)

const (
	websocketConnectionKeyBase     = "ws-connection:"
	websocketConnectionMapTimeout  = 300
	websocketConnectionLockTimeout = 1

	// WSConnLimitKey for app client update connection timeout.
	WSConnLimitKey = "ws-conn-limit-key"
)

// ReachWebsocketAPIIP10RPS returns boolean indicates particular IP
// reaches 10 requests/second on API entries or not.
func ReachWebsocketAPIIP10RPS(ip string) bool {
	key := fmt.Sprintf("websocket-ip-rate-limit:%v", ip)
	ret := ReachLimitation(NewLimiter(10, 1), key)
	return ret.Reached
}

// ClearWebsocketAPIIP10RPS by ip if neccessary.
func ClearWebsocketAPIIP10RPS(ip string) error {
	key := fmt.Sprintf("websocket-ip-rate-limit:%v", ip)
	err := cache.GetRedis().Delete(key)
	if err != nil {
		logger.Error("Fail to clear api limit by ip: %v. Err: %v\n", ip, err)
	}
	return err
}

// ReachWesocketIPBlackList return boolean indicates ip is in jail
func ReachWesocketIPBlackList(ip string) bool {
	data, err := cache.GetRedis().Get("ws-ip-black-list:" + ip)
	cacheErrorCode := cache.ParseCacheErrorCode(err)
	if err != nil && cacheErrorCode != cache.ErrNilKey &&
		cacheErrorCode != cache.ErrNoHost {
		logger.Info("get ws-ip-black-list err: %v", err)
		return true
	}

	switch val := data.(type) {
	case string:
		if limitTime, err := strconv.ParseInt(val, 10, 64); err == nil {
			if limitTime > time.Now().Unix() {
				return true
			}
		}
	}
	return false
}

// PutWebsocketIPBlackList return boolean indicates put ip into jail success
func PutWebsocketIPBlackList(ip string) bool {
	now := fmt.Sprintf("%v", time.Now().Add(15*time.Minute).Unix())
	err := cache.GetRedis().Set("ws-ip-black-list:"+ip, now, 15*60)
	if err != nil {
		logger.Info("set ws-ip-black-list err: %v", err)
		return false
	}
	return true
}

// ClearWebsocketIPBlackList by ip if in blacklist.
func ClearWebsocketIPBlackList(ip string) error {
	err := cache.GetRedis().Delete("ws-ip-black-list:" + ip)
	if err != nil {
		logger.Error("Fail to clear api limit by ip: %v. Err: %v\n", ip, err)
	}
	return err
}

// ReachWebsocketConnectionlimit return bool indicates connection amount reach limit
func ReachWebsocketConnectionlimit(key, sessionID string, limit int) bool {
	logger := logging.NewLoggerTag(sessionID)

	redis := cache.GetRedis()
	redis.Lock(key, sessionID, websocketConnectionLockTimeout)
	defer redis.UnLock(key, sessionID)

	// Get connection count
	data, err := redis.GetMap(websocketConnectionKeyBase + key)
	cacheErrorCode := cache.ParseCacheErrorCode(err)
	if err != nil && cacheErrorCode != cache.ErrNilKey &&
		cacheErrorCode != cache.ErrNoHost {
		logger.Error("get websocket-connection err: %v", err)
		return true
	}

	if data != nil {
		logger.Debug("connections count: %v", len(data))
		if len(data) >= limit {
			logger.Error("Websocket connection exceeded %d", limit)
			return true
		}
	}
	// set connection into list
	redis.SetFieldOfMap(
		websocketConnectionKeyBase+key,
		sessionID,
		strconv.FormatInt(time.Now().Unix(), 10),
		websocketConnectionMapTimeout,
	)

	return false
}

// UpdateWebsocketConnectionExpireTime update expire time for connection
func UpdateWebsocketConnectionExpireTime(key, sessionID string) {
	redis := cache.GetRedis()
	redis.Lock(key, sessionID, websocketConnectionLockTimeout)
	defer redis.UnLock(key, sessionID)

	redis.SetFieldOfMap(
		websocketConnectionKeyBase+key,
		sessionID,
		strconv.FormatInt(time.Now().Unix(), 10),
		websocketConnectionMapTimeout,
	)
}

// RemoveWebsocketConnection remove connection record from map.
func RemoveWebsocketConnection(key, sessionID string) error {
	redis := cache.GetRedis()
	redis.Lock(key, sessionID, websocketConnectionLockTimeout)
	defer redis.UnLock(key, sessionID)

	err := redis.RemoveFieldFromMap(websocketConnectionKeyBase+key, sessionID)
	if err != nil {
		logger.Error("Fail to remove api limit by connection: %v:%v. Err: %v\n",
			key, sessionID, err)
	}
	return err
}

// ClearWebsocketConnectionLimit by key if in blacklist.
func ClearWebsocketConnectionLimit(key string) error {
	err := cache.GetRedis().Delete(websocketConnectionKeyBase + key)
	if err != nil {
		logger.Error("Fail to clear api limit by connection: %v. Err: %v\n",
			key, err)
	}
	return err
}
