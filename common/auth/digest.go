// This package is refer to https://github.com/xinsnake/go-http-digest-auth-client
// and https://github.com/delphinus/go-digest-request

package auth

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jiarung/mochi/common/logging"
)

var (
	logger = logging.NewLoggerTag("digest")
)

// DigestRequest defines request for digest auth.
type DigestRequest struct {
	Body     string
	Method   string
	Password string
	URI      string
	Username string
}

// DigestAuthParams defines parameters for digest auth.
type DigestAuthParams struct {
	Algorithm string
	Domain    string
	Cnonce    string
	Nc        int
	Nonce     string
	Opaque    string
	Qop       string
	Realm     string
	Response  string
	URI       string
	Userhash  bool
	Username  string
	Stale     bool
}

// NewDigestAuthParams creates parameter set from www header.
func NewDigestAuthParams(waHeader string) *DigestAuthParams {
	s := strings.SplitN(waHeader, " ", 2)
	if len(s) != 2 || s[0] != "Digest" {
		return nil
	}

	param := s[1]

	result := &DigestAuthParams{Nc: 1}

	algorithmRegex := regexp.MustCompile(`algorithm=([^ ,]+)`)
	algorithmMatch := algorithmRegex.FindStringSubmatch(param)
	if algorithmMatch != nil {
		result.Algorithm = algorithmMatch[1]
	}

	domainRegex := regexp.MustCompile(`domain="(.+?)"`)
	domainMatch := domainRegex.FindStringSubmatch(param)
	if domainMatch != nil {
		result.Domain = domainMatch[1]
	}

	nonceRegex := regexp.MustCompile(`nonce="(.+?)"`)
	nonceMatch := nonceRegex.FindStringSubmatch(param)
	if nonceMatch != nil {
		result.Nonce = nonceMatch[1]
	}

	opaqueRegex := regexp.MustCompile(`opaque="(.+?)"`)
	opaqueMatch := opaqueRegex.FindStringSubmatch(param)
	if opaqueMatch != nil {
		result.Opaque = opaqueMatch[1]
	}

	qopRegex := regexp.MustCompile(`qop="(.+?)"`)
	qopMatch := qopRegex.FindStringSubmatch(param)
	if qopMatch != nil {
		result.Qop = qopMatch[1]
	}

	realmRegex := regexp.MustCompile(`realm="(.+?)"`)
	realmMatch := realmRegex.FindStringSubmatch(param)
	if realmMatch != nil {
		result.Realm = realmMatch[1]
	}

	staleRegex := regexp.MustCompile(`stale=([^ ,])"`)
	staleMatch := staleRegex.FindStringSubmatch(param)
	if staleMatch != nil {
		result.Stale = (strings.ToLower(staleMatch[1]) == "true")
	}

	userhashRegex := regexp.MustCompile(`userhash=([^ ,])"`)
	userhashMatch := userhashRegex.FindStringSubmatch(param)
	if userhashMatch != nil {
		result.Userhash = (strings.ToLower(userhashMatch[1]) == "true")
	}

	return result
}

func (ah *DigestAuthParams) computeA1(dr *DigestRequest) (ret string, err error) {

	switch ah.Algorithm {
	case "", "MD5", "SHA-256":
		ret = fmt.Sprintf("%s:%s:%s", dr.Username, ah.Realm, dr.Password)
	case "MD5-sess", "SHA-256-sess":
		upHash, err := ah.computeHash(fmt.Sprintf("%s:%s:%s", dr.Username, ah.Realm, dr.Password))
		if err != nil {
			return "", err
		}
		ret = fmt.Sprintf("%s:%s:%s", upHash, ah.Nonce, ah.Cnonce)
	default:
		err = fmt.Errorf("unknown digest algorithm: %v", ah.Algorithm)
	}

	return
}

func (ah *DigestAuthParams) computeA2(dr *DigestRequest) (string, error) {

	if matched, _ := regexp.MatchString("auth-int", ah.Qop); matched {
		ah.Qop = "auth-int"
		hashedBody, err := ah.computeHash(dr.Body)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%s:%s", dr.Method, ah.URI, hashedBody), nil
	}

	if ah.Qop == "auth" || ah.Qop == "" {
		ah.Qop = "auth"
		return fmt.Sprintf("%s:%s", dr.Method, ah.URI), nil
	}

	return "", fmt.Errorf("unknwon qop: %v", ah.Qop)
}

func (ah *DigestAuthParams) computeResponse(dr *DigestRequest) error {
	ah.Username = dr.Username
	cnonce, err := ah.computeHash(fmt.Sprintf("%d:%s:my_value", time.Now().UnixNano(), dr.Username))
	if err != nil {
		return err
	}
	ah.Cnonce = cnonce
	ah.URI = dr.URI

	A1, err := ah.computeA1(dr)
	if err != nil {
		return err
	}

	kdSecret, err := ah.computeHash(A1)
	if err != nil {
		return err
	}

	A2, err := ah.computeA2(dr)
	if err != nil {
		return err
	}

	A2Hash, err := ah.computeHash(A2)
	if err != nil {
		return err
	}

	kdData := fmt.Sprintf("%s:%08x:%s:%s:%s", ah.Nonce, ah.Nc, ah.Cnonce, ah.Qop, A2Hash)

	resp, err := ah.computeHash(fmt.Sprintf("%s:%s", kdSecret, kdData))
	if err != nil {
		return err
	}

	ah.Response = resp

	return nil
}

func (ah *DigestAuthParams) computeHash(a string) (s string, err error) {

	var h hash.Hash

	switch ah.Algorithm {
	case "", "MD5", "MD5-sess":
		h = md5.New()
	case "SHA-256", "SHA-256-sess":
		h = sha256.New()
	default:
		err = fmt.Errorf("unknown digest algorithm: %v", ah.Algorithm)
	}

	io.WriteString(h, a)
	s = hex.EncodeToString(h.Sum(nil))

	return
}

// String converts DigestAuthParams to http header.
func (ah *DigestAuthParams) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("Digest ")

	if ah.Algorithm != "" {
		buffer.WriteString(fmt.Sprintf("algorithm=%s, ", ah.Algorithm))
	}

	if ah.Cnonce != "" {
		buffer.WriteString(fmt.Sprintf("cnonce=\"%s\", ", ah.Cnonce))
	}

	if ah.Nc != 0 {
		buffer.WriteString(fmt.Sprintf("nc=%08x, ", ah.Nc))
	}

	if ah.Opaque != "" {
		buffer.WriteString(fmt.Sprintf("opaque=\"%s\", ", ah.Opaque))
	}

	if ah.Nonce != "" {
		buffer.WriteString(fmt.Sprintf("nonce=\"%s\", ", ah.Nonce))
	}

	if ah.Qop != "" {
		buffer.WriteString(fmt.Sprintf("qop=%s, ", ah.Qop))
	}

	if ah.Realm != "" {
		buffer.WriteString(fmt.Sprintf("realm=\"%s\", ", ah.Realm))
	}

	if ah.Response != "" {
		buffer.WriteString(fmt.Sprintf("response=\"%s\", ", ah.Response))
	}

	if ah.URI != "" {
		buffer.WriteString(fmt.Sprintf("uri=\"%s\", ", ah.URI))
	}

	if ah.Userhash {
		buffer.WriteString("userhash=true, ")
	}

	if ah.Username != "" {
		buffer.WriteString(fmt.Sprintf("username=\"%s\", ", ah.Username))
	}

	s := buffer.String()

	return strings.TrimSuffix(s, ", ")
}

// DoDigestAuth do http request with digest auth.
func DoDigestAuth(ctx context.Context, r *http.Response,
	endpoint, method, URI, body, payload, user, pass string) (*bytes.Buffer, error) {
	// Header WWW-Authenticate will be convert to Www-Authenticate
	// in function CanonicalMIMEHeaderKey().
	// It calls when reading response.
	// https://golang.org/pkg/net/textproto/#CanonicalMIMEHeaderKey
	authParam := NewDigestAuthParams(r.Header.Get("Www-Authenticate"))
	if authParam == nil {
		return nil, errors.New("parse header failed")
	}

	dr := &DigestRequest{
		Body:     body,
		Username: user,
		Password: pass,
		Method:   method,
		URI:      URI,
	}

	if err := authParam.computeResponse(dr); err != nil {
		return nil, err
	}

	var req *http.Request
	req, err := http.NewRequest(method, endpoint,
		strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	req = req.WithContext(ctx)
	req.Header.Add("Authorization", authParam.String())
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(rsp.Body)
	rsp.Body.Close()
	return buf, nil
}
