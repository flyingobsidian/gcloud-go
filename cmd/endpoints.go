package cmd

import "github.com/spf13/cobra"

// --- gcloud endpoints (#892-#894) ---
//
// The top-level `endpoints` command aggregates the Endpoints/Service
// Management subgroups. All subgroups (configs, operations, services) are
// implemented in endpoints_all.go.

func (g *endpointsCmds) initEndpoints() {

}

// endpointsCmds holds the commands and flag values of one invocation's "endpoints"
// command group, so invocations don't share state (#2070).
type endpointsCmds struct {
	inv *invocation

	endpointsCmd           *cobra.Command
	flagEPFormat           string
	flagEPService          string
	flagEPAsync            bool
	flagEPValidate         bool
	flagEPForce            bool
	flagEPMember           string
	flagEPRole             string
	flagEPPageSize         int64
	endpointsConfigsCmd    *cobra.Command
	endpointsOperationsCmd *cobra.Command
	endpointsServicesCmd   *cobra.Command
	epConfigDescribeCmd    *cobra.Command
	epConfigListCmd        *cobra.Command
	epOpDescribeCmd        *cobra.Command
	epOpListCmd            *cobra.Command
	epOpWaitCmd            *cobra.Command
	epSvcDeleteCmd         *cobra.Command
	epSvcUndeployCmd       *cobra.Command
	epSvcDescribeCmd       *cobra.Command
	epSvcListCmd           *cobra.Command
	epSvcDeployCmd         *cobra.Command
	epSvcGetConfigNameCmd  *cobra.Command
	epSvcEnableCmd         *cobra.Command
	epSvcDisableCmd        *cobra.Command
	epSvcAddIAMCmd         *cobra.Command
	epSvcRemoveIAMCmd      *cobra.Command
	epSvcGetIAMCmd         *cobra.Command
	epSvcSetIAMCmd         *cobra.Command
	epSvcCheckIAMCmd       *cobra.Command
}

// newEndpointsCmd returns the "endpoints" command group for inv.
func newEndpointsCmd(inv *invocation) *cobra.Command {
	return newEndpointsCmds(inv).endpointsCmd
}

// newEndpointsCmds builds the "endpoints" command group for inv and returns its state.
func newEndpointsCmds(inv *invocation) *endpointsCmds {
	g := &endpointsCmds{inv: inv}
	g.endpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Endpoints services"}
	g.endpointsConfigsCmd = &cobra.Command{Use: "configs", Short: "View service configurations"}
	g.endpointsOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Service Management operations"}
	g.endpointsServicesCmd = &cobra.Command{Use: "services", Short: "Manage Endpoints services"}
	g.epConfigDescribeCmd = &cobra.Command{
		Use: "describe CONFIG_ID", Short: "Describe a service configuration",
		Args: cobra.ExactArgs(1), RunE: g.runEPConfigDescribe,
	}
	g.epConfigListCmd = &cobra.Command{
		Use: "list", Short: "List configurations for a service",
		Args: cobra.NoArgs, RunE: g.runEPConfigList,
	}
	g.epOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Service Management operation",
		Args: cobra.ExactArgs(1), RunE: g.runEPOpDescribe,
	}
	g.epOpListCmd = &cobra.Command{
		Use: "list", Short: "List Service Management operations",
		Args: cobra.NoArgs, RunE: g.runEPOpList,
	}
	g.epOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Service Management operation to complete",
		Args: cobra.ExactArgs(1), RunE: g.runEPOpWait,
	}
	g.epSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete an Endpoints service",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcDelete,
	}
	g.epSvcUndeployCmd = &cobra.Command{
		Use: "undeploy SERVICE", Short: "Undeploy an Endpoints service (alias for delete)",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcDelete,
	}
	g.epSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe an Endpoints service",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcDescribe,
	}
	g.epSvcListCmd = &cobra.Command{
		Use: "list", Short: "List Endpoints services owned by the current project",
		Args: cobra.NoArgs, RunE: g.runEPSvcList,
	}
	g.epSvcDeployCmd = &cobra.Command{
		Use: "deploy CONFIG_FILE [CONFIG_FILE ...]", Short: "Deploy a service configuration",
		Args: cobra.MinimumNArgs(1), RunE: g.runEPSvcDeploy,
	}
	g.epSvcGetConfigNameCmd = &cobra.Command{
		Use: "get-config-name SERVICE", Short: "Print the active service configuration ID",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcGetConfigName,
	}
	g.epSvcEnableCmd = &cobra.Command{
		Use: "enable SERVICE", Short: "Enable an Endpoints service for the project",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcEnable,
	}
	g.epSvcDisableCmd = &cobra.Command{
		Use: "disable SERVICE", Short: "Disable an Endpoints service for the project",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcDisable,
	}
	g.epSvcAddIAMCmd = &cobra.Command{
		Use: "add-iam-policy-binding SERVICE", Short: "Add an IAM policy binding to a service",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcAddIAM,
	}
	g.epSvcRemoveIAMCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SERVICE", Short: "Remove an IAM policy binding from a service",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcRemoveIAM,
	}
	g.epSvcGetIAMCmd = &cobra.Command{
		Use: "get-iam-policy SERVICE", Short: "Get the IAM policy for a service",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcGetIAM,
	}
	g.epSvcSetIAMCmd = &cobra.Command{
		Use: "set-iam-policy SERVICE POLICY_FILE", Short: "Set the IAM policy for a service",
		Args: cobra.ExactArgs(2), RunE: g.runEPSvcSetIAM,
	}
	g.epSvcCheckIAMCmd = &cobra.Command{
		Use: "check-iam-policy SERVICE", Short: "Check IAM permissions on a service",
		Args: cobra.ExactArgs(1), RunE: g.runEPSvcCheckIAM,
	}
	g.initEndpoints()
	g.initEndpointsAll()
	return g
}
