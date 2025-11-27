package writer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
)

// FileWriter defines the behavior for writing markdown files.
// Interface is defined at the usage point (consumer-defined interface).
type FileWriter interface {
	Write(ctx context.Context, path, content string) (int64, error)
}

// OSFileWriter implements FileWriter using the OS file system with atomic writes.
type OSFileWriter struct{}

// NewOSFileWriter creates a new OS file system writer.
func NewOSFileWriter() *OSFileWriter {
	return &OSFileWriter{}
}

// Write writes content to a file atomically using a temporary file and rename.
// This ensures the file is either fully written or not written at all.
func (w *OSFileWriter) Write(ctx context.Context, path, content string) (int64, error) {
	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrDirCreateFailed, err)
	}

	// Create temporary file in the same directory for atomic rename
	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+".tmp.*")
	if err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrWriteFailed, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Cleanup temp file on error

	// Write content to temporary file
	n, err := tmpFile.WriteString(content)
	if err != nil {
		tmpFile.Close()
		return 0, fmt.Errorf("%w: %v", domain.ErrWriteFailed, err)
	}

	// Close the temp file before renaming
	if err := tmpFile.Close(); err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrWriteFailed, err)
	}

	// Atomically rename temp file to target path
	if err := os.Rename(tmpPath, path); err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrWriteFailed, err)
	}

	return int64(n), nil
}

// InMemoryFileWriter is a fake implementation for testing.
// It stores files in memory rather than on disk.
type InMemoryFileWriter struct {
	Files map[string]string // path -> content
}

// NewInMemoryFileWriter creates a new in-memory file writer for testing.
func NewInMemoryFileWriter() *InMemoryFileWriter {
	return &InMemoryFileWriter{
		Files: make(map[string]string),
	}
}

// Write stores the content in memory.
func (w *InMemoryFileWriter) Write(ctx context.Context, path, content string) (int64, error) {
	w.Files[path] = content
	return int64(len(content)), nil
}
