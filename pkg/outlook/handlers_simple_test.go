package outlook

import (
	"testing"
)

func TestFormatMessageListSimple(t *testing.T) {
	// Test empty message list
	result := formatMessageList([]Message{})
	expected := "No messages found."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGetImportanceStringSimple(t *testing.T) {
	tests := []struct {
		importance int
		expected   string
	}{
		{0, "Low"},
		{1, "Normal"},
		{2, "High"},
		{999, "Unknown"},
	}

	for _, tt := range tests {
		result := getImportanceString(tt.importance)
		if result != tt.expected {
			t.Errorf("getImportanceString(%d) = %s, expected %s", tt.importance, result, tt.expected)
		}
	}
}

func TestOutlookServerScriptEmbedded(t *testing.T) {
	if outlookServerScript == "" {
		t.Error("Embedded PowerShell script should not be empty")
	}

	// Verify script contains expected PowerShell content
	expectedContent := []string{
		"New-Object -ComObject Outlook.Application",
		"$listener = New-Object System.Net.HttpListener",
		"/messages",
		"/search",
	}

	for _, content := range expectedContent {
		if !containsSubstring(outlookServerScript, content) {
			t.Errorf("PowerShell script should contain: %s", content)
		}
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
