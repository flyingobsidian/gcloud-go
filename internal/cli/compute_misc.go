package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- project-info ---

// --- forwarding-rules ---

// --- addresses ---

func (s *computeCmds) initComputeMisc() {
	// project-info
	s.projectInfoRemoveMetadataCmd.Flags().StringVar(&s.flagRemoveMetadataKeys, "keys", "", "Comma-separated list of metadata keys to remove")
	s.projectInfoRemoveMetadataCmd.Flags().BoolVar(&s.flagRemoveMetadataAll, "all", false, "Remove all metadata")
	s.projectInfoAddMetadataCmd.Flags().StringToStringVar(&s.flagAddProjectMetadata, "metadata", nil, "Metadata key=value pairs to add")
	s.projectInfoAddMetadataCmd.MarkFlagRequired("metadata")
	s.projectInfoCmd.AddCommand(s.projectInfoDescribeCmd)
	s.projectInfoCmd.AddCommand(s.projectInfoAddMetadataCmd)
	s.projectInfoCmd.AddCommand(s.projectInfoRemoveMetadataCmd)
	s.computeCmd.AddCommand(s.projectInfoCmd)

	// forwarding-rules
	s.forwardingRulesListCmd.Flags().StringVar(&s.flagFRListFormat, "format", "", "Output format (e.g. json)")
	s.forwardingRulesListCmd.Flags().StringVar(&s.flagFRListFilter, "filter", "", "Filter expression")
	s.forwardingRulesListCmd.Flags().BoolVar(&s.flagFRListURI, "uri", false, "Print resource URIs")
	s.forwardingRulesDescribeCmd.Flags().StringVar(&s.flagFRDescribeRegion, "region", "", "Region")
	s.forwardingRulesCreateCmd.Flags().StringVar(&s.flagFRRegion, "region", "", "Region")
	s.forwardingRulesCreateCmd.Flags().StringVar(&s.flagFRTarget, "target-pool", "", "Target pool")
	s.forwardingRulesCreateCmd.Flags().StringVar(&s.flagFRIPAddress, "address", "", "IP address")
	s.forwardingRulesCreateCmd.Flags().StringVar(&s.flagFRIPProtocol, "ip-protocol", "", "IP protocol (TCP, UDP)")
	s.forwardingRulesCreateCmd.Flags().StringVar(&s.flagFRPortRange, "port-range", "", "Port range (e.g. 80-80)")
	s.forwardingRulesDeleteCmd.Flags().StringVar(&s.flagFRDeleteRegion, "region", "", "Region")
	s.forwardingRulesCmd.AddCommand(s.forwardingRulesListCmd)
	s.forwardingRulesCmd.AddCommand(s.forwardingRulesDescribeCmd)
	s.forwardingRulesCmd.AddCommand(s.forwardingRulesCreateCmd)
	s.forwardingRulesCmd.AddCommand(s.forwardingRulesDeleteCmd)
	s.computeCmd.AddCommand(s.forwardingRulesCmd)

	// addresses
	s.addressesCreateCmd.Flags().StringVar(&s.flagAddressRegion, "region", "", "Region for the address")
	s.addressesCreateCmd.Flags().StringVar(&s.flagAddressNetworkTier, "network-tier", "", "Network tier (PREMIUM or STANDARD)")
	s.addressesCreateCmd.Flags().StringVar(&s.flagAddressSubnet, "subnet", "", "Subnet for internal address")
	s.addressesCreateCmd.Flags().StringVar(&s.flagAddressNetwork, "network", "", "Network for internal address")
	s.addressesCreateCmd.Flags().StringVar(&s.flagAddressPurpose, "purpose", "", "Purpose of the address")
	s.addressesCreateCmd.Flags().StringVar(&s.flagAddressType, "address-type", "", "Address type (INTERNAL or EXTERNAL)")
	s.addressesCreateCmd.Flags().StringVar(&s.flagAddressAddresses, "addresses", "", "Specific IP address to reserve")
	s.addressesCreateCmd.Flags().StringVar(&s.flagAddressDescription, "description", "", "Description for the address")
	s.addressesCreateCmd.Flags().BoolVar(&s.flagAddressGlobal, "global", false, "Create a global address")
	s.addressesCreateCmd.Flags().BoolVar(&s.flagAddressCreateAsync, "async", false, "Return immediately without waiting")
	s.addressesListCmd.Flags().StringVar(&s.flagAddressListFormat, "format", "", "Output format (e.g. json)")
	s.addressesListCmd.Flags().BoolVar(&s.flagAddressListURI, "uri", false, "Print resource URIs")
	s.addressesDeleteCmd.Flags().StringVar(&s.flagAddrDeleteRegion, "region", "", "Region")
	s.addressesDescribeCmd.Flags().StringVar(&s.flagAddrDescribeRegion, "region", "", "Region")
	s.addressesCmd.AddCommand(s.addressesCreateCmd)
	s.addressesCmd.AddCommand(s.addressesListCmd)
	s.addressesCmd.AddCommand(s.addressesDeleteCmd)
	s.addressesCmd.AddCommand(s.addressesDescribeCmd)
	s.computeCmd.AddCommand(s.addressesCmd)
}

func (s *computeCmds) runProjectInfoDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	proj, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project: %w", err)
	}

	// #1742: honour the global --format flag. gcloud-python defaults describe
	// commands to yaml; use that when no format is specified.
	format := s.inv.flags.format
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, proj, format)
}

func (s *computeCmds) runProjectInfoAddMetadata(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
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

	for k, v := range s.flagAddProjectMetadata {
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
	fmt.Fprintln(s.inv.stdout, "Updated project metadata.")
	return nil
}

func (s *computeCmds) runProjectInfoRemoveMetadata(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	// Get current project metadata.
	proj, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting project metadata: %w", err)
	}

	var newItems []*compute.MetadataItems
	if s.flagRemoveMetadataAll {
		if len(proj.CommonInstanceMetadata.Items) == 0 {
			fmt.Fprintln(s.inv.stderr, "No change requested; skipping update.")
			return nil
		}
		newItems = nil
	} else if s.flagRemoveMetadataKeys == "" {
		return fmt.Errorf("one of --keys or --all is required")
	} else {
		keysToRemove := strings.Split(s.flagRemoveMetadataKeys, ",")
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
			fmt.Fprintln(s.inv.stderr, "No change requested; skipping update.")
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

	fmt.Fprintln(s.inv.stdout, "Updated project metadata.")
	return nil
}

func (s *computeCmds) runForwardingRulesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	var allRules []*compute.ForwardingRule
	pageToken := ""
	for {
		call := svc.ForwardingRules.AggregatedList(project).Context(ctx)
		if s.flagFRListFilter != "" {
			call = call.Filter(s.flagFRListFilter)
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

	if s.flagFRListURI {
		for _, fr := range allRules {
			fmt.Fprintln(s.inv.stdout, fr.SelfLink)
		}
		return nil
	}

	if s.flagFRListFormat == "json" {
		return s.inv.printJSON(allRules)
	}

	printTableHeadingTo(s.inv.stdout, len(allRules), "%-30s %-20s %-15s %-10s %-30s\n", "NAME", "REGION", "IP_ADDRESS", "IP_PROTOCOL", "TARGET")
	for _, fr := range allRules {
		fmt.Fprintf(s.inv.stdout, "%-30s %-20s %-15s %-10s %-30s\n", fr.Name, fr.Region, fr.IPAddress, fr.IPProtocol, fr.Target)
	}
	return nil
}

func (s *computeCmds) runAddressesCreate(cmd *cobra.Command, args []string) error {
	// Cross-flag validation (#148).
	if s.flagAddressNetwork != "" && s.flagAddressSubnet != "" {
		return fmt.Errorf("--network and --subnet are mutually exclusive")
	}

	name := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	addr := &compute.Address{
		Name: name,
	}
	if s.flagAddressNetworkTier != "" {
		addr.NetworkTier = s.flagAddressNetworkTier
	}
	if s.flagAddressSubnet != "" {
		addr.Subnetwork = s.flagAddressSubnet
	}
	if s.flagAddressNetwork != "" {
		addr.Network = s.flagAddressNetwork
	}
	if s.flagAddressPurpose != "" {
		addr.Purpose = s.flagAddressPurpose
	}
	if s.flagAddressType != "" {
		addr.AddressType = s.flagAddressType
	}
	if s.flagAddressAddresses != "" {
		addr.Address = s.flagAddressAddresses
	}
	if s.flagAddressDescription != "" {
		addr.Description = s.flagAddressDescription
	}

	if s.flagAddressGlobal {
		op, err := svc.GlobalAddresses.Insert(project, addr).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating global address: %w", err)
		}
		if s.flagAddressCreateAsync {
			fmt.Fprintf(s.inv.stdout, "Create operation started for [%s]: %s\n", name, op.Name)
			return nil
		}
		if err := waitForGlobalOp(ctx, svc, project, op.Name); err != nil {
			return err
		}
		fmt.Fprintf(s.inv.stdout, "Created global address [%s].\n", name)
		return nil
	}

	region := s.flagAddressRegion
	if region == "" {
		props, err := s.inv.properties()
		if err != nil {
			return err
		}
		region = s.inv.resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required (or use --global)")
		}
	}

	op, err := svc.Addresses.Insert(project, region, addr).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating address: %w", err)
	}

	if s.flagAddressCreateAsync {
		fmt.Fprintf(s.inv.stdout, "Create operation started for [%s]: %s\n", name, op.Name)
		return nil
	}

	if err := waitForRegionOp(ctx, svc, project, region, op.Name); err != nil {
		return err
	}

	fmt.Fprintf(s.inv.stdout, "Created address [%s].\n", name)
	return nil
}

func (s *computeCmds) runAddressesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
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

	if s.flagAddressListURI {
		for _, a := range allAddrs {
			fmt.Fprintln(s.inv.stdout, a.SelfLink)
		}
		return nil
	}

	if s.flagAddressListFormat == "json" {
		return s.inv.printJSON(allAddrs)
	}

	printTableHeadingTo(s.inv.stdout, len(allAddrs), "%-30s %-20s %-18s %-10s %-10s\n", "NAME", "REGION", "ADDRESS", "STATUS", "TYPE")
	for _, a := range allAddrs {
		fmt.Fprintf(s.inv.stdout, "%-30s %-20s %-18s %-10s %-10s\n", a.Name, a.Region, a.Address, a.Status, a.AddressType)
	}
	return nil
}

func (s *computeCmds) runAddressesDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	region := s.flagAddrDeleteRegion
	if region == "" {
		props, err := s.inv.properties()
		if err != nil {
			return err
		}
		region = s.inv.resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Deleted address [%s].\n", name)
	return nil
}

func (s *computeCmds) runAddressesDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	region := s.flagAddrDescribeRegion
	if region == "" {
		props, err := s.inv.properties()
		if err != nil {
			return err
		}
		region = s.inv.resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	addr, err := svc.Addresses.Get(project, region, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing address: %w", err)
	}

	return s.inv.formatOutput(addr, "")
}

func (s *computeCmds) runForwardingRulesDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	region := s.flagFRDescribeRegion
	if region == "" {
		props, err := s.inv.properties()
		if err != nil {
			return err
		}
		region = s.inv.resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	fr, err := svc.ForwardingRules.Get(project, region, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing forwarding rule: %w", err)
	}

	return s.inv.formatOutput(fr, "")
}

func (s *computeCmds) runForwardingRulesCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	region := s.flagFRRegion
	if region == "" {
		props, err := s.inv.properties()
		if err != nil {
			return err
		}
		region = s.inv.resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	fr := &compute.ForwardingRule{
		Name: name,
	}
	if s.flagFRTarget != "" {
		fr.Target = s.flagFRTarget
	}
	if s.flagFRIPAddress != "" {
		fr.IPAddress = s.flagFRIPAddress
	}
	if s.flagFRIPProtocol != "" {
		fr.IPProtocol = s.flagFRIPProtocol
	}
	if s.flagFRPortRange != "" {
		fr.PortRange = s.flagFRPortRange
	}

	op, err := svc.ForwardingRules.Insert(project, region, fr).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating forwarding rule: %w", err)
	}

	if err := waitForRegionOp(ctx, svc, project, region, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Created forwarding rule [%s].\n", name)
	return nil
}

func (s *computeCmds) runForwardingRulesDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	region := s.flagFRDeleteRegion
	if region == "" {
		props, err := s.inv.properties()
		if err != nil {
			return err
		}
		region = s.inv.resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
		if region == "" {
			return fmt.Errorf("--region is required")
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Deleted forwarding rule [%s].\n", name)
	return nil
}

// --- Helpers ---

func (s *computeCmds) loadProps() (*config.Properties, error) {
	return s.inv.properties()
}

func (s *computeCmds) resolveProjectOnly(props *config.Properties) string {
	return s.inv.resolve(s.inv.flags.project, "CLOUDSDK_CORE_PROJECT", props.Core.Project)
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
