package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	vmmigration "google.golang.org/api/vmmigration/v1"
)

// --- gcloud migration vms machine-image-imports, disk-migrations (#1714) ---
//
// machine-image-imports shares the same underlying REST resource as
// image-imports (ImageImport), distinguished by MachineImageTargetDefaults
// instead of DiskImageTargetDefaults on the create body.
//
// disk-migrations is a per-source nested resource
// (projects/*/locations/*/sources/*/diskMigrationJobs/*).

// --- machine-image-imports ---

// --- disk-migrations ---

func (s *migrationCmds) initMigrationVmsExtra() {
	addLoc := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagMVMLocation, "location", "", "Location (region)")
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
	addAsync := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().BoolVar(&s.flagMVMAsync, "async", false, "Do not wait for the operation to finish")
		}
	}

	// machine-image-imports
	addLoc(s.mvmMIICreateCmd, s.mvmMIIDeleteCmd, s.mvmMIIDescribeCmd, s.mvmMIIListCmd)
	addFmt(s.mvmMIICreateCmd, s.mvmMIIDescribeCmd, s.mvmMIIListCmd)
	addFilter(s.mvmMIIListCmd)
	addAsync(s.mvmMIICreateCmd, s.mvmMIIDeleteCmd)
	s.mvmMIICreateCmd.Flags().StringVar(&s.flagMVMMIISourceFile, "source-file", "", "Cloud Storage URI of the source machine image (required)")
	_ = s.mvmMIICreateCmd.MarkFlagRequired("source-file")
	s.mvmMIICreateCmd.Flags().StringVar(&s.flagMVMMIIName, "machine-image-name", "", "Target Compute Engine machine image name (defaults to import ID)")
	s.mvmMIICreateCmd.Flags().StringVar(&s.flagMVMMIITarget, "target-project", "", "Target project resource path (required)")
	_ = s.mvmMIICreateCmd.MarkFlagRequired("target-project")
	s.mvmMIICreateCmd.Flags().StringVar(&s.flagMVMConfigFile, "config-file", "", "JSON/YAML ImageImport body override")
	s.mvmMachineImageImportsCmd.AddCommand(s.mvmMIICreateCmd, s.mvmMIIDeleteCmd, s.mvmMIIDescribeCmd, s.mvmMIIListCmd)
	s.migrationVMsCmd.AddCommand(s.mvmMachineImageImportsCmd)

	// disk-migrations
	addLoc(s.mvmDMCreateCmd, s.mvmDMDeleteCmd, s.mvmDMDescribeCmd, s.mvmDMListCmd, s.mvmDMCancelCmd, s.mvmDMRunCmd, s.mvmDMUpdateCmd)
	addFmt(s.mvmDMCreateCmd, s.mvmDMDescribeCmd, s.mvmDMListCmd, s.mvmDMUpdateCmd)
	addFilter(s.mvmDMListCmd)
	addAsync(s.mvmDMCreateCmd, s.mvmDMDeleteCmd, s.mvmDMCancelCmd, s.mvmDMRunCmd, s.mvmDMUpdateCmd)
	for _, c := range []*cobra.Command{s.mvmDMCreateCmd, s.mvmDMDeleteCmd, s.mvmDMDescribeCmd, s.mvmDMListCmd, s.mvmDMCancelCmd, s.mvmDMRunCmd, s.mvmDMUpdateCmd} {
		c.Flags().StringVar(&s.flagMVMDMSource, "source", "", "Migration source resource path or ID (required)")
		_ = c.MarkFlagRequired("source")
	}
	s.mvmDMCreateCmd.Flags().StringVar(&s.flagMVMDMDescription, "description", "", "Optional description")
	s.mvmDMCreateCmd.Flags().StringVar(&s.flagMVMConfigFile, "config-file", "", "JSON/YAML DiskMigrationJob body override")
	s.mvmDMUpdateCmd.Flags().StringVar(&s.flagMVMUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	s.mvmDMUpdateCmd.Flags().StringVar(&s.flagMVMDMDescription, "description", "", "Update the description")
	s.mvmDMUpdateCmd.Flags().StringVar(&s.flagMVMConfigFile, "config-file", "", "JSON/YAML DiskMigrationJob body override")
	s.mvmDiskMigrationsCmd.AddCommand(s.mvmDMCreateCmd, s.mvmDMDeleteCmd, s.mvmDMDescribeCmd, s.mvmDMListCmd, s.mvmDMCancelCmd, s.mvmDMRunCmd, s.mvmDMUpdateCmd)
	s.migrationVMsCmd.AddCommand(s.mvmDiskMigrationsCmd)
}

// --- machine-image-imports impl ---

func (s *migrationCmds) runMVMMIICreate(cmd *cobra.Command, args []string) error {
	parent, err := s.mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.ImageImport{CloudStorageUri: s.flagMVMMIISourceFile}
	if s.flagMVMConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagMVMConfigFile, body); err != nil {
			return err
		}
	}
	if body.MachineImageTargetDefaults == nil {
		body.MachineImageTargetDefaults = &vmmigration.MachineImageTargetDetails{}
	}
	target := body.MachineImageTargetDefaults
	if s.flagMVMMIIName != "" {
		target.MachineImageName = s.flagMVMMIIName
	} else if target.MachineImageName == "" {
		target.MachineImageName = args[0]
	}
	target.TargetProject = s.flagMVMMIITarget
	op, err := svc.Projects.Locations.ImageImports.Create(parent, body).ImageImportId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating machine image import: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Create machine image import", args[0])
}

func (s *migrationCmds) runMVMMIIDelete(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("deleting machine image import: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Delete machine image import", args[0])
}

func (s *migrationCmds) runMVMMIIDescribe(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("describing machine image import: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMVMFormat)
}

func (s *migrationCmds) runMVMMIIList(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("listing machine image imports: %w", err)
		}
		for _, ii := range resp.ImageImports {
			if ii.MachineImageTargetDefaults != nil {
				all = append(all, ii)
			}
		}
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

// --- disk-migrations helpers ---

func (s *migrationCmds) mvmDMSourceParent() (string, error) {
	parent, err := s.mvmResolveLocationParent(false)
	if err != nil {
		return "", err
	}
	if s.flagMVMDMSource == "" {
		return "", fmt.Errorf("--source is required")
	}
	if strings.HasPrefix(s.flagMVMDMSource, "projects/") {
		return s.flagMVMDMSource, nil
	}
	return fmt.Sprintf("%s/sources/%s", parent, s.flagMVMDMSource), nil
}

// --- disk-migrations impl ---

func (s *migrationCmds) runMVMDMCreate(cmd *cobra.Command, args []string) error {
	sourceParent, err := s.mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.DiskMigrationJob{}
	if s.flagMVMConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagMVMConfigFile, body); err != nil {
			return err
		}
	}
	op, err := svc.Projects.Locations.Sources.DiskMigrationJobs.Create(sourceParent, body).DiskMigrationJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating disk migration job: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Create disk migration job", args[0])
}

func (s *migrationCmds) runMVMDMDelete(cmd *cobra.Command, args []string) error {
	sourceParent, err := s.mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/diskMigrationJobs/%s", sourceParent, args[0])
	if strings.HasPrefix(args[0], "projects/") {
		name = args[0]
	}
	op, err := svc.Projects.Locations.Sources.DiskMigrationJobs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting disk migration job: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Delete disk migration job", args[0])
}

func (s *migrationCmds) runMVMDMDescribe(cmd *cobra.Command, args []string) error {
	sourceParent, err := s.mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/diskMigrationJobs/%s", sourceParent, args[0])
	if strings.HasPrefix(args[0], "projects/") {
		name = args[0]
	}
	got, err := svc.Projects.Locations.Sources.DiskMigrationJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing disk migration job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMVMFormat)
}

func (s *migrationCmds) runMVMDMList(cmd *cobra.Command, args []string) error {
	sourceParent, err := s.mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	var all []*vmmigration.DiskMigrationJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.Sources.DiskMigrationJobs.List(sourceParent).Context(ctx)
		if s.flagMVMFilter != "" {
			call = call.Filter(s.flagMVMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing disk migration jobs: %w", err)
		}
		all = append(all, resp.DiskMigrationJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagMVMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagMVMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-40s %s\n", "NAME", "STATE", "CREATE_TIME")
	for _, j := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-40s %s\n", path.Base(j.Name), j.State, j.CreateTime)
	}
	return nil
}

func (s *migrationCmds) runMVMDMCancel(cmd *cobra.Command, args []string) error {
	sourceParent, err := s.mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/diskMigrationJobs/%s", sourceParent, args[0])
	if strings.HasPrefix(args[0], "projects/") {
		name = args[0]
	}
	op, err := svc.Projects.Locations.Sources.DiskMigrationJobs.Cancel(name, &vmmigration.CancelDiskMigrationJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("canceling disk migration job: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Cancel disk migration job", args[0])
}

func (s *migrationCmds) runMVMDMRun(cmd *cobra.Command, args []string) error {
	sourceParent, err := s.mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/diskMigrationJobs/%s", sourceParent, args[0])
	if strings.HasPrefix(args[0], "projects/") {
		name = args[0]
	}
	op, err := svc.Projects.Locations.Sources.DiskMigrationJobs.Run(name, &vmmigration.RunDiskMigrationJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running disk migration job: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Run disk migration job", args[0])
}

func (s *migrationCmds) runMVMDMUpdate(cmd *cobra.Command, args []string) error {
	sourceParent, err := s.mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.mvmService(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/diskMigrationJobs/%s", sourceParent, args[0])
	if strings.HasPrefix(args[0], "projects/") {
		name = args[0]
	}
	body := &vmmigration.DiskMigrationJob{}
	if s.flagMVMConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagMVMConfigFile, body); err != nil {
			return err
		}
	}
	call := svc.Projects.Locations.Sources.DiskMigrationJobs.Patch(name, body).Context(ctx)
	if s.flagMVMUpdateMask != "" {
		call = call.UpdateMask(s.flagMVMUpdateMask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating disk migration job: %w", err)
	}
	return s.mvmFinishOp(ctx, svc, op, "Update disk migration job", args[0])
}
