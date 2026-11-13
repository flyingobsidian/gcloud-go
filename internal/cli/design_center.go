package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud design-center (#329) ---
//
// The Design Center API (`designcenter.googleapis.com`, v1alpha) is not
// exposed by google.golang.org/api. Uses the shared restClient from
// rest_helpers.go with a per-service endpoint.

func designCenterRest() *restClient {
	return newRESTClient("https://designcenter.googleapis.com/v1alpha")
}

func (s *designCenterCmds) initDesignCenter() {

}

func dcLocationName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func dcJoin(parts ...string) string {
	out := parts[0]
	for _, p := range parts[1:] {
		out = out + "/" + strings.TrimPrefix(p, "/")
	}
	return out
}

// designCenterCmds holds the commands and flag values of one invocation's "designCenter"
// command group, so invocations don't share state (#2070).
type designCenterCmds struct {
	inv *invocation

	designCenterCmd       *cobra.Command
	dcLocCmd              *cobra.Command
	flagDCLocFormat       string
	flagDCLocFilter       string
	flagDCLocPageSize     int64
	dcLocDescribeCmd      *cobra.Command
	dcLocListCmd          *cobra.Command
	dcOpCmd               *cobra.Command
	flagDCOpLocation      string
	flagDCOpFormat        string
	flagDCOpFilter        string
	flagDCOpPageSize      int64
	flagDCOpTimeout       time.Duration
	dcOpCancelCmd         *cobra.Command
	dcOpDeleteCmd         *cobra.Command
	dcOpDescribeCmd       *cobra.Command
	dcOpListCmd           *cobra.Command
	dcOpWaitCmd           *cobra.Command
	dcSpaceCmd            *cobra.Command
	flagDCSpaceLocation   string
	flagDCSpaceFormat     string
	flagDCSpaceConfigFile string
	flagDCSpaceUpdateMask string
	flagDCSpaceFilter     string
	flagDCSpacePageSize   int64
	flagDCSpaceForce      bool
	flagDCSpaceTestPerms  []string
	dcSpaceCreateCmd      *cobra.Command
	dcSpaceDeleteCmd      *cobra.Command
	dcSpaceDescribeCmd    *cobra.Command
	dcSpaceListCmd        *cobra.Command
	dcSpaceUpdateCmd      *cobra.Command
	dcSpaceGetIamCmd      *cobra.Command
	dcSpaceSetIamCmd      *cobra.Command
	dcSpaceTestIamCmd     *cobra.Command
}

// newDesignCenterCmd returns the "designCenter" command group for inv.
func newDesignCenterCmd(inv *invocation) *cobra.Command {
	return newDesignCenterCmds(inv).designCenterCmd
}

// newDesignCenterCmds builds the "designCenter" command group for inv and returns its state.
func newDesignCenterCmds(inv *invocation) *designCenterCmds {
	s := &designCenterCmds{inv: inv}
	s.designCenterCmd = &cobra.Command{Use: "design-center", Short: "Manage Google Cloud Design Center"}
	s.dcLocCmd = &cobra.Command{Use: "locations", Short: "Manage Design Center locations"}
	s.dcLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Design Center location",
		Args: cobra.ExactArgs(1), RunE: s.runDCLocDescribe,
	}
	s.dcLocListCmd = &cobra.Command{
		Use: "list", Short: "List Design Center locations for the current project",
		Args: cobra.NoArgs, RunE: s.runDCLocList,
	}
	s.dcOpCmd = &cobra.Command{Use: "operations", Short: "Manage Design Center operations"}
	s.dcOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Design Center operation",
		Args: cobra.ExactArgs(1), RunE: s.runDCOpCancel,
	}
	s.dcOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Design Center operation record",
		Args: cobra.ExactArgs(1), RunE: s.runDCOpDelete,
	}
	s.dcOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Design Center operation",
		Args: cobra.ExactArgs(1), RunE: s.runDCOpDescribe,
	}
	s.dcOpListCmd = &cobra.Command{
		Use: "list", Short: "List Design Center operations",
		Args: cobra.NoArgs, RunE: s.runDCOpList,
	}
	s.dcOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Design Center operation to finish",
		Args: cobra.ExactArgs(1), RunE: s.runDCOpWait,
	}
	s.dcSpaceCmd = &cobra.Command{Use: "spaces", Short: "Manage Design Center spaces"}
	s.dcSpaceCreateCmd = &cobra.Command{
		Use: "create SPACE", Short: "Create a Design Center space",
		Args: cobra.ExactArgs(1), RunE: s.runDCSpaceCreate,
	}
	s.dcSpaceDeleteCmd = &cobra.Command{
		Use: "delete SPACE", Short: "Delete a Design Center space",
		Args: cobra.ExactArgs(1), RunE: s.runDCSpaceDelete,
	}
	s.dcSpaceDescribeCmd = &cobra.Command{
		Use: "describe SPACE", Short: "Describe a Design Center space",
		Args: cobra.ExactArgs(1), RunE: s.runDCSpaceDescribe,
	}
	s.dcSpaceListCmd = &cobra.Command{
		Use: "list", Short: "List Design Center spaces",
		Args: cobra.NoArgs, RunE: s.runDCSpaceList,
	}
	s.dcSpaceUpdateCmd = &cobra.Command{
		Use: "update SPACE", Short: "Update a Design Center space",
		Args: cobra.ExactArgs(1), RunE: s.runDCSpaceUpdate,
	}
	s.dcSpaceGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SPACE", Short: "Get the IAM policy for a space",
		Args: cobra.ExactArgs(1), RunE: s.runDCSpaceGetIam,
	}
	s.dcSpaceSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SPACE POLICY_FILE", Short: "Set the IAM policy for a space",
		Args: cobra.ExactArgs(2), RunE: s.runDCSpaceSetIam,
	}
	s.dcSpaceTestIamCmd = &cobra.Command{
		Use: "test-iam-permissions SPACE", Short: "Test the caller's permissions on a space",
		Args: cobra.ExactArgs(1), RunE: s.runDCSpaceTestIam,
	}
	s.initDesignCenter()
	s.initDesignCenterLocations()
	s.initDesignCenterOperations()
	s.initDesignCenterSpaces()
	return s
}
