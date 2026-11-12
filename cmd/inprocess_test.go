package cmd

import "testing"

func TestHasPositional(t *testing.T) {
	inv, _, _ := newTestInvocation(t, nil, "")
	root := newRootCmd(inv, false)
	tests := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"--version"}, false},
		{[]string{"--project", "p"}, false},
		{[]string{"--project=p", "-q"}, false},
		{[]string{"--log-http", "--format", "json"}, false},
		{[]string{"--project", "p", "secrets"}, true},
		{[]string{"secrets", "list"}, true},
		{[]string{"-q", "secrets"}, true},
		{[]string{"-"}, true},
		{[]string{"--"}, false},
		{[]string{"--", "x"}, true},
	}
	for _, tt := range tests {
		if got := hasPositional(root, tt.args); got != tt.want {
			t.Errorf("hasPositional(%q) = %v, want %v", tt.args, got, tt.want)
		}
	}
}
