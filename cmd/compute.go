package cmd

import (
	"github.com/spf13/cobra"
)

func (s *computeCmds) initCompute() {
	s.instanceGroupsCmd.AddCommand(s.unmanagedCmd)
	s.computeCmd.AddCommand(s.instancesCmd)
	s.computeCmd.AddCommand(s.instanceGroupsCmd)
	s.computeCmd.AddCommand(s.disksCmd)
	// Preserve the previous stub-group behaviour for `describe` so existing
	// users don't see a regression on that subcommand path — it still returns
	// "not yet implemented" but at least appears in --help. `list` is in
	// compute_lists.go.
	registerStubCommand(s.disksCmd, "describe", "Not yet implemented")

	// Stub registrations for compute subgroups present in gcloud-python but
	// not yet implemented in gcloud-go (#539). Each is a placeholder so
	// `--help` lists the surface; per-subgroup PRs replace these with real
	// behavior. Grouped in registration to keep the diff scannable.
	for name, short := range map[string]string{
		"accelerator-types":       "GPU / accelerator types",
		"advice":                  "Advisor recommendations",
		"backend-buckets":         "Manage backend buckets",
		"commitments":             "Manage committed use discounts",
		"composite-health-checks": "Manage composite health checks",
		"connect-to-serial-port":  "Connect to a VM's serial port",
		"copy-files":              "Copy files to/from a VM (deprecated in favor of scp)",
		"diagnose":                "Diagnose instance issues",
		// "disks" is a real group registered separately (see disksCmd above).
		"disk-types":                     "List available disk types",
		"external-vpn-gateways":          "Manage external VPN gateways",
		"firewall-policies":              "Manage hierarchical firewall policies",
		"future-reservations":            "Manage future capacity reservations",
		"global-vm-extension-policies":   "Manage global VM extension policies",
		"health-aggregation-policies":    "Manage health aggregation policies",
		"health-sources":                 "Manage health sources",
		"http-health-checks":             "Manage legacy HTTP health checks",
		"https-health-checks":            "Manage legacy HTTPS health checks",
		"instant-snapshot-groups":        "Manage instant snapshot groups",
		"instant-snapshots":              "Manage instant snapshots",
		"interconnects":                  "Manage Cloud Interconnect",
		"machine-images":                 "Manage machine images",
		"machine-types":                  "List machine types",
		"migration":                      "Migrate to Google Cloud",
		"network-attachments":            "Manage network attachments",
		"network-edge-security-services": "Manage network edge security services",
		"network-endpoint-groups":        "Manage network endpoint groups",
		"network-firewall-policies":      "Manage network firewall policies",
		"network-profiles":               "Manage network profiles",
		"operations":                     "Manage long-running operations",
		"org-security-policies":          "Manage organization security policies",
		"os-config":                      "Manage OS Config resources",
		"os-login":                       "Manage OS Login",
		"packet-mirrorings":              "Manage packet mirroring",
		"preview-features":               "Manage preview features",
		"project-zonal-metadata":         "Manage project zonal metadata",
		"public-advertised-prefixes":     "Manage public advertised prefixes",
		"public-delegated-prefixes":      "Manage public delegated prefixes",
		"regions":                        "List Compute Engine regions",
		"reservations":                   "Manage capacity reservations",
		"resource-policies":              "Manage resource policies",
		"rollout-plans":                  "Manage rollout plans",
		"rollouts":                       "Manage rollouts",
		"routes":                         "Manage VPC routes",
		"security-policies":              "Manage Cloud Armor security policies",
		"service-attachments":            "Manage service attachments",
		"shared-vpc":                     "Manage Shared VPC",
		"snapshot-settings":              "Manage snapshot settings",
		"sole-tenancy":                   "Manage sole-tenant nodes",
		"ssl-policies":                   "Manage SSL policies",
		"storage-pools":                  "Manage storage pools",
		"storage-pool-types":             "List storage pool types",
		"target-grpc-proxies":            "Manage target gRPC proxies",
		"target-http-proxies":            "Manage target HTTP proxies",
		"target-https-proxies":           "Manage target HTTPS proxies",
		"target-instances":               "Manage target instances",
		"target-pools":                   "Manage target pools",
		"target-ssl-proxies":             "Manage target SSL proxies",
		"target-tcp-proxies":             "Manage target TCP proxies",
		"target-vpn-gateways":            "Manage Classic VPN gateways",
		"tpus":                           "Manage Cloud TPU nodes",
		"vpn-gateways":                   "Manage HA VPN gateways",
		"vpn-tunnels":                    "Manage VPN tunnels",
		"zones":                          "List Compute Engine zones",
		"zone-vm-extension-policies":     "Manage zonal VM extension policies",
	} {
		registerStubGroup(s.computeCmd, name, short, "list", "describe")
	}
	for name, short := range map[string]string{
		"config-ssh":             "Populate ~/.ssh/config with SSH aliases for VMs",
		"reset-windows-password": "Reset a Windows VM's password",
		"sign-url":               "Sign a URL for Cloud CDN backend buckets",
		"start-iap-tunnel":       "Open an IAP TCP tunnel to a VM",
	} {
		registerStubCommand(s.computeCmd, name, short)
	}

}

// computeCmds holds the commands and flag values of one invocation's "compute"
// command group, so invocations don't share state (#2070).
type computeCmds struct {
	inv *invocation

	computeCmd        *cobra.Command
	instancesCmd      *cobra.Command
	instanceGroupsCmd *cobra.Command
	unmanagedCmd      *cobra.Command
	// disksCmd is the real `compute disks` group. Its subcommands are attached
	// from their own files (see disks_resize.go for #1727 and compute_lists.go
	// for list). "disks" is left out of the stub-registration loop below so it
	// doesn't shadow this real group.
	disksCmd                        *cobra.Command
	flagImagesListStandard          bool
	flagImagesListNoStandard        bool
	flagImagesListShowDeprecate     bool
	flagImagesListPreview           bool
	flagImagesListURI               bool
	projectInfoCmd                  *cobra.Command
	projectInfoDescribeCmd          *cobra.Command
	projectInfoAddMetadataCmd       *cobra.Command
	projectInfoRemoveMetadataCmd    *cobra.Command
	flagRemoveMetadataKeys          string
	flagRemoveMetadataAll           bool
	flagAddProjectMetadata          map[string]string
	forwardingRulesCmd              *cobra.Command
	forwardingRulesListCmd          *cobra.Command
	flagFRListFormat                string
	addressesCmd                    *cobra.Command
	addressesCreateCmd              *cobra.Command
	addressesListCmd                *cobra.Command
	addressesDeleteCmd              *cobra.Command
	addressesDescribeCmd            *cobra.Command
	forwardingRulesDescribeCmd      *cobra.Command
	forwardingRulesCreateCmd        *cobra.Command
	forwardingRulesDeleteCmd        *cobra.Command
	flagAddressRegion               string
	flagAddressNetworkTier          string
	flagAddressSubnet               string
	flagAddressNetwork              string
	flagAddressPurpose              string
	flagAddressType                 string
	flagAddressAddresses            string
	flagAddressDescription          string
	flagAddressGlobal               bool
	flagAddressListFormat           string
	flagAddrDeleteRegion            string
	flagAddrDescribeRegion          string
	flagFRRegion                    string
	flagFRDescribeRegion            string
	flagFRTarget                    string
	flagFRIPAddress                 string
	flagFRIPProtocol                string
	flagFRPortRange                 string
	flagFRListFilter                string
	flagFRDeleteRegion              string
	flagFRListURI                   bool
	flagAddressListURI              bool
	flagAddressCreateAsync          bool
	flagSubnetsListRegions          []string
	flagSubnetsListNetwork          string
	flagSubnetsListURI              bool
	flagDiskResizeSize              string
	flagDiskResizeZone              string
	flagDiskResizeRegion            string
	flagDiskResizeAsync             bool
	flagDiskResizeFormat            string
	disksResizeCmd                  *cobra.Command
	unmanagedListInstancesCmd       *cobra.Command
	unmanagedAddInstancesCmd        *cobra.Command
	flagInstances                   []string
	flagUnmanagedLIFilter           string
	flagUnmanagedLIFormat           string
	managedCmd                      *cobra.Command
	managedListInstancesCmd         *cobra.Command
	flagManagedRegion               string
	flagManagedListFormat           string
	managedDescribeCmd              *cobra.Command
	flagManagedDescribeRegion       string
	managedResizeCmd                *cobra.Command
	flagManagedResizeRegion         string
	flagManagedResizeSize           int64
	managedCreateCmd                *cobra.Command
	flagManagedTemplate             string
	flagManagedSize                 int64
	flagManagedCreateRegion         string
	managedDeleteCmd                *cobra.Command
	flagManagedDeleteRegion         string
	managedListCmd                  *cobra.Command
	flagManagedListURI              bool
	managedUpdateCmd                *cobra.Command
	flagManagedUpdateRegion         string
	flagManagedUpdateTemplate       string
	managedSetTemplateCmd           *cobra.Command
	flagManagedSetTemplateRegion    string
	flagManagedSetTemplateName      string
	managedRecreateCmd              *cobra.Command
	flagManagedRecreateRegion       string
	flagManagedRecreateInstances    []string
	managedSetAutoscalingCmd        *cobra.Command
	flagAutoscalingRegion           string
	flagAutoscalingMin              int64
	flagAutoscalingMax              int64
	flagAutoscalingCPUTarget        float64
	managedStopAutoscalingCmd       *cobra.Command
	flagStopAutoscalingRegion       string
	flagManagedListZones            []string
	flagManagedListRegions          []string
	unmanagedCreateCmd              *cobra.Command
	flagUnmanagedDescription        string
	unmanagedDeleteCmd              *cobra.Command
	unmanagedRemoveInstancesCmd     *cobra.Command
	flagRemoveInstances             []string
	unmanagedListCmd                *cobra.Command
	flagUnmanagedListFormat         string
	flagUnmanagedListURI            bool
	unmanagedDescribeCmd            *cobra.Command
	unmanagedSetNamedPortsCmd       *cobra.Command
	flagNamedPorts                  []string
	unmanagedGetNamedPortsCmd       *cobra.Command
	instancesStopCmd                *cobra.Command
	instancesStartCmd               *cobra.Command
	flagAsync                       bool
	flagStopDiscardLocalSSD         bool
	instancesAttachDiskCmd          *cobra.Command
	flagAttachDisk                  string
	flagAttachDiskMode              string
	flagAttachDeviceName            string
	instancesDetachDiskCmd          *cobra.Command
	flagDetachDisk                  string
	instancesSetDiskAutoDeleteCmd   *cobra.Command
	flagAutoDeleteDisk              string
	flagAutoDelete                  bool
	flagNoAutoDelete                bool
	instancesDescribeCmd            *cobra.Command
	flagDescribeFormat              string
	instancesListCmd                *cobra.Command
	flagListFilter                  string
	flagListFormat                  string
	flagListURI                     bool
	instancesCreateCmd              *cobra.Command
	flagMachineType                 string
	flagNetwork                     string
	flagSubnet                      string
	flagImageFamily                 string
	flagImageProject                string
	flagBootDiskSize                string
	flagBootDiskType                string
	flagTags                        []string
	flagMetadata                    map[string]string
	flagMetadataFromFile            map[string]string
	flagServiceAccountEmail         string
	flagScopes                      []string
	flagNoAddress                   bool
	flagPreemptible                 bool
	flagProvisioningModel           string
	flagCreateLabels                map[string]string
	flagCreateDeletionProtect       bool
	flagCreateAsync                 bool
	flagCreateDisk                  string
	flagCreateNewDisk               string
	flagPrivateNetworkIP            string
	flagCreateAddress               string
	flagAccelerator                 string
	flagShieldedSecureBoot          bool
	flagShieldedVTPM                bool
	flagShieldedIntegrity           bool
	flagCanIPForward                bool
	flagMinCPUPlatform              string
	instancesDeleteCmd              *cobra.Command
	flagDeleteDisks                 string
	flagKeepDisks                   string
	flagDeleteAsync                 bool
	instancesAddLabelsCmd           *cobra.Command
	instancesRemoveLabelsCmd        *cobra.Command
	flagAddLabels                   map[string]string
	flagRemoveLabelsKeys            string
	flagRemoveLabelsAll             bool
	instancesAddMetadataCmd         *cobra.Command
	instancesRemoveMetadataCmd      *cobra.Command
	flagAddMetadataKV               map[string]string
	instancesResetCmd               *cobra.Command
	instancesSetMachineTypeCmd      *cobra.Command
	flagSetMachineType              string
	instancesGetSerialPortOutputCmd *cobra.Command
	flagSerialPort                  int64
	flagSerialStart                 int64
	flagResetAsync                  bool
	instancesSuspendCmd             *cobra.Command
	instancesResumeCmd              *cobra.Command
	instancesSetSchedulingCmd       *cobra.Command
	flagRestartOnFailure            bool
	flagNoRestartOnFailure          bool
	flagMaintenancePolicy           string
	instancesAddTagsCmd             *cobra.Command
	instancesRemoveTagsCmd          *cobra.Command
	flagInstanceAddTags             []string
	flagInstanceRemoveTags          []string
	flagRemoveAllTags               bool
	instancesUpdateCmd              *cobra.Command
	flagUpdateLabels                map[string]string
	flagUpdateRemoveLabels          string
	flagDeletionProtection          bool
	flagNoDeletionProtection        bool
	networksCmd                     *cobra.Command
	networksListCmd                 *cobra.Command
	flagNetworksListFormat          string
	flagNetworksListFilter          string
	flagNetworksListURI             bool
	scpCmd                          *cobra.Command
	flagSCPTunnelThroughIAP         bool
	flagSCPInternalIP               bool
	flagSCPKeyFile                  string
	flagSCPRecurse                  bool
	flagSCPPort                     int
	flagSCPCompress                 bool
	flagSCPFlag                     []string
	flagSCPPlain                    bool
	flagSCPStrictHostKeyChecking    string
	flagSCPDryRun                   bool
	sshCmd                          *cobra.Command
	flagSSHTunnelThroughIAP         bool
	flagSSHInternalIP               bool
	flagSSHKeyFile                  string
	flagSSHCommand                  string
	flagSSHFlag                     []string
	flagSSHDryRun                   bool
	flagSSHPlain                    bool
	flagSSHStrictHostKeyChecking    string
	flagSSHKeyExpiration            string
	flagSSHKeyExpireAfter           string
	flagSSHTroubleshoot             bool
	flagSSHForceKeyOverwrite        bool
	flagSSHContainer                string
}

// newComputeCmd returns the "compute" command group for inv.
func newComputeCmd(inv *invocation) *cobra.Command {
	return newComputeCmds(inv).computeCmd
}

// newComputeCmds builds the "compute" command group for inv and returns its state.
func newComputeCmds(inv *invocation) *computeCmds {
	s := &computeCmds{inv: inv}
	s.computeCmd = &cobra.Command{
		Use:   "compute",
		Short: "Compute Engine commands",
	}
	s.instancesCmd = &cobra.Command{
		Use:   "instances",
		Short: "Manage Compute Engine instances",
	}
	s.instanceGroupsCmd = &cobra.Command{
		Use:   "instance-groups",
		Short: "Manage instance groups",
	}
	s.unmanagedCmd = &cobra.Command{
		Use:   "unmanaged",
		Short: "Manage unmanaged instance groups",
	}
	s.disksCmd = &cobra.Command{
		Use:   "disks",
		Short: "Manage Compute Engine disks",
	}
	s.projectInfoCmd = &cobra.Command{
		Use:   "project-info",
		Short: "Manage project-level information",
	}
	s.projectInfoDescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe project-level information",
		Args:  cobra.NoArgs,
		RunE:  s.runProjectInfoDescribe,
	}
	s.projectInfoAddMetadataCmd = &cobra.Command{
		Use:   "add-metadata",
		Short: "Add or update project-level metadata",
		Args:  cobra.NoArgs,
		RunE:  s.runProjectInfoAddMetadata,
	}
	s.projectInfoRemoveMetadataCmd = &cobra.Command{
		Use:   "remove-metadata",
		Short: "Remove project-level metadata",
		Args:  cobra.NoArgs,
		RunE:  s.runProjectInfoRemoveMetadata,
	}
	s.forwardingRulesCmd = &cobra.Command{
		Use:   "forwarding-rules",
		Short: "Manage forwarding rules",
	}
	s.forwardingRulesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List forwarding rules",
		Args:  cobra.NoArgs,
		RunE:  s.runForwardingRulesList,
	}
	s.addressesCmd = &cobra.Command{
		Use:   "addresses",
		Short: "Manage static IP addresses",
	}
	s.addressesCreateCmd = &cobra.Command{
		Use:   "create ADDRESS_NAME",
		Short: "Reserve a static IP address",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAddressesCreate,
	}
	s.addressesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List static IP addresses",
		Args:  cobra.NoArgs,
		RunE:  s.runAddressesList,
	}
	s.addressesDeleteCmd = &cobra.Command{
		Use:   "delete ADDRESS_NAME",
		Short: "Delete a static IP address",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAddressesDelete,
	}
	s.addressesDescribeCmd = &cobra.Command{
		Use:   "describe ADDRESS_NAME",
		Short: "Describe a static IP address",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAddressesDescribe,
	}
	s.forwardingRulesDescribeCmd = &cobra.Command{
		Use:   "describe RULE_NAME",
		Short: "Describe a forwarding rule",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runForwardingRulesDescribe,
	}
	s.forwardingRulesCreateCmd = &cobra.Command{
		Use:   "create RULE_NAME",
		Short: "Create a forwarding rule",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runForwardingRulesCreate,
	}
	s.forwardingRulesDeleteCmd = &cobra.Command{
		Use:   "delete RULE_NAME",
		Short: "Delete a forwarding rule",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runForwardingRulesDelete,
	}
	s.disksResizeCmd = &cobra.Command{
		Use:   "resize DISK_NAME [DISK_NAME ...]",
		Short: "Resize (grow) one or more Compute Engine disks",
		Long: "Resize one or more Compute Engine persistent disks. Disk size may " +
			"only be increased; shrinking is not supported by the Compute API. " +
			"This change is not reversible.",
		Args: cobra.MinimumNArgs(1),
		RunE: s.runDisksResize,
	}
	s.unmanagedListInstancesCmd = &cobra.Command{
		Use:   "list-instances INSTANCE_GROUP",
		Short: "List instances in an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runUnmanagedListInstances,
	}
	s.unmanagedAddInstancesCmd = &cobra.Command{
		Use:   "add-instances INSTANCE_GROUP",
		Short: "Add instances to an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runUnmanagedAddInstances,
	}
	s.managedCmd = &cobra.Command{
		Use:   "managed",
		Short: "Manage managed instance groups",
	}
	s.managedListInstancesCmd = &cobra.Command{
		Use:   "list-instances INSTANCE_GROUP",
		Short: "List instances in a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedListInstances,
	}
	s.managedDescribeCmd = &cobra.Command{
		Use:   "describe INSTANCE_GROUP",
		Short: "Describe a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedDescribe,
	}
	s.managedResizeCmd = &cobra.Command{
		Use:   "resize INSTANCE_GROUP",
		Short: "Resize a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedResize,
	}
	s.managedCreateCmd = &cobra.Command{
		Use:   "create INSTANCE_GROUP",
		Short: "Create a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedCreate,
	}
	s.managedDeleteCmd = &cobra.Command{
		Use:   "delete INSTANCE_GROUP",
		Short: "Delete a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedDelete,
	}
	s.managedListCmd = &cobra.Command{
		Use:   "list",
		Short: "List managed instance groups",
		Args:  cobra.NoArgs,
		RunE:  s.runManagedList,
	}
	s.managedUpdateCmd = &cobra.Command{
		Use:   "update INSTANCE_GROUP",
		Short: "Update a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedUpdate,
	}
	s.managedSetTemplateCmd = &cobra.Command{
		Use:   "set-instance-template INSTANCE_GROUP",
		Short: "Set the instance template for a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedSetTemplate,
	}
	s.managedRecreateCmd = &cobra.Command{
		Use:   "recreate-instances INSTANCE_GROUP",
		Short: "Recreate instances in a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedRecreateInstances,
	}
	s.managedSetAutoscalingCmd = &cobra.Command{
		Use:   "set-autoscaling INSTANCE_GROUP",
		Short: "Set autoscaling for a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedSetAutoscaling,
	}
	s.managedStopAutoscalingCmd = &cobra.Command{
		Use:   "stop-autoscaling INSTANCE_GROUP",
		Short: "Stop autoscaling for a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runManagedStopAutoscaling,
	}
	s.unmanagedCreateCmd = &cobra.Command{
		Use:   "create INSTANCE_GROUP",
		Short: "Create an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runUnmanagedCreate,
	}
	s.unmanagedDeleteCmd = &cobra.Command{
		Use:   "delete INSTANCE_GROUP",
		Short: "Delete an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runUnmanagedDelete,
	}
	s.unmanagedRemoveInstancesCmd = &cobra.Command{
		Use:   "remove-instances INSTANCE_GROUP",
		Short: "Remove instances from an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runUnmanagedRemoveInstances,
	}
	s.unmanagedListCmd = &cobra.Command{
		Use:   "list",
		Short: "List unmanaged instance groups",
		Args:  cobra.NoArgs,
		RunE:  s.runUnmanagedList,
	}
	s.unmanagedDescribeCmd = &cobra.Command{
		Use:   "describe INSTANCE_GROUP",
		Short: "Describe an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runUnmanagedDescribe,
	}
	s.unmanagedSetNamedPortsCmd = &cobra.Command{
		Use:   "set-named-ports INSTANCE_GROUP",
		Short: "Set named ports for an instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runUnmanagedSetNamedPorts,
	}
	s.unmanagedGetNamedPortsCmd = &cobra.Command{
		Use:   "get-named-ports INSTANCE_GROUP",
		Short: "Get named ports for an instance group",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runUnmanagedGetNamedPorts,
	}
	s.instancesStopCmd = &cobra.Command{
		Use:   "stop INSTANCE_NAME [INSTANCE_NAME ...]",
		Short: "Stop a Compute Engine instance",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runInstancesStop,
	}
	s.instancesStartCmd = &cobra.Command{
		Use:   "start INSTANCE_NAME [INSTANCE_NAME ...]",
		Short: "Start a Compute Engine instance",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runInstancesStart,
	}
	s.instancesAttachDiskCmd = &cobra.Command{
		Use:   "attach-disk INSTANCE_NAME",
		Short: "Attach a disk to an instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesAttachDisk,
	}
	s.instancesDetachDiskCmd = &cobra.Command{
		Use:   "detach-disk INSTANCE_NAME",
		Short: "Detach a disk from an instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesDetachDisk,
	}
	s.instancesSetDiskAutoDeleteCmd = &cobra.Command{
		Use:   "set-disk-auto-delete INSTANCE_NAME",
		Short: "Set auto-delete policy for a disk on an instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesSetDiskAutoDelete,
	}
	s.instancesDescribeCmd = &cobra.Command{
		Use:   "describe INSTANCE_NAME",
		Short: "Describe a Compute Engine instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesDescribe,
	}
	s.instancesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Compute Engine instances",
		Args:  cobra.NoArgs,
		RunE:  s.runInstancesList,
	}
	s.instancesCreateCmd = &cobra.Command{
		Use:   "create INSTANCE_NAME [INSTANCE_NAME ...]",
		Short: "Create a Compute Engine instance",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runInstancesCreate,
	}
	s.instancesDeleteCmd = &cobra.Command{
		Use:   "delete INSTANCE_NAME [INSTANCE_NAME ...]",
		Short: "Delete a Compute Engine instance",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runInstancesDelete,
	}
	s.instancesAddLabelsCmd = &cobra.Command{
		Use:   "add-labels INSTANCE_NAME",
		Short: "Add labels to an instance",
		Long: `Add or update labels on an instance.

Examples:
  gcloud compute instances add-labels my-instance --labels=env=prod,team=backend --zone=us-central1-a`,
		Args: cobra.ExactArgs(1),
		RunE: s.runInstancesAddLabels,
	}
	s.instancesRemoveLabelsCmd = &cobra.Command{
		Use:   "remove-labels INSTANCE_NAME",
		Short: "Remove labels from an instance",
		Long: `Remove labels from an instance by key.

Examples:
  gcloud compute instances remove-labels my-instance --labels=env,team --zone=us-central1-a
  gcloud compute instances remove-labels my-instance --all --zone=us-central1-a`,
		Args: cobra.ExactArgs(1),
		RunE: s.runInstancesRemoveLabels,
	}
	s.instancesAddMetadataCmd = &cobra.Command{
		Use:   "add-metadata INSTANCE_NAME",
		Short: "Add or update metadata for an instance",
		Long: `Add or update instance metadata. Existing keys are overwritten.

Examples:
  gcloud compute instances add-metadata my-instance --metadata=key1=value1,key2=value2 --zone=us-central1-a`,
		Args: cobra.ExactArgs(1),
		RunE: s.runInstancesAddMetadata,
	}
	s.instancesRemoveMetadataCmd = &cobra.Command{
		Use:   "remove-metadata INSTANCE_NAME",
		Short: "Remove metadata from an instance",
		Long: `Remove instance metadata by key.

Examples:
  gcloud compute instances remove-metadata my-instance --keys=key1,key2 --zone=us-central1-a
  gcloud compute instances remove-metadata my-instance --all --zone=us-central1-a`,
		Args: cobra.ExactArgs(1),
		RunE: s.runInstancesRemoveMetadata,
	}
	s.instancesResetCmd = &cobra.Command{
		Use:   "reset INSTANCE_NAME [INSTANCE_NAME ...]",
		Short: "Reset a Compute Engine instance",
		Long:  "Perform a hard reboot of an instance (distinct from stop+start).",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runInstancesReset,
	}
	s.instancesSetMachineTypeCmd = &cobra.Command{
		Use:   "set-machine-type INSTANCE_NAME",
		Short: "Change the machine type of a stopped instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesSetMachineType,
	}
	s.instancesGetSerialPortOutputCmd = &cobra.Command{
		Use:   "get-serial-port-output INSTANCE_NAME",
		Short: "Read serial console output from an instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesGetSerialPortOutput,
	}
	s.instancesSuspendCmd = &cobra.Command{
		Use:   "suspend INSTANCE_NAME [INSTANCE_NAME ...]",
		Short: "Suspend a Compute Engine instance",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runInstancesSuspend,
	}
	s.instancesResumeCmd = &cobra.Command{
		Use:   "resume INSTANCE_NAME [INSTANCE_NAME ...]",
		Short: "Resume a suspended Compute Engine instance",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runInstancesResume,
	}
	s.instancesSetSchedulingCmd = &cobra.Command{
		Use:   "set-scheduling INSTANCE_NAME",
		Short: "Set scheduling options for an instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesSetScheduling,
	}
	s.instancesAddTagsCmd = &cobra.Command{
		Use:   "add-tags INSTANCE_NAME",
		Short: "Add network tags to an instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesAddTags,
	}
	s.instancesRemoveTagsCmd = &cobra.Command{
		Use:   "remove-tags INSTANCE_NAME",
		Short: "Remove network tags from an instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesRemoveTags,
	}
	s.instancesUpdateCmd = &cobra.Command{
		Use:   "update INSTANCE_NAME",
		Short: "Update a Compute Engine instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runInstancesUpdate,
	}
	s.networksCmd = &cobra.Command{
		Use:   "networks",
		Short: "Manage VPC networks",
	}
	s.networksListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Compute Engine networks",
		Args:  cobra.NoArgs,
		RunE:  s.runNetworksList,
	}
	s.scpCmd = &cobra.Command{
		Use:   "scp [[USER@]INSTANCE:]SRC [[[USER@]INSTANCE:]SRC ...] [[USER@]INSTANCE:]DEST",
		Short: "Copy files to/from a Compute Engine instance via SCP",
		Args:  cobra.MinimumNArgs(2),
		RunE:  s.runSCP,
	}
	s.sshCmd = &cobra.Command{
		Use:   "ssh [USER@]INSTANCE [-- REMOTE_COMMAND]",
		Short: "SSH into a Compute Engine instance",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runSSH,
	}
	s.initCompute()
	s.initComputeImagesList()
	s.initComputeLists()
	s.initComputeMisc()
	s.initComputeSubnetsList()
	s.initDisksResize()
	s.initInstanceGroups()
	s.initInstanceGroupsManaged()
	s.initInstanceGroupsManagedExtra()
	s.initInstanceGroupsManagedList()
	s.initInstanceGroupsUnmanaged()
	s.initInstances()
	s.initInstancesDisk()
	s.initInstancesExtra()
	s.initInstancesLabels()
	s.initInstancesMetadata()
	s.initInstancesOps()
	s.initInstancesScheduling()
	s.initInstancesTags()
	s.initInstancesUpdate()
	s.initNetworks()
	s.initScp()
	s.initSsh()
	return s
}
