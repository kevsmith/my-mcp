package outlook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	_ "embed"
)

//go:embed scripts/outlook-server.ps1
var outlookServerScript string

// Manager handles the PowerShell server process and REST API communication
type Manager struct {
	port          int
	cmd           *exec.Cmd
	baseURL       string
	client        *http.Client
	supervisorCtx context.Context
	cancelFunc    context.CancelFunc
	restartChan   chan bool
	isShutdown    bool
}

// NewManager creates a new Outlook manager and starts the PowerShell server
func NewManager() (*Manager, error) {
	port := 8080
	if portEnv := os.Getenv("OUTLOOK_SERVER_PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			port = p
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		port:          port,
		baseURL:       fmt.Sprintf("http://localhost:%d", port),
		client:        &http.Client{Timeout: 30 * time.Second},
		supervisorCtx: ctx,
		cancelFunc:    cancel,
		restartChan:   make(chan bool, 1),
		isShutdown:    false,
	}

	if err := m.startPowerShellServer(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start PowerShell server: %w", err)
	}

	// Wait for server to be ready
	if err := m.waitForServer(); err != nil {
		m.Stop()
		return nil, fmt.Errorf("server failed to start: %w", err)
	}

	// Start process supervisor
	go m.supervisorLoop()

	return m, nil
}

// startPowerShellServer starts the PowerShell server process
func (m *Manager) startPowerShellServer() error {
	// Create temp file for the script
	tmpFile, err := os.CreateTemp("", "outlook-server-*.ps1")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write the embedded script to the temp file
	if _, err := tmpFile.WriteString(outlookServerScript); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to write script: %w", err)
	}
	tmpFile.Close()

	// Set environment variable for port
	env := append(os.Environ(), fmt.Sprintf("OUTLOOK_SERVER_PORT=%d", m.port))

	// Start PowerShell process
	m.cmd = exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", tmpFile.Name())
	m.cmd.Env = env
	// Note: SysProcAttr configuration is Windows-specific and would be set at runtime

	// Start the process
	if err := m.cmd.Start(); err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to start PowerShell: %w", err)
	}

	// Clean up temp file in a goroutine after a delay
	go func() {
		time.Sleep(5 * time.Second)
		os.Remove(tmpFile.Name())
	}()

	// Start process monitor goroutine
	go m.monitorProcess()

	return nil
}

// waitForServer waits for the PowerShell server to be ready
func (m *Manager) waitForServer() error {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		req, _ := http.NewRequestWithContext(ctx, "GET", m.baseURL+"/messages?page=1", nil)
		resp, err := m.client.Do(req)
		cancel()

		if err == nil {
			resp.Body.Close()
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("server did not start within timeout period")
}

// Stop gracefully stops the PowerShell server and supervisor
func (m *Manager) Stop() error {
	m.isShutdown = true

	// Cancel supervisor context to stop all monitoring goroutines
	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	if m.cmd != nil && m.cmd.Process != nil {
		// Send interrupt signal
		if err := m.cmd.Process.Signal(os.Interrupt); err != nil {
			// Force kill if interrupt fails
			return m.cmd.Process.Kill()
		}

		// Wait for process to exit with timeout
		done := make(chan error, 1)
		go func() {
			done <- m.cmd.Wait()
		}()

		select {
		case <-done:
			return nil
		case <-time.After(5 * time.Second):
			return m.cmd.Process.Kill()
		}
	}
	return nil
}

// makeRequest makes an HTTP request to the PowerShell server
func (m *Manager) makeRequest(endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", m.baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if json.Unmarshal(body, &errorResp) == nil {
			return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, errorResp.Error)
		}
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ListMessages retrieves messages from the inbox with pagination
func (m *Manager) ListMessages(page int) (*MessageListResponse, error) {
	if page < 1 {
		page = 1
	}

	endpoint := fmt.Sprintf("/messages?page=%d", page)
	body, err := m.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response MessageListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetMessage retrieves full details of a specific message
func (m *Manager) GetMessage(messageID string) (*Message, error) {
	endpoint := fmt.Sprintf("/messages/%s", url.PathEscape(messageID))
	body, err := m.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &message, nil
}

// GetMessageBody retrieves the readable text content of a message
func (m *Manager) GetMessageBody(messageID string) (*MessageBodyResponse, error) {
	endpoint := fmt.Sprintf("/messages/%s/body", url.PathEscape(messageID))
	body, err := m.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response MessageBodyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetMessageBodyRaw retrieves the raw body content of a message
func (m *Manager) GetMessageBodyRaw(messageID string) (*MessageBodyRawResponse, error) {
	endpoint := fmt.Sprintf("/messages/%s/body/raw", url.PathEscape(messageID))
	body, err := m.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response MessageBodyRawResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// SearchMessages searches for messages matching the query
func (m *Manager) SearchMessages(query string) (*SearchResponse, error) {
	endpoint := fmt.Sprintf("/search?q=%s", url.QueryEscape(query))
	body, err := m.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// supervisorLoop monitors the PowerShell process and restarts it if needed
func (m *Manager) supervisorLoop() {
	for {
		select {
		case <-m.supervisorCtx.Done():
			// Supervisor context cancelled, exit
			return
		case <-m.restartChan:
			if m.isShutdown {
				return
			}

			fmt.Fprintf(os.Stderr, "PowerShell server crashed, attempting restart...\n")

			// Wait a moment before restarting to avoid rapid restart loops
			time.Sleep(2 * time.Second)

			// Attempt to restart the server
			if err := m.restartPowerShellServer(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to restart PowerShell server: %v\n", err)
				// Wait longer before next attempt
				time.Sleep(10 * time.Second)
				// Trigger another restart attempt
				select {
				case m.restartChan <- true:
				default:
				}
			} else {
				fmt.Fprintf(os.Stderr, "PowerShell server restarted successfully\n")
			}
		}
	}
}

// monitorProcess watches the PowerShell process and signals restart if it dies
func (m *Manager) monitorProcess() {
	if m.cmd == nil {
		return
	}

	// Wait for the process to exit
	err := m.cmd.Wait()

	// If we're shutting down, don't attempt restart
	if m.isShutdown {
		return
	}

	fmt.Fprintf(os.Stderr, "PowerShell process exited with error: %v\n", err)

	// Signal supervisor to restart the process
	select {
	case m.restartChan <- true:
	default:
		// Channel full, restart already pending
	}
}

// restartPowerShellServer restarts the PowerShell server process
func (m *Manager) restartPowerShellServer() error {
	// Clean up the old process
	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Kill()
	}

	// Start a new PowerShell server
	if err := m.startPowerShellServer(); err != nil {
		return fmt.Errorf("failed to start new PowerShell server: %w", err)
	}

	// Wait for the new server to be ready
	if err := m.waitForServer(); err != nil {
		return fmt.Errorf("new PowerShell server failed to start: %w", err)
	}

	return nil
}
