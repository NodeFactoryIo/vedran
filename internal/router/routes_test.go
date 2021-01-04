package router

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"testing"

	"github.com/NodeFactoryIo/vedran/internal/controllers"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestApiRouter(t *testing.T) {
	nodeRepoMock := mocks.NodeRepository{}
	pingRepoMock := mocks.PingRepository{}
	metricsRepoMock := mocks.MetricsRepository{}
	recordRepoMock := mocks.RecordRepository{}
	apiController := controllers.NewApiController(
		false,
		repositories.Repos{
			NodeRepo:    &nodeRepoMock,
			PingRepo:    &pingRepoMock,
			MetricsRepo: &metricsRepoMock,
			RecordRepo:  &recordRepoMock,
		},
		nil,
	)

	tests := []struct {
		name    string
		url     string
		methods []string
	}{
		{name: "Test register route", url: "/api/v1/nodes", methods: []string{"POST"}},
		{name: "Test ping route", url: "/api/v1/nodes/pings", methods: []string{"POST"}},
		{name: "Test metrics route", url: "/api/v1/nodes/metrics", methods: []string{"PUT"}},
	}

	router := mux.NewRouter()
	createRoutes(apiController, router, "")

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
