package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/robertbagge/markdown-writer-mcp/internal/pathutil"
	"github.com/robertbagge/markdown-writer-mcp/internal/verifier"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// VerifyTool defines the verify tool metadata
var VerifyTool = &mcp.Tool{
	Name:        "verify",
	Description: "Verify that a markdown file exists and get its statistics",
}

// VerifyArgs defines the input parameters for the verify tool
type VerifyArgs struct {
	Path string `json:"path" jsonschema:"Absolute or relative path to the markdown file to verify"`
}

// VerifyOutput defines the output structure for the verify tool
type VerifyOutput struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Lines int    `json:"lines"`
}

// fileVerifier is injected via SetFileVerifier (DIP - dependency injection)
var fileVerifier verifier.FileVerifier

// SetFileVerifier injects the file verifier implementation.
// This follows the Dependency Inversion Principle.
func SetFileVerifier(v verifier.FileVerifier) {
	fileVerifier = v
}

// VerifyHandler handles the verify tool invocation
func VerifyHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args VerifyArgs,
) (*mcp.CallToolResult, VerifyOutput, error) {
	// Resolve path (validates and converts to absolute)
	absPath, err := pathutil.Resolve(args.Path)
	if err != nil {
		return nil, VerifyOutput{}, err
	}

	slog.Info("verify tool called",
		slog.String("path", absPath),
	)

	// Verify file using injected verifier
	info, err := fileVerifier.Verify(ctx, absPath)
	if err != nil {
		return nil, VerifyOutput{}, err
	}

	output := VerifyOutput{
		Path:  info.Path,
		Size:  info.Size,
		Lines: info.Lines,
	}

	message := fmt.Sprintf("File verified: %s (%d bytes, %d lines)", info.Path, info.Size, info.Lines)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}

	return result, output, nil
}
