package server

import (
	"testing"
)

func TestSetup(t *testing.T) {
	server := ExcelSetup()

	if server == nil {
		t.Fatal("Setup returned nil server")
	}

	// We can only test that the server was created successfully
	// since the Name and Version fields are unexported
}

func TestSetupCreatesServer(t *testing.T) {
	server := ExcelSetup()

	if server == nil {
		t.Fatal("Setup returned nil server")
	}

	// Test that ExcelSetup() consistently returns a server instance
	server2 := ExcelSetup()
	if server2 == nil {
		t.Fatal("Second call to Setup returned nil server")
	}

	// Both calls should return valid server instances
	// (they will be different instances, which is expected)
}
