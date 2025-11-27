package pathutil

import (
	"path/filepath"
	"strings"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
)

// Resolve converts a relative or absolute path to a clean absolute path
// and validates that it doesn't contain path traversal attempts.
func Resolve(path string) (string, error) {
	if path == "" {
		return "", domain.ErrInvalidPath
	}

	// Prevent path traversal attacks
	if strings.Contains(path, "..") {
		return "", domain.ErrPathTraversal
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", domain.ErrInvalidPath
	}

	// Clean the path (remove redundant separators, resolve . and ..)
	cleanPath := filepath.Clean(absPath)

	return cleanPath, nil
}
