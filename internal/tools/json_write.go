package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
	"github.com/robertbagge/markdown-writer-mcp/internal/pathutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// JSONWriteTool defines the json_write tool metadata
var JSONWriteTool = &mcp.Tool{
	Name:        "json_write",
	Description: "Write JSON content to a file path with validation and atomic writes",
}

// JSONWriteArgs defines the input parameters for the json_write tool
type JSONWriteArgs struct {
	Path    string `json:"path" jsonschema:"Absolute or relative path to the JSON file to write"`
	Content string `json:"content" jsonschema:"JSON content to write to the file (must be valid JSON)"`
}

// JSONWriteOutput defines the output structure for the json_write tool
type JSONWriteOutput struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// JSONWriteHandler handles the json_write tool invocation
func JSONWriteHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args JSONWriteArgs,
) (*mcp.CallToolResult, JSONWriteOutput, error) {
	// Resolve path (validates and converts to absolute)
	absPath, err := pathutil.Resolve(args.Path)
	if err != nil {
		return nil, JSONWriteOutput{}, err
	}

	// Validate that content is valid JSON
	if !json.Valid([]byte(args.Content)) {
		return nil, JSONWriteOutput{}, domain.ErrInvalidJSON
	}

	slog.Info("json_write tool called",
		slog.String("path", absPath),
		slog.Int("content_length", len(args.Content)),
	)

	// Write file using the existing fileWriter (reused from write.go)
	size, err := fileWriter.Write(ctx, absPath, args.Content)
	if err != nil {
		return nil, JSONWriteOutput{}, err
	}

	output := JSONWriteOutput{
		Path: absPath,
		Size: size,
	}

	message := fmt.Sprintf("Successfully wrote %d bytes of JSON to %s", size, absPath)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}

	return result, output, nil
}
