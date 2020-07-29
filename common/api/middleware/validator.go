package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"

	cobxtypes "github.com/jiarung/mochi/apps/exchange/cobx-types"
)

type hasT interface {
	T() *testing.T
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

const (
	// ValidateRequestKey defines validate request control key.
	ValidateRequestKey = "Do_not_scan_req"
	// ValidateResponseKey defines validate response control key.
	ValidateResponseKey = "Do_not_scan_res"
)

// DoNotScanRequest sets header to bypass request validation.
func DoNotScanRequest(req *http.Request) {
	req.Header.Set(ValidateRequestKey, "1")
}

// DoNotScanResponse sets header to bypass response validation.
func DoNotScanResponse(req *http.Request) {
	req.Header.Set(ValidateResponseKey, "1")
}

func isHeaderNotExists(ctx *gin.Context, key string) bool {
	s := ctx.GetHeader(key)
	return s == ""
}

func validateJSON(schemaPath string,
	body interface{}) error {
	_, err := os.Stat(schemaPath)
	if err != nil {
		// Stat() may returns `Not Exists` error, we have to exclude this case.
		if !os.IsNotExist(err) {
			return err
		}

		// If body has any data and the schema json file is not exists.
		if body != nil && os.IsNotExist(err) {
			return fmt.Errorf("request schema is undefined: %v", err)
		}
	} else {
		requestPath := "file://" + schemaPath
		schemaLoader := gojsonschema.NewReferenceLoader(requestPath)

		validator, err := gojsonschema.Validate(
			schemaLoader,
			gojsonschema.NewRawLoader(body))

		if err != nil {
			return fmt.Errorf("validator creation failed: %v", err)
		}

		if !validator.Valid() {
			var str strings.Builder
			str.WriteString("schema is invalid:\n")
			for _, e := range validator.Errors() {
				str.WriteString("\t\t")
				str.WriteString(e.String())
				str.WriteString("\n")
			}

			return errors.New(str.String())
		}
	}

	return nil
}

// validateFail prints fail reason and related information then failed.
func validateFail(
	t hasT, path, file string, actual interface{}, reason error) {
	buf, _ := json.MarshalIndent(actual, "\t", "\t")

	var msg strings.Builder
	msg.WriteString("Validation failed:\n")
	fmt.Fprintf(&msg, "\tPath: %v\n", path)
	fmt.Fprintf(&msg, "\tValidator File: %v\n", file)
	fmt.Fprintf(&msg, "\tActual: \n\t%s\n", buf)
	fmt.Fprintf(&msg, "\tReason: %v\n", reason.Error())

	assert.Fail(t.T(), msg.String())
}

// APIValidator is a middleware to support query parameter, request,
// response check.
func APIValidator(t hasT, service cobxtypes.ServiceName) func(*gin.Context) {
	return func(ctx *gin.Context) {
		// From Reqeust Header.
		scanRequest := isHeaderNotExists(ctx, ValidateRequestKey)
		scanResponse := isHeaderNotExists(ctx, ValidateResponseKey)

		// Build prefix.
		endpoint := strings.Split(ctx.Request.URL.Path, "/")
		prefix := os.Getenv("GITROOT") + "/tmp/validator/" + string(service)
		for _, comp := range endpoint {
			if comp != "" {
				_, err := os.Stat(prefix + "/" + comp)
				if err == nil {
					prefix += "/" + comp
				} else {
					prefix += "/::cobin::"
				}
			}
		}
		prefix += "/" + ctx.Request.Method

		var request interface{}
		// Prevent panic.
		if ctx.Request.Body != nil {
			buf, _ := ioutil.ReadAll(ctx.Request.Body)

			rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
			decoder := json.NewDecoder(rdr1)

			decoder.Decode(&request)
			ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		}

		// Validate request schema.
		if scanRequest {
			schemaFile := prefix + "/Request.json"
			err := validateJSON(schemaFile, request)
			if err != nil {
				validateFail(t, ctx.Request.URL.String(),
					schemaFile, request, err)
			}
		}

		// Inject our response writer.
		hijackWriter := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: ctx.Writer}
		ctx.Writer = hijackWriter
		ctx.Next()

		// Check response.
		if !scanResponse {
			return
		}

		var response interface{}
		rdr1 := ioutil.NopCloser(
			bytes.NewBuffer(
				hijackWriter.body.Bytes()))

		decoder := json.NewDecoder(rdr1)
		decoder.Decode(&response)

		// Failure response mustn't matches schema, so we ignore this case.
		if responseMap, ok := response.(map[string]interface{}); ok {
			if val, ok := responseMap["success"]; ok {
				if success, ok := val.(bool); ok && !success {
					return
				}
			}
		}

		schemaFile := prefix + "/Response.json"
		err := validateJSON(schemaFile, response)
		if err != nil {
			validateFail(t, ctx.Request.URL.String(),
				schemaFile, response, err)
		}
	}
}
