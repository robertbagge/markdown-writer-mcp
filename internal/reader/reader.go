package reader

import (
	"context"
	"fmt"
	"os"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
)

// FileReader defines the behavior for reading files.
// Interface is defined at the usage point (consumer-defined interface).
type FileReader interface {
	Read(ctx context.Context, path string) (string, error)
}

// OSFileReader implements FileReader using the OS file system.
type OSFileReader struct{}

// NewOSFileReader creates a new OS file system reader.
func NewOSFileReader() *OSFileReader {
	return &OSFileReader{}
}

// Read reads the entire content of a file and returns it as a string.
func (r *OSFileReader) Read(ctx context.Context, path string) (string, error) {
	// Check for context cancellation before reading
	if err := ctx.Err(); err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", domain.ErrFileNotFound
		}
		return "", fmt.Errorf("%w: %v", domain.ErrReadFailed, err)
	}
	return string(content), nil
}

// InMemoryFileReader is a fake implementation for testing.
// It stores files in memory rather than on disk.
type InMemoryFileReader struct {
	Files map[string]string // path -> content
}

// NewInMemoryFileReader creates a new in-memory file reader for testing.
func NewInMemoryFileReader() *InMemoryFileReader {
	return &InMemoryFileReader{
		Files: make(map[string]string),
	}
}

// Read returns the content from memory.
func (r *InMemoryFileReader) Read(ctx context.Context, path string) (string, error) {
	content, ok := r.Files[path]
	if !ok {
		return "", domain.ErrFileNotFound
	}
	return content, nil
}
