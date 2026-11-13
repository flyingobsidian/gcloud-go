package cmd

import "github.com/spf13/cobra"

// --- gcloud ids (#347) ---

func (s *idsCmds) initIds() {

}

// idsCmds holds the commands and flag values of one invocation's "ids"
// command group, so invocations don't share state (#2070).
type idsCmds struct {
	inv *invocation

	idsCmd                  *cobra.Command
	idsEndpointsCmd         *cobra.Command
	flagIDSZone             string
	flagIDSNetwork          string
	flagIDSDescription      string
	flagIDSSeverity         string
	flagIDSThreatExceptions []string
	flagIDSTrafficLogs      bool
	flagIDSLabels           []string
	flagIDSFormat           string
	flagIDSFilter           string
	flagIDSPageSize         int64
	flagIDSUpdateMask       string
	idsEndpointsCreateCmd   *cobra.Command
	idsEndpointsDeleteCmd   *cobra.Command
	idsEndpointsDescribeCmd *cobra.Command
	idsEndpointsListCmd     *cobra.Command
	idsEndpointsUpdateCmd   *cobra.Command
}

// newIdsCmd returns the "ids" command group for inv.
func newIdsCmd(inv *invocation) *cobra.Command {
	return newIdsCmds(inv).idsCmd
}

// newIdsCmds builds the "ids" command group for inv and returns its state.
func newIdsCmds(inv *invocation) *idsCmds {
	s := &idsCmds{inv: inv}
	s.idsCmd = &cobra.Command{Use: "ids", Short: "Manage Cloud IDS"}
	s.idsEndpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Cloud IDS endpoints"}
	s.idsEndpointsCreateCmd = &cobra.Command{
		Use: "create ENDPOINT", Short: "Create a Cloud IDS endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runIDSCreate,
	}
	s.idsEndpointsDeleteCmd = &cobra.Command{
		Use: "delete ENDPOINT", Short: "Delete a Cloud IDS endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runIDSDelete,
	}
	s.idsEndpointsDescribeCmd = &cobra.Command{
		Use: "describe ENDPOINT", Short: "Describe a Cloud IDS endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runIDSDescribe,
	}
	s.idsEndpointsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud IDS endpoints",
		Args: cobra.NoArgs, RunE: s.runIDSList,
	}
	s.idsEndpointsUpdateCmd = &cobra.Command{
		Use: "update ENDPOINT", Short: "Update a Cloud IDS endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runIDSUpdate,
	}
	s.initIds()
	s.initIdsEndpoints()
	return s
}
