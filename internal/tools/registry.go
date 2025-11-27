package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all available tools with the MCP server
func RegisterAll(server *mcp.Server) error {
	// Register write tool
	mcp.AddTool(server, WriteTool, WriteHandler)

	// Register verify tool
	mcp.AddTool(server, VerifyTool, VerifyHandler)

	return nil
}
