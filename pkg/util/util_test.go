package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testRequest struct {
	Id string `json:"id"`
}

func TestDecodeJSONBody(t *testing.T) {

	validTestRequest := testRequest{Id: "1"}
	validTestRequestBody, _ := json.Marshal(validTestRequest)

	tests := []struct {
		name        string
		contentType string
		request     testRequest
		err         MalformedRequest
		bodyContent string
		success     bool
	}{
		{
			name:        "body successfully decoded",
			contentType: "application/json",
			request:     validTestRequest,
			err:         MalformedRequest{},
			bodyContent: string(validTestRequestBody),
			success:     true,
		},
		{
			name:        "wrong Content-Type header",
			contentType: "txt",
			request:     validTestRequest,
			err: MalformedRequest{
				Status: http.StatusUnsupportedMediaType,
				Msg:    "Content-Type header is not application/json",
			},
			bodyContent: string(validTestRequestBody),
			success:     false,
		},
		{
			name:        "unknown field in json body content",
			contentType: "application/json",
			request:     validTestRequest,
			err: MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Request body contains unknown field \"prop\"",
			},
			bodyContent: "{\"prop\":\"value\"}",
			success:     false,
		},
		{
			name:        "baldy-formed JSON in json body content",
			contentType: "application/json",
			request:     validTestRequest,
			err: MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Request body contains badly-formed JSON (at position 2)",
			},
			bodyContent: "{prop:\"value\"}",
			success:     false,
		},
		{
			name:        "empty json body content",
			contentType: "application/json",
			request:     validTestRequest,
			err: MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Request body must not be empty",
			},
			bodyContent: "",
			success:     false,
		},
		{
			name:        "invalid value for field in json body content",
			contentType: "application/json",
			request:     validTestRequest,
			err: MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Request body contains an invalid value for the \"id\" field (at position 8)",
			},
			bodyContent: "{\"id\":10}",
			success:     false,
		},
	}
	for _, test := range tests {
		var testPrefix string
		if test.success {
			testPrefix = "Valid request"
		} else {
			testPrefix = "Invalid request"
		}
		t.Run(fmt.Sprintf("%s %s", testPrefix, test.name), func(t *testing.T) {

			req, _ := http.NewRequest("POST", "/api/v1/", bytes.NewReader([]byte(test.bodyContent)))
			req.Header.Set("Content-Type", test.contentType)
			rr := httptest.NewRecorder()
			var parsedRequest testRequest

			err := DecodeJSONBody(rr, req, &parsedRequest)

			if test.success {
				assert.Equal(t, err, nil)
				assert.Equal(t, test.request, parsedRequest)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, test.err.Msg)
				assert.Equal(t, testRequest{}, parsedRequest)
			}
		})
	}
}
