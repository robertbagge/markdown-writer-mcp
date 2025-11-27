# markdown-writer-mcp

An MCP (Model Context Protocol) server for writing markdown files with atomic writes.

## Features

- **Atomic writes** - Files are written using a temporary file and rename pattern, ensuring the file is either fully written or not written at all
- **Path validation** - Protects against path traversal attacks
- **Auto-creates directories** - Parent directories are created automatically if they don't exist

## Installation

### Option 1: Using `go run` (no pre-installation required)

Add to your `.mcp.json`:

```json
{
  "mcpServers": {
    "markdown-writer": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "github.com/robertbagge/markdown-writer-mcp/cmd/markdown-writer-mcp@latest"]
    }
  }
}
```

### Option 2: Install binary first

```bash
go install github.com/robertbagge/markdown-writer-mcp/cmd/markdown-writer-mcp@latest
```

Then add to your `.mcp.json`:

```json
{
  "mcpServers": {
    "markdown-writer": {
      "type": "stdio",
      "command": "markdown-writer-mcp",
      "args": []
    }
  }
}
```

## Tools

### write

Write markdown content to a file path with atomic writes.

**Parameters:**
- `path` (string, required) - Absolute or relative path to the markdown file to write
- `content` (string, required) - Markdown content to write to the file

**Returns:**
- `path` - The resolved absolute path where the file was written
- `size` - Number of bytes written

### verify

Verify that a markdown file exists and get its statistics.

**Parameters:**
- `path` (string, required) - Absolute or relative path to the markdown file to verify

**Returns:**
- `path` - The resolved absolute path of the file
- `size` - File size in bytes
- `lines` - Number of lines in the file

## Development

```bash
# Run in development mode
task dev

# Build binary
task build

# Run tests
task test

# Tidy modules
task tidy
```

## License

MIT
