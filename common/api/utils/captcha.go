package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jiarung/mochi/common/config/secret"
	"github.com/jiarung/mochi/common/config/thirdparty"
)

var (
	recaptchaURL = "https://www.google.com/recaptcha/api/siteverify"
	necaptchaID  = thirdparty.NecaptchaId()
	necaptchaURL = "http://c.dun.163yun.com/api/v2/verify"
)

// Defines captcah type
var (
	RECAPTCHA = "recaptcha"
	NECAPTCHA = "necaptcha"
)

// VerifyRecaptcha by google
func VerifyRecaptcha(platform string, response string) error {
	req, err := http.NewRequest(http.MethodPost, recaptchaURL, nil)
	if err != nil {
		return err
	}

	var secretKey string

	// FIXME(xnum): don't directly read from secret.
	if platform == "iOS" || platform == "Web" {
		secretKey = secret.Get("RECAPTCHA_IOS_OR_WEB_SECRET")
	} else if platform == "Android" {
		secretKey = secret.Get("RECAPTCHA_ANDROID_SECRET")
	}

	q := req.URL.Query()
	q.Add("secret", secretKey)
	q.Add("response", response)

	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data map[string]interface{}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	success, ok := data["success"].(bool)

	if !ok {
		return errors.New("Type assertion fail ")
	} else if !success {
		return errors.New("Verify fail ")
	}

	return nil
}

// VerifyNECaptcha by 168yun
func VerifyNECaptcha(platform, neToken string) error {
	// FIXME(xnum): don't directly read from secret.
	necaptchaSecretID := secret.Get("NECAPTCHA_SECRET_ID")
	necaptchaSecretKey := secret.Get("NECAPTCHA_SECRET_KEY")
	timestamp := fmt.Sprintf("%v", time.Now().Unix()*1000)
	params := map[string]string{
		"captchaId": necaptchaID,
		"validate":  neToken,
		"user":      platform,
		"secretId":  necaptchaSecretID,
		"version":   "v2",
		"timestamp": timestamp,
		"nonce":     timestamp,
	}

	// generate ne captcha required signature
	sig := genSignature(necaptchaSecretKey, params)

	form := url.Values{}
	form.Add("captchaId", necaptchaID)
	form.Add("validate", neToken)
	form.Add("user", platform)
	form.Add("secretId", necaptchaSecretID)
	form.Add("version", "v2")
	form.Add("timestamp", timestamp)
	form.Add("nonce", timestamp)
	form.Add("signature", sig)

	request, _ := http.NewRequest(
		http.MethodPost,
		necaptchaURL,
		strings.NewReader(form.Encode()),
	)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Failed to verify ne captcha (%v)", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to verify ne captcha (%v)", err)
	}

	response := struct {
		Msg    string
		Result bool
		Error  int
	}{}

	json.Unmarshal(body, &response)
	if response.Result == true {
		return nil
	}
	return fmt.Errorf("Failed to verify ne captcha (%+v)", response)
}

func genSignature(secretKey string, params map[string]string) string {
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	buf := bytes.NewBufferString("")
	for _, key := range keys {
		buf.WriteString(key + params[key])
	}
	buf.WriteString(secretKey)
	has := md5.Sum(buf.Bytes())
	return fmt.Sprintf("%x", has)
}

// SelectCaptcha set the captcha type
func SelectCaptcha(ctx *gin.Context) string {
	if ctx.Request.Header.Get("g-recaptcha-token") != "" {
		return RECAPTCHA
	}
	return NECAPTCHA
}
