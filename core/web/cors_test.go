package web_test

import (
	"net/http"
	"testing"

	"github.com/SeerLink/seerlink/core/services/eth"

	"github.com/SeerLink/seerlink/core/internal/cltest"

	"github.com/stretchr/testify/require"
)

func TestCors_DefaultOrigins(t *testing.T) {
	t.Parallel()

	config, _ := cltest.NewConfig(t)
	config.Set("ALLOW_ORIGINS", "http://localhost:3000,http://localhost:6689")
	rpcClient, gethClient, _, assertMocksCalled := cltest.NewEthMocksWithStartupAssertions(t)
	defer assertMocksCalled()
	app, cleanup := cltest.NewApplicationWithConfigAndKey(t,
		config,
		eth.NewClientWith(rpcClient, gethClient),
	)
	defer cleanup()
	require.NoError(t, app.Start())

	client := app.NewHTTPClient()

	tests := []struct {
		origin     string
		statusCode int
	}{
		{"http://localhost:3000", http.StatusOK},
		{"http://localhost:6689", http.StatusOK},
		{"http://localhost:1234", http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.origin, func(t *testing.T) {
			headers := map[string]string{"Origin": test.origin}
			resp, cleanup := client.Get("/v2/config", headers)
			defer cleanup()
			cltest.AssertServerResponse(t, resp, test.statusCode)
		})
	}
}

func TestCors_OverrideOrigins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		allow      string
		origin     string
		statusCode int
	}{
		{"http://seerlink.com", "http://seerlink.com", http.StatusOK},
		{"http://seerlink.com", "http://localhost:3000", http.StatusForbidden},
		{"*", "http://seerlink.com", http.StatusOK},
		{"*", "http://localhost:3000", http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.origin, func(t *testing.T) {
			config, _ := cltest.NewConfig(t)
			config.Set("ALLOW_ORIGINS", test.allow)

			rpcClient, gethClient, _, assertMocksCalled := cltest.NewEthMocksWithStartupAssertions(t)
			defer assertMocksCalled()
			app, cleanup := cltest.NewApplicationWithConfigAndKey(t,
				config,
				eth.NewClientWith(rpcClient, gethClient),
			)
			defer cleanup()
			require.NoError(t, app.Start())

			client := app.NewHTTPClient()

			headers := map[string]string{"Origin": test.origin}
			resp, cleanup := client.Get("/v2/config", headers)
			defer cleanup()
			cltest.AssertServerResponse(t, resp, test.statusCode)
		})
	}
}
