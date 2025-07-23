package shared

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
)

// JSONEncoder provides optimized JSON encoding with pooled resources
type JSONEncoder struct {
	bufferPool  sync.Pool
	encoderPool sync.Pool
}

// NewJSONEncoder creates a new optimized JSON encoder
func NewJSONEncoder() *JSONEncoder {
	je := &JSONEncoder{
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 1024)) // Pre-allocate 1KB
			},
		},
		encoderPool: sync.Pool{
			New: func() interface{} {
				return json.NewEncoder(bytes.NewBuffer(nil))
			},
		},
	}
	return je
}

// Marshal performs optimized JSON marshaling using pooled buffers
func (je *JSONEncoder) Marshal(v interface{}) ([]byte, error) {
	buf := je.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer je.bufferPool.Put(buf)

	encoder := json.NewEncoder(buf)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	// Remove trailing newline added by json.Encoder.Encode
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Make a copy since buf will be returned to pool
	output := make([]byte, len(result))
	copy(output, result)
	return output, nil
}

// MarshalIndent performs optimized indented JSON marshaling
func (je *JSONEncoder) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	buf := je.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer je.bufferPool.Put(buf)

	encoder := json.NewEncoder(buf)
	encoder.SetIndent(prefix, indent)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	// Remove trailing newline added by json.Encoder.Encode
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Make a copy since buf will be returned to pool
	output := make([]byte, len(result))
	copy(output, result)
	return output, nil
}

// Unmarshal performs optimized JSON unmarshaling with type checking
func (je *JSONEncoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// UnmarshalRequest unmarshals MCP request arguments into a struct with validation
func (je *JSONEncoder) UnmarshalRequest(request mcp.CallToolRequest, args interface{}) error {
	// Use reflection to validate the args type
	argType := reflect.TypeOf(args)
	if argType.Kind() != reflect.Ptr || argType.Elem().Kind() != reflect.Struct {
		return &json.UnmarshalTypeError{
			Value:  "request arguments",
			Type:   argType,
			Offset: 0,
		}
	}

	// Marshal the request arguments first (this is required by the MCP framework)
	argBytes, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return err
	}

	// Unmarshal into the provided struct
	return je.Unmarshal(argBytes, args)
}

// NewToolResultJSON creates an MCP tool result with optimized JSON response
func (je *JSONEncoder) NewToolResultJSON(data interface{}) (*mcp.CallToolResult, error) {
	jsonBytes, err := je.Marshal(data)
	if err != nil {
		return mcp.NewToolResultError("Failed to marshal response: " + err.Error()), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// NewToolResultJSONIndent creates an MCP tool result with indented JSON response
func (je *JSONEncoder) NewToolResultJSONIndent(data interface{}, prefix, indent string) (*mcp.CallToolResult, error) {
	jsonBytes, err := je.MarshalIndent(data, prefix, indent)
	if err != nil {
		return mcp.NewToolResultError("Failed to marshal response: " + err.Error()), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// Global optimized encoder instance
var GlobalJSONEncoder = NewJSONEncoder()

// Convenience functions using the global encoder

// OptimizedMarshal provides globally optimized JSON marshaling
func OptimizedMarshal(v interface{}) ([]byte, error) {
	return GlobalJSONEncoder.Marshal(v)
}

// OptimizedMarshalIndent provides globally optimized indented JSON marshaling
func OptimizedMarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return GlobalJSONEncoder.MarshalIndent(v, prefix, indent)
}

// OptimizedUnmarshal provides globally optimized JSON unmarshaling
func OptimizedUnmarshal(data []byte, v interface{}) error {
	return GlobalJSONEncoder.Unmarshal(data, v)
}

// OptimizedUnmarshalRequest unmarshals MCP request arguments with global encoder
func OptimizedUnmarshalRequest(request mcp.CallToolRequest, args interface{}) error {
	return GlobalJSONEncoder.UnmarshalRequest(request, args)
}

// OptimizedToolResultJSON creates optimized JSON tool result with global encoder
func OptimizedToolResultJSON(data interface{}) (*mcp.CallToolResult, error) {
	return GlobalJSONEncoder.NewToolResultJSON(data)
}

// OptimizedToolResultJSONIndent creates optimized indented JSON tool result with global encoder
func OptimizedToolResultJSONIndent(data interface{}, prefix, indent string) (*mcp.CallToolResult, error) {
	return GlobalJSONEncoder.NewToolResultJSONIndent(data, prefix, indent)
}
