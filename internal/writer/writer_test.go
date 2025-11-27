package writer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/robertbagge/markdown-writer-mcp/internal/writer"
)

func TestInMemoryFileWriter_Write(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		content     string
		wantSize    int64
		wantContent string
	}{
		{
			name:        "simple write",
			path:        "/tmp/test.md",
			content:     "# Hello World",
			wantSize:    13,
			wantContent: "# Hello World",
		},
		{
			name:        "multi-line content",
			path:        "/test/multi.md",
			content:     "# Title\n\n## Section\n\nContent here",
			wantSize:    33,
			wantContent: "# Title\n\n## Section\n\nContent here",
		},
		{
			name:        "empty content",
			path:        "/empty.md",
			content:     "",
			wantSize:    0,
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := writer.NewInMemoryFileWriter()

			size, err := w.Write(context.Background(), tt.path, tt.content)
			if err != nil {
				t.Errorf("Write() error = %v", err)
				return
			}

			if size != tt.wantSize {
				t.Errorf("Write() size = %v, want %v", size, tt.wantSize)
			}

			if got := w.Files[tt.path]; got != tt.wantContent {
				t.Errorf("Write() content = %v, want %v", got, tt.wantContent)
			}
		})
	}
}

func TestOSFileWriter_Write(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		content string
		wantErr bool
	}{
		{
			name:    "simple write",
			path:    filepath.Join(tmpDir, "test.md"),
			content: "# Hello World",
			wantErr: false,
		},
		{
			name:    "write with subdirectory creation",
			path:    filepath.Join(tmpDir, "sub", "dir", "test.md"),
			content: "# Nested",
			wantErr: false,
		},
		{
			name:    "overwrite existing file",
			path:    filepath.Join(tmpDir, "overwrite.md"),
			content: "original",
			wantErr: false,
		},
	}

	w := writer.NewOSFileWriter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := w.Write(context.Background(), tt.path, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify file exists and has correct content
			content, err := os.ReadFile(tt.path)
			if err != nil {
				t.Errorf("Failed to read written file: %v", err)
				return
			}

			if string(content) != tt.content {
				t.Errorf("File content = %v, want %v", string(content), tt.content)
			}

			if size != int64(len(tt.content)) {
				t.Errorf("Write() size = %v, want %v", size, len(tt.content))
			}
		})
	}

	// Test overwrite behavior
	t.Run("overwrite updates content", func(t *testing.T) {
		path := filepath.Join(tmpDir, "overwrite-test.md")

		// Write first time
		_, err := w.Write(context.Background(), path, "first")
		if err != nil {
			t.Fatalf("First write failed: %v", err)
		}

		// Overwrite
		_, err = w.Write(context.Background(), path, "second")
		if err != nil {
			t.Fatalf("Second write failed: %v", err)
		}

		// Verify content was overwritten
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if string(content) != "second" {
			t.Errorf("Content after overwrite = %v, want %v", string(content), "second")
		}
	})
}
