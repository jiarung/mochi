package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// PostSlackMsg send msg to given slack hook url.
func PostSlackMsg(url, msg string) error {
	payload := map[string]string{}
	payload["text"] = msg
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	rsp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		return errors.New(rsp.Status)
	}
	return nil
}
