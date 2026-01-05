package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/robertbagge/markdown-writer-mcp/internal/domain"
	"github.com/robertbagge/markdown-writer-mcp/internal/reader"
	"github.com/robertbagge/markdown-writer-mcp/internal/tools"
)

// Test data as specified in the plan
const testDataJSON = `[
  {
    "id": "sequoia-capital",
    "name": "Sequoia Capital",
    "type": "vc",
    "regions": ["san-francisco-bay-area"],
    "profileUpdatedAt": "2026-01-05T12:00:00Z"
  },
  {
    "id": "intel-capital",
    "name": "Intel Capital",
    "type": "corporate-vc",
    "regions": ["san-francisco-bay-area", "portland"],
    "profileUpdatedAt": null
  },
  {
    "id": "angel-list-syndicate",
    "name": "AngelList Access Fund",
    "type": "angel-syndicate",
    "regions": ["new-york-city"],
    "profileUpdatedAt": null
  },
  {
    "id": "a16z",
    "name": "Andreessen Horowitz",
    "type": "vc",
    "regions": ["san-francisco-bay-area", "new-york-city"],
    "profileUpdatedAt": "2026-01-04T10:00:00Z"
  }
]`

const nestedDataJSON = `{
  "metadata": { "version": "1.0" },
  "data": {
    "investors": [
      { "id": "a", "type": "vc" },
      { "id": "b", "type": "angel-syndicate" }
    ]
  }
}`

func intPtr(i int) *int {
	return &i
}

func TestJSONQueryHandler(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		args       tools.JSONQueryArgs
		wantErr    error
		wantCount  int
		wantIDs    []string // Expected IDs in results (for verification)
	}{
		// Test Case 1: Filter by region (contains)
		{
			name:  "filter by region contains",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path: "/tmp/test.json",
				Filters: []tools.Filter{
					{Field: "regions", Op: "contains", Value: "san-francisco-bay-area"},
				},
			},
			wantErr:   nil,
			wantCount: 3,
			wantIDs:   []string{"sequoia-capital", "intel-capital", "a16z"},
		},
		// Test Case 2: Filter by type (eq)
		{
			name:  "filter by type eq",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path: "/tmp/test.json",
				Filters: []tools.Filter{
					{Field: "type", Op: "eq", Value: "vc"},
				},
			},
			wantErr:   nil,
			wantCount: 2,
			wantIDs:   []string{"sequoia-capital", "a16z"},
		},
		// Test Case 3: Filter unprofiled (is_null)
		{
			name:  "filter unprofiled is_null",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path: "/tmp/test.json",
				Filters: []tools.Filter{
					{Field: "profileUpdatedAt", Op: "is_null"},
				},
			},
			wantErr:   nil,
			wantCount: 2,
			wantIDs:   []string{"intel-capital", "angel-list-syndicate"},
		},
		// Test Case 4: Filter profiled (is_not_null)
		{
			name:  "filter profiled is_not_null",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path: "/tmp/test.json",
				Filters: []tools.Filter{
					{Field: "profileUpdatedAt", Op: "is_not_null"},
				},
			},
			wantErr:   nil,
			wantCount: 2,
			wantIDs:   []string{"sequoia-capital", "a16z"},
		},
		// Test Case 5: Multiple filters (AND logic) - should return 0
		{
			name:  "multiple filters AND logic returns empty",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path: "/tmp/test.json",
				Filters: []tools.Filter{
					{Field: "regions", Op: "contains", Value: "san-francisco-bay-area"},
					{Field: "type", Op: "eq", Value: "vc"},
					{Field: "profileUpdatedAt", Op: "is_null"},
				},
			},
			wantErr:   nil,
			wantCount: 0,
			wantIDs:   []string{},
		},
		// Test Case 6: Unprofiled corporate-vc in SF
		{
			name:  "unprofiled corporate-vc in SF",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path: "/tmp/test.json",
				Filters: []tools.Filter{
					{Field: "regions", Op: "contains", Value: "san-francisco-bay-area"},
					{Field: "type", Op: "eq", Value: "corporate-vc"},
					{Field: "profileUpdatedAt", Op: "is_null"},
				},
			},
			wantErr:   nil,
			wantCount: 1,
			wantIDs:   []string{"intel-capital"},
		},
		// Test Case 7: neq operator
		{
			name:  "neq operator",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path: "/tmp/test.json",
				Filters: []tools.Filter{
					{Field: "type", Op: "neq", Value: "vc"},
				},
			},
			wantErr:   nil,
			wantCount: 2,
			wantIDs:   []string{"intel-capital", "angel-list-syndicate"},
		},
		// Test Case 8: No filters (return all)
		{
			name:  "no filters return all",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path:    "/tmp/test.json",
				Filters: []tools.Filter{},
			},
			wantErr:   nil,
			wantCount: 4,
			wantIDs:   []string{"sequoia-capital", "intel-capital", "angel-list-syndicate", "a16z"},
		},
		// Test Case 9: With limit
		{
			name:  "with limit",
			files: map[string]string{"/tmp/test.json": testDataJSON},
			args: tools.JSONQueryArgs{
				Path:    "/tmp/test.json",
				Filters: []tools.Filter{},
				Limit:   intPtr(2),
			},
			wantErr:   nil,
			wantCount: 2,
			wantIDs:   []string{"sequoia-capital", "intel-capital"},
		},
		// Test Case 10: Invalid path (error handling)
		{
			name:  "invalid path file not found",
			files: map[string]string{},
			args: tools.JSONQueryArgs{
				Path:    "/tmp/nonexistent.json",
				Filters: []tools.Filter{},
			},
			wantErr:   domain.ErrFileNotFound,
			wantCount: 0,
			wantIDs:   []string{},
		},
		// Test Case 11: arrayPath for nested data
		{
			name:  "arrayPath for nested data",
			files: map[string]string{"/tmp/nested.json": nestedDataJSON},
			args: tools.JSONQueryArgs{
				Path:      "/tmp/nested.json",
				ArrayPath: []string{"data", "investors"},
				Filters: []tools.Filter{
					{Field: "type", Op: "eq", Value: "vc"},
				},
			},
			wantErr:   nil,
			wantCount: 1,
			wantIDs:   []string{"a"},
		},
		// Test Case 12: Invalid arrayPath (error handling)
		{
			name:  "invalid arrayPath",
			files: map[string]string{"/tmp/nested.json": nestedDataJSON},
			args: tools.JSONQueryArgs{
				Path:      "/tmp/nested.json",
				ArrayPath: []string{"data", "nonexistent"},
				Filters:   []tools.Filter{},
			},
			wantErr:   domain.ErrArrayPathNotFound,
			wantCount: 0,
			wantIDs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup in-memory reader with test files
			memReader := reader.NewInMemoryFileReader()
			memReader.Files = tt.files
			tools.SetFileReader(memReader)

			result, output, err := tools.JSONQueryHandler(
				context.Background(),
				&mcp.CallToolRequest{},
				tt.args,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("JSONQueryHandler() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("JSONQueryHandler() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Error("JSONQueryHandler() result is nil")
				return
			}

			// Verify count
			if output.Count != tt.wantCount {
				t.Errorf("JSONQueryHandler() count = %v, want %v", output.Count, tt.wantCount)
			}

			// Verify result length matches count
			if len(output.Result) != tt.wantCount {
				t.Errorf("JSONQueryHandler() result length = %v, want %v", len(output.Result), tt.wantCount)
			}

			// Verify IDs if specified
			if len(tt.wantIDs) > 0 {
				gotIDs := make([]string, 0, len(output.Result))
				for _, item := range output.Result {
					if m, ok := item.(map[string]any); ok {
						if id, ok := m["id"].(string); ok {
							gotIDs = append(gotIDs, id)
						}
					}
				}

				if len(gotIDs) != len(tt.wantIDs) {
					t.Errorf("JSONQueryHandler() got IDs = %v, want %v", gotIDs, tt.wantIDs)
				} else {
					for i, wantID := range tt.wantIDs {
						if gotIDs[i] != wantID {
							t.Errorf("JSONQueryHandler() got IDs[%d] = %v, want %v", i, gotIDs[i], wantID)
						}
					}
				}
			}

			// Verify result has text content
			if len(result.Content) == 0 {
				t.Error("JSONQueryHandler() result has no content")
				return
			}

			_, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Error("JSONQueryHandler() result content is not TextContent")
			}
		})
	}
}
