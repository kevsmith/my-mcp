# My MCP Architecture Documentation

## Overview

This project implements four Model Context Protocol (MCP) servers that provide specialized tools for working with different types of files and data. The architecture follows a modular design where each server focuses on a specific domain: document processing, Excel manipulation, filesystem operations, and Outlook message management.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     MCP Client (Claude)                     │
└─────────────────────┬───────────────────────────────────────┘
                      │ JSON-RPC over stdio
          ┌───────────┼───────────┬───────────┐
          │           │           │           │
    ┌─────▼─────┐ ┌───▼────┐ ┌────▼────────┐ │
    │Document   │ │Excel   │ │Filesystem   │ │
    │MCP Server │ │MCP     │ │MCP Server   │ │
    │           │ │Server  │ │             │ │
    └───────────┘ └────────┘ └─────────────┘ │
                                             │
                                      ┌─────▼──────┐
                                      │Outlook     │
                                      │MCP Server  │
                                      │(Windows)   │
                                      └────────────┘
```

## Project Structure

```
my-mcp/
├── cmd/                        # Entry points for each MCP server
│   ├── document-mcp/main.go    # Document server executable
│   ├── excel-mcp/main.go       # Excel server executable
│   ├── fs-mcp/main.go          # Filesystem server executable
│   └── outlook-mcp/main.go     # Outlook server executable (Windows only)
├── pkg/                        # Shared packages and server implementations
│   ├── common/                 # Common utilities (if any)
│   ├── document/               # Document processing logic
│   ├── excel/                  # Excel manipulation logic
│   ├── filesystem/             # Filesystem operations logic
│   ├── outlook/                # Outlook message management (Windows only)
│   └── server/                 # Server setup and configuration
├── build/                      # Build artifacts
├── Taskfile.yaml              # Task automation (build, test, run)
├── go.mod                     # Go module dependencies
└── go.sum                     # Go module checksums
```

## Server Components

### 1. Document MCP Server (`cmd/document-mcp`)

**Purpose**: Extract clean prose text and metadata from document files (PDF, Word, PowerPoint)

**Key Files**:
- `pkg/document/definitions.go` - Tool definitions
- `pkg/document/handlers.go` - Tool implementations
- `pkg/document/manager.go` - Document processing logic with clean text extraction
- `pkg/document/manager_test.go` - Comprehensive text extraction and cleanup tests
- `pkg/server/document_setup.go` - Server configuration

**Tools Provided**:
- `extract_text` - Extract clean prose text from .pdf, .docx, .pptx files (removes XML markup and formatting)
- `get_document_info` - Get metadata and information about documents

**Text Extraction Features**:
- **Clean Prose Output**: Extracts readable text without XML markup, formatting tags, or document structure
- **Multi-Format Support**: Handles PDF, Word documents (.docx), and PowerPoint presentations (.pptx)
- **Advanced XML Parsing**: Custom XML parser for DOCX files to extract only character data
- **Text Normalization**: Removes excessive whitespace, control characters, and artifacts
- **Legacy Format Handling**: Clear error messages for unsupported .doc and .ppt formats

**Dependencies**:
- `github.com/ledongthuc/pdf` - PDF text extraction
- `github.com/nguyenthenguyen/docx` - DOCX document processing
- `code.sajari.com/docconv` - PowerPoint (.pptx) text extraction

### 2. Excel MCP Server (`cmd/excel-mcp`)

**Purpose**: Read and manipulate Excel spreadsheets

**Key Files**:
- `pkg/excel/definitions.go` - Tool definitions
- `pkg/excel/handlers.go` - Tool implementations  
- `pkg/excel/manager.go` - Excel file management
- `pkg/server/excel_setup.go` - Server configuration

**Tools Provided**:
- `enumerate_columns` - List all columns in a sheet
- `enumerate_rows` - List all rows in a sheet
- `get_cell_value` - Get value of a specific cell
- `get_range_values` - Get values from a cell range
- `list_sheets` - List all sheets in a workbook
- `set_current_sheet` - Set active sheet for operations
- `get_column` - Get all values in a column
- `get_row` - Get all values in a row

**Dependencies**:
- `github.com/xuri/excelize/v2` - Excel file processing

### 3. Filesystem MCP Server (`cmd/fs-mcp`) - v2.0

**Purpose**: Provide safe filesystem access across multiple allowed root directories with shell-like navigation

**Key Files**:
- `pkg/filesystem/definitions.go` - Tool definitions
- `pkg/filesystem/handlers.go` - Tool implementations
- `pkg/filesystem/handler.go` - Multi-root handler with CWD support
- `pkg/server/fs_setup.go` - Server configuration
- `pkg/filesystem/filesystem_test.go` - Comprehensive security and functionality tests

**Navigation Tools**:
- `change_directory` - Navigate between directories like shell `cd` command
- `get_current_directory` - Get current working directory like `pwd` command
- `get_directory_info` - Show CWD and list all allowed root directories

**File Operation Tools**:
- `list_directory` - List files and directories (optional path, defaults to CWD)
- `read_file` - Read file contents (relative to CWD or absolute within roots)
- `get_file_info` - Get file/directory metadata with absolute paths
- `glob` - Find files matching wildcard patterns from CWD

**Multi-Root Architecture**:
- **Multiple Allowed Roots**: Access multiple top-level directories simultaneously
- **Current Working Directory**: Maintains session state for intuitive navigation
- **Flexible Path Resolution**: Accepts both relative (from CWD) and absolute paths
- **Consistent Output**: Always returns absolute paths with optional relative display

**Advanced Security Features**:
- **Multi-Layer Path Validation**: `filepath.Clean()` + absolute path resolution + prefix checking
- **Path Traversal Prevention**: Mathematically impossible to escape allowed roots via `../../../`
- **Root Boundary Enforcement**: All operations restricted to specified allowed roots
- **Symlink Protection**: Optional validation against symlink-based escapes
- **Comprehensive Testing**: Full test coverage for attack vectors and edge cases

**Usage Examples**:
```bash
# Multi-root server startup
fs-mcp /Users/kevsmith/repos /Users/kevsmith/Documents /etc

# Shell-like navigation
change_directory("my-project/src")
list_directory()                    # Lists current directory contents
read_file("main.go")               # Relative to current directory

# Cross-root access
change_directory("/Users/kevsmith/Documents")
read_file("/etc/hosts")            # Absolute path within allowed roots
```

### 4. Outlook MCP Server (`cmd/outlook-mcp`) - Windows Only

**Purpose**: Provide access to Microsoft Outlook inbox for message navigation, metadata retrieval, and search

**Key Files**:
- `pkg/outlook/definitions.go` - Tool definitions for Outlook operations
- `pkg/outlook/handlers.go` - Tool implementations for message access
- `pkg/outlook/manager.go` - PowerShell process lifecycle and REST client
- `pkg/outlook/types.go` - Type definitions for Outlook data structures
- `pkg/outlook/scripts/outlook-server.ps1` - Embedded PowerShell REST API server
- `pkg/server/outlook_setup.go` - Server configuration and setup

**MCP Tools Provided**:
- `list_messages` - List inbox messages with pagination (page size: 10)
- `get_message` - Get full message details including metadata and preview
- `get_message_body` - Get readable text content of a message (cooked)
- `get_message_body_raw` - Get raw message body content (HTML and plain text)
- `search_messages` - Search messages by subject, body, or sender

**Architecture Components**:
- **Embedded PowerShell Server**: REST API server embedded as Go binary resource
- **COM Object Integration**: Direct access to Outlook via COM automation objects
- **Process Lifecycle Management**: Automatic PowerShell server startup/shutdown
- **REST API Bridge**: HTTP client in Go communicates with PowerShell REST endpoints
- **Graceful Degradation**: Continues operation with error responses when Outlook unavailable

**REST API Endpoints** (Internal PowerShell Server):
- `GET /messages?page=N` - Paginated message listing
- `GET /messages/{id}` - Full message details with preview
- `GET /messages/{id}/body` - Readable message body text
- `GET /messages/{id}/body/raw` - Raw message body (HTML/plain text)
- `GET /search?q={query}` - Message search functionality

**Security & Configuration**:
- **Windows-Only Operation**: Runtime OS validation prevents non-Windows execution  
- **Localhost Binding**: PowerShell REST API only accessible from localhost
- **Configurable Port**: Uses `OUTLOOK_SERVER_PORT` environment variable (default: 8080)
- **Process Isolation**: PowerShell server runs in separate process with proper cleanup
- **Temporary Script Management**: Embedded script written to temp file and cleaned up

**Usage Examples**:
```bash
# Start Outlook server (Windows only)
outlook-mcp.exe

# Configure custom port
set OUTLOOK_SERVER_PORT=9090
outlook-mcp.exe

# Development mode
task dev-outlook
```

**Error Handling**:
- Server starts even when Outlook is unavailable
- Clear error messages returned when Outlook COM objects cannot be accessed
- Proper HTTP status codes and structured error responses
- Graceful PowerShell process termination on shutdown

## Core Dependencies

### MCP Framework
- `github.com/mark3labs/mcp-go v0.34.0` - Go implementation of Model Context Protocol

### Document Processing
- `github.com/ledongthuc/pdf` - PDF text extraction
- `github.com/nguyenthenguyen/docx` - DOCX document processing  
- `code.sajari.com/docconv v1.3.8` - PowerPoint (.pptx) text extraction and document conversion

### Excel Processing  
- `github.com/xuri/excelize/v2 v2.9.1` - Excel file manipulation

### Utilities
- `github.com/google/uuid v1.6.0` - UUID generation
- `github.com/spf13/cast v1.7.1` - Type casting utilities

## Communication Protocol

All servers communicate using the Model Context Protocol (MCP) over standard input/output:

1. **Transport**: JSON-RPC 2.0 over stdio
2. **Initialization**: Server capabilities negotiation
3. **Tool Discovery**: Client queries available tools
4. **Tool Execution**: Client invokes tools with parameters
5. **Response**: Server returns structured results

## Build and Deployment

### Build System
- **Task Runner**: `Taskfile.yaml` provides consistent build commands
- **Multi-Platform**: Supports Linux, macOS, and Windows builds
- **Release Builds**: Optimized binaries with `-ldflags="-s -w"`

### Key Build Commands
```bash
task build           # Build all servers
task build-excel     # Build Excel server only
task build-fs        # Build filesystem server only  
task build-document  # Build document server only
task build-outlook   # Build Outlook server only (Windows)
task build-release   # Cross-platform release builds
```

### Development Commands
```bash
task dev-excel       # Run Excel server in development mode
task dev-fs          # Run filesystem server v2.0 in development mode with current directory
task dev-document    # Run document server in development mode
task dev-outlook     # Run Outlook server in development mode (Windows)
```

### Server Usage Examples
```bash
# Filesystem Server v2.0 - Multi-root access
fs-mcp /Users/kevsmith/repos /Users/kevsmith/Documents /etc

# Filesystem Server v2.0 - Single root (backward compatible)
fs-mcp /Users/kevsmith/projects

# Excel Server - Spreadsheet processing
excel-mcp

# Document Server - Clean text extraction from PDF, Word, PowerPoint
document-mcp

# Outlook Server - Windows Outlook message access
outlook-mcp.exe
```

## Security Considerations

### Filesystem Server v2.0
- **Multi-Root Security**: All operations restricted to explicitly allowed root directories
- **Advanced Path Traversal Prevention**: Multi-layer validation using `filepath.Clean()` + absolute resolution + prefix matching
- **Mathematical Escape Prevention**: Impossible to traverse outside roots via `../../../` due to absolute path validation
- **Current Directory Isolation**: CWD changes cannot escape allowed roots
- **Comprehensive Security Testing**: Full test coverage for attack vectors and edge cases
- **Platform Security**: Separate implementations for Unix and Windows with consistent security model

### Document Server  
- **Read-Only Operations**: All tools are marked as read-only
- **File Type Validation**: Only processes supported document formats

### Excel Server
- **Read-Only Default**: Most operations are read-only
- **State Management**: Careful handling of current sheet state

## Testing Strategy

- **Unit Tests**: Each package has corresponding `*_test.go` files
- **Integration Tests**: End-to-end testing of MCP protocol
- **Coverage Reports**: `task test-coverage` generates HTML coverage reports
- **Benchmarks**: Performance testing with `task benchmark`

## Extension Points

The architecture is designed for extensibility:

1. **New Tool Types**: Add new tools by extending definitions and handlers
2. **New File Formats**: Document server can be extended for additional formats
3. **New Servers**: Follow the same pattern to create servers for other domains
4. **Enhanced Security**: Additional validation and sandboxing can be added

## Configuration

- **Server Metadata**: Name and version defined in setup functions
- **Tool Capabilities**: Configured per server (read-only hints, etc.)
- **Base Paths**: Filesystem server accepts runtime base directory configuration
- **Logging**: Optional logging capabilities available