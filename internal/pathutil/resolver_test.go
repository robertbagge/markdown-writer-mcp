package pathutil_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
	"github.com/robertbagge/markdown-writer-mcp/internal/pathutil"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "valid absolute path",
			path:    "/tmp/test.md",
			wantErr: nil,
		},
		{
			name:    "valid relative path",
			path:    "test.md",
			wantErr: nil,
		},
		{
			name:    "path traversal with ..",
			path:    "../etc/passwd",
			wantErr: domain.ErrPathTraversal,
		},
		{
			name:    "path traversal in middle",
			path:    "/tmp/../etc/passwd",
			wantErr: domain.ErrPathTraversal,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: domain.ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pathutil.Resolve(tt.path)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Resolve() unexpected error = %v", err)
				return
			}

			// Result should be absolute
			if !filepath.IsAbs(result) {
				t.Errorf("Resolve() result = %v is not absolute", result)
			}
		})
	}
}
