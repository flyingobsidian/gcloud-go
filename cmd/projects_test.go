package cmd

import (
	"strings"
	"testing"
)

func TestProjectResourceName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"my-project", "projects/my-project"},
		{"projects/my-project", "projects/my-project"},
		{"1234567890", "projects/1234567890"},
	}
	for _, c := range cases {
		if got := projectResourceName(c.in); got != c.want {
			t.Errorf("projectResourceName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestProjectParent(t *testing.T) {
	cases := []struct {
		name           string
		folder, org    string
		want           string
		wantErrSubstr  string
	}{
		{name: "neither", want: ""},
		{name: "folder only, bare id", folder: "123", want: "folders/123"},
		{name: "folder only, prefixed", folder: "folders/123", want: "folders/123"},
		{name: "org only, bare id", org: "456", want: "organizations/456"},
		{name: "org only, prefixed", org: "organizations/456", want: "organizations/456"},
		{name: "both set", folder: "1", org: "2", wantErrSubstr: "only one of"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := projectParent(c.folder, c.org)
			if c.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", c.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), c.wantErrSubstr) {
					t.Errorf("error = %q, want substring %q", err.Error(), c.wantErrSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestProjectsCommandRegistered(t *testing.T) {
	got := map[string]bool{}
	for _, c := range projectsCmd.Commands() {
		got[c.Name()] = true
	}
	want := []string{
		"add-iam-policy-binding",
		"create",
		"delete",
		"describe",
		"get-ancestors",
		"get-ancestors-iam-policy",
		"get-iam-policy",
		"list",
		"remove-iam-policy-binding",
		"set-iam-policy",
		"undelete",
		"update",
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("projects subcommand %q not registered", name)
		}
	}
}
