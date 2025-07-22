# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project implementing four Model Context Protocol (MCP) servers for specialized file operations:
- **Document MCP Server**: Clean text extraction from PDF, Word, and PowerPoint files
- **Excel MCP Server**: Spreadsheet reading and manipulation
- **Filesystem MCP Server**: Safe multi-root filesystem access with shell-like navigation
- **Outlook MCP Server**: Windows-only server for Outlook inbox access and message management

## Common Commands

### Build Commands
```bash
# Build all servers
task build

# Build individual servers
task build-excel
task build-fs  
task build-document
task build-outlook       # Windows only

# Cross-platform release builds
task build-release
```

### Development Commands
```bash
# Run servers in development mode (no build step)
task dev-excel
task dev-fs        # Runs with current directory as base
task dev-document
task dev-outlook   # Windows only

# Built and run servers
task run-excel
task run-fs        # Runs with current directory as base  
task run-document
task run-outlook   # Windows only
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
├── fs-mcp/             # Filesystem server executable
└── outlook-mcp/        # Outlook server executable (Windows only)

pkg/                     # Server implementations and shared code
├── document/            # Document processing (PDF, Word, PowerPoint)
├── excel/              # Excel manipulation with excelize
├── filesystem/         # Multi-root filesystem with CWD support
├── outlook/            # Outlook message access (Windows only)
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
- LRU cache with TTL for performance (configurable via env vars and CLI args)
- Tools for reading cells, ranges, columns, rows, and sheets
- Statistical analysis with `get_sheet_stats` tool
- Formula extraction and translation with `explain_formula` tool
- Header discovery for intelligent formula explanations
- Memory management with manual cache flushing
- State management for current sheet operations

**Filesystem Server v2.0** (`pkg/filesystem/`):
- Multi-root directory access with shell-like `cd` and `pwd` functionality
- Current working directory state maintained per session
- Advanced security with path traversal prevention and root boundary enforcement
- Platform-specific implementations (`platform_unix.go`, `platform_windows.go`)

**Outlook Server** (`pkg/outlook/`):
- Windows-only server for Outlook inbox access via COM objects
- PowerShell REST API bridge with embedded script management
- Message navigation, metadata retrieval, and full-text search
- Process lifecycle management with graceful shutdown

### MCP Protocol Implementation
All servers use `github.com/mark3labs/mcp-go v0.34.0` for JSON-RPC communication over stdio. Each server defines tools in `definitions.go` and implements handlers in `handlers.go`.

## Development Notes

### Server Usage Examples
```bash
# Filesystem server with multiple roots
./fs-mcp /Users/kevsmith/repos /Users/kevsmith/Documents /etc

# Excel server with default caching (10 files, 5-minute TTL)
./excel-mcp

# Excel server with custom cache configuration
./excel-mcp --cache-size 20 --cache-ttl 10
EXCEL_CACHE_MAX_SIZE=15 EXCEL_CACHE_TTL_MINUTES=3 ./excel-mcp

# Document server (no arguments needed)  
./document-mcp

# Outlook server (Windows only, no arguments needed)
./outlook-mcp.exe
```

### Excel Server Features

**Cache Management**:
- LRU cache with TTL prevents memory bloat from large Excel files
- Configurable via environment variables or command-line arguments
- Manual cache flushing available via `flush_cache` tool
- Automatic cleanup ticker removes expired entries every minute

**Formula Intelligence**:
- `explain_formula` tool converts cell references to human-readable names
- Example: "=A5*B5" becomes "=quantity*cost" based on headers
- Header discovery searches upward (columns) and leftward (rows) from data cells
- Intelligent caching of header lookups for performance

**Statistical Analysis**:
- `get_sheet_stats` provides comprehensive sheet analytics
- Data type classification (integer, number, text, boolean, date)
- Non-empty cell counts and data boundaries
- First/last data rows and columns identification

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