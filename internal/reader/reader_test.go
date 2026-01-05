package reader_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
	"github.com/robertbagge/markdown-writer-mcp/internal/reader"
)

func TestInMemoryFileReader_Read(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		path        string
		wantContent string
		wantErr     error
	}{
		{
			name:        "read existing file",
			files:       map[string]string{"/tmp/test.json": `{"key": "value"}`},
			path:        "/tmp/test.json",
			wantContent: `{"key": "value"}`,
			wantErr:     nil,
		},
		{
			name:        "read non-existent file",
			files:       map[string]string{},
			path:        "/tmp/missing.json",
			wantContent: "",
			wantErr:     domain.ErrFileNotFound,
		},
		{
			name:        "read empty file",
			files:       map[string]string{"/empty.json": ""},
			path:        "/empty.json",
			wantContent: "",
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reader.NewInMemoryFileReader()
			r.Files = tt.files

			content, err := r.Read(context.Background(), tt.path)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Read() unexpected error = %v", err)
				return
			}

			if content != tt.wantContent {
				t.Errorf("Read() content = %v, want %v", content, tt.wantContent)
			}
		})
	}
}

func TestOSFileReader_Read(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("read existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.json")
		content := `{"name": "test", "value": 123}`

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		r := reader.NewOSFileReader()
		got, err := r.Read(context.Background(), path)
		if err != nil {
			t.Errorf("Read() error = %v", err)
			return
		}

		if got != content {
			t.Errorf("Read() content = %v, want %v", got, content)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		r := reader.NewOSFileReader()
		_, err := r.Read(context.Background(), filepath.Join(tmpDir, "missing.json"))

		if !errors.Is(err, domain.ErrFileNotFound) {
			t.Errorf("Read() error = %v, want %v", err, domain.ErrFileNotFound)
		}
	})

	t.Run("read multi-line content", func(t *testing.T) {
		path := filepath.Join(tmpDir, "multiline.json")
		content := `{
  "name": "test",
  "items": [1, 2, 3]
}`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		r := reader.NewOSFileReader()
		got, err := r.Read(context.Background(), path)
		if err != nil {
			t.Errorf("Read() error = %v", err)
			return
		}

		if got != content {
			t.Errorf("Read() content = %v, want %v", got, content)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		r := reader.NewOSFileReader()
		_, err := r.Read(ctx, filepath.Join(tmpDir, "any.json"))

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Read() error = %v, want context.Canceled", err)
		}
	})
}
