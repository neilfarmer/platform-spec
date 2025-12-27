package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_HTTPTest(t *testing.T) {
	tests := []struct {
		name         string
		httpTest     core.HTTPTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "successful GET request with 200",
			httpTest: core.HTTPTest{
				Name:       "API health check",
				URL:        "http://localhost:8080/health",
				StatusCode: 200,
				Method:     "GET",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:8080/health'", "{\"status\":\"ok\"}\n200", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "returned status 200",
		},
		{
			name: "POST request with custom status code",
			httpTest: core.HTTPTest{
				Name:       "Webhook endpoint",
				URL:        "https://api.example.com/webhook",
				StatusCode: 202,
				Method:     "POST",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -X POST -w $'\\n%{http_code}' 'https://api.example.com/webhook'", "{\"accepted\":true}\n202", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "returned status 202",
		},
		{
			name: "request with insecure flag",
			httpTest: core.HTTPTest{
				Name:       "Self-signed cert",
				URL:        "https://internal.local/api",
				StatusCode: 200,
				Method:     "GET",
				Insecure:   true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -k -w $'\\n%{http_code}' 'https://internal.local/api'", "{\"data\":\"test\"}\n200", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "returned status 200",
		},
		{
			name: "response with content validation",
			httpTest: core.HTTPTest{
				Name:       "API response validation",
				URL:        "http://localhost:3000/status",
				StatusCode: 200,
				Method:     "GET",
				Contains:   []string{"healthy", "version"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:3000/status'", "{\"status\":\"healthy\",\"version\":\"1.2.3\"}\n200", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "with all expected content (2 strings)",
		},
		{
			name: "wrong status code",
			httpTest: core.HTTPTest{
				Name:       "Expect 200 but got 404",
				URL:        "http://localhost:8080/missing",
				StatusCode: 200,
				Method:     "GET",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:8080/missing'", "Not Found\n404", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Status code is 404, expected 200",
		},
		{
			name: "missing content in response",
			httpTest: core.HTTPTest{
				Name:       "Missing expected string",
				URL:        "http://localhost:8080/api",
				StatusCode: 200,
				Method:     "GET",
				Contains:   []string{"expected_field"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:8080/api'", "{\"other\":\"data\"}\n200", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Response body missing expected strings: 'expected_field'",
		},
		{
			name: "curl command fails",
			httpTest: core.HTTPTest{
				Name:       "Connection refused",
				URL:        "http://localhost:9999",
				StatusCode: 200,
				Method:     "GET",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:9999'", "", "curl: (7) Failed to connect", 7, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "HTTP request failed",
		},
		{
			name: "redirect (302) when expecting 200",
			httpTest: core.HTTPTest{
				Name:       "No redirect expected",
				URL:        "https://example.com/old",
				StatusCode: 200,
				Method:     "GET",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'https://example.com/old'", "<html>Moved</html>\n302", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Status code is 302, expected 200",
		},
		{
			name: "follow redirects and get final 200",
			httpTest: core.HTTPTest{
				Name:            "Follow redirect to final page",
				URL:             "https://example.com/redirect",
				StatusCode:      200,
				Method:          "GET",
				FollowRedirects: true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -L -w $'\\n%{http_code}' 'https://example.com/redirect'", "<html>Final page</html>\n200", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "returned status 200",
		},
		{
			name: "follow redirects with insecure flag",
			httpTest: core.HTTPTest{
				Name:            "Follow redirect with self-signed cert",
				URL:             "https://internal.local/old",
				StatusCode:      200,
				Method:          "GET",
				FollowRedirects: true,
				Insecure:        true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -L -k -w $'\\n%{http_code}' 'https://internal.local/old'", "<html>Redirected page</html>\n200", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "returned status 200",
		},
		{
			name: "follow redirects with POST method",
			httpTest: core.HTTPTest{
				Name:            "POST with redirect",
				URL:             "https://api.example.com/create",
				StatusCode:      201,
				Method:          "POST",
				FollowRedirects: true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("curl -s -X POST -L -w $'\\n%{http_code}' 'https://api.example.com/create'", "{\"id\":123}\n201", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "returned status 201",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeHTTPTest(ctx, mock, tt.httpTest)

			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}
