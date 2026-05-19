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

var projectInfoRemoveMetadataCmd = &cobra.Command{
	Use:   "remove-metadata",
	Short: "Remove project-level metadata",
	Args:  cobra.NoArgs,
	RunE:  runProjectInfoRemoveMetadata,
}

var (
	flagRemoveMetadataKeys string
	flagRemoveMetadataAll  bool
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

var (
	flagAddressRegion      string
	flagAddressNetworkTier string
	flagAddressSubnet      string
	flagAddressNetwork     string
	flagAddressPurpose     string
	flagAddressType        string
	flagAddressAddresses   string
	flagAddressDescription string
)

func init() {
	// project-info remove-metadata
	projectInfoRemoveMetadataCmd.Flags().StringVar(&flagRemoveMetadataKeys, "keys", "", "Comma-separated list of metadata keys to remove")
	projectInfoRemoveMetadataCmd.Flags().BoolVar(&flagRemoveMetadataAll, "all", false, "Remove all metadata")
	projectInfoCmd.AddCommand(projectInfoRemoveMetadataCmd)
	computeCmd.AddCommand(projectInfoCmd)

	// forwarding-rules list
	forwardingRulesListCmd.Flags().StringVar(&flagFRListFormat, "format", "", "Output format (e.g. json)")
	forwardingRulesCmd.AddCommand(forwardingRulesListCmd)
	computeCmd.AddCommand(forwardingRulesCmd)

	// addresses create
	addressesCreateCmd.Flags().StringVar(&flagAddressRegion, "region", "", "Region for the address")
	addressesCreateCmd.Flags().StringVar(&flagAddressNetworkTier, "network-tier", "", "Network tier (PREMIUM or STANDARD)")
	addressesCreateCmd.Flags().StringVar(&flagAddressSubnet, "subnet", "", "Subnet for internal address")
	addressesCreateCmd.Flags().StringVar(&flagAddressNetwork, "network", "", "Network for internal address")
	addressesCreateCmd.Flags().StringVar(&flagAddressPurpose, "purpose", "", "Purpose of the address")
	addressesCreateCmd.Flags().StringVar(&flagAddressType, "address-type", "", "Address type (INTERNAL or EXTERNAL)")
	addressesCreateCmd.Flags().StringVar(&flagAddressAddresses, "addresses", "", "Specific IP address to reserve")
	addressesCreateCmd.Flags().StringVar(&flagAddressDescription, "description", "", "Description for the address")
	addressesCmd.AddCommand(addressesCreateCmd)
	computeCmd.AddCommand(addressesCmd)
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
		newItems = nil
	} else if flagRemoveMetadataKeys == "" {
		return fmt.Errorf("one of --keys or --all is required")
	} else {
		keysToRemove := strings.Split(flagRemoveMetadataKeys, ",")
		removeSet := make(map[string]bool)
		for _, k := range keysToRemove {
			removeSet[strings.TrimSpace(k)] = true
		}
		for _, item := range proj.CommonInstanceMetadata.Items {
			if !removeSet[item.Key] {
				newItems = append(newItems, item)
			}
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

	if flagFRListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allRules)
	}

	fmt.Printf("%-40s %-20s %-15s %-10s\n", "NAME", "REGION", "IP_ADDRESS", "IP_PROTOCOL")
	for _, fr := range allRules {
		fmt.Printf("%-40s %-20s %-15s %-10s\n", fr.Name, fr.Region, fr.IPAddress, fr.IPProtocol)
	}
	return nil
}

func runAddressesCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	region := flagAddressRegion
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

	op, err := svc.Addresses.Insert(project, region, addr).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating address: %w", err)
	}

	// Wait for region operation.
	if err := waitForRegionOp(ctx, svc, project, region, op.Name); err != nil {
		return err
	}

	fmt.Printf("Created address [%s].\n", name)
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
