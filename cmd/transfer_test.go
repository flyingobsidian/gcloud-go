package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func transferSubgroup(name string) *cobra.Command {
	for _, c := range transferCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestTransferAgentPoolsSubcommands(t *testing.T) {
	g := transferSubgroup("agent-pools")
	if g == nil {
		t.Fatal("transfer agent-pools missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestTransferAgentsSubcommands(t *testing.T) {
	g := transferSubgroup("agents")
	if g == nil {
		t.Fatal("transfer agents missing")
	}
	assertSubcommands(t, g, []string{"install", "delete"})
}

func TestTransferJobsSubcommands(t *testing.T) {
	g := transferSubgroup("jobs")
	if g == nil {
		t.Fatal("transfer jobs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update", "run", "monitor"})
}

func TestTransferOperationsSubcommands(t *testing.T) {
	g := transferSubgroup("operations")
	if g == nil {
		t.Fatal("transfer operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list", "pause", "resume"})
}

func TestTransferAgentsInstallS3CompatibleModeDeprecated(t *testing.T) {
	g := transferSubgroup("agents")
	if g == nil {
		t.Fatal("transfer agents missing")
	}
	var install *cobra.Command
	for _, c := range g.Commands() {
		if c.Name() == "install" {
			install = c
			break
		}
	}
	if install == nil {
		t.Fatal("transfer agents install missing")
	}
	flag := install.Flags().Lookup("s3-compatible-mode")
	if flag == nil {
		t.Fatal("--s3-compatible-mode missing on transfer agents install")
	}
	if flag.Deprecated == "" {
		t.Error("--s3-compatible-mode should be marked deprecated")
	}
}

func TestTransferAuthorizeExists(t *testing.T) {
	if transferSubgroup("authorize") == nil {
		t.Fatal("transfer authorize missing")
	}
}

func TestTransferEndpointParsing(t *testing.T) {
	cases := []struct {
		url    string
		bucket string
		prefix string
	}{
		{"gs://bkt", "bkt", ""},
		{"gs://bkt/", "bkt", ""},
		{"gs://bkt/foo/bar", "bkt", "foo/bar/"},
	}
	for _, tc := range cases {
		spec := &struct{}{}
		_ = spec
		b, p := splitBucketPrefix(tc.url[len("gs://"):])
		if b != tc.bucket || p != tc.prefix {
			t.Errorf("splitBucketPrefix(%q) = %q,%q want %q,%q", tc.url, b, p, tc.bucket, tc.prefix)
		}
	}
}

func TestTransferIamMemberFor(t *testing.T) {
	cases := []struct {
		email string
		isSA  bool
		want  string
	}{
		{"alice@example.com", false, "user:alice@example.com"},
		{"sa@my-project.iam.gserviceaccount.com", false, "serviceAccount:sa@my-project.iam.gserviceaccount.com"},
		{"sa@my-project.iam.gserviceaccount.com", true, "serviceAccount:sa@my-project.iam.gserviceaccount.com"},
		{"user:alice@example.com", false, "user:alice@example.com"},
	}
	for _, tc := range cases {
		got := xferIamMemberFor(tc.email, tc.isSA)
		if got != tc.want {
			t.Errorf("xferIamMemberFor(%q,%v) = %q want %q", tc.email, tc.isSA, got, tc.want)
		}
	}
}
