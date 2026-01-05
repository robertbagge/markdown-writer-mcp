package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all available tools with the MCP server
func RegisterAll(server *mcp.Server) error {
	// Register write tool (markdown)
	mcp.AddTool(server, WriteTool, WriteHandler)

	// Register verify tool (markdown)
	mcp.AddTool(server, VerifyTool, VerifyHandler)

	// Register json_read tool
	mcp.AddTool(server, JSONReadTool, JSONReadHandler)

	// Register json_write tool
	mcp.AddTool(server, JSONWriteTool, JSONWriteHandler)

	return nil
}
