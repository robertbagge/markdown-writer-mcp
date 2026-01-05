package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
	"github.com/robertbagge/markdown-writer-mcp/internal/reader"
	"github.com/robertbagge/markdown-writer-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestJSONReadHandler(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		args        tools.JSONReadArgs
		wantErr     error
		wantContent string
		wantSize    int64
	}{
		{
			name:        "read existing JSON file",
			files:       map[string]string{"/tmp/test.json": `{"name": "test"}`},
			args:        tools.JSONReadArgs{Path: "/tmp/test.json"},
			wantErr:     nil,
			wantContent: `{"name": "test"}`,
			wantSize:    16,
		},
		{
			name:        "read JSON array",
			files:       map[string]string{"/tmp/array.json": `[1, 2, 3]`},
			args:        tools.JSONReadArgs{Path: "/tmp/array.json"},
			wantErr:     nil,
			wantContent: `[1, 2, 3]`,
			wantSize:    9,
		},
		{
			name:        "read non-existent file",
			files:       map[string]string{},
			args:        tools.JSONReadArgs{Path: "/tmp/missing.json"},
			wantErr:     domain.ErrFileNotFound,
			wantContent: "",
			wantSize:    0,
		},
		{
			name:        "path traversal attempt",
			files:       map[string]string{},
			args:        tools.JSONReadArgs{Path: "/tmp/../etc/passwd"},
			wantErr:     domain.ErrPathTraversal,
			wantContent: "",
			wantSize:    0,
		},
		{
			name:        "empty path",
			files:       map[string]string{},
			args:        tools.JSONReadArgs{Path: ""},
			wantErr:     domain.ErrInvalidPath,
			wantContent: "",
			wantSize:    0,
		},
		{
			name:        "read multi-line JSON",
			files:       map[string]string{"/tmp/pretty.json": "{\n  \"name\": \"test\"\n}"},
			args:        tools.JSONReadArgs{Path: "/tmp/pretty.json"},
			wantErr:     nil,
			wantContent: "{\n  \"name\": \"test\"\n}",
			wantSize:    20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup in-memory reader with test files
			memReader := reader.NewInMemoryFileReader()
			memReader.Files = tt.files
			tools.SetFileReader(memReader)

			result, output, err := tools.JSONReadHandler(
				context.Background(),
				&mcp.CallToolRequest{},
				tt.args,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("JSONReadHandler() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("JSONReadHandler() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Error("JSONReadHandler() result is nil")
				return
			}

			// Verify output content
			if output.Content != tt.wantContent {
				t.Errorf("JSONReadHandler() content = %v, want %v", output.Content, tt.wantContent)
			}

			// Verify output size
			if output.Size != tt.wantSize {
				t.Errorf("JSONReadHandler() size = %v, want %v", output.Size, tt.wantSize)
			}

			// Verify result has text content
			if len(result.Content) == 0 {
				t.Error("JSONReadHandler() result has no content")
				return
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Error("JSONReadHandler() result content is not TextContent")
				return
			}

			if textContent.Text != tt.wantContent {
				t.Errorf("JSONReadHandler() result text = %v, want %v", textContent.Text, tt.wantContent)
			}
		})
	}
}
