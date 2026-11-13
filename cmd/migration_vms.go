package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	vmmigration "google.golang.org/api/vmmigration/v1"
)

// --- gcloud migration vms target-projects and image-imports (#895, #896) ---

func (s *migrationCmds) mvmService(ctx context.Context) (*vmmigration.Service, error) {
	return s.inv.clients().VMMigration(ctx)
}

// mvmLocationParent returns projects/PROJECT/locations/LOCATION.
func mvmLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func mvmResourceName(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func (s *migrationCmds) mvmResolveLocationParent(defaultToGlobal bool) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	loc := s.flagMVMLocation
	if loc == "" && defaultToGlobal {
		loc = "global"
	}
	if loc == "" {
		return "", fmt.Errorf("--location is required")
	}
	return mvmLocationParent(project, loc), nil
}

func mvmWaitOp(ctx context.Context, svc *vmmigration.Service, op *vmmigration.Operation) (*vmmigration.Operation, error) {
	for !op.Done {
		time.Sleep(3 * time.Second)
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (s *migrationCmds) mvmFinishOp(ctx context.Context, svc *vmmigration.Service, op *vmmigration.Operation, verb, target string) error {
	if s.flagMVMAsync {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagMVMFormat)
	}
	final, err := mvmWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, target)
	return emitFormattedTo(s.inv.stdout, final, s.flagMVMFormat)
}

// --- target-projects ---

// --- image-imports ---

func (s *migrationCmds) initMigrationVms() {
	addLoc := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagMVMLocation, "location", "", "Location (region); defaults to 'global' for target-projects")
		}
	}
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagMVMFormat, "format", "", "Output format")
		}
	}
	addFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagMVMFilter, "filter", "", "Server-side list filter")
		}
	}

	// target-projects
	addLoc(s.mvmTPAddCmd, s.mvmTPDeleteCmd, s.mvmTPDescribeCmd, s.mvmTPListCmd, s.mvmTPUpdateCmd)
	addFmt(s.mvmTPAddCmd, s.mvmTPDescribeCmd, s.mvmTPListCmd, s.mvmTPUpdateCmd)
	addFilter(s.mvmTPListCmd)
	s.mvmTPAddCmd.Flags().StringVar(&s.flagMVMTargetProj, "target-project", "", "The Compute Engine project ID or number (required)")
	_ = s.mvmTPAddCmd.MarkFlagRequired("target-project")
	s.mvmTPAddCmd.Flags().StringVar(&s.flagMVMDescription, "description", "", "Optional description")
	s.mvmTPAddCmd.Flags().BoolVar(&s.flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	s.mvmTPDeleteCmd.Flags().BoolVar(&s.flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	s.mvmTPUpdateCmd.Flags().StringVar(&s.flagMVMTargetProj, "target-project", "", "Update the Compute Engine project ID")
	s.mvmTPUpdateCmd.Flags().StringVar(&s.flagMVMDescription, "description", "", "Update the description")
	s.mvmTPUpdateCmd.Flags().StringVar(&s.flagMVMUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	s.mvmTPUpdateCmd.Flags().BoolVar(&s.flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	s.mvmTargetProjectsCmd.AddCommand(s.mvmTPAddCmd, s.mvmTPDeleteCmd, s.mvmTPDescribeCmd, s.mvmTPListCmd, s.mvmTPUpdateCmd)
	s.migrationVMsCmd.AddCommand(s.mvmTargetProjectsCmd)

	// image-imports
	addLoc(s.mvmIICreateCmd, s.mvmIIDeleteCmd, s.mvmIIDescribeCmd, s.mvmIIListCmd)
	addFmt(s.mvmIICreateCmd, s.mvmIIDescribeCmd, s.mvmIIListCmd)
	addFilter(s.mvmIIListCmd)
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMSourceFile, "source-file", "", "Cloud Storage URI (gs://bucket/file.vmdk) of the source image (required)")
	_ = s.mvmIICreateCmd.MarkFlagRequired("source-file")
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMImageName, "image-name", "", "Target Compute Engine image name (defaults to the import resource ID)")
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMImportTP, "target-project", "", "Target project resource path (projects/*/locations/global/targetProjects/*) (required)")
	_ = s.mvmIICreateCmd.MarkFlagRequired("target-project")
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMDescription, "description", "", "Optional description for the target image")
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMFamily, "family-name", "", "Compute Engine image family")
	s.mvmIICreateCmd.Flags().StringSliceVar(&s.flagMVMLabels, "labels", nil, "KEY=VALUE labels on the target image")
	s.mvmIICreateCmd.Flags().StringSliceVar(&s.flagMVMAddLicenses, "additional-licenses", nil, "Additional Compute Engine license URIs")
	s.mvmIICreateCmd.Flags().BoolVar(&s.flagMVMSingleRegion, "single-region-storage", false, "Store image in a single region")
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMKmsKey, "kms-key", "", "Cloud KMS key used to encrypt the image import (projects/.../cryptoKeys/...)")
	s.mvmIICreateCmd.Flags().BoolVar(&s.flagMVMSkipOSAdapt, "skip-os-adaptation", false, "Skip OS adaptation")
	s.mvmIICreateCmd.Flags().BoolVar(&s.flagMVMGeneralize, "generalize", false, "Generalize the image (Windows)")
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMLicenseType, "license-type", "", "License type (default, payg, byol)")
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMBootConv, "boot-conversion", "", "Boot conversion (none, bios-to-efi)")
	s.mvmIICreateCmd.Flags().StringVar(&s.flagMVMConfigFile, "config-file", "", "JSON/YAML ImageImport body override")
	s.mvmIICreateCmd.Flags().BoolVar(&s.flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	s.mvmIIDeleteCmd.Flags().BoolVar(&s.flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	s.mvmImageImportsCmd.AddCommand(s.mvmIICreateCmd, s.mvmIIDeleteCmd, s.mvmIIDescribeCmd, s.mvmIIListCmd)
	s.migrationVMsCmd.AddCommand(s.mvmImageImportsCmd)
}

// --- target-projects impl ---

func (s *migrationCmds) runMVMTPCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.TargetProject{
		Project:     s.flagMVMTargetProj,
		Description: s.flagMVMDescription,
	}
	op, err := svc.Projects.Locations.TargetProjects.Create(parent, body).TargetProjectId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating target project: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Create target project", args[0])
}

func (s *migrationCmds) runMVMTPDelete(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "targetProjects", args[0])
	op, err := svc.Projects.Locations.TargetProjects.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting target project: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Delete target project", args[0])
}

func (s *migrationCmds) runMVMTPDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "targetProjects", args[0])
	got, err := svc.Projects.Locations.TargetProjects.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing target project: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMVMFormat)
}

func (s *migrationCmds) runMVMTPList(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	var all []*vmmigration.TargetProject
	pageToken := ""
	for {
		call := svc.Projects.Locations.TargetProjects.List(parent).Context(ctx)
		if s.flagMVMFilter != "" {
			call = call.Filter(s.flagMVMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing target projects: %w", err)
		}
		all = append(all, resp.TargetProjects...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagMVMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagMVMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-30s %s\n", "NAME", "PROJECT", "CREATE_TIME")
	for _, tp := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-30s %s\n", path.Base(tp.Name), tp.Project, tp.CreateTime)
	}
	return nil
}

func (s *migrationCmds) runMVMTPUpdate(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.TargetProject{}
	var mask []string
	if s.flagMVMTargetProj != "" {
		body.Project = s.flagMVMTargetProj
		mask = append(mask, "project")
	}
	if s.flagMVMDescription != "" {
		body.Description = s.flagMVMDescription
		mask = append(mask, "description")
	}
	if s.flagMVMUpdateMask != "" {
		mask = strings.Split(s.flagMVMUpdateMask, ",")
	}
	name := mvmResourceName(parent, "targetProjects", args[0])
	call := svc.Projects.Locations.TargetProjects.Patch(name, body).Context(ctx)
	if len(mask) > 0 {
		call = call.UpdateMask(strings.Join(mask, ","))
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating target project: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Update target project", args[0])
}

// --- image-imports impl ---

func (s *migrationCmds) runMVMIICreate(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.ImageImport{
		CloudStorageUri: s.flagMVMSourceFile,
	}
	if s.flagMVMConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagMVMConfigFile, body); err != nil {
			return err
		}
	}
	if body.DiskImageTargetDefaults == nil {
		body.DiskImageTargetDefaults = &vmmigration.DiskImageTargetDetails{}
	}
	target := body.DiskImageTargetDefaults
	if s.flagMVMImageName != "" {
		target.ImageName = s.flagMVMImageName
	} else if target.ImageName == "" {
		target.ImageName = args[0]
	}
	target.TargetProject = s.flagMVMImportTP
	if s.flagMVMDescription != "" {
		target.Description = s.flagMVMDescription
	}
	if s.flagMVMFamily != "" {
		target.FamilyName = s.flagMVMFamily
	}
	if len(s.flagMVMLabels) > 0 {
		if target.Labels == nil {
			target.Labels = map[string]string{}
		}
		for _, kv := range s.flagMVMLabels {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return fmt.Errorf("invalid --labels entry %q (want KEY=VALUE)", kv)
			}
			target.Labels[k] = v
		}
	}
	if len(s.flagMVMAddLicenses) > 0 {
		target.AdditionalLicenses = append(target.AdditionalLicenses, s.flagMVMAddLicenses...)
	}
	target.SingleRegionStorage = s.flagMVMSingleRegion
	if s.flagMVMKmsKey != "" {
		body.Encryption = &vmmigration.Encryption{KmsKey: s.flagMVMKmsKey}
	}
	if s.flagMVMSkipOSAdapt {
		target.DataDiskImageImport = &vmmigration.DataDiskImageImport{}
		target.OsAdaptationParameters = nil
	} else if s.flagMVMGeneralize || s.flagMVMLicenseType != "" || s.flagMVMBootConv != "" {
		if target.OsAdaptationParameters == nil {
			target.OsAdaptationParameters = &vmmigration.ImageImportOsAdaptationParameters{}
		}
		p := target.OsAdaptationParameters
		p.Generalize = s.flagMVMGeneralize
		switch strings.ToLower(s.flagMVMLicenseType) {
		case "":
		case "default":
			p.LicenseType = "COMPUTE_ENGINE_LICENSE_TYPE_DEFAULT"
		case "payg":
			p.LicenseType = "COMPUTE_ENGINE_LICENSE_TYPE_PAYG"
		case "byol":
			p.LicenseType = "COMPUTE_ENGINE_LICENSE_TYPE_BYOL"
		default:
			return fmt.Errorf("invalid --license-type %q (want default, payg, or byol)", s.flagMVMLicenseType)
		}
		switch strings.ToLower(s.flagMVMBootConv) {
		case "":
		case "none":
			p.BootConversion = "NONE"
		case "bios-to-efi":
			p.BootConversion = "BIOS_TO_EFI"
		default:
			return fmt.Errorf("invalid --boot-conversion %q (want none or bios-to-efi)", s.flagMVMBootConv)
		}
	}
	op, err := svc.Projects.Locations.ImageImports.Create(parent, body).ImageImportId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating image import: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Create image import", args[0])
}

func (s *migrationCmds) runMVMIIDelete(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "imageImports", args[0])
	op, err := svc.Projects.Locations.ImageImports.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting image import: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Delete image import", args[0])
}

func (s *migrationCmds) runMVMIIDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "imageImports", args[0])
	got, err := svc.Projects.Locations.ImageImports.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing image import: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMVMFormat)
}

func (s *migrationCmds) runMVMIIList(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	var all []*vmmigration.ImageImport
	pageToken := ""
	for {
		call := svc.Projects.Locations.ImageImports.List(parent).Context(ctx)
		if s.flagMVMFilter != "" {
			call = call.Filter(s.flagMVMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing image imports: %w", err)
		}
		all = append(all, resp.ImageImports...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagMVMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagMVMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-60s %s\n", "NAME", "CLOUD_STORAGE_URI", "CREATE_TIME")
	for _, ii := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-60s %s\n", path.Base(ii.Name), ii.CloudStorageUri, ii.CreateTime)
	}
	return nil
}
