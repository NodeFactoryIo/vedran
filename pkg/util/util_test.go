package util

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testRequest struct {
	Id string `json:"id"`
}

func TestDecodeJSONBody(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		request     testRequest
		err         MalformedRequest
		success     bool
	}{
		{name: "Valid body", contentType: "application/json", request: testRequest{Id: "1"}, err: MalformedRequest{
				// empty error
		}, success: true},
		{name: "Valid body", contentType: "txt", request: testRequest{Id: "1"}, err: MalformedRequest{
			Status: http.StatusUnsupportedMediaType,
			Msg:    "Content-Type header is not application/json",
		}, success: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rb, _ := json.Marshal(test.request)
			req, _ := http.NewRequest("POST", "/api/v1/", bytes.NewReader(rb))
			req.Header.Set("Content-Type", test.contentType)
			rr := httptest.NewRecorder()
			var parsedRequest testRequest

			err := DecodeJSONBody(rr, req, &parsedRequest)

			if test.success {
				assert.Equal(t, err, nil)
				assert.Equal(t, parsedRequest, test.request)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, test.err.Msg)
				assert.Equal(t, parsedRequest, testRequest{})
			}

		})
	}
}
