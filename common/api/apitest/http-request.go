package apitest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// PerformRequest is stolen from
// https://github.com/gin-gonic/gin/blob/4b5ec517daa247f8f3bd0448f82ca55c0d14fa0a/routes_test.go#L19
// for testing middleware or request
func PerformRequest(r http.Handler,
	method,
	path string,
	req *http.Request) *httptest.ResponseRecorder {
	if req == nil {
		req, _ = http.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// HTTPParameter take the needing parameter of SendRequest
type HTTPParameter struct {
	Method          string
	URL             string
	Body            io.Reader
	StatusCode      int
	Headers         map[string]string
	HeaderOperators []func(*http.Request)
}

// SendRequest takes the template of gin.Engine to handle general http routing
func SendRequest(router *gin.Engine, param HTTPParameter) ([]byte, error) {
	request := httptest.NewRequest(param.Method, param.URL, param.Body)
	for k, v := range param.Headers {
		request.Header.Set(k, v)
	}
	for _, operator := range param.HeaderOperators {
		operator(request)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != param.StatusCode {
		errMsg := fmt.Sprintf("expected status %d but got %d",
			param.StatusCode,
			recorder.Code)
		return nil, errors.New(errMsg)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(recorder.Result().Body)
	return buf.Bytes(), nil
}
