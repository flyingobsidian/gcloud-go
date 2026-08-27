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
		name          string
		folder, org   string
		want          string
		wantErrSubstr string
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

// TestGetIamPolicyPipeline exercises the exact --flatten/--filter/--format
// invocation from #1728 against a fake policy value.
func TestGetIamPolicyPipeline(t *testing.T) {
	// Fake policy, matching the CRM v3 Policy JSON shape.
	policy := map[string]any{
		"bindings": []any{
			map[string]any{
				"role":    "roles/some.roleName",
				"members": []any{"user:human@example.com", "serviceAccount:MY_SA@MY_PROJECT.iam.gserviceaccount.com"},
			},
			map[string]any{
				"role":    "roles/other.role",
				"members": []any{"user:human@example.com"},
			},
			map[string]any{
				"role":    "roles/someOther.roleName",
				"members": []any{"serviceAccount:MY_SA@MY_PROJECT.iam.gserviceaccount.com"},
			},
		},
	}

	// Save & restore globals.
	prevFlatten, prevFilter, prevFormat := flagFlatten, flagFilter, flagFormat
	defer func() { flagFlatten, flagFilter, flagFormat = prevFlatten, prevFilter, prevFormat }()

	flagFlatten = []string{"bindings[].members"}
	flagFilter = `bindings.members:serviceAccount:MY_SA@MY_PROJECT.iam.gserviceaccount.com`
	flagFormat = "table(bindings.role)"

	var emitErr error
	out := captureStdout(t, func() { emitErr = emitWithPipeline(policy) })
	if emitErr != nil {
		t.Fatalf("emitWithPipeline: %v", emitErr)
	}

	if !strings.Contains(out, "ROLE") {
		t.Errorf("expected table header 'ROLE' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "roles/some.roleName") {
		t.Errorf("expected roles/some.roleName in output, got:\n%s", out)
	}
	if !strings.Contains(out, "roles/someOther.roleName") {
		t.Errorf("expected roles/someOther.roleName in output, got:\n%s", out)
	}
	if strings.Contains(out, "roles/other.role") {
		t.Errorf("did not expect the role without matching member, got:\n%s", out)
	}
}

// TestGetIamPolicyNoPipeline confirms the fallback path (no --flatten/filter/
// format flags) still produces a yaml dump byte-for-byte identical to
// yamlEncode(policy) — protecting existing users.
func TestGetIamPolicyNoPipeline(t *testing.T) {
	policy := map[string]any{"bindings": []any{map[string]any{"role": "roles/A"}}}

	prevFlatten, prevFilter, prevFormat := flagFlatten, flagFilter, flagFormat
	defer func() { flagFlatten, flagFilter, flagFormat = prevFlatten, prevFilter, prevFormat }()
	flagFlatten, flagFilter, flagFormat = nil, "", ""

	var err1, err2 error
	got := captureStdout(t, func() { err1 = emitWithPipeline(policy) })
	want := captureStdout(t, func() { err2 = yamlEncode(policy) })
	if err1 != nil || err2 != nil {
		t.Fatalf("encode errors: pipeline=%v, yaml=%v", err1, err2)
	}
	if got != want {
		t.Errorf("emitWithPipeline(no flags) output diverged from yamlEncode:\n--got--\n%s\n--want--\n%s", got, want)
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
