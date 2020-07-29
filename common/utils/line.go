// Copyright 2018-2019 Cobinhood Inc. All rights reserved.

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client
var err error
var mux sync.Mutex

// GetLineBotConfig returns line bot config.
func GetLineBotConfig() (lineBotConfig map[string]interface{}) {
	configName := "LINEBOT_CONFIG"
	confStr := os.Getenv(configName)
	err = json.Unmarshal([]byte(confStr), &lineBotConfig)
	if err != nil {
		e := fmt.Sprintf("failed to get %s err=%s", configName, err)
		panic(errors.New(e))
	}
	return
}

func getChannelConfig() (channelSecret, channelToken string) {
	lineBotConfig := GetLineBotConfig()

	var v string
	var ok bool
	if v, ok = lineBotConfig["channel_secret"].(string); !ok {
		err = errors.New("failed to get channel_secret")
		panic(err)
	}
	channelSecret = v
	if v, ok = lineBotConfig["channel_token"].(string); !ok {
		err = errors.New("failed to get channel_token")
		panic(err)
	}
	channelToken = v
	return
}

func newLinebot() *linebot.Client {
	s, t := getChannelConfig()

	bot, err = linebot.New(s, t)
	if err != nil {
		e := fmt.Sprintf("failed to new linebot err=%s", err)
		panic(e)
	}
	return bot
}

// PostLineMsg sends msg to given ID.
func PostLineMsg(id, msg string) error {
	bot = GetLinebot()
	_, err = bot.PushMessage(id, linebot.NewTextMessage(msg)).Do()
	return err
}

// MulticastLineMsg sends msg to given IDs.
func MulticastLineMsg(ids []string, msg string) error {
	bot = GetLinebot()
	_, err = bot.Multicast(ids, linebot.NewTextMessage(msg)).Do()
	return err
}

// GetLinebot returns line bot client.
func GetLinebot() *linebot.Client {
	if bot == nil {
		mux.Lock()
		if bot == nil {
			bot = newLinebot()
		}
		mux.Unlock()
	}
	return bot
}

// ParseLinebot parse body to linebot's events.
func ParseLinebot(body []byte) ([]*linebot.Event, error) {
	request := &struct {
		Events []*linebot.Event `json:"events"`
	}{}
	if err = json.Unmarshal(body, request); err != nil {
		return nil, err
	}
	return request.Events, nil
}
