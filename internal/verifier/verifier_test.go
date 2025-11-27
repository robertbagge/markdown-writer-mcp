package verifier_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
	"github.com/robertbagge/markdown-writer-mcp/internal/verifier"
)

func TestInMemoryFileVerifier_Verify(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		path      string
		wantSize  int64
		wantLines int
		wantErr   error
	}{
		{
			name: "single line",
			files: map[string]string{
				"/test.md": "# Hello",
			},
			path:      "/test.md",
			wantSize:  7,
			wantLines: 1,
			wantErr:   nil,
		},
		{
			name: "multiple lines",
			files: map[string]string{
				"/multi.md": "# Title\n\n## Section\n\nContent",
			},
			path:      "/multi.md",
			wantSize:  28,
			wantLines: 5,
			wantErr:   nil,
		},
		{
			name: "empty file",
			files: map[string]string{
				"/empty.md": "",
			},
			path:      "/empty.md",
			wantSize:  0,
			wantLines: 0,
			wantErr:   nil,
		},
		{
			name: "file not found",
			files: map[string]string{
				"/exists.md": "content",
			},
			path:      "/notfound.md",
			wantSize:  0,
			wantLines: 0,
			wantErr:   domain.ErrFileNotFound,
		},
		{
			name: "content with trailing newline",
			files: map[string]string{
				"/trail.md": "line1\nline2\n",
			},
			path:      "/trail.md",
			wantSize:  12,
			wantLines: 2,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := verifier.NewInMemoryFileVerifier(tt.files)

			info, err := v.Verify(context.Background(), tt.path)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Verify() unexpected error = %v", err)
				return
			}

			if info.Size != tt.wantSize {
				t.Errorf("Verify() size = %v, want %v", info.Size, tt.wantSize)
			}

			if info.Lines != tt.wantLines {
				t.Errorf("Verify() lines = %v, want %v", info.Lines, tt.wantLines)
			}

			if info.Path != tt.path {
				t.Errorf("Verify() path = %v, want %v", info.Path, tt.path)
			}
		})
	}
}

func TestOSFileVerifier_Verify(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		wantLines int
		wantErr   error
	}{
		{
			name:      "simple file",
			content:   "# Hello World",
			wantLines: 1,
			wantErr:   nil,
		},
		{
			name:      "multi-line file",
			content:   "# Title\n\n## Section\n\nContent here",
			wantLines: 5,
			wantErr:   nil,
		},
		{
			name:      "empty file",
			content:   "",
			wantLines: 0,
			wantErr:   nil,
		},
	}

	v := verifier.NewOSFileVerifier()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			path := filepath.Join(tmpDir, tt.name+".md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			info, err := v.Verify(context.Background(), path)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Verify() unexpected error = %v", err)
				return
			}

			if info.Size != int64(len(tt.content)) {
				t.Errorf("Verify() size = %v, want %v", info.Size, len(tt.content))
			}

			if info.Lines != tt.wantLines {
				t.Errorf("Verify() lines = %v, want %v", info.Lines, tt.wantLines)
			}

			if info.Path != path {
				t.Errorf("Verify() path = %v, want %v", info.Path, path)
			}
		})
	}

	t.Run("file not found", func(t *testing.T) {
		path := filepath.Join(tmpDir, "nonexistent.md")
		_, err := v.Verify(context.Background(), path)
		if !errors.Is(err, domain.ErrFileNotFound) {
			t.Errorf("Verify() error = %v, want %v", err, domain.ErrFileNotFound)
		}
	})
}
