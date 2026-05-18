package cmd

import "testing"

func TestStatusToAPIFilter(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"active", "ACTIVE"},
		{"Active", "ACTIVE"},
		{"ACTIVE", "ACTIVE"},
		{"terminated", "TERMINATED"},
		{"all", "ALL"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := statusToAPIFilter(tt.input)
			if got != tt.want {
				t.Errorf("statusToAPIFilter(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
