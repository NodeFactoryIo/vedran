package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/magiconair/properties/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterHandler(t *testing.T) {
	tests := []struct{
		name string
		registerRequest RegisterRequest
		httpStatus int
	}{
		{
			name:            "Valid registration test",
			registerRequest: RegisterRequest{
				Id:            "1",
				ConfigHash:    "dadf2e32dwq12",
				NodeUrl:       "node.test.url",
				PayoutAddress: "0xdafe2cdscdsa",
			},
			httpStatus: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create test request
			rb, _ := json.Marshal(test.registerRequest)
			req, err := http.NewRequest("POST", "/api/v1/node", bytes.NewReader(rb))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(RegisterHandler)
			// invoke test request
			handler.ServeHTTP(rr, req)
			// asserts
			assert.Equal(t, rr.Code, test.httpStatus, "")
		})
	}
}