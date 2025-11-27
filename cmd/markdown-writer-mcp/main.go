package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/robertbagge/markdown-writer-mcp/internal/tools"
	"github.com/robertbagge/markdown-writer-mcp/internal/verifier"
	"github.com/robertbagge/markdown-writer-mcp/internal/writer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Setup structured logging
	setupLogger()

	slog.Info("Starting markdown-writer MCP server")

	// Wire dependencies (constructor injection following DIP)
	fileWriter := writer.NewOSFileWriter()
	fileVerifier := verifier.NewOSFileVerifier()

	// Inject dependencies into tools
	tools.SetFileWriter(fileWriter)
	tools.SetFileVerifier(fileVerifier)

	// Create MCP server instance
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "markdown-writer",
			Version: "1.0.0",
		},
		nil, // options
	)

	// Register all tools
	if err := tools.RegisterAll(server); err != nil {
		slog.Error("failed to register tools", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("tools registered successfully")

	// Run server with stdio transport
	ctx := context.Background()
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		slog.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}
}

func setupLogger() {
	// JSON handler for structured logging - writes to stderr
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
