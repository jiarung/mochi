package notification

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Slack message buffer should contain valid JSON object.
// ref: https://api.slack.com/incoming-webhooks

// SlackWebHook defines config to webhook.
type SlackWebHook struct {
	URL string
}

func slackHookPublish(url string, msg *bytes.Buffer) (err error) {
	client := &http.Client{}
	resp, err := client.Post(url, "application/json", msg)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Sucess
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = fmt.Errorf("Slack error: %s", body)
	return
}

// PublishToSlackOpsNotification to Slack channel #ops-notification on production
//  and #operation-test on dev and staging.
func (s *SlackWebHook) PublishToSlackOpsNotification(message *bytes.Buffer) error {
	return slackHookPublish(s.URL, message)
}
