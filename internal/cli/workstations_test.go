package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func workstationsSubgroup(name string) *cobra.Command {
	grp := newWorkstationsCmds(legacyTestInvocation())
	for _, c := range grp.workstationsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestWorkstationsClustersSubcommands(t *testing.T) {
	g := workstationsSubgroup("clusters")
	if g == nil {
		t.Fatal("workstations clusters missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestWorkstationsConfigsSubcommands(t *testing.T) {
	g := workstationsSubgroup("configs")
	if g == nil {
		t.Fatal("workstations configs missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy",
	})
}

func TestWorkstationsEndpointMode(t *testing.T) {
	for _, tc := range []struct {
		name, flag, property, location, want, wantErr string
	}{
		{name: "unset", location: "us-central1"},
		{name: "global", flag: "global", location: "us-central1"},
		{name: "regional", flag: "regional", location: "europe-west1", want: "https://workstations.europe-west1.rep.googleapis.com/"},
		{name: "regional-preferred known", flag: "regional-preferred", location: "us-east4", want: "https://workstations.us-east4.rep.googleapis.com/"},
		{name: "regional-preferred unknown", flag: "regional-preferred", location: "europe-west1"},
		{name: "property", property: "regional", location: "us-west1", want: "https://workstations.us-west1.rep.googleapis.com/"},
		{name: "flag overrides property", flag: "global", property: "regional", location: "us-west1"},
		{name: "invalid", flag: "nearby", location: "us-west1", wantErr: "argument --endpoint-mode: Invalid choice: 'nearby'.\n\nValid choices are [global, regional, regional-preferred]."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vars := map[string]string{}
			if tc.property != "" {
				vars["CLOUDSDK_REGIONAL_ENDPOINT_MODE"] = tc.property
			}
			inv, _, _ := newTestInvocation(t, vars, "")
			s := &workstationsCmds{inv: inv, flagWsEndpointMode: tc.flag}
			got, err := s.wsEndpoint(tc.location)
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

func TestWorkstationsEndpointModeFlag(t *testing.T) {
	cmds := newWorkstationsCmds(legacyTestInvocation())
	for _, c := range []*cobra.Command{
		cmds.workstationsClCreateCmd, cmds.workstationsClDeleteCmd, cmds.workstationsClDescribeCmd,
		cmds.workstationsClListCmd, cmds.workstationsClUpdateCmd,
		cmds.workstationsCfgCreateCmd, cmds.workstationsCfgDeleteCmd, cmds.workstationsCfgDescribeCmd,
		cmds.workstationsCfgListCmd, cmds.workstationsCfgUpdateCmd, cmds.workstationsCfgGetIamCmd,
		cmds.workstationsCfgSetIamCmd,
	} {
		if c.Flags().Lookup("endpoint-mode") == nil {
			t.Errorf("%s has no --endpoint-mode", c.CommandPath())
		}
	}
}
