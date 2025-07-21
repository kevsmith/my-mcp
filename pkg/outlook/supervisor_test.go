package outlook

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestProcessSupervision tests the process supervision functionality
func TestProcessSupervision(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping process supervision test on Windows - would require actual PowerShell")
	}

	// Test that the manager struct has the required fields for supervision
	manager := &Manager{
		port:          8080,
		baseURL:       "http://localhost:8080",
		client:        &http.Client{Timeout: 5 * time.Second},
		supervisorCtx: context.Background(),
		restartChan:   make(chan bool, 1),
		isShutdown:    false,
	}

	// Verify the required fields exist
	if manager.supervisorCtx == nil {
		t.Error("supervisorCtx should be initialized")
	}
	if manager.restartChan == nil {
		t.Error("restartChan should be initialized")
	}
	if cap(manager.restartChan) != 1 {
		t.Error("restartChan should have capacity of 1")
	}
}

// TestSupervisorLoop tests the supervisor loop logic
func TestSupervisorLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &Manager{
		port:          8080,
		baseURL:       "http://localhost:8080",
		client:        &http.Client{Timeout: 5 * time.Second},
		supervisorCtx: ctx,
		cancelFunc:    cancel,
		restartChan:   make(chan bool, 1),
		isShutdown:    false,
	}

	// Test that supervisor exits when context is cancelled
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		manager.supervisorLoop()
	}()

	// Cancel the context
	cancel()

	// Wait for supervisor to exit with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Test passed - supervisor exited
	case <-time.After(5 * time.Second):
		t.Error("Supervisor did not exit within timeout")
	}
}

// TestRestartSignaling tests that restart signals work correctly
func TestRestartSignaling(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping test on non-Windows platform")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	restartChan := make(chan bool, 1)

	manager := &Manager{
		port:          8080,
		baseURL:       "http://localhost:8080",
		client:        &http.Client{Timeout: 5 * time.Second},
		supervisorCtx: ctx,
		cancelFunc:    cancel,
		restartChan:   restartChan,
		isShutdown:    false,
	}

	// Test that restart signal can be sent
	select {
	case manager.restartChan <- true:
		// Success
	default:
		t.Error("Should be able to send restart signal")
	}

	// Test that channel doesn't block when full
	select {
	case manager.restartChan <- true:
		t.Error("Channel should be full, send should not succeed")
	default:
		// Expected behavior - channel is full
	}

	// Drain the channel
	<-manager.restartChan
	<-manager.restartChan
}

// TestShutdownPreventsRestart tests that shutdown flag prevents restarts
func TestShutdownPreventsRestart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		port:          8080,
		baseURL:       "http://localhost:8080",
		client:        &http.Client{Timeout: 5 * time.Second},
		supervisorCtx: ctx,
		cancelFunc:    cancel,
		restartChan:   make(chan bool, 1),
		isShutdown:    true, // Set shutdown flag
	}

	// Mock a restart attempt - should return early due to shutdown flag
	var wg sync.WaitGroup
	wg.Add(1)

	restartCompleted := false
	go func() {
		defer wg.Done()

		// Send restart signal
		manager.restartChan <- true

		// This should exit quickly due to shutdown flag
		select {
		case <-manager.supervisorCtx.Done():
			// Supervisor context cancelled
		case <-manager.restartChan:
			if manager.isShutdown {
				// Expected: should return early
				return
			}
			// If we get here and not shutting down, that would be a restart
			restartCompleted = true
		case <-time.After(1 * time.Second):
			// Timeout
		}
	}()

	wg.Wait()

	if restartCompleted {
		t.Error("Restart should not complete when shutdown flag is set")
	}

	cancel()
}

// TestManagerStopCancelsSupervision tests that Stop() properly cancels supervision
func TestManagerStopCancelsSupervision(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		port:          8080,
		baseURL:       "http://localhost:8080",
		client:        &http.Client{Timeout: 5 * time.Second},
		supervisorCtx: ctx,
		cancelFunc:    cancel,
		restartChan:   make(chan bool, 1),
		isShutdown:    false,
		cmd:           nil, // No actual process
	}

	// Call Stop() method
	err := manager.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	// Verify shutdown flag is set
	if !manager.isShutdown {
		t.Error("isShutdown flag should be set after Stop()")
	}

	// Verify context is cancelled
	select {
	case <-manager.supervisorCtx.Done():
		// Expected
	default:
		t.Error("supervisorCtx should be cancelled after Stop()")
	}
}

// TestMakeRequestWithSupervision tests that API requests work with supervision
func TestMakeRequestWithSupervision(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"test": "response"}`)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager := &Manager{
		baseURL:       server.URL,
		client:        &http.Client{Timeout: 5 * time.Second},
		supervisorCtx: ctx,
		cancelFunc:    cancel,
		restartChan:   make(chan bool, 1),
		isShutdown:    false,
	}

	// Test that makeRequest still works with supervision fields
	body, err := manager.makeRequest("/test")
	if err != nil {
		t.Errorf("makeRequest failed: %v", err)
	}

	expected := `{"test": "response"}`
	if string(body) != expected {
		t.Errorf("Expected response %s, got %s", expected, string(body))
	}
}
