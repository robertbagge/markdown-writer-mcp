package verifier

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
)

// FileInfo contains information about a verified file.
type FileInfo struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Lines int    `json:"lines"`
}

// FileVerifier defines the behavior for verifying markdown files.
// Interface is defined at the usage point (consumer-defined interface).
type FileVerifier interface {
	Verify(ctx context.Context, path string) (*FileInfo, error)
}

// OSFileVerifier implements FileVerifier using the OS file system.
type OSFileVerifier struct{}

// NewOSFileVerifier creates a new OS file system verifier.
func NewOSFileVerifier() *OSFileVerifier {
	return &OSFileVerifier{}
}

// Verify checks if a file exists and returns its statistics.
func (v *OSFileVerifier) Verify(ctx context.Context, path string) (*FileInfo, error) {
	// Check if file exists and get stats
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrFileNotFound
		}
		return nil, fmt.Errorf("verify failed: %w", err)
	}

	// Open file to count lines
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("verify failed: %w", err)
	}
	defer file.Close()

	// Count lines
	lines := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("verify failed: %w", err)
	}

	return &FileInfo{
		Path:  path,
		Size:  stat.Size(),
		Lines: lines,
	}, nil
}

// InMemoryFileVerifier is a fake implementation for testing.
// It verifies files stored in memory.
type InMemoryFileVerifier struct {
	Files map[string]string // path -> content
}

// NewInMemoryFileVerifier creates a new in-memory file verifier for testing.
func NewInMemoryFileVerifier(files map[string]string) *InMemoryFileVerifier {
	if files == nil {
		files = make(map[string]string)
	}
	return &InMemoryFileVerifier{
		Files: files,
	}
}

// Verify checks if a file exists in memory and returns its statistics.
func (v *InMemoryFileVerifier) Verify(ctx context.Context, path string) (*FileInfo, error) {
	content, ok := v.Files[path]
	if !ok {
		return nil, domain.ErrFileNotFound
	}

	// Count lines (including last line even if it doesn't end with newline)
	lines := 0
	if len(content) > 0 {
		lines = strings.Count(content, "\n") + 1
		// If content ends with newline, don't count the extra empty line
		if strings.HasSuffix(content, "\n") {
			lines--
		}
		// Ensure at least one line if content is not empty
		if lines == 0 {
			lines = 1
		}
	}

	return &FileInfo{
		Path:  path,
		Size:  int64(len(content)),
		Lines: lines,
	}, nil
}
