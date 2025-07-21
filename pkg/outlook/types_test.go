package outlook

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageSerialization(t *testing.T) {
	// Create a test message
	receivedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	sentTime := time.Date(2024, 1, 15, 10, 25, 0, 0, time.UTC)

	msg := Message{
		ID:              "test-id-123",
		Subject:         "Test Message",
		Sender:          "John Doe",
		SenderEmail:     "john@example.com",
		ReceivedTime:    receivedTime,
		SentOn:          &sentTime,
		Size:            1024,
		Unread:          true,
		Importance:      1,
		HasAttachments:  true,
		AttachmentCount: 2,
		BodyPreview:     "This is a test message preview...",
	}

	// Serialize to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	// Deserialize from JSON
	var deserializedMsg Message
	err = json.Unmarshal(data, &deserializedMsg)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	// Verify fields
	if deserializedMsg.ID != msg.ID {
		t.Errorf("ID mismatch: expected %s, got %s", msg.ID, deserializedMsg.ID)
	}
	if deserializedMsg.Subject != msg.Subject {
		t.Errorf("Subject mismatch: expected %s, got %s", msg.Subject, deserializedMsg.Subject)
	}
	if !deserializedMsg.ReceivedTime.Equal(msg.ReceivedTime) {
		t.Errorf("ReceivedTime mismatch: expected %v, got %v", msg.ReceivedTime, deserializedMsg.ReceivedTime)
	}
	if deserializedMsg.SentOn == nil || !deserializedMsg.SentOn.Equal(*msg.SentOn) {
		t.Errorf("SentOn mismatch: expected %v, got %v", msg.SentOn, deserializedMsg.SentOn)
	}
}

func TestMessageListResponse(t *testing.T) {
	messages := []Message{
		{
			ID:           "msg1",
			Subject:      "First Message",
			Sender:       "Alice",
			SenderEmail:  "alice@example.com",
			ReceivedTime: time.Now(),
			Unread:       false,
		},
		{
			ID:           "msg2",
			Subject:      "Second Message",
			Sender:       "Bob",
			SenderEmail:  "bob@example.com",
			ReceivedTime: time.Now().Add(-1 * time.Hour),
			Unread:       true,
		},
	}

	pagination := Pagination{
		Page:        1,
		PageSize:    10,
		Total:       25,
		HasNext:     true,
		HasPrevious: false,
	}

	response := MessageListResponse{
		Messages:   messages,
		Pagination: pagination,
	}

	// Test serialization
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal MessageListResponse: %v", err)
	}

	var deserializedResponse MessageListResponse
	err = json.Unmarshal(data, &deserializedResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal MessageListResponse: %v", err)
	}

	if len(deserializedResponse.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(deserializedResponse.Messages))
	}
	if deserializedResponse.Pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", deserializedResponse.Pagination.Page)
	}
	if !deserializedResponse.Pagination.HasNext {
		t.Error("Expected HasNext to be true")
	}
}

func TestErrorResponseHandling(t *testing.T) {
	errorResp := ErrorResponse{
		Error: "Outlook is not available",
		Code:  "OUTLOOK_UNAVAILABLE",
	}

	data, err := json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse: %v", err)
	}

	var deserializedError ErrorResponse
	err = json.Unmarshal(data, &deserializedError)
	if err != nil {
		t.Fatalf("Failed to unmarshal ErrorResponse: %v", err)
	}

	if deserializedError.Error != errorResp.Error {
		t.Errorf("Error message mismatch: expected %s, got %s", errorResp.Error, deserializedError.Error)
	}
	if deserializedError.Code != errorResp.Code {
		t.Errorf("Error code mismatch: expected %s, got %s", errorResp.Code, deserializedError.Code)
	}
}

func TestSearchResponse(t *testing.T) {
	results := []Message{
		{
			ID:           "search1",
			Subject:      "Meeting reminder",
			Sender:       "Calendar",
			SenderEmail:  "calendar@company.com",
			ReceivedTime: time.Now(),
		},
	}

	searchResp := SearchResponse{
		Query:   "meeting",
		Results: results,
		Count:   1,
	}

	data, err := json.Marshal(searchResp)
	if err != nil {
		t.Fatalf("Failed to marshal SearchResponse: %v", err)
	}

	var deserializedSearch SearchResponse
	err = json.Unmarshal(data, &deserializedSearch)
	if err != nil {
		t.Fatalf("Failed to unmarshal SearchResponse: %v", err)
	}

	if deserializedSearch.Query != "meeting" {
		t.Errorf("Query mismatch: expected 'meeting', got %s", deserializedSearch.Query)
	}
	if deserializedSearch.Count != 1 {
		t.Errorf("Count mismatch: expected 1, got %d", deserializedSearch.Count)
	}
	if len(deserializedSearch.Results) != 1 {
		t.Errorf("Results length mismatch: expected 1, got %d", len(deserializedSearch.Results))
	}
}
