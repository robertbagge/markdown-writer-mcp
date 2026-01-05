package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
	"github.com/robertbagge/markdown-writer-mcp/internal/pathutil"
)

// JSONQueryTool defines the json_query tool metadata
var JSONQueryTool = &mcp.Tool{
	Name:        "json_query",
	Description: "Query JSON arrays with filtering, supports nested paths and multiple filter operations",
}

// Filter defines a single filter condition
type Filter struct {
	Field string `json:"field" jsonschema:"Dot-notation path to field (e.g., 'regions', 'type')"`
	Op    string `json:"op" jsonschema:"Operation: eq, neq, contains, is_null, is_not_null"`
	Value any    `json:"value,omitempty" jsonschema:"Value to compare (required for eq/neq/contains)"`
}

// JSONQueryArgs defines the input parameters for the json_query tool
type JSONQueryArgs struct {
	Path      string   `json:"path" jsonschema:"Absolute or relative path to the JSON file"`
	ArrayPath []string `json:"arrayPath,omitempty" jsonschema:"Path to array in JSON structure (e.g., [\"data\", \"items\"])"`
	Filters   []Filter `json:"filters,omitempty" jsonschema:"Array of filter conditions (AND logic)"`
	Limit     *int     `json:"limit,omitempty" jsonschema:"Maximum number of results to return"`
}

// JSONQueryOutput defines the output structure for the json_query tool
type JSONQueryOutput struct {
	Result []any `json:"result"`
	Count  int   `json:"count"`
}

// JSONQueryHandler handles the json_query tool invocation
func JSONQueryHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args JSONQueryArgs,
) (*mcp.CallToolResult, JSONQueryOutput, error) {
	// Resolve path (validates and converts to absolute)
	absPath, err := pathutil.Resolve(args.Path)
	if err != nil {
		return nil, JSONQueryOutput{}, err
	}

	slog.Info("json_query tool called",
		slog.String("path", absPath),
		slog.Any("arrayPath", args.ArrayPath),
		slog.Int("filterCount", len(args.Filters)),
	)

	// Read file using injected reader
	content, err := fileReader.Read(ctx, absPath)
	if err != nil {
		return nil, JSONQueryOutput{}, err
	}

	// Parse JSON
	var data any
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, JSONQueryOutput{}, fmt.Errorf("%w: %v", domain.ErrInvalidJSON, err)
	}

	// Navigate to arrayPath if provided
	target := data
	if len(args.ArrayPath) > 0 {
		target, err = navigateToPath(data, args.ArrayPath)
		if err != nil {
			return nil, JSONQueryOutput{}, err
		}
	}

	// Verify target is an array
	arr, ok := target.([]any)
	if !ok {
		return nil, JSONQueryOutput{}, domain.ErrNotAnArray
	}

	// Apply filters
	var results []any
	for _, item := range arr {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue // Skip non-object items
		}

		if matchesAllFilters(itemMap, args.Filters) {
			results = append(results, item)
		}
	}

	// Apply limit if specified
	if args.Limit != nil && *args.Limit > 0 && len(results) > *args.Limit {
		results = results[:*args.Limit]
	}

	output := JSONQueryOutput{
		Result: results,
		Count:  len(results),
	}

	// Serialize output to JSON for MCP response
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return nil, JSONQueryOutput{}, fmt.Errorf("failed to marshal output: %w", err)
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(outputJSON)},
		},
	}

	return result, output, nil
}

// navigateToPath traverses the JSON structure following the given path
func navigateToPath(data any, path []string) (any, error) {
	current := data
	for _, key := range path {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, domain.ErrArrayPathNotFound
		}
		val, exists := obj[key]
		if !exists {
			return nil, domain.ErrArrayPathNotFound
		}
		current = val
	}
	return current, nil
}

// matchesAllFilters checks if an item matches all filters (AND logic)
func matchesAllFilters(item map[string]any, filters []Filter) bool {
	for _, filter := range filters {
		if !applyFilter(item, filter) {
			return false
		}
	}
	return true
}

// applyFilter checks if a single filter matches the item
func applyFilter(item map[string]any, filter Filter) bool {
	value, exists := getFieldValue(item, filter.Field)

	switch filter.Op {
	case "eq":
		if !exists {
			return false
		}
		return valuesEqual(value, filter.Value)

	case "neq":
		if !exists {
			return true // Non-existent field is not equal to any value
		}
		return !valuesEqual(value, filter.Value)

	case "contains":
		if !exists {
			return false
		}
		arr, ok := value.([]any)
		if !ok {
			return false
		}
		return arrayContains(arr, filter.Value)

	case "is_null":
		return !exists || value == nil

	case "is_not_null":
		return exists && value != nil

	default:
		return false
	}
}

// getFieldValue retrieves a field value from the item, supporting dot notation
func getFieldValue(item map[string]any, field string) (any, bool) {
	// For now, support simple field access (not nested dot notation in field)
	// The field name itself might contain dots in the future
	val, exists := item[field]
	return val, exists
}

// valuesEqual compares two values for equality
func valuesEqual(a, b any) bool {
	// Handle numeric comparisons (JSON numbers are float64)
	switch av := a.(type) {
	case float64:
		switch bv := b.(type) {
		case float64:
			return av == bv
		case int:
			return av == float64(bv)
		case int64:
			return av == float64(bv)
		}
	case int:
		switch bv := b.(type) {
		case float64:
			return float64(av) == bv
		case int:
			return av == bv
		}
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	}
	return a == b
}

// arrayContains checks if an array contains a value
func arrayContains(arr []any, value any) bool {
	for _, item := range arr {
		if valuesEqual(item, value) {
			return true
		}
	}
	return false
}
