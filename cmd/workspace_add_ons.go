package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// --- gcloud workspace-add-ons (#399) ---
//
// The Google Workspace Add-ons API (`gsuiteaddons.googleapis.com`, v1) is not
// exposed by google.golang.org/api; subgroups use the shared restClient from
// rest_helpers.go with a per-service endpoint.

var workspaceAddOnsRest = newRESTClient("https://gsuiteaddons.googleapis.com/v1")

func (s *workspaceAddOnsCmds) initWorkspaceAddOns() {
	s.workspaceAddOnsGetAuthCmd.Flags().StringVar(&s.flagWAOGetAuthFormat, "format", "", "Output format")
	s.workspaceAddOnsCmd.AddCommand(s.workspaceAddOnsGetAuthCmd)

}

func (s *workspaceAddOnsCmds) runWAOGetAuthorization(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(workspaceAddOnsRest).do(ctx, http.MethodGet, fmt.Sprintf("/projects/%s/authorization", project), nil, nil, &got); err != nil {
		return fmt.Errorf("getting workspace add-ons authorization: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagWAOGetAuthFormat)
}

// workspaceAddOnsCmds holds the commands and flag values of one invocation's "workspaceAddOns"
// command group, so invocations don't share state (#2070).
type workspaceAddOnsCmds struct {
	inv *invocation

	workspaceAddOnsCmd        *cobra.Command
	flagWAOGetAuthFormat      string
	workspaceAddOnsGetAuthCmd *cobra.Command
	waoDepCmd                 *cobra.Command
	flagWAODepFormat          string
	flagWAODepConfigFile      string
	flagWAODepPageSize        int64
	waoDepCreateCmd           *cobra.Command
	waoDepDeleteCmd           *cobra.Command
	waoDepDescribeCmd         *cobra.Command
	waoDepListCmd             *cobra.Command
	waoDepReplaceCmd          *cobra.Command
	waoDepInstallCmd          *cobra.Command
	waoDepUninstallCmd        *cobra.Command
	waoDepInstallStatusCmd    *cobra.Command
}

// newWorkspaceAddOnsCmd returns the "workspaceAddOns" command group for inv.
func newWorkspaceAddOnsCmd(inv *invocation) *cobra.Command {
	return newWorkspaceAddOnsCmds(inv).workspaceAddOnsCmd
}

// newWorkspaceAddOnsCmds builds the "workspaceAddOns" command group for inv and returns its state.
func newWorkspaceAddOnsCmds(inv *invocation) *workspaceAddOnsCmds {
	s := &workspaceAddOnsCmds{inv: inv}
	s.workspaceAddOnsCmd = &cobra.Command{Use: "workspace-add-ons", Short: "Manage Google Workspace Add-ons"}
	s.workspaceAddOnsGetAuthCmd = &cobra.Command{
		Use: "get-authorization", Short: "Get the authorization information for deployments in a project",
		Args: cobra.NoArgs, RunE: s.runWAOGetAuthorization,
	}
	s.waoDepCmd = &cobra.Command{Use: "deployments", Short: "Manage Google Workspace Add-ons deployments"}
	s.waoDepCreateCmd = &cobra.Command{
		Use: "create DEPLOYMENT", Short: "Create a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: s.runWAODepCreate,
	}
	s.waoDepDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: s.runWAODepDelete,
	}
	s.waoDepDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: s.runWAODepDescribe,
	}
	s.waoDepListCmd = &cobra.Command{
		Use: "list", Short: "List Workspace Add-ons deployments",
		Args: cobra.NoArgs, RunE: s.runWAODepList,
	}
	s.waoDepReplaceCmd = &cobra.Command{
		Use: "replace DEPLOYMENT", Short: "Replace a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: s.runWAODepReplace,
	}
	s.waoDepInstallCmd = &cobra.Command{
		Use: "install DEPLOYMENT", Short: "Install a Workspace Add-ons deployment in developer mode",
		Args: cobra.ExactArgs(1), RunE: s.runWAODepInstall,
	}
	s.waoDepUninstallCmd = &cobra.Command{
		Use: "uninstall DEPLOYMENT", Short: "Uninstall a Workspace Add-ons deployment from developer mode",
		Args: cobra.ExactArgs(1), RunE: s.runWAODepUninstall,
	}
	s.waoDepInstallStatusCmd = &cobra.Command{
		Use: "install-status DEPLOYMENT", Short: "Get the install status of a Workspace Add-ons deployment",
		Args: cobra.ExactArgs(1), RunE: s.runWAODepInstallStatus,
	}
	s.initWorkspaceAddOns()
	s.initWorkspaceAddOnsDeployments()
	return s
}
