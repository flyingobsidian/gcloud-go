package cli

import (
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud telco-automation (#390) ---
//
// The Telco Automation API (`telcoautomation.googleapis.com`, v1) is not
// exposed by google.golang.org/api; subgroups use the shared restClient from
// rest_helpers.go with a per-service endpoint.

func telcoAutomationRest() *restClient {
	return newRESTClient("https://telcoautomation.googleapis.com/v1")
}

func (s *telcoAutomationCmds) initTelcoAutomation() {

}

// telcoAutomationCmds holds the commands and flag values of one invocation's "telcoAutomation"
// command group, so invocations don't share state (#2070).
type telcoAutomationCmds struct {
	inv *invocation

	telcoAutomationCmd *cobra.Command
	taOpCmd            *cobra.Command
	flagTAOpLocation   string
	flagTAOpFormat     string
	flagTAOpTimeout    time.Duration
	taOpDescribeCmd    *cobra.Command
	taOpWaitCmd        *cobra.Command
	taOCCmd            *cobra.Command
	flagTAOCLocation   string
	flagTAOCFormat     string
	flagTAOCConfigFile string
	flagTAOCPageSize   int64
	taOCCreateCmd      *cobra.Command
	taOCDeleteCmd      *cobra.Command
	taOCDescribeCmd    *cobra.Command
	taOCListCmd        *cobra.Command
}

// newTelcoAutomationCmd returns the "telcoAutomation" command group for inv.
func newTelcoAutomationCmd(inv *invocation) *cobra.Command {
	return newTelcoAutomationCmds(inv).telcoAutomationCmd
}

// newTelcoAutomationCmds builds the "telcoAutomation" command group for inv and returns its state.
func newTelcoAutomationCmds(inv *invocation) *telcoAutomationCmds {
	s := &telcoAutomationCmds{inv: inv}
	s.telcoAutomationCmd = &cobra.Command{Use: "telco-automation", Short: "Manage Telco Automation"}
	s.taOpCmd = &cobra.Command{Use: "operations", Short: "Manage Telco Automation operations"}
	s.taOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Telco Automation operation",
		Args: cobra.ExactArgs(1), RunE: s.runTAOpDescribe,
	}
	s.taOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Telco Automation operation to finish",
		Args: cobra.ExactArgs(1), RunE: s.runTAOpWait,
	}
	s.taOCCmd = &cobra.Command{Use: "orchestration-cluster", Short: "Manage Telco Automation orchestration clusters"}
	s.taOCCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a Telco Automation orchestration cluster",
		Args: cobra.ExactArgs(1), RunE: s.runTAOCCreate,
	}
	s.taOCDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a Telco Automation orchestration cluster",
		Args: cobra.ExactArgs(1), RunE: s.runTAOCDelete,
	}
	s.taOCDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a Telco Automation orchestration cluster",
		Args: cobra.ExactArgs(1), RunE: s.runTAOCDescribe,
	}
	s.taOCListCmd = &cobra.Command{
		Use: "list", Short: "List Telco Automation orchestration clusters",
		Args: cobra.NoArgs, RunE: s.runTAOCList,
	}
	s.initTelcoAutomation()
	s.initTelcoAutomationOperations()
	s.initTelcoAutomationOrchestrationCluster()
	return s
}
