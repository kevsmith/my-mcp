# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project implementing three Model Context Protocol (MCP) servers for specialized file operations:
- **Document MCP Server**: Clean text extraction from PDF, Word, and PowerPoint files
- **Excel MCP Server**: Spreadsheet reading and manipulation
- **Filesystem MCP Server**: Safe multi-root filesystem access with shell-like navigation

## Common Commands

### Build Commands
```bash
# Build all servers
task build

# Build individual servers
task build-excel
task build-fs  
task build-document

# Cross-platform release builds
task build-release
```

### Development Commands
```bash
# Run servers in development mode (no build step)
task dev-excel
task dev-fs        # Runs with current directory as base
task dev-document

# Built and run servers
task run-excel
task run-fs        # Runs with current directory as base  
task run-document
```

### Testing Commands
```bash
# Run all tests
task test

# Run tests with coverage report (generates coverage.html)
task test-coverage

# Run format, vet, and test together
task check

# Run benchmarks
task benchmark
```

### Code Quality Commands
```bash
# Format Go code
task fmt

# Run go vet
task vet

# Clean build artifacts and test cache
task clean

# Update dependencies
task mod-update
```

## Architecture

### High-Level Structure
```
cmd/                     # Entry points for each MCP server
├── document-mcp/        # Document server executable
├── excel-mcp/          # Excel server executable  
└── fs-mcp/             # Filesystem server executable

pkg/                     # Server implementations and shared code
├── document/            # Document processing (PDF, Word, PowerPoint)
├── excel/              # Excel manipulation with excelize
├── filesystem/         # Multi-root filesystem with CWD support
└── server/             # Server setup and configuration
```

### Key Components

**Document Server** (`pkg/document/`):
- Clean prose text extraction from .pdf, .docx, .pptx files
- XML markup removal and text normalization
- Manager handles document processing with comprehensive cleanup
- Uses `github.com/ledongthuc/pdf`, `github.com/nguyenthenguyen/docx`, and `code.sajari.com/docconv`

**Excel Server** (`pkg/excel/`):
- Spreadsheet operations using `github.com/xuri/excelize/v2`
- Tools for reading cells, ranges, columns, rows, and sheets
- State management for current sheet operations

**Filesystem Server v2.0** (`pkg/filesystem/`):
- Multi-root directory access with shell-like `cd` and `pwd` functionality
- Current working directory state maintained per session
- Advanced security with path traversal prevention and root boundary enforcement
- Platform-specific implementations (`platform_unix.go`, `platform_windows.go`)

### MCP Protocol Implementation
All servers use `github.com/mark3labs/mcp-go v0.34.0` for JSON-RPC communication over stdio. Each server defines tools in `definitions.go` and implements handlers in `handlers.go`.

## Development Notes

### Server Usage Examples
```bash
# Filesystem server with multiple roots
./fs-mcp /Users/kevsmith/repos /Users/kevsmith/Documents /etc

# Excel server (no arguments needed)
./excel-mcp

# Document server (no arguments needed)  
./document-mcp
```

### Testing Strategy
- Unit tests in `*_test.go` files alongside implementation
- Comprehensive security testing for filesystem server
- Coverage reports available via `task test-coverage`
- Integration tests verify MCP protocol compliance

### Security Considerations
The filesystem server implements multi-layer security:
- `filepath.Clean()` + absolute path resolution + prefix checking
- Mathematical prevention of path traversal attacks
- All operations restricted to explicitly allowed root directories
- Optional symlink validation