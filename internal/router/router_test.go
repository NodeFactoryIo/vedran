package router

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApiRouter(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		methods []string
	}{
		{name: "Test register route", url: "/api/v1/nodes", methods: []string{"POST"}},
	}
	// pass nil as db instance as only routes are tested
	router := CreateNewApiRouter(nil)
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s::%s", test.name, test.url), func(t *testing.T) {
			rout := router.GetRoute(test.url)
			assert.NotNil(t, rout, fmt.Sprintf("Assert that API rout %s is defined", test.url))
			if rout != nil {
				methods, _ := rout.GetMethods()
				assert.Equal(
					t, methods, test.methods,
					fmt.Sprintf("Assert that API rout %s handles methods: %v", test.url, test.methods))
			}
		})
	}
}
