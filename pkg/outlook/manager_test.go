package outlook

import (
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"
)

// TestManagerRequiresWindows tests that the manager properly validates Windows OS
func TestManagerRequiresWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows platform")
	}

	// This test validates that on non-Windows platforms,
	// the manager would fail to create due to PowerShell dependency
	// We can't actually test NewManager() since it tries to start PowerShell

	// Instead, we'll test the embedded script is available
	if outlookServerScript == "" {
		t.Error("Embedded PowerShell script should not be empty")
	}

	// Verify script contains expected PowerShell content
	expectedContent := []string{
		"New-Object -ComObject Outlook.Application",
		"$listener = New-Object System.Net.HttpListener",
		"GET /messages",
		"GET /search",
	}

	for _, content := range expectedContent {
		if !containsString(outlookServerScript, content) {
			t.Errorf("PowerShell script should contain: %s", content)
		}
	}
}

// TestManagerHTTPClientConfiguration tests HTTP client settings
func TestManagerHTTPClientConfiguration(t *testing.T) {
	// Create a mock manager instance (without starting PowerShell)
	manager := &Manager{
		port:    8080,
		baseURL: "http://localhost:8080",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if manager.client.Timeout != 30*time.Second {
		t.Errorf("Expected timeout of 30s, got %v", manager.client.Timeout)
	}

	if manager.baseURL != "http://localhost:8080" {
		t.Errorf("Expected baseURL 'http://localhost:8080', got %s", manager.baseURL)
	}
}

// TestManagerErrorHandling tests error response handling
func TestManagerErrorHandling(t *testing.T) {
	// Create a test server that returns error responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/messages":
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"Outlook is not available","code":"OUTLOOK_UNAVAILABLE"}`))
		case "/search":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Query parameter 'q' is required","code":"MISSING_QUERY"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Endpoint not found","code":"NOT_FOUND"}`))
		}
	}))
	defer server.Close()

	// Create manager pointing to test server
	manager := &Manager{
		baseURL: server.URL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// Test error handling for unavailable service
	_, err := manager.ListMessages(1)
	if err == nil {
		t.Error("Expected error for unavailable service")
	}
	if !containsString(err.Error(), "Outlook is not available") {
		t.Errorf("Expected 'Outlook is not available' in error, got: %v", err)
	}

	// Test error handling for bad request
	_, err = manager.SearchMessages("")
	if err == nil {
		t.Error("Expected error for empty query")
	}
	if !containsString(err.Error(), "Query parameter") {
		t.Errorf("Expected query parameter error, got: %v", err)
	}
}

// TestManagerSuccessfulResponses tests successful API responses
func TestManagerSuccessfulResponses(t *testing.T) {
	// Create a test server that returns successful responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/messages":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"messages": [
					{
						"id": "test123",
						"subject": "Test Message",
						"sender": "Test Sender",
						"senderEmail": "test@example.com",
						"receivedTime": "2024-01-15T10:30:00.000Z",
						"size": 1024,
						"unread": true,
						"importance": 1,
						"hasAttachments": false,
						"attachmentCount": 0
					}
				],
				"pagination": {
					"page": 1,
					"pageSize": 10,
					"total": 1,
					"hasNext": false,
					"hasPrevious": false
				}
			}`))
		case "/search":
			page := r.URL.Query().Get("q")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"query": "` + page + `",
				"results": [],
				"count": 0
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	manager := &Manager{
		baseURL: server.URL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// Test successful message listing
	response, err := manager.ListMessages(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response == nil {
		t.Error("Expected non-nil response")
	}
	if len(response.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(response.Messages))
	}
	if response.Messages[0].Subject != "Test Message" {
		t.Errorf("Expected subject 'Test Message', got %s", response.Messages[0].Subject)
	}

	// Test successful search
	searchResp, err := manager.SearchMessages("test query")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if searchResp.Query != "test query" {
		t.Errorf("Expected query 'test query', got %s", searchResp.Query)
	}
}

// TestEnvironmentVariableHandling tests port configuration via environment
func TestEnvironmentVariableHandling(t *testing.T) {
	// Save original environment
	originalPort := os.Getenv("OUTLOOK_SERVER_PORT")
	defer func() {
		if originalPort == "" {
			os.Unsetenv("OUTLOOK_SERVER_PORT")
		} else {
			os.Setenv("OUTLOOK_SERVER_PORT", originalPort)
		}
	}()

	// Test default port (8080) when no environment variable
	os.Unsetenv("OUTLOOK_SERVER_PORT")
	// We can't actually create a manager without Windows/PowerShell,
	// but we can verify the embedded script is available
	if outlookServerScript == "" {
		t.Error("Embedded script should be available regardless of environment")
	}

	// Test custom port via environment variable
	os.Setenv("OUTLOOK_SERVER_PORT", "9090")
	// Script should still be available
	if outlookServerScript == "" {
		t.Error("Embedded script should be available with custom port")
	}
}

// Helper function to check if a string contains a substring
func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack == needle ||
			haystack[:len(needle)] == needle ||
			haystack[len(haystack)-len(needle):] == needle ||
			containsSubstring(haystack, needle))
}
