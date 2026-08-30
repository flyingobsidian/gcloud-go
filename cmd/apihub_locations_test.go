package cmd

import (
	"strings"
	"testing"
)

// TestParseMcpToolFlag covers the two --mcp-tools value shapes accepted by
// `apihub locations configure-and-deploy-server`: the default comma-delimited
// form and the alternate-delimiter escape (`^|^…`).
func TestParseMcpToolFlag(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantID  string
		wantDes string
		wantOp  string
		wantErr string
	}{
		{
			name:    "default_delim",
			in:      "tool-id=greet,description=Greet the user,operation=projects/p/locations/us-central1/apis/hello/versions/v1/operations/greet",
			wantID:  "greet",
			wantDes: "Greet the user",
			wantOp:  "projects/p/locations/us-central1/apis/hello/versions/v1/operations/greet",
		},
		{
			name:    "alternate_delim_with_comma_in_value",
			in:      "^|^tool-id=lookup|description=Lookup by id, name, or email|operation=projects/p/locations/us-central1/apis/users/versions/v1/operations/lookup",
			wantID:  "lookup",
			wantDes: "Lookup by id, name, or email",
			wantOp:  "projects/p/locations/us-central1/apis/users/versions/v1/operations/lookup",
		},
		{
			name:    "missing_operation",
			in:      "tool-id=x,description=y",
			wantErr: "operation",
		},
		{
			name:    "unknown_key",
			in:      "tool-id=x,description=y,operation=z,extra=v",
			wantErr: "unrecognised --mcp-tools key",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseMcpToolFlag(tc.in)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q; got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ToolId != tc.wantID {
				t.Errorf("ToolId=%q, want %q", got.ToolId, tc.wantID)
			}
			if got.Description != tc.wantDes {
				t.Errorf("Description=%q, want %q", got.Description, tc.wantDes)
			}
			if got.Operation == nil || got.Operation.Operation != tc.wantOp {
				t.Errorf("Operation=%+v, want %q", got.Operation, tc.wantOp)
			}
		})
	}
}

// TestBuildApihubMcpToolsMutex ensures --mcp-tools and --mcp-tools-from-file
// are mutually exclusive and that one of them is required.
func TestBuildApihubMcpToolsMutex(t *testing.T) {
	t.Cleanup(func() {
		flagApihubCDMcpTools = nil
		flagApihubCDMcpToolsFromFile = ""
	})

	flagApihubCDMcpTools = []string{"tool-id=x,description=y,operation=z"}
	flagApihubCDMcpToolsFromFile = "tools.yaml"
	if _, err := buildApihubMcpTools(); err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("want mutually-exclusive error; got %v", err)
	}

	flagApihubCDMcpTools = nil
	flagApihubCDMcpToolsFromFile = ""
	if _, err := buildApihubMcpTools(); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("want required error; got %v", err)
	}
}
