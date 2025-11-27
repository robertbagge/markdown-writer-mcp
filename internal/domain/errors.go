package domain

import "errors"

// Domain-level errors represent business concerns, not infrastructure details.
// These errors are returned by the application and can be checked with errors.Is.
var (
	// ErrInvalidPath indicates the provided file path is invalid
	ErrInvalidPath = errors.New("invalid file path")

	// ErrPathTraversal indicates a path traversal attempt was detected
	ErrPathTraversal = errors.New("path traversal detected")

	// ErrFileNotFound indicates the requested file does not exist
	ErrFileNotFound = errors.New("file not found")

	// ErrWriteFailed indicates a file write operation failed
	ErrWriteFailed = errors.New("write operation failed")

	// ErrDirCreateFailed indicates directory creation failed
	ErrDirCreateFailed = errors.New("directory creation failed")
)
