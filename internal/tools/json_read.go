package tools

import (
	"context"
	"log/slog"

	"github.com/robertbagge/markdown-writer-mcp/internal/pathutil"
	"github.com/robertbagge/markdown-writer-mcp/internal/reader"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// JSONReadTool defines the json_read tool metadata
var JSONReadTool = &mcp.Tool{
	Name:        "json_read",
	Description: "Read JSON file contents from a file path",
}

// JSONReadArgs defines the input parameters for the json_read tool
type JSONReadArgs struct {
	Path string `json:"path" jsonschema:"Absolute or relative path to the JSON file to read"`
}

// JSONReadOutput defines the output structure for the json_read tool
type JSONReadOutput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Size    int64  `json:"size"`
}

// fileReader is injected via SetFileReader (DIP - dependency injection)
var fileReader reader.FileReader

// SetFileReader injects the file reader implementation.
// This follows the Dependency Inversion Principle.
func SetFileReader(r reader.FileReader) {
	fileReader = r
}

// JSONReadHandler handles the json_read tool invocation
func JSONReadHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args JSONReadArgs,
) (*mcp.CallToolResult, JSONReadOutput, error) {
	// Resolve path (validates and converts to absolute)
	absPath, err := pathutil.Resolve(args.Path)
	if err != nil {
		return nil, JSONReadOutput{}, err
	}

	slog.Info("json_read tool called",
		slog.String("path", absPath),
	)

	// Read file using injected reader
	content, err := fileReader.Read(ctx, absPath)
	if err != nil {
		return nil, JSONReadOutput{}, err
	}

	output := JSONReadOutput{
		Path:    absPath,
		Content: content,
		Size:    int64(len(content)),
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: content},
		},
	}

	return result, output, nil
}
