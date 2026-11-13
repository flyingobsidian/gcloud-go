package cli

import "github.com/spf13/cobra"

// --- gcloud apphub (#300) ---

func (g *apphubCmds) initApphub() {
	// All subgroups (applications with nested services + workloads, boundary,
	// discovered-services, discovered-workloads, locations, operations,
	// service-projects) are implemented in apphub_all.go.

}

// apphubCmds holds the commands and flag values of one invocation's "apphub"
// command group, so invocations don't share state (#2070).
type apphubCmds struct {
	inv *invocation

	apphubCmd                    *cobra.Command
	flagAHLocation               string
	flagAHConfigFile             string
	flagAHUpdateMask             string
	flagAHFormat                 string
	flagAHAsync                  bool
	apphubLocationsCmd           *cobra.Command
	ahLocDescribeCmd             *cobra.Command
	ahLocListCmd                 *cobra.Command
	apphubOperationsCmd          *cobra.Command
	ahOpCancelCmd                *cobra.Command
	ahOpDeleteCmd                *cobra.Command
	ahOpDescribeCmd              *cobra.Command
	ahOpListCmd                  *cobra.Command
	apphubBoundaryCmd            *cobra.Command
	ahBoundaryDescribeCmd        *cobra.Command
	ahBoundaryUpdateCmd          *cobra.Command
	apphubApplicationsCmd        *cobra.Command
	ahAppCreateCmd               *cobra.Command
	ahAppDeleteCmd               *cobra.Command
	ahAppDescribeCmd             *cobra.Command
	ahAppListCmd                 *cobra.Command
	ahAppUpdateCmd               *cobra.Command
	apphubAppServicesCmd         *cobra.Command
	flagAHAppID                  string
	ahAppSvcCreateCmd            *cobra.Command
	ahAppSvcDeleteCmd            *cobra.Command
	ahAppSvcDescribeCmd          *cobra.Command
	ahAppSvcListCmd              *cobra.Command
	ahAppSvcUpdateCmd            *cobra.Command
	apphubAppWorkloadsCmd        *cobra.Command
	ahAppWLCreateCmd             *cobra.Command
	ahAppWLDeleteCmd             *cobra.Command
	ahAppWLDescribeCmd           *cobra.Command
	ahAppWLListCmd               *cobra.Command
	ahAppWLUpdateCmd             *cobra.Command
	apphubDiscoveredServicesCmd  *cobra.Command
	ahDSDescribeCmd              *cobra.Command
	ahDSListCmd                  *cobra.Command
	ahDSLookupCmd                *cobra.Command
	flagAHDSURI                  string
	apphubDiscoveredWorkloadsCmd *cobra.Command
	ahDWDescribeCmd              *cobra.Command
	ahDWListCmd                  *cobra.Command
	ahDWLookupCmd                *cobra.Command
	flagAHDWURI                  string
	apphubServiceProjectsCmd     *cobra.Command
	ahSPAddCmd                   *cobra.Command
	ahSPRemoveCmd                *cobra.Command
	ahSPDescribeCmd              *cobra.Command
	ahSPListCmd                  *cobra.Command
	ahSPLookupCmd                *cobra.Command
}

// newApphubCmd returns the "apphub" command group for inv.
func newApphubCmd(inv *invocation) *cobra.Command {
	return newApphubCmds(inv).apphubCmd
}

// newApphubCmds builds the "apphub" command group for inv and returns its state.
func newApphubCmds(inv *invocation) *apphubCmds {
	g := &apphubCmds{inv: inv}
	g.apphubCmd = &cobra.Command{Use: "apphub", Short: "Manage App Hub"}
	g.apphubLocationsCmd = &cobra.Command{Use: "locations", Short: "Explore App Hub locations"}
	g.ahLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe an App Hub location",
		Args: cobra.ExactArgs(1), RunE: g.runAHLocDescribe,
	}
	g.ahLocListCmd = &cobra.Command{
		Use: "list", Short: "List App Hub locations",
		Args: cobra.NoArgs, RunE: g.runAHLocList,
	}
	g.apphubOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage App Hub operations"}
	g.ahOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an App Hub operation",
		Args: cobra.ExactArgs(1), RunE: g.runAHOpCancel,
	}
	g.ahOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an App Hub operation",
		Args: cobra.ExactArgs(1), RunE: g.runAHOpDelete,
	}
	g.ahOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an App Hub operation",
		Args: cobra.ExactArgs(1), RunE: g.runAHOpDescribe,
	}
	g.ahOpListCmd = &cobra.Command{
		Use: "list", Short: "List App Hub operations in a location",
		Args: cobra.NoArgs, RunE: g.runAHOpList,
	}
	g.apphubBoundaryCmd = &cobra.Command{Use: "boundary", Short: "Manage App Hub location boundaries"}
	g.ahBoundaryDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the boundary for a location",
		Args: cobra.NoArgs, RunE: g.runAHBoundaryDescribe,
	}
	g.ahBoundaryUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update the boundary from a --config-file",
		Args: cobra.NoArgs, RunE: g.runAHBoundaryUpdate,
	}
	g.apphubApplicationsCmd = &cobra.Command{Use: "applications", Short: "Manage App Hub applications"}
	g.ahAppCreateCmd = &cobra.Command{
		Use: "create APPLICATION", Short: "Create an application from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppCreate,
	}
	g.ahAppDeleteCmd = &cobra.Command{
		Use: "delete APPLICATION", Short: "Delete an application",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppDelete,
	}
	g.ahAppDescribeCmd = &cobra.Command{
		Use: "describe APPLICATION", Short: "Describe an application",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppDescribe,
	}
	g.ahAppListCmd = &cobra.Command{
		Use: "list", Short: "List applications in a location",
		Args: cobra.NoArgs, RunE: g.runAHAppList,
	}
	g.ahAppUpdateCmd = &cobra.Command{
		Use: "update APPLICATION", Short: "Update an application from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppUpdate,
	}
	g.apphubAppServicesCmd = &cobra.Command{Use: "services", Short: "Manage application services"}
	g.ahAppSvcCreateCmd = &cobra.Command{
		Use: "create SERVICE", Short: "Create a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppSvcCreate,
	}
	g.ahAppSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a service",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppSvcDelete,
	}
	g.ahAppSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppSvcDescribe,
	}
	g.ahAppSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services in an application",
		Args: cobra.NoArgs, RunE: g.runAHAppSvcList,
	}
	g.ahAppSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppSvcUpdate,
	}
	g.apphubAppWorkloadsCmd = &cobra.Command{Use: "workloads", Short: "Manage application workloads"}
	g.ahAppWLCreateCmd = &cobra.Command{
		Use: "create WORKLOAD", Short: "Create a workload from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppWLCreate,
	}
	g.ahAppWLDeleteCmd = &cobra.Command{
		Use: "delete WORKLOAD", Short: "Delete a workload",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppWLDelete,
	}
	g.ahAppWLDescribeCmd = &cobra.Command{
		Use: "describe WORKLOAD", Short: "Describe a workload",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppWLDescribe,
	}
	g.ahAppWLListCmd = &cobra.Command{
		Use: "list", Short: "List workloads in an application",
		Args: cobra.NoArgs, RunE: g.runAHAppWLList,
	}
	g.ahAppWLUpdateCmd = &cobra.Command{
		Use: "update WORKLOAD", Short: "Update a workload from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runAHAppWLUpdate,
	}
	g.apphubDiscoveredServicesCmd = &cobra.Command{Use: "discovered-services", Short: "View discovered App Hub services"}
	g.ahDSDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a discovered service",
		Args: cobra.ExactArgs(1), RunE: g.runAHDSDescribe,
	}
	g.ahDSListCmd = &cobra.Command{
		Use: "list", Short: "List discovered services in a location",
		Args: cobra.NoArgs, RunE: g.runAHDSList,
	}
	g.ahDSLookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up a discovered service by URI",
		Args: cobra.NoArgs, RunE: g.runAHDSLookup,
	}
	g.apphubDiscoveredWorkloadsCmd = &cobra.Command{Use: "discovered-workloads", Short: "View discovered App Hub workloads"}
	g.ahDWDescribeCmd = &cobra.Command{
		Use: "describe WORKLOAD", Short: "Describe a discovered workload",
		Args: cobra.ExactArgs(1), RunE: g.runAHDWDescribe,
	}
	g.ahDWListCmd = &cobra.Command{
		Use: "list", Short: "List discovered workloads in a location",
		Args: cobra.NoArgs, RunE: g.runAHDWList,
	}
	g.ahDWLookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up a discovered workload by URI",
		Args: cobra.NoArgs, RunE: g.runAHDWLookup,
	}
	g.apphubServiceProjectsCmd = &cobra.Command{Use: "service-projects", Short: "Manage App Hub service project attachments"}
	g.ahSPAddCmd = &cobra.Command{
		Use: "add SERVICE_PROJECT", Short: "Attach a service project",
		Args: cobra.ExactArgs(1), RunE: g.runAHSPAdd,
	}
	g.ahSPRemoveCmd = &cobra.Command{
		Use: "remove SERVICE_PROJECT", Short: "Detach a service project",
		Args: cobra.ExactArgs(1), RunE: g.runAHSPRemove,
	}
	g.ahSPDescribeCmd = &cobra.Command{
		Use: "describe SERVICE_PROJECT", Short: "Describe a service project attachment",
		Args: cobra.ExactArgs(1), RunE: g.runAHSPDescribe,
	}
	g.ahSPListCmd = &cobra.Command{
		Use: "list", Short: "List service project attachments in a location",
		Args: cobra.NoArgs, RunE: g.runAHSPList,
	}
	g.ahSPLookupCmd = &cobra.Command{
		Use: "lookup SERVICE_PROJECT", Short: "Look up a service project attachment by project ID",
		Args: cobra.ExactArgs(1), RunE: g.runAHSPLookup,
	}
	g.initApphub()
	g.initApphubAll()
	g.initApphubAll2()
	g.initApphubAll3()
	g.initApphubAll4()
	g.initApphubAll5()
	g.initApphubAll6()
	g.initApphubAll7()
	return g
}
