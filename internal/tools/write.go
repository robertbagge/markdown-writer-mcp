package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/robertbagge/markdown-writer-mcp/internal/pathutil"
	"github.com/robertbagge/markdown-writer-mcp/internal/writer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WriteTool defines the write tool metadata
var WriteTool = &mcp.Tool{
	Name:        "write",
	Description: "Write markdown content to a file path with atomic writes",
}

// WriteArgs defines the input parameters for the write tool
type WriteArgs struct {
	Path    string `json:"path" jsonschema:"Absolute or relative path to the markdown file to write"`
	Content string `json:"content" jsonschema:"Markdown content to write to the file"`
}

// WriteOutput defines the output structure for the write tool
type WriteOutput struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// fileWriter is injected via SetFileWriter (DIP - dependency injection)
var fileWriter writer.FileWriter

// SetFileWriter injects the file writer implementation.
// This follows the Dependency Inversion Principle.
func SetFileWriter(w writer.FileWriter) {
	fileWriter = w
}

// WriteHandler handles the write tool invocation
func WriteHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args WriteArgs,
) (*mcp.CallToolResult, WriteOutput, error) {
	// Resolve path (validates and converts to absolute)
	absPath, err := pathutil.Resolve(args.Path)
	if err != nil {
		return nil, WriteOutput{}, err
	}

	slog.Info("write tool called",
		slog.String("path", absPath),
		slog.Int("content_length", len(args.Content)),
	)

	// Write file using injected writer
	size, err := fileWriter.Write(ctx, absPath, args.Content)
	if err != nil {
		return nil, WriteOutput{}, err
	}

	output := WriteOutput{
		Path: absPath,
		Size: size,
	}

	message := fmt.Sprintf("Successfully wrote %d bytes to %s", size, absPath)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}

	return result, output, nil
}
