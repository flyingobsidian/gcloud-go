package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud bms (#310) ---

func (s *bmsCmds) initBms() {

}

// bmsLocationParent returns projects/PROJECT/locations/LOCATION.
func (s *bmsCmds) bmsLocationParent(location string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// bmsResource returns projects/.../COLLECTION/ID.
func (s *bmsCmds) bmsResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.bmsLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// bmsCmds holds the commands and flag values of one invocation's "bms"
// command group, so invocations don't share state (#2070).
type bmsCmds struct {
	inv *invocation

	bmsCmd                  *cobra.Command
	bmsInstancesCmd         *cobra.Command
	flagBmsInstLocation     string
	flagBmsInstFormat       string
	flagBmsInstConfigFile   string
	flagBmsInstUpdateMask   string
	flagBmsInstNewName      string
	flagBmsInstPageSize     int64
	bmsInstDescribeCmd      *cobra.Command
	bmsInstListCmd          *cobra.Command
	bmsInstUpdateCmd        *cobra.Command
	bmsInstResetCmd         *cobra.Command
	bmsInstStartCmd         *cobra.Command
	bmsInstStopCmd          *cobra.Command
	bmsInstEnableSerialCmd  *cobra.Command
	bmsInstDisableSerialCmd *cobra.Command
	bmsInstRenameCmd        *cobra.Command
	bmsNetworksCmd          *cobra.Command
	flagBmsNetLocation      string
	flagBmsNetFormat        string
	flagBmsNetConfigFile    string
	flagBmsNetUpdateMask    string
	flagBmsNetNewName       string
	flagBmsNetPageSize      int64
	bmsNetDescribeCmd       *cobra.Command
	bmsNetListCmd           *cobra.Command
	bmsNetUpdateCmd         *cobra.Command
	bmsNetRenameCmd         *cobra.Command
	bmsNfsSharesCmd         *cobra.Command
	flagBmsNfsLocation      string
	flagBmsNfsFormat        string
	flagBmsNfsConfigFile    string
	flagBmsNfsUpdateMask    string
	flagBmsNfsNewName       string
	flagBmsNfsPageSize      int64
	bmsNfsDescribeCmd       *cobra.Command
	bmsNfsListCmd           *cobra.Command
	bmsNfsUpdateCmd         *cobra.Command
	bmsNfsRenameCmd         *cobra.Command
	bmsOperationsCmd        *cobra.Command
	flagBmsOpLocation       string
	flagBmsOpFormat         string
	bmsOpDescribeCmd        *cobra.Command
	bmsOsImagesCmd          *cobra.Command
	flagBmsOsLocation       string
	flagBmsOsFormat         string
	flagBmsOsPageSize       int64
	bmsOsListCmd            *cobra.Command
	bmsSshKeysCmd           *cobra.Command
	flagBmsSshLocation      string
	flagBmsSshFormat        string
	flagBmsSshConfigFile    string
	flagBmsSshPageSize      int64
	bmsSshCreateCmd         *cobra.Command
	bmsSshDeleteCmd         *cobra.Command
	bmsSshListCmd           *cobra.Command
	bmsVolumesCmd           *cobra.Command
	flagBmsVolLocation      string
	flagBmsVolFormat        string
	flagBmsVolConfigFile    string
	flagBmsVolUpdateMask    string
	flagBmsVolNewName       string
	flagBmsVolSizeGib       int64
	flagBmsVolPageSize      int64
	bmsVolDescribeCmd       *cobra.Command
	bmsVolListCmd           *cobra.Command
	bmsVolUpdateCmd         *cobra.Command
	bmsVolRenameCmd         *cobra.Command
	bmsVolResizeCmd         *cobra.Command
}

// newBmsCmd returns the "bms" command group for inv.
func newBmsCmd(inv *invocation) *cobra.Command {
	return newBmsCmds(inv).bmsCmd
}

// newBmsCmds builds the "bms" command group for inv and returns its state.
func newBmsCmds(inv *invocation) *bmsCmds {
	s := &bmsCmds{inv: inv}
	s.bmsCmd = &cobra.Command{Use: "bms", Short: "Manage Bare Metal Solution"}
	s.bmsInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage bare metal instances"}
	s.bmsInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: s.runBmsInstDescribe,
	}
	s.bmsInstListCmd = &cobra.Command{
		Use: "list", Short: "List bare metal instances in a location",
		Args: cobra.NoArgs, RunE: s.runBmsInstList,
	}
	s.bmsInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: s.runBmsInstUpdate,
	}
	s.bmsInstResetCmd = &cobra.Command{
		Use: "reset INSTANCE", Short: "Reset a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: s.runBmsInstReset,
	}
	s.bmsInstStartCmd = &cobra.Command{
		Use: "start INSTANCE", Short: "Start a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: s.runBmsInstStart,
	}
	s.bmsInstStopCmd = &cobra.Command{
		Use: "stop INSTANCE", Short: "Stop a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: s.runBmsInstStop,
	}
	s.bmsInstEnableSerialCmd = &cobra.Command{
		Use: "enable-interactive-serial-console INSTANCE", Short: "Enable interactive serial console access",
		Args: cobra.ExactArgs(1), RunE: s.runBmsInstEnableSerial,
	}
	s.bmsInstDisableSerialCmd = &cobra.Command{
		Use: "disable-interactive-serial-console INSTANCE", Short: "Disable interactive serial console access",
		Args: cobra.ExactArgs(1), RunE: s.runBmsInstDisableSerial,
	}
	s.bmsInstRenameCmd = &cobra.Command{
		Use: "rename INSTANCE", Short: "Rename a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: s.runBmsInstRename,
	}
	s.bmsNetworksCmd = &cobra.Command{Use: "networks", Short: "Manage bare metal networks"}
	s.bmsNetDescribeCmd = &cobra.Command{
		Use: "describe NETWORK", Short: "Describe a bare metal network",
		Args: cobra.ExactArgs(1), RunE: s.runBmsNetDescribe,
	}
	s.bmsNetListCmd = &cobra.Command{
		Use: "list", Short: "List bare metal networks in a location",
		Args: cobra.NoArgs, RunE: s.runBmsNetList,
	}
	s.bmsNetUpdateCmd = &cobra.Command{
		Use: "update NETWORK", Short: "Update a bare metal network",
		Args: cobra.ExactArgs(1), RunE: s.runBmsNetUpdate,
	}
	s.bmsNetRenameCmd = &cobra.Command{
		Use: "rename NETWORK", Short: "Rename a bare metal network",
		Args: cobra.ExactArgs(1), RunE: s.runBmsNetRename,
	}
	s.bmsNfsSharesCmd = &cobra.Command{Use: "nfs-shares", Short: "Manage bare metal NFS shares"}
	s.bmsNfsDescribeCmd = &cobra.Command{
		Use: "describe NFS_SHARE", Short: "Describe an NFS share",
		Args: cobra.ExactArgs(1), RunE: s.runBmsNfsDescribe,
	}
	s.bmsNfsListCmd = &cobra.Command{
		Use: "list", Short: "List NFS shares in a location",
		Args: cobra.NoArgs, RunE: s.runBmsNfsList,
	}
	s.bmsNfsUpdateCmd = &cobra.Command{
		Use: "update NFS_SHARE", Short: "Update an NFS share",
		Args: cobra.ExactArgs(1), RunE: s.runBmsNfsUpdate,
	}
	s.bmsNfsRenameCmd = &cobra.Command{
		Use: "rename NFS_SHARE", Short: "Rename an NFS share",
		Args: cobra.ExactArgs(1), RunE: s.runBmsNfsRename,
	}
	s.bmsOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage bare metal long-running operations"}
	s.bmsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a bare metal long-running operation",
		Args: cobra.ExactArgs(1), RunE: s.runBmsOpDescribe,
	}
	s.bmsOsImagesCmd = &cobra.Command{Use: "os-images", Short: "Manage bare metal OS images"}
	s.bmsOsListCmd = &cobra.Command{
		Use: "list", Short: "List available bare metal OS images",
		Args: cobra.NoArgs, RunE: s.runBmsOsList,
	}
	s.bmsSshKeysCmd = &cobra.Command{Use: "ssh-keys", Short: "Manage BMS SSH keys"}
	s.bmsSshCreateCmd = &cobra.Command{
		Use: "create SSH_KEY", Short: "Create an SSH key",
		Args: cobra.ExactArgs(1), RunE: s.runBmsSshCreate,
	}
	s.bmsSshDeleteCmd = &cobra.Command{
		Use: "delete SSH_KEY", Short: "Delete an SSH key",
		Args: cobra.ExactArgs(1), RunE: s.runBmsSshDelete,
	}
	s.bmsSshListCmd = &cobra.Command{
		Use: "list", Short: "List SSH keys in a location",
		Args: cobra.NoArgs, RunE: s.runBmsSshList,
	}
	s.bmsVolumesCmd = &cobra.Command{Use: "volumes", Short: "Manage bare metal volumes"}
	s.bmsVolDescribeCmd = &cobra.Command{
		Use: "describe VOLUME", Short: "Describe a bare metal volume",
		Args: cobra.ExactArgs(1), RunE: s.runBmsVolDescribe,
	}
	s.bmsVolListCmd = &cobra.Command{
		Use: "list", Short: "List bare metal volumes in a location",
		Args: cobra.NoArgs, RunE: s.runBmsVolList,
	}
	s.bmsVolUpdateCmd = &cobra.Command{
		Use: "update VOLUME", Short: "Update a bare metal volume",
		Args: cobra.ExactArgs(1), RunE: s.runBmsVolUpdate,
	}
	s.bmsVolRenameCmd = &cobra.Command{
		Use: "rename VOLUME", Short: "Rename a bare metal volume",
		Args: cobra.ExactArgs(1), RunE: s.runBmsVolRename,
	}
	s.bmsVolResizeCmd = &cobra.Command{
		Use: "resize VOLUME", Short: "Resize a bare metal volume",
		Args: cobra.ExactArgs(1), RunE: s.runBmsVolResize,
	}
	s.initBms()
	s.initBmsInstances()
	s.initBmsNetworks()
	s.initBmsNfsShares()
	s.initBmsOperations()
	s.initBmsOsImages()
	s.initBmsSshKeys()
	s.initBmsVolumes()
	return s
}
