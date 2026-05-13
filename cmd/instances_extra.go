package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	icompute "github.com/flyingobsidian/gcloud-golang-cli/internal/compute"
	"github.com/flyingobsidian/gcloud-golang-cli/internal/config"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- instances describe ---

var instancesDescribeCmd = &cobra.Command{
	Use:   "describe INSTANCE_NAME",
	Short: "Describe a Compute Engine instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesDescribe,
}

var flagDescribeFormat string

// --- instances list ---

var instancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Compute Engine instances",
	Args:  cobra.NoArgs,
	RunE:  runInstancesList,
}

var (
	flagListFilter string
	flagListFormat string
)

// --- instances create ---

var instancesCreateCmd = &cobra.Command{
	Use:   "create INSTANCE_NAME",
	Short: "Create a Compute Engine instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesCreate,
}

var (
	flagMachineType    string
	flagNetwork        string
	flagSubnet         string
	flagImageFamily    string
	flagImageProject   string
	flagBootDiskSize   string
	flagBootDiskType   string
	flagTags           []string
	flagMetadata       map[string]string
	flagMetadataFromFile map[string]string
	flagServiceAccountEmail string
	flagScopes         []string
	flagNoAddress      bool
)

// --- instances delete ---

var instancesDeleteCmd = &cobra.Command{
	Use:   "delete INSTANCE_NAME",
	Short: "Delete a Compute Engine instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesDelete,
}

var flagDeleteQuiet bool

func init() {
	// describe
	instancesDescribeCmd.Flags().StringVar(&flagDescribeFormat, "format", "", "Output format (e.g. json, 'get(STATUS)')")
	instancesCmd.AddCommand(instancesDescribeCmd)

	// list
	instancesListCmd.Flags().StringVar(&flagListFilter, "filter", "", "Filter expression")
	instancesListCmd.Flags().StringVar(&flagListFormat, "format", "", "Output format (e.g. json, 'get(STATUS)')")
	instancesCmd.AddCommand(instancesListCmd)

	// create
	instancesCreateCmd.Flags().StringVar(&flagMachineType, "machine-type", "n1-standard-1", "Machine type")
	instancesCreateCmd.Flags().StringVar(&flagNetwork, "network", "", "Network")
	instancesCreateCmd.Flags().StringVar(&flagSubnet, "subnet", "", "Subnet")
	instancesCreateCmd.Flags().StringVar(&flagImageFamily, "image-family", "debian-12", "Image family")
	instancesCreateCmd.Flags().StringVar(&flagImageProject, "image-project", "debian-cloud", "Image project")
	instancesCreateCmd.Flags().StringVar(&flagBootDiskSize, "boot-disk-size", "", "Boot disk size (e.g. 10GB)")
	instancesCreateCmd.Flags().StringVar(&flagBootDiskType, "boot-disk-type", "", "Boot disk type (e.g. pd-standard)")
	instancesCreateCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "Network tags")
	instancesCreateCmd.Flags().StringToStringVar(&flagMetadata, "metadata", nil, "Instance metadata key=value pairs")
	instancesCreateCmd.Flags().StringToStringVar(&flagMetadataFromFile, "metadata-from-file", nil, "Instance metadata key=filepath pairs")
	instancesCreateCmd.Flags().StringVar(&flagServiceAccountEmail, "service-account", "", "Service account email")
	instancesCreateCmd.Flags().StringSliceVar(&flagScopes, "scopes", nil, "Service account scopes")
	instancesCreateCmd.Flags().BoolVar(&flagNoAddress, "no-address", false, "No external IP address")
	instancesCmd.AddCommand(instancesCreateCmd)

	// delete
	instancesDeleteCmd.Flags().BoolVar(&flagDeleteQuiet, "quiet", false, "Suppress confirmation prompt")
	instancesCmd.AddCommand(instancesDeleteCmd)
}

func runInstancesDescribe(cmd *cobra.Command, args []string) error {
	instance := args[0]
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}

	return formatOutput(inst, flagDescribeFormat)
}

func runInstancesList(cmd *cobra.Command, args []string) error {
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	call := svc.Instances.List(project, zone).Context(ctx)
	if flagListFilter != "" {
		call = call.Filter(flagListFilter)
	}

	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}

	if flagListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp.Items)
	}

	if isGetFormat(flagListFormat) {
		field := extractGetField(flagListFormat)
		for _, inst := range resp.Items {
			fmt.Println(getInstanceField(inst, field))
		}
		return nil
	}

	fmt.Printf("%-30s %-15s %-20s %-15s %-15s\n", "NAME", "ZONE", "MACHINE_TYPE", "INTERNAL_IP", "STATUS")
	for _, inst := range resp.Items {
		mt := path.Base(inst.MachineType)
		ip := getInternalIP(inst)
		z := path.Base(inst.Zone)
		fmt.Printf("%-30s %-15s %-20s %-15s %-15s\n", inst.Name, z, mt, ip, inst.Status)
	}
	return nil
}

func runInstancesCreate(cmd *cobra.Command, args []string) error {
	instance := args[0]
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	machineTypeURL := fmt.Sprintf("zones/%s/machineTypes/%s", zone, flagMachineType)
	sourceImage := fmt.Sprintf("projects/%s/global/images/family/%s", flagImageProject, flagImageFamily)

	inst := &compute.Instance{
		Name:        instance,
		MachineType: machineTypeURL,
		Disks: []*compute.AttachedDisk{
			{
				Boot:       true,
				AutoDelete: true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: sourceImage,
					DiskSizeGb:  parseDiskSize(flagBootDiskSize),
					DiskType:    formatDiskType(zone, flagBootDiskType),
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			buildNetworkInterface(project, flagNetwork, flagSubnet, flagNoAddress),
		},
	}

	if len(flagTags) > 0 {
		inst.Tags = &compute.Tags{Items: flagTags}
	}

	meta := buildMetadata(flagMetadata, flagMetadataFromFile)
	if meta != nil {
		inst.Metadata = meta
	}

	if flagServiceAccountEmail != "" || len(flagScopes) > 0 {
		sa := &compute.ServiceAccount{
			Email:  flagServiceAccountEmail,
			Scopes: flagScopes,
		}
		if sa.Email == "" {
			sa.Email = "default"
		}
		if len(sa.Scopes) == 0 {
			sa.Scopes = []string{"https://www.googleapis.com/auth/cloud-platform"}
		}
		inst.ServiceAccounts = []*compute.ServiceAccount{sa}
	}

	fmt.Printf("Creating instance [%s] in zone [%s]...\n", instance, zone)
	op, err := svc.Instances.Insert(project, zone, inst).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Created instance [%s].\n", instance)
	return nil
}

func runInstancesDelete(cmd *cobra.Command, args []string) error {
	instance := args[0]
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	if !flagDeleteQuiet {
		fmt.Printf("The following instances will be deleted: [%s]\n", instance)
		fmt.Print("Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "" && answer != "y" && answer != "Y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting instance [%s]...\n", instance)
	op, err := svc.Instances.Delete(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Deleted instance [%s].\n", instance)
	return nil
}

// --- Helpers ---

func resolveRegion() (string, string, error) {
	props, err := config.Load()
	if err != nil {
		return "", "", err
	}
	project := config.Resolve(flagProject, "CLOUDSDK_CORE_PROJECT", props.Core.Project)
	region := config.Resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
	if project == "" {
		return "", "", fmt.Errorf("project is required")
	}
	return project, region, nil
}

func formatOutput(v any, format string) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func isGetFormat(format string) bool {
	return len(format) > 5 && format[:4] == "get(" && format[len(format)-1] == ')'
}

func extractGetField(format string) string {
	return format[4 : len(format)-1]
}

func getInstanceField(inst *compute.Instance, field string) string {
	switch field {
	case "STATUS":
		return inst.Status
	case "name":
		return inst.Name
	case "networkInterfaces[0].networkIP":
		return getInternalIP(inst)
	case "networkInterfaces[0].accessConfigs[0].natIP":
		return getExternalIP(inst)
	default:
		return inst.Name
	}
}

func parseDiskSize(s string) int64 {
	if s == "" {
		return 0
	}
	var size int64
	fmt.Sscanf(s, "%dGB", &size)
	if size == 0 {
		fmt.Sscanf(s, "%d", &size)
	}
	return size
}

func formatDiskType(zone, diskType string) string {
	if diskType == "" {
		return ""
	}
	return fmt.Sprintf("zones/%s/diskTypes/%s", zone, diskType)
}

func buildNetworkInterface(project, network, subnet string, noAddress bool) *compute.NetworkInterface {
	ni := &compute.NetworkInterface{}
	if network != "" {
		ni.Network = fmt.Sprintf("projects/%s/global/networks/%s", project, network)
	}
	if subnet != "" {
		ni.Subnetwork = subnet
	}
	if !noAddress {
		ni.AccessConfigs = []*compute.AccessConfig{
			{Name: "External NAT", Type: "ONE_TO_ONE_NAT"},
		}
	}
	return ni
}

func buildMetadata(kv map[string]string, fromFile map[string]string) *compute.Metadata {
	var items []*compute.MetadataItems
	for k, v := range kv {
		val := v
		items = append(items, &compute.MetadataItems{Key: k, Value: &val})
	}
	for k, f := range fromFile {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		val := string(data)
		items = append(items, &compute.MetadataItems{Key: k, Value: &val})
	}
	if len(items) == 0 {
		return nil
	}
	return &compute.Metadata{Items: items}
}
