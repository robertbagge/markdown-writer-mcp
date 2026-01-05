package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
	"github.com/robertbagge/markdown-writer-mcp/internal/tools"
	"github.com/robertbagge/markdown-writer-mcp/internal/writer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestJSONWriteHandler(t *testing.T) {
	// Setup in-memory writer
	memWriter := writer.NewInMemoryFileWriter()
	tools.SetFileWriter(memWriter)

	tests := []struct {
		name        string
		args        tools.JSONWriteArgs
		wantErr     error
		wantContent string
	}{
		{
			name: "valid JSON object",
			args: tools.JSONWriteArgs{
				Path:    "/tmp/test.json",
				Content: `{"name": "test", "value": 123}`,
			},
			wantErr:     nil,
			wantContent: `{"name": "test", "value": 123}`,
		},
		{
			name: "valid JSON array",
			args: tools.JSONWriteArgs{
				Path:    "/tmp/array.json",
				Content: `[1, 2, 3, "four"]`,
			},
			wantErr:     nil,
			wantContent: `[1, 2, 3, "four"]`,
		},
		{
			name: "valid JSON string",
			args: tools.JSONWriteArgs{
				Path:    "/tmp/string.json",
				Content: `"hello world"`,
			},
			wantErr:     nil,
			wantContent: `"hello world"`,
		},
		{
			name: "valid JSON number",
			args: tools.JSONWriteArgs{
				Path:    "/tmp/number.json",
				Content: `42`,
			},
			wantErr:     nil,
			wantContent: `42`,
		},
		{
			name: "invalid JSON - missing quote",
			args: tools.JSONWriteArgs{
				Path:    "/tmp/invalid.json",
				Content: `{"name": "test`,
			},
			wantErr:     domain.ErrInvalidJSON,
			wantContent: "",
		},
		{
			name: "invalid JSON - plain text",
			args: tools.JSONWriteArgs{
				Path:    "/tmp/invalid2.json",
				Content: `hello world`,
			},
			wantErr:     domain.ErrInvalidJSON,
			wantContent: "",
		},
		{
			name: "path traversal attempt",
			args: tools.JSONWriteArgs{
				Path:    "/tmp/../etc/passwd",
				Content: `{}`,
			},
			wantErr:     domain.ErrPathTraversal,
			wantContent: "",
		},
		{
			name: "empty path",
			args: tools.JSONWriteArgs{
				Path:    "",
				Content: `{}`,
			},
			wantErr:     domain.ErrInvalidPath,
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous writes
			memWriter.Files = make(map[string]string)

			result, output, err := tools.JSONWriteHandler(
				context.Background(),
				&mcp.CallToolRequest{},
				tt.args,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("JSONWriteHandler() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("JSONWriteHandler() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Error("JSONWriteHandler() result is nil")
				return
			}

			// Verify output size matches content length
			if output.Size != int64(len(tt.wantContent)) {
				t.Errorf("JSONWriteHandler() size = %v, want %v", output.Size, len(tt.wantContent))
			}

			// Verify content was written correctly
			if got := memWriter.Files[output.Path]; got != tt.wantContent {
				t.Errorf("JSONWriteHandler() written content = %v, want %v", got, tt.wantContent)
			}
		})
	}
}
