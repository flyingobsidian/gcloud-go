package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- project-info ---

var projectInfoCmd = &cobra.Command{
	Use:   "project-info",
	Short: "Manage project-level information",
}

var projectInfoDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe project-level information",
	Args:  cobra.NoArgs,
	RunE:  runProjectInfoDescribe,
}

var projectInfoAddMetadataCmd = &cobra.Command{
	Use:   "add-metadata",
	Short: "Add or update project-level metadata",
	Args:  cobra.NoArgs,
	RunE:  runProjectInfoAddMetadata,
}

var projectInfoRemoveMetadataCmd = &cobra.Command{
	Use:   "remove-metadata",
	Short: "Remove project-level metadata",
	Args:  cobra.NoArgs,
	RunE:  runProjectInfoRemoveMetadata,
}

var (
	flagRemoveMetadataKeys string
	flagRemoveMetadataAll  bool
	flagAddProjectMetadata map[string]string
)

// --- forwarding-rules ---

var forwardingRulesCmd = &cobra.Command{
	Use:   "forwarding-rules",
	Short: "Manage forwarding rules",
}

var forwardingRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List forwarding rules",
	Args:  cobra.NoArgs,
	RunE:  runForwardingRulesList,
}

var flagFRListFormat string

// --- addresses ---

var addressesCmd = &cobra.Command{
	Use:   "addresses",
	Short: "Manage static IP addresses",
}

var addressesCreateCmd = &cobra.Command{
	Use:   "create ADDRESS_NAME",
	Short: "Reserve a static IP address",
	Args:  cobra.ExactArgs(1),
	RunE:  runAddressesCreate,
}

var addressesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List static IP addresses",
	Args:  cobra.NoArgs,
	RunE:  runAddressesList,
}

var addressesDeleteCmd = &cobra.Command{
	Use:   "delete ADDRESS_NAME",
	Short: "Delete a static IP address",
	Args:  cobra.ExactArgs(1),
	RunE:  runAddressesDelete,
}

var addressesDescribeCmd = &cobra.Command{
	Use:   "describe ADDRESS_NAME",
	Short: "Describe a static IP address",
	Args:  cobra.ExactArgs(1),
	RunE:  runAddressesDescribe,
}

var forwardingRulesDescribeCmd = &cobra.Command{
	Use:   "describe RULE_NAME",
	Short: "Describe a forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE:  runForwardingRulesDescribe,
}

var forwardingRulesCreateCmd = &cobra.Command{
	Use:   "create RULE_NAME",
	Short: "Create a forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE:  runForwardingRulesCreate,
}

var forwardingRulesDeleteCmd = &cobra.Command{
	Use:   "delete RULE_NAME",
	Short: "Delete a forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE:  runForwardingRulesDelete,
}

var (
	flagAddressRegion      string
	flagAddressNetworkTier string
	flagAddressSubnet      string
	flagAddressNetwork     string
	flagAddressPurpose     string
	flagAddressType        string
	flagAddressAddresses   string
	flagAddressDescription string
	flagAddressGlobal      bool
	flagAddressListFormat  string
	flagAddrDeleteRegion   string
	flagAddrDescribeRegion string
	flagFRRegion           string
	flagFRDescribeRegion   string
	flagFRTarget           string
	flagFRIPAddress        string
	flagFRIPProtocol       string
	flagFRPortRange        string
	flagFRListFilter       string
	flagFRDeleteRegion     string
	flagFRListURI          bool
	flagAddressListURI     bool
	flagAddressCreateAsync bool
)

func init() {
	// project-info
	projectInfoRemoveMetadataCmd.Flags().StringVar(&flagRemoveMetadataKeys, "keys", "", "Comma-separated list of metadata keys to remove")
	projectInfoRemoveMetadataCmd.Flags().BoolVar(&flagRemoveMetadataAll, "all", false, "Remove all metadata")
	projectInfoAddMetadataCmd.Flags().StringToStringVar(&flagAddProjectMetadata, "metadata", nil, "Metadata key=value pairs to add")
	projectInfoAddMetadataCmd.MarkFlagRequired("metadata")
	projectInfoCmd.AddCommand(projectInfoDescribeCmd)
	projectInfoCmd.AddCommand(projectInfoAddMetadataCmd)
	projectInfoCmd.AddCommand(projectInfoRemoveMetadataCmd)
	computeCmd.AddCommand(projectInfoCmd)

	// forwarding-rules
	forwardingRulesListCmd.Flags().StringVar(&flagFRListFormat, "format", "", "Output format (e.g. json)")
	forwardingRulesListCmd.Flags().StringVar(&flagFRListFilter, "filter", "", "Filter expression")
	forwardingRulesListCmd.Flags().BoolVar(&flagFRListURI, "uri", false, "Print resource URIs")
	forwardingRulesDescribeCmd.Flags().StringVar(&flagFRDescribeRegion, "region", "", "Region")
	forwardingRulesCreateCmd.Flags().StringVar(&flagFRRegion, "region", "", "Region")
	forwardingRulesCreateCmd.Flags().StringVar(&flagFRTarget, "target-pool", "", "Target pool")
	forwardingRulesCreateCmd.Flags().StringVar(&flagFRIPAddress, "address", "", "IP address")
	forwardingRulesCreateCmd.Flags().StringVar(&flagFRIPProtocol, "ip-protocol", "", "IP protocol (TCP, UDP)")
	forwardingRulesCreateCmd.Flags().StringVar(&flagFRPortRange, "port-range", "", "Port range (e.g. 80-80)")
	forwardingRulesDeleteCmd.Flags().StringVar(&flagFRDeleteRegion, "region", "", "Region")
	forwardingRulesCmd.AddCommand(forwardingRulesListCmd)
	forwardingRulesCmd.AddCommand(forwardingRulesDescribeCmd)
	forwardingRulesCmd.AddCommand(forwardingRulesCreateCmd)
	forwardingRulesCmd.AddCommand(forwardingRulesDeleteCmd)
	computeCmd.AddCommand(forwardingRulesCmd)

	// addresses
	addressesCreateCmd.Flags().StringVar(&flagAddressRegion, "region", "", "Region for the address")
	addressesCreateCmd.Flags().StringVar(&flagAddressNetworkTier, "network-tier", "", "Network tier (PREMIUM or STANDARD)")
	addressesCreateCmd.Flags().StringVar(&flagAddressSubnet, "subnet", "", "Subnet for internal address")
	addressesCreateCmd.Flags().StringVar(&flagAddressNetwork, "network", "", "Network for internal address")
	addressesCreateCmd.Flags().StringVar(&flagAddressPurpose, "purpose", "", "Purpose of the address")
	addressesCreateCmd.Flags().StringVar(&flagAddressType, "address-type", "", "Address type (INTERNAL or EXTERNAL)")
	addressesCreateCmd.Flags().StringVar(&flagAddressAddresses, "addresses", "", "Specific IP address to reserve")
	addressesCreateCmd.Flags().StringVar(&flagAddressDescription, "description", "", "Description for the address")
	addressesCreateCmd.Flags().BoolVar(&flagAddressGlobal, "global", false, "Create a global address")
	addressesCreateCmd.Flags().BoolVar(&flagAddressCreateAsync, "async", false, "Return immediately without waiting")
	addressesListCmd.Flags().StringVar(&flagAddressListFormat, "format", "", "Output format (e.g. json)")
	addressesListCmd.Flags().BoolVar(&flagAddressListURI, "uri", false, "Print resource URIs")
	addressesDeleteCmd.Flags().StringVar(&flagAddrDeleteRegion, "region", "", "Region")
	addressesDescribeCmd.Flags().StringVar(&flagAddrDescribeRegion, "region", "", "Region")
	addressesCmd.AddCommand(addressesCreateCmd)
	addressesCmd.AddCommand(addressesListCmd)
	addressesCmd.AddCommand(addressesDeleteCmd)
	addressesCmd.AddCommand(addressesDescribeCmd)
	computeCmd.AddCommand(addressesCmd)
}

func runProjectInfoDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	proj, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project: %w", err)
	}

	return formatOutput(proj, "")
}

func runProjectInfoAddMetadata(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	proj, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting project metadata: %w", err)
	}

	metadata := proj.CommonInstanceMetadata
	if metadata == nil {
		metadata = &compute.Metadata{}
	}

	for k, v := range flagAddProjectMetadata {
		val := v
		found := false
		for _, item := range metadata.Items {
			if item.Key == k {
				item.Value = &val
				found = true
				break
			}
		}
		if !found {
			metadata.Items = append(metadata.Items, &compute.MetadataItems{Key: k, Value: &val})
		}
	}

	op, err := svc.Projects.SetCommonInstanceMetadata(project, metadata).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting metadata: %w", err)
	}

	if err := waitForGlobalOp(ctx, svc, project, op.Name); err != nil {
		return err
	}
	fmt.Println("Updated project metadata.")
	return nil
}

func runProjectInfoRemoveMetadata(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Get current project metadata.
	proj, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting project metadata: %w", err)
	}

	var newItems []*compute.MetadataItems
	if flagRemoveMetadataAll {
		if len(proj.CommonInstanceMetadata.Items) == 0 {
			fmt.Fprintln(os.Stderr, "No change requested; skipping update.")
			return nil
		}
		newItems = nil
	} else if flagRemoveMetadataKeys == "" {
		return fmt.Errorf("one of --keys or --all is required")
	} else {
		keysToRemove := strings.Split(flagRemoveMetadataKeys, ",")
		removeSet := make(map[string]bool)
		for _, k := range keysToRemove {
			removeSet[strings.TrimSpace(k)] = true
		}
		found := false
		for _, item := range proj.CommonInstanceMetadata.Items {
			if removeSet[item.Key] {
				found = true
			} else {
				newItems = append(newItems, item)
			}
		}
		if !found {
			fmt.Fprintln(os.Stderr, "No change requested; skipping update.")
			return nil
		}
	}

	proj.CommonInstanceMetadata.Items = newItems
	op, err := svc.Projects.SetCommonInstanceMetadata(project, proj.CommonInstanceMetadata).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing metadata: %w", err)
	}

	// Wait for global operation.
	if err := waitForGlobalOp(ctx, svc, project, op.Name); err != nil {
		return err
	}

	fmt.Println("Updated project metadata.")
	return nil
}

func runForwardingRulesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var allRules []*compute.ForwardingRule
	pageToken := ""
	for {
		call := svc.ForwardingRules.AggregatedList(project).Context(ctx)
		if flagFRListFilter != "" {
			call = call.Filter(flagFRListFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing forwarding rules: %w", err)
		}
		for _, scoped := range resp.Items {
			allRules = append(allRules, scoped.ForwardingRules...)
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagFRListURI {
		for _, fr := range allRules {
			fmt.Println(fr.SelfLink)
		}
		return nil
	}

	if flagFRListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allRules)
	}

	fmt.Printf("%-30s %-20s %-15s %-10s %-30s\n", "NAME", "REGION", "IP_ADDRESS", "IP_PROTOCOL", "TARGET")
	for _, fr := range allRules {
		fmt.Printf("%-30s %-20s %-15s %-10s %-30s\n", fr.Name, fr.Region, fr.IPAddress, fr.IPProtocol, fr.Target)
	}
	return nil
}

func runAddressesCreate(cmd *cobra.Command, args []string) error {
	// Cross-flag validation (#148).
	if flagAddressNetwork != "" && flagAddressSubnet != "" {
		return fmt.Errorf("--network and --subnet are mutually exclusive")
	}

	name := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	addr := &compute.Address{
		Name: name,
	}
	if flagAddressNetworkTier != "" {
		addr.NetworkTier = flagAddressNetworkTier
	}
	if flagAddressSubnet != "" {
		addr.Subnetwork = flagAddressSubnet
	}
	if flagAddressNetwork != "" {
		addr.Network = flagAddressNetwork
	}
	if flagAddressPurpose != "" {
		addr.Purpose = flagAddressPurpose
	}
	if flagAddressType != "" {
		addr.AddressType = flagAddressType
	}
	if flagAddressAddresses != "" {
		addr.Address = flagAddressAddresses
	}
	if flagAddressDescription != "" {
		addr.Description = flagAddressDescription
	}

	if flagAddressGlobal {
		op, err := svc.GlobalAddresses.Insert(project, addr).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating global address: %w", err)
		}
		if flagAddressCreateAsync {
			fmt.Printf("Create operation started for [%s]: %s\n", name, op.Name)
			return nil
		}
		if err := waitForGlobalOp(ctx, svc, project, op.Name); err != nil {
			return err
		}
		fmt.Printf("Created global address [%s].\n", name)
		return nil
	}

	region := flagAddressRegion
	if region == "" {
		props, err := config.Load()
		if err != nil {
			return err
		}
		region = config.Resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required (or use --global)")
		}
	}

	op, err := svc.Addresses.Insert(project, region, addr).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating address: %w", err)
	}

	if flagAddressCreateAsync {
		fmt.Printf("Create operation started for [%s]: %s\n", name, op.Name)
		return nil
	}

	if err := waitForRegionOp(ctx, svc, project, region, op.Name); err != nil {
		return err
	}

	fmt.Printf("Created address [%s].\n", name)
	return nil
}

func runAddressesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var allAddrs []*compute.Address
	pageToken := ""
	for {
		call := svc.Addresses.AggregatedList(project).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing addresses: %w", err)
		}
		for _, scoped := range resp.Items {
			allAddrs = append(allAddrs, scoped.Addresses...)
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagAddressListURI {
		for _, a := range allAddrs {
			fmt.Println(a.SelfLink)
		}
		return nil
	}

	if flagAddressListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allAddrs)
	}

	fmt.Printf("%-30s %-20s %-18s %-10s %-10s\n", "NAME", "REGION", "ADDRESS", "STATUS", "TYPE")
	for _, a := range allAddrs {
		fmt.Printf("%-30s %-20s %-18s %-10s %-10s\n", a.Name, a.Region, a.Address, a.Status, a.AddressType)
	}
	return nil
}

func runAddressesDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	region := flagAddrDeleteRegion
	if region == "" {
		props, err := config.Load()
		if err != nil {
			return err
		}
		region = config.Resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	op, err := svc.Addresses.Delete(project, region, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting address: %w", err)
	}

	if err := waitForRegionOp(ctx, svc, project, region, op.Name); err != nil {
		return err
	}
	fmt.Printf("Deleted address [%s].\n", name)
	return nil
}

func runAddressesDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	region := flagAddrDescribeRegion
	if region == "" {
		props, err := config.Load()
		if err != nil {
			return err
		}
		region = config.Resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	addr, err := svc.Addresses.Get(project, region, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing address: %w", err)
	}

	return formatOutput(addr, "")
}

func runForwardingRulesDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	region := flagFRDescribeRegion
	if region == "" {
		props, err := config.Load()
		if err != nil {
			return err
		}
		region = config.Resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	fr, err := svc.ForwardingRules.Get(project, region, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing forwarding rule: %w", err)
	}

	return formatOutput(fr, "")
}

func runForwardingRulesCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	region := flagFRRegion
	if region == "" {
		props, err := config.Load()
		if err != nil {
			return err
		}
		region = config.Resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	fr := &compute.ForwardingRule{
		Name: name,
	}
	if flagFRTarget != "" {
		fr.Target = flagFRTarget
	}
	if flagFRIPAddress != "" {
		fr.IPAddress = flagFRIPAddress
	}
	if flagFRIPProtocol != "" {
		fr.IPProtocol = flagFRIPProtocol
	}
	if flagFRPortRange != "" {
		fr.PortRange = flagFRPortRange
	}

	op, err := svc.ForwardingRules.Insert(project, region, fr).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating forwarding rule: %w", err)
	}

	if err := waitForRegionOp(ctx, svc, project, region, op.Name); err != nil {
		return err
	}
	fmt.Printf("Created forwarding rule [%s].\n", name)
	return nil
}

func runForwardingRulesDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	region := flagFRDeleteRegion
	if region == "" {
		props, err := config.Load()
		if err != nil {
			return err
		}
		region = config.Resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	op, err := svc.ForwardingRules.Delete(project, region, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting forwarding rule: %w", err)
	}

	if err := waitForRegionOp(ctx, svc, project, region, op.Name); err != nil {
		return err
	}
	fmt.Printf("Deleted forwarding rule [%s].\n", name)
	return nil
}

// --- Helpers ---

func loadProps() (*config.Properties, error) {
	return config.Load()
}

func resolveProjectOnly(props *config.Properties) string {
	return config.Resolve(flagProject, "CLOUDSDK_CORE_PROJECT", props.Core.Project)
}

func waitForGlobalOp(ctx context.Context, svc *compute.Service, project, opName string) error {
	deadline := time.Now().Add(30 * time.Minute)
	for {
		op, err := svc.GlobalOperations.Get(project, opName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Status == "DONE" {
			if op.Error != nil && len(op.Error.Errors) > 0 {
				return fmt.Errorf("operation failed: %s", op.Error.Errors[0].Message)
			}
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for operation %s", opName)
		}
		time.Sleep(2 * time.Second)
	}
}

func waitForRegionOp(ctx context.Context, svc *compute.Service, project, region, opName string) error {
	deadline := time.Now().Add(30 * time.Minute)
	for {
		op, err := svc.RegionOperations.Get(project, region, opName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Status == "DONE" {
			if op.Error != nil && len(op.Error.Errors) > 0 {
				return fmt.Errorf("operation failed: %s", op.Error.Errors[0].Message)
			}
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for operation %s", opName)
		}
		time.Sleep(2 * time.Second)
	}
}
