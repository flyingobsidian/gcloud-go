package cmd

import (
	"context"
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

var (
	flagMVMMIISourceFile string
	flagMVMMIIName       string
	flagMVMMIITarget     string
	flagMVMDMSource      string
	flagMVMDMDescription string
)

// --- machine-image-imports ---

var mvmMachineImageImportsCmd = &cobra.Command{
	Use: "machine-image-imports", Short: "Manage VM Migration machine image imports",
}

var (
	mvmMIICreateCmd = &cobra.Command{
		Use: "create MACHINE_IMAGE_IMPORT", Short: "Create a machine image import",
		Args: cobra.ExactArgs(1), RunE: runMVMMIICreate,
	}
	mvmMIIDeleteCmd = &cobra.Command{
		Use: "delete MACHINE_IMAGE_IMPORT", Short: "Delete a machine image import",
		Args: cobra.ExactArgs(1), RunE: runMVMMIIDelete,
	}
	mvmMIIDescribeCmd = &cobra.Command{
		Use: "describe MACHINE_IMAGE_IMPORT", Short: "Describe a machine image import",
		Args: cobra.ExactArgs(1), RunE: runMVMMIIDescribe,
	}
	mvmMIIListCmd = &cobra.Command{
		Use: "list", Short: "List machine image imports",
		Args: cobra.NoArgs, RunE: runMVMMIIList,
	}
)

// --- disk-migrations ---

var mvmDiskMigrationsCmd = &cobra.Command{
	Use: "disk-migrations", Short: "Migrate disks to Compute Engine",
}

var (
	mvmDMCreateCmd = &cobra.Command{
		Use: "create DISK_MIGRATION_JOB", Short: "Create a disk migration job",
		Args: cobra.ExactArgs(1), RunE: runMVMDMCreate,
	}
	mvmDMDeleteCmd = &cobra.Command{
		Use: "delete DISK_MIGRATION_JOB", Short: "Delete a disk migration job",
		Args: cobra.ExactArgs(1), RunE: runMVMDMDelete,
	}
	mvmDMDescribeCmd = &cobra.Command{
		Use: "describe DISK_MIGRATION_JOB", Short: "Describe a disk migration job",
		Args: cobra.ExactArgs(1), RunE: runMVMDMDescribe,
	}
	mvmDMListCmd = &cobra.Command{
		Use: "list", Short: "List disk migration jobs for a source",
		Args: cobra.NoArgs, RunE: runMVMDMList,
	}
	mvmDMCancelCmd = &cobra.Command{
		Use: "cancel DISK_MIGRATION_JOB", Short: "Cancel a running disk migration job",
		Args: cobra.ExactArgs(1), RunE: runMVMDMCancel,
	}
	mvmDMRunCmd = &cobra.Command{
		Use: "run DISK_MIGRATION_JOB", Short: "Run a disk migration job",
		Args: cobra.ExactArgs(1), RunE: runMVMDMRun,
	}
	mvmDMUpdateCmd = &cobra.Command{
		Use: "update DISK_MIGRATION_JOB", Short: "Update a disk migration job",
		Args: cobra.ExactArgs(1), RunE: runMVMDMUpdate,
	}
)

func init() {
	addLoc := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagMVMLocation, "location", "", "Location (region)")
		}
	}
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagMVMFormat, "format", "", "Output format")
		}
	}
	addFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagMVMFilter, "filter", "", "Server-side list filter")
		}
	}
	addAsync := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().BoolVar(&flagMVMAsync, "async", false, "Do not wait for the operation to finish")
		}
	}

	// machine-image-imports
	addLoc(mvmMIICreateCmd, mvmMIIDeleteCmd, mvmMIIDescribeCmd, mvmMIIListCmd)
	addFmt(mvmMIICreateCmd, mvmMIIDescribeCmd, mvmMIIListCmd)
	addFilter(mvmMIIListCmd)
	addAsync(mvmMIICreateCmd, mvmMIIDeleteCmd)
	mvmMIICreateCmd.Flags().StringVar(&flagMVMMIISourceFile, "source-file", "", "Cloud Storage URI of the source machine image (required)")
	_ = mvmMIICreateCmd.MarkFlagRequired("source-file")
	mvmMIICreateCmd.Flags().StringVar(&flagMVMMIIName, "machine-image-name", "", "Target Compute Engine machine image name (defaults to import ID)")
	mvmMIICreateCmd.Flags().StringVar(&flagMVMMIITarget, "target-project", "", "Target project resource path (required)")
	_ = mvmMIICreateCmd.MarkFlagRequired("target-project")
	mvmMIICreateCmd.Flags().StringVar(&flagMVMConfigFile, "config-file", "", "JSON/YAML ImageImport body override")
	mvmMachineImageImportsCmd.AddCommand(mvmMIICreateCmd, mvmMIIDeleteCmd, mvmMIIDescribeCmd, mvmMIIListCmd)
	migrationVMsCmd.AddCommand(mvmMachineImageImportsCmd)

	// disk-migrations
	addLoc(mvmDMCreateCmd, mvmDMDeleteCmd, mvmDMDescribeCmd, mvmDMListCmd, mvmDMCancelCmd, mvmDMRunCmd, mvmDMUpdateCmd)
	addFmt(mvmDMCreateCmd, mvmDMDescribeCmd, mvmDMListCmd, mvmDMUpdateCmd)
	addFilter(mvmDMListCmd)
	addAsync(mvmDMCreateCmd, mvmDMDeleteCmd, mvmDMCancelCmd, mvmDMRunCmd, mvmDMUpdateCmd)
	for _, c := range []*cobra.Command{mvmDMCreateCmd, mvmDMDeleteCmd, mvmDMDescribeCmd, mvmDMListCmd, mvmDMCancelCmd, mvmDMRunCmd, mvmDMUpdateCmd} {
		c.Flags().StringVar(&flagMVMDMSource, "source", "", "Migration source resource path or ID (required)")
		_ = c.MarkFlagRequired("source")
	}
	mvmDMCreateCmd.Flags().StringVar(&flagMVMDMDescription, "description", "", "Optional description")
	mvmDMCreateCmd.Flags().StringVar(&flagMVMConfigFile, "config-file", "", "JSON/YAML DiskMigrationJob body override")
	mvmDMUpdateCmd.Flags().StringVar(&flagMVMUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	mvmDMUpdateCmd.Flags().StringVar(&flagMVMDMDescription, "description", "", "Update the description")
	mvmDMUpdateCmd.Flags().StringVar(&flagMVMConfigFile, "config-file", "", "JSON/YAML DiskMigrationJob body override")
	mvmDiskMigrationsCmd.AddCommand(mvmDMCreateCmd, mvmDMDeleteCmd, mvmDMDescribeCmd, mvmDMListCmd, mvmDMCancelCmd, mvmDMRunCmd, mvmDMUpdateCmd)
	migrationVMsCmd.AddCommand(mvmDiskMigrationsCmd)
}

// --- machine-image-imports impl ---

func runMVMMIICreate(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.ImageImport{CloudStorageUri: flagMVMMIISourceFile}
	if flagMVMConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagMVMConfigFile, body); err != nil {
			return err
		}
	}
	if body.MachineImageTargetDefaults == nil {
		body.MachineImageTargetDefaults = &vmmigration.MachineImageTargetDetails{}
	}
	target := body.MachineImageTargetDefaults
	if flagMVMMIIName != "" {
		target.MachineImageName = flagMVMMIIName
	} else if target.MachineImageName == "" {
		target.MachineImageName = args[0]
	}
	target.TargetProject = flagMVMMIITarget
	op, err := svc.Projects.Locations.ImageImports.Create(parent, body).ImageImportId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating machine image import: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Create machine image import", args[0])
}

func runMVMMIIDelete(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "imageImports", args[0])
	op, err := svc.Projects.Locations.ImageImports.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting machine image import: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Delete machine image import", args[0])
}

func runMVMMIIDescribe(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "imageImports", args[0])
	got, err := svc.Projects.Locations.ImageImports.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing machine image import: %w", err)
	}
	return emitFormatted(got, flagMVMFormat)
}

func runMVMMIIList(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	var all []*vmmigration.ImageImport
	pageToken := ""
	for {
		call := svc.Projects.Locations.ImageImports.List(parent).Context(ctx)
		if flagMVMFilter != "" {
			call = call.Filter(flagMVMFilter)
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
	if flagMVMFormat != "" {
		return emitFormatted(all, flagMVMFormat)
	}
	fmt.Printf("%-40s %-60s %s\n", "NAME", "CLOUD_STORAGE_URI", "CREATE_TIME")
	for _, ii := range all {
		fmt.Printf("%-40s %-60s %s\n", path.Base(ii.Name), ii.CloudStorageUri, ii.CreateTime)
	}
	return nil
}

// --- disk-migrations helpers ---

func mvmDMSourceParent() (string, error) {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return "", err
	}
	if flagMVMDMSource == "" {
		return "", fmt.Errorf("--source is required")
	}
	if strings.HasPrefix(flagMVMDMSource, "projects/") {
		return flagMVMDMSource, nil
	}
	return fmt.Sprintf("%s/sources/%s", parent, flagMVMDMSource), nil
}

// --- disk-migrations impl ---

func runMVMDMCreate(cmd *cobra.Command, args []string) error {
	sourceParent, err := mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.DiskMigrationJob{}
	if flagMVMConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagMVMConfigFile, body); err != nil {
			return err
		}
	}
	op, err := svc.Projects.Locations.Sources.DiskMigrationJobs.Create(sourceParent, body).DiskMigrationJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating disk migration job: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Create disk migration job", args[0])
}

func runMVMDMDelete(cmd *cobra.Command, args []string) error {
	sourceParent, err := mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
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
	return mvmFinishOp(ctx, svc, op, "Delete disk migration job", args[0])
}

func runMVMDMDescribe(cmd *cobra.Command, args []string) error {
	sourceParent, err := mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
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
	return emitFormatted(got, flagMVMFormat)
}

func runMVMDMList(cmd *cobra.Command, args []string) error {
	sourceParent, err := mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	var all []*vmmigration.DiskMigrationJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.Sources.DiskMigrationJobs.List(sourceParent).Context(ctx)
		if flagMVMFilter != "" {
			call = call.Filter(flagMVMFilter)
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
	if flagMVMFormat != "" {
		return emitFormatted(all, flagMVMFormat)
	}
	fmt.Printf("%-40s %-40s %s\n", "NAME", "STATE", "CREATE_TIME")
	for _, j := range all {
		fmt.Printf("%-40s %-40s %s\n", path.Base(j.Name), j.State, j.CreateTime)
	}
	return nil
}

func runMVMDMCancel(cmd *cobra.Command, args []string) error {
	sourceParent, err := mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
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
	return mvmFinishOp(ctx, svc, op, "Cancel disk migration job", args[0])
}

func runMVMDMRun(cmd *cobra.Command, args []string) error {
	sourceParent, err := mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
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
	return mvmFinishOp(ctx, svc, op, "Run disk migration job", args[0])
}

func runMVMDMUpdate(cmd *cobra.Command, args []string) error {
	sourceParent, err := mvmDMSourceParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/diskMigrationJobs/%s", sourceParent, args[0])
	if strings.HasPrefix(args[0], "projects/") {
		name = args[0]
	}
	body := &vmmigration.DiskMigrationJob{}
	if flagMVMConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagMVMConfigFile, body); err != nil {
			return err
		}
	}
	call := svc.Projects.Locations.Sources.DiskMigrationJobs.Patch(name, body).Context(ctx)
	if flagMVMUpdateMask != "" {
		call = call.UpdateMask(flagMVMUpdateMask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating disk migration job: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Update disk migration job", args[0])
}
