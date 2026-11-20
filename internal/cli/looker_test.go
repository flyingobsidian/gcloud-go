package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func lookerSubgroup(name string) *cobra.Command {
	grp := newLookerCmds(legacyTestInvocation())
	for _, c := range grp.lookerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestLookerRegionsSubcommands(t *testing.T) {
	g := lookerSubgroup("regions")
	if g == nil {
		t.Fatal("looker regions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestLookerOperationsSubcommands(t *testing.T) {
	g := lookerSubgroup("operations")
	if g == nil {
		t.Fatal("looker operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list"})
}

func TestLookerInstancesSubcommands(t *testing.T) {
	g := lookerSubgroup("instances")
	if g == nil {
		t.Fatal("looker instances missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "export", "import", "list", "restart", "restore", "update",
	})
}

func TestLookerBackupsSubcommands(t *testing.T) {
	g := lookerSubgroup("backups")
	if g == nil {
		t.Fatal("looker backups missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestLookerEndpointMode(t *testing.T) {
	for _, tc := range []struct {
		name, flag, property, region, want, wantErr string
	}{
		{name: "unset", region: "us-central1"},
		{name: "global", flag: "global", region: "us-central1"},
		{name: "regional", flag: "regional", region: "europe-west2", want: "https://looker.europe-west2.rep.googleapis.com/"},
		{name: "regional-preferred known", flag: "regional-preferred", region: "us-east7", want: "https://looker.us-east7.rep.googleapis.com/"},
		{name: "regional-preferred unknown", flag: "regional-preferred", region: "europe-west2"},
		{name: "property", property: "regional", region: "us-west2", want: "https://looker.us-west2.rep.googleapis.com/"},
		{name: "flag overrides property", flag: "global", property: "regional", region: "us-west2"},
		{name: "invalid", flag: "foo", region: "us-west2", wantErr: "argument --endpoint-mode: Invalid choice: 'foo'.\n\nValid choices are [global, regional, regional-preferred]."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vars := map[string]string{}
			if tc.property != "" {
				vars["CLOUDSDK_REGIONAL_ENDPOINT_MODE"] = tc.property
			}
			inv, _, _ := newTestInvocation(t, vars, "")
			s := &lookerCmds{inv: inv, flagLKEndpointMode: tc.flag, flagLKRegion: tc.region}
			got, err := s.lkEndpoint()
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("err = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("endpoint = %q, want %q", got, tc.want)
			}
		})
	}
}

// Every looker command gcloud-python 589.0.0 marks as supporting regional
// endpoints has --endpoint-mode; regions list, which doesn't, has none.
func TestLookerEndpointModeFlag(t *testing.T) {
	for _, group := range []string{"backups", "instances", "operations", "regions"} {
		for _, c := range lookerSubgroup(group).Commands() {
			has := c.Flags().Lookup("endpoint-mode") != nil
			if want := group != "regions"; has != want {
				t.Errorf("looker %s %s: --endpoint-mode present = %v, want %v", group, c.Name(), has, want)
			}
		}
	}
}

func TestLookerRegionsListRejectsRegionalProperty(t *testing.T) {
	inv, _, _ := newTestInvocation(t, map[string]string{"CLOUDSDK_REGIONAL_ENDPOINT_MODE": "regional"}, "")
	err := runInvocation(inv, "looker", "regions", "list", "--project", "p")
	want := "The regional/endpoint_mode property is currently set to [regional], but this command does not support regional endpoints."
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}
