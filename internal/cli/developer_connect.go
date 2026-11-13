package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud developer-connect (#330) ---

func (s *developerConnectCmds) initDeveloperConnect() {

}

// devConnLocationParent returns "projects/PROJECT/locations/LOCATION".
func (s *developerConnectCmds) devConnLocationParent(location string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// devConnResourceName joins a parent, collection, and id into a fully qualified
// name. If id is already fully qualified it is returned as-is.
func (s *developerConnectCmds) devConnResourceName(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.devConnLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// developerConnectCmds holds the commands and flag values of one invocation's "developerConnect"
// command group, so invocations don't share state (#2070).
type developerConnectCmds struct {
	inv *invocation

	developerConnectCmd                *cobra.Command
	developerConnectConnectionsCmd     *cobra.Command
	flagDcConnLocation                 string
	flagDcConnFormat                   string
	flagDcConnConfigFile               string
	flagDcConnUpdateMask               string
	flagDcConnPageSize                 int64
	developerConnectConnCreateCmd      *cobra.Command
	developerConnectConnDeleteCmd      *cobra.Command
	developerConnectConnDescribeCmd    *cobra.Command
	developerConnectConnListCmd        *cobra.Command
	developerConnectConnUpdateCmd      *cobra.Command
	developerConnectInsightsConfigsCmd *cobra.Command
	flagDcIcLocation                   string
	flagDcIcFormat                     string
	flagDcIcConfigFile                 string
	flagDcIcPageSize                   int64
	developerConnectIcCreateCmd        *cobra.Command
	developerConnectIcDeleteCmd        *cobra.Command
	developerConnectIcDescribeCmd      *cobra.Command
	developerConnectIcListCmd          *cobra.Command
	developerConnectIcUpdateCmd        *cobra.Command
	developerConnectOperationsCmd      *cobra.Command
	flagDcOpLocation                   string
	flagDcOpFormat                     string
	flagDcOpFilter                     string
	flagDcOpPageSize                   int64
	flagDcOpTimeout                    time.Duration
	developerConnectOpCancelCmd        *cobra.Command
	developerConnectOpDeleteCmd        *cobra.Command
	developerConnectOpDescribeCmd      *cobra.Command
	developerConnectOpListCmd          *cobra.Command
	developerConnectOpWaitCmd          *cobra.Command
}

// newDeveloperConnectCmd returns the "developerConnect" command group for inv.
func newDeveloperConnectCmd(inv *invocation) *cobra.Command {
	return newDeveloperConnectCmds(inv).developerConnectCmd
}

// newDeveloperConnectCmds builds the "developerConnect" command group for inv and returns its state.
func newDeveloperConnectCmds(inv *invocation) *developerConnectCmds {
	s := &developerConnectCmds{inv: inv}
	s.developerConnectCmd = &cobra.Command{Use: "developer-connect", Short: "Manage Developer Connect"}
	s.developerConnectConnectionsCmd = &cobra.Command{Use: "connections", Short: "Manage Developer Connect connections"}
	s.developerConnectConnCreateCmd = &cobra.Command{
		Use: "create CONNECTION", Short: "Create a Developer Connect connection",
		Args: cobra.ExactArgs(1), RunE: s.runDcConnCreate,
	}
	s.developerConnectConnDeleteCmd = &cobra.Command{
		Use: "delete CONNECTION", Short: "Delete a Developer Connect connection",
		Args: cobra.ExactArgs(1), RunE: s.runDcConnDelete,
	}
	s.developerConnectConnDescribeCmd = &cobra.Command{
		Use: "describe CONNECTION", Short: "Describe a Developer Connect connection",
		Args: cobra.ExactArgs(1), RunE: s.runDcConnDescribe,
	}
	s.developerConnectConnListCmd = &cobra.Command{
		Use: "list", Short: "List Developer Connect connections in a location",
		Args: cobra.NoArgs, RunE: s.runDcConnList,
	}
	s.developerConnectConnUpdateCmd = &cobra.Command{
		Use: "update CONNECTION", Short: "Update a Developer Connect connection",
		Args: cobra.ExactArgs(1), RunE: s.runDcConnUpdate,
	}
	s.developerConnectInsightsConfigsCmd = &cobra.Command{Use: "insights-configs", Short: "Manage Developer Connect insights configs"}
	s.developerConnectIcCreateCmd = &cobra.Command{
		Use: "create CONFIG", Short: "Create an insights config",
		Args: cobra.ExactArgs(1), RunE: s.runDcIcCreate,
	}
	s.developerConnectIcDeleteCmd = &cobra.Command{
		Use: "delete CONFIG", Short: "Delete an insights config",
		Args: cobra.ExactArgs(1), RunE: s.runDcIcDelete,
	}
	s.developerConnectIcDescribeCmd = &cobra.Command{
		Use: "describe CONFIG", Short: "Describe an insights config",
		Args: cobra.ExactArgs(1), RunE: s.runDcIcDescribe,
	}
	s.developerConnectIcListCmd = &cobra.Command{
		Use: "list", Short: "List insights configs in a location",
		Args: cobra.NoArgs, RunE: s.runDcIcList,
	}
	s.developerConnectIcUpdateCmd = &cobra.Command{
		Use: "update CONFIG", Short: "Update an insights config",
		Args: cobra.ExactArgs(1), RunE: s.runDcIcUpdate,
	}
	s.developerConnectOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Developer Connect long-running operations"}
	s.developerConnectOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Developer Connect operation",
		Args: cobra.ExactArgs(1), RunE: s.runDcOpCancel,
	}
	s.developerConnectOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Developer Connect operation",
		Args: cobra.ExactArgs(1), RunE: s.runDcOpDelete,
	}
	s.developerConnectOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Developer Connect operation",
		Args: cobra.ExactArgs(1), RunE: s.runDcOpDescribe,
	}
	s.developerConnectOpListCmd = &cobra.Command{
		Use: "list", Short: "List Developer Connect operations in a location",
		Args: cobra.NoArgs, RunE: s.runDcOpList,
	}
	s.developerConnectOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Developer Connect operation to complete",
		Args: cobra.ExactArgs(1), RunE: s.runDcOpWait,
	}
	s.initDeveloperConnect()
	s.initDeveloperConnectConnections()
	s.initDeveloperConnectInsightsConfigs()
	s.initDeveloperConnectOperations()
	return s
}
