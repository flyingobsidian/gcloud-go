package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/flyingobsidian/gcloud-go/internal/config"
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
	flagListURI    bool
)

// --- instances create ---

var instancesCreateCmd = &cobra.Command{
	Use:   "create INSTANCE_NAME [INSTANCE_NAME ...]",
	Short: "Create a Compute Engine instance",
	Args:  cobra.MinimumNArgs(1),
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
	flagScopes           []string
	flagNoAddress            bool
	flagPreemptible          bool
	flagProvisioningModel    string
	flagCreateLabels         map[string]string
	flagCreateDeletionProtect bool
	flagCreateAsync          bool
	flagCreateDisk           string
	flagCreateNewDisk        string
	flagPrivateNetworkIP     string
	flagCreateAddress        string
	flagAccelerator          string
	flagShieldedSecureBoot   bool
	flagShieldedVTPM         bool
	flagShieldedIntegrity    bool
	flagCanIPForward         bool
	flagMinCPUPlatform       string
)

// --- instances delete ---

var instancesDeleteCmd = &cobra.Command{
	Use:   "delete INSTANCE_NAME [INSTANCE_NAME ...]",
	Short: "Delete a Compute Engine instance",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runInstancesDelete,
}

var (
	flagDeleteQuiet bool
	flagDeleteDisks string
	flagKeepDisks   string
)

func init() {
	// describe
	instancesDescribeCmd.Flags().StringVar(&flagDescribeFormat, "format", "", "Output format (e.g. json, 'get(STATUS)')")
	instancesCmd.AddCommand(instancesDescribeCmd)

	// list
	instancesListCmd.Flags().StringVar(&flagListFilter, "filter", "", "Filter expression")
	instancesListCmd.Flags().StringVar(&flagListFormat, "format", "", "Output format (e.g. json, 'get(STATUS)')")
	instancesListCmd.Flags().BoolVar(&flagListURI, "uri", false, "Print self-links (URIs) instead of the default table")
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
	instancesCreateCmd.Flags().BoolVar(&flagPreemptible, "preemptible", false, "Create a preemptible instance")
	instancesCreateCmd.Flags().StringVar(&flagProvisioningModel, "provisioning-model", "", "Provisioning model: STANDARD or SPOT")
	instancesCreateCmd.Flags().StringToStringVar(&flagCreateLabels, "labels", nil, "Labels key=value pairs")
	instancesCreateCmd.Flags().BoolVar(&flagCreateDeletionProtect, "deletion-protection", false, "Enable deletion protection")
	instancesCreateCmd.Flags().BoolVar(&flagCreateAsync, "async", false, "Return immediately without waiting")
	instancesCreateCmd.Flags().StringVar(&flagCreateDisk, "disk", "", "Attach an existing disk (name=NAME,mode=rw|ro,boot=yes|no)")
	instancesCreateCmd.Flags().StringVar(&flagCreateNewDisk, "create-disk", "", "Create and attach a new disk (name=NAME,size=SIZE,type=TYPE)")
	instancesCreateCmd.Flags().StringVar(&flagPrivateNetworkIP, "private-network-ip", "", "Private network IP address")
	instancesCreateCmd.Flags().StringVar(&flagCreateAddress, "address", "", "External IP address or name of a reserved address")
	instancesCreateCmd.Flags().StringVar(&flagAccelerator, "accelerator", "", "Accelerator spec (e.g. type=nvidia-tesla-t4,count=1)")
	instancesCreateCmd.Flags().BoolVar(&flagShieldedSecureBoot, "shielded-secure-boot", false, "Enable Secure Boot for shielded VM")
	instancesCreateCmd.Flags().BoolVar(&flagShieldedVTPM, "shielded-vtpm", true, "Enable vTPM for shielded VM")
	instancesCreateCmd.Flags().BoolVar(&flagShieldedIntegrity, "shielded-integrity-monitoring", true, "Enable integrity monitoring for shielded VM")
	instancesCreateCmd.Flags().BoolVar(&flagCanIPForward, "can-ip-forward", false, "Allow IP forwarding")
	instancesCreateCmd.Flags().StringVar(&flagMinCPUPlatform, "min-cpu-platform", "", "Minimum CPU platform")
	instancesCmd.AddCommand(instancesCreateCmd)

	// delete
	instancesDeleteCmd.Flags().BoolVar(&flagDeleteQuiet, "quiet", false, "Suppress confirmation prompt")
	instancesDeleteCmd.Flags().StringVar(&flagDeleteDisks, "delete-disks", "", "Disk types to delete: all, data, or boot")
	instancesDeleteCmd.Flags().StringVar(&flagKeepDisks, "keep-disks", "", "Disk types to keep: all, data, or boot")
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
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Resolve zone: if set, list in that zone; otherwise aggregate across all zones.
	zone := resolveZone()
	var instances []*compute.Instance
	if zone != "" {
		pageToken := ""
		for {
			call := svc.Instances.List(project, zone).Context(ctx)
			if flagListFilter != "" {
				call = call.Filter(flagListFilter)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing instances: %w", err)
			}
			instances = append(instances, resp.Items...)
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
	} else {
		pageToken := ""
		for {
			call := svc.Instances.AggregatedList(project).Context(ctx)
			if flagListFilter != "" {
				call = call.Filter(flagListFilter)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing instances: %w", err)
			}
			for _, scoped := range resp.Items {
				instances = append(instances, scoped.Instances...)
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
	}

	return formatInstanceList(instances)
}

func runInstancesCreate(cmd *cobra.Command, args []string) error {
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	for _, instance := range args {
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

		meta, err := buildMetadata(flagMetadata, flagMetadataFromFile)
		if err != nil {
			return err
		}
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

		if flagPreemptible || flagProvisioningModel != "" {
			if inst.Scheduling == nil {
				inst.Scheduling = &compute.Scheduling{}
			}
			if flagPreemptible {
				inst.Scheduling.Preemptible = true
			}
			if flagProvisioningModel != "" {
				inst.Scheduling.ProvisioningModel = flagProvisioningModel
			}
		}

		if len(flagCreateLabels) > 0 {
			inst.Labels = flagCreateLabels
		}

		if flagCreateDeletionProtect {
			inst.DeletionProtection = true
		}

		if flagCanIPForward {
			inst.CanIpForward = true
		}

		if flagMinCPUPlatform != "" {
			inst.MinCpuPlatform = flagMinCPUPlatform
		}

		if flagShieldedSecureBoot || flagShieldedVTPM || flagShieldedIntegrity {
			inst.ShieldedInstanceConfig = &compute.ShieldedInstanceConfig{
				EnableSecureBoot:          flagShieldedSecureBoot,
				EnableVtpm:                flagShieldedVTPM,
				EnableIntegrityMonitoring: flagShieldedIntegrity,
			}
		}

		if flagPrivateNetworkIP != "" && len(inst.NetworkInterfaces) > 0 {
			inst.NetworkInterfaces[0].NetworkIP = flagPrivateNetworkIP
		}

		if flagCreateAddress != "" && len(inst.NetworkInterfaces) > 0 {
			if len(inst.NetworkInterfaces[0].AccessConfigs) > 0 {
				inst.NetworkInterfaces[0].AccessConfigs[0].NatIP = flagCreateAddress
			}
		}

		if flagAccelerator != "" {
			accel := parseAccelerator(flagAccelerator, project, zone)
			if accel != nil {
				inst.GuestAccelerators = []*compute.AcceleratorConfig{accel}
				if inst.Scheduling == nil {
					inst.Scheduling = &compute.Scheduling{}
				}
				inst.Scheduling.OnHostMaintenance = "TERMINATE"
			}
		}

		if flagCreateDisk != "" {
			diskSpec := parseKeyValueSpec(flagCreateDisk)
			diskName := diskSpec["name"]
			if diskName == "" {
				return fmt.Errorf("--disk requires name= parameter")
			}
			diskURL := fmt.Sprintf("projects/%s/zones/%s/disks/%s", project, zone, diskName)
			mode := "READ_WRITE"
			if diskSpec["mode"] == "ro" {
				mode = "READ_ONLY"
			}
			boot := diskSpec["boot"] == "yes"
			inst.Disks = append(inst.Disks, &compute.AttachedDisk{
				Source: diskURL,
				Mode:   mode,
				Boot:   boot,
			})
		}

		if flagCreateNewDisk != "" {
			diskSpec := parseKeyValueSpec(flagCreateNewDisk)
			diskName := diskSpec["name"]
			params := &compute.AttachedDiskInitializeParams{
				DiskName:   diskName,
				DiskSizeGb: parseDiskSize(diskSpec["size"]),
				DiskType:   formatDiskType(zone, diskSpec["type"]),
			}
			inst.Disks = append(inst.Disks, &compute.AttachedDisk{
				AutoDelete:       true,
				InitializeParams: params,
			})
		}

		fmt.Printf("Creating instance [%s] in zone [%s]...\n", instance, zone)
		op, err := svc.Instances.Insert(project, zone, inst).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating instance %s: %w", instance, err)
		}

		if flagCreateAsync {
			fmt.Printf("Create operation started for [%s]: %s\n", instance, op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Printf("Created instance [%s].\n", instance)
	}
	return nil
}

func parseAccelerator(spec, project, zone string) *compute.AcceleratorConfig {
	parts := parseKeyValueSpec(spec)
	accelType := parts["type"]
	if accelType == "" {
		return nil
	}
	count := int64(1)
	if c := parts["count"]; c != "" {
		fmt.Sscanf(c, "%d", &count)
	}
	return &compute.AcceleratorConfig{
		AcceleratorType:  fmt.Sprintf("projects/%s/zones/%s/acceleratorTypes/%s", project, zone, accelType),
		AcceleratorCount: count,
	}
}

func parseKeyValueSpec(spec string) map[string]string {
	result := make(map[string]string)
	for _, part := range strings.Split(spec, ",") {
		k, v, ok := strings.Cut(part, "=")
		if ok {
			result[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return result
}

func runInstancesDelete(cmd *cobra.Command, args []string) error {
	if flagDeleteDisks != "" && flagKeepDisks != "" {
		return fmt.Errorf("--delete-disks and --keep-disks are mutually exclusive")
	}
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	if !flagDeleteQuiet {
		fmt.Printf("The following instances will be deleted: %v\n", args)
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

	for _, instance := range args {
		if flagDeleteDisks != "" || flagKeepDisks != "" {
			inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("getting instance %s: %w", instance, err)
			}
			for _, d := range inst.Disks {
				want := d.AutoDelete
				if shouldModifyDisk(d, flagDeleteDisks, true) {
					want = true
				}
				if shouldModifyDisk(d, flagKeepDisks, false) {
					want = false
				}
				if want != d.AutoDelete {
					op, err := svc.Instances.SetDiskAutoDelete(project, zone, instance, want, d.DeviceName).Context(ctx).Do()
					if err != nil {
						return fmt.Errorf("setting auto-delete on disk %s: %w", d.DeviceName, err)
					}
					if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
						return err
					}
				}
			}
		}

		fmt.Printf("Deleting instance [%s]...\n", instance)
		op, err := svc.Instances.Delete(project, zone, instance).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting instance %s: %w", instance, err)
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Printf("Deleted instance [%s].\n", instance)
	}
	return nil
}

func resolveZone() string {
	if flagZone != "" {
		return flagZone
	}
	props, _ := config.Load()
	if props != nil {
		return config.Resolve("", "CLOUDSDK_COMPUTE_ZONE", props.Compute.Zone)
	}
	return ""
}

func formatInstanceList(instances []*compute.Instance) error {
	if flagListURI {
		for _, inst := range instances {
			fmt.Println(inst.SelfLink)
		}
		return nil
	}

	if flagListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(instances)
	}

	if isGetFormat(flagListFormat) {
		field := extractGetField(flagListFormat)
		for _, inst := range instances {
			fmt.Println(getInstanceField(inst, field))
		}
		return nil
	}

	if isCsvFormat(flagListFormat) {
		fields := extractCsvFields(flagListFormat)
		fmt.Println(strings.Join(fields, ","))
		for _, inst := range instances {
			var vals []string
			for _, f := range fields {
				vals = append(vals, getInstanceField(inst, f))
			}
			fmt.Println(strings.Join(vals, ","))
		}
		return nil
	}

	fmt.Printf("%-30s %-15s %-20s %-12s %-16s %-16s %-10s\n", "NAME", "ZONE", "MACHINE_TYPE", "PREEMPTIBLE", "INTERNAL_IP", "EXTERNAL_IP", "STATUS")
	for _, inst := range instances {
		mt := path.Base(inst.MachineType)
		intIP := getInternalIP(inst)
		extIP := getExternalIP(inst)
		z := path.Base(inst.Zone)
		preempt := ""
		if inst.Scheduling != nil && inst.Scheduling.Preemptible {
			preempt = "true"
		}
		fmt.Printf("%-30s %-15s %-20s %-12s %-16s %-16s %-10s\n", inst.Name, z, mt, preempt, intIP, extIP, inst.Status)
	}
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
	if isGetFormat(format) {
		field := extractGetField(format)
		if inst, ok := v.(*compute.Instance); ok {
			fmt.Println(getInstanceField(inst, field))
			return nil
		}
	}
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

func isCsvFormat(format string) bool {
	return len(format) > 5 && strings.HasPrefix(format, "csv(") && format[len(format)-1] == ')'
}

func extractCsvFields(format string) []string {
	inner := format[4 : len(format)-1]
	var fields []string
	for _, f := range strings.Split(inner, ",") {
		fields = append(fields, strings.TrimSpace(f))
	}
	return fields
}

func getInstanceField(inst *compute.Instance, field string) string {
	switch strings.ToUpper(field) {
	case "STATUS":
		return inst.Status
	case "NAME":
		return inst.Name
	case "ZONE":
		return path.Base(inst.Zone)
	case "MACHINE_TYPE":
		return path.Base(inst.MachineType)
	case "INTERNAL_IP":
		return getInternalIP(inst)
	case "EXTERNAL_IP":
		return getExternalIP(inst)
	}
	// Handle dotted-path field access (case-sensitive).
	switch field {
	case "networkInterfaces[0].networkIP":
		return getInternalIP(inst)
	case "networkInterfaces[0].accessConfigs[0].natIP":
		return getExternalIP(inst)
	default:
		return ""
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

func shouldModifyDisk(d *compute.AttachedDisk, mode string, _ bool) bool {
	switch mode {
	case "all":
		return true
	case "boot":
		return d.Boot
	case "data":
		return !d.Boot
	}
	return false
}

func buildMetadata(kv map[string]string, fromFile map[string]string) (*compute.Metadata, error) {
	var items []*compute.MetadataItems
	for k, v := range kv {
		val := v
		items = append(items, &compute.MetadataItems{Key: k, Value: &val})
	}
	for k, f := range fromFile {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("reading metadata file for key %q: %w", k, err)
		}
		val := string(data)
		items = append(items, &compute.MetadataItems{Key: k, Value: &val})
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &compute.Metadata{Items: items}, nil
}
