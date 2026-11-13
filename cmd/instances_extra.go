package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/flyingobsidian/gcloud-go/internal/gfilter"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- instances describe ---

// --- instances list ---

// --- instances create ---

// --- instances delete ---

func (s *computeCmds) initInstancesExtra() {
	// describe
	s.instancesDescribeCmd.Flags().StringVar(&s.flagDescribeFormat, "format", "", "Output format (e.g. json, 'get(STATUS)')")
	s.instancesCmd.AddCommand(s.instancesDescribeCmd)

	// list
	s.instancesListCmd.Flags().StringVar(&s.flagListFilter, "filter", "", "Filter expression")
	s.instancesListCmd.Flags().StringVar(&s.flagListFormat, "format", "", "Output format (e.g. json, 'get(STATUS)')")
	s.instancesListCmd.Flags().BoolVar(&s.flagListURI, "uri", false, "Print self-links (URIs) instead of the default table")
	s.instancesCmd.AddCommand(s.instancesListCmd)

	// create
	s.instancesCreateCmd.Flags().StringVar(&s.flagMachineType, "machine-type", "n1-standard-1", "Machine type")
	s.instancesCreateCmd.Flags().StringVar(&s.flagNetwork, "network", "", "Network")
	s.instancesCreateCmd.Flags().StringVar(&s.flagSubnet, "subnet", "", "Subnet")
	s.instancesCreateCmd.Flags().StringVar(&s.flagImageFamily, "image-family", "debian-12", "Image family")
	s.instancesCreateCmd.Flags().StringVar(&s.flagImageProject, "image-project", "debian-cloud", "Image project")
	s.instancesCreateCmd.Flags().StringVar(&s.flagBootDiskSize, "boot-disk-size", "", "Boot disk size (e.g. 10GB)")
	s.instancesCreateCmd.Flags().StringVar(&s.flagBootDiskType, "boot-disk-type", "", "Boot disk type (e.g. pd-standard)")
	s.instancesCreateCmd.Flags().StringSliceVar(&s.flagTags, "tags", nil, "Network tags")
	s.instancesCreateCmd.Flags().StringToStringVar(&s.flagMetadata, "metadata", nil, "Instance metadata key=value pairs")
	s.instancesCreateCmd.Flags().StringToStringVar(&s.flagMetadataFromFile, "metadata-from-file", nil, "Instance metadata key=filepath pairs")
	s.instancesCreateCmd.Flags().StringVar(&s.flagServiceAccountEmail, "service-account", "", "Service account email")
	s.instancesCreateCmd.Flags().StringSliceVar(&s.flagScopes, "scopes", nil, "Service account scopes")
	s.instancesCreateCmd.Flags().BoolVar(&s.flagNoAddress, "no-address", false, "No external IP address")
	s.instancesCreateCmd.Flags().BoolVar(&s.flagPreemptible, "preemptible", false, "Create a preemptible instance")
	s.instancesCreateCmd.Flags().StringVar(&s.flagProvisioningModel, "provisioning-model", "", "Provisioning model: STANDARD or SPOT")
	s.instancesCreateCmd.Flags().StringToStringVar(&s.flagCreateLabels, "labels", nil, "Labels key=value pairs")
	s.instancesCreateCmd.Flags().BoolVar(&s.flagCreateDeletionProtect, "deletion-protection", false, "Enable deletion protection")
	s.instancesCreateCmd.Flags().BoolVar(&s.flagCreateAsync, "async", false, "Return immediately without waiting")
	s.instancesCreateCmd.Flags().StringVar(&s.flagCreateDisk, "disk", "", "Attach an existing disk (name=NAME,mode=rw|ro,boot=yes|no)")
	s.instancesCreateCmd.Flags().StringVar(&s.flagCreateNewDisk, "create-disk", "", "Create and attach a new disk (name=NAME,size=SIZE,type=TYPE)")
	s.instancesCreateCmd.Flags().StringVar(&s.flagPrivateNetworkIP, "private-network-ip", "", "Private network IP address")
	s.instancesCreateCmd.Flags().StringVar(&s.flagCreateAddress, "address", "", "External IP address or name of a reserved address")
	s.instancesCreateCmd.Flags().StringVar(&s.flagAccelerator, "accelerator", "", "Accelerator spec (e.g. type=nvidia-tesla-t4,count=1)")
	s.instancesCreateCmd.Flags().BoolVar(&s.flagShieldedSecureBoot, "shielded-secure-boot", false, "Enable Secure Boot for shielded VM")
	s.instancesCreateCmd.Flags().BoolVar(&s.flagShieldedVTPM, "shielded-vtpm", true, "Enable vTPM for shielded VM")
	s.instancesCreateCmd.Flags().BoolVar(&s.flagShieldedIntegrity, "shielded-integrity-monitoring", true, "Enable integrity monitoring for shielded VM")
	s.instancesCreateCmd.Flags().BoolVar(&s.flagCanIPForward, "can-ip-forward", false, "Allow IP forwarding")
	s.instancesCreateCmd.Flags().StringVar(&s.flagMinCPUPlatform, "min-cpu-platform", "", "Minimum CPU platform")
	s.instancesCmd.AddCommand(s.instancesCreateCmd)

	// delete
	// --quiet is provided by the global persistent flag
	s.instancesDeleteCmd.Flags().StringVar(&s.flagDeleteDisks, "delete-disks", "", "Disk types to delete: all, data, or boot")
	s.instancesDeleteCmd.Flags().StringVar(&s.flagKeepDisks, "keep-disks", "", "Disk types to keep: all, data, or boot")
	s.instancesDeleteCmd.Flags().BoolVar(&s.flagDeleteAsync, "async", false, "Return immediately without waiting")
	s.instancesCmd.AddCommand(s.instancesDeleteCmd)
}

func (s *computeCmds) runInstancesDescribe(cmd *cobra.Command, args []string) error {
	instance := args[0]
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}

	return s.inv.formatOutput(inst, s.flagDescribeFormat)
}

func (s *computeCmds) runInstancesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	// --filter is a gcloud client-side filter (see `gcloud topic filters`) —
	// distinct from the Compute API's server-side filter (`field eq value`
	// syntax). Passing gcloud syntax like `foo:bar` to the server returns
	// no results (issue #1725). Compile up-front and apply after fetching.
	filter, err := gfilter.Compile(s.flagListFilter)
	if err != nil {
		return fmt.Errorf("invalid --filter expression: %w", err)
	}
	if filter.UsesContainsOperator() {
		fmt.Fprintf(s.inv.stderr,
			"WARNING: --filter : operator evaluation is changing for consistency across Google APIs.  %s currently matches but will not match in the near future.  Run `gcloud topic filters` for details.\n",
			s.flagListFilter)
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	// Resolve zone: if set, list in that zone; otherwise aggregate across all zones.
	zone := s.resolveZone()
	var instances []*compute.Instance
	if zone != "" {
		pageToken := ""
		for {
			call := svc.Instances.List(project, zone).Context(ctx)
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

	// Apply the client-side --filter to the aggregated list.
	if s.flagListFilter != "" {
		kept := instances[:0]
		for _, inst := range instances {
			if filter.Match(inst) {
				kept = append(kept, inst)
			}
		}
		instances = kept
	}

	return s.formatInstanceList(instances)
}

func (s *computeCmds) runInstancesCreate(cmd *cobra.Command, args []string) error {
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	for _, instance := range args {
		machineTypeURL := fmt.Sprintf("zones/%s/machineTypes/%s", zone, s.flagMachineType)
		sourceImage := fmt.Sprintf("projects/%s/global/images/family/%s", s.flagImageProject, s.flagImageFamily)

		inst := &compute.Instance{
			Name:        instance,
			MachineType: machineTypeURL,
			Disks: []*compute.AttachedDisk{
				{
					Boot:       true,
					AutoDelete: true,
					InitializeParams: &compute.AttachedDiskInitializeParams{
						SourceImage: sourceImage,
						DiskSizeGb:  parseDiskSize(s.flagBootDiskSize),
						DiskType:    formatDiskType(zone, s.flagBootDiskType),
					},
				},
			},
			NetworkInterfaces: []*compute.NetworkInterface{
				buildNetworkInterface(project, s.flagNetwork, s.flagSubnet, s.flagNoAddress),
			},
		}

		if len(s.flagTags) > 0 {
			inst.Tags = &compute.Tags{Items: s.flagTags}
		}

		meta, err := buildMetadata(s.flagMetadata, s.flagMetadataFromFile)
		if err != nil {
			return err
		}
		if meta != nil {
			inst.Metadata = meta
		}

		if s.flagServiceAccountEmail != "" || len(s.flagScopes) > 0 {
			sa := &compute.ServiceAccount{
				Email:  s.flagServiceAccountEmail,
				Scopes: s.flagScopes,
			}
			if sa.Email == "" {
				sa.Email = "default"
			}
			if len(sa.Scopes) == 0 {
				sa.Scopes = []string{"https://www.googleapis.com/auth/cloud-platform"}
			}
			inst.ServiceAccounts = []*compute.ServiceAccount{sa}
		}

		if s.flagPreemptible || s.flagProvisioningModel != "" {
			if inst.Scheduling == nil {
				inst.Scheduling = &compute.Scheduling{}
			}
			if s.flagPreemptible {
				inst.Scheduling.Preemptible = true
			}
			if s.flagProvisioningModel != "" {
				inst.Scheduling.ProvisioningModel = s.flagProvisioningModel
			}
		}

		if len(s.flagCreateLabels) > 0 {
			inst.Labels = s.flagCreateLabels
		}

		if s.flagCreateDeletionProtect {
			inst.DeletionProtection = true
		}

		if s.flagCanIPForward {
			inst.CanIpForward = true
		}

		if s.flagMinCPUPlatform != "" {
			inst.MinCpuPlatform = s.flagMinCPUPlatform
		}

		if s.flagShieldedSecureBoot || s.flagShieldedVTPM || s.flagShieldedIntegrity {
			inst.ShieldedInstanceConfig = &compute.ShieldedInstanceConfig{
				EnableSecureBoot:          s.flagShieldedSecureBoot,
				EnableVtpm:                s.flagShieldedVTPM,
				EnableIntegrityMonitoring: s.flagShieldedIntegrity,
			}
		}

		if s.flagPrivateNetworkIP != "" && len(inst.NetworkInterfaces) > 0 {
			inst.NetworkInterfaces[0].NetworkIP = s.flagPrivateNetworkIP
		}

		if s.flagCreateAddress != "" && len(inst.NetworkInterfaces) > 0 {
			if len(inst.NetworkInterfaces[0].AccessConfigs) > 0 {
				inst.NetworkInterfaces[0].AccessConfigs[0].NatIP = s.flagCreateAddress
			}
		}

		if s.flagAccelerator != "" {
			accel := parseAccelerator(s.flagAccelerator, project, zone)
			if accel != nil {
				inst.GuestAccelerators = []*compute.AcceleratorConfig{accel}
				if inst.Scheduling == nil {
					inst.Scheduling = &compute.Scheduling{}
				}
				inst.Scheduling.OnHostMaintenance = "TERMINATE"
			}
		}

		if s.flagCreateDisk != "" {
			diskSpec := parseKeyValueSpec(s.flagCreateDisk)
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

		if s.flagCreateNewDisk != "" {
			diskSpec := parseKeyValueSpec(s.flagCreateNewDisk)
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

		fmt.Fprintf(s.inv.stdout, "Creating instance [%s] in zone [%s]...\n", instance, zone)
		op, err := svc.Instances.Insert(project, zone, inst).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating instance %s: %w", instance, err)
		}

		if s.flagCreateAsync {
			fmt.Fprintf(s.inv.stdout, "Create operation started for [%s]: %s\n", instance, op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Fprintf(s.inv.stdout, "Created instance [%s].\n", instance)
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

func (s *computeCmds) runInstancesDelete(cmd *cobra.Command, args []string) error {
	if s.flagDeleteDisks != "" && s.flagKeepDisks != "" {
		return fmt.Errorf("--delete-disks and --keep-disks are mutually exclusive")
	}
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	if !s.inv.flags.quiet {
		fmt.Fprintf(s.inv.stdout, "The following instances will be deleted: %v\n", args)
		fmt.Fprint(s.inv.stdout, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Fscanln(s.inv.stdin, &answer)
		if answer != "" && answer != "y" && answer != "Y" && answer != "yes" {
			fmt.Fprintln(s.inv.stdout, "Aborted.")
			return nil
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	for _, instance := range args {
		if s.flagDeleteDisks != "" || s.flagKeepDisks != "" {
			inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("getting instance %s: %w", instance, err)
			}
			for _, d := range inst.Disks {
				want := d.AutoDelete
				if shouldModifyDisk(d, s.flagDeleteDisks, true) {
					want = true
				}
				if shouldModifyDisk(d, s.flagKeepDisks, false) {
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

		fmt.Fprintf(s.inv.stdout, "Deleting instance [%s]...\n", instance)
		op, err := svc.Instances.Delete(project, zone, instance).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting instance %s: %w", instance, err)
		}

		if s.flagDeleteAsync {
			fmt.Fprintf(s.inv.stdout, "Delete operation started for [%s]: %s\n", instance, op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Fprintf(s.inv.stdout, "Deleted instance [%s].\n", instance)
	}
	return nil
}

func (s *computeCmds) resolveZone() string {
	if s.inv.flags.zone != "" {
		return s.inv.flags.zone
	}
	props, _ := s.inv.properties()
	if props != nil {
		return s.inv.resolve("", "CLOUDSDK_COMPUTE_ZONE", props.Compute.Zone)
	}
	return ""
}

func (s *computeCmds) formatInstanceList(instances []*compute.Instance) error {
	if s.flagListURI {
		for _, inst := range instances {
			fmt.Fprintln(s.inv.stdout, inst.SelfLink)
		}
		return nil
	}

	if s.flagListFormat == "json" {
		return s.inv.printJSON(instances)
	}

	if isGetFormat(s.flagListFormat) {
		field := extractGetField(s.flagListFormat)
		for _, inst := range instances {
			fmt.Fprintln(s.inv.stdout, getInstanceField(inst, field))
		}
		return nil
	}

	if isCsvFormat(s.flagListFormat) {
		fields := extractCsvFields(s.flagListFormat)
		fmt.Fprintln(s.inv.stdout, strings.Join(fields, ","))
		for _, inst := range instances {
			var vals []string
			for _, f := range fields {
				vals = append(vals, getInstanceField(inst, f))
			}
			fmt.Fprintln(s.inv.stdout, strings.Join(vals, ","))
		}
		return nil
	}

	printTableHeadingTo(s.inv.stdout, len(instances), "%-30s %-15s %-20s %-12s %-16s %-16s %-10s\n", "NAME", "ZONE", "MACHINE_TYPE", "PREEMPTIBLE", "INTERNAL_IP", "EXTERNAL_IP", "STATUS")
	for _, inst := range instances {
		mt := path.Base(inst.MachineType)
		intIP := getInternalIP(inst)
		extIP := getExternalIP(inst)
		z := path.Base(inst.Zone)
		preempt := ""
		if inst.Scheduling != nil && inst.Scheduling.Preemptible {
			preempt = "true"
		}
		fmt.Fprintf(s.inv.stdout, "%-30s %-15s %-20s %-12s %-16s %-16s %-10s\n", inst.Name, z, mt, preempt, intIP, extIP, inst.Status)
	}
	return nil
}

// --- Helpers ---

// formatOutputTo renders v to w for the compute-instance command family.
// It keeps the compute.Instance get(FIELD) alias table wired to
// getInstanceField (which handles gcloud aliases like INTERNAL_IP), and
// defers all other formatting to emitFormattedTo so there is a single
// implementation of yaml/json/table/etc. across the codebase.
func formatOutputTo(w io.Writer, v any, format string) error {
	if isGetFormat(format) {
		field := extractGetField(format)
		if inst, ok := v.(*compute.Instance); ok {
			fmt.Fprintln(w, getInstanceField(inst, field))
			return nil
		}
	}
	if format == "" {
		format = "json"
	}
	return emitFormattedTo(w, v, format)
}

func isTableFormat(format string) bool {
	return len(format) > 7 && strings.HasPrefix(format, "table(") && format[len(format)-1] == ')'
}

func extractTableFields(format string) []string {
	inner := format[6 : len(format)-1]
	var fields []string
	for _, f := range strings.Split(inner, ",") {
		fields = append(fields, strings.TrimSpace(f))
	}
	return fields
}

func isValueFormat(format string) bool {
	return len(format) > 7 && strings.HasPrefix(format, "value(") && format[len(format)-1] == ')'
}

func extractValueFields(format string) []string {
	inner := format[6 : len(format)-1]
	var fields []string
	for _, f := range strings.Split(inner, ",") {
		fields = append(fields, strings.TrimSpace(f))
	}
	return fields
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
